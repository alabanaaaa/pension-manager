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
    <div className="space-y-8 animate-fade-in-up">
      <div className="flex flex-col sm:flex-row sm:items-end sm:justify-between gap-4">
        <div>
          <h1 className="text-3xl font-semibold tracking-tight text-neutral-900">Sponsors</h1>
          <p className="text-neutral-500 mt-2 text-base">{sponsors.length} sponsors</p>
        </div>
        <Link to="/sponsors/new" className="btn-hover flex items-center gap-2 px-4 py-2.5 bg-neutral-900 text-white rounded-xl text-sm font-medium hover:bg-neutral-800 transition-all">
          <Plus size={15} /> Add Sponsor
        </Link>
      </div>

      {loading ? (
        <div className="p-16 text-center"><Loader2 size={24} className="animate-spin mx-auto text-neutral-300" /><p className="text-sm text-neutral-400 mt-3">Loading...</p></div>
      ) : sponsors.length === 0 ? (
        <div className="bg-white rounded-2xl border border-neutral-50 p-16 text-center">
          <Building2 size={48} className="mx-auto text-neutral-200 mb-4" />
          <p className="text-neutral-500">No sponsors yet</p>
        </div>
      ) : (
        <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {sponsors.map((s, i) => (
            <div key={s.id} className="bg-white rounded-2xl border border-neutral-50 p-6 hover:shadow-sm transition-all animate-fade-in" style={{ animationDelay: `${i * 0.05}s` }}>
              <div className="flex items-start gap-4">
                <div className="p-2.5 rounded-xl bg-neutral-50 flex-shrink-0">
                  <Building2 size={20} className="text-neutral-600" />
                </div>
                <div className="flex-1 min-w-0">
                  <h3 className="font-semibold text-neutral-900 truncate">{s.name}</h3>
                  {s.contact_person && <p className="text-sm text-neutral-400 mt-1">{s.contact_person}</p>}
                  {s.phone && <p className="text-sm text-neutral-400 mt-0.5">{s.phone}</p>}
                  <div className="flex items-center gap-4 mt-4 text-sm text-neutral-500">
                    <span className="flex items-center gap-1.5"><Users size={13} /> {s.total_members || 0} members</span>
                    <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${s.status === 'active' ? 'bg-emerald-50 text-emerald-700' : 'bg-neutral-100 text-neutral-600'}`}>{s.status}</span>
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
