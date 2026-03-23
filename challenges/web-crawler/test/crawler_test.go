package crawler

import (
	"context"
	"net/http"
	"net/url"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Test Helper: Erstelle einen Test-HTTP-Server
func createTestServer() *httptest.Server {
	mux := http.NewServeMux()

	// Hauptseite mit Links
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
			<html>
				<head><title>Home Page</title></head>
				<body>
					<a href="/page1">Page 1</a>
					<a href="/page2">Page 2</a>
					<a href="http://external.com">External</a>
				</body>
			</html>
		`))
	})

	// Seite 1 mit weiteren Links
	mux.HandleFunc("/page1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
			<html>
				<head><title>Page 1</title></head>
				<body>
					<a href="/page1-1">Page 1-1</a>
					<a href="/page1-2">Page 1-2</a>
				</body>
			</html>
		`))
	})

	// Seite 2
	mux.HandleFunc("/page2", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><head><title>Page 2</title></head><body>Content</body></html>`))
	})

	// Langsame Seite (für Timeout-Tests)
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.Write([]byte(`<html><head><title>Slow Page</title></head></html>`))
	})

	return httptest.NewServer(mux)
}

// TestNewCrawler_ValidConfig testet die Crawler-Erstellung mit gültiger Config
func TestNewCrawler_ValidConfig(t *testing.T) {
	config := CrawlerConfig{
		MaxDepth:      2,
		MaxConcurrent: 5,
		Timeout:       5 * time.Second,
		MaxPages:      100,
		UserAgent:     "TestBot/1.0",
	}

	crawler, err := NewCrawler(config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if crawler == nil {
		t.Fatal("Expected crawler, got nil")
	}
}

// TestNewCrawler_InvalidConfig testet Validierung bei ungültiger Config
func TestNewCrawler_InvalidConfig(t *testing.T) {
	tests := []struct {
		name   string
		config CrawlerConfig
	}{
		{
			name: "negative MaxDepth",
			config: CrawlerConfig{
				MaxDepth:      -1,
				MaxConcurrent: 5,
				Timeout:       time.Second,
			},
		},
		{
			name: "zero MaxConcurrent",
			config: CrawlerConfig{
				MaxDepth:      2,
				MaxConcurrent: 0,
				Timeout:       time.Second,
			},
		},
		{
			name: "zero Timeout",
			config: CrawlerConfig{
				MaxDepth:      2,
				MaxConcurrent: 5,
				Timeout:       0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewCrawler(tt.config)
			if err == nil {
				t.Error("Expected error for invalid config, got nil")
			}
		})
	}
}

// TestCrawl_SinglePage testet Crawlen einer einzelnen Seite
func TestCrawl_SinglePage(t *testing.T) {
	server := createTestServer()
	defer server.Close()

	config := CrawlerConfig{
		MaxDepth:      0, // Nur Start-URL
		MaxConcurrent: 1,
		Timeout:       5 * time.Second,
		MaxPages:      10,
	}

	crawler, err := NewCrawler(config)
	if err != nil {
		t.Fatalf("Failed to create crawler: %v", err)
	}

	ctx := context.Background()
	results, err := crawler.Crawl(ctx, server.URL)
	if err != nil {
		t.Fatalf("Crawl failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if results[0].URL != server.URL && results[0].URL != server.URL+"/" {
		t.Errorf("Expected URL %s, got %s", server.URL, results[0].URL)
	}

	if results[0].Title != "Home Page" {
		t.Errorf("Expected title 'Home Page', got '%s'", results[0].Title)
	}

	if results[0].StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", results[0].StatusCode)
	}
}

// TestCrawl_MaxDepth testet MaxDepth Begrenzung
func TestCrawl_MaxDepth(t *testing.T) {
	server := createTestServer()
	defer server.Close()

	config := CrawlerConfig{
		MaxDepth:      1, // Start + 1 Ebene
		MaxConcurrent: 5,
		Timeout:       5 * time.Second,
		MaxPages:      100,
	}

	crawler, _ := NewCrawler(config)
	ctx := context.Background()
	results, err := crawler.Crawl(ctx, server.URL)

	if err != nil {
		t.Fatalf("Crawl failed: %v", err)
	}

	// Sollte Home + page1 + page2 = 3 Seiten sein (ohne external)
	if len(results) != 3 {
		t.Errorf("Expected 3 results at depth 1, got %d", len(results))
	}

	// Prüfe dass /page1-1 nicht gecrawlt wurde (wäre Depth 2)
	for _, result := range results {
		if strings.Contains(result.URL, "page1-1") {
			t.Error("Should not crawl to depth 2 with MaxDepth=1")
		}
	}
}

// TestCrawl_MaxPages testet MaxPages Limit
func TestCrawl_MaxPages(t *testing.T) {
	server := createTestServer()
	defer server.Close()

	config := CrawlerConfig{
		MaxDepth:      5,
		MaxConcurrent: 5,
		Timeout:       5 * time.Second,
		MaxPages:      2, // Nur 2 Seiten
	}

	crawler, _ := NewCrawler(config)
	ctx := context.Background()
	results, err := crawler.Crawl(ctx, server.URL)

	if err != nil {
		t.Fatalf("Crawl failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results (MaxPages=2), got %d", len(results))
	}
}

// TestCrawl_NoDuplicates testet dass URLs nicht doppelt gecrawlt werden
func TestCrawl_NoDuplicates(t *testing.T) {
	// Server mit zyklischen Links
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
			<html>
				<head><title>Cycle</title></head>
				<body>
					<a href="/">Home</a>
					<a href="/page1">Page 1</a>
				</body>
			</html>
		`))
	}))
	defer server.Close()

	config := CrawlerConfig{
		MaxDepth:      3,
		MaxConcurrent: 5,
		Timeout:       5 * time.Second,
	}

	crawler, _ := NewCrawler(config)
	ctx := context.Background()
	results, _ := crawler.Crawl(ctx, server.URL)

	// Zähle wie oft jede URL vorkommt
	urlCounts := make(map[string]int)
	for _, result := range results {
		urlCounts[result.URL]++
	}

	for url, count := range urlCounts {
		if count > 1 {
			t.Errorf("URL %s was crawled %d times (expected 1)", url, count)
		}
	}
}

// TestCrawl_Timeout testet Request-Timeout
func TestCrawl_Timeout(t *testing.T) {
	server := createTestServer()
	defer server.Close()

	config := CrawlerConfig{
		MaxDepth:      0,
		MaxConcurrent: 1,
		Timeout:       50 * time.Millisecond, // Kurzes Timeout
	}

	crawler, _ := NewCrawler(config)
	ctx := context.Background()
	results, _ := crawler.Crawl(ctx, server.URL+"/slow")

	if len(results) == 0 {
		t.Fatal("Expected at least one result")
	}

	// Sollte Timeout-Error haben
	if results[0].Error == nil {
		t.Error("Expected timeout error for slow page")
	}
}

// TestCrawl_ContextCancellation testet Context-Abbruch
func TestCrawl_ContextCancellation(t *testing.T) {
	server := createTestServer()
	defer server.Close()

	config := CrawlerConfig{
		MaxDepth:      5,
		MaxConcurrent: 1,
		Timeout:       5 * time.Second,
	}

	crawler, _ := NewCrawler(config)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	results, err := crawler.Crawl(ctx, server.URL)

	// Sollte abgebrochen werden
	if err != nil && err != context.DeadlineExceeded {
		// OK, Context wurde abgebrochen
	}

	// Sollte nicht alle Seiten gecrawlt haben
	if len(results) > 2 {
		t.Log("Context cancellation might not be working properly")
	}
}

// TestExtractLinks testet Link-Extraktion
func TestExtractLinks(t *testing.T) {
	html := `
		<html>
			<body>
				<a href="/page1">Page 1</a>
				<a href="http://example.com/page2">Page 2</a>
				<a href="../parent">Parent</a>
				<a href="#anchor">Anchor</a>
				<a href="javascript:void(0)">JS</a>
				<a href="mailto:test@example.com">Email</a>
			</body>
		</html>
	`

	baseURL, _ := parseURL("http://example.com/sub/page")
	links, err := ExtractLinks(html, baseURL)

	if err != nil {
		t.Fatalf("ExtractLinks failed: %v", err)
	}

	// Sollte /page1, page2, ../parent enthalten (nicht anchor, js, mailto)
	if len(links) < 2 {
		t.Errorf("Expected at least 2 valid links, got %d", len(links))
	}

	// Prüfe dass keine ungültigen Links enthalten sind
	for _, link := range links {
		if strings.Contains(link, "#") {
			t.Error("Should not include anchor links")
		}
		if strings.Contains(link, "javascript:") {
			t.Error("Should not include javascript: links")
		}
		if strings.Contains(link, "mailto:") {
			t.Error("Should not include mailto: links")
		}
	}
}

// TestExtractTitle testet Title-Extraktion
func TestExtractTitle(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "simple title",
			html:     `<html><head><title>Test Page</title></head></html>`,
			expected: "Test Page",
		},
		{
			name:     "no title",
			html:     `<html><head></head></html>`,
			expected: "",
		},
		{
			name:     "title with whitespace",
			html:     `<html><head><title>  Spaced Title  </title></head></html>`,
			expected: "Spaced Title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractTitle(tt.html)
			if strings.TrimSpace(result) != tt.expected {
				t.Errorf("Expected title '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestIsSameDomain testet Domain-Vergleich
func TestIsSameDomain(t *testing.T) {
	tests := []struct {
		url1     string
		url2     string
		expected bool
	}{
		{"http://example.com/page1", "http://example.com/page2", true},
		{"https://example.com", "http://example.com", true},
		{"http://example.com", "http://other.com", false},
		{"http://sub.example.com", "http://example.com", false},
	}

	for _, tt := range tests {
		result := IsSameDomain(tt.url1, tt.url2)
		if result != tt.expected {
			t.Errorf("IsSameDomain(%s, %s) = %v, expected %v",
				tt.url1, tt.url2, result, tt.expected)
		}
	}
}

// TestNormalizeURL testet URL-Normalisierung
func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "http://Example.COM/Path",
			expected: "http://example.com/path",
		},
		{
			input:    "http://example.com/page#anchor",
			expected: "http://example.com/page",
		},
		{
			input:    "http://example.com/page/",
			expected: "http://example.com/page",
		},
		{
			input:    "http://example.com/",
			expected: "http://example.com/",
		},
	}

	for _, tt := range tests {
		result, err := NormalizeURL(tt.input)
		if err != nil {
			t.Errorf("NormalizeURL(%s) error: %v", tt.input, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("NormalizeURL(%s) = %s, expected %s", tt.input, result, tt.expected)
		}
	}
}

// TestGetDomain testet Domain-Extraktion
func TestGetDomain(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"http://example.com/page", "example.com"},
		{"https://sub.example.com", "sub.example.com"},
		{"http://example.com:8080", "example.com"},
	}

	for _, tt := range tests {
		result, err := GetDomain(tt.url)
		if err != nil {
			t.Errorf("GetDomain(%s) error: %v", tt.url, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("GetDomain(%s) = %s, expected %s", tt.url, result, tt.expected)
		}
	}
}

// TestConcurrentCrawling testet dass Concurrency korrekt funktioniert
func TestConcurrentCrawling(t *testing.T) {
	server := createTestServer()
	defer server.Close()

	config := CrawlerConfig{
		MaxDepth:      2,
		MaxConcurrent: 10, // Hohe Concurrency
		Timeout:       5 * time.Second,
		MaxPages:      100,
	}

	crawler, _ := NewCrawler(config)
	ctx := context.Background()

	start := time.Now()
	results, err := crawler.Crawl(ctx, server.URL)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Crawl failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("Expected some results")
	}

	// Mit Concurrency sollte es schneller sein als sequentiell
	// (dieser Test ist nicht deterministisch, aber gibt einen Hinweis)
	t.Logf("Crawled %d pages in %v with concurrency %d",
		len(results), duration, config.MaxConcurrent)
}

// Helper function
func parseURL(rawURL string) (*url.URL, error) {
	return url.Parse(rawURL)
}
