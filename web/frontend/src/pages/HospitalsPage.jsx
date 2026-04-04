import { useState, useEffect } from 'react';
import { hospitals } from '../lib/api';
import { Link } from 'react-router-dom';
import { Plus, Hospital, AlertTriangle, Loader2, ChevronRight, Phone, Mail } from 'lucide-react';

export default function HospitalsPage() {
  const [hospitals, setHospitals] = useState([]);
  const [alerts, setAlerts] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.all([
      hospitals.list().catch(() => []),
      hospitals.getAlerts().catch(() => null),
    ]).then(([h, a]) => {
      setHospitals(h.data || []);
      setAlerts(a?.data);
      setLoading(false);
    });
  }, []);

  return (
    <div className="space-y-8">
      <div className="flex flex-col sm:flex-row sm:items-end sm:justify-between gap-5">
        <div>
          <h1 className="text-3xl font-semibold tracking-tight text-neutral-900">Hospitals</h1>
          <p className="text-neutral-500 mt-1">{hospitals.length} hospitals</p>
        </div>
        <Link to="/hospitals/new" className="btn-hover flex items-center gap-2 px-4 py-2.5 bg-neutral-900 text-white rounded-2xl text-sm font-medium hover:bg-neutral-800 transition-all">
          <Plus size={15} /> Add Hospital
        </Link>
      </div>

      {/* Alerts */}
      {alerts && alerts.pending_bills > 0 && (
        <div className="bg-amber-50 border border-amber-100 rounded-2xl p-5">
          <div className="flex items-center gap-5">
            <div className="p-3 bg-amber-100 rounded-2xl">
              <AlertTriangle size={20} className="text-amber-600" />
            </div>
            <div className="flex-1">
              <h3 className="font-medium text-amber-900">{alerts.pending_bills} Pending Bills</h3>
              <p className="text-sm text-amber-600 mt-0.5">
                {alerts.high_urgency_bills} high urgency · Total: KES {(alerts.total_pending_amount / 100).toLocaleString()}
              </p>
            </div>
            <Link to="/medical-expenditures" className="btn-hover px-5 py-2.5 bg-amber-600 text-white rounded-2xl text-sm font-medium hover:bg-amber-700 transition-all">
              View Bills
            </Link>
          </div>
        </div>
      )}

      {loading ? (
        <div className="p-16 text-center"><Loader2 size={24} className="animate-spin mx-auto text-neutral-300" /><p className="text-sm text-neutral-400 mt-3">Loading...</p></div>
      ) : hospitals.length === 0 ? (
        <div className="bg-white rounded-2xl border border-[#e8e9eb] p-16 text-center">
          <p className="text-sm text-neutral-400 mb-4">No hospitals yet</p>
          <Link to="/hospitals/new" className="text-sm text-neutral-900 font-medium hover:underline">Add one to get started</Link>
        </div>
      ) : (
        <div className="grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
          {hospitals.map(h => (
            <Link key={h.id} to={`/hospitals/${h.id}`} className="btn-hover bg-white rounded-2xl border border-[#e8e9eb] p-5 hover:shadow-sm transition-all">
              <div className="flex items-start gap-5">
                <div className="p-2.5 bg-neutral-50 rounded-2xl flex-shrink-0">
                  <Hospital size={20} className="text-neutral-600" />
                </div>
                <div className="flex-1 min-w-0">
                  <h3 className="font-semibold text-neutral-900 truncate">{h.name}</h3>
                  {h.address && <p className="text-sm text-neutral-400 mt-1">{h.address}</p>}
                  <div className="flex items-center gap-5 mt-3 text-xs text-neutral-400">
                    {h.phone && <span className="flex items-center gap-1"><Phone size={12} /> {h.phone}</span>}
                  </div>
                  <div className="mt-4 flex items-center justify-between">
                    <span className={`px-2.5 py-1 rounded-full text-xs font-medium ${h.status === 'active' ? 'bg-emerald-50 text-emerald-700' : 'bg-neutral-100 text-neutral-600'}`}>
                      {h.status}
                    </span>
                    <span className="text-sm font-mono text-neutral-600">KES {((h.account_balance || 0) / 100).toLocaleString()}</span>
                  </div>
                </div>
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
