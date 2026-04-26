import { useState, useEffect } from 'react';
import { portal } from '../lib/api';
import { Loader2, MessageSquare, Send, CheckCircle, Clock } from 'lucide-react';

export default function PortalFeedbackPage() {
  const [feedbacks, setFeedbacks] = useState([]);
  const [loading, setLoading] = useState(true);
  const [subject, setSubject] = useState('');
  const [message, setMessage] = useState('');
  const [sending, setSending] = useState(false);
  const [sent, setSent] = useState(false);

  useEffect(() => {
    portal.getFeedback()
      .then(r => setFeedbacks(Array.isArray(r.data) ? r.data : []))
      .catch(() => setFeedbacks([]))
      .finally(() => setLoading(false));
  }, []);

  const handleSubmit = async () => {
    if (!subject || !message) return;
    setSending(true);
    try {
      await portal.submitFeedback({ subject, message });
      setSent(true);
      setSubject('');
      setMessage('');
      setTimeout(() => setSent(false), 3000);
      // Refresh feedback list
      const r = await portal.getFeedback();
      setFeedbacks(Array.isArray(r.data) ? r.data : []);
    } catch (err) {
      console.error(err);
    }
    finally { setSending(false); }
  };

  const statusColors = {
    open: 'bg-blue-50 text-blue-700',
    in_progress: 'bg-amber-50 text-amber-700',
    resolved: 'bg-emerald-50 text-emerald-700',
  };

  return (
    <div className="space-y-8 animate-fade-in-up">
      <div>
        <h1 className="text-2xl font-bold tracking-tight text-black">Feedback</h1>
        <p className="text-neutral-500 mt-2 text-base">Submit feedback or inquiries</p>
      </div>

      {sent && (
        <div className="bg-emerald-50 border border-emerald-100 rounded-2xl p-5 flex items-center gap-3">
          <CheckCircle size={18} className="text-emerald-600" />
          <p className="text-sm text-emerald-700">Feedback submitted successfully</p>
        </div>
      )}

      {/* Submit Form */}
      <div className="bg-white rounded-2xl border border-neutral-50 overflow-hidden">
        <div className="flex items-center justify-between px-6 py-5 border-b border-neutral-50">
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-blue-50"><MessageSquare size={20} className="text-blue-600" /></div>
            <h2 className="text-lg font-semibold tracking-tight text-neutral-900">New Feedback</h2>
          </div>
        </div>
        <div className="p-6 space-y-6">
          <div>
            <label className="block text-sm font-medium text-neutral-700 mb-2">Subject</label>
            <input
              type="text"
              value={subject}
              onChange={e => setSubject(e.target.value)}
              placeholder="What is this about?"
              className="w-full px-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all placeholder:text-neutral-300"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-neutral-700 mb-2">Message</label>
            <textarea
              value={message}
              onChange={e => setMessage(e.target.value)}
              placeholder="Describe your feedback or inquiry..."
              rows={4}
              className="w-full px-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 resize-none transition-all placeholder:text-neutral-300"
            />
          </div>
          <button
            onClick={handleSubmit}
            disabled={sending || !subject || !message}
            className="btn-hover flex items-center gap-2 px-6 py-3.5 bg-neutral-900 text-white rounded-xl text-sm font-medium hover:bg-neutral-800 disabled:opacity-50 transition-all"
          >
            {sending ? <Loader2 size={16} className="animate-spin" /> : <Send size={16} />}
            Submit Feedback
          </button>
        </div>
      </div>

      {/* Feedback History */}
      <div className="bg-white rounded-2xl border border-neutral-50 overflow-hidden">
        <div className="flex items-center justify-between px-6 py-5 border-b border-neutral-50">
          <h2 className="text-lg font-semibold tracking-tight text-neutral-900">My Feedback</h2>
        </div>
        {loading ? (
          <div className="p-16 text-center"><Loader2 size={24} className="animate-spin mx-auto text-neutral-300" /><p className="text-sm text-neutral-400 mt-3">Loading...</p></div>
        ) : feedbacks.length === 0 ? (
          <div className="p-16 text-center"><p className="text-neutral-500">No feedback submitted yet</p></div>
        ) : (
          <div className="divide-y divide-neutral-50">
            {feedbacks.map((fb, i) => (
              <div key={fb.id} className="px-6 py-5 hover:bg-neutral-50/50 transition-colors">
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <h3 className="font-medium text-neutral-900">{fb.subject}</h3>
                    <p className="text-sm text-neutral-500 mt-1">{fb.message}</p>
                    <p className="text-xs text-neutral-400 mt-2">{new Date(fb.created_at).toLocaleString()}</p>
                  </div>
                  <span className={`px-2.5 py-1 rounded-full text-xs font-medium capitalize ${statusColors[fb.status] || 'bg-neutral-50 text-neutral-600'}`}>
                    {fb.status?.replace('_', ' ')}
                  </span>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
