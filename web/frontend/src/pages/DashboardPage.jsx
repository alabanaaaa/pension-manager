import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { dashboard, pendingChanges } from '../lib/api';
import {
  Users, CreditCard, FileText, Vote, TrendingUp,
  AlertTriangle, CheckCircle, Clock, ArrowUpRight,
  ArrowDownRight, Activity
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
    {
      label: 'Total Members',
      value: stats?.total_members ?? '—',
      icon: Users,
      color: 'blue',
      change: stats?.member_growth,
      href: '/members',
    },
    {
      label: 'Active Members',
      value: stats?.active_members ?? '—',
      icon: Activity,
      color: 'green',
      href: '/members',
    },
    {
      label: 'Total Contributions',
      value: stats?.total_contributions ? `KES ${(stats.total_contributions / 1000000).toFixed(1)}M` : '—',
      icon: CreditCard,
      color: 'purple',
      href: '/contributions',
    },
    {
      label: 'Pending Claims',
      value: stats?.pending_claims ?? '—',
      icon: FileText,
      color: 'orange',
      href: '/claims',
    },
    {
      label: 'Active Elections',
      value: stats?.active_elections ?? '—',
      icon: Vote,
      color: 'indigo',
      href: '/voting',
    },
    {
      label: 'Pending Approvals',
      value: pendingCount?.total ?? '—',
      icon: Clock,
      color: 'yellow',
      href: '/maker-checker',
    },
  ];

  const colorMap = {
    blue: 'bg-blue-50 text-blue-700',
    green: 'bg-green-50 text-green-700',
    purple: 'bg-purple-50 text-purple-700',
    orange: 'bg-orange-50 text-orange-700',
    indigo: 'bg-indigo-50 text-indigo-700',
    yellow: 'bg-yellow-50 text-yellow-700',
  };

  const iconColorMap = {
    blue: 'text-blue-500',
    green: 'text-green-500',
    purple: 'text-purple-500',
    orange: 'text-orange-500',
    indigo: 'text-indigo-500',
    yellow: 'text-yellow-500',
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
        <p className="text-gray-500 mt-1">Overview of your pension management system</p>
      </div>

      {/* Stats grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        {statCards.map((stat) => (
          <Link
            key={stat.label}
            to={stat.href}
            className="bg-white rounded-xl border p-5 hover:shadow-md transition-shadow"
          >
            <div className="flex items-start justify-between">
              <div>
                <p className="text-sm text-gray-500">{stat.label}</p>
                <p className="text-2xl font-bold mt-1">{loading ? '...' : stat.value}</p>
                {stat.change !== undefined && (
                  <div className={`flex items-center gap-1 mt-2 text-sm ${stat.change >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                    {stat.change >= 0 ? <ArrowUpRight size={14} /> : <ArrowDownRight size={14} />}
                    <span>{Math.abs(stat.change)}% from last month</span>
                  </div>
                )}
              </div>
              <div className={`p-3 rounded-lg ${colorMap[stat.color]}`}>
                <stat.icon size={24} className={iconColorMap[stat.color]} />
              </div>
            </div>
          </Link>
        ))}
      </div>

      {/* Quick actions */}
      <div className="bg-white rounded-xl border p-6">
        <h2 className="text-lg font-semibold mb-4">Quick Actions</h2>
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
          {[
            { label: 'Add Member', href: '/members/new', color: 'blue' },
            { label: 'Record Contribution', href: '/contributions/new', color: 'green' },
            { label: 'New Claim', href: '/claims/new', color: 'orange' },
            { label: 'Create Election', href: '/voting/new', color: 'indigo' },
          ].map((action) => (
            <Link
              key={action.label}
              to={action.href}
              className={`p-4 rounded-lg border-2 border-dashed border-${action.color}-200 hover:border-${action.color}-400 hover:bg-${action.color}-50 transition-colors text-center`}
            >
              <span className={`text-sm font-medium text-${action.color}-700`}>{action.label}</span>
            </Link>
          ))}
        </div>
      </div>

      {/* Pending approvals */}
      {pendingCount && pendingCount.total > 0 && (
        <div className="bg-amber-50 border border-amber-200 rounded-xl p-5">
          <div className="flex items-center gap-3">
            <AlertTriangle size={24} className="text-amber-600" />
            <div>
              <h3 className="font-semibold text-amber-800">Pending Approvals</h3>
              <p className="text-sm text-amber-600">
                {pendingCount.members} member changes, {pendingCount.beneficiaries} beneficiary changes, {pendingCount.claims} claim changes awaiting review
              </p>
            </div>
            <Link
              to="/maker-checker"
              className="ml-auto bg-amber-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-amber-700"
            >
              Review
            </Link>
          </div>
        </div>
      )}
    </div>
  );
}
