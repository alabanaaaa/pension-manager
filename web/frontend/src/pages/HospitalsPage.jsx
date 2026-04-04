import { useState, useEffect, useCallback } from 'react';
import { Link } from 'react-router-dom';
import { hospitals } from '../lib/api';
import {
  Plus, Hospital, AlertTriangle, Loader2, Phone, Mail,
  TrendingUp, TrendingDown, Download, Search, Filter,
  ChevronLeft, ChevronRight, Building2, DollarSign, Users,
  FileSpreadsheet, BarChart3, Clock, ArrowUpRight, ArrowDownRight
} from 'lucide-react';

export default function HospitalsPage() {
  const [hospitalList, setHospitalList] = useState([]);
  const [alerts, setAlerts] = useState(null);
  const [pendingBills, setPendingBills] = useState([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [activeTab, setActiveTab] = useState('hospitals');

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const [hRes, aRes, bRes] = await Promise.all([
        hospitals.list().then(r => Array.isArray(r.data) ? r.data : []).catch(() => []),
        hospitals.getAlerts().then(r => r.data || null).catch(() => null),
        hospitals.getPendingBills().then(r => Array.isArray(r.data) ? r.data : []).catch(() => []),
      ]);
      setHospitalList(hRes);
      setAlerts(aRes);
      setPendingBills(bRes);
    } catch {
      setHospitalList([]);
      setAlerts(null);
      setPendingBills([]);
    }
    finally { setLoading(false); }
  }, []);

  useEffect(() => { fetchData(); }, [fetchData]);

  const filteredHospitals = hospitalList.filter(h =>
    h.name?.toLowerCase().includes(search.toLowerCase()) ||
    h.address?.toLowerCase().includes(search.toLowerCase())
  );

  const totalBalance = hospitalList.reduce((sum, h) => sum + (h.account_balance || 0), 0);
  const activeHospitals = hospitalList.filter(h => h.status === 'active').length;

  const tabs = [
    { id: 'hospitals', label: 'Hospitals', icon: Building2, count: hospitalList.length },
    { id: 'pending', label: 'Pending Bills', icon: Clock, count: pendingBills.length },
    { id: 'alerts', label: 'Alerts', icon: AlertTriangle, count: alerts?.pending_bills || 0 },
  ];

  return (
    <div className="space-y-8 animate-fade-in-up">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-end sm:justify-between gap-4">
        <div>
          <h1 className="text-3xl font-semibold tracking-tight text-neutral-900">Hospital Management</h1>
          <p className="text-neutral-500 mt-2 text-base">Manage hospital accounts, medical expenditures, and pending bills</p>
        </div>
        <div className="flex gap-3">
          <button className="btn-hover flex items-center gap-2 px-4 py-2.5 border border-neutral-200 rounded-xl text-sm font-medium hover:bg-neutral-50 transition-all">
            <FileSpreadsheet size={15} /> Export Excel
          </button>
          <Link to="/hospitals/new" className="btn-hover flex items-center gap-2 px-4 py-2.5 bg-neutral-900 text-white rounded-xl text-sm font-medium hover:bg-neutral-800 transition-all">
            <Plus size={15} /> Add Hospital
          </Link>
        </div>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
        <div className="bg-white rounded-2xl border border-neutral-50 p-6">
          <div className="flex items-center justify-between mb-4">
            <div className="p-2.5 rounded-xl bg-blue-50"><Building2 size={20} className="text-blue-600" /></div>
          </div>
          <p className="text-sm text-neutral-500">Total Hospitals</p>
          <p className="text-2xl font-semibold tracking-tight text-neutral-900 mt-1">{loading ? '—' : hospitalList.length}</p>
          <p className="text-xs text-emerald-600 mt-1 flex items-center gap-1">
            <ArrowUpRight size={12} /> {activeHospitals} active
          </p>
        </div>

        <div className="bg-white rounded-2xl border border-neutral-50 p-6">
          <div className="flex items-center justify-between mb-4">
            <div className="p-2.5 rounded-xl bg-emerald-50"><DollarSign size={20} className="text-emerald-600" /></div>
          </div>
          <p className="text-sm text-neutral-500">Total Hospital Balances</p>
          <p className="text-2xl font-semibold tracking-tight text-neutral-900 mt-1">
            {loading ? '—' : `KES ${(totalBalance / 100).toLocaleString()}`}
          </p>
        </div>

        <div className="bg-white rounded-2xl border border-neutral-50 p-6">
          <div className="flex items-center justify-between mb-4">
            <div className="p-2.5 rounded-xl bg-amber-50"><Clock size={20} className="text-amber-600" /></div>
          </div>
          <p className="text-sm text-neutral-500">Pending Bills (&gt;45 days)</p>
          <p className="text-2xl font-semibold tracking-tight text-neutral-900 mt-1">
            {loading ? '—' : alerts?.pending_bills || 0}
          </p>
          <p className="text-xs text-amber-600 mt-1">
            KES {((alerts?.total_pending_amount || 0) / 100).toLocaleString()} total
          </p>
        </div>

        <div className="bg-white rounded-2xl border border-neutral-50 p-6">
          <div className="flex items-center justify-between mb-4">
            <div className="p-2.5 rounded-xl bg-red-50"><AlertTriangle size={20} className="text-red-600" /></div>
          </div>
          <p className="text-sm text-neutral-500">High Urgency (&gt;60 days)</p>
          <p className="text-2xl font-semibold tracking-tight text-neutral-900 mt-1">
            {loading ? '—' : alerts?.high_urgency_bills || 0}
          </p>
          <p className="text-xs text-red-600 mt-1 flex items-center gap-1">
            <ArrowDownRight size={12} /> Requires immediate attention
          </p>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 bg-neutral-100 p-1 rounded-xl w-fit">
        {tabs.map(tab => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-all ${
              activeTab === tab.id ? 'bg-white text-neutral-900 shadow-sm' : 'text-neutral-400 hover:text-neutral-600'
            }`}
          >
            <tab.icon size={14} />
            {tab.label}
            {tab.count > 0 && (
              <span className={`px-2 py-0.5 rounded-full text-xs ${
                activeTab === tab.id ? 'bg-neutral-100 text-neutral-600' : 'bg-neutral-200 text-neutral-500'
              }`}>
                {tab.count}
              </span>
            )}
          </button>
        ))}
      </div>

      {/* Hospitals Tab */}
      {activeTab === 'hospitals' && (
        <div className="space-y-6">
          {/* Search */}
          <div className="relative">
            <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-neutral-300" />
            <input
              type="text"
              placeholder="Search hospitals by name or address..."
              value={search}
              onChange={e => setSearch(e.target.value)}
              className="w-full pl-9 pr-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all placeholder:text-neutral-300"
            />
          </div>

          {loading ? (
            <div className="p-16 text-center">
              <Loader2 size={24} className="animate-spin mx-auto text-neutral-300" />
              <p className="text-sm text-neutral-400 mt-3">Loading hospitals...</p>
            </div>
          ) : filteredHospitals.length === 0 ? (
            <div className="bg-white rounded-2xl border border-neutral-50 p-16 text-center">
              <Hospital size={48} className="mx-auto text-neutral-200 mb-4" />
              <p className="text-neutral-500 mb-4">No hospitals found</p>
              <Link to="/hospitals/new" className="text-sm text-neutral-900 font-medium hover:underline">Add one to get started</Link>
            </div>
          ) : (
            <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
              {filteredHospitals.map((h, i) => (
                <Link
                  key={h.id}
                  to={`/hospitals/${h.id}`}
                  className="btn-hover bg-white rounded-2xl border border-neutral-50 p-6 hover:shadow-sm transition-all animate-fade-in"
                  style={{ animationDelay: `${i * 0.03}s` }}
                >
                  <div className="flex items-start gap-4">
                    <div className="p-2.5 rounded-xl bg-neutral-50 flex-shrink-0">
                      <Hospital size={20} className="text-neutral-600" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-start justify-between gap-2">
                        <h3 className="font-semibold text-neutral-900 truncate">{h.name}</h3>
                        <span className={`px-2 py-0.5 rounded-full text-xs font-medium flex-shrink-0 ${
                          h.status === 'active' ? 'bg-emerald-50 text-emerald-700' :
                          h.status === 'suspended' ? 'bg-red-50 text-red-700' :
                          'bg-neutral-100 text-neutral-600'
                        }`}>
                          {h.status}
                        </span>
                      </div>
                      {h.address && <p className="text-sm text-neutral-400 mt-1 truncate">{h.address}</p>}
                      <div className="flex flex-wrap gap-4 mt-3 text-xs text-neutral-400">
                        {h.phone && <span className="flex items-center gap-1"><Phone size={12} /> {h.phone}</span>}
                        {h.email && <span className="flex items-center gap-1"><Mail size={12} /> {h.email}</span>}
                      </div>
                      <div className="mt-4 pt-4 border-t border-neutral-50 flex items-center justify-between">
                        <span className="text-xs text-neutral-400">Account Balance</span>
                        <span className="text-sm font-mono font-semibold text-neutral-900">
                          KES {((h.account_balance || 0) / 100).toLocaleString()}
                        </span>
                      </div>
                    </div>
                  </div>
                </Link>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Pending Bills Tab */}
      {activeTab === 'pending' && (
        <div className="bg-white rounded-2xl border border-neutral-50 overflow-hidden">
          <div className="flex items-center justify-between px-6 py-5 border-b border-neutral-50">
            <div className="flex items-center gap-3">
              <div className="p-2.5 rounded-xl bg-amber-50"><Clock size={20} className="text-amber-600" /></div>
              <div>
                <h2 className="text-lg font-semibold tracking-tight text-neutral-900">Pending Bills (&gt;45 days)</h2>
                <p className="text-sm text-neutral-400">Medical expenditures awaiting payment</p>
              </div>
            </div>
            <button className="btn-hover flex items-center gap-2 px-4 py-2.5 border border-neutral-200 rounded-xl text-sm font-medium hover:bg-neutral-50 transition-all">
              <Download size={15} /> Export
            </button>
          </div>
          {loading ? (
            <div className="p-16 text-center"><Loader2 size={24} className="animate-spin mx-auto text-neutral-300" /><p className="text-sm text-neutral-400 mt-3">Loading...</p></div>
          ) : pendingBills.length === 0 ? (
            <div className="p-16 text-center"><p className="text-neutral-500">No pending bills over 45 days</p></div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-neutral-50">
                    <th className="text-left px-6 py-4 font-medium text-neutral-400 text-xs uppercase tracking-wider">Member</th>
                    <th className="text-left px-6 py-4 font-medium text-neutral-400 text-xs uppercase tracking-wider">Service Type</th>
                    <th className="text-left px-6 py-4 font-medium text-neutral-400 text-xs uppercase tracking-wider">Date Submitted</th>
                    <th className="text-right px-6 py-4 font-medium text-neutral-400 text-xs uppercase tracking-wider">Amount Charged</th>
                    <th className="text-right px-6 py-4 font-medium text-neutral-400 text-xs uppercase tracking-wider">Amount Covered</th>
                    <th className="text-right px-6 py-4 font-medium text-neutral-400 text-xs uppercase tracking-wider">Member Pays</th>
                    <th className="text-left px-6 py-4 font-medium text-neutral-400 text-xs uppercase tracking-wider">Days Overdue</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-neutral-50">
                  {pendingBills.map((bill, i) => {
                    const daysOverdue = Math.floor((Date.now() - new Date(bill.date_submitted).getTime()) / 86400000);
                    return (
                      <tr key={bill.id} className="hover:bg-neutral-50/50 transition-colors">
                        <td className="px-6 py-4 font-medium text-neutral-900">{bill.member_name || bill.member_id}</td>
                        <td className="px-6 py-4 text-neutral-500 capitalize">{bill.service_type}</td>
                        <td className="px-6 py-4 text-neutral-400 text-xs">{new Date(bill.date_submitted).toLocaleDateString()}</td>
                        <td className="px-6 py-4 text-right font-mono text-xs text-neutral-600">KES {((bill.amount_charged || 0) / 100).toLocaleString()}</td>
                        <td className="px-6 py-4 text-right font-mono text-xs text-emerald-600">KES {((bill.amount_covered || 0) / 100).toLocaleString()}</td>
                        <td className="px-6 py-4 text-right font-mono text-xs text-amber-600">KES {((bill.member_responsibility || 0) / 100).toLocaleString()}</td>
                        <td className="px-6 py-4">
                          <span className={`px-2 py-1 rounded-full text-xs font-medium ${
                            daysOverdue > 60 ? 'bg-red-50 text-red-700' : 'bg-amber-50 text-amber-700'
                          }`}>
                            {daysOverdue} days
                          </span>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {/* Alerts Tab */}
      {activeTab === 'alerts' && (
        <div className="space-y-6">
          {alerts?.high_urgency_bills > 0 && (
            <div className="bg-red-50 border border-red-100 rounded-2xl p-6">
              <div className="flex items-center gap-4">
                <div className="p-3 bg-red-100 rounded-xl"><AlertTriangle size={20} className="text-red-600" /></div>
                <div className="flex-1">
                  <h3 className="font-semibold text-red-900">{alerts.high_urgency_bills} High Urgency Bills (&gt;60 days)</h3>
                  <p className="text-sm text-red-600 mt-1">These bills require immediate attention and payment processing</p>
                </div>
              </div>
            </div>
          )}

          {alerts?.medium_urgency_bills > 0 && (
            <div className="bg-amber-50 border border-amber-100 rounded-2xl p-6">
              <div className="flex items-center gap-4">
                <div className="p-3 bg-amber-100 rounded-xl"><Clock size={20} className="text-amber-600" /></div>
                <div className="flex-1">
                  <h3 className="font-semibold text-amber-900">{alerts.medium_urgency_bills} Medium Urgency Bills (45-60 days)</h3>
                  <p className="text-sm text-amber-600 mt-1">These bills are approaching the critical threshold</p>
                </div>
              </div>
            </div>
          )}

          {alerts && (
            <div className="bg-white rounded-2xl border border-neutral-50 overflow-hidden">
              <div className="flex items-center justify-between px-6 py-5 border-b border-neutral-50">
                <div className="flex items-center gap-3">
                  <div className="p-2.5 rounded-xl bg-blue-50"><BarChart3 size={20} className="text-blue-600" /></div>
                  <h2 className="text-lg font-semibold tracking-tight text-neutral-900">Expenditure Summary</h2>
                </div>
              </div>
              <div className="p-6 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
                <div className="text-center p-4 bg-neutral-50 rounded-xl">
                  <p className="text-3xl font-bold text-neutral-900">{alerts.pending_bills}</p>
                  <p className="text-sm text-neutral-500 mt-1">Total Pending</p>
                </div>
                <div className="text-center p-4 bg-red-50 rounded-xl">
                  <p className="text-3xl font-bold text-red-600">{alerts.high_urgency_bills}</p>
                  <p className="text-sm text-red-500 mt-1">High Urgency</p>
                </div>
                <div className="text-center p-4 bg-amber-50 rounded-xl">
                  <p className="text-3xl font-bold text-amber-600">{alerts.medium_urgency_bills}</p>
                  <p className="text-sm text-amber-500 mt-1">Medium Urgency</p>
                </div>
                <div className="text-center p-4 bg-emerald-50 rounded-xl">
                  <p className="text-3xl font-bold text-emerald-600">{alerts.low_urgency_bills}</p>
                  <p className="text-sm text-emerald-500 mt-1">Low Urgency</p>
                </div>
              </div>
              <div className="px-6 pb-6">
                <div className="p-4 bg-neutral-50 rounded-xl flex items-center justify-between">
                  <span className="text-sm text-neutral-500">Total Pending Amount</span>
                  <span className="text-xl font-mono font-semibold text-neutral-900">
                    KES {((alerts.total_pending_amount || 0) / 100).toLocaleString()}
                  </span>
                </div>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
