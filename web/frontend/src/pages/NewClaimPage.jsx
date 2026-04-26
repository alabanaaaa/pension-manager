import { useState, useEffect } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { claims, members } from '../lib/api';
import { ArrowLeft, Save, Loader2, FileText, Calendar, User, DollarSign } from 'lucide-react';

export default function NewClaimPage() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [membersList, setMembersList] = useState([]);
  const [form, setForm] = useState({
    member_id: '',
    claim_type: 'medical_claim',
    date_of_claim: new Date().toISOString().split('T')[0],
    date_of_leaving: '',
    leaving_reason: '',
    amount: 0,
    notes: '',
  });

  useEffect(() => {
    members.list({ limit: 100 })
      .then(res => setMembersList(Array.isArray(res.data) ? res.data : []))
      .catch(() => setMembersList([]));
  }, []);

  const handleChange = (e) => {
    const { name, value, type } = e.target;
    setForm(prev => ({
      ...prev,
      [name]: type === 'number' ? parseInt(value) || 0 : value
    }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    
    try {
      const res = await claims.create(form);
      navigate(`/claims/${res.data.id}`);
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to create claim');
    } finally {
      setLoading(false);
    }
  };

  const claimTypes = [
    { value: 'normal_retirement', label: 'Normal Retirement' },
    { value: 'early_retirement', label: 'Early Retirement' },
    { value: 'late_retirement', label: 'Late Retirement' },
    { value: 'ill_health_retirement', label: 'Ill Health Retirement' },
    { value: 'death_in_service', label: 'Death in Service' },
    { value: 'leaving_service', label: 'Leaving Service' },
    { value: 'deferred_retirement', label: 'Deferred Benefit' },
    { value: 'medical_claim', label: 'Medical Claim' },
    { value: 'ex_gratia', label: 'Ex-Gratia' },
  ];

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Link to="/claims" className="btn-hover p-2 rounded-xl hover:bg-neutral-100">
          <ArrowLeft size={20} className="text-neutral-600" />
        </Link>
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-black">New Claim</h1>
          <p className="text-neutral-500 mt-1">Submit a new benefit claim</p>
        </div>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-100 rounded-2xl p-4 text-red-600 text-sm">
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-8">
        {/* Claim Details */}
        <div className="bg-white rounded-2xl border border-neutral-100 overflow-hidden">
          <div className="px-6 py-4 border-b border-neutral-50 bg-neutral-50/50">
            <div className="flex items-center gap-2">
              <FileText size={18} className="text-neutral-500" />
              <h2 className="text-lg font-medium text-neutral-900">Claim Details</h2>
            </div>
          </div>
          <div className="p-6 grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <label className="block text-sm font-medium text-neutral-700 mb-1.5">Member *</label>
              <select
                name="member_id"
                value={form.member_id}
                onChange={handleChange}
                required
                className="w-full px-4 py-2.5 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10"
              >
                <option value="">Select member</option>
                {membersList.map(m => (
                  <option key={m.id} value={m.id}>
                    {m.member_no} - {m.first_name} {m.last_name}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-neutral-700 mb-1.5">Claim Type *</label>
              <select
                name="claim_type"
                value={form.claim_type}
                onChange={handleChange}
                required
                className="w-full px-4 py-2.5 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10"
              >
                {claimTypes.map(t => (
                  <option key={t.value} value={t.value}>{t.label}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-neutral-700 mb-1.5">Date of Claim *</label>
              <input
                type="date"
                name="date_of_claim"
                value={form.date_of_claim}
                onChange={handleChange}
                required
                className="w-full px-4 py-2.5 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-neutral-700 mb-1.5">Amount (KES)</label>
              <input
                type="number"
                name="amount"
                value={form.amount}
                onChange={handleChange}
                className="w-full px-4 py-2.5 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10"
              />
            </div>
            {['normal_retirement', 'early_retirement', 'leaving_service', 'deferred_retirement'].includes(form.claim_type) && (
              <>
                <div>
                  <label className="block text-sm font-medium text-neutral-700 mb-1.5">Date of Leaving</label>
                  <input
                    type="date"
                    name="date_of_leaving"
                    value={form.date_of_leaving}
                    onChange={handleChange}
                    className="w-full px-4 py-2.5 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-neutral-700 mb-1.5">Leaving Reason</label>
                  <input
                    type="text"
                    name="leaving_reason"
                    value={form.leaving_reason}
                    onChange={handleChange}
                    className="w-full px-4 py-2.5 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10"
                  />
                </div>
              </>
            )}
            <div className="md:col-span-2">
              <label className="block text-sm font-medium text-neutral-700 mb-1.5">Notes</label>
              <textarea
                name="notes"
                value={form.notes}
                onChange={handleChange}
                rows={3}
                className="w-full px-4 py-2.5 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10"
              />
            </div>
          </div>
        </div>

        {/* Submit */}
        <div className="flex justify-end gap-3">
          <Link to="/claims" className="px-6 py-2.5 border border-neutral-200 rounded-xl text-sm font-medium hover:bg-neutral-50">
            Cancel
          </Link>
          <button
            type="submit"
            disabled={loading}
            className="flex items-center gap-2 px-6 py-2.5 bg-neutral-900 text-white rounded-xl text-sm font-medium hover:bg-neutral-800 disabled:opacity-50"
          >
            {loading ? <Loader2 size={16} className="animate-spin" /> : <Save size={16} />}
            Submit Claim
          </button>
        </div>
      </form>
    </div>
  );
}
