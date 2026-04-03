import { useState, useEffect } from 'react';
import { claims } from '../lib/api';
import { Link } from 'react-router-dom';
import { Plus, Search, Filter, Eye, CheckCircle, XCircle, Clock, Loader2 } from 'lucide-react';

export default function ClaimsPage() {
  const [data, setData] = useState([]);
  const [loading, setLoading] = useState(true);
  const [statusFilter, setStatusFilter] = useState('');

  useEffect(() => {
    setLoading(true);
    claims.list({ status: statusFilter })
      .then(res => setData(res.data || []))
      .catch(() => setData([]))
      .finally(() => setLoading(false));
  }, [statusFilter]);

  const statusConfig = {
    submitted: { color: 'bg-yellow-100 text-yellow-800', icon: Clock },
    accepted: { color: 'bg-green-100 text-green-800', icon: CheckCircle },
    rejected: { color: 'bg-red-100 text-red-800', icon: XCircle },
    paid: { color: 'bg-blue-100 text-blue-800', icon: CheckCircle },
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Claims</h1>
          <p className="text-gray-500 mt-1">{data.length} claims</p>
        </div>
        <Link
          to="/claims/new"
          className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700"
        >
          <Plus size={16} />
          New Claim
        </Link>
      </div>

      <div className="flex gap-2">
        {['', 'submitted', 'accepted', 'rejected', 'paid'].map(s => (
          <button
            key={s}
            onClick={() => setStatusFilter(s)}
            className={`px-3 py-1.5 rounded-lg text-sm font-medium ${statusFilter === s ? 'bg-blue-600 text-white' : 'bg-white border hover:bg-gray-50'}`}
          >
            {s || 'All'}
          </button>
        ))}
      </div>

      <div className="bg-white rounded-xl border overflow-hidden">
        {loading ? (
          <div className="p-12 text-center"><Loader2 size={32} className="animate-spin mx-auto text-gray-400" /></div>
        ) : data.length === 0 ? (
          <div className="p-12 text-center text-gray-500">No claims found</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="bg-gray-50 border-b">
                <tr>
                  <th className="text-left px-4 py-3 font-medium text-gray-500">Claim No</th>
                  <th className="text-left px-4 py-3 font-medium text-gray-500">Member</th>
                  <th className="text-left px-4 py-3 font-medium text-gray-500">Type</th>
                  <th className="text-left px-4 py-3 font-medium text-gray-500">Date</th>
                  <th className="text-left px-4 py-3 font-medium text-gray-500">Status</th>
                  <th className="text-right px-4 py-3 font-medium text-gray-500">Amount</th>
                  <th className="text-right px-4 py-3 font-medium text-gray-500">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y">
                {data.map(c => {
                  const cfg = statusConfig[c.status] || { color: 'bg-gray-100 text-gray-800', icon: Clock };
                  return (
                    <tr key={c.id} className="hover:bg-gray-50">
                      <td className="px-4 py-3 font-mono text-xs">{c.claim_form_no}</td>
                      <td className="px-4 py-3">{c.member_name || c.member_id}</td>
                      <td className="px-4 py-3 capitalize">{c.claim_type}</td>
                      <td className="px-4 py-3 text-gray-500">{new Date(c.date_of_claim).toLocaleDateString()}</td>
                      <td className="px-4 py-3">
                        <span className={`px-2 py-1 rounded-full text-xs font-medium ${cfg.color}`}>{c.status}</span>
                      </td>
                      <td className="px-4 py-3 text-right font-mono text-xs">
                        {c.amount ? `KES ${(c.amount / 100).toLocaleString()}` : '—'}
                      </td>
                      <td className="px-4 py-3 text-right">
                        <Link to={`/claims/${c.id}`} className="p-1.5 hover:bg-gray-100 rounded">
                          <Eye size={16} className="text-gray-500" />
                        </Link>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
