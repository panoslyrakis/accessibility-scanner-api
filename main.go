package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// LighthouseResult represents the structure of Lighthouse API response
type LighthouseResult struct {
	LighthouseResult struct {
		Categories struct {
			Accessibility struct {
				Score float64 `json:"score"`
				Title string  `json:"title"`
			} `json:"accessibility"`
		} `json:"categories"`
		Audits map[string]struct {
			ID               string  `json:"id"`
			Title            string  `json:"title"`
			Description      string  `json:"description"`
			Score            float64 `json:"score"`
			ScoreDisplayMode string  `json:"scoreDisplayMode"`
			Details          struct {
				Type  string `json:"type"`
				Items []struct {
					Node struct {
						Type     string `json:"type"`
						Selector string `json:"selector"`
						Snippet  string `json:"snippet"`
					} `json:"node"`
					Impact      string `json:"impact"`
					Description string `json:"description"`
				} `json:"items"`
			} `json:"details"`
		} `json:"audits"`
	} `json:"lighthouseResult"`
}

// AccessibilityIssue represents a single accessibility issue
type AccessibilityIssue struct {
	AuditID     string `json:"audit_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
	Selector    string `json:"selector"`
	Snippet     string `json:"snippet"`
}

// PageResult represents the accessibility results for a single page
type PageResult struct {
	URL                string               `json:"url"`
	AccessibilityScore float64              `json:"accessibility_score"`
	Issues             []AccessibilityIssue `json:"issues"`
	Error              string               `json:"error,omitempty"`
}

// ScanConfig represents the configuration used for scanning
type ScanConfig struct {
	MaxPages int `json:"max_pages"`
	Offset   int `json:"offset"`
	Limit    int `json:"limit"`
}

// ScanResult represents the complete scan results
type ScanResult struct {
	BaseURL        string       `json:"base_url"`
	ScanTime       time.Time    `json:"scan_time"`
	TotalPages     int          `json:"total_pages"`
	PageResults    []PageResult `json:"page_results"`
	UrlsDiscovered []string     `json:"urls_discovered"`
	UrlsVisited    []string     `json:"urls_visited"`
	ScanConfig     ScanConfig   `json:"scan_config"`
	Status         string       `json:"status"` // "completed", "failed", "partial"
}

// ScanRequest represents an API scan request
type ScanRequest struct {
	URL      string `json:"url"`
	MaxPages int    `json:"max_pages,omitempty"`
	Offset   int    `json:"offset,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// AccessibilityScanner handles the scanning process
type AccessibilityScanner struct {
	apiKey         string
	baseURL        string
	maxPages       int
	offset         int
	limit          int
	visited        map[string]bool
	urlsDiscovered []string
	client         *http.Client
}

// NewAccessibilityScanner creates a new scanner instance
func NewAccessibilityScanner(apiKey, baseURL string, maxPages, offset, limit int) *AccessibilityScanner {
	return &AccessibilityScanner{
		apiKey:         apiKey,
		baseURL:        baseURL,
		maxPages:       maxPages,
		offset:         offset,
		limit:          limit,
		visited:        make(map[string]bool),
		urlsDiscovered: make([]string, 0),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// loadEnvFile loads environment variables from a .env file
func loadEnvFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if (strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`)) ||
			(strings.HasPrefix(value, `'`) && strings.HasSuffix(value, `'`)) {
			value = value[1 : len(value)-1]
		}

		os.Setenv(key, value)
	}

	return scanner.Err()
}

// getAPIKey tries to get the API key from various sources
func getAPIKey() string {
	if key := os.Getenv("GOOGLE_API_KEY"); key != "" {
		return key
	}
	if key := os.Getenv("PAGESPEED_API_KEY"); key != "" {
		return key
	}
	if key := os.Getenv("LIGHTHOUSE_API_KEY"); key != "" {
		return key
	}
	return ""
}

// scanPageWithLighthouse scans a single page using Lighthouse API
func (s *AccessibilityScanner) scanPageWithLighthouse(pageURL string) PageResult {
	result := PageResult{URL: pageURL}

	lighthouseURL := fmt.Sprintf(
		"https://www.googleapis.com/pagespeedonline/v5/runPagespeed?url=%s&category=accessibility&key=%s",
		url.QueryEscape(pageURL),
		s.apiKey,
	)

	resp, err := s.client.Get(lighthouseURL)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to call Lighthouse API: %v", err)
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		result.Error = fmt.Sprintf("Lighthouse API error (status %d): %s", resp.StatusCode, string(body))
		return result
	}

	var lighthouseResult LighthouseResult
	if err := json.NewDecoder(resp.Body).Decode(&lighthouseResult); err != nil {
		result.Error = fmt.Sprintf("Failed to decode Lighthouse response: %v", err)
		return result
	}

	result.AccessibilityScore = lighthouseResult.LighthouseResult.Categories.Accessibility.Score

	for auditID, audit := range lighthouseResult.LighthouseResult.Audits {
		if audit.ScoreDisplayMode == "binary" && audit.Score < 1.0 {
			for _, item := range audit.Details.Items {
				issue := AccessibilityIssue{
					AuditID:     auditID,
					Title:       audit.Title,
					Description: audit.Description,
					Impact:      item.Impact,
					Selector:    item.Node.Selector,
					Snippet:     item.Node.Snippet,
				}
				result.Issues = append(result.Issues, issue)
			}

			if len(audit.Details.Items) == 0 {
				issue := AccessibilityIssue{
					AuditID:     auditID,
					Title:       audit.Title,
					Description: audit.Description,
					Impact:      "unknown",
					Selector:    "",
					Snippet:     "",
				}
				result.Issues = append(result.Issues, issue)
			}
		}
	}

	return result
}

// extractLinks extracts all internal links from an HTML page
func (s *AccessibilityScanner) extractLinks(pageURL string) ([]string, error) {
	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		return nil, err
	}

	customUA := "WPMUDEVAccessibilityScannerBot/1.0 (+mailto:panos.lyrakis@incsub.com; Purpose: Website Accessibility Testing)"
	req.Header.Set("User-Agent", customUA)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Connection", "keep-alive")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP error %d", resp.StatusCode)
	}

	baseURLParsed, err := url.Parse(s.baseURL)
	if err != nil {
		return nil, err
	}

	currentURLParsed, err := url.Parse(pageURL)
	if err != nil {
		return nil, err
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	var links []string

	var findLinks func(*html.Node)
	findLinks = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					linkURL, err := url.Parse(attr.Val)
					if err != nil {
						continue
					}

					absoluteURL := currentURLParsed.ResolveReference(linkURL)

					if absoluteURL.Host == baseURLParsed.Host {
						cleanURL := &url.URL{
							Scheme: absoluteURL.Scheme,
							Host:   absoluteURL.Host,
							Path:   absoluteURL.Path,
						}
						finalURL := cleanURL.String()

						isDuplicate := false
						for _, existing := range links {
							if existing == finalURL {
								isDuplicate = true
								break
							}
						}

						if !isDuplicate {
							links = append(links, finalURL)

							alreadyDiscovered := false
							for _, discovered := range s.urlsDiscovered {
								if discovered == finalURL {
									alreadyDiscovered = true
									break
								}
							}
							if !alreadyDiscovered {
								s.urlsDiscovered = append(s.urlsDiscovered, finalURL)
							}
						}
					}
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findLinks(c)
		}
	}

	findLinks(doc)
	return links, nil
}

// crawlAndScan performs the scanning with context support for cancellation
func (s *AccessibilityScanner) crawlAndScan(ctx context.Context) ScanResult {
	result := ScanResult{
		BaseURL:  s.baseURL,
		ScanTime: time.Now(),
		ScanConfig: ScanConfig{
			MaxPages: s.maxPages,
			Offset:   s.offset,
			Limit:    s.limit,
		},
		Status: "completed",
	}

	queue := []string{s.baseURL}
	s.visited[s.baseURL] = true
	s.urlsDiscovered = append(s.urlsDiscovered, s.baseURL)

	urlIndex := 0

	for len(queue) > 0 && len(result.PageResults) < s.limit {
		select {
		case <-ctx.Done():
			result.Status = "cancelled"
			break
		default:
		}

		currentURL := queue[0]
		queue = queue[1:]

		if urlIndex < s.offset {
			urlIndex++
			if len(queue) < s.maxPages {
				links, err := s.extractLinks(currentURL)
				if err == nil {
					for _, link := range links {
						if !s.visited[link] && len(queue) < s.maxPages {
							s.visited[link] = true
							queue = append(queue, link)
						}
					}
				}
			}
			continue
		}

		urlIndex++
		pageResult := s.scanPageWithLighthouse(currentURL)
		result.PageResults = append(result.PageResults, pageResult)

		time.Sleep(1 * time.Second)

		if pageResult.Error == "" && len(queue) < s.maxPages {
			links, err := s.extractLinks(currentURL)
			if err == nil {
				for _, link := range links {
					if !s.visited[link] && len(queue) < s.maxPages {
						s.visited[link] = true
						queue = append(queue, link)
					}
				}
			}
		}
	}

	result.TotalPages = len(result.PageResults)
	result.UrlsDiscovered = s.urlsDiscovered

	for _, pageResult := range result.PageResults {
		result.UrlsVisited = append(result.UrlsVisited, pageResult.URL)
	}

	if result.Status != "cancelled" && len(result.PageResults) == 0 {
		result.Status = "failed"
	} else if result.Status != "cancelled" {
		hasErrors := false
		for _, page := range result.PageResults {
			if page.Error != "" {
				hasErrors = true
				break
			}
		}
		if hasErrors {
			result.Status = "partial"
		}
	}

	return result
}

// API Handlers

// handleScan handles POST /api/v1/scan requests
func handleScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed, "Only POST method is supported")
		return
	}

	var req ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid JSON", http.StatusBadRequest, "Request body must be valid JSON")
		return
	}

	// Validate URL
	if req.URL == "" {
		sendError(w, "Missing URL", http.StatusBadRequest, "URL is required")
		return
	}

	if _, err := url.Parse(req.URL); err != nil {
		sendError(w, "Invalid URL", http.StatusBadRequest, "URL must be valid")
		return
	}

	// Set defaults
	if req.MaxPages == 0 {
		req.MaxPages = 50
	}
	if req.Limit == 0 {
		req.Limit = 5
	}

	// Validate ranges
	if req.MaxPages < 1 || req.MaxPages > 1000 {
		sendError(w, "Invalid max_pages", http.StatusBadRequest, "max_pages must be between 1 and 1000")
		return
	}
	if req.Limit < 1 || req.Limit > 100 {
		sendError(w, "Invalid limit", http.StatusBadRequest, "limit must be between 1 and 100")
		return
	}
	if req.Offset < 0 {
		sendError(w, "Invalid offset", http.StatusBadRequest, "offset cannot be negative")
		return
	}

	// Get API key
	apiKey := getAPIKey()
	if apiKey == "" {
		sendError(w, "Configuration error", http.StatusInternalServerError, "Google API key not configured")
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
	defer cancel()

	// Run scan
	scanner := NewAccessibilityScanner(apiKey, req.URL, req.MaxPages, req.Offset, req.Limit)
	result := scanner.crawlAndScan(ctx)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleHealth handles GET /health requests
func handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
		"service":   "accessibility-scanner",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// handleRoot handles GET / requests with API documentation
func handleRoot(w http.ResponseWriter, r *http.Request) {
	docs := map[string]interface{}{
		"service": "WPMUDEV Accessibility Scanner API",
		"version": "1.0.0",
		"endpoints": map[string]interface{}{
			"POST /api/v1/scan": map[string]interface{}{
				"description": "Scan a website for accessibility issues",
				"body": map[string]interface{}{
					"url":       "Website URL to scan (required)",
					"max_pages": "Maximum pages to discover (default: 50, max: 1000)",
					"offset":    "Skip first N pages (default: 0)",
					"limit":     "Maximum pages to scan (default: 5, max: 100)",
				},
				"example": map[string]interface{}{
					"url":       "https://example.com",
					"max_pages": 100,
					"offset":    10,
					"limit":     20,
				},
			},
			"GET /health": map[string]interface{}{
				"description": "Health check endpoint",
			},
		},
		"user_agent": "WPMUDEVAccessibilityScannerBot/1.0",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(docs)
}

// sendError sends a standardized error response
func sendError(w http.ResponseWriter, error string, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	response := ErrorResponse{
		Error:   error,
		Code:    code,
		Message: message,
	}

	json.NewEncoder(w).Encode(response)
}

// CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Logging middleware
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)

		log.Printf("%s %s %v", r.Method, r.URL.Path, duration)
	})
}

func main() {
	// Load environment variables
	if err := loadEnvFile(".env"); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Validate API key exists
	if getAPIKey() == "" {
		log.Fatal("Google API key not found. Please set GOOGLE_API_KEY environment variable or add to .env file.")
	}

	// Setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRoot)
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/api/v1/scan", handleScan)

	// Apply middleware
	handler := corsMiddleware(loggingMiddleware(mux))

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Accessibility Scanner API starting on port %s", port)
	log.Printf("🔑 Google API key configured: %t", getAPIKey() != "")
	log.Printf("🌐 Endpoints available:")
	log.Printf("   GET  / - API documentation")
	log.Printf("   GET  /health - Health check")
	log.Printf("   POST /api/v1/scan - Scan website")
	log.Printf("📡 Server ready on port %s", port)

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
