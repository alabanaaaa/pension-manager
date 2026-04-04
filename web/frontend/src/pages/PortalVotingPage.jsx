import { useState, useEffect } from 'react';
import { voting } from '../lib/api';
import { Loader2, Vote, CheckCircle, Clock, ChevronRight, AlertCircle } from 'lucide-react';

export default function PortalVotingPage() {
  const [elections, setElections] = useState([]);
  const [myVotes, setMyVotes] = useState({});
  const [loading, setLoading] = useState(true);
  const [voting, setVoting] = useState(null);
  const [selectedCandidate, setSelectedCandidate] = useState('');
  const [processing, setProcessing] = useState(false);
  const [result, setResult] = useState(null);

  useEffect(() => {
    voting.memberListElections()
      .then(r => setElections(Array.isArray(r.data) ? r.data : []))
      .catch(() => setElections([]))
      .finally(() => setLoading(false));
  }, []);

  const handleVote = async (electionId) => {
    if (!selectedCandidate) return;
    setProcessing(true);
    setResult(null);
    try {
      await voting.memberCastVote(electionId, { candidate_id: selectedCandidate });
      setResult({ success: true });
      setVoting(null);
      setSelectedCandidate('');
    } catch (err) {
      setResult({ success: false, error: err.response?.data?.error || 'Vote failed' });
    }
    finally { setProcessing(false); }
  };

  if (loading) {
    return (
      <div className="p-16 text-center">
        <Loader2 size={24} className="animate-spin mx-auto text-neutral-300" />
        <p className="text-sm text-neutral-400 mt-3">Loading elections...</p>
      </div>
    );
  }

  return (
    <div className="space-y-8 animate-fade-in-up">
      <div>
        <h1 className="text-3xl font-semibold tracking-tight text-neutral-900">Voting</h1>
        <p className="text-neutral-500 mt-2 text-base">Cast your vote in open elections</p>
      </div>

      {elections.length === 0 ? (
        <div className="bg-white rounded-2xl border border-neutral-50 p-16 text-center">
          <Vote size={48} className="mx-auto text-neutral-200 mb-4" />
          <p className="text-neutral-500">No open elections</p>
        </div>
      ) : (
        <div className="space-y-4">
          {elections.map((e, i) => (
            <div key={e.id} className="bg-white rounded-2xl border border-neutral-50 p-6 animate-fade-in" style={{ animationDelay: `${i * 0.05}s` }}>
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <div className="flex items-center gap-3">
                    <h3 className="font-semibold text-neutral-900">{e.title}</h3>
                    <span className="px-2.5 py-1 rounded-full text-xs font-medium bg-emerald-50 text-emerald-700">Open</span>
                  </div>
                  {e.description && <p className="text-sm text-neutral-400 mt-1">{e.description}</p>}
                  <p className="text-xs text-neutral-400 mt-2">Max {e.max_candidates || 3} vote(s) allowed</p>
                </div>
                <button
                  onClick={() => setVoting(e.id)}
                  className="btn-hover flex items-center gap-2 px-4 py-2.5 bg-neutral-900 text-white rounded-xl text-sm font-medium hover:bg-neutral-800 transition-all"
                >
                  Vote <ChevronRight size={14} />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Voting Modal */}
      {voting && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30 backdrop-blur-sm" onClick={() => setVoting(null)}>
          <div className="bg-white rounded-2xl p-6 w-full max-w-md mx-4 shadow-xl animate-scale-in" onClick={e => e.stopPropagation()}>
            <h3 className="text-lg font-semibold text-neutral-900 mb-4">Cast Your Vote</h3>
            <select
              value={selectedCandidate}
              onChange={e => setSelectedCandidate(e.target.value)}
              className="w-full px-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all mb-4"
            >
              <option value="">Select a candidate...</option>
              <option value="cand-1">Candidate 1</option>
              <option value="cand-2">Candidate 2</option>
              <option value="cand-3">Candidate 3</option>
            </select>
            <div className="flex gap-3">
              <button onClick={() => setVoting(null)} className="flex-1 px-4 py-2.5 border border-neutral-200 rounded-xl text-sm font-medium hover:bg-neutral-50 transition-all">
                Cancel
              </button>
              <button
                onClick={() => handleVote(voting)}
                disabled={!selectedCandidate || processing}
                className="flex-1 px-4 py-2.5 bg-neutral-900 text-white rounded-xl text-sm font-medium hover:bg-neutral-800 disabled:opacity-50 transition-all"
              >
                {processing ? 'Voting...' : 'Confirm Vote'}
              </button>
            </div>
            {result && (
              <div className={`mt-4 p-3 rounded-xl flex items-center gap-2 text-sm ${result.success ? 'bg-emerald-50 text-emerald-700' : 'bg-red-50 text-red-700'}`}>
                {result.success ? <CheckCircle size={16} /> : <AlertCircle size={16} />}
                {result.success ? 'Vote recorded successfully!' : result.error}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
