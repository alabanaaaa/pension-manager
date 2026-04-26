import { useState } from 'react';
import { portal } from '../lib/api';
import { Loader2, TrendingUp, CheckCircle, AlertCircle, Calculator, DollarSign, Calendar } from 'lucide-react';

export default function PortalProjectionsPage() {
  const [params, setParams] = useState({
    retirement_age: 60,
    salary_growth_rate: 5,
    investment_return: 8,
    scheme_type: 'dc',
  });
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState(null);

  const handleProject = async () => {
    setLoading(true);
    setResult(null);
    try {
      const res = await portal.projectBenefits(params);
      setResult(res.data || null);
    } catch (err) {
      setResult({ error: err.response?.data?.error || 'Projection failed' });
    }
    finally { setLoading(false); }
  };

  return (
    <div className="space-y-8 animate-fade-in-up">
      <div>
        <h1 className="text-2xl font-bold tracking-tight text-black">Benefit Projections</h1>
        <p className="text-neutral-500 mt-2 text-base">Project your retirement benefits based on current savings</p>
      </div>

      {/* Parameters */}
      <div className="bg-white rounded-2xl border border-neutral-50 overflow-hidden">
        <div className="flex items-center justify-between px-6 py-5 border-b border-neutral-50">
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-blue-50"><Calculator size={20} className="text-blue-600" /></div>
            <h2 className="text-lg font-semibold tracking-tight text-neutral-900">Projection Parameters</h2>
          </div>
        </div>
        <div className="p-6 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
          <div>
            <label className="block text-xs text-neutral-400 uppercase tracking-wider mb-2">Retirement Age</label>
            <input
              type="number"
              value={params.retirement_age}
              onChange={e => setParams(p => ({ ...p, retirement_age: parseInt(e.target.value) }))}
              className="w-full px-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all"
            />
          </div>
          <div>
            <label className="block text-xs text-neutral-400 uppercase tracking-wider mb-2">Salary Growth (%)</label>
            <input
              type="number"
              value={params.salary_growth_rate}
              onChange={e => setParams(p => ({ ...p, salary_growth_rate: parseFloat(e.target.value) }))}
              className="w-full px-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all"
            />
          </div>
          <div>
            <label className="block text-xs text-neutral-400 uppercase tracking-wider mb-2">Investment Return (%)</label>
            <input
              type="number"
              value={params.investment_return}
              onChange={e => setParams(p => ({ ...p, investment_return: parseFloat(e.target.value) }))}
              className="w-full px-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all"
            />
          </div>
          <div>
            <label className="block text-xs text-neutral-400 uppercase tracking-wider mb-2">Scheme Type</label>
            <select
              value={params.scheme_type}
              onChange={e => setParams(p => ({ ...p, scheme_type: e.target.value }))}
              className="w-full px-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all"
            >
              <option value="dc">Defined Contribution (DC)</option>
              <option value="db">Defined Benefit (DB)</option>
            </select>
          </div>
        </div>
        <div className="px-6 pb-6">
          <button
            onClick={handleProject}
            disabled={loading}
            className="btn-hover flex items-center gap-2 px-6 py-3.5 bg-neutral-900 text-white rounded-xl text-sm font-medium hover:bg-neutral-800 disabled:opacity-50 transition-all"
          >
            {loading ? <Loader2 size={16} className="animate-spin" /> : <TrendingUp size={16} />}
            Calculate Projection
          </button>
        </div>
      </div>

      {/* Results */}
      {result && !result.error && (
        <div className="space-y-6 animate-fade-in-up">
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
            <div className="bg-white rounded-2xl border border-neutral-50 p-6">
              <div className="flex items-center gap-3 mb-3">
                <div className="p-2.5 rounded-xl bg-emerald-50"><DollarSign size={20} className="text-emerald-600" /></div>
              </div>
              <p className="text-sm text-neutral-500">Projected Balance</p>
              <p className="text-xl font-semibold text-neutral-900 mt-1">KES {((result.projected_balance || 0) / 100).toLocaleString()}</p>
            </div>
            <div className="bg-white rounded-2xl border border-neutral-50 p-6">
              <div className="flex items-center gap-3 mb-3">
                <div className="p-2.5 rounded-xl bg-blue-50"><TrendingUp size={20} className="text-blue-600" /></div>
              </div>
              <p className="text-sm text-neutral-500">Total Contributions</p>
              <p className="text-xl font-semibold text-neutral-900 mt-1">KES {((result.total_contributions || 0) / 100).toLocaleString()}</p>
            </div>
            <div className="bg-white rounded-2xl border border-neutral-50 p-6">
              <div className="flex items-center gap-3 mb-3">
                <div className="p-2.5 rounded-xl bg-violet-50"><Calendar size={20} className="text-violet-600" /></div>
              </div>
              <p className="text-sm text-neutral-500">Monthly Pension</p>
              <p className="text-xl font-semibold text-neutral-900 mt-1">KES {((result.estimated_monthly || 0) / 100).toLocaleString()}</p>
            </div>
            <div className="bg-white rounded-2xl border border-neutral-50 p-6">
              <div className="flex items-center gap-3 mb-3">
                <div className="p-2.5 rounded-xl bg-amber-50"><DollarSign size={20} className="text-amber-600" /></div>
              </div>
              <p className="text-sm text-neutral-500">Lump Sum</p>
              <p className="text-xl font-semibold text-neutral-900 mt-1">KES {((result.estimated_lump_sum || 0) / 100).toLocaleString()}</p>
            </div>
          </div>

          {/* Year by Year */}
          {result.year_by_year && result.year_by_year.length > 0 && (
            <div className="bg-white rounded-2xl border border-neutral-50 overflow-hidden">
              <div className="flex items-center justify-between px-6 py-5 border-b border-neutral-50">
                <h2 className="text-lg font-semibold tracking-tight text-neutral-900">Year-by-Year Projection</h2>
              </div>
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b border-neutral-50">
                      <th className="text-left px-6 py-4 font-medium text-neutral-400 text-xs uppercase tracking-wider">Year</th>
                      <th className="text-left px-6 py-4 font-medium text-neutral-400 text-xs uppercase tracking-wider">Age</th>
                      <th className="text-right px-6 py-4 font-medium text-neutral-400 text-xs uppercase tracking-wider">Salary</th>
                      <th className="text-right px-6 py-4 font-medium text-neutral-400 text-xs uppercase tracking-wider">Member Contrib</th>
                      <th className="text-right px-6 py-4 font-medium text-neutral-400 text-xs uppercase tracking-wider">Sponsor Contrib</th>
                      <th className="text-right px-6 py-4 font-medium text-neutral-400 text-xs uppercase tracking-wider">Interest</th>
                      <th className="text-right px-6 py-4 font-medium text-neutral-400 text-xs uppercase tracking-wider">End Balance</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-neutral-50">
                    {result.year_by_year.slice(0, 10).map((y, i) => (
                      <tr key={i} className="hover:bg-neutral-50/50 transition-colors">
                        <td className="px-6 py-4 text-neutral-500">{y.year}</td>
                        <td className="px-6 py-4 text-neutral-500">{y.age}</td>
                        <td className="px-6 py-4 text-right font-mono text-xs text-neutral-600">KES {(y.salary / 100).toLocaleString()}</td>
                        <td className="px-6 py-4 text-right font-mono text-xs text-neutral-600">KES {(y.member_contribution / 100).toLocaleString()}</td>
                        <td className="px-6 py-4 text-right font-mono text-xs text-neutral-600">KES {(y.sponsor_contribution / 100).toLocaleString()}</td>
                        <td className="px-6 py-4 text-right font-mono text-xs text-emerald-600">KES {(y.interest / 100).toLocaleString()}</td>
                        <td className="px-6 py-4 text-right font-mono text-xs font-semibold text-neutral-900">KES {(y.end_balance / 100).toLocaleString()}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </div>
      )}

      {result?.error && (
        <div className="bg-red-50 border border-red-100 rounded-2xl p-5 flex items-center gap-3">
          <AlertCircle size={18} className="text-red-600" />
          <p className="text-sm text-red-700">{result.error}</p>
        </div>
      )}
    </div>
  );
}
