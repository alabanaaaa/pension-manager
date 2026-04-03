import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { Eye, EyeOff, Loader2, AlertCircle, ArrowRight } from 'lucide-react';
import bankLogo from '/bank-logo.svg';

export default function LoginPage() {
  const { login, error } = useAuth();
  const navigate = useNavigate();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [loading, setLoading] = useState(false);
  const [formError, setFormError] = useState('');

  const handleSubmit = async (e) => {
    e.preventDefault();
    setFormError('');
    if (!email || !password) {
      setFormError('Please enter both email and password');
      return;
    }
    setLoading(true);
    try {
      await login(email, password);
      navigate('/');
    } catch {
      // Error handled in context
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
        <div className="relative z-10 flex flex-col justify-center px-16 text-white">
          <div className="mb-12">
            <img src={bankLogo} alt="Logo" className="w-14 h-14 opacity-90" />
          </div>
          <h2 className="text-5xl font-light leading-tight tracking-tight mb-6">
            Pension Fund<br />
            <span className="font-semibold">Management</span>
          </h2>
          <p className="text-neutral-400 text-lg max-w-md leading-relaxed">
            Securely manage contributions, claims, voting, and member benefits — all in one platform.
          </p>
          <div className="mt-16 space-y-6">
            {[
              { label: 'Event-sourced audit trails', sub: 'Tamper-proof hash-chained records' },
              { label: 'Maker-checker workflow', sub: 'Every change requires approval' },
              { label: 'Multi-channel voting', sub: 'Web, USSD, and URL-based voting' },
            ].map((item, i) => (
              <div key={i} className="flex items-start gap-4">
                <div className="w-1 h-1 rounded-full bg-blue-400 mt-2.5 flex-shrink-0" />
                <div>
                  <p className="text-sm font-medium">{item.label}</p>
                  <p className="text-xs text-neutral-500 mt-0.5">{item.sub}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Right - Form */}
      <div className="flex-1 flex flex-col items-center justify-center px-8 py-12 lg:px-16">
        <div className="w-full max-w-sm">
          {/* Mobile logo */}
          <div className="lg:hidden mb-12">
            <img src={bankLogo} alt="Logo" className="w-12 h-12" />
          </div>

          <div className="mb-12 animate-fade-in-up">
            <h1 className="text-3xl font-semibold tracking-tight text-neutral-900">Welcome back</h1>
            <p className="text-neutral-500 mt-2 text-base">Enter your credentials to access your account</p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-8">
            {(formError || error) && (
              <div className="flex items-center gap-2 p-3 bg-red-50 text-red-600 rounded-lg text-sm animate-fade-in">
                <AlertCircle size={15} className="flex-shrink-0" />
                <span>{formError || error}</span>
              </div>
            )}

            <div className="space-y-6 animate-fade-in-up stagger-1">
              <div>
                <label className="block text-xs font-medium text-neutral-500 uppercase tracking-wider mb-2">Email</label>
                <input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="you@example.com"
                  className="w-full px-0 py-3 bg-transparent border-b border-neutral-200 text-base focus:outline-none focus:border-neutral-900 transition-colors placeholder:text-neutral-300"
                  required
                />
              </div>

              <div>
                <label className="block text-xs font-medium text-neutral-500 uppercase tracking-wider mb-2">Password</label>
                <div className="relative">
                  <input
                    type={showPassword ? 'text' : 'password'}
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    placeholder="Enter your password"
                    className="w-full px-0 py-3 pr-10 bg-transparent border-b border-neutral-200 text-base focus:outline-none focus:border-neutral-900 transition-colors placeholder:text-neutral-300"
                    required
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-0 top-1/2 -translate-y-1/2 text-neutral-300 hover:text-neutral-600 p-1.5 transition-colors"
                  >
                    {showPassword ? <EyeOff size={16} /> : <Eye size={16} />}
                  </button>
                </div>
              </div>
            </div>

            <div className="flex items-center justify-between animate-fade-in-up stagger-2">
              <label className="flex items-center gap-2 cursor-pointer">
                <input type="checkbox" className="w-4 h-4 rounded border-neutral-300 text-neutral-900 focus:ring-neutral-900/20" />
                <span className="text-sm text-neutral-500">Remember me</span>
              </label>
              <Link to="/forgot-password" className="text-sm text-neutral-900 hover:underline font-medium">
                Forgot password?
              </Link>
            </div>

            <button
              type="submit"
              disabled={loading}
              className="btn-hover w-full bg-neutral-900 text-white py-4 rounded-xl text-sm font-medium hover:bg-neutral-800 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2 mt-10"
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

          <p className="text-center text-sm text-neutral-400 mt-10 animate-fade-in-up stagger-3">
            Don't have an account?{' '}
            <Link to="/register" className="text-neutral-900 hover:underline font-medium">
              Create one
            </Link>
          </p>
          <p className="text-center text-xs text-neutral-300 mt-6">
            Powered by <span className="font-medium text-neutral-400">minidb</span>
          </p>
        </div>
      </div>
    </div>
  );
}
