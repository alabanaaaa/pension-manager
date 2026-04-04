import { useState, useEffect, useRef } from 'react';
import { Outlet, Link, useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import {
  LayoutDashboard, Users, CreditCard, FileText, Vote,
  Building2, Shield, BarChart3, Settings, LogOut,
  Menu, X, Bell, Search, Hospital, MessageSquare,
  TrendingUp, UserCheck, ClipboardList, Newspaper,
  Lock, Send, UserCircle, ChevronDown
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
    <div className="min-h-screen bg-[#f5f6f8] flex">
      {/* Mobile overlay */}
      {sidebarOpen && (
        <div className="lg:hidden fixed inset-0 z-40 bg-black/30" onClick={() => setSidebarOpen(false)} />
      )}

      {/* Sidebar */}
      <aside className={`
        fixed lg:static inset-y-0 left-0 z-50
        ${sidebarOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'}
        w-[260px] bg-white border-r border-[#e8e9eb] transition-transform duration-300 flex flex-col
      `}>
        {/* Brand */}
        <div className="px-5 py-[20px] border-b border-[#e8e9eb] flex items-center gap-2.5">
          <div className="w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0 overflow-hidden">
            <img src={bankLogo} alt="Logo" className="w-8 h-8 object-contain" />
          </div>
          <h1 className="text-[1rem] font-bold text-neutral-900 tracking-tight">Pension Manager</h1>
        </div>

        {/* Nav */}
        <nav className="flex-1 overflow-y-auto py-4 px-3">
          {navItems.map((item) => {
            const isActive = location.pathname === item.href || (item.href !== '/' && location.pathname.startsWith(item.href));
            return (
              <Link
                key={item.href}
                to={item.href}
                onClick={() => setSidebarOpen(false)}
                className={`
                  flex items-center gap-2.5 px-3 py-[9px] rounded-lg text-[0.875rem] transition-all mb-0.5
                  ${isActive
                    ? 'bg-[#e8f5d8] text-[#6a9a2e] font-medium'
                    : 'text-[#6b7280] hover:bg-[#f5f6f8] hover:text-neutral-900'
                  }
                `}
              >
                <item.icon size={18} className="flex-shrink-0" />
                <span>{item.label}</span>
              </Link>
            );
          })}
        </nav>

        {/* User */}
        <div className="px-3 py-4 border-t border-[#e8e9eb]">
          <div className="flex items-center gap-2.5 px-2.5 py-2 rounded-lg">
            <div className="w-8 h-8 rounded-full bg-[#e8f5d8] text-[#6a9a2e] flex items-center justify-center flex-shrink-0 text-[0.7rem] font-semibold">
              {user?.name?.[0]?.toUpperCase()}
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-[0.8rem] font-medium text-neutral-900 truncate">{user?.name}</p>
              <p className="text-[0.65rem] text-[#9ca3af] capitalize">{user?.role?.replace('_', ' ')}</p>
            </div>
          </div>
        </div>
      </aside>

      {/* Main */}
      <div className="flex-1 flex flex-col min-w-0">
        {/* Topbar */}
        <header className="h-[60px] border-b border-[#e8e9eb] flex items-center justify-between px-6 bg-white sticky top-0 z-30">
          <div className="flex items-center gap-4">
            <button onClick={() => setSidebarOpen(true)} className="lg:hidden p-2 hover:bg-[#f5f6f8] rounded-lg">
              <Menu size={18} className="text-neutral-500" />
            </button>
            <h2 className="text-[1.1rem] font-semibold text-neutral-900 tracking-tight">
              {location.pathname === '/' ? 'Dashboard' : navItems.find(n => n.href === location.pathname)?.label || 'Pension Manager'}
            </h2>
          </div>
          <div className="flex items-center gap-3">
            <div className="relative hidden md:block">
              <Search size={15} className="absolute left-3 top-1/2 -translate-y-1/2 text-[#9ca3af]" />
              <input
                type="text"
                placeholder="Search..."
                className="pl-9 pr-3.5 py-2 bg-white border border-[#e8e9eb] rounded-lg text-sm w-56 focus:outline-none focus:border-[#82b440] focus:ring-2 focus:ring-[#e8f5d8] transition-all placeholder:text-[#9ca3af]"
              />
            </div>
            <button className="relative p-2 hover:bg-[#f5f6f8] rounded-lg transition-colors">
              <Bell size={17} className="text-[#9ca3af]" />
              <span className="absolute top-1.5 right-1.5 w-1.5 h-1.5 bg-red-500 rounded-full" />
            </button>
            <div className="relative" ref={userMenuRef}>
              <button onClick={() => setUserMenuOpen(!userMenuOpen)} className="flex items-center gap-2 p-1.5 hover:bg-[#f5f6f8] rounded-lg transition-colors">
                <div className="w-7 h-7 bg-neutral-900 rounded-full flex items-center justify-center">
                  <span className="text-xs font-medium text-white">{user?.name?.[0]?.toUpperCase()}</span>
                </div>
                <ChevronDown size={13} className="text-[#9ca3af]" />
              </button>
              {userMenuOpen && (
                <div className="absolute right-0 mt-2 w-48 bg-white rounded-xl shadow-lg border border-[#e8e9eb] py-2 animate-fade-in-up">
                  <div className="px-4 py-2.5 border-b border-[#e8e9eb]">
                    <p className="text-sm font-medium text-neutral-900">{user?.name}</p>
                    <p className="text-xs text-[#9ca3af]">{user?.email}</p>
                  </div>
                  <Link to="/settings" className="flex items-center gap-2 px-4 py-2 text-sm text-neutral-600 hover:bg-[#f5f6f8] transition-colors">
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
        <main className="flex-1 p-6 overflow-auto">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
