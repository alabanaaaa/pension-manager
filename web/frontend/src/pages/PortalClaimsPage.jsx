import { useState, useEffect } from 'react';
import { claims } from '../lib/api';
import { Loader2, FileText, CheckCircle, XCircle, Clock, Calendar, DollarSign } from 'lucide-react';

export default function PortalClaimsPage() {
  const [claimsList, setClaimsList] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    claims.list()
      .then(r => setClaimsList(Array.isArray(r.data) ? r.data : []))
      .catch(() => setClaimsList([]))
      .finally(() => setLoading(false));
  }, []);

  const statusConfig = {
    submitted: { label: 'Submitted', cls: 'bg-amber-50 text-amber-700', icon: Clock },
    accepted: { label: 'Accepted', cls: 'bg-emerald-50 text-emerald-700', icon: CheckCircle },
    rejected: { label: 'Rejected', cls: 'bg-red-50 text-red-700', icon: XCircle },
    paid: { label: 'Paid', cls: 'bg-blue-50 text-blue-700', icon: CheckCircle },
  };

  if (loading) {
    return (
      <div className="p-16 text-center">
        <Loader2 size={24} className="animate-spin mx-auto text-neutral-300" />
        <p className="text-sm text-neutral-400 mt-3">Loading claims...</p>
      </div>
    );
  }

  return (
    <div className="space-y-8 animate-fade-in-up">
      <div>
        <h1 className="text-2xl font-bold tracking-tight text-black">My Claims</h1>
        <p className="text-neutral-500 mt-2 text-base">Check the status of your claims</p>
      </div>

      {claimsList.length === 0 ? (
        <div className="bg-white rounded-2xl border border-neutral-50 p-16 text-center">
          <FileText size={48} className="mx-auto text-neutral-200 mb-4" />
          <p className="text-neutral-500">No claims found</p>
        </div>
      ) : (
        <div className="space-y-4">
          {claimsList.map((claim, i) => {
            const cfg = statusConfig[claim.status] || { label: claim.status, cls: 'bg-neutral-50 text-neutral-600', icon: Clock };
            return (
              <div key={claim.id} className="bg-white rounded-2xl border border-neutral-50 p-6 hover:shadow-sm transition-all animate-fade-in" style={{ animationDelay: `${i * 0.03}s` }}>
                <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4">
                  <div className="flex-1">
                    <div className="flex items-center gap-3">
                      <h3 className="font-semibold text-neutral-900">{claim.claim_form_no || claim.id}</h3>
                      <span className={`px-2.5 py-1 rounded-full text-xs font-medium ${cfg.cls}`}>{cfg.label}</span>
                    </div>
                    <p className="text-sm text-neutral-400 mt-1 capitalize">{claim.claim_type} claim</p>
                    <div className="flex flex-wrap gap-4 mt-3 text-sm text-neutral-400">
                      <span className="flex items-center gap-1.5"><Calendar size={13} /> {new Date(claim.date_of_claim).toLocaleDateString()}</span>
                      {claim.amount && (
                        <span className="flex items-center gap-1.5"><DollarSign size={13} /> KES {(claim.amount / 100).toLocaleString()}</span>
                      )}
                    </div>
                    {claim.rejection_reason && (
                      <p className="text-sm text-red-600 mt-2">Reason: {claim.rejection_reason}</p>
                    )}
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
