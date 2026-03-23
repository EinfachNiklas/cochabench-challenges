# Concurrent Web Crawler

## Task

Implement a complete concurrent web crawler in Go that crawls pages in parallel, extracts links, and respects multiple operational constraints.

Implement and preserve the following public types:

```go
type CrawlResult struct {
    URL          string
    Title        string
    Links        []string
    StatusCode   int
    ResponseTime time.Duration
    Error        error
    Depth        int
}

type CrawlerConfig struct {
    MaxDepth       int
    MaxConcurrent  int
    Timeout        time.Duration
    MaxPages       int
    RateLimitDelay time.Duration
    UserAgent      string
    FollowExternal bool
}

type Crawler struct {
    config CrawlerConfig
    // ... internal fields
}
```

Implement the following functions and behavior:

### `NewCrawler(config CrawlerConfig) (*Crawler, error)`

Create a new crawler and validate configuration:

- `MaxDepth >= 0`
- `MaxConcurrent > 0`
- `Timeout > 0`
- `MaxPages >= 0`, where `0` means unlimited

### `Crawl(ctx context.Context, startURL string) ([]CrawlResult, error)`

Main crawl function with the following features:

1. Concurrent crawling
   Use a worker-pool model with `MaxConcurrent` goroutines and a channel-based work queue.

2. Duplicate prevention
   Crawl each normalized URL at most once with thread-safe visited tracking.

3. Depth limiting
   The start URL has depth `0`, discovered links use `parent depth + 1`, and crawling stops at `MaxDepth`.

4. Page limiting
   Stop after `MaxPages` crawled pages, where `0` means unlimited.

5. Rate limiting
   Respect `RateLimitDelay` between requests to the same domain.

6. Context support
   Respect cancellation and timeout through the provided context and stop gracefully.

7. Error handling
   Store HTTP and crawl errors in `CrawlResult.Error` without failing the entire crawl for a single bad URL.

### `ExtractLinks(htmlContent string, baseURL *url.URL) ([]string, error)`

Extract all links from HTML:

- Find all `<a href="...">` elements
- Convert relative URLs into absolute URLs
- Filter invalid schemes such as `javascript:` and `mailto:`
- Remove URL fragments
- Return absolute URLs

Use `golang.org/x/net/html` for HTML parsing.

### `ExtractTitle(htmlContent string) string`

Extract the content of the `<title>` tag:

- Return an empty string when no title is found
- Trim surrounding whitespace

### `IsSameDomain(url1, url2 string) bool`

Compare whether two URLs belong to the same domain:

- `http://example.com` and `https://example.com` should be treated as the same domain
- `example.com` and `sub.example.com` should not

### `NormalizeURL(rawURL string) (string, error)`

Normalize a URL for duplicate detection:

- Lowercase scheme and host
- Remove fragments
- Remove trailing slash except for the root path `/`
- Sort query parameters if needed

### `GetDomain(rawURL string) (string, error)`

Extract the domain without the port:

- `http://example.com:8080/path` -> `example.com`

## Context

Web crawlers systematically visit pages and collect information from them. This challenge combines several core Go topics:

- Concurrency with goroutines, channels, and wait groups
- Context-based timeout and cancellation
- HTTP request and response handling
- HTML parsing for links and page titles
- Thread-safe coordination of shared state

The challenge is intended to exercise concurrent orchestration, URL handling, and resilient crawling behavior.

## Dependencies

- Go
- `golang.org/x/net/html` for HTML parsing

Typical local command:

```bash
go test ./...
```

## Constraints

- Do not change the provided public API
- Do not modify the tests
- Preserve the expected crawler behavior around concurrency, depth, page limits, and URL deduplication
- Respect context cancellation and request timeout semantics
- Keep shared state safe under concurrent access

Implementation guidance:

- A worker-pool design is recommended
- Track visited URLs with a mutex-protected map or `sync.Map`
- Track per-domain request timing for rate limiting
- Normalize URLs before duplicate checks

Expected complexity targets:

- Time: `O(n)` for `n` crawled pages under ideal concurrency
- Space: `O(n)` for visited URLs and results
- HTTP concurrency: up to `MaxConcurrent` simultaneous requests

## Edge Cases

- Invalid crawler configuration
- Invalid start URL
- Pages with relative links
- Duplicate links and cyclic navigation
- Reaching `MaxDepth`
- Reaching `MaxPages`
- Slow or timing-out responses
- External links when `FollowExternal` is disabled
- Invalid or unsupported URL schemes
- Pages without a `<title>` tag
