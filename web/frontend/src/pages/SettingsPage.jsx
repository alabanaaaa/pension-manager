import { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import { Loader2, User, Bell, Shield, Database, Settings as SettingsIcon, CheckCircle } from 'lucide-react';

export default function SettingsPage() {
  const { user } = useAuth();
  const [saved, setSaved] = useState(false);

  const handleSave = () => {
    setSaved(true);
    setTimeout(() => setSaved(false), 3000);
  };

  return (
    <div className="space-y-8 animate-fade-in-up">
      <div>
        <h1 className="text-3xl font-semibold tracking-tight text-neutral-900">Settings</h1>
        <p className="text-neutral-500 mt-2 text-base">Manage your account and system preferences</p>
      </div>

      {saved && (
        <div className="bg-emerald-50 border border-emerald-100 rounded-2xl p-5 flex items-center gap-3">
          <CheckCircle size={18} className="text-emerald-600" />
          <p className="text-sm text-emerald-700">Settings saved successfully</p>
        </div>
      )}

      {/* Profile */}
      <div className="bg-white rounded-2xl border border-neutral-50 overflow-hidden">
        <div className="flex items-center justify-between px-6 py-5 border-b border-neutral-50">
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-blue-50"><User size={20} className="text-blue-600" /></div>
            <h2 className="text-lg font-semibold tracking-tight text-neutral-900">Profile</h2>
          </div>
        </div>
        <div className="p-6 space-y-6">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
            <div>
              <label className="block text-sm font-medium text-neutral-700 mb-2">Full Name</label>
              <input type="text" defaultValue={user?.name} className="w-full px-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all" />
            </div>
            <div>
              <label className="block text-sm font-medium text-neutral-700 mb-2">Email</label>
              <input type="email" defaultValue={user?.email} className="w-full px-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all" />
            </div>
            <div>
              <label className="block text-sm font-medium text-neutral-700 mb-2">Role</label>
              <input type="text" defaultValue={user?.role} disabled className="w-full px-4 py-3.5 bg-neutral-100 border border-neutral-200 rounded-xl text-sm text-neutral-500" />
            </div>
          </div>
          <button onClick={handleSave} className="btn-hover px-6 py-3.5 bg-neutral-900 text-white rounded-xl text-sm font-medium hover:bg-neutral-800 transition-all">
            Save Changes
          </button>
        </div>
      </div>

      {/* Notifications */}
      <div className="bg-white rounded-2xl border border-neutral-50 overflow-hidden">
        <div className="flex items-center justify-between px-6 py-5 border-b border-neutral-50">
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-amber-50"><Bell size={20} className="text-amber-600" /></div>
            <h2 className="text-lg font-semibold tracking-tight text-neutral-900">Notifications</h2>
          </div>
        </div>
        <div className="p-6 space-y-4">
          {[
            { label: 'Email notifications for pending approvals', defaultChecked: true },
            { label: 'SMS alerts for claim status changes', defaultChecked: true },
            { label: 'Daily contribution summary', defaultChecked: false },
            { label: 'Weekly security report', defaultChecked: true },
          ].map((item, i) => (
            <label key={i} className="flex items-center justify-between py-3 border-b border-neutral-50 last:border-0 cursor-pointer">
              <span className="text-sm text-neutral-700">{item.label}</span>
              <input type="checkbox" defaultChecked={item.defaultChecked} className="w-4 h-4 rounded border-neutral-300 text-neutral-900 focus:ring-neutral-900/20" />
            </label>
          ))}
        </div>
      </div>

      {/* Security */}
      <div className="bg-white rounded-2xl border border-neutral-50 overflow-hidden">
        <div className="flex items-center justify-between px-6 py-5 border-b border-neutral-50">
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-red-50"><Shield size={20} className="text-red-600" /></div>
            <h2 className="text-lg font-semibold tracking-tight text-neutral-900">Security</h2>
          </div>
        </div>
        <div className="p-6 space-y-6">
          <div>
            <label className="block text-sm font-medium text-neutral-700 mb-2">Current Password</label>
            <input type="password" className="w-full px-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all" />
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
            <div>
              <label className="block text-sm font-medium text-neutral-700 mb-2">New Password</label>
              <input type="password" className="w-full px-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all" />
            </div>
            <div>
              <label className="block text-sm font-medium text-neutral-700 mb-2">Confirm New Password</label>
              <input type="password" className="w-full px-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all" />
            </div>
          </div>
          <button onClick={handleSave} className="btn-hover px-6 py-3.5 bg-neutral-900 text-white rounded-xl text-sm font-medium hover:bg-neutral-800 transition-all">
            Update Password
          </button>
        </div>
      </div>

      {/* System */}
      <div className="bg-white rounded-2xl border border-neutral-50 overflow-hidden">
        <div className="flex items-center justify-between px-6 py-5 border-b border-neutral-50">
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-violet-50"><Database size={20} className="text-violet-600" /></div>
            <h2 className="text-lg font-semibold tracking-tight text-neutral-900">System</h2>
          </div>
        </div>
        <div className="p-6 space-y-4">
          {[
            { label: 'Database Connection', value: 'Connected', status: 'ok' },
            { label: 'M-Pesa Integration', value: 'Sandbox', status: 'ok' },
            { label: 'SMS Gateway', value: 'Mock Mode', status: 'ok' },
            { label: 'News API', value: 'Mock Mode', status: 'ok' },
          ].map((item, i) => (
            <div key={i} className="flex items-center justify-between py-3 border-b border-neutral-50 last:border-0">
              <span className="text-sm text-neutral-600">{item.label}</span>
              <div className="flex items-center gap-2">
                <div className={`w-2 h-2 rounded-full ${item.status === 'ok' ? 'bg-emerald-500' : 'bg-red-500'}`} />
                <span className="text-sm text-neutral-500">{item.value}</span>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
