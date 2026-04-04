import { useState, useEffect } from 'react';
import { news } from '../lib/api';
import { Loader2, ExternalLink, RefreshCw, Calendar, Clock } from 'lucide-react';

export default function NewsPage() {
  const [articles, setArticles] = useState([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [fetchedAt, setFetchedAt] = useState(null);

  const fetchNews = async (refresh = false) => {
    if (refresh) setRefreshing(true);
    else setLoading(true);
    try {
      const res = await news.get();
      const data = res.data || {};
      setArticles(data.articles || []);
      setFetchedAt(data.fetched_at);
    } catch { setArticles([]); }
    finally { setLoading(false); setRefreshing(false); }
  };

  useEffect(() => { fetchNews(); }, []);

  return (
    <div className="space-y-8 animate-fade-in-up">
      <div className="flex items-end justify-between">
        <div>
          <h1 className="text-3xl font-semibold tracking-tight text-neutral-900">Kenya Government News</h1>
          <p className="text-neutral-500 mt-2 text-base">Latest updates relevant to pension management</p>
        </div>
        <button
          onClick={() => fetchNews(true)}
          disabled={refreshing}
          className="btn-hover flex items-center gap-2 px-4 py-2.5 border border-neutral-200 rounded-xl text-sm font-medium hover:bg-neutral-50 disabled:opacity-50 transition-all"
        >
          <RefreshCw size={15} className={refreshing ? 'animate-spin' : ''} />
          Refresh
        </button>
      </div>

      {loading ? (
        <div className="p-16 text-center">
          <Loader2 size={24} className="animate-spin mx-auto text-neutral-300" />
          <p className="text-sm text-neutral-400 mt-3">Loading news...</p>
        </div>
      ) : articles.length === 0 ? (
        <div className="bg-white rounded-2xl border border-neutral-50 p-16 text-center">
          <p className="text-neutral-500">No news articles available</p>
        </div>
      ) : (
        <>
          {fetchedAt && (
            <p className="text-xs text-neutral-400 flex items-center gap-1.5">
              <Clock size={12} /> Last updated: {new Date(fetchedAt).toLocaleString()}
            </p>
          )}
          <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
            {articles.map((article, i) => (
              <div key={i} className="bg-white rounded-2xl border border-neutral-50 overflow-hidden hover:shadow-sm transition-all animate-fade-in" style={{ animationDelay: `${i * 0.05}s` }}>
                {article.url_to_image && (
                  <img src={article.url_to_image} alt="" className="w-full h-40 object-cover" onError={e => e.target.style.display = 'none'} />
                )}
                <div className="p-5">
                  <div className="flex items-center gap-2 text-xs text-neutral-400 mb-3">
                    <Calendar size={12} />
                    <span>{article.published_at ? new Date(article.published_at).toLocaleDateString() : 'Unknown'}</span>
                    {article.source && <span>· {article.source}</span>}
                  </div>
                  <h3 className="font-semibold text-neutral-900 leading-snug mb-2 line-clamp-3">{article.title}</h3>
                  {article.description && <p className="text-sm text-neutral-500 line-clamp-3 mb-4">{article.description}</p>}
                  {article.url && (
                    <a href={article.url} target="_blank" rel="noopener noreferrer" className="inline-flex items-center gap-1.5 text-sm text-blue-600 hover:underline font-medium">
                      Read more <ExternalLink size={13} />
                    </a>
                  )}
                </div>
              </div>
            ))}
          </div>
        </>
      )}
    </div>
  );
}
