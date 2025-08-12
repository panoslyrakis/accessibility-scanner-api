# WPMUDEV Accessibility Scanner API

Website accessibility scanning service using Google Lighthouse via REST API.

## 🚀 Live API

**Production API:** [https://accessibility-scanner-api-production.up.railway.app](https://accessibility-scanner-api-production.up.railway.app)  
**Hosted on:** Railway (panoskatws@gmail.com)  
**Google PageSpeed API:** panos.lyrakis@gmail.com

## 📋 API Endpoints

### `POST /api/v1/scan`
Scan a website for accessibility issues.

**Request Body:**
```json
{
  "url": "https://example.com",
  "max_pages": 50,
  "offset": 0,
  "limit": 5
}
```

**Response:**
```json
{
  "base_url": "https://example.com",
  "scan_time": "2025-08-08T12:00:00Z",
  "status": "completed",
  "total_pages": 5,
  "scan_config": {
    "max_pages": 50,
    "offset": 0,
    "limit": 5
  },
  "urls_discovered": ["https://example.com/", "https://example.com/about"],
  "urls_visited": ["https://example.com/", "https://example.com/about"],
  "page_results": [
    {
      "url": "https://example.com/",
      "accessibility_score": 0.91,
      "issues": [
        {
          "audit_id": "heading-order",
          "title": "Heading elements are not in sequentially-descending order",
          "description": "...",
          "impact": "moderate",
          "selector": "h4.title",
          "snippet": "<h4>Title</h4>"
        }
      ]
    }
  ]
}
```

### `GET /health`
Health check endpoint.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-08-08T12:00:00Z",
  "version": "1.0.0",
  "service": "accessibility-scanner"
}
```

### `GET /`
API documentation and service information.

## 🎯 Usage Examples

### Basic Scan
```bash
curl -X POST https://accessibility-scanner-api-production.up.railway.app/api/v1/scan \
  -H "Content-Type: application/json" \
  -d '{"url": "https://dev3.candybits.eu/", "limit": 3}'
```

### Advanced Scan with Pagination
```bash
# Skip first 5 pages, scan next 10 pages
curl -X POST https://accessibility-scanner-api-production.up.railway.app/api/v1/scan \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://dev3.candybits.eu/",
    "max_pages": 100,
    "offset": 5,
    "limit": 10
  }'
```

### JavaScript/Frontend Usage
```javascript
const scanWebsite = async (url, options = {}) => {
  const response = await fetch('/api/v1/scan', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      url,
      max_pages: options.maxPages || 50,
      offset: options.offset || 0,
      limit: options.limit || 5
    })
  });
  
  return await response.json();
};

// Use it
const results = await scanWebsite('https://example.com', {
  limit: 10
});
console.log(`Found ${results.page_results.length} pages with accessibility data`);
```

## 🕷️ How the Scanner Works

The scanner uses a **queue-based breadth-first search**:

1. **Start with Homepage** → Add to queue
2. **Process Homepage** → Extract all internal links → Add new links to queue
3. **Process next URL in queue** → Extract links → Add new ones
4. **Continue until** limit reached or queue empty

### Parameters Explained

- **`max_pages`** (default: 50, max: 1000) - Maximum URLs to discover during crawling
- **`offset`** (default: 0) - Skip the first N discovered URLs  
- **`limit`** (default: 5, max: 100) - Maximum pages to actually scan with PageSpeed API
- **`url`** - Website URL to scan (required)

### Example Crawl Process

For a site with structure: Homepage → About → Contact → Blog → Posts...

**With `limit=5`:**
1. **Homepage** ✅ (scanned)
2. **About** ✅ (scanned)  
3. **Contact** ✅ (scanned)
4. **Blog** ✅ (scanned)
5. **Post 1** ✅ (scanned)

**Discovered but not scanned:** Post 2, Post 3, etc.

**To continue:** Use `offset=5, limit=5` for next batch.

## 🐳 Docker Development

### Quick Start with Docker
```bash
# Clone repository
git clone https://github.com/panoslyrakis/accessibility-scanner-api.git
cd accessibility-scanner-api

# Create environment file
cp .env.example .env
# Edit .env and add your Google API key

# Build and run with Docker Compose
docker-compose up --build

# API will be available at http://localhost:3001
```

### Docker Commands
```bash
# Build only
docker-compose build

# Run in background
docker-compose up -d

# View logs
docker-compose logs -f

# Stop service
docker-compose down

# Rebuild after code changes
docker-compose up --build
```

### Manual Docker Build
```bash
# Build image
docker build -t accessibility-api .

# Run container
docker run -p 3001:3001 \
  -e GOOGLE_API_KEY=your_key \
  -e PORT=3001 \
  accessibility-api
```

## ⚙️ Environment Variables

Create a `.env` file with:

```env
# Google PageSpeed Insights API Key (required)
GOOGLE_API_KEY=your_api_key_here

# Server port (default: 8080)
PORT=3001

# Environment setting
GO_ENV=development
```

### Getting Google PageSpeed API Key

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing one
3. Enable the **PageSpeed Insights API**
4. Go to **Credentials** → **Create Credentials** → **API Key**
5. Copy the API key to your `.env` file

**Current API Account:** panos.lyrakis@gmail.com  
**Project Link:** https://console.cloud.google.com/apis/credentials?project=pagespeed-accessibility-go

## 🔧 Local Development

### Prerequisites
- Go 1.23+
- Git

### Setup
```bash
# Clone repository
git clone https://github.com/panoslyrakis/accessibility-scanner-api.git
cd accessibility-scanner-api

# Install dependencies
go mod download

# Create environment file
cp .env.example .env
# Edit .env and add your API key

# Run development server
go run main.go

# Server starts on port from .env (default: 8080)
```

### Build Binary
```bash
# Build for current platform
go build -o accessibility-api main.go

# Run binary
./accessibility-api
```

### Cross-Platform Builds
```bash
# Windows 64-bit
GOOS=windows GOARCH=amd64 go build -o accessibility-api.exe main.go

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -o accessibility-api-macos main.go

# macOS Apple Silicon  
GOOS=darwin GOARCH=arm64 go build -o accessibility-api-macos-arm64 main.go

# Linux 64-bit
GOOS=linux GOARCH=amd64 go build -o accessibility-api-linux main.go
```

## 🏗️ Architecture

### Modular Design
- **Scanner Service** (this repo) - Returns JSON via REST API
- **UI Module** (separate) - Consumes API, displays results  
- **Storage Module** (separate) - Saves results to database/files

### Key Features
- **RESTful API** with proper HTTP methods
- **CORS enabled** for web applications  
- **Request validation** with helpful error messages
- **Context cancellation** for timeouts
- **Structured logging** for debugging
- **Status tracking** (completed/failed/partial/cancelled)
- **Rate limiting** respect for Google API
- **Custom User-Agent** for bot identification

### Bot Information
**User-Agent:** `WPMUDEVAccessibilityScannerBot/1.0 (+mailto:panos.lyrakis@incsub.com; Purpose: Website Accessibility Testing)`

If your site blocks this scanner, whitelist this User-Agent in your server/Cloudflare settings.

## 📊 Response Status Codes

- **200** - Success
- **400** - Bad Request (invalid parameters)
- **500** - Internal Server Error (API key issues, etc.)

### Response Status Field
- **`"completed"`** - All pages scanned successfully
- **`"partial"`** - Some pages had errors  
- **`"failed"`** - Scan failed completely
- **`"cancelled"`** - Scan was cancelled (timeout)

## 🚀 Deployment

### Railway (Current Production)
**Account:** panoskatws@gmail.com  
**URL:** https://accessibility-scanner-api-production.up.railway.app

Environment Variables in Railway:
- `GOOGLE_API_KEY` - Your PageSpeed Insights API key
- `PORT` - Automatically set by Railway

### Other Deployment Options
- **Fly.io** - Native Go support
- **Google Cloud Run** - Serverless, pay-per-request  
- **Heroku** - Traditional PaaS
- **DigitalOcean App Platform** - Simple deployment

## 📝 API Rate Limits

- **Google PageSpeed Insights API**: 25,000 requests/day (free tier)
- **This API**: No artificial limits (respects Google's limits)
- **Scan delays**: 1 second between PageSpeed API calls

## 🐛 Troubleshooting

### Common Issues

**"Google API key not found"**
- Check `.env` file contains `GOOGLE_API_KEY=your_key`
- Verify API key is enabled for PageSpeed Insights API

**"HTTP error 403 - site may be blocking"**  
- Website blocks the bot User-Agent
- Contact site owner to whitelist the bot
- Try scanning a different site to verify API works

**"Docker build fails"**
- Ensure `.env.example` file exists
- Check Go version matches (1.23+)

### Debug Mode
```bash
# Run with verbose logging
GO_ENV=development go run main.go

# Check API key configuration
curl http://localhost:3001/health
```

## 📄 License

© 2025 WPMUDEV - Incsub  
For internal use and development purposes.

## 🤝 Contributing

This is an internal WPMUDEV tool. For issues or enhancements:
1. Create GitHub issue
2. Submit pull request  
3. Contact: panos.lyrakis@incsub.com

---

**Made with ❤️ by WPMUDEV for better web accessibility** 🌐