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
    draft: { label: 'Draft', cls: 'bg-neutral-100 text-neutral-600' },
    open: { label: 'Open', cls: 'bg-emerald-50 text-emerald-700' },
    closed: { label: 'Closed', cls: 'bg-red-50 text-red-700' },
    archived: { label: 'Archived', cls: 'bg-blue-50 text-blue-700' },
  };

  return (
    <div className="space-y-8">
      <div className="flex flex-col sm:flex-row sm:items-end sm:justify-between gap-5">
        <div>
          <h1 className="text-3xl font-semibold tracking-tight text-neutral-900">Elections</h1>
          <p className="text-neutral-500 mt-1">{elections.length} elections</p>
        </div>
        <Link to="/voting/new" className="btn-hover flex items-center gap-2 px-4 py-2.5 bg-neutral-900 text-white rounded-xl text-sm font-medium hover:bg-neutral-800 transition-all">
          <Plus size={15} /> New Election
        </Link>
      </div>

      {loading ? (
        <div className="p-16 text-center"><Loader2 size={24} className="animate-spin mx-auto text-neutral-300" /><p className="text-sm text-neutral-400 mt-3">Loading...</p></div>
      ) : elections.length === 0 ? (
        <div className="bg-white rounded-xl border border-[#e8e9eb] p-16 text-center">
          <p className="text-sm text-neutral-400 mb-4">No elections yet</p>
          <Link to="/voting/new" className="text-sm text-neutral-900 font-medium hover:underline">Create one to get started</Link>
        </div>
      ) : (
        <div className="grid gap-5">
          {elections.map(e => {
            const st = statusMap[e.status] || { label: e.status, cls: 'bg-neutral-50 text-neutral-600' };
            return (
              <div key={e.id} className="bg-white rounded-xl border border-[#e8e9eb] p-5 hover:shadow-sm transition-all">
                <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-5">
                  <div className="flex-1">
                    <div className="flex items-center gap-3">
                      <h3 className="text-lg font-semibold tracking-tight text-neutral-900">{e.title}</h3>
                      <span className={`px-2.5 py-1 rounded-full text-xs font-medium ${st.cls}`}>{st.label}</span>
                    </div>
                    {e.description && <p className="text-sm text-neutral-400 mt-1">{e.description}</p>}
                    <div className="flex gap-5 mt-3 text-sm text-neutral-400">
                      <span className="flex items-center gap-1.5"><Users size={14} /> {e.total_voters} voters</span>
                      <span className="flex items-center gap-1.5"><CheckCircle size={14} /> {e.total_votes} votes</span>
                      <span className="flex items-center gap-1.5"><Clock size={14} /> Max {e.max_candidates} votes</span>
                    </div>
                  </div>
                  <div className="flex gap-2 flex-shrink-0">
                    <Link to={`/voting/${e.id}/results`} className="btn-hover flex items-center gap-1.5 px-4 py-2.5 border border-neutral-200 rounded-xl text-sm hover:bg-neutral-50 transition-all">
                      <BarChart3 size={14} /> Results
                    </Link>
                    <Link to={`/voting/${e.id}`} className="btn-hover flex items-center gap-1.5 px-4 py-2.5 bg-neutral-900 text-white rounded-xl text-sm hover:bg-neutral-800 transition-all">
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
