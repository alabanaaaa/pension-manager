import { useState, useEffect } from 'react';
import { tax } from '../lib/api';
import { Loader2, TrendingUp, Shield, Calculator, AlertCircle } from 'lucide-react';

export default function TaxPage() {
  const [brackets, setBrackets] = useState([]);
  const [reliefs, setReliefs] = useState([]);
  const [expiring, setExpiring] = useState([]);
  const [overdue, setOverdue] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.all([
      tax.getBrackets().then(r => setBrackets(Array.isArray(r.data) ? r.data : [])).catch(() => []),
      tax.getReliefs().then(r => setReliefs(Array.isArray(r.data) ? r.data : [])).catch(() => []),
      tax.getExpiring(30).then(r => setExpiring(Array.isArray(r.data) ? r.data : [])).catch(() => []),
      tax.getOverdue().then(r => setOverdue(Array.isArray(r.data) ? r.data : [])).catch(() => []),
    ]).finally(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <div className="p-16 text-center">
        <Loader2 size={24} className="animate-spin mx-auto text-neutral-300" />
        <p className="text-sm text-neutral-400 mt-3">Loading tax data...</p>
      </div>
    );
  }

  return (
    <div className="space-y-8 animate-fade-in-up">
      <div>
        <h1 className="text-3xl font-semibold tracking-tight text-neutral-900">Tax Management</h1>
        <p className="text-neutral-500 mt-2 text-base">KRA tax computation and exemption tracking</p>
      </div>

      {/* Alerts */}
      {expiring.length > 0 && (
        <div className="bg-amber-50 border border-amber-100 rounded-2xl p-6">
          <div className="flex items-center gap-4">
            <div className="p-3 bg-amber-100 rounded-xl"><AlertCircle size={20} className="text-amber-600" /></div>
            <div className="flex-1">
              <h3 className="font-medium text-amber-900">{expiring.length} Tax Exemptions Expiring Soon</h3>
              <p className="text-sm text-amber-600 mt-0.5">Members with KRA certificates expiring within 30 days</p>
            </div>
          </div>
        </div>
      )}

      {overdue.length > 0 && (
        <div className="bg-red-50 border border-red-100 rounded-2xl p-6">
          <div className="flex items-center gap-4">
            <div className="p-3 bg-red-100 rounded-xl"><AlertCircle size={20} className="text-red-600" /></div>
            <div className="flex-1">
              <h3 className="font-medium text-red-900">{overdue.length} Expired Tax Exemptions</h3>
              <p className="text-sm text-red-600 mt-0.5">Members with expired KRA exemption certificates</p>
            </div>
          </div>
        </div>
      )}

      {/* Tax Brackets */}
      <div className="bg-white rounded-2xl border border-neutral-50 overflow-hidden">
        <div className="flex items-center justify-between px-6 py-5 border-b border-neutral-50">
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-blue-50"><TrendingUp size={20} className="text-blue-600" /></div>
            <h2 className="text-lg font-semibold tracking-tight text-neutral-900">KRA PAYE Tax Brackets</h2>
          </div>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-neutral-50">
                <th className="text-left px-6 py-4 font-medium text-neutral-400 text-xs uppercase tracking-wider">Min (KES)</th>
                <th className="text-left px-6 py-4 font-medium text-neutral-400 text-xs uppercase tracking-wider">Max (KES)</th>
                <th className="text-right px-6 py-4 font-medium text-neutral-400 text-xs uppercase tracking-wider">Rate</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-neutral-50">
              {brackets.map((b, i) => (
                <tr key={i} className="hover:bg-neutral-50/50 transition-colors">
                  <td className="px-6 py-4 font-mono text-sm">{b.Min?.toLocaleString() || '0'}</td>
                  <td className="px-6 py-4 font-mono text-sm">{b.Max ? b.Max.toLocaleString() : 'No limit'}</td>
                  <td className="px-6 py-4 text-right font-semibold">{(b.Rate * 100).toFixed(1)}%</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Tax Reliefs */}
      <div className="bg-white rounded-2xl border border-neutral-50 overflow-hidden">
        <div className="flex items-center justify-between px-6 py-5 border-b border-neutral-50">
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-emerald-50"><Shield size={20} className="text-emerald-600" /></div>
            <h2 className="text-lg font-semibold tracking-tight text-neutral-900">Available Tax Reliefs</h2>
          </div>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-neutral-50">
                <th className="text-left px-6 py-4 font-medium text-neutral-400 text-xs uppercase tracking-wider">Relief</th>
                <th className="text-left px-6 py-4 font-medium text-neutral-400 text-xs uppercase tracking-wider">Description</th>
                <th className="text-right px-6 py-4 font-medium text-neutral-400 text-xs uppercase tracking-wider">Amount (KES/yr)</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-neutral-50">
              {reliefs.map((r, i) => (
                <tr key={i} className="hover:bg-neutral-50/50 transition-colors">
                  <td className="px-6 py-4 font-medium capitalize">{r.Name}</td>
                  <td className="px-6 py-4 text-neutral-500">{r.Description}</td>
                  <td className="px-6 py-4 text-right font-mono text-sm">{(r.Amount / 100).toLocaleString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
