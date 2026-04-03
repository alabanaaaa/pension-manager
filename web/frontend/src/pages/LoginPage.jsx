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
    <div className="min-h-screen bg-neutral-50 flex flex-col items-center justify-center px-6 py-12">
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
          <h2 className="text-3xl font-semibold text-neutral-900 tracking-tight">Welcome back</h2>
          <p className="text-sm text-neutral-500 mt-2">Sign in to continue to your account</p>
        </div>

        {/* Card */}
        <div className="bg-white rounded-2xl border border-neutral-200 p-8 animate-fade-in-up stagger-2">
          <form onSubmit={handleSubmit} className="space-y-6">
            {(formError || error) && (
              <div className="flex items-center gap-2.5 p-3.5 bg-red-50 text-red-600 rounded-xl text-sm animate-fade-in">
                <AlertCircle size={15} className="flex-shrink-0" />
                <span>{formError || error}</span>
              </div>
            )}

            <div className="animate-fade-in-up stagger-3">
              <label className="block text-sm font-medium text-neutral-700 mb-2">Email</label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="you@example.com"
                className="input-focus w-full px-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all placeholder:text-neutral-400"
                required
              />
            </div>

            <div className="animate-fade-in-up stagger-4">
              <label className="block text-sm font-medium text-neutral-700 mb-2">Password</label>
              <div className="relative">
                <input
                  type={showPassword ? 'text' : 'password'}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="Enter your password"
                  className="input-focus w-full px-4 py-3.5 pr-12 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all placeholder:text-neutral-400"
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

            <div className="flex items-center justify-between animate-fade-in-up stagger-5">
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
              className="btn-hover w-full bg-neutral-900 text-white py-3.5 rounded-xl text-sm font-medium hover:bg-neutral-800 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2 mt-8"
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
        </div>

        {/* Footer */}
        <div className="mt-8 animate-fade-in-up stagger-5">
          <p className="text-center text-sm text-neutral-400">
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
