package news

import (
	"context"
	"testing"
	"time"
)

func TestMockProvider_FetchKenyaNews(t *testing.T) {
	provider := NewMockProvider()

	resp, err := provider.FetchKenyaNews(context.Background(), "", 0)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("Expected status 'ok', got: %s", resp.Status)
	}

	if len(resp.Articles) == 0 {
		t.Error("Expected articles, got none")
	}

	// Verify article structure
	for i, article := range resp.Articles {
		if article.Title == "" {
			t.Errorf("Article %d has empty title", i)
		}
		if article.Description == "" {
			t.Errorf("Article %d has empty description", i)
		}
		if article.URL == "" {
			t.Errorf("Article %d has empty URL", i)
		}
		if article.Source == "" {
			t.Errorf("Article %d has empty source", i)
		}
		if article.PublishedAt.IsZero() {
			t.Errorf("Article %d has zero published_at", i)
		}
	}
}

func TestMockProvider_FetchByCategory(t *testing.T) {
	provider := NewMockProvider()

	resp, err := provider.FetchKenyaNews(context.Background(), "business", 0)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(resp.Articles) == 0 {
		t.Error("Expected business articles, got none")
	}

	for _, article := range resp.Articles {
		if article.Category != "business" {
			t.Errorf("Expected category 'business', got: %s", article.Category)
		}
	}
}

func TestMockProvider_FetchWithPageSize(t *testing.T) {
	provider := NewMockProvider()

	resp, err := provider.FetchKenyaNews(context.Background(), "", 2)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(resp.Articles) != 2 {
		t.Errorf("Expected 2 articles, got: %d", len(resp.Articles))
	}
}

func TestMockProvider_Name(t *testing.T) {
	provider := NewMockProvider()

	if provider.Name() != "mock" {
		t.Errorf("Expected name 'mock', got: %s", provider.Name())
	}
}

func TestService_GetKenyaNews(t *testing.T) {
	provider := NewMockProvider()
	service := NewService(provider, 15*time.Minute)

	resp, err := service.GetKenyaNews(context.Background(), "", 5)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if resp.TotalResults == 0 {
		t.Error("Expected results, got 0")
	}

	if resp.FetchedAt.IsZero() {
		t.Error("Expected fetched_at to be set")
	}
}

func TestService_Cache(t *testing.T) {
	provider := NewMockProvider()
	service := NewService(provider, 15*time.Minute)

	// First fetch
	resp1, err := service.GetKenyaNews(context.Background(), "", 5)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Second fetch should return cached data
	resp2, err := service.GetKenyaNews(context.Background(), "", 5)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if resp1.FetchedAt != resp2.FetchedAt {
		t.Error("Expected same fetched_at for cached response")
	}
}

func TestService_ClearCache(t *testing.T) {
	provider := NewMockProvider()
	service := NewService(provider, 15*time.Minute)

	// Fetch and cache
	_, err := service.GetKenyaNews(context.Background(), "", 5)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	cachedAt := service.GetCachedAt()
	if cachedAt.IsZero() {
		t.Error("Expected cache timestamp")
	}

	// Clear cache
	service.ClearCache()

	if !service.GetCachedAt().IsZero() {
		t.Error("Expected zero timestamp after clearing cache")
	}
}

func TestService_ProviderName(t *testing.T) {
	provider := NewMockProvider()
	service := NewService(provider, 15*time.Minute)

	if service.ProviderName() != "mock" {
		t.Errorf("Expected provider name 'mock', got: %s", service.ProviderName())
	}
}

func TestService_DefaultCacheTTL(t *testing.T) {
	provider := NewMockProvider()
	service := NewService(provider, 0)

	if service.cacheTTL != 15*time.Minute {
		t.Errorf("Expected default cache TTL 15m, got: %v", service.cacheTTL)
	}
}

func TestNewsResponse_JSONFields(t *testing.T) {
	provider := NewMockProvider()
	resp, _ := provider.FetchKenyaNews(context.Background(), "", 1)

	article := resp.Articles[0]

	// Verify all expected fields are populated
	if article.Title == "" {
		t.Error("Title should not be empty")
	}
	if article.Description == "" {
		t.Error("Description should not be empty")
	}
	if article.Content == "" {
		t.Error("Content should not be empty")
	}
	if article.URL == "" {
		t.Error("URL should not be empty")
	}
	if article.Source == "" {
		t.Error("Source should not be empty")
	}
	if article.Category == "" {
		t.Error("Category should not be empty")
	}
}
