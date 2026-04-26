package news

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
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

	// Build query based on category - all pension-relevant for Kenya
	query := buildPensionNewsQuery(category)

	// Build URL with proper query encoding
	baseURL := "https://newsapi.org/v2/everything"
	reqURL := fmt.Sprintf("%s?q=%s&language=en&pageSize=%d&sortBy=publishedAt&apiKey=%s",
		baseURL, url.QueryEscape(query), pageSize, p.APIKey)

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
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
		detectedCategory := detectNewsCategory(a.Title, a.Description)
		articles = append(articles, Article{
			Title:       a.Title,
			Description: a.Description,
			Content:     a.Content,
			URL:         a.URL,
			URLToImage:  a.URLToImage,
			Source:      a.Source.Name,
			Author:      a.Author,
			PublishedAt: pubTime,
			Category:    detectedCategory,
		})
	}

	return &NewsResponse{
		Status:       "ok",
		TotalResults: apiResp.TotalResults,
		Articles:     articles,
		FetchedAt:    time.Now(),
	}, nil
}

func buildPensionNewsQuery(category string) string {
	// Base pension-related keywords for Kenya
	baseKeywords := "pension OR retirement OR \"retirement benefits\" OR \"RBA Kenya\" OR \"NSSF Kenya\" OR \"National Social Security Fund\""

	switch category {
	case "business":
		return fmt.Sprintf("(%s) AND (economy OR \"investment\" OR \"Treasury\" OR \"CBK\" OR \"Central Bank\" OR financial OR markets OR stock)", baseKeywords)
	case "politics":
		return fmt.Sprintf("(%s) AND (Parliament OR Senate OR \"National Assembly\" OR legislation OR government OR regulation OR \"RBA\" OR \"Authority\")", baseKeywords)
	case "health":
		return fmt.Sprintf("(%s) AND (health OR medical OR NHIF OR SHA OR \"Social Health\" OR hospital OR healthcare)", baseKeywords)
	case "technology":
		return fmt.Sprintf("(%s) AND (digital OR technology OR fintech OR mobile OR app OR blockchain OR \"online\" OR cybersecurity)", baseKeywords)
	default:
		return baseKeywords
	}
}

func detectNewsCategory(title, description string) string {
	text := strings.ToLower(title + " " + description)

	categoryKeywords := map[string][]string{
		"business":   {"economy", "investment", "treasury", "cbk", "central bank", "market", "stock", "shilling", "finance", "banking"},
		"politics":   {"parliament", "senate", "legislation", "bill", "government", "minister", "cs ", "mp ", "lawmakers"},
		"health":     {"health", "medical", "nhif", "sha", "hospital", "healthcare", "doctor", "clinic"},
		"technology": {"digital", "tech", "app", "mobile", "online", "blockchain", "cyber", "software", "platform"},
	}

	maxMatches := 0
	detected := "general"

	for cat, keywords := range categoryKeywords {
		matches := 0
		for _, kw := range keywords {
			if strings.Contains(text, kw) {
				matches++
			}
		}
		if matches > maxMatches {
			maxMatches = matches
			detected = cat
		}
	}

	return detected
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
	now := time.Now()

	// All articles with their categories
	allArticles := []Article{
		// BUSINESS & ECONOMY - Pension-related business news
		{
			Title:       "RBA Announces New Guidelines for Retirement Scheme Investments",
			Description: "The Retirement Benefits Authority has released updated investment guidelines affecting how pension funds can allocate assets, with new provisions for infrastructure investments.",
			URL:         "https://www.businessdaily.co.ke/news/rba-investment-guidelines",
			URLToImage:  "https://images.unsplash.com/photo-1611974789855-9c2a0a7236a3?w=800",
			Source:      "Business Daily",
			Author:      "Sarah Mutua",
			PublishedAt: now.Add(-1 * time.Hour),
			Category:    "business",
		},
		{
			Title:       "Treasury Proposes Tax Relief for Low-Income Pension Contributors",
			Description: "The National Treasury is considering amendments to the Income Tax Act that would increase tax relief thresholds for workers contributing to retirement schemes.",
			URL:         "https://www.standardmedia.co.ke/business/tax-relief-proposal",
			URLToImage:  "https://images.unsplash.com/photo-1554224155-6726b3ff858f?w=800",
			Source:      "The Standard",
			Author:      "James Ochieng",
			PublishedAt: now.Add(-3 * time.Hour),
			Category:    "business",
		},
		{
			Title:       "Pension Funds Warn of Challenges Amid Rising Interest Rates",
			Description: "Retirement benefits schemes report压力 as interest rate environment impacts investment returns and actuarial valuations.",
			URL:         "https://www.nation.co.ke/business/pension-funds-interest-rates",
			URLToImage:  "https://images.unsplash.com/photo-1579532537598-459ecdaf39cc?w=800",
			Source:      "Nation Media",
			Author:      "David Mwangi",
			PublishedAt: now.Add(-5 * time.Hour),
			Category:    "business",
		},
		{
			Title:       "CBA Reports Surge in Retirement Account Openings",
			Description: "Commercial banks report a 40% increase in individual retirement account openings as workers seek additional retirement savings beyond employer schemes.",
			URL:         "https://www.theeastafrican.co.ke/business/cba-retirement-accounts",
			URLToImage:  "https://images.unsplash.com/photo-1579621970563-ebec7560ff3e?w=800",
			Source:      "The EastAfrican",
			Author:      "Grace Wanjiku",
			PublishedAt: now.Add(-8 * time.Hour),
			Category:    "business",
		},
		{
			Title:       "NSE Sees Increased Pension Fund Investments in Blue Chip Stocks",
			Description: "Retirement benefits schemes have increased their equity allocations on the Nairobi Securities Exchange, driving up demand for blue-chip shares.",
			URL:         "https://www.businessdaily.co.ke/markets/nse-pension-investments",
			URLToImage:  "https://images.unsplash.com/photo-1590283603385-17ffb3a7f29f?w=800",
			Source:      "Business Daily",
			Author:      "Peter Kimani",
			PublishedAt: now.Add(-12 * time.Hour),
			Category:    "business",
		},
		// POLITICS & LEGISLATION
		{
			Title:       "Parliament Passes Amendment to Retirement Benefits Act",
			Description: "National Assembly approves changes to enhance trustee accountability and improve scheme governance standards across the pension sector.",
			URL:         "https://www.parliament.go.ke/news/pension-amendment",
			URLToImage:  "https://images.unsplash.com/photo-1541872703-74c5e44368f9?w=800",
			Source:      "KBC News",
			Author:      "Margaret Akinyi",
			PublishedAt: now.Add(-2 * time.Hour),
			Category:    "politics",
		},
		{
			Title:       "Senate Committee Reviews Portable Retirement Benefits Bill",
			Description: "New legislation proposed to allow workers to transfer pension benefits between schemes when changing jobs, improving labor mobility.",
			URL:         "https://www.standardmedia.co.ke/politics/portable-pensions-bill",
			URLToImage:  "https://images.unsplash.com/photo-1589829545856-d10d557cf95f?w=800",
			Source:      "The Standard",
			Author:      "Johnstone Kiptoo",
			PublishedAt: now.Add(-4 * time.Hour),
			Category:    "politics",
		},
		{
			Title:       "Ministry of Labour Announces Review of Minimum Pension Guarantees",
			Description: "Government undertaking comprehensive review of minimum pension guarantees to ensure adequate retirement income for low-wage workers.",
			URL:         "https://www.nation.co.ke/politics/minimum-pension-review",
			URLToImage:  "https://images.unsplash.com/photo-1507679799987-c73779587ccf?w=800",
			Source:      "Nation Media",
			Author:      "Elizabeth Njeri",
			PublishedAt: now.Add(-7 * time.Hour),
			Category:    "politics",
		},
		{
			Title:       "National Treasury Flags Delayed Regulations for Supplementary Schemes",
			Description: "New regulations for industry-wide, umbrella, and master trust schemes remain pending, prompting calls for expedited implementation.",
			URL:         "https://www.businessdaily.co.ke/politics/treasury-regulations-delay",
			URLToImage:  "https://images.unsplash.com/photo-1450101499163-c8848c66ca85?w=800",
			Source:      "Business Daily",
			Author:      "Michael Odhiambo",
			PublishedAt: now.Add(-10 * time.Hour),
			Category:    "politics",
		},
		{
			Title:       "County Governments Seek Clarity on Staff Pension Matters",
			Description: "Devolved units pushing for clear guidelines on pension responsibilities following constitutional transition of health workers to counties.",
			URL:         "https://www.theeastafrican.co.ke/news/county-pension-clarity",
			URLToImage:  "https://images.unsplash.com/photo-1529107386315-e1a2ed48a620?w=800",
			Source:      "The EastAfrican",
			Author:      "Catherine Muthoni",
			PublishedAt: now.Add(-14 * time.Hour),
			Category:    "politics",
		},
		// HEALTH
		{
			Title:       "SHA Implementation: Impact on Pensioners' Medical Coverage",
			Description: "As Social Health Authority takes over from NHIF, retirees need to understand changes to their medical scheme benefits and contribution requirements.",
			URL:         "https://www.the-star.co.ke/health/sha-pensioners-guide",
			URLToImage:  "https://images.unsplash.com/photo-1576091160550-2173dba999ef?w=800",
			Source:      "The Star",
			Author:      "Dr. Amina Hassan",
			PublishedAt: now.Add(-1 * time.Hour),
			Category:    "health",
		},
		{
			Title:       "Study Reveals High Healthcare Costs Among Kenyan Retirees",
			Description: "Research shows pensioners spend up to 30% of their retirement income on medical expenses, highlighting need for adequate health cover.",
			URL:         "https://www.nation.co.ke/health/retiree-healthcare-costs",
			URLToImage:  "https://images.unsplash.com/photo-1460925895917-afdab827c52f?w=800",
			Source:      "Nation Media",
			Author:      "Dr. Joseph Kariuki",
			PublishedAt: now.Add(-6 * time.Hour),
			Category:    "health",
		},
		{
			Title:       "Mental Health Services to Be Covered Under Basic Health Insurance",
			Description: "New regulations require all health schemes to include mental health coverage, benefiting retirees managing age-related conditions.",
			URL:         "https://www.standardmedia.co.ke/health/mental-health-coverage",
			URLToImage:  "https://images.unsplash.com/photo-1507679799987-c73779587ccf?w=800",
			Source:      "The Standard",
			Author:      "Faith Muthoni",
			PublishedAt: now.Add(-9 * time.Hour),
			Category:    "health",
		},
		{
			Title:       "Ministry of Health Urges Seniors to Register for SHA",
			Description: "Older persons encouraged to complete Social Health Authority registration before deadline to avoid disruption in medical coverage.",
			URL:         "https://www.kbc.co.ke/news/sha-senior-registration",
			URLToImage:  "https://images.unsplash.com/photo-1584515933487-779824d29309?w=800",
			Source:      "KBC News",
			Author:      "Robert Otieno",
			PublishedAt: now.Add(-11 * time.Hour),
			Category:    "health",
		},
		// TECHNOLOGY
		{
			Title:       "RBA Launches Digital Platform for Scheme Member Verification",
			Description: "New online portal allows pension scheme members to verify their contribution records and scheme registration status in real-time.",
			URL:         "https://www.businessdaily.co.ke/tech/rba-digital-platform",
			URLToImage:  "https://images.unsplash.com/photo-1551288049-bebda4e38f71?w=800",
			Source:      "Business Daily",
			Author:      "Kevin Ochieng",
			PublishedAt: now.Add(-2 * time.Hour),
			Category:    "technology",
		},
		{
			Title:       "Pension Schemes Embrace Blockchain for Improved Transparency",
			Description: "Several retirement benefits schemes are piloting blockchain technology to enhance audit trails and member data security.",
			URL:         "https://www.nation.co.ke/tech/pension-blockchain",
			URLToImage:  "https://images.unsplash.com/photo-1639762681485-074b7f938ba0?w=800",
			Source:      "Nation Media",
			Author:      "Alex Murimi",
			PublishedAt: now.Add(-5 * time.Hour),
			Category:    "technology",
		},
		{
			Title:       "Mobile App Allows Kenyans to Track Pension Contributions",
			Description: "New smartphone application enables workers to monitor their pension contributions and projected retirement benefits on the go.",
			URL:         "https://www.standardmedia.co.ke/tech/pension-mobile-app",
			URLToImage:  "https://images.unsplash.com/photo-1563986768609-322da13575f3?w=800",
			Source:      "The Standard",
			Author:      "Diana Wanjiku",
			PublishedAt: now.Add(-8 * time.Hour),
			Category:    "technology",
		},
		{
			Title:       "e-Government Services Portal Adds Pension-Related Self-Service",
			Description: "Citizens can now access pension statement requests and scheme transfer applications through the unified e-citizen platform.",
			URL:         "https://www.theeastafrican.co.ke/tech/e-government-pension",
			URLToImage:  "https://images.unsplash.com/photo-1556742049-0cfed4f6a45d?w=800",
			Source:      "The EastAfrican",
			Author:      "Patrick Njoroge",
			PublishedAt: now.Add(-13 * time.Hour),
			Category:    "technology",
		},
		// GENERAL
		{
			Title:       "Over 2 Million Kenyans Now Covered by Formal Pension Schemes",
			Description: "RBA reports significant growth in pension coverage, though penetration remains below regional averages despite recent regulatory reforms.",
			URL:         "https://www.businessdaily.co.ke/news/pension-coverage-growth",
			URLToImage:  "https://images.unsplash.com/photo-1552664730-d307ca884978?w=800",
			Source:      "Business Daily",
			Author:      "Maryanne Wangari",
			PublishedAt: now.Add(-1 * time.Hour),
			Category:    "general",
		},
		{
			Title:       "Financial Advisors Warn of Risks in DIY Retirement Planning",
			Description: "Experts caution workers against withdrawing pension savings early, citing long-term financial security concerns and tax implications.",
			URL:         "https://www.standardmedia.co.ke/life/diy-retirement-risks",
			URLToImage:  "https://images.unsplash.com/photo-1579621970563-ebec7560ff3e?w=800",
			Source:      "The Standard",
			Author:      "Samuel Kamau",
			PublishedAt: now.Add(-4 * time.Hour),
			Category:    "general",
		},
		{
			Title:       "Women in Kenya Face Higher Retirement Savings Gap",
			Description: "New study shows female workers have 35% less pension savings than male counterparts due to career breaks and wage disparities.",
			URL:         "https://www.nation.co.ke/news/women-pension-gap",
			URLToImage:  "https://images.unsplash.com/photo-1573497019940-1c28c88b4f3e?w=800",
			Source:      "Nation Media",
			Author:      "Lucy Achieng",
			PublishedAt: now.Add(-7 * time.Hour),
			Category:    "general",
		},
		{
			Title:       "Retired Civil Servants Receive Bonuses from Consolidated Fund",
			Description: "Government releases Ksh 15 billion to settle terminal benefits and bonuses for retired public sector employees.",
			URL:         "https://www.kbc.co.ke/news/civil-servant-bonuses",
			URLToImage:  "https://images.unsplash.com/photo-1521791136064-7986c2920216?w=800",
			Source:      "KBC News",
			Author:      "Christine Adhiambo",
			PublishedAt: now.Add(-10 * time.Hour),
			Category:    "general",
		},
		{
			Title:       "Industry Players Call for Public Awareness Campaign on Pension",
			Description: "Retirement benefits sector stakeholders urge government to invest in financial literacy programs to boost pension awareness.",
			URL:         "https://www.theeastafrican.co.ke/business/pension-awareness",
			URLToImage:  "https://images.unsplash.com/photo-1557804506-669a67965ba0?w=800",
			Source:      "The EastAfrican",
			Author:      "George Maina",
			PublishedAt: now.Add(-15 * time.Hour),
			Category:    "general",
		},
	}

	// Filter by category if specified
	var filtered []Article
	if category == "" || category == "all" || category == "general" {
		filtered = allArticles
	} else {
		for _, a := range allArticles {
			if a.Category == category {
				filtered = append(filtered, a)
			}
		}
		// If no articles in category, return general articles
		if len(filtered) == 0 {
			filtered = allArticles
		}
	}

	// Apply page size limit
	if pageSize > 0 && pageSize < len(filtered) {
		filtered = filtered[:pageSize]
	}

	return &NewsResponse{
		Status:       "ok",
		TotalResults: len(filtered),
		Articles:     filtered,
		FetchedAt:    now,
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
