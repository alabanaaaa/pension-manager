import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { auth } from '../lib/api';
import { Eye, EyeOff, Loader2, AlertCircle, ArrowRight, User, Shield } from 'lucide-react';
import bankLogo from '/bank-logo.svg';

export default function LoginPage() {
  const { setUser } = useAuth();
  const navigate = useNavigate();
  const [loginType, setLoginType] = useState('admin');
  const [identifier, setIdentifier] = useState('');
  const [password, setPassword] = useState('');
  const [pin, setPin] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [loading, setLoading] = useState(false);
  const [formError, setFormError] = useState('');

  const handleSubmit = async (e) => {
    e.preventDefault();
    setFormError('');

    const currentIdentifier = identifier.trim();
    const currentPassword = password;
    const currentPin = pin;

    if (loginType === 'admin') {
      if (!currentIdentifier || !currentPassword) {
        setFormError('Please enter email and password');
        return;
      }
    } else {
      if (!currentIdentifier || !currentPin) {
        setFormError('Please enter member number and PIN');
        return;
      }
    }

    setLoading(true);
    try {
      if (loginType === 'admin') {
        console.log('Admin login attempt:', { email: currentIdentifier, password: '***' });
        const response = await auth.login(currentIdentifier, currentPassword);
        const { access_token, refresh_token, user_id, name, role, scheme_id } = response.data;
        
        localStorage.setItem('access_token', access_token);
        localStorage.setItem('refresh_token', refresh_token);
        
        const userData = { id: user_id, name, email: currentIdentifier, role, scheme_id };
        setUser(userData);
        localStorage.setItem('user', JSON.stringify(userData));
        
        navigate('/');
      } else {
        console.log('Member login attempt:', { member_no: currentIdentifier, pin: '***' });
        const response = await auth.memberLogin(currentIdentifier, currentPin);
        console.log('Member login success:', response.data);
        const { access_token, refresh_token, member_id, name, role, scheme_id } = response.data;
        
        localStorage.setItem('access_token', access_token);
        localStorage.setItem('refresh_token', refresh_token);
        
        const userData = { id: member_id, name, email: '', role, scheme_id, isMember: true };
        setUser(userData);
        localStorage.setItem('user', JSON.stringify(userData));
        
        navigate('/portal');
      }
    } catch (err) {
      console.error('Login error:', err);
      setFormError(err.response?.data?.error || 'Login failed. Please check your credentials.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-white flex">
      {/* Left - Branding */}
      <div className="hidden lg:flex lg:w-1/2 bg-neutral-950 relative overflow-hidden">
        <div className="absolute inset-0">
          <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-blue-600/10 rounded-full blur-3xl" />
          <div className="absolute bottom-1/4 right-1/4 w-80 h-80 bg-indigo-600/10 rounded-full blur-3xl" />
        </div>
        <div className="relative z-10 flex flex-col justify-center px-20 text-white">
          <div className="mb-16">
            <img src={bankLogo} alt="Logo" className="w-16 h-16 opacity-90" />
          </div>
          <h2 className="text-6xl font-light leading-[1.15] tracking-tight mb-8">
            Pension Fund<br />
            <span className="font-semibold">Management</span>
          </h2>
          <p className="text-neutral-400 text-lg max-w-md leading-relaxed">
            Securely manage contributions, claims, voting, and member benefits — all in one platform.
          </p>
          <div className="mt-20 space-y-8">
            {[
              { label: 'Event-sourced audit trails', sub: 'Tamper-proof hash-chained records' },
              { label: 'Maker-checker workflow', sub: 'Every change requires approval' },
              { label: 'Multi-channel voting', sub: 'Web, USSD, and URL-based voting' },
            ].map((item, i) => (
              <div key={i} className="flex items-start gap-5">
                <div className="w-1.5 h-1.5 rounded-full bg-blue-400 mt-3 flex-shrink-0" />
                <div>
                  <p className="text-sm font-medium">{item.label}</p>
                  <p className="text-xs text-neutral-500 mt-1">{item.sub}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Right - Form */}
      <div className="flex-1 flex flex-col items-center justify-center px-10 py-16 lg:px-20">
        <div className="w-full max-w-sm">
          {/* Mobile logo */}
          <div className="lg:hidden mb-16">
            <img src={bankLogo} alt="Logo" className="w-14 h-14" />
          </div>

          <div className="mb-10 animate-fade-in-up">
            <h1 className="text-4xl font-semibold tracking-tight text-neutral-900">Welcome back</h1>
            <p className="text-neutral-500 mt-3 text-base">Sign in to your account</p>
          </div>

          {/* Login Type Tabs */}
          <div className="flex mb-8 bg-neutral-100 rounded-xl p-1">
            <button
              type="button"
              onClick={() => { setLoginType('admin'); setFormError(''); setIdentifier(''); setPassword(''); setPin(''); }}
              className={`flex-1 flex items-center justify-center gap-2 py-3 px-4 rounded-lg text-sm font-medium transition-all ${
                loginType === 'admin'
                  ? 'bg-white text-neutral-900 shadow-sm'
                  : 'text-neutral-500 hover:text-neutral-700'
              }`}
            >
              <Shield size={16} />
              Admin
            </button>
            <button
              type="button"
              onClick={() => { setLoginType('member'); setFormError(''); setIdentifier(''); setPassword(''); setPin(''); }}
              className={`flex-1 flex items-center justify-center gap-2 py-3 px-4 rounded-lg text-sm font-medium transition-all ${
                loginType === 'member'
                  ? 'bg-white text-neutral-900 shadow-sm'
                  : 'text-neutral-500 hover:text-neutral-700'
              }`}
            >
              <User size={16} />
              Member
            </button>
          </div>

          <form onSubmit={handleSubmit} className="space-y-6">
            {formError && (
              <div className="flex items-center gap-2.5 p-4 bg-red-50 text-red-600 rounded-xl text-sm animate-fade-in">
                <AlertCircle size={16} className="flex-shrink-0" />
                <span>{formError}</span>
              </div>
            )}

            <div className="space-y-6 animate-fade-in-up stagger-1">
              <div>
                <label className="block text-xs font-medium text-neutral-500 uppercase tracking-widest mb-3">
                  {loginType === 'admin' ? 'Email' : 'Member Number'}
                </label>
                <input
                  type={loginType === 'admin' ? 'email' : 'text'}
                  value={identifier}
                  onChange={(e) => setIdentifier(e.target.value)}
                  placeholder={loginType === 'admin' ? 'you@example.com' : 'e.g., MEM001'}
                  className="w-full px-0 py-4 bg-transparent border-b border-neutral-200 text-base focus:outline-none focus:border-neutral-900 transition-colors placeholder:text-neutral-300"
                  required
                />
              </div>

              {loginType === 'admin' ? (
                <div>
                  <label className="block text-xs font-medium text-neutral-500 uppercase tracking-widest mb-3">Password</label>
                  <div className="relative">
                    <input
                      type={showPassword ? 'text' : 'password'}
                      value={password}
                      onChange={(e) => setPassword(e.target.value)}
                      placeholder="Enter your password"
                      className="w-full px-0 py-4 pr-12 bg-transparent border-b border-neutral-200 text-base focus:outline-none focus:border-neutral-900 transition-colors placeholder:text-neutral-300"
                      required
                    />
                    <button
                      type="button"
                      onClick={() => setShowPassword(!showPassword)}
                      className="absolute right-0 top-1/2 -translate-y-1/2 text-neutral-300 hover:text-neutral-600 p-2 transition-colors"
                    >
                      {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
                    </button>
                  </div>
                </div>
              ) : (
                <div>
                  <label className="block text-xs font-medium text-neutral-500 uppercase tracking-widest mb-3">PIN</label>
                  <div className="relative">
                    <input
                      type={showPassword ? 'text' : 'password'}
                      value={pin}
                      onChange={(e) => setPin(e.target.value)}
                      placeholder="Enter your PIN"
                      maxLength={10}
                      className="w-full px-0 py-4 pr-12 bg-transparent border-b border-neutral-200 text-base focus:outline-none focus:border-neutral-900 transition-colors placeholder:text-neutral-300"
                      required
                    />
                    <button
                      type="button"
                      onClick={() => setShowPassword(!showPassword)}
                      className="absolute right-0 top-1/2 -translate-y-1/2 text-neutral-300 hover:text-neutral-600 p-2 transition-colors"
                    >
                      {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
                    </button>
                  </div>
                </div>
              )}
            </div>

            <button
              type="submit"
              disabled={loading}
              className="w-full bg-neutral-900 text-white py-4 rounded-xl text-sm font-medium hover:bg-neutral-800 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2.5 mt-8 transition-all hover:scale-[1.02] active:scale-[0.98]"
            >
              {loading ? (
                <>
                  <Loader2 size={16} className="animate-spin" />
                  Signing in...
                </>
              ) : (
                <>
                  Sign In
                  <ArrowRight size={16} />
                </>
              )}
            </button>
          </form>

          <p className="text-center text-xs text-neutral-300 mt-8">
            Powered by <span className="font-medium text-neutral-400">Pension Manager</span>
          </p>
        </div>
      </div>
    </div>
  );
}
