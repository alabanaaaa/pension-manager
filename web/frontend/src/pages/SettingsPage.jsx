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
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight text-black">Settings</h1>
        <p className="text-sm text-gray-500 mt-1">Manage your account and system preferences</p>
      </div>

      {saved && (
        <div className="bg-black text-white rounded-lg p-4 flex items-center gap-3">
          <CheckCircle size={18} />
          <p className="text-sm">Settings saved successfully</p>
        </div>
      )}

      {/* Profile */}
      <div className="card">
        <div className="card-header flex items-center gap-3">
          <User size={18} className="text-gray-500" />
          <h2 className="text-base font-semibold text-black">Profile</h2>
        </div>
        <div className="card-body space-y-4">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div className="form-group mb-0">
              <label className="label">Full Name</label>
              <input type="text" defaultValue={user?.name} className="input" />
            </div>
            <div className="form-group mb-0">
              <label className="label">Email</label>
              <input type="email" defaultValue={user?.email} className="input" />
            </div>
            <div className="form-group mb-0">
              <label className="label">Role</label>
              <input type="text" defaultValue={user?.role} disabled className="input" />
            </div>
          </div>
          <button onClick={handleSave} className="btn btn-primary">
            Save Changes
          </button>
        </div>
      </div>

      {/* Notifications */}
      <div className="card">
        <div className="card-header flex items-center gap-3">
          <Bell size={18} className="text-gray-500" />
          <h2 className="text-base font-semibold text-black">Notifications</h2>
        </div>
        <div className="card-body space-y-0">
          {[
            { label: 'Email notifications for pending approvals', defaultChecked: true },
            { label: 'SMS alerts for claim status changes', defaultChecked: true },
            { label: 'Daily contribution summary', defaultChecked: false },
            { label: 'Weekly security report', defaultChecked: true },
          ].map((item, i) => (
            <label key={i} className="flex items-center justify-between py-3 border-b border-gray-100 last:border-0 cursor-pointer">
              <span className="text-sm text-black">{item.label}</span>
              <input type="checkbox" defaultChecked={item.defaultChecked} className="w-4 h-4 rounded border-gray-300 text-black focus:ring-black" />
            </label>
          ))}
        </div>
      </div>

      {/* Security */}
      <div className="card">
        <div className="card-header flex items-center gap-3">
          <Shield size={18} className="text-gray-500" />
          <h2 className="text-base font-semibold text-black">Security</h2>
        </div>
        <div className="card-body space-y-4">
          <div className="form-group mb-0">
            <label className="label">Current Password</label>
            <input type="password" className="input" placeholder="Enter current password" />
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div className="form-group mb-0">
              <label className="label">New Password</label>
              <input type="password" className="input" placeholder="Enter new password" />
            </div>
            <div className="form-group mb-0">
              <label className="label">Confirm New Password</label>
              <input type="password" className="input" placeholder="Confirm new password" />
            </div>
          </div>
          <button onClick={handleSave} className="btn btn-primary">
            Update Password
          </button>
        </div>
      </div>

      {/* System */}
      <div className="card">
        <div className="card-header flex items-center gap-3">
          <Database size={18} className="text-gray-500" />
          <h2 className="text-base font-semibold text-black">System</h2>
        </div>
        <div className="card-body space-y-0">
          {[
            { label: 'Database Connection', value: 'Connected', status: 'ok' },
            { label: 'M-Pesa Integration', value: 'Sandbox', status: 'ok' },
            { label: 'SMS Gateway', value: 'Mock Mode', status: 'ok' },
            { label: 'News API', value: 'Mock Mode', status: 'ok' },
          ].map((item, i) => (
            <div key={i} className="flex items-center justify-between py-3 border-b border-gray-100 last:border-0">
              <span className="text-sm text-gray-600">{item.label}</span>
              <div className="flex items-center gap-2">
                <div className={`w-2 h-2 rounded-full ${item.status === 'ok' ? 'bg-green-600' : 'bg-red-600'}`} />
                <span className="text-sm text-gray-500">{item.value}</span>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
