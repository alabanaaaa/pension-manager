import { useState, useEffect } from 'react';
import { sponsor } from '../lib/api';
import { Link } from 'react-router-dom';
import { Plus, Building2, Loader2, TrendingUp, Users, CreditCard } from 'lucide-react';

export default function SponsorsPage() {
  const [sponsors, setSponsors] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    sponsor.list()
      .then(res => setSponsors(Array.isArray(res.data) ? res.data : []))
      .catch(() => setSponsors([]))
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="space-y-8">
      <div className="flex flex-col sm:flex-row sm:items-end sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-black">Sponsors</h1>
          <p className="text-sm text-gray-500 mt-1">{sponsors.length} sponsors</p>
        </div>
        <Link to="/sponsors/new" className="btn-primary flex items-center gap-2">
          <Plus size={15} /> Add Sponsor
        </Link>
      </div>

      {loading ? (
        <div className="p-16 text-center"><Loader2 size={24} className="animate-spin mx-auto text-gray-300" /><p className="text-sm text-gray-400 mt-3">Loading...</p></div>
      ) : sponsors.length === 0 ? (
        <div className="card p-16 text-center">
          <Building2 size={48} className="mx-auto text-gray-200 mb-4" />
          <p className="text-gray-500">No sponsors yet</p>
        </div>
      ) : (
        <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {sponsors.map((s, i) => (
            <div key={s.id} className="card p-6 hover:border-black transition-all" style={{ animationDelay: `${i * 0.05}s` }}>
              <div className="flex items-start gap-4">
                <div className="p-2.5 border border-gray-200 rounded flex-shrink-0">
                  <Building2 size={20} className="text-black" />
                </div>
                <div className="flex-1 min-w-0">
                  <h3 className="font-semibold text-black truncate">{s.name}</h3>
                  {s.contact_person && <p className="text-sm text-gray-400 mt-1">{s.contact_person}</p>}
                  {s.phone && <p className="text-sm text-gray-400 mt-0.5">{s.phone}</p>}
                  <div className="flex items-center gap-4 mt-4 text-sm text-gray-500">
                    <span className="flex items-center gap-1.5"><Users size={13} /> {s.total_members || 0} members</span>
                    <span className={`badge ${s.status === 'active' ? 'badge-success' : 'badge-warning'}`}>{s.status}</span>
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
