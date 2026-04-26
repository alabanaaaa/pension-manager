import { useState, useEffect } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { members } from '../lib/api';
import { ArrowLeft, Save, Loader2, User, Mail, Phone, Calendar, CreditCard, FileText, Shield } from 'lucide-react';

export default function EditMemberPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [form, setForm] = useState({
    first_name: '', last_name: '', email: '', phone: '', id_number: '',
    member_no: '', department: '', membership_status: 'active',
    date_joined_scheme: '', account_balance: 0
  });

  useEffect(() => {
    if (!id) { setLoading(false); return; }
    members.get(id)
      .then(res => {
        const m = res.data;
        setForm({
          first_name: m.first_name || '',
          last_name: m.last_name || '',
          email: m.email || '',
          phone: m.phone || '',
          id_number: m.id_number || '',
          member_no: m.member_no || '',
          department: m.department || '',
          membership_status: m.membership_status || 'active',
          date_joined_scheme: m.date_joined_scheme?.split('T')[0] || '',
          account_balance: m.account_balance || 0
        });
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [id]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setSaving(true);
    try {
      await members.update(id, form);
      navigate(`/members/${id}`);
    } catch (err) { console.error(err); }
    finally { setSaving(false); }
  };

  if (loading) return <div className="p-8 text-center"><Loader2 className="animate-spin mx-auto" /></div>;

  return (
    <div className="space-y-8">
      <div className="flex items-center gap-4">
        <Link to={`/members/${id}`} className="p-2 hover:bg-neutral-100 rounded-lg"><ArrowLeft size={20} /></Link>
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-black">Edit Member</h1>
          <p className="text-neutral-500 mt-1">Update member information</p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="bg-white rounded-2xl border border-neutral-100 p-6 space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div>
            <label className="block text-sm font-medium text-neutral-700 mb-1">First Name</label>
            <input type="text" value={form.first_name} onChange={e => setForm({...form, first_name: e.target.value})}
              className="w-full px-4 py-2.5 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10" />
          </div>
          <div>
            <label className="block text-sm font-medium text-neutral-700 mb-1">Last Name</label>
            <input type="text" value={form.last_name} onChange={e => setForm({...form, last_name: e.target.value})}
              className="w-full px-4 py-2.5 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10" />
          </div>
          <div>
            <label className="block text-sm font-medium text-neutral-700 mb-1">Email</label>
            <input type="email" value={form.email} onChange={e => setForm({...form, email: e.target.value})}
              className="w-full px-4 py-2.5 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10" />
          </div>
          <div>
            <label className="block text-sm font-medium text-neutral-700 mb-1">Phone</label>
            <input type="tel" value={form.phone} onChange={e => setForm({...form, phone: e.target.value})}
              className="w-full px-4 py-2.5 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10" />
          </div>
          <div>
            <label className="block text-sm font-medium text-neutral-700 mb-1">ID Number</label>
            <input type="text" value={form.id_number} onChange={e => setForm({...form, id_number: e.target.value})}
              className="w-full px-4 py-2.5 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10" />
          </div>
          <div>
            <label className="block text-sm font-medium text-neutral-700 mb-1">Member No</label>
            <input type="text" value={form.member_no} onChange={e => setForm({...form, member_no: e.target.value})}
              className="w-full px-4 py-2.5 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10" />
          </div>
          <div>
            <label className="block text-sm font-medium text-neutral-700 mb-1">Department</label>
            <input type="text" value={form.department} onChange={e => setForm({...form, department: e.target.value})}
              className="w-full px-4 py-2.5 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10" />
          </div>
          <div>
            <label className="block text-sm font-medium text-neutral-700 mb-1">Status</label>
            <select value={form.membership_status} onChange={e => setForm({...form, membership_status: e.target.value})}
              className="w-full px-4 py-2.5 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10">
              <option value="active">Active</option>
              <option value="retired">Retired</option>
              <option value="deceased">Deceased</option>
              <option value="deferred">Deferred</option>
              <option value="withdrawn">Withdrawn</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-neutral-700 mb-1">Date Joined</label>
            <input type="date" value={form.date_joined_scheme} onChange={e => setForm({...form, date_joined_scheme: e.target.value})}
              className="w-full px-4 py-2.5 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10" />
          </div>
          <div>
            <label className="block text-sm font-medium text-neutral-700 mb-1">Account Balance</label>
            <input type="number" value={form.account_balance} onChange={e => setForm({...form, account_balance: e.target.value})}
              className="w-full px-4 py-2.5 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10" />
          </div>
        </div>

        <div className="flex justify-end gap-3 pt-4">
          <Link to={`/members/${id}`} className="px-5 py-2.5 border border-neutral-200 rounded-xl text-sm font-medium hover:bg-neutral-50">Cancel</Link>
          <button type="submit" disabled={saving} className="px-5 py-2.5 bg-neutral-900 text-white rounded-xl text-sm font-medium hover:bg-neutral-800 flex items-center gap-2">
            {saving ? <Loader2 className="animate-spin w-4 h-4" /> : <Save size={16} />}
            Save Changes
          </button>
        </div>
      </form>
    </div>
  );
}
