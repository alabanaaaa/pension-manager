import { useState, useEffect } from 'react';
import { reports } from '../lib/api';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, LineChart, Line, PieChart, Pie, Cell } from 'recharts';
import { Download, Loader2 } from 'lucide-react';

const COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6'];

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
    { id: 'breakdown', label: 'Breakdown' },
    { id: 'trends', label: 'Trends' },
    { id: 'ytd', label: 'YTD' },
  ];

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Reports</h1>
          <p className="text-gray-500 mt-1">Contribution analysis and trends</p>
        </div>
        <button className="flex items-center gap-2 px-4 py-2 border rounded-lg text-sm font-medium hover:bg-gray-50">
          <Download size={16} /> Export
        </button>
      </div>

      <div className="flex gap-1 bg-gray-100 p-1 rounded-lg w-fit">
        {tabs.map(tab => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${activeTab === tab.id ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-500 hover:text-gray-700'}`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {loading ? (
        <div className="p-12 text-center"><Loader2 size={32} className="animate-spin mx-auto text-gray-400" /></div>
      ) : (
        <>
          {activeTab === 'breakdown' && (
            <div className="space-y-6">
              <div className="bg-white rounded-xl border p-5">
                <h2 className="text-lg font-semibold mb-4">Monthly Breakdown</h2>
                {breakdown.length > 0 ? (
                  <ResponsiveContainer width="100%" height={300}>
                    <BarChart data={breakdown}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey="period" fontSize={12} />
                      <YAxis fontSize={12} />
                      <Tooltip formatter={(v) => `KES ${(v / 100).toLocaleString()}`} />
                      <Bar dataKey="employee_total" fill="#3b82f6" name="Employee" />
                      <Bar dataKey="employer_total" fill="#10b981" name="Employer" />
                    </BarChart>
                  </ResponsiveContainer>
                ) : <p className="text-gray-500 text-center py-8">No data available</p>}
              </div>
              <div className="bg-white rounded-xl border overflow-hidden">
                <table className="w-full text-sm">
                  <thead className="bg-gray-50 border-b">
                    <tr>
                      <th className="text-left px-4 py-3 font-medium text-gray-500">Period</th>
                      <th className="text-right px-4 py-3 font-medium text-gray-500">Employees</th>
                      <th className="text-right px-4 py-3 font-medium text-gray-500">Employee Total</th>
                      <th className="text-right px-4 py-3 font-medium text-gray-500">Employer Total</th>
                      <th className="text-right px-4 py-3 font-medium text-gray-500">Grand Total</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y">
                    {breakdown.map((b, i) => (
                      <tr key={i} className="hover:bg-gray-50">
                        <td className="px-4 py-3">{b.period}</td>
                        <td className="px-4 py-3 text-right">{b.employee_count}</td>
                        <td className="px-4 py-3 text-right font-mono text-xs">KES {(b.employee_total / 100).toLocaleString()}</td>
                        <td className="px-4 py-3 text-right font-mono text-xs">KES {(b.employer_total / 100).toLocaleString()}</td>
                        <td className="px-4 py-3 text-right font-mono text-xs font-semibold">KES {(b.grand_total / 100).toLocaleString()}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {activeTab === 'trends' && (
            <div className="bg-white rounded-xl border p-5">
              <h2 className="text-lg font-semibold mb-4">Contribution Trends</h2>
              {trends.length > 0 ? (
                <ResponsiveContainer width="100%" height={300}>
                  <LineChart data={trends}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="month" fontSize={12} />
                    <YAxis fontSize={12} />
                    <Tooltip formatter={(v) => `KES ${(v / 100).toLocaleString()}`} />
                    <Line type="monotone" dataKey="total_amount" stroke="#3b82f6" strokeWidth={2} name="Total" />
                    <Line type="monotone" dataKey="member_count" stroke="#10b981" strokeWidth={2} name="Members" yAxisId="right" />
                  </LineChart>
                </ResponsiveContainer>
              ) : <p className="text-gray-500 text-center py-8">No data available</p>}
            </div>
          )}

          {activeTab === 'ytd' && (
            <div className="bg-white rounded-xl border overflow-hidden">
              <table className="w-full text-sm">
                <thead className="bg-gray-50 border-b">
                  <tr>
                    <th className="text-left px-4 py-3 font-medium text-gray-500">Member</th>
                    <th className="text-right px-4 py-3 font-medium text-gray-500">Employee YTD</th>
                    <th className="text-right px-4 py-3 font-medium text-gray-500">Employer YTD</th>
                    <th className="text-right px-4 py-3 font-medium text-gray-500">Total YTD</th>
                  </tr>
                </thead>
                <tbody className="divide-y">
                  {ytd.slice(0, 20).map((y, i) => (
                    <tr key={i} className="hover:bg-gray-50">
                      <td className="px-4 py-3 font-medium">{y.full_name}</td>
                      <td className="px-4 py-3 text-right font-mono text-xs">KES {(y.employee_ytd / 100).toLocaleString()}</td>
                      <td className="px-4 py-3 text-right font-mono text-xs">KES {(y.employer_ytd / 100).toLocaleString()}</td>
                      <td className="px-4 py-3 text-right font-mono text-xs font-semibold">KES {(y.total_ytd / 100).toLocaleString()}</td>
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
