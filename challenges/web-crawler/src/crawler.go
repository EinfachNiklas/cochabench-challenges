package crawler

import (
	"context"
	"net/url"
	"time"
)

// CrawlResult represents the result of a crawl operation for a URL
type CrawlResult struct {
	URL          string        // The crawled URL
	Title        string        // Page title (from <title> tag)
	Links        []string      // Discovered links (absolute URLs)
	StatusCode   int           // HTTP status code
	ResponseTime time.Duration // Response time
	Error        error         // Error if one occurred
	Depth        int           // Depth in the crawl tree (start = 0)
}

// CrawlerConfig holds the configuration for the crawler
type CrawlerConfig struct {
	MaxDepth       int           // Maximum crawl depth (0 = start URL only)
	MaxConcurrent  int           // Maximum number of concurrent requests
	Timeout        time.Duration // Timeout per request
	MaxPages       int           // Maximum number of pages to crawl (0 = unlimited)
	RateLimitDelay time.Duration // Minimum time between requests to the same domain
	UserAgent      string        // User-agent string
	FollowExternal bool          // Follow external links?
}

// Crawler is the main crawler struct
type Crawler struct {
	config CrawlerConfig
	// TODO: Add internal fields (e.g. visited URLs, rate limiter, etc.)
}

// NewCrawler creates a new crawler with the given configuration
//
// Validation:
// - MaxDepth must be >= 0
// - MaxConcurrent must be > 0
// - Timeout must be > 0
// - MaxPages must be >= 0 (0 = unlimited)
//
// Returns:
//   - Initialized crawler
//   - error for invalid configuration
func NewCrawler(config CrawlerConfig) (*Crawler, error) {
	// TODO: Implement validation and initialization
	return nil, nil
}

// Crawl starts the crawl operation from the given start URL
//
// The crawler:
// - Crawls the start URL and follows discovered links up to MaxDepth
// - Respects MaxConcurrent (worker pool pattern)
// - Stops when MaxPages pages have been crawled
// - Crawls each URL only once (duplicate prevention)
// - Respects RateLimitDelay per domain
// - Can be cancelled early via context
//
// Args:
//   - ctx: Context for timeout/cancellation
//   - startURL: The start URL (must be a valid HTTP(S) URL)
//
// Returns:
//   - Slice of all CrawlResults (including errors)
//   - error for fatal errors (e.g. invalid start URL)
func (c *Crawler) Crawl(ctx context.Context, startURL string) ([]CrawlResult, error) {
	// TODO: Implement the crawl algorithm
	// Tips:
	// - Use goroutines for concurrency
	// - Channel for work queue and result collection
	// - sync.WaitGroup or context for coordination
	// - Map + mutex for visited URLs (or sync.Map)
	// - time.Ticker for rate limiting
	return nil, nil
}

// ExtractLinks extracts all links from an HTML page
//
// The function:
// - Parses HTML and finds all <a href="..."> tags
// - Converts relative URLs to absolute URLs
// - Filters invalid URLs (e.g. javascript:, mailto:, #anchors)
// - Normalizes URLs (removes fragments, etc.)
//
// Args:
//   - htmlContent: The HTML content as a string
//   - baseURL: The base URL for relative links
//
// Returns:
//   - Slice of absolute URLs
//   - error on parse failures
func ExtractLinks(htmlContent string, baseURL *url.URL) ([]string, error) {
	// TODO: Implement link extraction
	// Tip: Use golang.org/x/net/html for HTML parsing
	return nil, nil
}

// ExtractTitle extracts the title of an HTML page
//
// Args:
//   - htmlContent: The HTML content as a string
//
// Returns:
//   - The content of the <title> tag (or "" if not found)
func ExtractTitle(htmlContent string) string {
	// TODO: Implement title extraction
	return ""
}

// IsSameDomain checks whether two URLs belong to the same domain
//
// Args:
//   - url1, url2: The URLs to compare
//
// Returns:
//   - true if both belong to the same domain
func IsSameDomain(url1, url2 string) bool {
	// TODO: Implement domain comparison
	return false
}

// NormalizeURL normalizes a URL
//
// Normalization:
// - Removes fragment (#...)
// - Removes trailing slash from paths (except root)
// - Converts to lowercase (scheme and host)
// - Sorts query parameters
//
// Args:
//   - rawURL: The URL to normalize
//
// Returns:
//   - Normalized URL as string
//   - error for invalid URLs
func NormalizeURL(rawURL string) (string, error) {
	// TODO: Implement URL normalization
	return "", nil
}

// GetDomain extracts the domain from a URL
//
// Args:
//   - rawURL: The URL
//
// Returns:
//   - Domain (e.g. "example.com")
//   - error for invalid URLs
func GetDomain(rawURL string) (string, error) {
	// TODO: Implement domain extraction
	return "", nil
}
