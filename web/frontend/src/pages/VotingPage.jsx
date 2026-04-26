import { useState, useEffect } from 'react';
import { voting } from '../lib/api';
import { Link } from 'react-router-dom';
import { Plus, BarChart3, Users, CheckCircle, Clock, Loader2, ChevronRight } from 'lucide-react';

export default function VotingPage() {
  const [elections, setElections] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    voting.listElections()
      .then(res => setElections(res.data || []))
      .catch(() => setElections([]))
      .finally(() => setLoading(false));
  }, []);

  const statusMap = {
    draft: { label: 'Draft', cls: 'badge-warning' },
    open: { label: 'Open', cls: 'badge-success' },
    closed: { label: 'Closed', cls: 'badge-error' },
    archived: { label: 'Archived', cls: 'badge-info' },
  };

  return (
    <div className="space-y-8">
      <div className="flex flex-col sm:flex-row sm:items-end sm:justify-between gap-5">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-black">Elections</h1>
          <p className="text-sm text-gray-500 mt-1">{elections.length} elections</p>
        </div>
        <Link to="/voting/new" className="btn-primary flex items-center gap-2">
          <Plus size={15} /> New Election
        </Link>
      </div>

      {loading ? (
        <div className="p-16 text-center"><Loader2 size={24} className="animate-spin mx-auto text-gray-300" /><p className="text-sm text-gray-400 mt-3">Loading...</p></div>
      ) : elections.length === 0 ? (
        <div className="card p-16 text-center">
          <p className="text-sm text-gray-400 mb-4">No elections yet</p>
          <Link to="/voting/new" className="text-sm font-medium text-black hover:underline">Create one to get started</Link>
        </div>
      ) : (
        <div className="grid gap-5">
          {elections.map(e => {
            const st = statusMap[e.status] || { label: e.status, cls: 'badge-warning' };
            return (
              <div key={e.id} className="card p-5 hover:border-black transition-all">
                <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-5">
                  <div className="flex-1">
                    <div className="flex items-center gap-3">
                      <h3 className="text-lg font-semibold tracking-tight text-black">{e.title}</h3>
                      <span className={`badge ${e.status === 'open' ? 'badge-success' : e.status === 'closed' ? 'badge-error' : e.status === 'archived' ? 'badge-info' : 'badge-warning'}`}>{st.label}</span>
                    </div>
                    {e.description && <p className="text-sm text-gray-400 mt-1">{e.description}</p>}
                    <div className="flex gap-5 mt-3 text-sm text-gray-400">
                      <span className="flex items-center gap-1.5"><Users size={14} /> {e.total_voters} voters</span>
                      <span className="flex items-center gap-1.5"><CheckCircle size={14} /> {e.total_votes} votes</span>
                      <span className="flex items-center gap-1.5"><Clock size={14} /> Max {e.max_candidates} votes</span>
                    </div>
                  </div>
                  <div className="flex gap-2 flex-shrink-0">
                    <Link to={`/voting/${e.id}/results`} className="btn-secondary flex items-center gap-1.5">
                      <BarChart3 size={14} /> Results
                    </Link>
                    <Link to={`/voting/${e.id}`} className="btn-primary flex items-center gap-1.5">
                      Manage <ChevronRight size={14} />
                    </Link>
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
