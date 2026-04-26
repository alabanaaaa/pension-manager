import { useState, useEffect } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { claims } from '../lib/api';
import { ArrowLeft, Loader2, FileText, User, Calendar, DollarSign, Check, X, Clock } from 'lucide-react';

export default function ClaimDetailsPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [claim, setClaim] = useState(null);
  const [actionLoading, setActionLoading] = useState(false);

  useEffect(() => {
    if (!id) { setLoading(false); return; }
    claims.get(id)
      .then(res => setClaim(res.data))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [id]);

  const handleApprove = async () => {
    setActionLoading(true);
    try {
      await claims.approve(id, 'Approved via dashboard');
      navigate('/claims');
    } catch (err) { console.error(err); }
    finally { setActionLoading(false); }
  };

  const handleReject = async () => {
    setActionLoading(true);
    try {
      await claims.reject(id, 'Rejected via dashboard');
      navigate('/claims');
    } catch (err) { console.error(err); }
    finally { setActionLoading(false); }
  };

  const handlePay = async () => {
    setActionLoading(true);
    try {
      await claims.pay(id, { payment_method: 'bank_transfer', reference: `PAY-${Date.now()}` });
      navigate('/claims');
    } catch (err) { console.error(err); }
    finally { setActionLoading(false); }
  };

  const statusColors = {
    submitted: 'bg-amber-50 text-amber-700',
    under_review: 'bg-blue-50 text-blue-700',
    accepted: 'bg-emerald-50 text-emerald-700',
    rejected: 'bg-red-50 text-red-700',
    paid: 'bg-green-50 text-green-700',
  };

  if (loading) return <div className="p-8 text-center"><Loader2 className="animate-spin mx-auto" /></div>;
  if (!claim) return <div className="p-8 text-center">Claim not found</div>;

  return (
    <div className="space-y-8">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Link to="/claims" className="p-2 hover:bg-neutral-100 rounded-lg"><ArrowLeft size={20} /></Link>
          <div>
            <h1 className="text-2xl font-bold tracking-tight text-black">Claim Details</h1>
            <p className="text-neutral-500 mt-1">Claim ID: {claim.id?.slice(0, 8)}</p>
          </div>
        </div>
        <span className={`px-3 py-1.5 rounded-full text-sm font-medium ${statusColors[claim.status] || 'bg-neutral-50 text-neutral-600'}`}>
          {claim.status}
        </span>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2 space-y-6">
          <div className="bg-white rounded-2xl border border-neutral-100 p-6">
            <h2 className="text-lg font-semibold text-neutral-900 mb-4">Claim Information</h2>
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div><p className="text-neutral-500">Claim Type</p><p className="font-medium">{claim.claim_type}</p></div>
              <div><p className="text-neutral-500">Amount</p><p className="font-medium">KES {(claim.amount / 100).toLocaleString()}</p></div>
              <div><p className="text-neutral-500">Date of Claim</p><p className="font-medium">{new Date(claim.date_of_claim).toLocaleDateString()}</p></div>
              <div><p className="text-neutral-500">Created</p><p className="font-medium">{new Date(claim.created_at).toLocaleDateString()}</p></div>
            </div>
          </div>

          <div className="bg-white rounded-2xl border border-neutral-100 p-6">
            <h2 className="text-lg font-semibold text-neutral-900 mb-4">Member Information</h2>
            <div className="flex items-center gap-4">
              <div className="w-12 h-12 bg-neutral-100 rounded-full flex items-center justify-center">
                <User size={20} className="text-neutral-500" />
              </div>
              <div>
                <p className="font-medium text-neutral-900">{claim.member_name || 'Unknown'}</p>
                <p className="text-sm text-neutral-500">Member No: {claim.member_no}</p>
              </div>
            </div>
          </div>
        </div>

        <div className="space-y-6">
          <div className="bg-white rounded-2xl border border-neutral-100 p-6">
            <h2 className="text-lg font-semibold text-neutral-900 mb-4">Actions</h2>
            <div className="space-y-3">
              {claim.status === 'submitted' && (
                <>
                  <button onClick={handleApprove} disabled={actionLoading} className="w-full flex items-center justify-center gap-2 px-4 py-2.5 bg-emerald-600 text-white rounded-xl text-sm font-medium hover:bg-emerald-700">
                    <Check size={16} /> Approve Claim
                  </button>
                  <button onClick={handleReject} disabled={actionLoading} className="w-full flex items-center justify-center gap-2 px-4 py-2.5 border border-red-200 text-red-600 rounded-xl text-sm font-medium hover:bg-red-50">
                    <X size={16} /> Reject Claim
                  </button>
                </>
              )}
              {claim.status === 'accepted' && (
                <button onClick={handlePay} disabled={actionLoading} className="w-full flex items-center justify-center gap-2 px-4 py-2.5 bg-neutral-900 text-white rounded-xl text-sm font-medium hover:bg-neutral-800">
                  <DollarSign size={16} /> Process Payment
                </button>
              )}
            </div>
          </div>

          <div className="bg-neutral-50 rounded-2xl border border-neutral-100 p-6">
            <h2 className="text-lg font-semibold text-neutral-900 mb-4">Timeline</h2>
            <div className="space-y-4">
              <div className="flex items-center gap-3">
                <div className="w-2 h-2 bg-neutral-400 rounded-full" />
                <div className="text-sm"><p className="font-medium">Created</p><p className="text-neutral-500">{new Date(claim.created_at).toLocaleString()}</p></div>
              </div>
              {claim.reviewed_at && claim.reviewed_at !== '0001-01-01T00:00:00Z' && (
                <div className="flex items-center gap-3">
                  <div className="w-2 h-2 bg-blue-400 rounded-full" />
                  <div className="text-sm"><p className="font-medium">Reviewed</p><p className="text-neutral-500">{new Date(claim.reviewed_at).toLocaleString()}</p></div>
                </div>
              )}
              {claim.paid_at && claim.paid_at !== '0001-01-01T00:00:00Z' && (
                <div className="flex items-center gap-3">
                  <div className="w-2 h-2 bg-green-400 rounded-full" />
                  <div className="text-sm"><p className="font-medium">Paid</p><p className="text-neutral-500">{new Date(claim.paid_at).toLocaleString()}</p></div>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
