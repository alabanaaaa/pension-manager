import { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import { portal, contributions, claims, voting } from '../lib/api';
import {
  User, CreditCard, FileText, Vote, TrendingUp,
  MessageSquare, Settings, Calendar, Clock, ArrowRight,
  Activity, Bell, ChevronRight, Loader2
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
    if (hour < 12) return `Good morning, ${name}`;
    if (hour < 17) return `Good afternoon, ${name}`;
    return `Good evening, ${name}`;
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
    { icon: User, label: 'My Profile', desc: 'View and update personal info', href: '/portal/profile' },
    { icon: CreditCard, label: 'Contributions', desc: 'View contribution history', href: '/portal/contributions' },
    { icon: FileText, label: 'My Claims', desc: 'Check claim status', href: '/portal/claims' },
    { icon: Vote, label: 'Vote', desc: 'Cast your vote', href: '/portal/voting' },
    { icon: TrendingUp, label: 'Projections', desc: 'Benefit projections', href: '/portal/projections' },
    { icon: MessageSquare, label: 'Feedback', desc: 'Submit feedback', href: '/portal/feedback' },
  ];

  if (loading) {
    return (
      <div className="p-16 text-center">
        <Loader2 size={24} className="animate-spin mx-auto text-gray-300" />
        <p className="text-sm text-gray-400 mt-3">Loading your dashboard...</p>
      </div>
    );
  }

  return (
    <div className="space-y-8 animate-fade-in-up">
      {/* Header */}
      <div className="flex items-end justify-between">
        <div>
          <p className="text-sm text-gray-500 mb-1">{greeting}</p>
          <h1 className="text-2xl font-bold tracking-tight text-black">My Dashboard</h1>
          <p className="text-sm text-gray-500 mt-1">Welcome to your member portal</p>
        </div>
        <div className="flex items-center gap-2 text-sm text-gray-400">
          <Calendar size={14} />
          <span>{new Date().toLocaleDateString('en-US', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' })}</span>
        </div>
      </div>

      {/* Stats - Uber Style */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        {[
          { icon: User, label: 'Member No', value: profile?.member_no || '—' },
          { icon: CreditCard, label: 'Contributions', value: contribCount },
          { icon: FileText, label: 'Pending Claims', value: claimCount },
          { icon: Vote, label: 'Open Elections', value: elections.length },
        ].map((stat, i) => (
          <div key={i} className="bg-white border border-gray-200 rounded-lg p-4">
            <div className="flex items-center gap-2 mb-3">
              <stat.icon size={14} className="text-gray-400" />
              <p className="text-xs font-medium text-gray-500 uppercase tracking-wide">{stat.label}</p>
            </div>
            <p className="text-xl font-bold text-black">{stat.value}</p>
          </div>
        ))}
      </div>

      {/* Quick Links */}
      <div className="bg-white border border-gray-200 rounded-lg">
        <div className="px-5 py-4 border-b border-gray-100">
          <h2 className="text-base font-semibold text-black">Quick Actions</h2>
        </div>
        <div className="p-5 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
          {quickLinks.map((link, i) => (
            <Link
              key={link.label}
              to={link.href}
              className="group flex items-center gap-4 p-4 border border-gray-200 rounded-lg hover:border-black hover:bg-gray-50 transition-all"
            >
              <div className="p-2 border border-gray-200 rounded">
                <link.icon size={18} className="text-black" />
              </div>
              <div className="flex-1">
                <p className="text-sm font-semibold text-black">{link.label}</p>
                <p className="text-xs text-gray-500 mt-0.5">{link.desc}</p>
              </div>
              <ArrowRight size={14} className="text-gray-300 group-hover:text-black group-hover:translate-x-0.5 transition-all" />
            </Link>
          ))}
        </div>
      </div>

      {/* Two Column Layout */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Account Summary */}
        <div className="bg-white border border-gray-200 rounded-lg">
          <div className="px-5 py-4 border-b border-gray-100">
            <h2 className="text-base font-semibold text-black">Account Summary</h2>
          </div>
          <div className="p-5">
            {[
              { label: 'Account Balance', value: profile?.account_summary?.account_balance ? `KES ${(profile.account_summary.account_balance / 100).toLocaleString()}` : '—' },
              { label: 'Basic Salary', value: profile?.employment_info?.basic_salary ? `KES ${(profile.employment_info.basic_salary / 100).toLocaleString()}` : '—' },
              { label: 'Department', value: profile?.employment_info?.department || '—' },
              { label: 'Membership Status', value: profile?.account_summary?.membership_status || '—' },
            ].map((item, i) => (
              <div key={i} className="flex items-center justify-between py-3 border-b border-gray-100 last:border-0">
                <span className="text-sm text-gray-500">{item.label}</span>
                <span className="text-sm font-medium text-black">{item.value}</span>
              </div>
            ))}
          </div>
        </div>

        {/* Login Activity */}
        <div className="bg-white border border-gray-200 rounded-lg">
          <div className="px-5 py-4 border-b border-gray-100">
            <h2 className="text-base font-semibold text-black">Login Activity</h2>
          </div>
          <div className="p-5">
            {[
              { label: 'Total Logins', value: loginStats?.total_logins ?? '—' },
              { label: 'Last 30 Days', value: loginStats?.logins_last_30_days ?? '—' },
              { label: 'Last Login', value: loginStats?.last_login ? new Date(loginStats.last_login).toLocaleString() : '—' },
            ].map((item, i) => (
              <div key={i} className="flex items-center justify-between py-3 border-b border-gray-100 last:border-0">
                <span className="text-sm text-gray-500">{item.label}</span>
                <span className="text-sm font-medium text-black">{item.value}</span>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Open Elections */}
      {elections.length > 0 && (
        <div className="bg-black text-white rounded-lg p-5">
          <div className="flex items-center gap-3 mb-4">
            <Vote size={18} />
            <h2 className="text-base font-semibold">Open Elections</h2>
          </div>
          <div className="space-y-4">
            {elections.map((e) => (
              <div key={e.id} className="flex items-center justify-between py-3 border-b border-white/10 last:border-0">
                <div>
                  <p className="font-medium">{e.title}</p>
                  <p className="text-sm text-gray-400 mt-0.5">{e.description || 'No description'}</p>
                </div>
                <Link to="/portal/voting" className="px-4 py-2 bg-white text-black text-sm font-semibold rounded hover:bg-gray-100 transition-colors">
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
