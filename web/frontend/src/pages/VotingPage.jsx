import { useState, useEffect } from 'react';
import { voting } from '../lib/api';
import { Link } from 'react-router-dom';
import { Plus, BarChart3, Users, CheckCircle, Clock, Loader2 } from 'lucide-react';

export default function VotingPage() {
  const [elections, setElections] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    voting.listElections()
      .then(res => setElections(res.data || []))
      .catch(() => setElections([]))
      .finally(() => setLoading(false));
  }, []);

  const statusColors = {
    draft: 'bg-gray-100 text-gray-800',
    open: 'bg-green-100 text-green-800',
    closed: 'bg-red-100 text-red-800',
    archived: 'bg-blue-100 text-blue-800',
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Elections</h1>
          <p className="text-gray-500 mt-1">{elections.length} elections</p>
        </div>
        <Link
          to="/voting/new"
          className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700"
        >
          <Plus size={16} />
          New Election
        </Link>
      </div>

      {loading ? (
        <div className="p-12 text-center"><Loader2 size={32} className="animate-spin mx-auto text-gray-400" /></div>
      ) : elections.length === 0 ? (
        <div className="bg-white rounded-xl border p-12 text-center text-gray-500">
          No elections yet. Create one to get started.
        </div>
      ) : (
        <div className="grid gap-4">
          {elections.map(e => (
            <div key={e.id} className="bg-white rounded-xl border p-5">
              <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4">
                <div className="flex-1">
                  <div className="flex items-center gap-3">
                    <h3 className="text-lg font-semibold">{e.title}</h3>
                    <span className={`px-2 py-1 rounded-full text-xs font-medium ${statusColors[e.status]}`}>
                      {e.status}
                    </span>
                  </div>
                  <p className="text-gray-500 text-sm mt-1">{e.description}</p>
                  <div className="flex gap-6 mt-3 text-sm text-gray-500">
                    <span className="flex items-center gap-1.5"><Users size={14} /> {e.total_voters} voters</span>
                    <span className="flex items-center gap-1.5"><CheckCircle size={14} /> {e.total_votes} votes</span>
                    <span className="flex items-center gap-1.5"><Clock size={14} /> {e.max_candidates} max votes</span>
                  </div>
                </div>
                <div className="flex gap-2">
                  <Link to={`/voting/${e.id}/results`} className="flex items-center gap-1.5 px-3 py-2 border rounded-lg text-sm hover:bg-gray-50">
                    <BarChart3 size={14} /> Results
                  </Link>
                  <Link to={`/voting/${e.id}`} className="flex items-center gap-1.5 px-3 py-2 bg-blue-600 text-white rounded-lg text-sm hover:bg-blue-700">
                    Manage
                  </Link>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
