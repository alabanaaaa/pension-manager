import { useState, useEffect } from 'react';
import { sms } from '../lib/api';
import { Loader2, Send, MessageSquare, BarChart3, CheckCircle, AlertCircle } from 'lucide-react';

export default function SMSPage() {
  const [provider, setProvider] = useState('');
  const [balance, setBalance] = useState(null);
  const [loading, setLoading] = useState(true);
  const [sending, setSending] = useState(false);
  const [message, setMessage] = useState('');
  const [phone, setPhone] = useState('');
  const [result, setResult] = useState(null);

  useEffect(() => {
    Promise.all([
      sms.getProvider().then(r => setProvider(r.data?.provider || '')).catch(() => {}),
      sms.getBalance().then(r => setBalance(r.data?.balance ?? null)).catch(() => {}),
    ]).finally(() => setLoading(false));
  }, []);

  const handleSend = async () => {
    if (!phone || !message) return;
    setSending(true);
    setResult(null);
    try {
      const res = await sms.send({ to: phone, message });
      setResult({ success: true, data: res.data });
    } catch (err) {
      setResult({ success: false, error: err.response?.data?.error || 'Failed to send' });
    }
    finally { setSending(false); }
  };

  if (loading) {
    return (
      <div className="p-16 text-center">
        <Loader2 size={24} className="animate-spin mx-auto text-neutral-300" />
        <p className="text-sm text-neutral-400 mt-3">Loading SMS data...</p>
      </div>
    );
  }

  return (
    <div className="space-y-8 animate-fade-in-up">
      <div>
        <h1 className="text-3xl font-semibold tracking-tight text-neutral-900">SMS Gateway</h1>
        <p className="text-neutral-500 mt-2 text-base">Send bulk messages and notifications</p>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-6">
        <div className="bg-white rounded-2xl border border-neutral-50 p-6 flex items-center gap-4">
          <div className="p-2.5 rounded-xl bg-blue-50"><MessageSquare size={20} className="text-blue-600" /></div>
          <div>
            <p className="text-sm text-neutral-500">Provider</p>
            <p className="text-lg font-semibold text-neutral-900 capitalize">{provider || 'Not configured'}</p>
          </div>
        </div>
        <div className="bg-white rounded-2xl border border-neutral-50 p-6 flex items-center gap-4">
          <div className="p-2.5 rounded-xl bg-emerald-50"><BarChart3 size={20} className="text-emerald-600" /></div>
          <div>
            <p className="text-sm text-neutral-500">Balance</p>
            <p className="text-lg font-semibold text-neutral-900">{balance !== null ? `KES ${balance}` : '—'}</p>
          </div>
        </div>
        <div className="bg-white rounded-2xl border border-neutral-50 p-6 flex items-center gap-4">
          <div className="p-2.5 rounded-xl bg-violet-50"><Send size={20} className="text-violet-600" /></div>
          <div>
            <p className="text-sm text-neutral-500">Status</p>
            <p className="text-lg font-semibold text-neutral-900">{provider ? 'Active' : 'Mock Mode'}</p>
          </div>
        </div>
      </div>

      {/* Send Message */}
      <div className="bg-white rounded-2xl border border-neutral-50 overflow-hidden">
        <div className="flex items-center justify-between px-6 py-5 border-b border-neutral-50">
          <h2 className="text-lg font-semibold tracking-tight text-neutral-900">Send Message</h2>
        </div>
        <div className="p-6 space-y-6">
          <div>
            <label className="block text-sm font-medium text-neutral-700 mb-2">Phone Number</label>
            <input
              type="tel"
              value={phone}
              onChange={e => setPhone(e.target.value)}
              placeholder="+254712345678"
              className="w-full px-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all placeholder:text-neutral-300"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-neutral-700 mb-2">Message</label>
            <textarea
              value={message}
              onChange={e => setMessage(e.target.value)}
              placeholder="Enter your message..."
              rows={4}
              className="w-full px-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 resize-none transition-all placeholder:text-neutral-300"
            />
          </div>
          <button
            onClick={handleSend}
            disabled={sending || !phone || !message}
            className="btn-hover flex items-center gap-2 px-6 py-3.5 bg-neutral-900 text-white rounded-xl text-sm font-medium hover:bg-neutral-800 disabled:opacity-50 transition-all"
          >
            {sending ? <Loader2 size={16} className="animate-spin" /> : <Send size={16} />}
            Send Message
          </button>

          {result && (
            <div className={`p-4 rounded-xl flex items-center gap-3 ${result.success ? 'bg-emerald-50 text-emerald-700' : 'bg-red-50 text-red-700'}`}>
              {result.success ? <CheckCircle size={18} /> : <AlertCircle size={18} />}
              <span className="text-sm">{result.success ? 'Message sent successfully' : result.error}</span>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
