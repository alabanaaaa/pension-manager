import { useState, useEffect, useRef } from 'react';
import { Outlet, Link, useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import {
  LayoutDashboard, Users, CreditCard, FileText, Vote,
  Building2, Shield, BarChart3, Settings, LogOut,
  Menu, X, Bell, Search, Hospital, MessageSquare,
  TrendingUp, UserCheck, ClipboardList, Newspaper,
  Lock, Send, UserCircle
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
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const sidebarRef = useRef(null);

  const navItems = isAdmin || isOfficer ? adminNav : memberNav;

  // Close sidebar when clicking outside
  useEffect(() => {
    const handleClickOutside = (e) => {
      if (sidebarOpen && sidebarRef.current && !sidebarRef.current.contains(e.target)) {
        setSidebarOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [sidebarOpen]);

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <div className="min-h-screen bg-white flex">
      {/* Overlay when sidebar is open */}
      {sidebarOpen && (
        <div className="fixed inset-0 z-30 bg-black/10 backdrop-blur-[2px] transition-opacity duration-300" onClick={() => setSidebarOpen(false)} />
      )}

      {/* Sidebar */}
      <aside
        ref={sidebarRef}
        className={`
          fixed top-0 left-0 bottom-0 z-40
          ${sidebarOpen ? 'translate-x-0 w-64' : '-translate-x-full w-64'}
          bg-neutral-950 text-white transition-all duration-300 ease-in-out flex flex-col shadow-2xl
        `}
      >
        {/* Brand */}
        <div className="px-5 py-5 flex items-center justify-between border-b border-white/10">
          <div className="flex items-center gap-3">
            <img src={bankLogo} alt="Logo" className="w-8 h-8 opacity-90" />
            <div>
              <h1 className="text-sm font-semibold tracking-tight">Pension Manager</h1>
              <p className="text-[10px] text-neutral-500 uppercase tracking-wider">Admin Portal</p>
            </div>
          </div>
          <button onClick={() => setSidebarOpen(false)} className="p-1.5 hover:bg-white/10 rounded-lg transition-colors">
            <X size={16} className="text-neutral-400" />
          </button>
        </div>

        {/* Nav */}
        <nav className="flex-1 overflow-y-auto py-4 px-3 space-y-0.5">
          {navItems.map((item, i) => {
            const isActive = location.pathname === item.href || (item.href !== '/' && location.pathname.startsWith(item.href));
            return (
              <Link
                key={item.href}
                to={item.href}
                onClick={() => { setSidebarOpen(false); setMobileMenuOpen(false); }}
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
                <span className="truncate">{item.label}</span>
              </Link>
            );
          })}
        </nav>

        {/* User */}
        <div className="p-3 border-t border-white/10">
          <div className="flex items-center gap-3 px-3 py-2.5 rounded-xl">
            <div className="w-8 h-8 bg-white rounded-full flex items-center justify-center flex-shrink-0">
              <span className="text-xs font-semibold text-neutral-950">{user?.name?.[0]?.toUpperCase()}</span>
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium truncate">{user?.name}</p>
              <p className="text-xs text-neutral-500 truncate capitalize">{user?.role?.replace('_', ' ')}</p>
            </div>
            <button onClick={handleLogout} className="text-neutral-500 hover:text-white p-1 transition-colors" title="Sign out">
              <LogOut size={16} />
            </button>
          </div>
        </div>
      </aside>

      {/* Main */}
      <div className="flex-1 flex flex-col min-w-0">
        {/* Topbar */}
        <header className="h-[60px] border-b border-neutral-100 flex items-center justify-between px-4 lg:px-6 bg-white sticky top-0 z-20">
          <div className="flex items-center gap-4">
            {/* Bank logo - click to open sidebar */}
            <button
              onClick={() => setSidebarOpen(!sidebarOpen)}
              className="p-1.5 hover:bg-neutral-50 rounded-xl transition-all duration-200 group"
              title="Toggle navigation"
            >
              <img
                src={bankLogo}
                alt="Menu"
                className="w-7 h-7 opacity-70 group-hover:opacity-100 group-hover:scale-105 transition-all duration-200"
              />
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
