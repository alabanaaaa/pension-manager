import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { auth } from '../lib/api';
import { useAuth } from '../context/AuthContext';
import { Loader2, User, Lock, LogIn } from 'lucide-react';

export default function MemberLoginPage() {
  const navigate = useNavigate();
  const { setUser } = useAuth();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [form, setForm] = useState({
    member_no: '',
    pin: '',
  });

  const handleChange = (e) => {
    setForm(prev => ({ ...prev, [e.target.name]: e.target.value }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      const res = await auth.memberLogin(form.member_no, form.pin);
      const { access_token, refresh_token, member_id, name, role, scheme_id } = res.data;
      
      localStorage.setItem('access_token', access_token);
      localStorage.setItem('refresh_token', refresh_token);
      
      const userData = { id: member_id, name, email: '', role, scheme_id, isMember: true };
      setUser(userData);
      localStorage.setItem('user', JSON.stringify(userData));
      
      navigate('/portal');
    } catch (err) {
      setError(err.response?.data?.error || 'Login failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', padding: '16px', backgroundColor: '#f7f7f7' }}>
      <div style={{ width: '100%', maxWidth: '448px' }}>
        <div style={{ textAlign: 'center', marginBottom: '32px' }}>
          <div style={{ width: '64px', height: '64px', backgroundColor: '#000', borderRadius: '16px', display: 'flex', alignItems: 'center', justifyContent: 'center', margin: '0 auto 16px' }}>
            <User size={32} color="#fff" />
          </div>
          <h1 style={{ fontSize: '24px', fontWeight: 600, color: '#000', marginBottom: '8px' }}>Member Portal</h1>
          <p style={{ color: '#666' }}>Sign in with your member number and PIN</p>
        </div>

        <div style={{ backgroundColor: '#fff', borderRadius: '16px', padding: '32px', border: '1px solid #e5e5e5' }}>
          {error && (
            <div style={{ marginBottom: '24px', padding: '16px', backgroundColor: '#fef2f2', borderRadius: '12px', color: '#dc2626', fontSize: '14px' }}>
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
            <div>
              <label style={{ display: 'block', fontSize: '14px', fontWeight: 500, color: '#374151', marginBottom: '6px' }}>Member Number</label>
              <div style={{ position: 'relative' }}>
                <User size={18} style={{ position: 'absolute', left: '12px', top: '50%', transform: 'translateY(-50%)', color: '#d1d5db' }} />
                <input
                  type="text"
                  name="member_no"
                  value={form.member_no}
                  onChange={handleChange}
                  required
                  style={{ width: '100%', paddingLeft: '40px', padding: '12px', border: '1px solid #e5e5e5', borderRadius: '12px', fontSize: '14px' }}
                  placeholder="e.g., MEM001"
                />
              </div>
            </div>

            <div>
              <label style={{ display: 'block', fontSize: '14px', fontWeight: 500, color: '#374151', marginBottom: '6px' }}>PIN</label>
              <div style={{ position: 'relative' }}>
                <Lock size={18} style={{ position: 'absolute', left: '12px', top: '50%', transform: 'translateY(-50%)', color: '#d1d5db' }} />
                <input
                  type="password"
                  name="pin"
                  value={form.pin}
                  onChange={handleChange}
                  required
                  style={{ width: '100%', paddingLeft: '40px', padding: '12px', border: '1px solid #e5e5e5', borderRadius: '12px', fontSize: '14px' }}
                  placeholder="Enter your PIN"
                  maxLength={6}
                />
              </div>
            </div>

            <button
              type="submit"
              disabled={loading}
              style={{ width: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '8px', padding: '12px', backgroundColor: '#000', color: '#fff', borderRadius: '12px', fontSize: '14px', fontWeight: 500, opacity: loading ? 0.5 : 1, cursor: loading ? 'not-allowed' : 'pointer' }}
            >
              {loading ? <Loader2 size={18} style={{ animation: 'spin 1s linear infinite' }} /> : <LogIn size={18} />}
              Sign In
            </button>
          </form>

          <div style={{ marginTop: '24px', paddingTop: '24px', borderTop: '1px solid #f3f4f6', textAlign: 'center' }}>
            <p style={{ fontSize: '14px', color: '#6b7280' }}>
              Admin login?{' '}
              <Link to="/login" style={{ color: '#000', fontWeight: 500 }}>
                Click here
              </Link>
            </p>
          </div>
        </div>

        <p style={{ textAlign: 'center', fontSize: '14px', color: '#9ca3af', marginTop: '24px' }}>
          Forgot your PIN? Contact your scheme administrator
        </p>
      </div>
    </div>
  );
}
