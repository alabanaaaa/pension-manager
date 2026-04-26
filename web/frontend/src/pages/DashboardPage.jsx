import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { dashboard, pendingChanges } from '../lib/api';
import { useAuth } from '../context/AuthContext';
import { AnimatedNumber } from '../components';
import {
  Users, CreditCard, FileText, Vote,
  AlertTriangle, Clock, Activity, Shield,
  ArrowRight, Calendar
} from 'lucide-react';

export default function DashboardPage() {
  const { user } = useAuth();
  const [stats, setStats] = useState(null);
  const [pendingCount, setPendingCount] = useState(null);
  const [loading, setLoading] = useState(true);

  const getGreeting = () => {
    const hour = new Date().getHours();
    const name = user?.name?.split(' ')[0] || 'User';
    if (hour < 12) return `Good morning, ${name}`;
    if (hour < 17) return `Good afternoon, ${name}`;
    return `Good evening, ${name}`;
  };

  useEffect(() => {
    Promise.all([
      dashboard.get().catch(() => null),
      pendingChanges.getCount().catch(() => null),
    ]).then(([dashRes, pendingRes]) => {
      if (dashRes?.data) setStats(dashRes.data);
      if (pendingRes?.data) setPendingCount(pendingRes.data);
      setLoading(false);
    });
  }, []);

  const greeting = getGreeting();

  const statCards = [
    { label: 'Total Members', value: stats?.total_members ?? 0, icon: Users, href: '/members' },
    { label: 'Active Members', value: stats?.active_members ?? 0, icon: Activity, href: '/members' },
    { label: 'Contributions', value: stats?.total_contributions ? Math.round(stats.total_contributions / 1000000 * 10) / 10 : 0, prefix: '', suffix: 'M', icon: CreditCard, href: '/contributions' },
    { label: 'Pending Claims', value: stats?.pending_claims ?? 0, icon: FileText, href: '/claims' },
    { label: 'Active Elections', value: stats?.active_elections ?? 0, icon: Vote, href: '/voting' },
    { label: 'Pending Approvals', value: pendingCount?.total ?? 0, icon: Clock, href: '/maker-checker' },
  ];

  const recentActivity = [
    { icon: Users, label: 'New member registered', time: '2 min ago' },
    { icon: CreditCard, label: 'Contribution received - KES 50,000', time: '15 min ago' },
    { icon: FileText, label: 'Claim #CLM-042 submitted', time: '1 hour ago' },
    { icon: Vote, label: 'Election "Trustee 2026" opened', time: '3 hours ago' },
    { icon: Shield, label: 'Member details updated', time: '5 hours ago' },
  ];

  const quickLinks = [
    { label: 'Add Member', desc: 'Register a new scheme member', href: '/members/new', icon: Users },
    { label: 'Record Contribution', desc: 'Log a member contribution', href: '/contributions/new', icon: CreditCard },
    { label: 'New Claim', desc: 'Process a benefit claim', href: '/claims/new', icon: FileText },
    { label: 'Create Election', desc: 'Set up a new voting election', href: '/voting/new', icon: Vote },
  ];

  return (
    <div className="space-y-8">
      {/* Header - Uber Minimal */}
      <div className="flex items-end justify-between">
        <div className="animate-fade-in-up">
          <p className="text-sm text-gray-500 mb-1">{greeting}</p>
          <h1 className="text-2xl font-bold tracking-tight text-black">Dashboard</h1>
          <p className="text-sm text-gray-500 mt-1">Overview of your pension management system</p>
        </div>
        <div className="hidden sm:flex items-center gap-2 text-sm text-gray-400 animate-fade-in">
          <Calendar size={14} />
          <span>{new Date().toLocaleDateString('en-US', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' })}</span>
        </div>
      </div>

      {/* Stats Grid - Uber Style */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        {statCards.map((stat, i) => (
          <Link
            key={stat.label}
            to={stat.href}
            className="group bg-white border border-gray-200 rounded-lg p-5 hover:border-black transition-all duration-150 cursor-pointer"
            style={{ animationDelay: `${i * 50}ms` }}
          >
            <div className="flex items-start justify-between mb-4">
              <div className="p-2 border border-gray-200 rounded group-hover:border-black group-hover:bg-black group-hover:text-white transition-all duration-150">
                <stat.icon size={18} className="text-black group-hover:text-white transition-colors" />
              </div>
              <ArrowRight size={14} className="text-gray-300 group-hover:text-black group-hover:translate-x-0.5 transition-all" />
            </div>
            <p className="text-xs font-medium text-gray-500 uppercase tracking-wide">{stat.label}</p>
            <p className="text-2xl font-bold tracking-tight text-black mt-1">
              {loading ? '—' : (
                stat.prefix ? `${stat.prefix}${stat.value}` : stat.value
              )}
            </p>
          </Link>
        ))}
      </div>

      {/* Main Content Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Quick Actions - Uber Style */}
        <div className="lg:col-span-2 bg-white border border-gray-200 rounded-lg animate-fade-in-up" style={{ animationDelay: '200ms' }}>
          <div className="flex items-center justify-between px-5 py-4 border-b border-gray-100">
            <h2 className="text-base font-semibold text-black">Quick Actions</h2>
          </div>
          <div className="p-5 grid grid-cols-1 sm:grid-cols-2 gap-3">
            {quickLinks.map((action, i) => (
              <Link
                key={action.label}
                to={action.href}
                className="group flex items-center gap-4 p-4 border border-gray-200 rounded-lg hover:border-black hover:bg-gray-50/50 transition-all duration-150"
                style={{ animationDelay: `${300 + i * 50}ms` }}
              >
                <div className="p-2.5 border border-gray-200 rounded group-hover:border-black group-hover:bg-black group-hover:text-white transition-all duration-150">
                  <action.icon size={18} className="text-black group-hover:text-white transition-colors" />
                </div>
                <div className="flex-1">
                  <p className="text-sm font-semibold text-black">{action.label}</p>
                  <p className="text-xs text-gray-500 mt-0.5">{action.desc}</p>
                </div>
                <ArrowRight size={14} className="text-gray-300 group-hover:text-black group-hover:translate-x-1 transition-all" />
              </Link>
            ))}
          </div>
        </div>

        {/* Recent Activity */}
        <div className="bg-white border border-gray-200 rounded-lg animate-fade-in-up" style={{ animationDelay: '250ms' }}>
          <div className="flex items-center justify-between px-5 py-4 border-b border-gray-100">
            <h2 className="text-base font-semibold text-black">Recent Activity</h2>
          </div>
          <div className="p-5 space-y-4">
            {recentActivity.map((item, i) => (
              <div 
                key={i} 
                className="flex items-start gap-3" 
                style={{ animationDelay: `${350 + i * 50}ms` }}
              >
                <div className="p-2 border border-gray-200 rounded flex-shrink-0 mt-0.5 group-hover:border-black transition-colors">
                  <item.icon size={14} className="text-black" />
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm text-black truncate">{item.label}</p>
                  <p className="text-xs text-gray-400 mt-0.5">{item.time}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Pending Approvals Alert - Uber Style */}
      {pendingCount && pendingCount.total > 0 && (
        <div className="bg-black text-white rounded-lg p-5 animate-fade-in-up" style={{ animationDelay: '300ms' }}>
          <div className="flex items-center gap-4">
            <div className="p-2.5 bg-white/10 rounded">
              <AlertTriangle size={20} className="text-white" />
            </div>
            <div className="flex-1">
              <h3 className="font-semibold text-white">Pending Approvals</h3>
              <p className="text-sm text-gray-300 mt-0.5">
                {pendingCount.members} member · {pendingCount.beneficiaries} beneficiary · {pendingCount.claims} claim
              </p>
            </div>
            <Link 
              to="/maker-checker" 
              className="px-5 py-2.5 bg-white text-black text-sm font-semibold rounded hover:bg-gray-100 transition-colors flex-shrink-0 hover:scale-105 transition-transform"
            >
              Review
            </Link>
          </div>
        </div>
      )}
    </div>
  );
}
