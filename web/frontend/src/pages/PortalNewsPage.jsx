import { useState, useEffect, useCallback } from 'react';
import { news } from '../lib/api';
import { Loader2, RefreshCw, Clock, Newspaper, AlertCircle, ExternalLink, Bookmark, Wifi } from 'lucide-react';

export default function PortalNewsPage() {
  const [articles, setArticles] = useState([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [category, setCategory] = useState('');
  const [isLive, setIsLive] = useState(false);
  const [bookmarks, setBookmarks] = useState([]);
  const [lastRefresh, setLastRefresh] = useState(Date.now());

  useEffect(() => {
    const saved = localStorage.getItem('news_bookmarks');
    if (saved) setBookmarks(JSON.parse(saved));
  }, []);

  const fetchNews = useCallback(async (forceRefresh = false) => {
    if (forceRefresh) {
      setRefreshing(true);
      setIsLive(false);
    } else {
      setLoading(true);
    }
    
    try {
      // Force clear cache and get fresh news
      if (forceRefresh) {
        await news.refresh().catch(() => {});
      }
      
      const res = await news.getPublic(category ? { category } : {});
      const data = res.data || {};
      
      let articlesData = [];
      if (Array.isArray(data)) {
        articlesData = data;
      } else if (data.articles) {
        articlesData = data.articles;
      }
      
      setArticles(articlesData);
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
    fetchNews(true);
  }, []);

  // Auto-refresh every 5 minutes
  useEffect(() => {
    const interval = setInterval(() => {
      fetchNews(true);
    }, 5 * 60 * 1000);
    
    return () => clearInterval(interval);
  }, [fetchNews]);

  const handleCategoryChange = (cat) => {
    setCategory(cat);
    fetchNews(true);
  };

  const categories = [
    { id: '', name: 'All' },
    { id: 'business', name: 'Business' },
    { id: 'politics', name: 'Legislation' },
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
    if (!dateStr) return '';
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

  const toggleBookmark = (article) => {
    const url = article.url || article.URL;
    if (bookmarks.includes(url)) {
      const updated = bookmarks.filter(b => b !== url);
      setBookmarks(updated);
      localStorage.setItem('news_bookmarks', JSON.stringify(updated));
    } else {
      const updated = [...bookmarks, url];
      setBookmarks(updated);
      localStorage.setItem('news_bookmarks', JSON.stringify(updated));
    }
  };

  const isBookmarked = (article) => {
    const url = article.url || article.URL;
    return bookmarks.includes(url);
  };

  const getTimeSinceRefresh = () => {
    const secs = Math.floor((Date.now() - lastRefresh) / 1000);
    if (secs < 60) return 'Just now';
    const mins = Math.floor(secs / 60);
    return `${mins}m ago`;
  };

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-end sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-black">News & Updates</h1>
          <p className="text-sm text-gray-500 mt-1">Latest news affecting your pension</p>
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
            {getTimeSinceRefresh()}
          </span>
          <button
            onClick={() => fetchNews(true)}
            disabled={refreshing}
            className="btn btn-secondary flex items-center gap-2"
          >
            <RefreshCw size={14} className={refreshing ? 'animate-spin' : ''} />
            {refreshing ? 'Updating...' : 'Update'}
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
                : 'bg-white border border-gray-200 text-gray-600 hover:border-black'
            }`}
          >
            {cat.name}
          </button>
        ))}
      </div>

      {/* Loading */}
      {loading && (
        <div className="space-y-4">
          {[1, 2, 3, 4, 5].map(i => (
            <div key={i} className="bg-white rounded-lg border border-gray-200 p-4 animate-pulse">
              <div className="flex gap-4">
                <div className="w-24 h-24 bg-gray-200 rounded-lg flex-shrink-0" />
                <div className="flex-1">
                  <div className="h-3 bg-gray-200 rounded w-1/4 mb-2" />
                  <div className="h-4 bg-gray-200 rounded w-3/4 mb-2" />
                  <div className="h-3 bg-gray-200 rounded w-1/2" />
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Empty */}
      {!loading && articles.length === 0 && (
        <div className="bg-white rounded-lg border border-gray-200 p-16 text-center">
          <Wifi size={32} className="mx-auto text-gray-300 mb-3" />
          <p className="text-gray-500">No live news available</p>
          <p className="text-sm text-gray-400 mt-1">Click Update to fetch the latest news</p>
          <button onClick={() => fetchNews(true)} className="btn btn-primary mt-4">
            Fetch Live News
          </button>
        </div>
      )}

      {/* Articles */}
      {!loading && articles.length > 0 && (
        <div className="space-y-4">
          {articles.map((article, i) => (
            <div 
              key={i} 
              className={`bg-white rounded-lg border border-gray-200 overflow-hidden hover:shadow-md transition-all ${
                i === 0 ? 'border-l-4 border-l-black' : ''
              }`}
            >
              <div className="flex gap-4 p-4">
                {/* Image */}
                {(article.urlToImage || article.URLToImage) && (
                  <a 
                    href={article.url || article.URL}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="flex-shrink-0 w-24 h-24 sm:w-32 sm:h-24 rounded-lg overflow-hidden bg-gray-100 hover:opacity-90 transition-opacity"
                  >
                    <img 
                      src={article.urlToImage || article.URLToImage} 
                      alt="" 
                      className="w-full h-full object-cover"
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
                  </a>
                )}

                {/* Content */}
                <div className="flex-1 min-w-0">
                  <div className="flex items-start justify-between gap-3">
                    <div className="flex-1">
                      {/* Meta */}
                      <div className="flex items-center gap-2 mb-2">
                        <span className={`px-2 py-0.5 text-xs font-medium rounded border ${
                          categoryColors[article.category || article.Category] || categoryColors.general
                        }`}>
                          {article.category || article.Category || 'General'}
                        </span>
                        {(article.source?.name || article.Source?.name) && (
                          <span className="text-xs text-gray-500">{article.source?.name || article.Source?.name}</span>
                        )}
                        <span className="text-xs text-gray-400 flex items-center gap-1">
                          <Clock size={10} />
                          {formatDate(article.publishedAt || article.PublishedAt)}
                        </span>
                      </div>

                      {/* Title */}
                      <a 
                        href={article.url || article.URL}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="block"
                      >
                        <h3 className="font-semibold text-black leading-snug mb-1 hover:underline">
                          {article.title || article.Title}
                        </h3>
                      </a>

                      {/* Description */}
                      {(article.description || article.Description) && (
                        <p className="text-sm text-gray-600 line-clamp-2">
                          {article.description || article.Description}
                        </p>
                      )}
                    </div>

                    {/* Actions */}
                    <div className="flex items-center gap-1 flex-shrink-0">
                      <button
                        onClick={() => toggleBookmark(article)}
                        className={`p-2 rounded-lg transition-all ${
                          isBookmarked(article)
                            ? 'bg-black text-white'
                            : 'hover:bg-gray-100 text-gray-400'
                        }`}
                        title={isBookmarked(article) ? 'Remove bookmark' : 'Save article'}
                      >
                        <Bookmark size={16} fill={isBookmarked(article) ? 'currentColor' : 'none'} />
                      </button>
                      <a 
                        href={article.url || article.URL}
                        target="_blank" 
                        rel="noopener noreferrer"
                        className="p-2 rounded-lg hover:bg-gray-100 text-gray-400 transition-all"
                        title="Read full article"
                      >
                        <ExternalLink size={16} />
                      </a>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Saved Articles */}
      {bookmarks.length > 0 && (
        <div className="mt-8 pt-6 border-t border-gray-200">
          <h2 className="text-lg font-semibold text-black mb-4 flex items-center gap-2">
            <Bookmark size={18} />
            Saved Articles ({bookmarks.length})
          </h2>
          <div className="bg-gray-50 rounded-lg p-4 border border-gray-200">
            <p className="text-sm text-gray-500">
              You have {bookmarks.length} saved article{bookmarks.length > 1 ? 's' : ''}. 
              Bookmarked articles are stored locally on your device.
            </p>
          </div>
        </div>
      )}
    </div>
  );
}
