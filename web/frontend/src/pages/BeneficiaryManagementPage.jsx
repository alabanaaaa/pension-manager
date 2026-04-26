import { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { members } from '../lib/api';
import { Loader2, Plus, Trash2, Edit, Users, AlertCircle } from 'lucide-react';

export default function BeneficiaryManagementPage() {
  const { id } = useParams();
  const [beneficiaries, setBeneficiaries] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [editing, setEditing] = useState(null);
  const [form, setForm] = useState({
    name: '',
    relationship: '',
    date_of_birth: '',
    id_number: '',
    phone: '',
    allocation_pct: 0,
  });

  useEffect(() => {
    if (id) {
      loadBeneficiaries();
    }
  }, [id]);

  const loadBeneficiaries = async () => {
    setLoading(true);
    try {
      const res = await members.getBeneficiaries(id);
      setBeneficiaries(Array.isArray(res.data) ? res.data : []);
    } catch { setBeneficiaries([]); }
    finally { setLoading(false); }
  };

  const handleChange = (e) => {
    const { name, value, type } = e.target;
    setForm(prev => ({
      ...prev,
      [name]: type === 'number' ? parseInt(value) || 0 : value
    }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      if (editing) {
        // Update would go here
      } else {
        await members.addBeneficiary(id, form);
      }
      setShowForm(false);
      setEditing(null);
      setForm({ name: '', relationship: '', date_of_birth: '', id_number: '', phone: '', allocation_pct: 0 });
      loadBeneficiaries();
    } catch (err) {
      alert(err.response?.data?.error || 'Failed to save beneficiary');
    }
  };

  const totalAllocation = beneficiaries.reduce((sum, b) => sum + (b.allocation_pct || 0), 0);

  const relationshipOptions = [
    { value: 'spouse', label: 'Spouse' },
    { value: 'child', label: 'Child' },
    { value: 'parent', label: 'Parent' },
    { value: 'sibling', label: 'Sibling' },
    { value: 'other', label: 'Other' },
  ];

  if (loading) {
    return (
      <div className="p-16 text-center">
        <Loader2 size={24} className="animate-spin mx-auto text-neutral-300" />
        <p className="text-sm text-neutral-400 mt-3">Loading...</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Users size={24} className="text-neutral-500" />
          <div>
            <h2 className="text-xl font-semibold text-neutral-900">Beneficiaries</h2>
            <p className="text-sm text-neutral-500">Manage dependent details</p>
          </div>
        </div>
        <button
          onClick={() => setShowForm(true)}
          className="flex items-center gap-2 px-4 py-2.5 bg-neutral-900 text-white rounded-xl text-sm font-medium"
        >
          <Plus size={16} /> Add Beneficiary
        </button>
      </div>

      {/* Allocation Warning */}
      {totalAllocation > 0 && totalAllocation !== 100 && (
        <div className="flex items-center gap-3 p-4 bg-amber-50 border border-amber-100 rounded-xl">
          <AlertCircle size={20} className="text-amber-600" />
          <div>
            <p className="text-sm font-medium text-amber-800">Allocation Warning</p>
            <p className="text-sm text-amber-600">Total allocation is {totalAllocation}%. Should equal 100%.</p>
          </div>
        </div>
      )}

      {/* Beneficiary List */}
      <div className="bg-white rounded-2xl border border-neutral-100 overflow-hidden">
        {beneficiaries.length === 0 ? (
          <div className="p-12 text-center">
            <Users size={32} className="mx-auto text-neutral-300 mb-3" />
            <p className="text-neutral-400">No beneficiaries added yet</p>
          </div>
        ) : (
          <table className="w-full text-sm">
            <thead className="bg-neutral-50 border-b border-neutral-100">
              <tr>
                <th className="text-left px-5 py-3 font-medium text-neutral-500">Name</th>
                <th className="text-left px-5 py-3 font-medium text-neutral-500">Relationship</th>
                <th className="text-left px-5 py-3 font-medium text-neutral-500">Date of Birth</th>
                <th className="text-left px-5 py-3 font-medium text-neutral-500">ID Number</th>
                <th className="text-right px-5 py-3 font-medium text-neutral-500">Allocation %</th>
                <th className="text-right px-5 py-3 font-medium text-neutral-500">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-neutral-50">
              {beneficiaries.map(b => (
                <tr key={b.id} className="hover:bg-neutral-50">
                  <td className="px-5 py-4 font-medium text-neutral-900">{b.name}</td>
                  <td className="px-5 py-4 text-neutral-600 capitalize">{b.relationship}</td>
                  <td className="px-5 py-4 text-neutral-600">
                    {b.date_of_birth ? new Date(b.date_of_birth).toLocaleDateString() : '-'}
                  </td>
                  <td className="px-5 py-4 text-neutral-600">{b.id_number || '-'}</td>
                  <td className="px-5 py-4 text-right font-medium text-neutral-900">{b.allocation_pct}%</td>
                  <td className="px-5 py-4 text-right">
                    <button className="p-2 hover:bg-neutral-100 rounded-lg">
                      <Edit size={16} className="text-neutral-500" />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {/* Add/Edit Modal */}
      {showForm && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-2xl w-full max-w-md p-6">
            <h3 className="text-lg font-semibold mb-4">{editing ? 'Edit' : 'Add'} Beneficiary</h3>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-neutral-700 mb-1">Full Name *</label>
                <input
                  type="text"
                  name="name"
                  value={form.name}
                  onChange={handleChange}
                  required
                  className="w-full px-3 py-2 border border-neutral-200 rounded-lg text-sm"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-neutral-700 mb-1">Relationship *</label>
                <select
                  name="relationship"
                  value={form.relationship}
                  onChange={handleChange}
                  required
                  className="w-full px-3 py-2 border border-neutral-200 rounded-lg text-sm"
                >
                  <option value="">Select</option>
                  {relationshipOptions.map(o => (
                    <option key={o.value} value={o.value}>{o.label}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-neutral-700 mb-1">Date of Birth</label>
                <input
                  type="date"
                  name="date_of_birth"
                  value={form.date_of_birth}
                  onChange={handleChange}
                  className="w-full px-3 py-2 border border-neutral-200 rounded-lg text-sm"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-neutral-700 mb-1">ID Number</label>
                <input
                  type="text"
                  name="id_number"
                  value={form.id_number}
                  onChange={handleChange}
                  className="w-full px-3 py-2 border border-neutral-200 rounded-lg text-sm"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-neutral-700 mb-1">Phone</label>
                <input
                  type="text"
                  name="phone"
                  value={form.phone}
                  onChange={handleChange}
                  className="w-full px-3 py-2 border border-neutral-200 rounded-lg text-sm"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-neutral-700 mb-1">Allocation % *</label>
                <input
                  type="number"
                  name="allocation_pct"
                  value={form.allocation_pct}
                  onChange={handleChange}
                  required
                  min={0}
                  max={100}
                  className="w-full px-3 py-2 border border-neutral-200 rounded-lg text-sm"
                />
              </div>
              <div className="flex gap-3 pt-4">
                <button
                  type="button"
                  onClick={() => { setShowForm(false); setEditing(null); }}
                  className="flex-1 px-4 py-2 border border-neutral-200 rounded-lg text-sm font-medium hover:bg-neutral-50"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="flex-1 px-4 py-2 bg-neutral-900 text-white rounded-lg text-sm font-medium"
                >
                  Save
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
