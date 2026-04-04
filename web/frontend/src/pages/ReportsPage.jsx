import { useState, useEffect } from 'react';
import { reports } from '../lib/api';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, LineChart, Line } from 'recharts';
import { Download, Loader2, BarChart3, TrendingUp, Table } from 'lucide-react';

export default function ReportsPage() {
  const [activeTab, setActiveTab] = useState('breakdown');
  const [breakdown, setBreakdown] = useState([]);
  const [trends, setTrends] = useState([]);
  const [ytd, setYtd] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    Promise.all([
      reports.breakdown().catch(() => []),
      reports.trends().catch(() => []),
      reports.ytd().catch(() => []),
    ]).then(([b, t, y]) => {
      setBreakdown(b.data || []);
      setTrends(t.data || []);
      setYtd(y.data || []);
      setLoading(false);
    });
  }, []);

  const tabs = [
    { id: 'breakdown', label: 'Breakdown', icon: BarChart3 },
    { id: 'trends', label: 'Trends', icon: TrendingUp },
    { id: 'ytd', label: 'YTD', icon: Table },
  ];

  return (
    <div className="space-y-8">
      <div className="flex flex-col sm:flex-row sm:items-end sm:justify-between gap-5">
        <div>
          <h1 className="text-3xl font-semibold tracking-tight text-neutral-900">Reports</h1>
          <p className="text-neutral-500 mt-1">Contribution analysis and trends</p>
        </div>
        <button className="btn-hover flex items-center gap-2 px-4 py-2.5 border border-neutral-200 rounded-2xl text-sm font-medium hover:bg-neutral-50 transition-all">
          <Download size={15} /> Export
        </button>
      </div>

      <div className="flex gap-1 bg-neutral-100 p-1 rounded-2xl w-fit">
        {tabs.map(tab => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-all ${activeTab === tab.id ? 'bg-white text-neutral-900 shadow-sm' : 'text-neutral-400 hover:text-neutral-600'}`}
          >
            <tab.icon size={14} />
            {tab.label}
          </button>
        ))}
      </div>

      {loading ? (
        <div className="p-16 text-center"><Loader2 size={24} className="animate-spin mx-auto text-neutral-300" /><p className="text-sm text-neutral-400 mt-3">Loading...</p></div>
      ) : (
        <>
          {activeTab === 'breakdown' && (
            <div className="space-y-8">
              <div className="bg-white rounded-2xl border border-[#e8e9eb] p-5">
                <h2 className="text-lg font-semibold tracking-tight text-neutral-900 mb-6">Monthly Breakdown</h2>
                {breakdown.length > 0 ? (
                  <ResponsiveContainer width="100%" height={280}>
                    <BarChart data={breakdown}>
                      <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#f5f5f4" />
                      <XAxis dataKey="period" fontSize={12} tick={{ fill: '#a8a29e' }} axisLine={false} tickLine={false} />
                      <YAxis fontSize={12} tick={{ fill: '#a8a29e' }} axisLine={false} tickLine={false} tickFormatter={(v) => `KES ${(v/1000).toFixed(0)}k`} />
                      <Tooltip formatter={(v) => `KES ${v.toLocaleString()}`} contentStyle={{ borderRadius: '12px', border: '1px solid #e7e5e4', boxShadow: 'none' }} />
                      <Bar dataKey="employee_total" fill="#171717" name="Employee" radius={[4, 4, 0, 0]} />
                      <Bar dataKey="employer_total" fill="#a8a29e" name="Employer" radius={[4, 4, 0, 0]} />
                    </BarChart>
                  </ResponsiveContainer>
                ) : <p className="text-neutral-400 text-center py-12">No data available</p>}
              </div>
              <div className="bg-white rounded-2xl border border-[#e8e9eb] overflow-hidden">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b border-neutral-50">
                      <th className="text-left px-5 py-[18px] font-medium text-neutral-400 text-xs uppercase tracking-wider">Period</th>
                      <th className="text-right px-5 py-[18px] font-medium text-neutral-400 text-xs uppercase tracking-wider">Employees</th>
                      <th className="text-right px-5 py-[18px] font-medium text-neutral-400 text-xs uppercase tracking-wider">Employee Total</th>
                      <th className="text-right px-5 py-[18px] font-medium text-neutral-400 text-xs uppercase tracking-wider">Employer Total</th>
                      <th className="text-right px-5 py-[18px] font-medium text-neutral-400 text-xs uppercase tracking-wider">Grand Total</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-neutral-50">
                    {breakdown.map((b, i) => (
                      <tr key={i} className="hover:bg-neutral-50/50 transition-colors">
                        <td className="px-5 py-[18px] font-medium text-neutral-900">{b.period}</td>
                        <td className="px-5 py-[18px] text-right text-neutral-500">{b.employee_count}</td>
                        <td className="px-5 py-[18px] text-right font-mono text-xs text-neutral-600">KES {(b.employee_total / 100).toLocaleString()}</td>
                        <td className="px-5 py-[18px] text-right font-mono text-xs text-neutral-600">KES {(b.employer_total / 100).toLocaleString()}</td>
                        <td className="px-5 py-[18px] text-right font-mono text-xs font-semibold text-neutral-900">KES {(b.grand_total / 100).toLocaleString()}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {activeTab === 'trends' && (
            <div className="bg-white rounded-2xl border border-[#e8e9eb] p-5">
              <h2 className="text-lg font-semibold tracking-tight text-neutral-900 mb-6">Contribution Trends</h2>
              {trends.length > 0 ? (
                <ResponsiveContainer width="100%" height={280}>
                  <LineChart data={trends}>
                    <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#f5f5f4" />
                    <XAxis dataKey="month" fontSize={12} tick={{ fill: '#a8a29e' }} axisLine={false} tickLine={false} />
                    <YAxis fontSize={12} tick={{ fill: '#a8a29e' }} axisLine={false} tickLine={false} tickFormatter={(v) => `KES ${(v/1000).toFixed(0)}k`} />
                    <Tooltip formatter={(v) => `KES ${v.toLocaleString()}`} contentStyle={{ borderRadius: '12px', border: '1px solid #e7e5e4', boxShadow: 'none' }} />
                    <Line type="monotone" dataKey="total_amount" stroke="#171717" strokeWidth={2} name="Total" dot={false} />
                  </LineChart>
                </ResponsiveContainer>
              ) : <p className="text-neutral-400 text-center py-12">No data available</p>}
            </div>
          )}

          {activeTab === 'ytd' && (
            <div className="bg-white rounded-2xl border border-[#e8e9eb] overflow-hidden">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-neutral-50">
                    <th className="text-left px-5 py-[18px] font-medium text-neutral-400 text-xs uppercase tracking-wider">Member</th>
                    <th className="text-right px-5 py-[18px] font-medium text-neutral-400 text-xs uppercase tracking-wider">Employee YTD</th>
                    <th className="text-right px-5 py-[18px] font-medium text-neutral-400 text-xs uppercase tracking-wider">Employer YTD</th>
                    <th className="text-right px-5 py-[18px] font-medium text-neutral-400 text-xs uppercase tracking-wider">Total YTD</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-neutral-50">
                  {ytd.slice(0, 20).map((y, i) => (
                    <tr key={i} className="hover:bg-neutral-50/50 transition-colors">
                      <td className="px-5 py-[18px] font-medium text-neutral-900">{y.full_name}</td>
                      <td className="px-5 py-[18px] text-right font-mono text-xs text-neutral-600">KES {(y.employee_ytd / 100).toLocaleString()}</td>
                      <td className="px-5 py-[18px] text-right font-mono text-xs text-neutral-600">KES {(y.employer_ytd / 100).toLocaleString()}</td>
                      <td className="px-5 py-[18px] text-right font-mono text-xs font-semibold text-neutral-900">KES {(y.total_ytd / 100).toLocaleString()}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </>
      )}
    </div>
  );
}
