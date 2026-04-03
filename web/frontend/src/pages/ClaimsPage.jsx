import { useState, useEffect } from 'react';
import { claims } from '../lib/api';
import { Link } from 'react-router-dom';
import { Plus, Search, Eye, CheckCircle, XCircle, Clock, Loader2, ChevronRight } from 'lucide-react';

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
    submitted: { label: 'Submitted', cls: 'bg-amber-50 text-amber-700', icon: Clock },
    accepted: { label: 'Accepted', cls: 'bg-emerald-50 text-emerald-700', icon: CheckCircle },
    rejected: { label: 'Rejected', cls: 'bg-red-50 text-red-700', icon: XCircle },
    paid: { label: 'Paid', cls: 'bg-blue-50 text-blue-700', icon: CheckCircle },
  };

  const filters = ['', 'submitted', 'accepted', 'rejected', 'paid'];
  const filterLabels = { '': 'All', submitted: 'Pending', accepted: 'Accepted', rejected: 'Rejected', paid: 'Paid' };

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-end sm:justify-between gap-4">
        <div>
          <h1 className="text-3xl font-semibold tracking-tight text-neutral-900">Claims</h1>
          <p className="text-neutral-500 mt-1">{data.length} claims</p>
        </div>
        <Link to="/claims/new" className="btn-hover flex items-center gap-2 px-4 py-2.5 bg-neutral-900 text-white rounded-xl text-sm font-medium hover:bg-neutral-800 transition-all">
          <Plus size={15} /> New Claim
        </Link>
      </div>

      {/* Filters */}
      <div className="flex gap-1.5 flex-wrap">
        {filters.map(s => (
          <button
            key={s}
            onClick={() => setStatusFilter(s)}
            className={`px-4 py-2 rounded-xl text-sm font-medium transition-all ${statusFilter === s ? 'bg-neutral-900 text-white' : 'bg-white border border-neutral-200 text-neutral-500 hover:bg-neutral-50'}`}
          >
            {filterLabels[s]}
          </button>
        ))}
      </div>

      {/* Table */}
      <div className="bg-white rounded-2xl border border-neutral-100 overflow-hidden">
        {loading ? (
          <div className="p-16 text-center"><Loader2 size={24} className="animate-spin mx-auto text-neutral-300" /><p className="text-sm text-neutral-400 mt-3">Loading...</p></div>
        ) : data.length === 0 ? (
          <div className="p-16 text-center"><p className="text-sm text-neutral-400">No claims found</p></div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-neutral-50">
                  <th className="text-left px-6 py-3.5 font-medium text-neutral-400 text-xs uppercase tracking-wider">Claim No</th>
                  <th className="text-left px-6 py-3.5 font-medium text-neutral-400 text-xs uppercase tracking-wider">Member</th>
                  <th className="text-left px-6 py-3.5 font-medium text-neutral-400 text-xs uppercase tracking-wider hidden sm:table-cell">Type</th>
                  <th className="text-left px-6 py-3.5 font-medium text-neutral-400 text-xs uppercase tracking-wider hidden md:table-cell">Date</th>
                  <th className="text-left px-6 py-3.5 font-medium text-neutral-400 text-xs uppercase tracking-wider">Status</th>
                  <th className="text-right px-6 py-3.5 font-medium text-neutral-400 text-xs uppercase tracking-wider hidden sm:table-cell">Amount</th>
                  <th className="text-right px-6 py-3.5 font-medium text-neutral-400 text-xs uppercase tracking-wider">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-neutral-50">
                {data.map(c => {
                  const cfg = statusConfig[c.status] || { label: c.status, cls: 'bg-neutral-50 text-neutral-600', icon: Clock };
                  return (
                    <tr key={c.id} className="hover:bg-neutral-50/50 transition-colors">
                      <td className="px-6 py-4 font-mono text-xs text-neutral-500">{c.claim_form_no}</td>
                      <td className="px-6 py-4 font-medium text-neutral-900">{c.member_name || c.member_id}</td>
                      <td className="px-6 py-4 text-neutral-500 capitalize hidden sm:table-cell">{c.claim_type}</td>
                      <td className="px-6 py-4 text-neutral-400 text-xs hidden md:table-cell">{new Date(c.date_of_claim).toLocaleDateString()}</td>
                      <td className="px-6 py-4">
                        <span className={`px-2.5 py-1 rounded-full text-xs font-medium ${cfg.cls}`}>{cfg.label}</span>
                      </td>
                      <td className="px-6 py-4 text-right font-mono text-xs text-neutral-600 hidden sm:table-cell">
                        {c.amount ? `KES ${(c.amount / 100).toLocaleString()}` : '—'}
                      </td>
                      <td className="px-6 py-4 text-right">
                        <Link to={`/claims/${c.id}`} className="p-2 hover:bg-neutral-100 rounded-lg transition-colors inline-flex">
                          <Eye size={15} className="text-neutral-400" />
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
