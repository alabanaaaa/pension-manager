import { useState, useEffect, useCallback } from 'react';
import { news } from '../lib/api';
import { Loader2, RefreshCw, Clock, Newspaper, AlertCircle, ExternalLink, Wifi } from 'lucide-react';

export default function NewsPage() {
  const [articles, setArticles] = useState([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [fetchedAt, setFetchedAt] = useState(null);
  const [category, setCategory] = useState('');
  const [isLive, setIsLive] = useState(true);
  const [lastRefresh, setLastRefresh] = useState(Date.now());

  const fetchNews = useCallback(async (forceRefresh = false) => {
    if (forceRefresh) {
      setRefreshing(true);
      setIsLive(false);
    } else {
      setLoading(true);
    }
    
    try {
      // Force clear cache by calling refresh endpoint first, then fetch fresh news
      if (forceRefresh) {
        await news.refresh().catch(() => {});
      }
      
      const res = await news.get(category ? { category } : {});
      const data = res.data || {};
      
      // Handle different response formats
      let articlesData = [];
      if (Array.isArray(data)) {
        articlesData = data;
      } else if (data.articles) {
        articlesData = data.articles;
      } else if (data.data && Array.isArray(data.data)) {
        articlesData = data.data;
      }
      
      setArticles(articlesData);
      setFetchedAt(data.fetched_at || data.fetchedAt || new Date().toISOString());
      setLastRefresh(Date.now());
      setIsLive(true);
    } catch (err) {
      console.error('Failed to fetch news:', err);
      setArticles([]);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, [category]);

  // Initial load - fetch live news
  useEffect(() => {
    fetchNews(true); // Force refresh on mount
  }, []);

  // Auto-refresh every 5 minutes
  useEffect(() => {
    const interval = setInterval(() => {
      fetchNews(true);
    }, 5 * 60 * 1000); // 5 minutes
    
    return () => clearInterval(interval);
  }, [fetchNews]);

  // Category change handler
  const handleCategoryChange = (cat) => {
    setCategory(cat);
    fetchNews(true);
  };

  const categories = [
    { id: '', name: 'All News' },
    { id: 'business', name: 'Business & Economy' },
    { id: 'politics', name: 'Politics & Legislation' },
    { id: 'health', name: 'Health' },
    { id: 'technology', name: 'Technology' },
  ];

  const categoryColors = {
    business: 'bg-emerald-100 text-emerald-700 border-emerald-200',
    politics: 'bg-blue-100 text-blue-700 border-blue-200',
    health: 'bg-red-100 text-red-700 border-red-200',
    technology: 'bg-purple-100 text-purple-700 border-purple-200',
    general: 'bg-gray-100 text-gray-700 border-gray-200',
  };

  const formatDate = (dateStr) => {
    if (!dateStr) return 'Unknown';
    const date = new Date(dateStr);
    const now = new Date();
    const diffHours = Math.floor((now - date) / (1000 * 60 * 60));
    const diffMins = Math.floor((now - date) / (1000 * 60));
    
    if (diffMins < 1) return 'Just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffHours < 48) return 'Yesterday';
    return date.toLocaleDateString('en-GB', { day: 'numeric', month: 'short' });
  };

  const getTimeSinceRefresh = () => {
    const secs = Math.floor((Date.now() - lastRefresh) / 1000);
    if (secs < 60) return 'Just now';
    const mins = Math.floor(secs / 60);
    return `${mins}m ago`;
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-end sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-black">News & Announcements</h1>
          <p className="text-sm text-gray-500 mt-1">Latest updates affecting your pension</p>
        </div>
        <div className="flex items-center gap-3">
          <div className={`flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-medium ${
            isLive ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'
          }`}>
            <span className={`w-2 h-2 rounded-full ${isLive ? 'bg-green-500 animate-pulse' : 'bg-gray-400'}`} />
            {isLive ? 'Live' : 'Loading...'}
          </div>
          <span className="text-xs text-gray-400 flex items-center gap-1">
            <Clock size={12} />
            Updated {getTimeSinceRefresh()}
          </span>
          <button
            onClick={() => fetchNews(true)}
            disabled={refreshing}
            className="btn btn-secondary flex items-center gap-2"
          >
            <RefreshCw size={14} className={refreshing ? 'animate-spin' : ''} />
            {refreshing ? 'Refreshing...' : 'Refresh'}
          </button>
        </div>
      </div>

      {/* Categories */}
      <div className="flex gap-2 flex-wrap">
        {categories.map(cat => (
          <button
            key={cat.id}
            onClick={() => handleCategoryChange(cat.id)}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${
              category === cat.id
                ? 'bg-black text-white'
                : 'bg-white border border-gray-200 text-gray-600 hover:border-black hover:bg-gray-50'
            }`}
          >
            {cat.name}
          </button>
        ))}
      </div>

      {/* Loading State */}
      {loading && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {[1, 2, 3, 4].map(i => (
            <div key={i} className="bg-white rounded-lg border border-gray-200 p-4 animate-pulse">
              <div className="h-40 bg-gray-200 rounded mb-4" />
              <div className="h-4 bg-gray-200 rounded w-3/4 mb-2" />
              <div className="h-4 bg-gray-200 rounded w-1/2" />
            </div>
          ))}
        </div>
      )}

      {/* Empty State */}
      {!loading && articles.length === 0 && (
        <div className="bg-white rounded-lg border border-gray-200 p-16 text-center">
          <Wifi size={32} className="mx-auto text-gray-300 mb-3" />
          <p className="text-gray-500">No live news available</p>
          <p className="text-sm text-gray-400 mt-1">Click Refresh to fetch the latest news</p>
          <button
            onClick={() => fetchNews(true)}
            className="btn btn-primary mt-4"
          >
            Fetch Live News
          </button>
        </div>
      )}

      {/* Articles */}
      {!loading && articles.length > 0 && (
        <>
          {/* Featured Article */}
          {articles[0] && (
            <a 
              href={articles[0].url || articles[0].URL} 
              target="_blank" 
              rel="noopener noreferrer"
              className="block bg-white rounded-lg border border-gray-200 overflow-hidden hover:shadow-lg transition-all group"
            >
              <div className="grid grid-cols-1 md:grid-cols-2">
                <div className="relative h-48 md:h-auto bg-gray-100">
                  {articles[0].urlToImage || articles[0].URLToImage ? (
                    <img 
                      src={articles[0].urlToImage || articles[0].URLToImage} 
                      alt="" 
                      className="w-full h-full object-cover"
                      onError={e => { 
                        e.target.parentElement.innerHTML = `
                          <div class="w-full h-full bg-gradient-to-br from-gray-800 to-gray-600 flex items-center justify-center">
                            <svg class="w-12 h-12 text-white/30" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 20H5a2 2 0 01-2-2V6a2 2 0 012-2h10a2 2 0 012 2v1m2 13a2 2 0 01-2-2V7m2 13a2 2 0 002-2V9a2 2 0 00-2-2h-2m-4-3H9M7 16h6M7 8h6v4H7V8z"></path>
                            </svg>
                          </div>
                        `;
                      }}
                    />
                  ) : (
                    <div className="w-full h-full bg-gradient-to-br from-gray-800 to-gray-600 flex items-center justify-center">
                      <Newspaper size={48} className="text-white/30" />
                    </div>
                  )}
                  {(articles[0].source?.name || articles[0].Source?.name) && (
                    <span className="absolute top-3 left-3 px-2.5 py-1 bg-black/80 text-white text-xs font-medium rounded backdrop-blur-sm">
                      {articles[0].source?.name || articles[0].Source?.name}
                    </span>
                  )}
                </div>
                <div className="p-6">
                  <div className="flex items-center gap-2 mb-3">
                    <span className={`px-2 py-0.5 text-xs font-medium rounded border ${
                      categoryColors[articles[0].category || articles[0].Category] || categoryColors.general
                    }`}>
                      {articles[0].category || articles[0].Category || 'General'}
                    </span>
                    <span className="text-xs text-gray-400 flex items-center gap-1">
                      <Clock size={10} />
                      {formatDate(articles[0].publishedAt || articles[0].PublishedAt)}
                    </span>
                  </div>
                  <h2 className="text-xl font-bold text-black leading-tight mb-3 group-hover:underline">
                    {articles[0].title || articles[0].Title}
                  </h2>
                  {(articles[0].description || articles[0].Description) && (
                    <p className="text-sm text-gray-600 line-clamp-3 mb-4">
                      {articles[0].description || articles[0].Description}
                    </p>
                  )}
                  <span className="inline-flex items-center gap-1.5 text-sm font-medium text-black">
                    Read full article <ExternalLink size={12} />
                  </span>
                </div>
              </div>
            </a>
          )}

          {/* Article Grid */}
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {articles.slice(1).map((article, i) => (
              <a 
                key={i} 
                href={article.url || article.URL} 
                target="_blank" 
                rel="noopener noreferrer"
                className="bg-white rounded-lg border border-gray-200 overflow-hidden hover:shadow-md transition-all group flex flex-col"
              >
                {(article.urlToImage || article.URLToImage) && (
                  <div className="relative h-36 bg-gray-100 overflow-hidden">
                    <img 
                      src={article.urlToImage || article.URLToImage} 
                      alt="" 
                      className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
                      onError={e => { 
                        e.target.parentElement.innerHTML = `
                          <div class="w-full h-full bg-gray-200 flex items-center justify-center">
                            <svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 20H5a2 2 0 01-2-2V6a2 2 0 012-2h10a2 2 0 012 2v1m2 13a2 2 0 01-2-2V7m2 13a2 2 0 002-2V9a2 2 0 00-2-2h-2m-4-3H9M7 16h6M7 8h6v4H7V8z"></path>
                            </svg>
                          </div>
                        `;
                      }}
                    />
                    {(article.source?.name || article.Source?.name) && (
                      <span className="absolute bottom-2 left-2 px-2 py-0.5 bg-white/90 text-xs font-medium text-gray-800 rounded shadow-sm">
                        {article.source?.name || article.Source?.name}
                      </span>
                    )}
                  </div>
                )}
                <div className="p-4 flex flex-col flex-1">
                  <div className="flex items-center gap-2 mb-2">
                    <span className={`px-2 py-0.5 text-xs font-medium rounded border ${
                      categoryColors[article.category || article.Category] || categoryColors.general
                    }`}>
                      {article.category || article.Category || 'General'}
                    </span>
                    <span className="text-xs text-gray-400">{formatDate(article.publishedAt || article.PublishedAt)}</span>
                  </div>
                  <h3 className="font-semibold text-black leading-snug mb-2 line-clamp-2 group-hover:underline">
                    {article.title || article.Title}
                  </h3>
                  {(article.description || article.Description) && (
                    <p className="text-xs text-gray-500 line-clamp-2 mt-auto">
                      {article.description || article.Description}
                    </p>
                  )}
                </div>
              </a>
            ))}
          </div>
        </>
      )}
    </div>
  );
}
