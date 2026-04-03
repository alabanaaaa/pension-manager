package news

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Article represents a single news article
type Article struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Content     string    `json:"content"`
	URL         string    `json:"url"`
	URLToImage  string    `json:"url_to_image"`
	Source      string    `json:"source"`
	Author      string    `json:"author"`
	PublishedAt time.Time `json:"published_at"`
	Category    string    `json:"category"`
}

// NewsResponse holds the news API response
type NewsResponse struct {
	Status       string    `json:"status"`
	TotalResults int       `json:"total_results"`
	Articles     []Article `json:"articles"`
	FetchedAt    time.Time `json:"fetched_at"`
}

// Provider is the interface for news API providers
type Provider interface {
	FetchKenyaNews(ctx context.Context, category string, pageSize int) (*NewsResponse, error)
	Name() string
}

// NewsAPIProvider implements Provider using NewsAPI.org
type NewsAPIProvider struct {
	APIKey string
}

// NewNewsAPIProvider creates a new NewsAPI.org provider
func NewNewsAPIProvider(apiKey string) *NewsAPIProvider {
	return &NewsAPIProvider{APIKey: apiKey}
}

func (p *NewsAPIProvider) Name() string {
	return "newsapi"
}

func (p *NewsAPIProvider) FetchKenyaNews(ctx context.Context, category string, pageSize int) (*NewsResponse, error) {
	if p.APIKey == "" {
		return nil, fmt.Errorf("NewsAPI key not configured")
	}

	url := fmt.Sprintf("https://newsapi.org/v2/everything?q=Kenya+government&country=ke&category=%s&pageSize=%d&sortBy=publishedAt&apiKey=%s",
		category, pageSize, p.APIKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch news: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("news API returned status %d", resp.StatusCode)
	}

	var apiResp struct {
		Status       string `json:"status"`
		TotalResults int    `json:"totalResults"`
		Articles     []struct {
			Source      struct{ Name string } `json:"source"`
			Author      string                `json:"author"`
			Title       string                `json:"title"`
			Description string                `json:"description"`
			URL         string                `json:"url"`
			URLToImage  string                `json:"urlToImage"`
			Content     string                `json:"content"`
			PublishedAt string                `json:"publishedAt"`
		} `json:"articles"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if apiResp.Status != "ok" {
		return nil, fmt.Errorf("news API error: %s", apiResp.Status)
	}

	articles := make([]Article, 0, len(apiResp.Articles))
	for _, a := range apiResp.Articles {
		pubTime, _ := time.Parse(time.RFC3339, a.PublishedAt)
		articles = append(articles, Article{
			Title:       a.Title,
			Description: a.Description,
			Content:     a.Content,
			URL:         a.URL,
			URLToImage:  a.URLToImage,
			Source:      a.Source.Name,
			Author:      a.Author,
			PublishedAt: pubTime,
			Category:    category,
		})
	}

	return &NewsResponse{
		Status:       "ok",
		TotalResults: apiResp.TotalResults,
		Articles:     articles,
		FetchedAt:    time.Now(),
	}, nil
}

// MockProvider returns sample Kenyan government news for testing
type MockProvider struct{}

func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

func (p *MockProvider) Name() string {
	return "mock"
}

func (p *MockProvider) FetchKenyaNews(ctx context.Context, category string, pageSize int) (*NewsResponse, error) {
	articles := []Article{
		{
			Title:       "Kenya Government Announces New Pension Reforms for 2026",
			Description: "The Treasury has unveiled comprehensive reforms to the pension sector aimed at improving retirement benefits for all Kenyans.",
			Content:     "The Cabinet Secretary for the Treasury announced new pension reforms that will affect both public and private sector employees...",
			URL:         "https://example.com/news/pension-reforms-2026",
			URLToImage:  "https://example.com/images/pension-reforms.jpg",
			Source:      "The Standard",
			Author:      "Finance Reporter",
			PublishedAt: time.Now().Add(-2 * time.Hour),
			Category:    "business",
		},
		{
			Title:       "RBA Updates Retirement Benefits Regulations",
			Description: "The Retirement Benefits Authority has issued new guidelines for scheme administrators and trustees.",
			Content:     "The RBA has released updated regulations governing the management of retirement benefits schemes in Kenya...",
			URL:         "https://example.com/news/rba-regulations",
			URLToImage:  "https://example.com/images/rba.jpg",
			Source:      "Business Daily",
			Author:      "Regulatory Correspondent",
			PublishedAt: time.Now().Add(-6 * time.Hour),
			Category:    "business",
		},
		{
			Title:       "Kenya Shilling Stabilizes Against Dollar Amid New CBK Measures",
			Description: "The Central Bank of Kenya's latest monetary policy measures have helped stabilize the local currency.",
			Content:     "The Kenya shilling has shown signs of stability against the US dollar following the Central Bank's intervention...",
			URL:         "https://example.com/news/shilling-stabilizes",
			URLToImage:  "https://example.com/images/cbk.jpg",
			Source:      "Nation Media",
			Author:      "Economics Editor",
			PublishedAt: time.Now().Add(-12 * time.Hour),
			Category:    "business",
		},
		{
			Title:       "Parliament Debates New Tax Amendment Bill Affecting Retirement Benefits",
			Description: "Members of Parliament are reviewing proposed changes to tax laws that could impact retirement benefit taxation.",
			Content:     "The National Assembly is currently debating the Finance Bill amendments that include provisions on retirement benefits taxation...",
			URL:         "https://example.com/news/tax-amendment",
			URLToImage:  "https://example.com/images/parliament.jpg",
			Source:      "KBC News",
			Author:      "Political Reporter",
			PublishedAt: time.Now().Add(-1 * time.Hour),
			Category:    "politics",
		},
		{
			Title:       "NHIF to SHA Transition: What Pensioners Need to Know",
			Description: "The transition from NHIF to the new Social Health Authority has implications for retiree medical coverage.",
			Content:     "As Kenya transitions from the National Hospital Insurance Fund to the Social Health Authority, pensioners should be aware of changes to their medical coverage...",
			URL:         "https://example.com/news/nhif-sha-transition",
			URLToImage:  "https://example.com/images/health.jpg",
			Source:      "The Star",
			Author:      "Health Correspondent",
			PublishedAt: time.Now().Add(-4 * time.Hour),
			Category:    "health",
		},
	}

	if category != "" && category != "general" {
		var filtered []Article
		for _, a := range articles {
			if a.Category == category {
				filtered = append(filtered, a)
			}
		}
		if len(filtered) > 0 {
			articles = filtered
		}
	}

	if pageSize > 0 && pageSize < len(articles) {
		articles = articles[:pageSize]
	}

	return &NewsResponse{
		Status:       "ok",
		TotalResults: len(articles),
		Articles:     articles,
		FetchedAt:    time.Now(),
	}, nil
}

// Service manages news fetching with caching
type Service struct {
	provider  Provider
	cache     *NewsResponse
	cacheMu   sync.RWMutex
	cacheTTL  time.Duration
	lastFetch time.Time
}

// NewService creates a new news service
func NewService(provider Provider, cacheTTL time.Duration) *Service {
	if cacheTTL == 0 {
		cacheTTL = 15 * time.Minute
	}
	return &Service{
		provider: provider,
		cacheTTL: cacheTTL,
	}
}

// GetKenyaNews fetches news with caching
func (s *Service) GetKenyaNews(ctx context.Context, category string, pageSize int) (*NewsResponse, error) {
	// Check cache
	s.cacheMu.RLock()
	if s.cache != nil && time.Since(s.lastFetch) < s.cacheTTL {
		s.cacheMu.RUnlock()
		return s.cache, nil
	}
	s.cacheMu.RUnlock()

	// Fetch fresh news
	news, err := s.provider.FetchKenyaNews(ctx, category, pageSize)
	if err != nil {
		// Return cached data if available, even if stale
		s.cacheMu.RLock()
		if s.cache != nil {
			s.cacheMu.RUnlock()
			return s.cache, nil
		}
		s.cacheMu.RUnlock()
		return nil, fmt.Errorf("fetch news: %w", err)
	}

	// Update cache
	s.cacheMu.Lock()
	s.cache = news
	s.lastFetch = time.Now()
	s.cacheMu.Unlock()

	return news, nil
}

// GetCachedAt returns when the cache was last updated
func (s *Service) GetCachedAt() time.Time {
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()
	return s.lastFetch
}

// ClearCache clears the news cache
func (s *Service) ClearCache() {
	s.cacheMu.Lock()
	s.cache = nil
	s.lastFetch = time.Time{}
	s.cacheMu.Unlock()
}

// ProviderName returns the name of the current news provider
func (s *Service) ProviderName() string {
	return s.provider.Name()
}
