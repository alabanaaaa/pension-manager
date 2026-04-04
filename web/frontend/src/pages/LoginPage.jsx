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
    <div className="min-h-screen bg-neutral-50 flex items-center justify-center px-6 py-12">
      <div className="w-full max-w-sm animate-fade-in-up">
        {/* Logo */}
        <div className="flex items-center gap-3 mb-12">
          <img src={bankLogo} alt="Logo" className="w-10 h-10" />
          <div>
            <h1 className="text-lg font-semibold text-neutral-900 tracking-tight">Pension Manager</h1>
            <p className="text-xs text-neutral-400">Fund Management System</p>
          </div>
        </div>

        {/* Heading */}
        <div className="mb-10 animate-fade-in-up stagger-1">
          <h2 className="text-3xl font-semibold tracking-tight text-neutral-900">Welcome back</h2>
          <p className="text-neutral-500 mt-2 text-base">Enter your credentials to access your account</p>
        </div>

        {/* Card */}
        <div className="bg-white rounded-2xl border border-neutral-200 overflow-hidden">
          {/* Header */}
          <div className="bg-gradient-to-r from-blue-600 to-indigo-600 px-8 py-10 text-center">
            <div className="inline-flex items-center justify-center w-16 h-16 bg-white/20 rounded-2xl mb-4 backdrop-blur-sm">
              <img src={bankLogo} alt="Logo" className="w-10 h-10" />
            </div>
            <h1 className="text-2xl font-bold text-white">Pension Manager</h1>
            <p className="text-blue-100 mt-1 text-sm">Sign in to your account</p>
          </div>

          {/* Form */}
          <div className="px-8 py-8">
            <form onSubmit={handleSubmit} className="space-y-6">
              {(formError || error) && (
                <div className="flex items-center gap-2.5 p-4 bg-red-50 text-red-600 rounded-xl text-sm animate-fade-in">
                  <AlertCircle size={16} className="flex-shrink-0" />
                  <span>{formError || error}</span>
                </div>
              )}

              <div className="animate-fade-in-up stagger-2">
                <label className="block text-sm font-medium text-neutral-700 mb-2">Email Address</label>
                <input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="admin@pension.go.ke"
                  className="w-full px-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all placeholder:text-neutral-400"
                  required
                />
              </div>

              <div className="animate-fade-in-up stagger-3">
                <label className="block text-sm font-medium text-neutral-700 mb-2">Password</label>
                <div className="relative">
                  <input
                    type={showPassword ? 'text' : 'password'}
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    placeholder="Enter your password"
                    className="w-full px-4 py-3.5 pr-12 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all placeholder:text-neutral-400"
                    required
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-neutral-400 hover:text-neutral-600 p-1.5 transition-colors"
                  >
                    {showPassword ? <EyeOff size={16} /> : <Eye size={16} />}
                  </button>
                </div>
              </div>

              <div className="flex items-center justify-between animate-fade-in-up stagger-4">
                <label className="flex items-center gap-2.5 cursor-pointer">
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
                className="btn-hover w-full bg-gradient-to-r from-blue-600 to-indigo-600 text-white py-3.5 rounded-xl text-sm font-medium hover:from-blue-700 hover:to-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2 mt-8 shadow-lg shadow-blue-600/25"
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

            <div className="mt-8 pt-6 border-t border-neutral-100 text-center">
              <p className="text-sm text-neutral-500">
                Don't have an account?{' '}
                <Link to="/register" className="text-blue-600 hover:underline font-semibold">
                  Create account
                </Link>
              </p>
            </div>
          </div>
        </div>

        <p className="text-center text-xs text-neutral-300 mt-8 animate-fade-in-up stagger-5">
          Powered by <span className="font-medium text-neutral-400">minidb</span>
        </p>
      </div>
    </div>
  );
}
