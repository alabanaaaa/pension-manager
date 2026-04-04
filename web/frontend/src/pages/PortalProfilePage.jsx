import { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import { portal } from '../lib/api';
import { Loader2, User, Mail, Phone, MapPin, Calendar, Building2, CreditCard, Save, CheckCircle } from 'lucide-react';

export default function PortalProfilePage() {
  const { user } = useAuth();
  const [profile, setProfile] = useState(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    portal.getProfile()
      .then(r => setProfile(r.data || null))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  const handleSave = async () => {
    setSaving(true);
    try {
      await portal.requestContactChange({
        email: profile?.contact_info?.email,
        mobile_number: profile?.contact_info?.mobile_number,
        postal_address: profile?.contact_info?.postal_address,
        postal_code: profile?.contact_info?.postal_code,
        town: profile?.contact_info?.town,
      });
      setSaved(true);
      setTimeout(() => setSaved(false), 3000);
    } catch (err) {
      console.error(err);
    }
    finally { setSaving(false); }
  };

  if (loading) {
    return (
      <div className="p-16 text-center">
        <Loader2 size={24} className="animate-spin mx-auto text-neutral-300" />
        <p className="text-sm text-neutral-400 mt-3">Loading profile...</p>
      </div>
    );
  }

  const personal = profile?.personal_info || {};
  const contact = profile?.contact_info || {};
  const employment = profile?.employment_info || {};
  const medical = profile?.medical_limits || {};
  const account = profile?.account_summary || {};

  return (
    <div className="space-y-8 animate-fade-in-up">
      <div>
        <h1 className="text-3xl font-semibold tracking-tight text-neutral-900">My Profile</h1>
        <p className="text-neutral-500 mt-2 text-base">View and update your personal information</p>
      </div>

      {saved && (
        <div className="bg-emerald-50 border border-emerald-100 rounded-2xl p-5 flex items-center gap-3">
          <CheckCircle size={18} className="text-emerald-600" />
          <p className="text-sm text-emerald-700">Update request submitted for approval</p>
        </div>
      )}

      {/* Personal Info */}
      <div className="bg-white rounded-2xl border border-neutral-50 overflow-hidden">
        <div className="flex items-center justify-between px-6 py-5 border-b border-neutral-50">
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-blue-50"><User size={20} className="text-blue-600" /></div>
            <h2 className="text-lg font-semibold tracking-tight text-neutral-900">Personal Information</h2>
          </div>
        </div>
        <div className="p-6 grid grid-cols-1 sm:grid-cols-2 gap-6">
          {[
            { label: 'Full Name', value: personal.full_name },
            { label: 'National ID', value: personal.national_id },
            { label: 'Gender', value: personal.gender },
            { label: 'Date of Birth', value: personal.date_of_birth ? new Date(personal.date_of_birth).toLocaleDateString() : '—' },
            { label: 'Age', value: personal.age ? `${personal.age} years` : '—' },
            { label: 'Nationality', value: personal.nationality },
            { label: 'KRA PIN', value: personal.kra_pin },
            { label: 'Marital Status', value: personal.marital_status },
            { label: 'Spouse Name', value: personal.spouse_name || '—' },
          ].map((item, i) => (
            <div key={i}>
              <p className="text-xs text-neutral-400 uppercase tracking-wider mb-1">{item.label}</p>
              <p className="text-sm font-medium text-neutral-900">{item.value || '—'}</p>
            </div>
          ))}
        </div>
      </div>

      {/* Contact Info (editable) */}
      <div className="bg-white rounded-2xl border border-neutral-50 overflow-hidden">
        <div className="flex items-center justify-between px-6 py-5 border-b border-neutral-50">
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-emerald-50"><Mail size={20} className="text-emerald-600" /></div>
            <h2 className="text-lg font-semibold tracking-tight text-neutral-900">Contact Details</h2>
          </div>
        </div>
        <div className="p-6 space-y-6">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
            <div>
              <label className="block text-xs text-neutral-400 uppercase tracking-wider mb-2">Email</label>
              <input
                type="email"
                value={contact.email || ''}
                onChange={e => setProfile(prev => ({ ...prev, contact_info: { ...prev.contact_info, email: e.target.value } }))}
                className="w-full px-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all"
              />
            </div>
            <div>
              <label className="block text-xs text-neutral-400 uppercase tracking-wider mb-2">Phone</label>
              <input
                type="tel"
                value={contact.mobile_number || ''}
                onChange={e => setProfile(prev => ({ ...prev, contact_info: { ...prev.contact_info, mobile_number: e.target.value } }))}
                className="w-full px-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all"
              />
            </div>
            <div>
              <label className="block text-xs text-neutral-400 uppercase tracking-wider mb-2">Postal Address</label>
              <input
                type="text"
                value={contact.postal_address || ''}
                onChange={e => setProfile(prev => ({ ...prev, contact_info: { ...prev.contact_info, postal_address: e.target.value } }))}
                className="w-full px-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all"
              />
            </div>
            <div>
              <label className="block text-xs text-neutral-400 uppercase tracking-wider mb-2">Town/City</label>
              <input
                type="text"
                value={contact.town || ''}
                onChange={e => setProfile(prev => ({ ...prev, contact_info: { ...prev.contact_info, town: e.target.value } }))}
                className="w-full px-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all"
              />
            </div>
          </div>
          <button
            onClick={handleSave}
            disabled={saving}
            className="btn-hover flex items-center gap-2 px-6 py-3.5 bg-neutral-900 text-white rounded-xl text-sm font-medium hover:bg-neutral-800 disabled:opacity-50 transition-all"
          >
            {saving ? <Loader2 size={16} className="animate-spin" /> : <Save size={16} />}
            Request Update
          </button>
        </div>
      </div>

      {/* Employment */}
      <div className="bg-white rounded-2xl border border-neutral-50 overflow-hidden">
        <div className="flex items-center justify-between px-6 py-5 border-b border-neutral-50">
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-violet-50"><Building2 size={20} className="text-violet-600" /></div>
            <h2 className="text-lg font-semibold tracking-tight text-neutral-900">Employment Details</h2>
          </div>
        </div>
        <div className="p-6 grid grid-cols-1 sm:grid-cols-2 gap-6">
          {[
            { label: 'Member No', value: employment.member_no },
            { label: 'Sponsor', value: employment.sponsor_name },
            { label: 'Department', value: employment.department },
            { label: 'Designation', value: employment.designation },
            { label: 'Date Joined', value: employment.date_joined_scheme ? new Date(employment.date_joined_scheme).toLocaleDateString() : '—' },
            { label: 'Basic Salary', value: employment.basic_salary ? `KES ${(employment.basic_salary / 100).toLocaleString()}` : '—' },
            { label: 'Bank', value: employment.bank_name },
            { label: 'Account', value: employment.bank_account },
          ].map((item, i) => (
            <div key={i}>
              <p className="text-xs text-neutral-400 uppercase tracking-wider mb-1">{item.label}</p>
              <p className="text-sm font-medium text-neutral-900">{item.value || '—'}</p>
            </div>
          ))}
        </div>
      </div>

      {/* Medical Limits */}
      <div className="bg-white rounded-2xl border border-neutral-50 overflow-hidden">
        <div className="flex items-center justify-between px-6 py-5 border-b border-neutral-50">
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-amber-50"><CreditCard size={20} className="text-amber-600" /></div>
            <h2 className="text-lg font-semibold tracking-tight text-neutral-900">Medical Limits</h2>
          </div>
        </div>
        <div className="p-6 grid grid-cols-1 sm:grid-cols-2 gap-6">
          <div>
            <p className="text-xs text-neutral-400 uppercase tracking-wider mb-1">Inpatient Limit</p>
            <p className="text-sm font-medium text-neutral-900">KES {((medical.inpatient_limit || 0) / 100).toLocaleString()}</p>
          </div>
          <div>
            <p className="text-xs text-neutral-400 uppercase tracking-wider mb-1">Outpatient Limit</p>
            <p className="text-sm font-medium text-neutral-900">KES {((medical.outpatient_limit || 0) / 100).toLocaleString()}</p>
          </div>
        </div>
      </div>
    </div>
  );
}
