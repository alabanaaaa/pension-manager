import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { pendingChanges } from '../lib/api';
import { Shield, CheckCircle, XCircle, Clock, Loader2, FileText, Users, ChevronRight } from 'lucide-react';

export default function MakerCheckerPage() {
  const [changes, setChanges] = useState([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState('pending');
  const [processing, setProcessing] = useState(null);
  const [reason, setReason] = useState('');
  const [showRejectModal, setShowRejectModal] = useState(null);

  useEffect(() => {
    fetchChanges();
  }, [filter]);

  const fetchChanges = async () => {
    setLoading(true);
    try {
      const res = await pendingChanges.list({ status: filter });
      const data = Array.isArray(res.data) ? res.data : [];
      setChanges(data);
    } catch { setChanges([]); }
    finally { setLoading(false); }
  };

  const handleApprove = async (id) => {
    setProcessing(id);
    try {
      await pendingChanges.approve(id, 'Approved');
      fetchChanges();
    } catch (err) {
      console.error(err);
    }
    finally { setProcessing(null); }
  };

  const handleReject = async (id) => {
    if (!reason.trim()) return;
    setProcessing(id);
    try {
      await pendingChanges.reject(id, reason);
      setShowRejectModal(null);
      setReason('');
      fetchChanges();
    } catch (err) {
      console.error(err);
    }
    finally { setProcessing(null); }
  };

  const typeIcons = {
    member: Users,
    beneficiary: Users,
    claim: FileText,
  };

  const statusColors = {
    pending: 'bg-amber-50 text-amber-700',
    approved: 'bg-emerald-50 text-emerald-700',
    rejected: 'bg-red-50 text-red-700',
  };

  return (
    <div className="space-y-8 animate-fade-in-up">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-semibold tracking-tight text-neutral-900">Maker-Checker</h1>
        <p className="text-neutral-500 mt-2 text-base">Review and approve pending changes</p>
      </div>

      {/* Filters */}
      <div className="flex gap-2">
        {['pending', 'approved', 'rejected'].map(s => (
          <button
            key={s}
            onClick={() => setFilter(s)}
            className={`px-4 py-2 rounded-xl text-sm font-medium capitalize transition-all ${filter === s ? 'bg-neutral-900 text-white' : 'bg-white border border-neutral-200 text-neutral-500 hover:bg-neutral-50'}`}
          >
            {s}
          </button>
        ))}
      </div>

      {/* Changes list */}
      {loading ? (
        <div className="p-16 text-center">
          <Loader2 size={24} className="animate-spin mx-auto text-neutral-300" />
          <p className="text-sm text-neutral-400 mt-3">Loading changes...</p>
        </div>
      ) : changes.length === 0 ? (
        <div className="bg-white rounded-2xl border border-neutral-50 p-16 text-center">
          <Shield size={48} className="mx-auto text-neutral-200 mb-4" />
          <p className="text-neutral-500">No {filter} changes</p>
        </div>
      ) : (
        <div className="space-y-4">
          {changes.map((change, i) => {
            const Icon = typeIcons[change.entity_type] || FileText;
            return (
              <div key={change.id} className="bg-white rounded-2xl border border-neutral-50 p-6 animate-fade-in" style={{ animationDelay: `${i * 0.03}s` }}>
                <div className="flex items-start justify-between">
                  <div className="flex items-start gap-4">
                    <div className="p-2.5 rounded-xl bg-neutral-50">
                      <Icon size={20} className="text-neutral-600" />
                    </div>
                    <div>
                      <div className="flex items-center gap-3">
                        <h3 className="font-medium text-neutral-900 capitalize">{change.entity_type} {change.change_type}</h3>
                        <span className={`px-2.5 py-1 rounded-full text-xs font-medium capitalize ${statusColors[change.status]}`}>
                          {change.status}
                        </span>
                      </div>
                      <p className="text-sm text-neutral-400 mt-1">
                        Entity ID: <code className="text-xs bg-neutral-50 px-1.5 py-0.5 rounded">{change.entity_id}</code>
                      </p>
                      <p className="text-xs text-neutral-400 mt-1">
                        Requested: {new Date(change.created_at).toLocaleString()}
                      </p>
                      {change.rejection_reason && (
                        <p className="text-sm text-red-600 mt-2">
                          Rejection reason: {change.rejection_reason}
                        </p>
                      )}
                    </div>
                  </div>
                  {change.status === 'pending' && (
                    <div className="flex gap-2">
                      <button
                        onClick={() => handleApprove(change.id)}
                        disabled={processing === change.id}
                        className="btn-hover flex items-center gap-2 px-4 py-2 bg-emerald-600 text-white rounded-xl text-sm font-medium hover:bg-emerald-700 disabled:opacity-50 transition-all"
                      >
                        {processing === change.id ? <Loader2 size={14} className="animate-spin" /> : <CheckCircle size={14} />}
                        Approve
                      </button>
                      <button
                        onClick={() => setShowRejectModal(change.id)}
                        disabled={processing === change.id}
                        className="btn-hover flex items-center gap-2 px-4 py-2 border border-neutral-200 text-neutral-600 rounded-xl text-sm font-medium hover:bg-red-50 hover:text-red-600 hover:border-red-200 disabled:opacity-50 transition-all"
                      >
                        <XCircle size={14} />
                        Reject
                      </button>
                    </div>
                  )}
                </div>

                {/* Data preview */}
                {change.after_data && (
                  <div className="mt-4 p-4 bg-neutral-50 rounded-xl">
                    <p className="text-xs font-medium text-neutral-500 uppercase tracking-wider mb-2">Proposed Changes</p>
                    <pre className="text-xs text-neutral-700 overflow-auto max-h-32">{JSON.stringify(change.after_data, null, 2)}</pre>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}

      {/* Reject modal */}
      {showRejectModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30 backdrop-blur-sm" onClick={() => setShowRejectModal(null)}>
          <div className="bg-white rounded-2xl p-6 w-full max-w-md mx-4 shadow-xl" onClick={e => e.stopPropagation()}>
            <h3 className="text-lg font-semibold text-neutral-900 mb-4">Reject Change</h3>
            <textarea
              value={reason}
              onChange={e => setReason(e.target.value)}
              placeholder="Enter rejection reason..."
              className="w-full px-4 py-3 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 resize-none h-24 placeholder:text-neutral-300"
            />
            <div className="flex gap-3 mt-4">
              <button onClick={() => setShowRejectModal(null)} className="flex-1 px-4 py-2.5 border border-neutral-200 rounded-xl text-sm font-medium hover:bg-neutral-50 transition-all">
                Cancel
              </button>
              <button
                onClick={() => handleReject(showRejectModal)}
                disabled={!reason.trim() || processing === showRejectModal}
                className="flex-1 px-4 py-2.5 bg-red-600 text-white rounded-xl text-sm font-medium hover:bg-red-700 disabled:opacity-50 transition-all"
              >
                {processing === showRejectModal ? 'Rejecting...' : 'Reject'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
