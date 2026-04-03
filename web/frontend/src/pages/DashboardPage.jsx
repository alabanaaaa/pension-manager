import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { dashboard, pendingChanges } from '../lib/api';
import {
  Users, CreditCard, FileText, Vote, TrendingUp,
  AlertTriangle, CheckCircle, Clock, ArrowUpRight,
  ArrowDownRight, Activity, Building2, Shield,
  ChevronRight, BarChart3, PieChart, Calendar
} from 'lucide-react';

export default function DashboardPage() {
  const [stats, setStats] = useState(null);
  const [pendingCount, setPendingCount] = useState(null);
  const [loading, setLoading] = useState(true);

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

  const statCards = [
    { label: 'Total Members', value: stats?.total_members ?? '—', icon: Users, color: 'blue', change: stats?.member_growth, href: '/members' },
    { label: 'Active Members', value: stats?.active_members ?? '—', icon: Activity, color: 'emerald', href: '/members' },
    { label: 'Contributions', value: stats?.total_contributions ? `KES ${(stats.total_contributions / 1000000).toFixed(1)}M` : '—', icon: CreditCard, color: 'violet', href: '/contributions' },
    { label: 'Pending Claims', value: stats?.pending_claims ?? '—', icon: FileText, color: 'amber', href: '/claims' },
    { label: 'Active Elections', value: stats?.active_elections ?? '—', icon: Vote, color: 'sky', href: '/voting' },
    { label: 'Pending Approvals', value: pendingCount?.total ?? '—', icon: Clock, color: 'rose', href: '/maker-checker' },
  ];

  const colorMap = {
    blue: 'bg-blue-50 text-blue-600',
    emerald: 'bg-emerald-50 text-emerald-600',
    violet: 'bg-violet-50 text-violet-600',
    amber: 'bg-amber-50 text-amber-600',
    sky: 'bg-sky-50 text-sky-600',
    rose: 'bg-rose-50 text-rose-600',
  };

  const recentActivity = [
    { icon: Users, label: 'New member registered', time: '2 min ago', color: 'blue' },
    { icon: CreditCard, label: 'Contribution received - KES 50,000', time: '15 min ago', color: 'emerald' },
    { icon: FileText, label: 'Claim #CLM-042 submitted', time: '1 hour ago', color: 'amber' },
    { icon: Vote, label: 'Election "Trustee 2026" opened', time: '3 hours ago', color: 'sky' },
    { icon: Shield, label: 'Member details updated (pending approval)', time: '5 hours ago', color: 'violet' },
  ];

  const quickLinks = [
    { label: 'Add Member', desc: 'Register a new scheme member', href: '/members/new', icon: Users, color: 'blue' },
    { label: 'Record Contribution', desc: 'Log a member contribution', href: '/contributions/new', icon: CreditCard, color: 'emerald' },
    { label: 'New Claim', desc: 'Process a benefit claim', href: '/claims/new', icon: FileText, color: 'amber' },
    { label: 'Create Election', desc: 'Set up a new voting election', href: '/voting/new', icon: Vote, color: 'sky' },
  ];

  return (
    <div className="space-y-10">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-end sm:justify-between gap-4">
        <div>
          <h1 className="text-3xl font-semibold tracking-tight text-neutral-900">Dashboard</h1>
          <p className="text-neutral-500 mt-1">Overview of your pension management system</p>
        </div>
        <div className="flex items-center gap-2 text-sm text-neutral-400">
          <Calendar size={14} />
          <span>{new Date().toLocaleDateString('en-US', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' })}</span>
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
        {statCards.map((stat, i) => (
          <Link
            key={stat.label}
            to={stat.href}
            className="btn-hover group bg-white rounded-2xl p-6 hover:shadow-sm border border-neutral-100 transition-all"
            style={{ animationDelay: `${i * 0.05}s` }}
          >
            <div className="flex items-start justify-between mb-4">
              <div className={`p-2.5 rounded-xl ${colorMap[stat.color]}`}>
                <stat.icon size={20} />
              </div>
              {stat.change !== undefined && (
                <div className={`flex items-center gap-0.5 text-xs font-medium px-2 py-1 rounded-full ${stat.change >= 0 ? 'bg-emerald-50 text-emerald-600' : 'bg-red-50 text-red-600'}`}>
                  {stat.change >= 0 ? <ArrowUpRight size={12} /> : <ArrowDownRight size={12} />}
                  {Math.abs(stat.change)}%
                </div>
              )}
            </div>
            <p className="text-sm text-neutral-500">{stat.label}</p>
            <p className="text-2xl font-semibold tracking-tight text-neutral-900 mt-1">{loading ? '—' : stat.value}</p>
          </Link>
        ))}
      </div>

      {/* Main content grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Quick actions */}
        <div className="lg:col-span-2 bg-white rounded-2xl p-6 border border-neutral-100">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-lg font-semibold tracking-tight text-neutral-900">Quick Actions</h2>
            <ChevronRight size={16} className="text-neutral-300" />
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            {quickLinks.map((action) => (
              <Link
                key={action.label}
                to={action.href}
                className="btn-hover group flex items-start gap-4 p-5 rounded-xl bg-neutral-50 hover:bg-neutral-100 transition-all"
              >
                <div className={`p-2.5 rounded-xl ${colorMap[action.color]} flex-shrink-0`}>
                  <action.icon size={18} />
                </div>
                <div>
                  <p className="text-sm font-medium text-neutral-900">{action.label}</p>
                  <p className="text-xs text-neutral-400 mt-0.5">{action.desc}</p>
                </div>
              </Link>
            ))}
          </div>
        </div>

        {/* Recent activity */}
        <div className="bg-white rounded-2xl p-6 border border-neutral-100">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-lg font-semibold tracking-tight text-neutral-900">Recent Activity</h2>
            <ChevronRight size={16} className="text-neutral-300" />
          </div>
          <div className="space-y-5">
            {recentActivity.map((item, i) => (
              <div key={i} className="flex items-start gap-3">
                <div className={`p-2 rounded-lg ${colorMap[item.color]} flex-shrink-0 mt-0.5`}>
                  <item.icon size={14} />
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm text-neutral-700 truncate">{item.label}</p>
                  <p className="text-xs text-neutral-400 mt-0.5">{item.time}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Pending approvals alert */}
      {pendingCount && pendingCount.total > 0 && (
        <div className="bg-amber-50 border border-amber-100 rounded-2xl p-6">
          <div className="flex items-center gap-4">
            <div className="p-3 bg-amber-100 rounded-xl">
              <AlertTriangle size={20} className="text-amber-600" />
            </div>
            <div className="flex-1">
              <h3 className="font-medium text-amber-900">Pending Approvals</h3>
              <p className="text-sm text-amber-600 mt-0.5">
                {pendingCount.members} member changes · {pendingCount.beneficiaries} beneficiary changes · {pendingCount.claims} claim changes
              </p>
            </div>
            <Link to="/maker-checker" className="btn-hover px-5 py-2.5 bg-amber-600 text-white rounded-xl text-sm font-medium hover:bg-amber-700 transition-all flex-shrink-0">
              Review All
            </Link>
          </div>
        </div>
      )}

      {/* Bottom stats row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        <div className="bg-white rounded-2xl p-6 border border-neutral-100">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-lg font-semibold tracking-tight text-neutral-900">Scheme Overview</h2>
          </div>
          <div className="space-y-4">
            {[
              { label: 'DB Scheme Members', value: '—', icon: Building2 },
              { label: 'DC Scheme Members', value: '—', icon: Building2 },
              { label: 'Medical Fund Members', value: '—', icon: Activity },
            ].map((item, i) => (
              <div key={i} className="flex items-center justify-between py-3 border-b border-neutral-50 last:border-0">
                <div className="flex items-center gap-3">
                  <item.icon size={16} className="text-neutral-400" />
                  <span className="text-sm text-neutral-600">{item.label}</span>
                </div>
                <span className="text-sm font-medium text-neutral-900">{item.value}</span>
              </div>
            ))}
          </div>
        </div>

        <div className="bg-white rounded-2xl p-6 border border-neutral-100">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-lg font-semibold tracking-tight text-neutral-900">System Status</h2>
          </div>
          <div className="space-y-4">
            {[
              { label: 'Database', status: 'Connected', ok: true },
              { label: 'M-Pesa Integration', status: 'Sandbox', ok: true },
              { label: 'SMS Gateway', status: 'Mock Mode', ok: true },
              { label: 'News API', status: 'Mock Mode', ok: true },
            ].map((item, i) => (
              <div key={i} className="flex items-center justify-between py-3 border-b border-neutral-50 last:border-0">
                <span className="text-sm text-neutral-600">{item.label}</span>
                <div className="flex items-center gap-2">
                  <div className={`w-1.5 h-1.5 rounded-full ${item.ok ? 'bg-emerald-500' : 'bg-red-500'}`} />
                  <span className="text-sm text-neutral-500">{item.status}</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
