import { useState, useEffect } from 'react';
import { contributions } from '../lib/api';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, LineChart, Line } from 'recharts';
import { CreditCard, TrendingUp, Loader2 } from 'lucide-react';

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
        // Group by month for chart
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
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Contributions</h1>
        <p className="text-gray-500 mt-1">{data.length} records</p>
      </div>

      {/* Chart */}
      {chartData.length > 0 && (
        <div className="bg-white rounded-xl border p-5">
          <h2 className="text-lg font-semibold mb-4">Monthly Trends</h2>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="month" fontSize={12} />
              <YAxis fontSize={12} />
              <Tooltip formatter={(v) => `KES ${v.toLocaleString()}`} />
              <Bar dataKey="employee" fill="#3b82f6" name="Employee" />
              <Bar dataKey="employer" fill="#10b981" name="Employer" />
            </BarChart>
          </ResponsiveContainer>
        </div>
      )}

      {/* Table */}
      <div className="bg-white rounded-xl border overflow-hidden">
        {loading ? (
          <div className="p-12 text-center"><Loader2 size={32} className="animate-spin mx-auto text-gray-400" /></div>
        ) : data.length === 0 ? (
          <div className="p-12 text-center text-gray-500">No contributions found</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="bg-gray-50 border-b">
                <tr>
                  <th className="text-left px-4 py-3 font-medium text-gray-500">Period</th>
                  <th className="text-left px-4 py-3 font-medium text-gray-500">Member</th>
                  <th className="text-right px-4 py-3 font-medium text-gray-500">Employee</th>
                  <th className="text-right px-4 py-3 font-medium text-gray-500">Employer</th>
                  <th className="text-right px-4 py-3 font-medium text-gray-500">AVC</th>
                  <th className="text-right px-4 py-3 font-medium text-gray-500">Total</th>
                  <th className="text-left px-4 py-3 font-medium text-gray-500">Status</th>
                </tr>
              </thead>
              <tbody className="divide-y">
                {data.slice(0, 20).map(c => (
                  <tr key={c.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3">{new Date(c.period).toLocaleDateString()}</td>
                    <td className="px-4 py-3">{c.member_name || c.member_id}</td>
                    <td className="px-4 py-3 text-right font-mono text-xs">KES {((c.employee_amount || 0) / 100).toLocaleString()}</td>
                    <td className="px-4 py-3 text-right font-mono text-xs">KES {((c.employer_amount || 0) / 100).toLocaleString()}</td>
                    <td className="px-4 py-3 text-right font-mono text-xs">KES {((c.avc_amount || 0) / 100).toLocaleString()}</td>
                    <td className="px-4 py-3 text-right font-mono text-xs font-semibold">KES {((c.total_amount || 0) / 100).toLocaleString()}</td>
                    <td className="px-4 py-3">
                      <span className={`px-2 py-1 rounded-full text-xs font-medium ${c.status === 'confirmed' ? 'bg-green-100 text-green-800' : 'bg-yellow-100 text-yellow-800'}`}>
                        {c.status}
                      </span>
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
