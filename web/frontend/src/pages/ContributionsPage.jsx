import { useState, useEffect } from 'react';
import { contributions } from '../lib/api';
import { 
  BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, 
  AreaChart, Area, PieChart, Pie, Cell
} from 'recharts';
import { CreditCard, TrendingUp, Loader2, Plus, Download, DollarSign, Users, Calendar, Building2 } from 'lucide-react';
import { Link } from 'react-router-dom';

const COLORS = ['#000000', '#333333', '#666666', '#999999', '#CCCCCC'];

export default function ContributionsPage() {
  const [data, setData] = useState([]);
  const [loading, setLoading] = useState(true);
  const [chartData, setChartData] = useState([]);
  const [stats, setStats] = useState(null);

  useEffect(() => {
    setLoading(true);
    contributions.list({ limit: 500 })
      .then(res => {
        const items = res.data || [];
        setData(items);
        
        const grouped = {};
        items.forEach(c => {
          const month = new Date(c.period).toLocaleDateString('en', { month: 'short', year: '2-digit' });
          if (!grouped[month]) grouped[month] = { month, employee: 0, employer: 0, avc: 0, total: 0 };
          grouped[month].employee += (c.employee_amount || 0) / 100;
          grouped[month].employer += (c.employer_amount || 0) / 100;
          grouped[month].avc += (c.avc_amount || 0) / 100;
          grouped[month].total += (c.total_amount || 0) / 100;
        });
        setChartData(Object.values(grouped).reverse());

        const total = items.reduce((sum, c) => sum + (c.total_amount || 0), 0);
        const employee = items.reduce((sum, c) => sum + (c.employee_amount || 0), 0);
        const employer = items.reduce((sum, c) => sum + (c.employer_amount || 0), 0);
        const avc = items.reduce((sum, c) => sum + (c.avc_amount || 0), 0);
        setStats({
          total: total / 100,
          employee: employee / 100,
          employer: employer / 100,
          avc: avc / 100,
          count: items.length,
          avg: items.length > 0 ? (total / items.length) / 100 : 0
        });
      })
      .catch(() => setData([]))
      .finally(() => setLoading(false));
  }, []);

  const formatCurrency = (value) => {
    if (value >= 1000000) return `KES ${(value / 1000000).toFixed(1)}M`;
    if (value >= 1000) return `KES ${(value / 1000).toFixed(0)}K`;
    return `KES ${value}`;
  };

  const paymentData = data.reduce((acc, c) => {
    const method = c.payment_method || 'other';
    if (!acc[method]) acc[method] = { name: method, value: 0, count: 0 };
    acc[method].value += (c.total_amount || 0) / 100;
    acc[method].count += 1;
    return acc;
  }, {});
  const paymentChart = Object.values(paymentData).slice(0, 5);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-end sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-black">Contributions</h1>
          <p className="text-sm text-gray-500 mt-1">{data.length} contribution records</p>
        </div>
        <div className="flex gap-2">
          <button className="btn btn-secondary flex items-center gap-2">
            <Download size={14} /> Export
          </button>
          <Link to="/contributions/new" className="btn btn-primary flex items-center gap-2">
            <Plus size={14} /> Record
          </Link>
        </div>
      </div>

      {/* Stats Cards - Uber Style */}
      {stats && (
        <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
          {[
            { icon: DollarSign, label: 'Total', value: formatCurrency(stats.total) },
            { icon: Users, label: 'Employee', value: formatCurrency(stats.employee) },
            { icon: Building2, label: 'Employer', value: formatCurrency(stats.employer) },
            { icon: TrendingUp, label: 'AVC', value: formatCurrency(stats.avc) },
            { icon: Calendar, label: 'Average', value: formatCurrency(stats.avg) },
          ].map((stat, i) => (
            <div key={i} className="bg-white border border-gray-200 rounded-lg p-4">
              <div className="flex items-center gap-2 mb-2">
                <stat.icon size={14} className="text-gray-400" />
                <p className="text-xs font-medium text-gray-500 uppercase tracking-wide">{stat.label}</p>
              </div>
              <p className="text-lg font-bold text-black">{stat.value}</p>
            </div>
          ))}
        </div>
      )}

      {/* Charts */}
      {chartData.length > 0 && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Trends Chart */}
          <div className="bg-white border border-gray-200 rounded-lg p-5">
            <h2 className="text-base font-semibold text-black mb-4">Contribution Trends</h2>
            <ResponsiveContainer width="100%" height={260}>
              <AreaChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#e5e5e5" />
                <XAxis dataKey="month" fontSize={11} tick={{ fill: '#666' }} axisLine={false} tickLine={false} />
                <YAxis fontSize={11} tick={{ fill: '#666' }} axisLine={false} tickLine={false} tickFormatter={(v) => `KES ${(v/1000).toFixed(0)}k`} />
                <Tooltip formatter={(v) => formatCurrency(v)} contentStyle={{ borderRadius: '4px', border: '1px solid #e5e5e5' }} />
                <Area type="monotone" dataKey="employee" stackId="1" stroke="#000" fill="#000" name="Employee" />
                <Area type="monotone" dataKey="employer" stackId="1" stroke="#666" fill="#666" name="Employer" />
                <Area type="monotone" dataKey="avc" stackId="1" stroke="#999" fill="#999" name="AVC" />
              </AreaChart>
            </ResponsiveContainer>
          </div>

          {/* Payment Methods */}
          {paymentChart.length > 0 && (
            <div className="bg-white border border-gray-200 rounded-lg p-5">
              <h2 className="text-base font-semibold text-black mb-4">By Payment Method</h2>
              <ResponsiveContainer width="100%" height={260}>
                <PieChart>
                  <Pie
                    data={paymentChart}
                    cx="50%"
                    cy="50%"
                    innerRadius={50}
                    outerRadius={90}
                    paddingAngle={2}
                    dataKey="value"
                  >
                    {paymentChart.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                    ))}
                  </Pie>
                  <Tooltip formatter={(v) => formatCurrency(v)} contentStyle={{ borderRadius: '4px', border: '1px solid #e5e5e5' }} />
                </PieChart>
              </ResponsiveContainer>
              <div className="flex flex-wrap gap-4 justify-center mt-4">
                {paymentChart.map((item, i) => (
                  <div key={i} className="flex items-center gap-2">
                    <div className="w-3 h-3 rounded-full" style={{ backgroundColor: COLORS[i % COLORS.length] }} />
                    <span className="text-xs text-gray-500 capitalize">{item.name}</span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      {/* Recent Contributions Table */}
      <div className="card">
        <div className="card-header">
          <h2 className="text-base font-semibold text-black">Recent Contributions</h2>
        </div>
        {loading ? (
          <div className="p-16 text-center">
            <Loader2 size={24} className="animate-spin mx-auto text-gray-300" />
          </div>
        ) : data.length === 0 ? (
          <div className="empty-state">
            <p className="text-sm text-gray-400">No contributions found</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="table">
              <thead>
                <tr>
                  <th className="text-left">Member</th>
                  <th className="text-left">Period</th>
                  <th className="text-right">Employee</th>
                  <th className="text-right">Employer</th>
                  <th className="text-right">AVC</th>
                  <th className="text-right">Total</th>
                  <th className="text-left">Status</th>
                </tr>
              </thead>
              <tbody>
                {data.slice(0, 10).map((c) => (
                  <tr key={c.id}>
                    <td className="font-medium">{c.member_name || c.member_id}</td>
                    <td className="text-gray-500">{new Date(c.period).toLocaleDateString()}</td>
                    <td className="text-right font-mono text-sm">KES {(c.employee_amount / 100).toLocaleString()}</td>
                    <td className="text-right font-mono text-sm">KES {(c.employer_amount / 100).toLocaleString()}</td>
                    <td className="text-right font-mono text-sm">KES {(c.avc_amount / 100).toLocaleString()}</td>
                    <td className="text-right font-mono text-sm font-semibold">KES {(c.total_amount / 100).toLocaleString()}</td>
                    <td>
                      <span className={`badge ${c.status === 'confirmed' ? 'badge-success' : 'badge-warning'}`}>
                        {c.status || 'pending'}
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
