import { Outlet, Link, useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import {
  LayoutDashboard, Users, CreditCard, FileText, Vote,
  Building2, Shield, BarChart3, Settings, LogOut,
  Menu, X, Bell, Search, Hospital, MessageSquare,
  TrendingUp, UserCheck, ClipboardList, Newspaper,
  Lock, Send, UserCircle
} from 'lucide-react';
import { useState } from 'react';

const adminNav = [
  { icon: LayoutDashboard, label: 'Dashboard', href: '/' },
  { icon: Users, label: 'Members', href: '/members' },
  { icon: CreditCard, label: 'Contributions', href: '/contributions' },
  { icon: FileText, label: 'Claims', href: '/claims' },
  { icon: Vote, label: 'Voting', href: '/voting' },
  { icon: Hospital, label: 'Hospitals', href: '/hospitals' },
  { icon: Building2, label: 'Sponsors', href: '/sponsors' },
  { icon: BarChart3, label: 'Reports', href: '/reports' },
  { icon: ClipboardList, label: 'Bulk Processing', href: '/bulk' },
  { icon: Shield, label: 'Maker-Checker', href: '/maker-checker' },
  { icon: TrendingUp, label: 'Tax', href: '/tax' },
  { icon: Send, label: 'SMS', href: '/sms' },
  { icon: Newspaper, label: 'News', href: '/news' },
  { icon: Lock, label: 'Security', href: '/security' },
  { icon: Settings, label: 'Settings', href: '/settings' },
];

const memberNav = [
  { icon: LayoutDashboard, label: 'My Dashboard', href: '/portal' },
  { icon: UserCircle, label: 'My Profile', href: '/portal/profile' },
  { icon: CreditCard, label: 'My Contributions', href: '/portal/contributions' },
  { icon: FileText, label: 'My Claims', href: '/portal/claims' },
  { icon: Vote, label: 'Voting', href: '/portal/voting' },
  { icon: TrendingUp, label: 'Projections', href: '/portal/projections' },
  { icon: MessageSquare, label: 'Feedback', href: '/portal/feedback' },
  { icon: Settings, label: 'Settings', href: '/portal/settings' },
];

export default function DashboardLayout() {
  const { user, logout, isAdmin, isOfficer } = useAuth();
  const location = useLocation();
  const navigate = useNavigate();
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  const navItems = isAdmin || isOfficer ? adminNav : memberNav;

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <div className="min-h-screen bg-gray-50 flex">
      {/* Mobile menu button */}
      <div className="lg:hidden fixed top-0 left-0 right-0 z-50 bg-white border-b px-4 py-3 flex items-center justify-between">
        <button onClick={() => setMobileMenuOpen(!mobileMenuOpen)} className="p-2">
          {mobileMenuOpen ? <X size={24} /> : <Menu size={24} />}
        </button>
        <h1 className="text-lg font-bold text-blue-900">Pension Manager</h1>
        <div className="w-10" />
      </div>

      {/* Mobile menu overlay */}
      {mobileMenuOpen && (
        <div className="lg:hidden fixed inset-0 z-40 bg-black/50" onClick={() => setMobileMenuOpen(false)} />
      )}

      {/* Sidebar */}
      <aside className={`
        fixed lg:static inset-y-0 left-0 z-40
        ${mobileMenuOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'}
        ${sidebarOpen ? 'w-64' : 'w-20'}
        bg-slate-900 text-white transition-all duration-300 flex flex-col
        pt-16 lg:pt-0
      `}>
        {/* Logo */}
        <div className="p-4 border-b border-slate-700 flex items-center gap-3">
          <div className="w-10 h-10 bg-blue-600 rounded-lg flex items-center justify-center flex-shrink-0">
            <Shield size={24} />
          </div>
          {sidebarOpen && (
            <div>
              <h1 className="font-bold text-lg">Pension Manager</h1>
              <p className="text-xs text-slate-400">{user?.role}</p>
            </div>
          )}
        </div>

        {/* Navigation */}
        <nav className="flex-1 overflow-y-auto p-3 space-y-1">
          {navItems.map((item) => {
            const isActive = location.pathname === item.href || location.pathname.startsWith(item.href + '/');
            return (
              <Link
                key={item.href}
                to={item.href}
                onClick={() => setMobileMenuOpen(false)}
                className={`
                  flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors
                  ${isActive
                    ? 'bg-blue-600 text-white'
                    : 'text-slate-300 hover:bg-slate-800 hover:text-white'
                  }
                `}
              >
                <item.icon size={20} className="flex-shrink-0" />
                {sidebarOpen && <span>{item.label}</span>}
              </Link>
            );
          })}
        </nav>

        {/* User section */}
        <div className="p-3 border-t border-slate-700">
          <div className={`flex items-center gap-3 px-3 py-2 ${!sidebarOpen && 'justify-center'}`}>
            <div className="w-8 h-8 bg-blue-600 rounded-full flex items-center justify-center flex-shrink-0">
              <span className="text-sm font-bold">{user?.name?.[0]?.toUpperCase()}</span>
            </div>
            {sidebarOpen && (
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium truncate">{user?.name}</p>
                <p className="text-xs text-slate-400 truncate">{user?.email}</p>
              </div>
            )}
            {sidebarOpen && (
              <button onClick={handleLogout} className="text-slate-400 hover:text-white">
                <LogOut size={18} />
              </button>
            )}
          </div>
        </div>
      </aside>

      {/* Main content */}
      <div className="flex-1 flex flex-col min-w-0">
        {/* Top bar */}
        <header className="bg-white border-b px-4 lg:px-6 py-3 flex items-center justify-between sticky top-0 z-30">
          <div className="flex items-center gap-4">
            <button
              onClick={() => setSidebarOpen(!sidebarOpen)}
              className="hidden lg:block p-2 hover:bg-gray-100 rounded-lg"
            >
              <Menu size={20} />
            </button>
            <div className="relative hidden md:block">
              <Search size={18} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
              <input
                type="text"
                placeholder="Search..."
                className="pl-10 pr-4 py-2 border rounded-lg text-sm w-64 focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
          </div>
          <div className="flex items-center gap-3">
            <button className="relative p-2 hover:bg-gray-100 rounded-lg">
              <Bell size={20} />
              <span className="absolute top-1 right-1 w-2 h-2 bg-red-500 rounded-full" />
            </button>
          </div>
        </header>

        {/* Page content */}
        <main className="flex-1 p-4 lg:p-6 overflow-auto pt-20 lg:pt-6">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
