import { useState, useEffect } from 'react';
import { voting } from '../lib/api';
import { Loader2, Vote, CheckCircle, Clock, ChevronRight, AlertCircle, X } from 'lucide-react';

export default function PortalVotingPage() {
  const [elections, setElections] = useState([]);
  const [candidates, setCandidates] = useState([]);
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

  const loadCandidates = async (electionId) => {
    try {
      const res = await voting.memberListCandidates(electionId);
      setCandidates(Array.isArray(res.data) ? res.data : []);
    } catch {
      setCandidates([]);
    }
  };

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

  const startVoting = (electionId) => {
    setVoting(electionId);
    loadCandidates(electionId);
    setSelectedCandidate('');
    setResult(null);
  };

  if (loading) {
    return (
      <div className="p-16 text-center">
        <Loader2 size={24} className="animate-spin mx-auto text-gray-300" />
        <p className="text-sm text-gray-400 mt-3">Loading elections...</p>
      </div>
    );
  }

  return (
    <div className="space-y-6 animate-fade-in-up">
      <div>
        <h1 className="text-2xl font-bold tracking-tight text-black">Voting</h1>
        <p className="text-sm text-gray-500 mt-1">Cast your vote in open elections</p>
      </div>

      {elections.length === 0 ? (
        <div className="bg-white border border-gray-200 rounded-lg p-16 text-center">
          <Vote size={48} className="mx-auto text-gray-200 mb-4" />
          <p className="text-gray-500">No open elections</p>
        </div>
      ) : (
        <div className="space-y-4">
          {elections.map((e, i) => (
            <div key={e.id} className="bg-white border border-gray-200 rounded-lg p-5 animate-fade-in">
              <div className="flex items-start justify-between gap-4">
                <div className="flex-1">
                  <div className="flex items-center gap-3 mb-2">
                    <h3 className="font-semibold text-black">{e.title}</h3>
                    <span className="badge badge-success">Open</span>
                  </div>
                  {e.description && <p className="text-sm text-gray-500 mb-2">{e.description}</p>}
                  <p className="text-xs text-gray-400">Max {e.max_candidates || 3} vote(s) allowed</p>
                </div>
                <button
                  onClick={() => startVoting(e.id)}
                  className="btn btn-primary flex items-center gap-2"
                >
                  Vote <ChevronRight size={14} />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Voting Modal - Uber Style */}
      {voting && (
        <div className="modal-overlay" onClick={() => setVoting(null)}>
          <div className="modal animate-scale-in" onClick={e => e.stopPropagation()}>
            <div className="modal-header">
              <h3 className="modal-title">Cast Your Vote</h3>
              <button className="modal-close" onClick={() => setVoting(null)}>
                <X size={18} />
              </button>
            </div>
            <div className="modal-body">
              <select
                value={selectedCandidate}
                onChange={e => setSelectedCandidate(e.target.value)}
                className="input select"
              >
                <option value="">Select a candidate...</option>
                {candidates.map(c => (
                  <option key={c.id} value={c.id}>{c.name}</option>
                ))}
              </select>
            </div>
            <div className="modal-footer">
              <button onClick={() => setVoting(null)} className="btn btn-secondary">
                Cancel
              </button>
              <button
                onClick={() => handleVote(voting)}
                disabled={!selectedCandidate || processing}
                className="btn btn-primary"
              >
                {processing ? 'Voting...' : 'Confirm Vote'}
              </button>
            </div>
            {result && (
              <div className={`p-4 mx-5 mb-5 rounded-lg flex items-center gap-2 text-sm ${
                result.success ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-700'
              }`}>
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
