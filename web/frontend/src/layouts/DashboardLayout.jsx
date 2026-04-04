import { useState, useEffect, useRef } from 'react';
import { Outlet, Link, useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import {
  LayoutDashboard, Users, CreditCard, FileText, Vote,
  Building2, Shield, BarChart3, Settings, LogOut,
  Menu, X, Bell, Search, Hospital, MessageSquare,
  TrendingUp, UserCheck, ClipboardList, Newspaper,
  Lock, Send, UserCircle, ChevronDown, ChevronLeft, ChevronRight
} from 'lucide-react';
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
    <div className="min-h-screen bg-white flex">
      {/* Mobile overlay */}
      {mobileMenuOpen && (
        <div className="lg:hidden fixed inset-0 z-40 bg-black/20 backdrop-blur-sm" onClick={() => setMobileMenuOpen(false)} />
      )}

      {/* Sidebar */}
      <aside className={`
        fixed lg:static inset-y-0 left-0 z-50
        ${mobileMenuOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'}
        ${sidebarOpen ? 'w-64' : 'w-[72px]'}
        bg-neutral-950 text-white transition-all duration-300 ease-in-out flex flex-col
      `}>
        {/* Brand */}
        <div className="px-5 py-5 flex items-center gap-3 border-b border-white/10">
          <img src={bankLogo} alt="Logo" className="w-8 h-8 flex-shrink-0 opacity-90" />
          {sidebarOpen && (
            <div className="overflow-hidden">
              <h1 className="text-sm font-semibold tracking-tight animate-fade-in">Pension Manager</h1>
              <p className="text-[10px] text-neutral-500 uppercase tracking-wider">Admin Portal</p>
            </div>
          )}
        </div>

        {/* Toggle button (desktop) */}
        <button
          onClick={() => setSidebarOpen(!sidebarOpen)}
          className="hidden lg:flex absolute -right-3 top-7 w-6 h-6 bg-neutral-950 border border-neutral-800 rounded-full items-center justify-center hover:bg-neutral-800 transition-colors z-50"
        >
          {sidebarOpen ? <ChevronLeft size={12} className="text-neutral-400" /> : <ChevronRight size={12} className="text-neutral-400" />}
        </button>

        {/* Nav */}
        <nav className="flex-1 overflow-y-auto py-4 px-3 space-y-0.5">
          {navItems.map((item, i) => {
            const isActive = location.pathname === item.href || (item.href !== '/' && location.pathname.startsWith(item.href));
            return (
              <Link
                key={item.href}
                to={item.href}
                onClick={() => setMobileMenuOpen(false)}
                className={`
                  flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm transition-all duration-200
                  ${isActive
                    ? 'bg-white text-neutral-950 font-medium'
                    : 'text-neutral-400 hover:bg-white/5 hover:text-white'
                  }
                `}
                style={{ animationDelay: `${i * 0.03}s` }}
              >
                <item.icon size={18} className="flex-shrink-0" />
                {sidebarOpen && <span className="truncate">{item.label}</span>}
              </Link>
            );
          })}
        </nav>

        {/* User */}
        <div className="p-3 border-t border-white/10">
          <div className={`flex items-center gap-3 px-3 py-2.5 rounded-xl ${!sidebarOpen && 'justify-center'}`}>
            <div className="w-8 h-8 bg-white rounded-full flex items-center justify-center flex-shrink-0">
              <span className="text-xs font-semibold text-neutral-950">{user?.name?.[0]?.toUpperCase()}</span>
            </div>
            {sidebarOpen && (
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium truncate">{user?.name}</p>
                <p className="text-xs text-neutral-500 truncate capitalize">{user?.role?.replace('_', ' ')}</p>
              </div>
            )}
            {sidebarOpen && (
              <button onClick={handleLogout} className="text-neutral-500 hover:text-white p-1 transition-colors" title="Sign out">
                <LogOut size={16} />
              </button>
            )}
          </div>
        </div>
      </aside>

      {/* Main */}
      <div className="flex-1 flex flex-col min-w-0">
        {/* Topbar */}
        <header className="h-[60px] border-b border-neutral-100 flex items-center justify-between px-4 lg:px-6 bg-white sticky top-0 z-30">
          <div className="flex items-center gap-4">
            <button onClick={() => setMobileMenuOpen(true)} className="lg:hidden p-2 hover:bg-neutral-50 rounded-lg transition-colors">
              <Menu size={18} className="text-neutral-500" />
            </button>
            <h2 className="text-[1.1rem] font-semibold text-neutral-900 tracking-tight">
              {location.pathname === '/' ? 'Dashboard' : navItems.find(n => n.href === location.pathname)?.label || 'Pension Manager'}
            </h2>
          </div>
          <div className="flex items-center gap-3">
            <div className="relative hidden md:block">
              <Search size={15} className="absolute left-3 top-1/2 -translate-y-1/2 text-neutral-300" />
              <input
                type="text"
                placeholder="Search..."
                className="pl-9 pr-3.5 py-2 bg-neutral-50 border border-neutral-200 rounded-xl text-sm w-56 focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all placeholder:text-neutral-300"
              />
            </div>
            <button className="relative p-2 hover:bg-neutral-50 rounded-lg transition-colors">
              <Bell size={17} className="text-neutral-400" />
              <span className="absolute top-1.5 right-1.5 w-1.5 h-1.5 bg-red-500 rounded-full animate-pulse" />
            </button>
            <div className="relative" ref={userMenuRef}>
              <button onClick={() => setUserMenuOpen(!userMenuOpen)} className="flex items-center gap-2 p-1.5 hover:bg-neutral-50 rounded-lg transition-colors">
                <div className="w-7 h-7 bg-neutral-900 rounded-full flex items-center justify-center">
                  <span className="text-xs font-medium text-white">{user?.name?.[0]?.toUpperCase()}</span>
                </div>
                <ChevronDown size={13} className="text-neutral-300 transition-transform duration-200" style={{ transform: userMenuOpen ? 'rotate(180deg)' : 'rotate(0)' }} />
              </button>
              {userMenuOpen && (
                <div className="absolute right-0 mt-2 w-48 bg-white rounded-xl shadow-lg border border-neutral-100 py-2 animate-scale-in">
                  <div className="px-4 py-2.5 border-b border-neutral-50">
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

        {/* Content */}
        <main className="flex-1 p-6 lg:p-8 overflow-auto">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
