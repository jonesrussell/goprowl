# Search Engine with Colly Integration

## Crawler Architecture

```plaintext
crawler/
├── collector/       # Colly collectors
│   ├── basic/      # Basic web crawler
│   ├── api/        # API crawler
│   └── sitemap/    # Sitemap crawler
├── parsers/        # Content parsers
│   ├── html/       # HTML content
│   ├── json/       # JSON data
│   └── feeds/      # RSS/Atom feeds
├── processors/     # Data processors
│   ├── extractor/  # Content extraction
│   ├── cleaner/    # Data cleaning
│   └── enricher/   # Data enrichment
└── queue/          # Crawl queue
    ├── memory/     # In-memory queue
    └── redis/      # Redis-backed queue
```

## Core Crawler Interfaces

```go
// Main crawler interface
type Crawler interface {
    // Configuration
    SetConfig(*CrawlerConfig) error
    
    // Crawling
    Start(seeds []string) error
    Stop() error
    Pause() error
    Resume() error
    
    // Status
    Status() *CrawlerStatus
    Stats() *CrawlerStats
}

// Colly-based collector
type WebCollector struct {
    collector *colly.Collector
    parser    ContentParser
    storage   SearchStorage
    
    // Configuration
    config    *CollectorConfig
    filters   []URLFilter
    rules     []CrawlRule
}

// Implementation example
func NewWebCollector(config *CollectorConfig) *WebCollector {
    c := colly.NewCollector(
        colly.AllowedDomains(config.AllowedDomains...),
        colly.MaxDepth(config.MaxDepth),
        colly.Async(config.Async),
    )
    
    return &WebCollector{
        collector: c,
        config:    config,
    }
}
```

## Crawling Rules

```go
// Define crawling behavior
type CrawlRule struct {
    // URL patterns
    URLPattern     string
    AllowSubpaths  bool
    
    // Extraction rules
    Selectors     map[string]string
    RequiredData  []string
    
    // Rate limiting
    RateLimit     time.Duration
    Parallelism   int
}

// Example configuration
var newsRule = &CrawlRule{
    URLPattern: `^https://blog\.domain\.com/\d{4}/\d{2}/.*$`,
    Selectors: map[string]string{
        "title":    "h1.post-title",
        "content":  "article.post-content",
        "author":   ".author-name",
        "date":     "time.published",
        "tags":     ".post-tags a",
    },
    RateLimit:   time.Second * 2,
    Parallelism: 3,
}
```

## Content Processing

```go
// Content processor interface
type ContentProcessor interface {
    // Processing
    Process(ctx context.Context, data *ScrapedData) (*SearchDocument, error)
    
    // Validation
    Validate(data *ScrapedData) error
    
    // Enrichment
    Enrich(doc *SearchDocument) error
}

// Implementation with Colly
func (w *WebCollector) setupCallbacks() {
    // Handle HTML
    w.collector.OnHTML("body", func(e *colly.HTMLElement) {
        data := &ScrapedData{
            URL:     e.Request.URL.String(),
            Title:   e.ChildText("h1.post-title"),
            Content: e.ChildText("article.post-content"),
            Author:  e.ChildText(".author-name"),
            Date:    e.ChildText("time.published"),
            Tags:    e.ChildTexts(".post-tags a"),
        }
        
        // Process the data
        processedData, err := w.parser.Parse(data)
        if err != nil {
            log.Errorf("Failed to parse content: %v", err)
            return
        }
        
        // Validate the data
        err = w.validator.Validate(processedData)
        if err != nil {
            log.Errorf("Failed to validate content: %v", err)
            return
        }
        
        // Enrich the data
        enrichedData, err := w.enricher.Enrich(processedData)
        if err != nil {
            log.Errorf("Failed to enrich content: %v", err)
            return
        }
        
        // Save the data to the storage
        err = w.storage.Save(enrichedData)
        if err != nil {
            log.Errorf("Failed to save content: %v", err)
            return
        }
    })
}
``` 