import { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import { portal, contributions, claims, voting } from '../lib/api';
import {
  User, CreditCard, FileText, Vote, TrendingUp,
  MessageSquare, Settings, Calendar, Clock, ArrowUpRight,
  ArrowDownRight, Activity, Bell, ChevronRight, Loader2
} from 'lucide-react';
import { Link } from 'react-router-dom';

export default function PortalDashboardPage() {
  const { user } = useAuth();
  const [profile, setProfile] = useState(null);
  const [contribCount, setContribCount] = useState(0);
  const [claimCount, setClaimCount] = useState(0);
  const [elections, setElections] = useState([]);
  const [loginStats, setLoginStats] = useState(null);
  const [loading, setLoading] = useState(true);

  const getGreeting = () => {
    const hour = new Date().getHours();
    const name = user?.name?.split(' ')[0] || 'Member';
    if (hour < 12) return { text: `Good morning, ${name}`, emoji: '☀️' };
    if (hour < 17) return { text: `Good afternoon, ${name}`, emoji: '🌤️' };
    return { text: `Good evening, ${name}`, emoji: '🌙' };
  };

  useEffect(() => {
    Promise.all([
      portal.getProfile().then(r => setProfile(r.data || null)).catch(() => {}),
      contributions.list().then(r => {
        const data = Array.isArray(r.data) ? r.data : [];
        setContribCount(data.length);
      }).catch(() => {}),
      claims.list().then(r => {
        const data = Array.isArray(r.data) ? r.data : [];
        setClaimCount(data.filter(c => c.status !== 'paid').length);
      }).catch(() => {}),
      voting.memberListElections().then(r => {
        const data = Array.isArray(r.data) ? r.data : [];
        setElections(data.filter(e => e.status === 'open'));
      }).catch(() => {}),
      portal.getLoginStats().then(r => setLoginStats(r.data || null)).catch(() => {}),
    ]).finally(() => setLoading(false));
  }, []);

  const greeting = getGreeting();

  const quickLinks = [
    { icon: User, label: 'My Profile', desc: 'View and update personal info', href: '/portal/profile', color: 'blue' },
    { icon: CreditCard, label: 'Contributions', desc: 'View contribution history', href: '/portal/contributions', color: 'emerald' },
    { icon: FileText, label: 'My Claims', desc: 'Check claim status', href: '/portal/claims', color: 'amber' },
    { icon: Vote, label: 'Vote', desc: 'Cast your vote', href: '/portal/voting', color: 'violet' },
    { icon: TrendingUp, label: 'Projections', desc: 'Benefit projections', href: '/portal/projections', color: 'sky' },
    { icon: MessageSquare, label: 'Feedback', desc: 'Submit feedback', href: '/portal/feedback', color: 'rose' },
  ];

  const colorMap = {
    blue: { bg: 'bg-blue-50', icon: 'text-blue-600' },
    emerald: { bg: 'bg-emerald-50', icon: 'text-emerald-600' },
    amber: { bg: 'bg-amber-50', icon: 'text-amber-600' },
    violet: { bg: 'bg-violet-50', icon: 'text-violet-600' },
    sky: { bg: 'bg-sky-50', icon: 'text-sky-600' },
    rose: { bg: 'bg-rose-50', icon: 'text-rose-600' },
  };

  if (loading) {
    return (
      <div className="p-16 text-center">
        <Loader2 size={24} className="animate-spin mx-auto text-neutral-300" />
        <p className="text-sm text-neutral-400 mt-3">Loading your dashboard...</p>
      </div>
    );
  }

  return (
    <div className="space-y-8 animate-fade-in-up">
      {/* Greeting */}
      <div className="flex items-end justify-between">
        <div>
          <p className="text-sm text-neutral-400 mb-1">{greeting.emoji} {greeting.text}</p>
          <h1 className="text-3xl font-semibold tracking-tight text-neutral-900">My Dashboard</h1>
          <p className="text-neutral-500 mt-2 text-base">Welcome to your member portal</p>
        </div>
        <div className="flex items-center gap-2 text-sm text-neutral-400">
          <Calendar size={14} />
          <span>{new Date().toLocaleDateString('en-US', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' })}</span>
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
        <div className="bg-white rounded-2xl border border-neutral-50 p-6">
          <div className="flex items-start justify-between mb-4">
            <div className="p-2.5 rounded-xl bg-blue-50"><User size={20} className="text-blue-600" /></div>
          </div>
          <p className="text-sm text-neutral-500">Member No</p>
          <p className="text-xl font-semibold tracking-tight text-neutral-900 mt-1">{profile?.member_no || '—'}</p>
        </div>
        <div className="bg-white rounded-2xl border border-neutral-50 p-6">
          <div className="flex items-start justify-between mb-4">
            <div className="p-2.5 rounded-xl bg-emerald-50"><CreditCard size={20} className="text-emerald-600" /></div>
          </div>
          <p className="text-sm text-neutral-500">Contributions</p>
          <p className="text-xl font-semibold tracking-tight text-neutral-900 mt-1">{contribCount}</p>
        </div>
        <div className="bg-white rounded-2xl border border-neutral-50 p-6">
          <div className="flex items-start justify-between mb-4">
            <div className="p-2.5 rounded-xl bg-amber-50"><FileText size={20} className="text-amber-600" /></div>
          </div>
          <p className="text-sm text-neutral-500">Pending Claims</p>
          <p className="text-xl font-semibold tracking-tight text-neutral-900 mt-1">{claimCount}</p>
        </div>
        <div className="bg-white rounded-2xl border border-neutral-50 p-6">
          <div className="flex items-start justify-between mb-4">
            <div className="p-2.5 rounded-xl bg-violet-50"><Vote size={20} className="text-violet-600" /></div>
          </div>
          <p className="text-sm text-neutral-500">Open Elections</p>
          <p className="text-xl font-semibold tracking-tight text-neutral-900 mt-1">{elections.length}</p>
        </div>
      </div>

      {/* Quick Links */}
      <div className="bg-white rounded-2xl border border-neutral-50 overflow-hidden">
        <div className="flex items-center justify-between px-6 py-5 border-b border-neutral-50">
          <h2 className="text-lg font-semibold tracking-tight text-neutral-900">Quick Actions</h2>
          <ChevronRight size={16} className="text-neutral-300" />
        </div>
        <div className="p-6 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {quickLinks.map((link, i) => (
            <Link
              key={link.label}
              to={link.href}
              className="btn-hover group flex items-start gap-4 p-5 rounded-xl bg-neutral-50 hover:bg-neutral-100 transition-all animate-fade-in"
              style={{ animationDelay: `${i * 0.05}s` }}
            >
              <div className={`p-2.5 rounded-xl ${colorMap[link.color].bg} flex-shrink-0`}>
                <link.icon size={18} className={colorMap[link.color].icon} />
              </div>
              <div>
                <p className="text-sm font-medium text-neutral-900">{link.label}</p>
                <p className="text-xs text-neutral-400 mt-0.5">{link.desc}</p>
              </div>
            </Link>
          ))}
        </div>
      </div>

      {/* Account Summary & Activity */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        {/* Account Summary */}
        <div className="bg-white rounded-2xl border border-neutral-50 overflow-hidden">
          <div className="flex items-center justify-between px-6 py-5 border-b border-neutral-50">
            <h2 className="text-lg font-semibold tracking-tight text-neutral-900">Account Summary</h2>
          </div>
          <div className="p-6 space-y-4">
            {[
              { label: 'Account Balance', value: profile?.account_summary?.account_balance ? `KES ${(profile.account_summary.account_balance / 100).toLocaleString()}` : '—' },
              { label: 'Basic Salary', value: profile?.employment_info?.basic_salary ? `KES ${(profile.employment_info.basic_salary / 100).toLocaleString()}` : '—' },
              { label: 'Department', value: profile?.employment_info?.department || '—' },
              { label: 'Membership Status', value: profile?.account_summary?.membership_status || '—' },
            ].map((item, i) => (
              <div key={i} className="flex items-center justify-between py-3 border-b border-neutral-50 last:border-0">
                <span className="text-sm text-neutral-600">{item.label}</span>
                <span className="text-sm font-medium text-neutral-900">{item.value}</span>
              </div>
            ))}
          </div>
        </div>

        {/* Login Activity */}
        <div className="bg-white rounded-2xl border border-neutral-50 overflow-hidden">
          <div className="flex items-center justify-between px-6 py-5 border-b border-neutral-50">
            <h2 className="text-lg font-semibold tracking-tight text-neutral-900">Login Activity</h2>
          </div>
          <div className="p-6 space-y-4">
            {[
              { label: 'Total Logins', value: loginStats?.total_logins ?? '—' },
              { label: 'Last 30 Days', value: loginStats?.logins_last_30_days ?? '—' },
              { label: 'Last Login', value: loginStats?.last_login ? new Date(loginStats.last_login).toLocaleString() : '—' },
            ].map((item, i) => (
              <div key={i} className="flex items-center justify-between py-3 border-b border-neutral-50 last:border-0">
                <span className="text-sm text-neutral-600">{item.label}</span>
                <span className="text-sm font-medium text-neutral-900">{item.value}</span>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Open Elections */}
      {elections.length > 0 && (
        <div className="bg-white rounded-2xl border border-neutral-50 overflow-hidden">
          <div className="flex items-center justify-between px-6 py-5 border-b border-neutral-50">
            <div className="flex items-center gap-3">
              <div className="p-2.5 rounded-xl bg-violet-50"><Vote size={20} className="text-violet-600" /></div>
              <h2 className="text-lg font-semibold tracking-tight text-neutral-900">Open Elections</h2>
            </div>
          </div>
          <div className="p-6 space-y-4">
            {elections.map((e, i) => (
              <div key={e.id} className="flex items-center justify-between py-3 border-b border-neutral-50 last:border-0">
                <div>
                  <p className="text-sm font-medium text-neutral-900">{e.title}</p>
                  <p className="text-xs text-neutral-400 mt-0.5">{e.description || 'No description'}</p>
                </div>
                <Link to="/portal/voting" className="btn-hover px-4 py-2 bg-violet-600 text-white rounded-xl text-xs font-medium hover:bg-violet-700 transition-all">
                  Vote Now
                </Link>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
