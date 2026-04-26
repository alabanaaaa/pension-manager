import { useState, useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { members, contributions, claims } from '../lib/api';
import { ArrowLeft, Edit, Loader2, User, Calendar, Landmark, CreditCard, Users, FileText } from 'lucide-react';

export default function MemberDetailsPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [member, setMember] = useState(null);
  const [beneficiaries, setBeneficiaries] = useState([]);
  const [memberContributions, setMemberContributions] = useState([]);
  const [memberClaims, setMemberClaims] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState('details');

  useEffect(() => {
    setLoading(true);
    members.get(id)
      .then(res => setMember(res.data))
      .catch(() => setMember(null))
      .finally(() => setLoading(false));
  }, [id]);

  useEffect(() => {
    if (activeTab === 'beneficiaries' && id) {
      members.getBeneficiaries(id)
        .then(res => setBeneficiaries(Array.isArray(res.data) ? res.data : []))
        .catch(() => setBeneficiaries([]));
    }
    if (activeTab === 'contributions' && id) {
      contributions.list({ member_id: id, limit: 100 })
        .then(res => setMemberContributions(Array.isArray(res.data) ? res.data : []))
        .catch(() => setMemberContributions([]));
    }
    if (activeTab === 'claims' && id) {
      claims.list({ member_id: id, limit: 100 })
        .then(res => setMemberClaims(Array.isArray(res.data) ? res.data : []))
        .catch(() => setMemberClaims([]));
    }
  }, [activeTab, id]);

  const formatCurrency = (value) => {
    return `KES ${(value / 100).toLocaleString()}`;
  };

  const statusColors = {
    submitted: 'bg-amber-50 text-amber-700',
    under_review: 'bg-blue-50 text-blue-700',
    accepted: 'bg-emerald-50 text-emerald-700',
    rejected: 'bg-red-50 text-red-700',
    paid: 'bg-green-50 text-green-700',
  };

  if (loading) {
    return (
      <div className="p-16 text-center">
        <Loader2 size={24} className="animate-spin mx-auto text-neutral-300" />
        <p className="text-sm text-neutral-400 mt-3">Loading member...</p>
      </div>
    );
  }

  if (!member) {
    return (
      <div className="p-16 text-center">
        <p className="text-neutral-400">Member not found</p>
        <Link to="/members" className="text-sm text-neutral-600 mt-2 inline-block">Back to Members</Link>
      </div>
    );
  }

  const statusMap = {
    active: { label: 'Active', cls: 'bg-emerald-50 text-emerald-700' },
    retired: { label: 'Retired', cls: 'bg-blue-50 text-blue-700' },
    deceased: { label: 'Deceased', cls: 'bg-neutral-100 text-neutral-600' },
    deferred: { label: 'Deferred', cls: 'bg-amber-50 text-amber-700' },
    withdrawn: { label: 'Withdrawn', cls: 'bg-red-50 text-red-700' },
  };

  const status = statusMap[member.membership_status] || { label: member.membership_status, cls: 'bg-neutral-50 text-neutral-600' };

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Link to="/members" className="btn-hover p-2 rounded-xl hover:bg-neutral-100">
            <ArrowLeft size={20} className="text-neutral-600" />
          </Link>
          <div>
            <div className="flex items-center gap-3">
              <h1 className="text-2xl font-bold tracking-tight text-black">
                {member.first_name} {member.last_name}
              </h1>
              <span className={`px-2.5 py-1 rounded-full text-xs font-medium ${status.cls}`}>
                {status.label}
              </span>
            </div>
            <p className="text-neutral-500 mt-1">{member.member_no}</p>
          </div>
        </div>
        <div className="flex gap-2">
          <Link to={`/members/${id}/edit`} className="flex items-center gap-2 px-4 py-2.5 border border-neutral-200 rounded-xl text-sm font-medium hover:bg-neutral-50">
            <Edit size={15} /> Edit
          </Link>
        </div>
      </div>

      {/* Quick Stats */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="bg-white rounded-2xl border border-neutral-100 p-5">
          <p className="text-sm text-neutral-500">Account Balance</p>
          <p className="text-2xl font-semibold text-neutral-900 mt-1">
            KES {Number(member.account_balance || 0).toLocaleString()}
          </p>
        </div>
        <div className="bg-white rounded-2xl border border-neutral-100 p-5">
          <p className="text-sm text-neutral-500">Basic Salary</p>
          <p className="text-2xl font-semibold text-neutral-900 mt-1">
            KES {Number(member.basic_salary || 0).toLocaleString()}
          </p>
        </div>
        <div className="bg-white rounded-2xl border border-neutral-100 p-5">
          <p className="text-sm text-neutral-500">Date Joined</p>
          <p className="text-lg font-semibold text-neutral-900 mt-1">
            {member.date_joined_scheme ? new Date(member.date_joined_scheme).toLocaleDateString() : '-'}
          </p>
        </div>
        <div className="bg-white rounded-2xl border border-neutral-100 p-5">
          <p className="text-sm text-neutral-500">Sponsor</p>
          <p className="text-lg font-semibold text-neutral-900 mt-1 truncate">
            {member.sponsor_name || '-'}
          </p>
        </div>
      </div>

      {/* Tabs */}
      <div className="border-b border-neutral-100">
        <div className="flex gap-1">
          {[
            { id: 'details', label: 'Details', icon: User },
            { id: 'contributions', label: `Contributions (${memberContributions.length})`, icon: CreditCard },
            { id: 'beneficiaries', label: `Beneficiaries (${beneficiaries.length})`, icon: Users },
            { id: 'claims', label: `Claims (${memberClaims.length})`, icon: FileText },
          ].map(tab => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`flex items-center gap-2 px-4 py-3 text-sm font-medium border-b-2 transition-all ${
                activeTab === tab.id
                  ? 'border-neutral-900 text-neutral-900'
                  : 'border-transparent text-neutral-500 hover:text-neutral-700'
              }`}
            >
              <tab.icon size={16} />
              {tab.label}
            </button>
          ))}
        </div>
      </div>

      {/* Tab Content */}
      {activeTab === 'details' && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Personal Info */}
          <div className="bg-white rounded-2xl border border-neutral-100 overflow-hidden">
            <div className="px-6 py-4 border-b border-neutral-50 bg-neutral-50/50">
              <h2 className="text-lg font-medium text-neutral-900">Personal Information</h2>
            </div>
            <div className="p-6 space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <p className="text-sm text-neutral-500">First Name</p>
                  <p className="font-medium text-neutral-900">{member.first_name}</p>
                </div>
                <div>
                  <p className="text-sm text-neutral-500">Last Name</p>
                  <p className="font-medium text-neutral-900">{member.last_name}</p>
                </div>
                <div>
                  <p className="text-sm text-neutral-500">Gender</p>
                  <p className="font-medium text-neutral-900">{member.gender || '-'}</p>
                </div>
                <div>
                  <p className="text-sm text-neutral-500">Date of Birth</p>
                  <p className="font-medium text-neutral-900">
                    {member.date_of_birth ? new Date(member.date_of_birth).toLocaleDateString() : '-'}
                  </p>
                </div>
                <div>
                  <p className="text-sm text-neutral-500">ID Number</p>
                  <p className="font-medium text-neutral-900">{member.id_number || '-'}</p>
                </div>
                <div>
                  <p className="text-sm text-neutral-500">KRA PIN</p>
                  <p className="font-medium text-neutral-900">{member.kra_pin || '-'}</p>
                </div>
                <div>
                  <p className="text-sm text-neutral-500">Nationality</p>
                  <p className="font-medium text-neutral-900">{member.nationality || '-'}</p>
                </div>
                <div>
                  <p className="text-sm text-neutral-500">Marital Status</p>
                  <p className="font-medium text-neutral-900">{member.marital_status || '-'}</p>
                </div>
              </div>
            </div>
          </div>

          {/* Contact Info */}
          <div className="bg-white rounded-2xl border border-neutral-100 overflow-hidden">
            <div className="px-6 py-4 border-b border-neutral-50 bg-neutral-50/50">
              <h2 className="text-lg font-medium text-neutral-900">Contact Information</h2>
            </div>
            <div className="p-6 space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <p className="text-sm text-neutral-500">Email</p>
                  <p className="font-medium text-neutral-900">{member.email || '-'}</p>
                </div>
                <div>
                  <p className="text-sm text-neutral-500">Phone</p>
                  <p className="font-medium text-neutral-900">{member.phone || '-'}</p>
                </div>
                <div>
                  <p className="text-sm text-neutral-500">Town</p>
                  <p className="font-medium text-neutral-900">{member.town || '-'}</p>
                </div>
                <div>
                  <p className="text-sm text-neutral-500">Postal Address</p>
                  <p className="font-medium text-neutral-900">{member.postal_address || '-'}</p>
                </div>
              </div>
            </div>
          </div>

          {/* Employment Info */}
          <div className="bg-white rounded-2xl border border-neutral-100 overflow-hidden">
            <div className="px-6 py-4 border-b border-neutral-50 bg-neutral-50/50">
              <h2 className="text-lg font-medium text-neutral-900">Employment Information</h2>
            </div>
            <div className="p-6 space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <p className="text-sm text-neutral-500">Payroll Number</p>
                  <p className="font-medium text-neutral-900">{member.payroll_no || '-'}</p>
                </div>
                <div>
                  <p className="text-sm text-neutral-500">Designation</p>
                  <p className="font-medium text-neutral-900">{member.designation || '-'}</p>
                </div>
                <div>
                  <p className="text-sm text-neutral-500">Department</p>
                  <p className="font-medium text-neutral-900">{member.department || '-'}</p>
                </div>
                <div>
                  <p className="text-sm text-neutral-500">Date First Appointed</p>
                  <p className="font-medium text-neutral-900">
                    {member.date_first_appt ? new Date(member.date_first_appt).toLocaleDateString() : '-'}
                  </p>
                </div>
                <div>
                  <p className="text-sm text-neutral-500">Expected Retirement</p>
                  <p className="font-medium text-neutral-900">
                    {member.expected_retirement ? new Date(member.expected_retirement).toLocaleDateString() : '-'}
                  </p>
                </div>
              </div>
            </div>
          </div>

          {/* Banking Info */}
          <div className="bg-white rounded-2xl border border-neutral-100 overflow-hidden">
            <div className="px-6 py-4 border-b border-neutral-50 bg-neutral-50/50">
              <h2 className="text-lg font-medium text-neutral-900">Banking Information</h2>
            </div>
            <div className="p-6 space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <p className="text-sm text-neutral-500">Bank Name</p>
                  <p className="font-medium text-neutral-900">{member.bank_name || '-'}</p>
                </div>
                <div>
                  <p className="text-sm text-neutral-500">Bank Branch</p>
                  <p className="font-medium text-neutral-900">{member.bank_branch || '-'}</p>
                </div>
                <div>
                  <p className="text-sm text-neutral-500">Bank Account</p>
                  <p className="font-medium text-neutral-900">{member.bank_account || '-'}</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {activeTab === 'contributions' && (
        <div className="bg-white rounded-2xl border border-neutral-100 overflow-hidden">
          <div className="px-6 py-4 border-b border-neutral-50">
            <h3 className="font-medium text-neutral-900">Contribution History</h3>
          </div>
          {memberContributions.length === 0 ? (
            <div className="p-12 text-center">
              <CreditCard size={32} className="mx-auto text-neutral-300 mb-4" />
              <p className="text-neutral-400">No contributions found for this member</p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead className="bg-neutral-50">
                  <tr>
                    <th className="text-left px-5 py-3 font-medium text-neutral-500">Period</th>
                    <th className="text-right px-5 py-3 font-medium text-neutral-500">Employee</th>
                    <th className="text-right px-5 py-3 font-medium text-neutral-500">Employer</th>
                    <th className="text-right px-5 py-3 font-medium text-neutral-500">AVC</th>
                    <th className="text-right px-5 py-3 font-medium text-neutral-500">Total</th>
                    <th className="text-left px-5 py-3 font-medium text-neutral-500">Status</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-neutral-50">
                  {memberContributions.slice(0, 20).map(c => (
                    <tr key={c.id} className="hover:bg-neutral-50">
                      <td className="px-5 py-3 text-neutral-600">{new Date(c.period).toLocaleDateString()}</td>
                      <td className="px-5 py-3 text-right font-mono text-xs">{formatCurrency(c.employee_amount || 0)}</td>
                      <td className="px-5 py-3 text-right font-mono text-xs">{formatCurrency(c.employer_amount || 0)}</td>
                      <td className="px-5 py-3 text-right font-mono text-xs">{formatCurrency(c.avc_amount || 0)}</td>
                      <td className="px-5 py-3 text-right font-mono text-xs font-semibold">{formatCurrency(c.total_amount || 0)}</td>
                      <td className="px-5 py-3">
                        <span className={`px-2 py-1 rounded-full text-xs ${c.status === 'confirmed' ? 'bg-emerald-50 text-emerald-700' : 'bg-amber-50 text-amber-700'}`}>
                          {c.status}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {activeTab === 'beneficiaries' && (
        <div className="bg-white rounded-2xl border border-neutral-100 overflow-hidden">
          <div className="px-6 py-4 border-b border-neutral-50 flex items-center justify-between">
            <h3 className="font-medium text-neutral-900">Beneficiaries ({beneficiaries.length})</h3>
          </div>
          {beneficiaries.length === 0 ? (
            <div className="p-12 text-center">
              <Users size={32} className="mx-auto text-neutral-300 mb-4" />
              <p className="text-neutral-400">No beneficiaries added yet</p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead className="bg-neutral-50 border-b border-neutral-100">
                  <tr>
                    <th className="text-left px-5 py-3 font-medium text-neutral-500">Name</th>
                    <th className="text-left px-5 py-3 font-medium text-neutral-500">Relationship</th>
                    <th className="text-left px-5 py-3 font-medium text-neutral-500">Date of Birth</th>
                    <th className="text-left px-5 py-3 font-medium text-neutral-500">ID Number</th>
                    <th className="text-right px-5 py-3 font-medium text-neutral-500">Allocation %</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-neutral-50">
                  {beneficiaries.map(b => (
                    <tr key={b.id} className="hover:bg-neutral-50">
                      <td className="px-5 py-4 font-medium text-neutral-900">{b.name}</td>
                      <td className="px-5 py-4 text-neutral-600 capitalize">{b.relationship}</td>
                      <td className="px-5 py-4 text-neutral-600">{b.date_of_birth || '-'}</td>
                      <td className="px-5 py-4 text-neutral-600">{b.id_number || '-'}</td>
                      <td className="px-5 py-4 text-right font-medium text-neutral-900">{b.allocation_pct}%</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {activeTab === 'claims' && (
        <div className="bg-white rounded-2xl border border-neutral-100 overflow-hidden">
          <div className="px-6 py-4 border-b border-neutral-50">
            <h3 className="font-medium text-neutral-900">Claim History</h3>
          </div>
          {memberClaims.length === 0 ? (
            <div className="p-12 text-center">
              <FileText size={32} className="mx-auto text-neutral-300 mb-4" />
              <p className="text-neutral-400">No claims found for this member</p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead className="bg-neutral-50">
                  <tr>
                    <th className="text-left px-5 py-3 font-medium text-neutral-500">Date</th>
                    <th className="text-left px-5 py-3 font-medium text-neutral-500">Type</th>
                    <th className="text-right px-5 py-3 font-medium text-neutral-500">Amount</th>
                    <th className="text-left px-5 py-3 font-medium text-neutral-500">Status</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-neutral-50">
                  {memberClaims.map(c => (
                    <tr key={c.id} className="hover:bg-neutral-50">
                      <td className="px-5 py-3 text-neutral-600">{new Date(c.date_of_claim).toLocaleDateString()}</td>
                      <td className="px-5 py-3 text-neutral-900 capitalize">{c.claim_type?.replace(/_/g, ' ')}</td>
                      <td className="px-5 py-3 text-right font-mono text-xs font-semibold">{formatCurrency(c.amount || 0)}</td>
                      <td className="px-5 py-3">
                        <span className={`px-2 py-1 rounded-full text-xs ${statusColors[c.status] || 'bg-neutral-100 text-neutral-600'}`}>
                          {c.status}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
