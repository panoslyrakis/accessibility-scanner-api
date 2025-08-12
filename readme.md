# Google Pagespeed api key :
AIzaSyDUwaLQoQO-8JnbdUQSBjmYXmunFIVGFuU
Google project : https://console.cloud.google.com/apis/credentials?inv=1&invt=Ab46ZQ&project=pagespeed-accessibility-go

## Run file :
go run main.go https://dev3.candybits.eu/ 5

## Compile app:
go build -o accessibility-scanner main.go

<br />
<br />

# Run compiled app | Common usage Patterns:
## Quick test (just homepage + few pages)
./accessibility-scanner -limit=3 https://dev3.candybits.eu/

## Skip homepage, scan next 5 pages  
./accessibility-scanner -offset=1 -limit=5 https://dev3.candybits.eu/

## Large site analysis
./accessibility-scanner -max-pages=200 -limit=50 https://dev3.candybits.eu/

## Resume from page 20, scan 10 more
./accessibility-scanner -offset=20 -limit=10 https://dev3.candybits.eu/

<br />
<br />

# Build App
To build executables, run the **build.sh** (which runs only on macos and linux) :
`./build.sh` 
Builds for all OSs will be in the *builds* folder which should contain the following:
```
├── builds/
│   ├── accessibility-scanner-windows-amd64.exe
│   ├── accessibility-scanner-windows-386.exe
│   ├── accessibility-scanner-macos-amd64
│   ├── accessibility-scanner-macos-arm64
│   ├── accessibility-scanner-linux-amd64
│   ├── accessibility-scanner-linux-386
│   └── accessibility-scanner-linux-arm64
```

## To run any of the executable we need to make it executable eg :
`chmod +x builds/accessibility-scanner-macos-arm64` 






<br />
<br />
<br />



## 🕷️ **How the Scanner Works:**

The scanner uses a **queue-based breadth-first search**:

1.  **Start with Homepage** → Add to queue
2.  **Process Homepage** → Extract all internal links → Add new links to queue
3.  **Process next URL in queue** → Extract links → Add new ones
4.  **Continue until** limit reached or queue empty

## 📊 **Example Crawl Order:**


### **Site Structure** (example)

```
Homepage
├── About, Contact, Who We Are, Cart, Latest Post

About
├── Homepage, About, Contact, Who We Are, Cart, Latest Post (nav)
├── Page A, Page B

Contact  
├── Homepage, About, Contact, Who We Are, Cart, Latest Post (nav)

Cart
├── Homepage, About, Contact, Who We Are, Cart, Latest Post (nav)

Latest Post
├── Homepage, About, Contact, Who We Are, Cart, Latest Post (nav)  
├── Post A, Post B, Post C

Post A
├── Homepage, About, Contact, Who We Are, Cart, Latest Post (nav)
├── Post B
```


### **Queue Processing:**


```
Queue Start: [Homepage]

Step 1: Process Homepage
├── Scan: Homepage ✅ (scanned #1)
├── Found links: About, Contact, Who-We-Are, Cart, Latest-Post
└── Queue: [About, Contact, Who-We-Are, Cart, Latest-Post]

Step 2: Process About  
├── Scan: About ✅ (scanned #2)
├── Found NEW links: Page-A, Page-B (others already discovered)
└── Queue: [Contact, Who-We-Are, Cart, Latest-Post, Page-A, Page-B]

Step 3: Process Contact
├── Scan: Contact ✅ (scanned #3) 
├── Found: No new links (all already discovered)
└── Queue: [Who-We-Are, Cart, Latest-Post, Page-A, Page-B]

Step 4: Process Who-We-Are
├── Scan: Who-We-Are ✅ (scanned #4)
├── Found: No new links
└── Queue: [Cart, Latest-Post, Page-A, Page-B]

Step 5: Process Cart
├── Scan: Cart ✅ (scanned #5)
├── LIMIT REACHED! Stop scanning
└── Queue: [Latest-Post, Page-A, Page-B] (discovered but not scanned)
```


### **Scan Order with `-limit=5`:**

1.  **Homepage** ✅
2.  **About** ✅
3.  **Contact** ✅
4.  **Who We Are** ✅
5.  **Cart** ✅

**Discovered but NOT scanned:** Latest-Post, Page-A, Page-B, Post-A, Post-B, Post-C

### **To scan ALL pages, you'd need multiple runs:**

**Run 1:** `-limit=5`

-   Scans: Homepage → About → Contact → Who-We-Are → Cart

**Run 2:** `-offset=5 -limit=5`

-   Scans: Latest-Post → Page-A → Page-B → Post-A → Post-B

**Run 3:** `-offset=10 -limit=5`

-   Scans: Post-C (and any other discovered pages)

## 📈 **Key Insights:**

### **Deduplication Works:**

-   Each URL only discovered **once**
-   Navigation links don't create duplicates
-   Queue grows efficiently

### **Breadth-First Benefits:**

-   **Prioritizes main navigation** (About, Contact, etc.)
-   **Finds structure quickly** before going deep
-   **Balanced coverage** across site sections

### **Discovery vs Scanning:**

-   **Discovery** is fast (just parsing HTML)
-   **Scanning** is slow (PageSpeed API calls)
-   You'll discover **way more** than you scan

## 🔍 **Example Output:**

```
📊 Configuration: max-pages=50, offset=0, limit=5
URLs Discovered: 12, URLs Scanned: 5

🔍 URLs discovered but not scanned (7):
  • https://site.com/latest-post
  • https://site.com/page-a  
  • https://site.com/page-b
  • https://site.com/post-a
  • https://site.com/post-b
  • https://site.com/post-c
```

**Perfect for incremental scanning of large sites!** 🚀


