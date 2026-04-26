import { useState, useEffect } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { voting } from '../lib/api';
import { ArrowLeft, Loader2, Users, Calendar, Clock, CheckCircle, XCircle, Edit, Trash2, BarChart3 } from 'lucide-react';

export default function ManageElectionPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [election, setElection] = useState(null);
  const [candidates, setCandidates] = useState([]);

  useEffect(() => {
    if (!id) { setLoading(false); return; }
    Promise.all([
      voting.getElection(id),
      voting.listCandidates(id)
    ]).then(([eRes, cRes]) => {
      setElection(eRes.data);
      setCandidates(cRes.data || []);
    }).catch(() => {})
    .finally(() => setLoading(false));
  }, [id]);

  const handleStatusChange = async (status) => {
    try {
      await voting.updateStatus(id, status);
      setElection({ ...election, status });
    } catch (err) { console.error(err); }
  };

  const statusColors = {
    draft: 'bg-neutral-100 text-neutral-600',
    active: 'bg-emerald-50 text-emerald-700',
    closed: 'bg-red-50 text-red-700',
    completed: 'bg-blue-50 text-blue-700',
  };

  if (loading) return <div className="p-8 text-center"><Loader2 className="animate-spin mx-auto" /></div>;
  if (!election) return <div className="p-8 text-center">Election not found</div>;

  return (
    <div className="space-y-8">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Link to="/voting" className="p-2 hover:bg-neutral-100 rounded-lg"><ArrowLeft size={20} /></Link>
          <div>
            <h1 className="text-2xl font-bold tracking-tight text-black">Manage Election</h1>
            <p className="text-neutral-500 mt-1">{election.title}</p>
          </div>
        </div>
        <span className={`px-3 py-1.5 rounded-full text-sm font-medium ${statusColors[election.status]}`}>{election.status}</span>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2 space-y-6">
          <div className="bg-white rounded-2xl border border-neutral-100 p-6">
            <h2 className="text-lg font-semibold text-neutral-900 mb-4">Election Details</h2>
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div><p className="text-neutral-500">Election Type</p><p className="font-medium">{election.election_type}</p></div>
              <div><p className="text-neutral-500">Total Voters</p><p className="font-medium">{election.total_voters || 0}</p></div>
              <div><p className="text-neutral-500">Start Date</p><p className="font-medium">{election.start_date ? new Date(election.start_date).toLocaleDateString() : '-'}</p></div>
              <div><p className="text-neutral-500">End Date</p><p className="font-medium">{election.end_date ? new Date(election.end_date).toLocaleDateString() : '-'}</p></div>
              <div><p className="text-neutral-500">Votes Cast</p><p className="font-medium">{election.votes_cast || 0}</p></div>
              <div><p className="text-neutral-500">Created</p><p className="font-medium">{election.created_at ? new Date(election.created_at).toLocaleDateString() : '-'}</p></div>
            </div>
          </div>

          <div className="bg-white rounded-2xl border border-neutral-100 p-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold text-neutral-900">Candidates</h2>
              <button className="px-3 py-1.5 bg-neutral-900 text-white rounded-lg text-sm font-medium hover:bg-neutral-800">Add Candidate</button>
            </div>
            {candidates.length === 0 ? (
              <p className="text-neutral-400 text-sm">No candidates added yet</p>
            ) : (
              <div className="space-y-3">
                {candidates.map(c => (
                  <div key={c.id} className="flex items-center justify-between p-3 bg-neutral-50 rounded-lg">
                    <div className="flex items-center gap-3">
                      <div className="w-10 h-10 bg-neutral-200 rounded-full flex items-center justify-center">
                        <Users size={16} className="text-neutral-500" />
                      </div>
                      <div>
                        <p className="font-medium text-neutral-900">{c.name}</p>
                        <p className="text-xs text-neutral-500">{c.position}</p>
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      <span className="text-sm text-neutral-500">{c.votes || 0} votes</span>
                      <button className="p-1 hover:bg-neutral-200 rounded"><Edit size={14} className="text-neutral-400" /></button>
                      <button className="p-1 hover:bg-neutral-200 rounded"><Trash2 size={14} className="text-neutral-400" /></button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>

        <div className="space-y-6">
          <div className="bg-white rounded-2xl border border-neutral-100 p-6">
            <h2 className="text-lg font-semibold text-neutral-900 mb-4">Actions</h2>
            <div className="space-y-3">
              {election.status === 'draft' && (
                <button onClick={() => handleStatusChange('active')} className="w-full flex items-center justify-center gap-2 px-4 py-2.5 bg-emerald-600 text-white rounded-xl text-sm font-medium hover:bg-emerald-700">
                  <CheckCircle size={16} /> Start Election
                </button>
              )}
              {election.status === 'active' && (
                <button onClick={() => handleStatusChange('closed')} className="w-full flex items-center justify-center gap-2 px-4 py-2.5 bg-red-600 text-white rounded-xl text-sm font-medium hover:bg-red-700">
                  <XCircle size={16} /> Close Election
                </button>
              )}
              <Link to={`/voting/${id}/results`} className="w-full flex items-center justify-center gap-2 px-4 py-2.5 border border-neutral-200 text-neutral-700 rounded-xl text-sm font-medium hover:bg-neutral-50">
                <BarChart3 size={16} /> View Results
              </Link>
            </div>
          </div>

          <div className="bg-neutral-50 rounded-2xl border border-neutral-100 p-6">
            <h2 className="text-lg font-semibold text-neutral-900 mb-4">Quick Stats</h2>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <span className="text-sm text-neutral-500">Total Voters</span>
                <span className="font-medium">{election.total_voters || 0}</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-neutral-500">Votes Cast</span>
                <span className="font-medium">{election.votes_cast || 0}</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-neutral-500">Turnout</span>
                <span className="font-medium">
                  {election.total_voters > 0 ? Math.round((election.votes_cast || 0) / election.total_voters * 100) : 0}%
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
