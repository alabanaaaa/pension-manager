import { useState, useEffect } from 'react';
import { hospitals } from '../lib/api';
import { Link } from 'react-router-dom';
import { Plus, Hospital, AlertTriangle, Loader2 } from 'lucide-react';

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
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Hospitals</h1>
          <p className="text-gray-500 mt-1">{hospitals.length} hospitals</p>
        </div>
        <Link to="/hospitals/new" className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700">
          <Plus size={16} /> Add Hospital
        </Link>
      </div>

      {/* Alerts */}
      {alerts && alerts.pending_bills > 0 && (
        <div className="bg-red-50 border border-red-200 rounded-xl p-5 flex items-center gap-3">
          <AlertTriangle size={24} className="text-red-600" />
          <div>
            <h3 className="font-semibold text-red-800">{alerts.pending_bills} Pending Bills</h3>
            <p className="text-sm text-red-600">
              {alerts.high_urgency_bills} high urgency ({alerts.high_urgency_bills > 60 ? '>60 days' : '45-60 days'})
              • Total: KES {(alerts.total_pending_amount / 100).toLocaleString()}
            </p>
          </div>
        </div>
      )}

      {loading ? (
        <div className="p-12 text-center"><Loader2 size={32} className="animate-spin mx-auto text-gray-400" /></div>
      ) : hospitals.length === 0 ? (
        <div className="bg-white rounded-xl border p-12 text-center text-gray-500">No hospitals yet</div>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {hospitals.map(h => (
            <Link key={h.id} to={`/hospitals/${h.id}`} className="bg-white rounded-xl border p-5 hover:shadow-md transition-shadow">
              <div className="flex items-start gap-3">
                <div className="p-2 bg-blue-50 rounded-lg"><Hospital size={20} className="text-blue-600" /></div>
                <div className="flex-1 min-w-0">
                  <h3 className="font-semibold truncate">{h.name}</h3>
                  <p className="text-sm text-gray-500 mt-1">{h.address || 'No address'}</p>
                  <p className="text-sm text-gray-500">{h.phone}</p>
                  <div className="mt-3 flex items-center justify-between">
                    <span className={`px-2 py-1 rounded-full text-xs font-medium ${h.status === 'active' ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'}`}>
                      {h.status}
                    </span>
                    <span className="text-sm font-mono">KES {((h.account_balance || 0) / 100).toLocaleString()}</span>
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
