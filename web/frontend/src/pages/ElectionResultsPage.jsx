import { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { voting } from '../lib/api';
import { ArrowLeft, Loader2, Users, BarChart3, Trophy } from 'lucide-react';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts';

const COLORS = ['#171717', '#404040', '#737373', '#a3a3a3'];

export default function ElectionResultsPage() {
  const { id } = useParams();
  const [loading, setLoading] = useState(true);
  const [election, setElection] = useState(null);
  const [results, setResults] = useState([]);
  const [stats, setStats] = useState(null);

  useEffect(() => {
    if (!id) { setLoading(false); return; }
    Promise.all([
      voting.getElection(id),
      voting.getResults(id),
      voting.getStats(id)
    ]).then(([eRes, rRes, sRes]) => {
      setElection(eRes.data);
      setResults(rRes.data || []);
      setStats(sRes.data);
    }).catch(() => {})
    .finally(() => setLoading(false));
  }, [id]);

  const statusColors = {
    draft: 'bg-neutral-100 text-neutral-600',
    active: 'bg-emerald-50 text-emerald-700',
    closed: 'bg-red-50 text-red-700',
    completed: 'bg-blue-50 text-blue-700',
  };

  if (loading) return <div className="p-8 text-center"><Loader2 className="animate-spin mx-auto" /></div>;
  if (!election) return <div className="p-8 text-center">Election not found</div>;

  const totalVotes = results.reduce((sum, r) => sum + (r.votes || 0), 0);

  return (
    <div className="space-y-8">
      <div className="flex items-center gap-4">
        <Link to="/voting" className="p-2 hover:bg-neutral-100 rounded-lg"><ArrowLeft size={20} /></Link>
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-black">Election Results</h1>
          <p className="text-neutral-500 mt-1">{election.title}</p>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div className="bg-white rounded-2xl border border-neutral-100 p-5">
          <p className="text-sm text-neutral-500">Total Voters</p>
          <p className="text-2xl font-semibold text-neutral-900">{election.total_voters || 0}</p>
        </div>
        <div className="bg-white rounded-2xl border border-neutral-100 p-5">
          <p className="text-sm text-neutral-500">Votes Cast</p>
          <p className="text-2xl font-semibold text-neutral-900">{election.votes_cast || 0}</p>
        </div>
        <div className="bg-white rounded-2xl border border-neutral-100 p-5">
          <p className="text-sm text-neutral-500">Turnout</p>
          <p className="text-2xl font-semibold text-neutral-900">
            {election.total_voters > 0 ? Math.round((election.votes_cast || 0) / election.total_voters * 100) : 0}%
          </p>
        </div>
        <div className="bg-white rounded-2xl border border-neutral-100 p-5">
          <p className="text-sm text-neutral-500">Status</p>
          <span className={`inline-block px-2.5 py-1 rounded-full text-sm font-medium ${statusColors[election.status]}`}>{election.status}</span>
        </div>
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-white rounded-2xl border border-neutral-100 p-6">
          <h2 className="text-lg font-semibold text-neutral-900 mb-4">Votes by Candidate</h2>
          {results.length > 0 ? (
            <ResponsiveContainer width="100%" height={300}>
              <BarChart data={results} layout="vertical">
                <CartesianGrid strokeDasharray="3 3" horizontal={false} stroke="#f5f5f4" />
                <XAxis type="number" fontSize={12} tick={{ fill: '#a8a29e' }} axisLine={false} />
                <YAxis type="category" dataKey="name" fontSize={12} tick={{ fill: '#a8a29e' }} axisLine={false} width={100} />
                <Tooltip />
                <Bar dataKey="votes" fill="#171717" radius={[0, 4, 4, 0]} />
              </BarChart>
            </ResponsiveContainer>
          ) : (
            <p className="text-neutral-400 text-center py-8">No votes recorded yet</p>
          )}
        </div>

        <div className="bg-white rounded-2xl border border-neutral-100 p-6">
          <h2 className="text-lg font-semibold text-neutral-900 mb-4">Vote Distribution</h2>
          {results.length > 0 ? (
            <ResponsiveContainer width="100%" height={300}>
              <PieChart>
                <Pie
                  data={results}
                  cx="50%"
                  cy="50%"
                  innerRadius={60}
                  outerRadius={100}
                  paddingAngle={2}
                  dataKey="votes"
                  nameKey="name"
                >
                  {results.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                  ))}
                </Pie>
                <Tooltip />
              </PieChart>
            </ResponsiveContainer>
          ) : (
            <p className="text-neutral-400 text-center py-8">No votes recorded yet</p>
          )}
          <div className="flex flex-wrap justify-center gap-4 mt-2">
            {results.map((r, i) => (
              <div key={r.id} className="flex items-center gap-2">
                <div className="w-3 h-3 rounded-full" style={{ backgroundColor: COLORS[i % COLORS.length] }} />
                <span className="text-xs text-neutral-600">{r.name}</span>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Results Table */}
      <div className="bg-white rounded-2xl border border-neutral-100 overflow-hidden">
        <div className="px-6 py-4 border-b border-neutral-50">
          <h2 className="text-lg font-semibold text-neutral-900">Candidate Results</h2>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead className="bg-neutral-50">
              <tr>
                <th className="text-left px-5 py-3 font-medium text-neutral-500">Rank</th>
                <th className="text-left px-5 py-3 font-medium text-neutral-500">Candidate</th>
                <th className="text-right px-5 py-3 font-medium text-neutral-500">Votes</th>
                <th className="text-right px-5 py-3 font-medium text-neutral-500">Percentage</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-neutral-50">
              {results.sort((a, b) => b.votes - a.votes).map((r, i) => (
                <tr key={r.id} className="hover:bg-neutral-50">
                  <td className="px-5 py-3">
                    {i === 0 ? <Trophy size={16} className="text-yellow-500" /> : <span className="text-neutral-500">#{i + 1}</span>}
                  </td>
                  <td className="px-5 py-3 font-medium text-neutral-900">{r.name}</td>
                  <td className="px-5 py-3 text-right">{r.votes || 0}</td>
                  <td className="px-5 py-3 text-right">
                    {totalVotes > 0 ? ((r.votes || 0) / totalVotes * 100).toFixed(1) : 0}%
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
