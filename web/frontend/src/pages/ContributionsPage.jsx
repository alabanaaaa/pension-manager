import { useState, useEffect } from 'react';
import { contributions } from '../lib/api';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, LineChart, Line } from 'recharts';
import { CreditCard, TrendingUp, Loader2, Plus, Download } from 'lucide-react';
import { Link } from 'react-router-dom';

export default function ContributionsPage() {
  const [data, setData] = useState([]);
  const [loading, setLoading] = useState(true);
  const [chartData, setChartData] = useState([]);

  useEffect(() => {
    setLoading(true);
    contributions.list()
      .then(res => {
        const items = res.data || [];
        setData(items);
        const grouped = {};
        items.slice(0, 50).forEach(c => {
          const month = new Date(c.period).toLocaleDateString('en', { month: 'short', year: '2-digit' });
          if (!grouped[month]) grouped[month] = { month, employee: 0, employer: 0, total: 0 };
          grouped[month].employee += (c.employee_amount || 0) / 100;
          grouped[month].employer += (c.employer_amount || 0) / 100;
          grouped[month].total += (c.total_amount || 0) / 100;
        });
        setChartData(Object.values(grouped).reverse());
      })
      .catch(() => setData([]))
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="space-y-8">
      <div className="flex flex-col sm:flex-row sm:items-end sm:justify-between gap-5">
        <div>
          <h1 className="text-3xl font-semibold tracking-tight text-neutral-900">Contributions</h1>
          <p className="text-neutral-500 mt-1">{data.length} records</p>
        </div>
        <div className="flex gap-2">
          <button className="btn-hover flex items-center gap-2 px-4 py-2.5 border border-neutral-200 rounded-xl text-sm font-medium hover:bg-neutral-50 transition-all">
            <Download size={15} /> Export
          </button>
          <Link to="/contributions/new" className="btn-hover flex items-center gap-2 px-4 py-2.5 bg-neutral-900 text-white rounded-xl text-sm font-medium hover:bg-neutral-800 transition-all">
            <Plus size={15} /> Record
          </Link>
        </div>
      </div>

      {/* Chart */}
      {chartData.length > 0 && (
        <div className="bg-white rounded-xl border border-[#e8e9eb] p-5">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-lg font-semibold tracking-tight text-neutral-900">Monthly Trends</h2>
          </div>
          <ResponsiveContainer width="100%" height={280}>
            <BarChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#f5f5f4" />
              <XAxis dataKey="month" fontSize={12} tick={{ fill: '#a8a29e' }} axisLine={false} tickLine={false} />
              <YAxis fontSize={12} tick={{ fill: '#a8a29e' }} axisLine={false} tickLine={false} tickFormatter={(v) => `KES ${(v/1000).toFixed(0)}k`} />
              <Tooltip formatter={(v) => `KES ${v.toLocaleString()}`} contentStyle={{ borderRadius: '12px', border: '1px solid #e7e5e4', boxShadow: 'none' }} />
              <Bar dataKey="employee" fill="#171717" name="Employee" radius={[4, 4, 0, 0]} />
              <Bar dataKey="employer" fill="#a8a29e" name="Employer" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>
      )}

      {/* Table */}
      <div className="bg-white rounded-xl border border-[#e8e9eb] overflow-hidden">
        {loading ? (
          <div className="p-16 text-center"><Loader2 size={24} className="animate-spin mx-auto text-neutral-300" /><p className="text-sm text-neutral-400 mt-3">Loading...</p></div>
        ) : data.length === 0 ? (
          <div className="p-16 text-center"><p className="text-sm text-neutral-400">No contributions found</p></div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-neutral-50">
                  <th className="text-left px-5 py-[18px] font-medium text-neutral-400 text-xs uppercase tracking-wider">Period</th>
                  <th className="text-left px-5 py-[18px] font-medium text-neutral-400 text-xs uppercase tracking-wider">Member</th>
                  <th className="text-right px-5 py-[18px] font-medium text-neutral-400 text-xs uppercase tracking-wider">Employee</th>
                  <th className="text-right px-5 py-[18px] font-medium text-neutral-400 text-xs uppercase tracking-wider">Employer</th>
                  <th className="text-right px-5 py-[18px] font-medium text-neutral-400 text-xs uppercase tracking-wider hidden sm:table-cell">AVC</th>
                  <th className="text-right px-5 py-[18px] font-medium text-neutral-400 text-xs uppercase tracking-wider">Total</th>
                  <th className="text-left px-5 py-[18px] font-medium text-neutral-400 text-xs uppercase tracking-wider">Status</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-neutral-50">
                {data.slice(0, 20).map(c => (
                  <tr key={c.id} className="hover:bg-neutral-50/50 transition-colors">
                    <td className="px-5 py-[18px] text-neutral-500">{new Date(c.period).toLocaleDateString()}</td>
                    <td className="px-5 py-[18px] font-medium text-neutral-900">{c.member_name || c.member_id}</td>
                    <td className="px-5 py-[18px] text-right font-mono text-xs text-neutral-600">KES {((c.employee_amount || 0) / 100).toLocaleString()}</td>
                    <td className="px-5 py-[18px] text-right font-mono text-xs text-neutral-600">KES {((c.employer_amount || 0) / 100).toLocaleString()}</td>
                    <td className="px-5 py-[18px] text-right font-mono text-xs text-neutral-600 hidden sm:table-cell">KES {((c.avc_amount || 0) / 100).toLocaleString()}</td>
                    <td className="px-5 py-[18px] text-right font-mono text-xs font-semibold text-neutral-900">KES {((c.total_amount || 0) / 100).toLocaleString()}</td>
                    <td className="px-5 py-[18px]">
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
