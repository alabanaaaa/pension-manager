import { Outlet, Link, useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import {
  LayoutDashboard, Users, CreditCard, FileText, Vote,
  Building2, Shield, BarChart3, Settings, LogOut,
  Menu, X, Bell, Search, Hospital, MessageSquare,
  TrendingUp, UserCheck, ClipboardList, Newspaper,
  Lock, Send, UserCircle, ChevronDown
} from 'lucide-react';
import { useState, useEffect, useRef } from 'react';
import bankLogo from '/bank-logo.svg';

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
  { icon: LayoutDashboard, label: 'Dashboard', href: '/portal' },
  { icon: UserCircle, label: 'My Profile', href: '/portal/profile' },
  { icon: CreditCard, label: 'Contributions', href: '/portal/contributions' },
  { icon: FileText, label: 'Claims', href: '/portal/claims' },
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
  const [userMenuOpen, setUserMenuOpen] = useState(false);
  const userMenuRef = useRef(null);

  const navItems = isAdmin || isOfficer ? adminNav : memberNav;

  useEffect(() => {
    const handleClickOutside = (e) => {
      if (userMenuRef.current && !userMenuRef.current.contains(e.target)) {
        setUserMenuOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <div className="min-h-screen bg-neutral-50 flex">
      {/* Mobile menu overlay */}
      {mobileMenuOpen && (
        <div className="lg:hidden fixed inset-0 z-40 bg-black/20 backdrop-blur-sm" onClick={() => setMobileMenuOpen(false)} />
      )}

      {/* Sidebar */}
      <aside className={`
        fixed lg:static inset-y-0 left-0 z-50
        ${mobileMenuOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'}
        ${sidebarOpen ? 'w-64' : 'w-[72px]'}
        bg-white border-r border-neutral-100 transition-all duration-300 flex flex-col
      `}>
        {/* Logo */}
        <div className="px-5 py-5 flex items-center gap-3 border-b border-neutral-50">
          <img src={bankLogo} alt="Logo" className="w-8 h-8 flex-shrink-0" />
          {sidebarOpen && (
            <div>
              <h1 className="text-sm font-semibold text-neutral-900 tracking-tight">Pension Manager</h1>
              <p className="text-[10px] text-neutral-400 uppercase tracking-wider">Admin Portal</p>
            </div>
          )}
        </div>

        {/* Navigation */}
        <nav className="flex-1 overflow-y-auto py-4 px-3 space-y-0.5">
          {navItems.map((item) => {
            const isActive = location.pathname === item.href || (item.href !== '/' && location.pathname.startsWith(item.href));
            return (
              <Link
                key={item.href}
                to={item.href}
                onClick={() => setMobileMenuOpen(false)}
                className={`
                  flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm transition-all
                  ${isActive
                    ? 'bg-neutral-900 text-white'
                    : 'text-neutral-500 hover:bg-neutral-50 hover:text-neutral-900'
                  }
                `}
              >
                <item.icon size={18} className="flex-shrink-0" />
                {sidebarOpen && <span className="font-medium">{item.label}</span>}
              </Link>
            );
          })}
        </nav>

        {/* User */}
        <div className="p-3 border-t border-neutral-50">
          <div className={`flex items-center gap-3 px-3 py-2.5 rounded-xl ${!sidebarOpen && 'justify-center'}`}>
            <div className="w-8 h-8 bg-neutral-900 rounded-full flex items-center justify-center flex-shrink-0">
              <span className="text-xs font-medium text-white">{user?.name?.[0]?.toUpperCase()}</span>
            </div>
            {sidebarOpen && (
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium text-neutral-900 truncate">{user?.name}</p>
                <p className="text-xs text-neutral-400 truncate">{user?.role?.replace('_', ' ')}</p>
              </div>
            )}
            {sidebarOpen && (
              <button onClick={handleLogout} className="text-neutral-300 hover:text-neutral-600 p-1 transition-colors" title="Sign out">
                <LogOut size={16} />
              </button>
            )}
          </div>
        </div>
      </aside>

      {/* Main content */}
      <div className="flex-1 flex flex-col min-w-0">
        {/* Top bar */}
        <header className="bg-white border-b border-neutral-100 px-4 lg:px-6 py-3 flex items-center justify-between sticky top-0 z-30">
          <div className="flex items-center gap-4">
            <button
              onClick={() => setMobileMenuOpen(true)}
              className="lg:hidden p-2 hover:bg-neutral-50 rounded-lg"
            >
              <Menu size={20} className="text-neutral-500" />
            </button>
            <button
              onClick={() => setSidebarOpen(!sidebarOpen)}
              className="hidden lg:block p-2 hover:bg-neutral-50 rounded-lg"
            >
              {sidebarOpen ? <X size={18} className="text-neutral-400" /> : <Menu size={18} className="text-neutral-400" />}
            </button>
            <div className="relative hidden md:block">
              <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-neutral-300" />
              <input
                type="text"
                placeholder="Search..."
                className="pl-9 pr-4 py-2 bg-neutral-50 rounded-xl text-sm w-56 focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:bg-white transition-all placeholder:text-neutral-300"
              />
            </div>
          </div>
          <div className="flex items-center gap-2">
            <button className="relative p-2 hover:bg-neutral-50 rounded-lg transition-colors">
              <Bell size={18} className="text-neutral-400" />
              <span className="absolute top-1.5 right-1.5 w-1.5 h-1.5 bg-red-500 rounded-full" />
            </button>
            <div className="relative" ref={userMenuRef}>
              <button
                onClick={() => setUserMenuOpen(!userMenuOpen)}
                className="flex items-center gap-2 p-1.5 hover:bg-neutral-50 rounded-lg transition-colors"
              >
                <div className="w-7 h-7 bg-neutral-900 rounded-full flex items-center justify-center">
                  <span className="text-xs font-medium text-white">{user?.name?.[0]?.toUpperCase()}</span>
                </div>
                <ChevronDown size={14} className="text-neutral-300" />
              </button>
              {userMenuOpen && (
                <div className="absolute right-0 mt-2 w-48 bg-white rounded-xl shadow-lg border border-neutral-100 py-2 animate-fade-in-up">
                  <div className="px-4 py-2 border-b border-neutral-50">
                    <p className="text-sm font-medium text-neutral-900">{user?.name}</p>
                    <p className="text-xs text-neutral-400">{user?.email}</p>
                  </div>
                  <Link to="/settings" className="flex items-center gap-2 px-4 py-2 text-sm text-neutral-600 hover:bg-neutral-50 transition-colors">
                    <Settings size={14} /> Settings
                  </Link>
                  <button onClick={handleLogout} className="flex items-center gap-2 w-full px-4 py-2 text-sm text-red-600 hover:bg-red-50 transition-colors">
                    <LogOut size={14} /> Sign out
                  </button>
                </div>
              )}
            </div>
          </div>
        </header>

        {/* Page content */}
        <main className="flex-1 p-6 lg:p-10 overflow-auto">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
