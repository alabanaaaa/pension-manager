import { useState, useEffect } from 'react';
import { claims } from '../lib/api';
import { Link } from 'react-router-dom';
import { Plus, Eye, CheckCircle, XCircle, Clock, Loader2 } from 'lucide-react';

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
    submitted: { label: 'Submitted', icon: Clock },
    accepted: { label: 'Accepted', icon: CheckCircle },
    rejected: { label: 'Rejected', icon: XCircle },
    paid: { label: 'Paid', icon: CheckCircle },
  };

  const filters = ['', 'submitted', 'accepted', 'rejected', 'paid'];
  const filterLabels = { '': 'All', submitted: 'Pending', accepted: 'Accepted', rejected: 'Rejected', paid: 'Paid' };

  return (
    <div className="space-y-8">
      <div className="flex flex-col sm:flex-row sm:items-end sm:justify-between gap-5">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-black">Claims</h1>
          <p className="text-sm text-gray-500 mt-1">{data.length} claims</p>
        </div>
        <Link to="/claims/new" className="btn-primary flex items-center gap-2">
          <Plus size={15} /> New Claim
        </Link>
      </div>

      <div className="flex gap-2 flex-wrap">
        {filters.map(s => (
          <button
            key={s}
            onClick={() => setStatusFilter(s)}
            className={`px-4 py-2 text-sm font-medium transition-all ${statusFilter === s ? 'btn-primary' : 'btn-secondary'}`}
          >
            {filterLabels[s]}
          </button>
        ))}
      </div>

      <div className="card overflow-hidden">
        {loading ? (
          <div className="p-16 text-center"><Loader2 size={24} className="animate-spin mx-auto text-gray-300" /><p className="text-sm text-gray-400 mt-3">Loading...</p></div>
        ) : data.length === 0 ? (
          <div className="p-16 text-center"><p className="text-sm text-gray-400">No claims found</p></div>
        ) : (
          <div className="overflow-x-auto">
            <table className="table">
              <thead>
                <tr>
                  <th className="text-left">Claim No</th>
                  <th className="text-left">Member</th>
                  <th className="text-left hidden sm:table-cell">Type</th>
                  <th className="text-left hidden md:table-cell">Date</th>
                  <th className="text-left">Status</th>
                  <th className="text-right hidden sm:table-cell">Amount</th>
                  <th className="text-right">Actions</th>
                </tr>
              </thead>
              <tbody>
                {data.map(c => {
                  const cfg = statusConfig[c.status] || { label: c.status, cls: 'badge-warning', icon: Clock };
                  return (
                    <tr key={c.id}>
                      <td className="font-mono text-xs text-gray-500">{c.claim_form_no}</td>
                      <td className="font-medium text-black">{c.member_name || c.member_id}</td>
                      <td className="text-gray-500 capitalize hidden sm:table-cell">{c.claim_type}</td>
                      <td className="text-gray-400 text-xs hidden md:table-cell">{new Date(c.date_of_claim).toLocaleDateString()}</td>
                      <td>
                        <span className={`badge ${c.status === 'accepted' ? 'badge-success' : c.status === 'rejected' ? 'badge-error' : c.status === 'paid' ? 'badge-info' : 'badge-warning'}`}>{cfg.label}</span>
                      </td>
                      <td className="text-right font-mono text-xs text-gray-600 hidden sm:table-cell">
                        {c.amount ? `KES ${(c.amount / 100).toLocaleString()}` : '—'}
                      </td>
                      <td className="text-right">
                        <Link to={`/claims/${c.id}`} className="btn-icon">
                          <Eye size={15} className="text-gray-400" />
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
