import { useState, useEffect } from 'react';
import { tax, taxExemptions } from '../lib/api';
import { Loader2, TrendingUp, Shield, Calculator, AlertCircle, Plus, Eye, Check, X } from 'lucide-react';

export default function TaxPage() {
  const [brackets, setBrackets] = useState([]);
  const [reliefs, setReliefs] = useState([]);
  const [expiring, setExpiring] = useState([]);
  const [overdue, setOverdue] = useState([]);
  const [exemptions, setExemptions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState('overview');
  const [showAddModal, setShowAddModal] = useState(false);

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    setLoading(true);
    try {
      const [bracketsRes, reliefsRes, expiringRes, overdueRes, exemptionsRes] = await Promise.all([
        tax.getBrackets().catch(() => ({ data: [] })),
        tax.getReliefs().catch(() => ({ data: [] })),
        tax.getExpiring(30).catch(() => ({ data: [] })),
        tax.getOverdue().catch(() => ({ data: [] })),
        taxExemptions.list({ limit: 100 }).catch(() => ({ data: [] })),
      ]);
      setBrackets(Array.isArray(bracketsRes.data) ? bracketsRes.data : []);
      setReliefs(Array.isArray(reliefsRes.data) ? reliefsRes.data : []);
      setExpiring(Array.isArray(expiringRes.data) ? expiringRes.data : []);
      setOverdue(Array.isArray(overdueRes.data) ? overdueRes.data : []);
      setExemptions(Array.isArray(exemptionsRes.data) ? exemptionsRes.data : []);
    } catch (e) {
      console.error('Failed to fetch tax data:', e);
    } finally {
      setLoading(false);
    }
  };

  const handleApproveExemption = async (id) => {
    try {
      await taxExemptions.approve(id);
      fetchData();
    } catch (e) {
      console.error('Failed to approve exemption:', e);
    }
  };

  const handleRejectExemption = async (id) => {
    const reason = prompt('Enter rejection reason:');
    if (reason) {
      try {
        await taxExemptions.reject(id, reason);
        fetchData();
      } catch (e) {
        console.error('Failed to reject exemption:', e);
      }
    }
  };

  const tabs = [
    { id: 'overview', label: 'Overview' },
    { id: 'exemptions', label: `Exemptions (${exemptions.length})` },
    { id: 'brackets', label: 'Tax Brackets' },
    { id: 'reliefs', label: 'Tax Reliefs' },
  ];

  if (loading) {
    return (
      <div className="p-16 text-center">
        <Loader2 size={24} className="animate-spin mx-auto text-neutral-300" />
        <p className="text-sm text-neutral-400 mt-3">Loading tax data...</p>
      </div>
    );
  }

  return (
    <div className="space-y-6 animate-fade-in-up">
      <div className="flex items-end justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-black">Tax Management</h1>
          <p className="text-sm text-gray-500 mt-1">KRA tax computation and exemption tracking</p>
        </div>
        {activeTab === 'exemptions' && (
          <button onClick={() => setShowAddModal(true)} className="btn btn-primary flex items-center gap-2">
            <Plus size={14} /> Add Exemption
          </button>
        )}
      </div>

      {/* Tabs */}
      <div className="flex gap-1 border-b border-gray-200">
        {tabs.map(tab => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`px-4 py-3 text-sm font-medium transition-all border-b-2 -mb-px ${
              activeTab === tab.id
                ? 'border-black text-black'
                : 'border-transparent text-gray-500 hover:text-black'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Alerts */}
      {(expiring.length > 0 || overdue.length > 0) && activeTab === 'overview' && (
        <div className="space-y-4">
          {expiring.length > 0 && (
            <div className="bg-amber-50 border border-amber-200 rounded-lg p-4">
              <div className="flex items-center gap-3">
                <AlertCircle size={20} className="text-amber-600" />
                <div>
                  <h3 className="font-medium text-amber-900">{expiring.length} Tax Exemptions Expiring Soon</h3>
                  <p className="text-sm text-amber-600">Members with KRA certificates expiring within 30 days</p>
                </div>
              </div>
            </div>
          )}
          {overdue.length > 0 && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-4">
              <div className="flex items-center gap-3">
                <AlertCircle size={20} className="text-red-600" />
                <div>
                  <h3 className="font-medium text-red-900">{overdue.length} Expired Tax Exemptions</h3>
                  <p className="text-sm text-red-600">Members with expired KRA exemption certificates</p>
                </div>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Tax Exemptions */}
      {activeTab === 'exemptions' && (
        <div className="card">
          {exemptions.length === 0 ? (
            <div className="p-16 text-center">
              <Shield size={32} className="mx-auto text-gray-300 mb-3" />
              <p className="text-sm text-gray-400">No tax exemptions found</p>
              <button onClick={() => setShowAddModal(true)} className="btn btn-secondary mt-4">
                Add First Exemption
              </button>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="table">
                <thead>
                  <tr>
                    <th className="text-left">Member</th>
                    <th className="text-left">Type</th>
                    <th className="text-left">Certificate No</th>
                    <th className="text-left">Expiry Date</th>
                    <th className="text-right">Monthly Limit</th>
                    <th className="text-left">Status</th>
                    <th className="text-right">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {exemptions.map(ex => {
                    const member = ex.member || {};
                    return (
                      <tr key={ex.id}>
                        <td className="font-medium">
                          {member.first_name} {member.last_name}
                          <span className="block text-xs text-gray-400">{member.member_no}</span>
                        </td>
                        <td className="text-gray-500">{ex.exemption_type || 'N/A'}</td>
                        <td className="font-mono text-sm">{ex.certificate_no || '—'}</td>
                        <td className="text-gray-500">
                          {ex.expiry_date ? new Date(ex.expiry_date).toLocaleDateString() : '—'}
                        </td>
                        <td className="text-right font-mono text-sm">
                          {ex.monthly_limit ? `KES ${(ex.monthly_limit / 100).toLocaleString()}` : '—'}
                        </td>
                        <td>
                          <span className={`badge ${
                            ex.status === 'approved' ? 'badge-success' :
                            ex.status === 'rejected' ? 'badge-error' :
                            'badge-warning'
                          }`}>
                            {ex.status || 'pending'}
                          </span>
                        </td>
                        <td className="text-right">
                          <div className="flex items-center justify-end gap-1">
                            {ex.status === 'pending' && (
                              <>
                                <button
                                  onClick={() => handleApproveExemption(ex.id)}
                                  className="action-menu text-green-600 hover:bg-green-50"
                                  title="Approve"
                                >
                                  <Check size={15} />
                                </button>
                                <button
                                  onClick={() => handleRejectExemption(ex.id)}
                                  className="action-menu text-red-600 hover:bg-red-50"
                                  title="Reject"
                                >
                                  <X size={15} />
                                </button>
                              </>
                            )}
                            <button className="action-menu" title="View">
                              <Eye size={15} />
                            </button>
                          </div>
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

      {/* Tax Brackets */}
      {activeTab === 'brackets' && (
        <div className="card">
          <div className="card-header">
            <h2 className="text-base font-semibold text-black">KRA PAYE Tax Brackets</h2>
          </div>
          {brackets.length === 0 ? (
            <div className="p-16 text-center">
              <TrendingUp size={32} className="mx-auto text-gray-300 mb-3" />
              <p className="text-sm text-gray-400">No tax brackets configured</p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="table">
                <thead>
                  <tr>
                    <th className="text-left">Min Income (KES)</th>
                    <th className="text-left">Max Income (KES)</th>
                    <th className="text-right">Rate</th>
                    <th className="text-right">Fixed Amount</th>
                  </tr>
                </thead>
                <tbody>
                  {brackets.map((b, i) => (
                    <tr key={i}>
                      <td className="font-mono">{(b.Min || 0).toLocaleString()}</td>
                      <td className="font-mono">{b.Max ? b.Max.toLocaleString() : 'No limit'}</td>
                      <td className="text-right font-semibold">{(b.Rate * 100).toFixed(1)}%</td>
                      <td className="text-right font-mono">
                        {b.FixedAmount ? `KES ${(b.FixedAmount / 100).toLocaleString()}` : '—'}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {/* Tax Reliefs */}
      {activeTab === 'reliefs' && (
        <div className="card">
          <div className="card-header">
            <h2 className="text-base font-semibold text-black">Available Tax Reliefs</h2>
          </div>
          {reliefs.length === 0 ? (
            <div className="p-16 text-center">
              <Shield size={32} className="mx-auto text-gray-300 mb-3" />
              <p className="text-sm text-gray-400">No tax reliefs configured</p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="table">
                <thead>
                  <tr>
                    <th className="text-left">Relief</th>
                    <th className="text-left">Description</th>
                    <th className="text-right">Annual Amount (KES)</th>
                  </tr>
                </thead>
                <tbody>
                  {reliefs.map((r, i) => (
                    <tr key={i}>
                      <td className="font-medium capitalize">{r.Name}</td>
                      <td className="text-gray-500">{r.Description}</td>
                      <td className="text-right font-mono">
                        {r.Amount ? `KES ${(r.Amount / 100).toLocaleString()}` : '—'}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {/* Add Exemption Modal */}
      {showAddModal && (
        <AddExemptionModal
          onClose={() => setShowAddModal(false)}
          onSuccess={() => {
            setShowAddModal(false);
            fetchData();
          }}
        />
      )}
    </div>
  );
}

function AddExemptionModal({ onClose, onSuccess }) {
  const [form, setForm] = useState({
    member_id: '',
    exemption_type: 'age_65_exemption',
    reason: '',
    certificate_no: '',
    expiry_date: '',
    monthly_limit: '',
    relief_amount: '',
    kra_reference: '',
  });
  const [saving, setSaving] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setSaving(true);
    try {
      await taxExemptions.create({
        ...form,
        monthly_limit: form.monthly_limit ? parseInt(form.monthly_limit) * 100 : 0,
        relief_amount: form.relief_amount ? parseInt(form.relief_amount) * 100 : 0,
      });
      onSuccess();
    } catch (err) {
      alert('Failed to create exemption: ' + (err.response?.data?.error || err.message));
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 animate-fade-in">
      <div className="bg-white rounded-lg w-full max-w-lg p-6 animate-fade-in-up">
        <h2 className="text-lg font-semibold text-black mb-4">Add Tax Exemption</h2>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Member ID</label>
            <input
              type="text"
              value={form.member_id}
              onChange={e => setForm({ ...form, member_id: e.target.value })}
              className="input w-full"
              required
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Exemption Type</label>
            <select
              value={form.exemption_type}
              onChange={e => setForm({ ...form, exemption_type: e.target.value })}
              className="input w-full"
            >
              <option value="age_65_exemption">Age 65+ Exemption</option>
              <option value="disability_exemption">Disability Exemption</option>
              <option value="medical_exemption">Medical Exemption</option>
              <option value="retirement_exemption">Retirement Exemption</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Certificate Number</label>
            <input
              type="text"
              value={form.certificate_no}
              onChange={e => setForm({ ...form, certificate_no: e.target.value })}
              className="input w-full"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Expiry Date</label>
            <input
              type="date"
              value={form.expiry_date}
              onChange={e => setForm({ ...form, expiry_date: e.target.value })}
              className="input w-full"
            />
          </div>
          <div className="flex gap-4">
            <div className="flex-1">
              <label className="block text-sm font-medium text-gray-700 mb-1">Monthly Limit</label>
              <input
                type="number"
                value={form.monthly_limit}
                onChange={e => setForm({ ...form, monthly_limit: e.target.value })}
                className="input w-full"
                placeholder="KES"
              />
            </div>
            <div className="flex-1">
              <label className="block text-sm font-medium text-gray-700 mb-1">Relief Amount</label>
              <input
                type="number"
                value={form.relief_amount}
                onChange={e => setForm({ ...form, relief_amount: e.target.value })}
                className="input w-full"
                placeholder="KES"
              />
            </div>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Reason</label>
            <textarea
              value={form.reason}
              onChange={e => setForm({ ...form, reason: e.target.value })}
              className="input w-full"
              rows={2}
            />
          </div>
          <div className="flex gap-3 pt-4">
            <button type="button" onClick={onClose} className="btn btn-secondary flex-1">Cancel</button>
            <button type="submit" disabled={saving} className="btn btn-primary flex-1">
              {saving ? 'Saving...' : 'Create'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
