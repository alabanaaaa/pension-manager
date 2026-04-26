import { useState, useEffect } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { voting } from '../lib/api';
import { ArrowLeft, Save, Loader2, Users, Calendar, Globe, Phone } from 'lucide-react';

export default function NewElectionPage() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [form, setForm] = useState({
    title: '',
    description: '',
    election_type: 'trustee',
    max_candidates: 3,
    allow_ussd: false,
    allow_web: true,
    start_date: '',
    end_date: '',
  });

  const handleChange = (e) => {
    const { name, value, type, checked } = e.target;
    setForm(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : (type === 'number' ? parseInt(value) || 0 : value)
    }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    
    try {
      const res = await voting.createElection(form);
      navigate(`/voting/${res.data.id}`);
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to create election');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Link to="/voting" className="btn-hover p-2 rounded-xl hover:bg-neutral-100">
          <ArrowLeft size={20} className="text-neutral-600" />
        </Link>
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-black">New Election</h1>
          <p className="text-neutral-500 mt-1">Create a new voting election</p>
        </div>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-100 rounded-2xl p-4 text-red-600 text-sm">
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-8">
        {/* Election Details */}
        <div className="bg-white rounded-2xl border border-neutral-100 overflow-hidden">
          <div className="px-6 py-4 border-b border-neutral-50 bg-neutral-50/50">
            <div className="flex items-center gap-2">
              <Users size={18} className="text-neutral-500" />
              <h2 className="text-lg font-medium text-neutral-900">Election Details</h2>
            </div>
          </div>
          <div className="p-6 grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <label className="block text-sm font-medium text-neutral-700 mb-1.5">Title *</label>
              <input
                type="text"
                name="title"
                value={form.title}
                onChange={handleChange}
                required
                className="w-full px-4 py-2.5 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10"
                placeholder="e.g., Trustee Election 2026"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-neutral-700 mb-1.5">Election Type *</label>
              <select
                name="election_type"
                value={form.election_type}
                onChange={handleChange}
                required
                className="w-full px-4 py-2.5 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10"
              >
                <option value="trustee">Trustee Election</option>
                <option value="agenda">Agenda Voting</option>
                <option value="agm">AGM Voting</option>
              </select>
            </div>
            <div className="md:col-span-2">
              <label className="block text-sm font-medium text-neutral-700 mb-1.5">Description</label>
              <textarea
                name="description"
                value={form.description}
                onChange={handleChange}
                rows={3}
                className="w-full px-4 py-2.5 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10"
                placeholder="Describe the election purpose..."
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-neutral-700 mb-1.5">Max Candidates per Voter</label>
              <input
                type="number"
                name="max_candidates"
                value={form.max_candidates}
                onChange={handleChange}
                min={1}
                max={10}
                className="w-full px-4 py-2.5 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10"
              />
            </div>
          </div>
        </div>

        {/* Voting Options */}
        <div className="bg-white rounded-2xl border border-neutral-100 overflow-hidden">
          <div className="px-6 py-4 border-b border-neutral-50 bg-neutral-50/50">
            <div className="flex items-center gap-2">
              <Globe size={18} className="text-neutral-500" />
              <h2 className="text-lg font-medium text-neutral-900">Voting Options</h2>
            </div>
          </div>
          <div className="p-6 space-y-4">
            <label className="flex items-center gap-3">
              <input
                type="checkbox"
                name="allow_web"
                checked={form.allow_web}
                onChange={handleChange}
                className="w-4 h-4 rounded border-neutral-300 text-neutral-900 focus:ring-neutral-900"
              />
              <div>
                <p className="font-medium text-neutral-900">Allow Web Voting</p>
                <p className="text-sm text-neutral-500">Members can vote via the web portal</p>
              </div>
            </label>
            <label className="flex items-center gap-3">
              <input
                type="checkbox"
                name="allow_ussd"
                checked={form.allow_ussd}
                onChange={handleChange}
                className="w-4 h-4 rounded border-neutral-300 text-neutral-900 focus:ring-neutral-900"
              />
              <div>
                <p className="font-medium text-neutral-900">Allow USSD Voting</p>
                <p className="text-sm text-neutral-500">Members can vote via USSD short code</p>
              </div>
            </label>
          </div>
        </div>

        {/* Schedule */}
        <div className="bg-white rounded-2xl border border-neutral-100 overflow-hidden">
          <div className="px-6 py-4 border-b border-neutral-50 bg-neutral-50/50">
            <div className="flex items-center gap-2">
              <Calendar size={18} className="text-neutral-500" />
              <h2 className="text-lg font-medium text-neutral-900">Schedule</h2>
            </div>
          </div>
          <div className="p-6 grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <label className="block text-sm font-medium text-neutral-700 mb-1.5">Start Date & Time *</label>
              <input
                type="datetime-local"
                name="start_date"
                value={form.start_date}
                onChange={handleChange}
                required
                className="w-full px-4 py-2.5 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-neutral-700 mb-1.5">End Date & Time *</label>
              <input
                type="datetime-local"
                name="end_date"
                value={form.end_date}
                onChange={handleChange}
                required
                className="w-full px-4 py-2.5 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10"
              />
            </div>
          </div>
        </div>

        {/* Submit */}
        <div className="flex justify-end gap-3">
          <Link to="/voting" className="px-6 py-2.5 border border-neutral-200 rounded-xl text-sm font-medium hover:bg-neutral-50">
            Cancel
          </Link>
          <button
            type="submit"
            disabled={loading}
            className="flex items-center gap-2 px-6 py-2.5 bg-neutral-900 text-white rounded-xl text-sm font-medium hover:bg-neutral-800 disabled:opacity-50"
          >
            {loading ? <Loader2 size={16} className="animate-spin" /> : <Save size={16} />}
            Create Election
          </button>
        </div>
      </form>
    </div>
  );
}
