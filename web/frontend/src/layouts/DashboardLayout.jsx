import { useState, useEffect, useRef } from 'react';
import { Outlet, Link, useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import {
  LayoutDashboard, Users, CreditCard, FileText, Vote,
  Building2, Shield, BarChart3, Settings, LogOut,
  Menu, Bell, Search, Hospital, MessageSquare,
  TrendingUp, ClipboardList, Newspaper,
  Lock, Send, ChevronRight, X, Minus
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
  { icon: Users, label: 'My Profile', href: '/portal/profile' },
  { icon: CreditCard, label: 'Contributions', href: '/portal/contributions' },
  { icon: FileText, label: 'Claims', href: '/portal/claims' },
  { icon: Vote, label: 'Voting', href: '/portal/voting' },
  { icon: TrendingUp, label: 'Projections', href: '/portal/projections' },
  { icon: Newspaper, label: 'News', href: '/portal/news' },
  { icon: MessageSquare, label: 'Feedback', href: '/portal/feedback' },
  { icon: Settings, label: 'Settings', href: '/portal/settings' },
];

export default function DashboardLayout() {
  const { user, logout, isAdmin, isOfficer } = useAuth();
  const location = useLocation();
  const navigate = useNavigate();
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const [activeDropdown, setActiveDropdown] = useState(null);
  const sidebarRef = useRef(null);

  const navItems = isAdmin || isOfficer ? adminNav : memberNav;

  useEffect(() => {
    const handleClickOutside = (e) => {
      if (sidebarOpen && sidebarRef.current && !sidebarRef.current.contains(e.target)) {
        setSidebarOpen(false);
      }
      if (activeDropdown) {
        setActiveDropdown(null);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [sidebarOpen, activeDropdown]);

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const getPageTitle = () => {
    const path = location.pathname;
    if (path === '/') return 'Dashboard';
    if (path === '/portal') return 'My Dashboard';
    const match = navItems.find(n => n.href === path || (n.href !== '/' && path.startsWith(n.href)));
    return match?.label || 'Pension Manager';
  };

  return (
    <div className="min-h-screen bg-white flex">
      {sidebarOpen && (
        <div 
          className="fixed inset-0 z-30 bg-black/40 backdrop-blur-sm transition-opacity" 
          onClick={() => setSidebarOpen(false)} 
        />
      )}

      {/* Sidebar - Uber Bold Black */}
      <aside
        ref={sidebarRef}
        className={`
          fixed top-0 left-0 bottom-0 z-40 w-60
          ${sidebarOpen ? 'translate-x-0' : '-translate-x-full'}
          bg-black text-white transition-transform duration-200 ease-out flex flex-col
        `}
      >
        {/* Logo */}
        <div className="px-5 py-4 flex items-center justify-between border-b border-white/10">
          <div className="flex items-center gap-3">
            <img src={bankLogo} alt="Logo" className="w-8 h-8" />
            <div>
              <h1 className="text-sm font-bold tracking-tight">PENSION</h1>
              <p className="text-[10px] text-gray-400 tracking-widest uppercase">Manager</p>
            </div>
          </div>
          <button 
            onClick={() => setSidebarOpen(false)} 
            className="p-1 hover:bg-white/10 rounded transition-colors"
          >
            <X size={16} className="text-gray-400" />
          </button>
        </div>

        {/* Navigation */}
        <nav className="flex-1 overflow-y-auto py-3 px-3">
          {navItems.map((item) => {
            const isActive = location.pathname === item.href || 
              (item.href !== '/' && item.href !== '/portal' && location.pathname.startsWith(item.href));
            return (
              <Link
                key={item.href}
                to={item.href}
                onClick={() => { setSidebarOpen(false); setMobileMenuOpen(false); }}
                className={`
                  flex items-center gap-3 px-3 py-2.5 mb-0.5 text-sm transition-all duration-150 rounded
                  ${isActive
                    ? 'bg-white text-black font-semibold'
                    : 'text-gray-300 hover:bg-white/10 hover:text-white'
                  }
                `}
              >
                <item.icon size={18} className="flex-shrink-0" />
                <span>{item.label}</span>
                {isActive && <ChevronRight size={14} className="ml-auto" />}
              </Link>
            );
          })}
        </nav>

        {/* User */}
        <div className="p-3 border-t border-white/10">
          <div className="flex items-center gap-3 px-3 py-3 rounded hover:bg-white/10 transition-colors cursor-pointer group">
            <div className="w-9 h-9 bg-white rounded-full flex items-center justify-center flex-shrink-0">
              <span className="text-xs font-bold text-black">{user?.name?.[0]?.toUpperCase()}</span>
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium truncate">{user?.name}</p>
              <p className="text-xs text-gray-400 truncate capitalize">{user?.role?.replace('_', ' ')}</p>
            </div>
          </div>
          <button 
            onClick={handleLogout} 
            className="w-full flex items-center gap-3 px-3 py-2.5 mt-1 text-sm text-gray-300 hover:bg-white/10 hover:text-white rounded transition-colors"
          >
            <LogOut size={16} />
            <span>Sign Out</span>
          </button>
        </div>
      </aside>

      {/* Main Content */}
      <div className="flex-1 flex flex-col min-w-0">
        {/* Header - Uber Minimal */}
        <header className="h-14 border-b border-gray-100 flex items-center justify-between px-6 bg-white/80 backdrop-blur-md sticky top-0 z-20">
          <div className="flex items-center gap-4">
            <button
              onClick={() => setSidebarOpen(true)}
              className="p-2 hover:bg-gray-50 rounded-lg transition-all active:scale-95"
              title="Open menu"
            >
              <img src={bankLogo} alt="Menu" className="w-6 h-6 opacity-60 hover:opacity-100 transition-opacity" />
            </button>
            <div className="flex items-center gap-2">
              <h2 className="text-base font-semibold text-black tracking-tight animate-fade-in">
                {getPageTitle()}
              </h2>
            </div>
          </div>
          
          <div className="flex items-center gap-2">
            {/* Search - Minimal */}
            <div className="relative hidden md:block group">
              <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-300 group-focus-within:text-gray-500 transition-colors" />
              <input
                type="text"
                placeholder="Search..."
                className="pl-8 pr-3 py-2 bg-gray-50 border border-gray-200 rounded-lg text-sm w-48 focus:outline-none focus:border-black focus:w-64 focus:bg-white transition-all placeholder:text-gray-400"
              />
            </div>
            
            {/* Notifications */}
            <button className="relative p-2 hover:bg-gray-50 rounded-lg transition-all active:scale-95">
              <Bell size={18} className="text-gray-500 hover:text-gray-700 transition-colors" />
              <span className="absolute top-1.5 right-1.5 w-2 h-2 bg-black rounded-full animate-pulse" />
            </button>
          </div>
        </header>

        {/* Page Content */}
        <main className="flex-1 p-6 overflow-auto">
          <div className="animate-fade-in-up">
            <Outlet />
          </div>
        </main>
      </div>
    </div>
  );
}
