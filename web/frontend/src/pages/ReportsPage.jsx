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
    <div className="space-y-6">
      <div className="flex items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-black">Reports</h1>
          <p className="text-sm text-gray-500 mt-1">Contribution analysis and trends</p>
        </div>
        <button className="btn btn-secondary flex items-center gap-2">
          <Download size={14} /> Export
        </button>
      </div>

      {/* Tabs */}
      <div className="tabs">
        {tabs.map(tab => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`tab ${activeTab === tab.id ? 'active' : ''}`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {loading ? (
        <div className="p-16 text-center">
          <Loader2 size={24} className="animate-spin mx-auto text-gray-300" />
          <p className="text-sm text-gray-400 mt-3">Loading...</p>
        </div>
      ) : (
        <>
          {activeTab === 'breakdown' && (
            <div className="space-y-6">
              <div className="card">
                <div className="card-header">
                  <h2 className="text-base font-semibold text-black">Monthly Breakdown</h2>
                </div>
                <div className="card-body">
                  {breakdown.length > 0 ? (
                    <ResponsiveContainer width="100%" height={280}>
                      <BarChart data={breakdown}>
                        <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#e5e5e5" />
                        <XAxis dataKey="period" fontSize={11} tick={{ fill: '#666' }} axisLine={false} tickLine={false} />
                        <YAxis fontSize={11} tick={{ fill: '#666' }} axisLine={false} tickLine={false} tickFormatter={(v) => `KES ${(v/1000).toFixed(0)}k`} />
                        <Tooltip formatter={(v) => `KES ${v.toLocaleString()}`} contentStyle={{ borderRadius: '4px', border: '1px solid #e5e5e5' }} />
                        <Bar dataKey="employee_total" fill="#000" name="Employee" radius={[2, 2, 0, 0]} />
                        <Bar dataKey="employer_total" fill="#999" name="Employer" radius={[2, 2, 0, 0]} />
                      </BarChart>
                    </ResponsiveContainer>
                  ) : <p className="text-gray-400 text-center py-12">No data available</p>}
                </div>
              </div>
              <div className="card">
                <table className="table">
                  <thead>
                    <tr>
                      <th className="text-left">Period</th>
                      <th className="text-right">Employees</th>
                      <th className="text-right">Employee Total</th>
                      <th className="text-right">Employer Total</th>
                      <th className="text-right">Grand Total</th>
                    </tr>
                  </thead>
                  <tbody>
                    {breakdown.map((b, i) => (
                      <tr key={i}>
                        <td className="font-medium text-black">{b.period}</td>
                        <td className="text-right text-gray-500">{b.employee_count}</td>
                        <td className="text-right font-mono text-sm">KES {(b.employee_total / 100).toLocaleString()}</td>
                        <td className="text-right font-mono text-sm">KES {(b.employer_total / 100).toLocaleString()}</td>
                        <td className="text-right font-mono text-sm font-semibold">KES {(b.grand_total / 100).toLocaleString()}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {activeTab === 'trends' && (
            <div className="card">
              <div className="card-header">
                <h2 className="text-base font-semibold text-black">Contribution Trends</h2>
              </div>
              <div className="card-body">
                {trends.length > 0 ? (
                  <ResponsiveContainer width="100%" height={280}>
                    <LineChart data={trends}>
                      <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#e5e5e5" />
                      <XAxis dataKey="month" fontSize={11} tick={{ fill: '#666' }} axisLine={false} tickLine={false} />
                      <YAxis fontSize={11} tick={{ fill: '#666' }} axisLine={false} tickLine={false} tickFormatter={(v) => `KES ${(v/1000).toFixed(0)}k`} />
                      <Tooltip formatter={(v) => `KES ${v.toLocaleString()}`} contentStyle={{ borderRadius: '4px', border: '1px solid #e5e5e5' }} />
                      <Line type="monotone" dataKey="total_amount" stroke="#000" strokeWidth={2} name="Total" dot={false} />
                    </LineChart>
                  </ResponsiveContainer>
                ) : <p className="text-gray-400 text-center py-12">No data available</p>}
              </div>
            </div>
          )}

          {activeTab === 'ytd' && (
            <div className="card">
              <table className="table">
                <thead>
                  <tr>
                    <th className="text-left">Member</th>
                    <th className="text-right">Employee YTD</th>
                    <th className="text-right">Employer YTD</th>
                    <th className="text-right">Total YTD</th>
                  </tr>
                </thead>
                <tbody>
                  {ytd.slice(0, 20).map((y, i) => (
                    <tr key={i}>
                      <td className="font-medium text-black">{y.full_name}</td>
                      <td className="text-right font-mono text-sm">KES {(y.employee_ytd / 100).toLocaleString()}</td>
                      <td className="text-right font-mono text-sm">KES {(y.employer_ytd / 100).toLocaleString()}</td>
                      <td className="text-right font-mono text-sm font-semibold">KES {(y.total_ytd / 100).toLocaleString()}</td>
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
