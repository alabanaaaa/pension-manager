import { useState, useEffect } from 'react';
import { portal } from '../lib/api';
import { Loader2, CreditCard, TrendingUp, Calendar, ArrowUpRight, ArrowDownRight, BarChart3 } from 'lucide-react';

export default function PortalContributionsPage() {
  const [contributions, setContributions] = useState([]);
  const [annual, setAnnual] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.all([
      portal.getContributions().then(r => setContributions(Array.isArray(r.data) ? r.data : [])).catch(() => []),
      portal.getAnnualContributions().then(r => setAnnual(Array.isArray(r.data) ? r.data : [])).catch(() => []),
    ]).finally(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <div className="p-16 text-center">
        <Loader2 size={24} className="animate-spin mx-auto text-neutral-300" />
        <p className="text-sm text-neutral-400 mt-3">Loading contributions...</p>
      </div>
    );
  }

  const totalEmployee = contributions.reduce((s, c) => s + (c.employee_amount || 0), 0);
  const totalEmployer = contributions.reduce((s, c) => s + (c.employer_amount || 0), 0);
  const totalAVC = contributions.reduce((s, c) => s + (c.avc_amount || 0), 0);

  return (
    <div className="space-y-8 animate-fade-in-up">
      <div>
        <h1 className="text-2xl font-bold tracking-tight text-black">My Contributions</h1>
        <p className="text-neutral-500 mt-2 text-base">View your contribution history</p>
      </div>

      {/* Summary */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-6">
        <div className="bg-white rounded-2xl border border-neutral-50 p-6">
          <div className="flex items-start justify-between mb-4">
            <div className="p-2.5 rounded-xl bg-blue-50"><ArrowUpRight size={20} className="text-blue-600" /></div>
          </div>
          <p className="text-sm text-neutral-500">Employee Total</p>
          <p className="text-xl font-semibold text-neutral-900 mt-1">KES {(totalEmployee / 100).toLocaleString()}</p>
        </div>
        <div className="bg-white rounded-2xl border border-neutral-50 p-6">
          <div className="flex items-start justify-between mb-4">
            <div className="p-2.5 rounded-xl bg-emerald-50"><ArrowDownRight size={20} className="text-emerald-600" /></div>
          </div>
          <p className="text-sm text-neutral-500">Employer Total</p>
          <p className="text-xl font-semibold text-neutral-900 mt-1">KES {(totalEmployer / 100).toLocaleString()}</p>
        </div>
        <div className="bg-white rounded-2xl border border-neutral-50 p-6">
          <div className="flex items-start justify-between mb-4">
            <div className="p-2.5 rounded-xl bg-violet-50"><BarChart3 size={20} className="text-violet-600" /></div>
          </div>
          <p className="text-sm text-neutral-500">AVC Total</p>
          <p className="text-xl font-semibold text-neutral-900 mt-1">KES {(totalAVC / 100).toLocaleString()}</p>
        </div>
      </div>

      {/* Contributions Table */}
      <div className="bg-white rounded-2xl border border-neutral-50 overflow-hidden">
        <div className="flex items-center justify-between px-6 py-5 border-b border-neutral-50">
          <h2 className="text-lg font-semibold tracking-tight text-neutral-900">Contribution History</h2>
        </div>
        {contributions.length === 0 ? (
          <div className="p-16 text-center"><p className="text-neutral-500">No contributions found</p></div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-neutral-50">
                  <th className="text-left px-6 py-4 font-medium text-neutral-400 text-xs uppercase tracking-wider">Period</th>
                  <th className="text-right px-6 py-4 font-medium text-neutral-400 text-xs uppercase tracking-wider">Employee</th>
                  <th className="text-right px-6 py-4 font-medium text-neutral-400 text-xs uppercase tracking-wider">Employer</th>
                  <th className="text-right px-6 py-4 font-medium text-neutral-400 text-xs uppercase tracking-wider">AVC</th>
                  <th className="text-right px-6 py-4 font-medium text-neutral-400 text-xs uppercase tracking-wider">Total</th>
                  <th className="text-left px-6 py-4 font-medium text-neutral-400 text-xs uppercase tracking-wider">Status</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-neutral-50">
                {contributions.slice(0, 50).map((c, i) => (
                  <tr key={c.id} className="hover:bg-neutral-50/50 transition-colors">
                    <td className="px-6 py-4 text-neutral-500">{new Date(c.period).toLocaleDateString()}</td>
                    <td className="px-6 py-4 text-right font-mono text-xs text-neutral-600">KES {((c.employee_amount || 0) / 100).toLocaleString()}</td>
                    <td className="px-6 py-4 text-right font-mono text-xs text-neutral-600">KES {((c.employer_amount || 0) / 100).toLocaleString()}</td>
                    <td className="px-6 py-4 text-right font-mono text-xs text-neutral-600">KES {((c.avc_amount || 0) / 100).toLocaleString()}</td>
                    <td className="px-6 py-4 text-right font-mono text-xs font-semibold text-neutral-900">KES {((c.total_amount || 0) / 100).toLocaleString()}</td>
                    <td className="px-6 py-4">
                      <span className={`px-2.5 py-1 rounded-full text-xs font-medium ${c.status === 'confirmed' ? 'bg-emerald-50 text-emerald-700' : 'bg-amber-50 text-amber-700'}`}>{c.status}</span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
