import { useState, useEffect } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { contributions, members, sponsor } from '../lib/api';
import { ArrowLeft, Save, Loader2, CreditCard, Calendar, User } from 'lucide-react';

export default function RecordContributionPage() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [membersList, setMembersList] = useState([]);
  const [sponsorsList, setSponsorsList] = useState([]);
  const [selectedMember, setSelectedMember] = useState(null);
  const [form, setForm] = useState({
    member_id: '',
    sponsor_id: '',
    period: new Date().toISOString().slice(0, 7),
    employee_amount: 0,
    employer_amount: 0,
    avc_amount: 0,
    payment_method: 'mpesa',
    payment_ref: '',
  });

  useEffect(() => {
    Promise.all([
      members.list({ limit: 200, status: 'active' }),
      sponsor.list(),
    ]).then(([mRes, sRes]) => {
      setMembersList(Array.isArray(mRes.data) ? mRes.data : []);
      setSponsorsList(Array.isArray(sRes.data) ? sRes.data : []);
    }).catch(() => {});
  }, []);

  useEffect(() => {
    if (form.member_id && membersList.length > 0) {
      const member = membersList.find(m => m.id === form.member_id);
      setSelectedMember(member);
      if (member?.sponsor_id) {
        setForm(prev => ({ ...prev, sponsor_id: member.sponsor_id }));
      }
    }
  }, [form.member_id, membersList]);

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
      const res = await contributions.create(form);
      navigate('/contributions');
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to record contribution');
    } finally {
      setLoading(false);
    }
  };

  const totalAmount = form.employee_amount + form.employer_amount + form.avc_amount;

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Link to="/contributions" className="btn-hover p-2 rounded-xl hover:bg-neutral-100">
          <ArrowLeft size={20} className="text-neutral-600" />
        </Link>
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-black">Record Contribution</h1>
          <p className="text-neutral-500 mt-1">Log a new member contribution</p>
        </div>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-100 rounded-2xl p-4 text-red-600 text-sm">
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-8">
        {/* Member Selection */}
        <div className="bg-white rounded-2xl border border-neutral-100 overflow-hidden">
          <div className="px-6 py-4 border-b border-neutral-50 bg-neutral-50/50">
            <div className="flex items-center gap-2">
              <User size={18} className="text-neutral-500" />
              <h2 className="text-lg font-medium text-neutral-900">Member & Period</h2>
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
              <label className="block text-sm font-medium text-neutral-700 mb-1.5">Sponsor</label>
              <select
                name="sponsor_id"
                value={form.sponsor_id}
                onChange={handleChange}
                className="w-full px-4 py-2.5 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10"
              >
                <option value="">Select sponsor</option>
                {sponsorsList.map(s => (
                  <option key={s.id} value={s.id}>{s.name}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-neutral-700 mb-1.5">Period (Month) *</label>
              <input
                type="month"
                name="period"
                value={form.period}
                onChange={handleChange}
                required
                className="w-full px-4 py-2.5 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-neutral-700 mb-1.5">Payment Method</label>
              <select
                name="payment_method"
                value={form.payment_method}
                onChange={handleChange}
                className="w-full px-4 py-2.5 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10"
              >
                <option value="mpesa">M-Pesa</option>
                <option value="bank_transfer">Bank Transfer</option>
                <option value="cheque">Cheque</option>
                <option value="cash">Cash</option>
                <option value="standing_order">Standing Order</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-neutral-700 mb-1.5">Payment Reference</label>
              <input
                type="text"
                name="payment_ref"
                value={form.payment_ref}
                onChange={handleChange}
                className="w-full px-4 py-2.5 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10"
                placeholder="Transaction ID, Cheque No, etc."
              />
            </div>
          </div>
        </div>

        {/* Contribution Amounts */}
        <div className="bg-white rounded-2xl border border-neutral-100 overflow-hidden">
          <div className="px-6 py-4 border-b border-neutral-50 bg-neutral-50/50">
            <div className="flex items-center gap-2">
              <CreditCard size={18} className="text-neutral-500" />
              <h2 className="text-lg font-medium text-neutral-900">Contribution Amounts</h2>
            </div>
          </div>
          <div className="p-6 grid grid-cols-1 md:grid-cols-3 gap-6">
            <div>
              <label className="block text-sm font-medium text-neutral-700 mb-1.5">Employee Contribution (KES)</label>
              <input
                type="number"
                name="employee_amount"
                value={form.employee_amount}
                onChange={handleChange}
                min={0}
                className="w-full px-4 py-2.5 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-neutral-700 mb-1.5">Employer Contribution (KES)</label>
              <input
                type="number"
                name="employer_amount"
                value={form.employer_amount}
                onChange={handleChange}
                min={0}
                className="w-full px-4 py-2.5 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-neutral-700 mb-1.5">AVC (KES)</label>
              <input
                type="number"
                name="avc_amount"
                value={form.avc_amount}
                onChange={handleChange}
                min={0}
                className="w-full px-4 py-2.5 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10"
              />
            </div>
          </div>
          <div className="px-6 pb-6">
            <div className="bg-neutral-50 rounded-xl p-4 flex justify-between items-center">
              <span className="text-neutral-600">Total Contribution</span>
              <span className="text-2xl font-semibold text-neutral-900">KES {totalAmount.toLocaleString()}</span>
            </div>
          </div>
        </div>

        {/* Member Info Preview */}
        {selectedMember && (
          <div className="bg-white rounded-2xl border border-neutral-100 overflow-hidden">
            <div className="px-6 py-4 border-b border-neutral-50 bg-neutral-50/50">
              <h2 className="text-lg font-medium text-neutral-900">Member Details</h2>
            </div>
            <div className="p-6 grid grid-cols-2 md:grid-cols-4 gap-4">
              <div>
                <p className="text-sm text-neutral-500">Member No</p>
                <p className="font-medium text-neutral-900">{selectedMember.member_no}</p>
              </div>
              <div>
                <p className="text-sm text-neutral-500">Name</p>
                <p className="font-medium text-neutral-900">{selectedMember.first_name} {selectedMember.last_name}</p>
              </div>
              <div>
                <p className="text-sm text-neutral-500">Department</p>
                <p className="font-medium text-neutral-900">{selectedMember.department || '-'}</p>
              </div>
              <div>
                <p className="text-sm text-neutral-500">Current Balance</p>
                <p className="font-medium text-neutral-900">KES {Number(selectedMember.account_balance || 0).toLocaleString()}</p>
              </div>
            </div>
          </div>
        )}

        {/* Submit */}
        <div className="flex justify-end gap-3">
          <Link to="/contributions" className="px-6 py-2.5 border border-neutral-200 rounded-xl text-sm font-medium hover:bg-neutral-50">
            Cancel
          </Link>
          <button
            type="submit"
            disabled={loading}
            className="flex items-center gap-2 px-6 py-2.5 bg-neutral-900 text-white rounded-xl text-sm font-medium hover:bg-neutral-800 disabled:opacity-50"
          >
            {loading ? <Loader2 size={16} className="animate-spin" /> : <Save size={16} />}
            Record Contribution
          </button>
        </div>
      </form>
    </div>
  );
}
