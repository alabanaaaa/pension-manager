import { useState, useEffect, useCallback } from 'react';
import { Link } from 'react-router-dom';
import { members } from '../lib/api';
import {
  Search, Plus, Filter, ChevronLeft, ChevronRight,
  Edit, Eye, Upload, Loader2, ChevronRight as ChevronRightIcon
} from 'lucide-react';

export default function MembersPage() {
  const [data, setData] = useState([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const limit = 20;

  const fetchMembers = useCallback(async () => {
    setLoading(true);
    try {
      const res = await members.list({ search, limit, offset: (page - 1) * limit });
      setData(res.data.members || []);
      setTotal(res.data.total || 0);
    } catch { setData([]); }
    finally { setLoading(false); }
  }, [search, page]);

  useEffect(() => { fetchMembers(); }, [fetchMembers]);

  const totalPages = Math.ceil(total / limit);

  const statusMap = {
    active: { label: 'Active', cls: 'bg-emerald-50 text-emerald-700' },
    retired: { label: 'Retired', cls: 'bg-blue-50 text-blue-700' },
    deceased: { label: 'Deceased', cls: 'bg-neutral-100 text-neutral-600' },
    deferred: { label: 'Deferred', cls: 'bg-amber-50 text-amber-700' },
    withdrawn: { label: 'Withdrawn', cls: 'bg-red-50 text-red-700' },
  };

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-end sm:justify-between gap-5">
        <div>
          <h1 className="text-3xl font-semibold tracking-tight text-neutral-900">Members</h1>
          <p className="text-neutral-500 mt-1">{total} total members</p>
        </div>
        <div className="flex gap-2">
          <Link to="/bulk/import" className="btn-hover flex items-center gap-2 px-4 py-2.5 border border-neutral-200 rounded-2xl text-sm font-medium hover:bg-neutral-50 transition-all">
            <Upload size={15} /> Import
          </Link>
          <Link to="/members/new" className="btn-hover flex items-center gap-2 px-4 py-2.5 bg-neutral-900 text-white rounded-2xl text-sm font-medium hover:bg-neutral-800 transition-all">
            <Plus size={15} /> Add Member
          </Link>
        </div>
      </div>

      {/* Search */}
      <div className="relative">
        <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-neutral-300" />
        <input
          type="text"
          placeholder="Search by name, member number, ID..."
          value={search}
          onChange={(e) => { setSearch(e.target.value); setPage(1); }}
          className="w-full pl-9 pr-4 py-3 bg-white border border-neutral-200 rounded-2xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all placeholder:text-neutral-300"
        />
      </div>

      {/* Table */}
      <div className="bg-white rounded-2xl border border-[#e8e9eb] overflow-hidden">
        {loading ? (
          <div className="p-16 text-center">
            <Loader2 size={24} className="animate-spin mx-auto text-neutral-300" />
            <p className="text-sm text-neutral-400 mt-3">Loading members...</p>
          </div>
        ) : data.length === 0 ? (
          <div className="p-16 text-center">
            <p className="text-sm text-neutral-400">No members found</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-neutral-50">
                  <th className="text-left px-5 py-[18px] font-medium text-neutral-400 text-xs uppercase tracking-wider">Member</th>
                  <th className="text-left px-5 py-[18px] font-medium text-neutral-400 text-xs uppercase tracking-wider hidden md:table-cell">Member No</th>
                  <th className="text-left px-5 py-[18px] font-medium text-neutral-400 text-xs uppercase tracking-wider hidden lg:table-cell">Department</th>
                  <th className="text-left px-5 py-[18px] font-medium text-neutral-400 text-xs uppercase tracking-wider">Status</th>
                  <th className="text-right px-5 py-[18px] font-medium text-neutral-400 text-xs uppercase tracking-wider hidden sm:table-cell">Balance</th>
                  <th className="text-right px-5 py-[18px] font-medium text-neutral-400 text-xs uppercase tracking-wider">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-neutral-50">
                {data.map((member) => {
                  const st = statusMap[member.membership_status] || { label: member.membership_status, cls: 'bg-neutral-50 text-neutral-600' };
                  return (
                    <tr key={member.id} className="hover:bg-neutral-50/50 transition-colors">
                      <td className="px-5 py-[18px]">
                        <div className="flex items-center gap-3">
                          <div className="w-9 h-9 bg-neutral-100 rounded-full flex items-center justify-center text-neutral-600 font-medium text-xs flex-shrink-0">
                            {member.first_name?.[0]}{member.last_name?.[0]}
                          </div>
                          <div>
                            <p className="font-medium text-neutral-900">{member.first_name} {member.last_name}</p>
                            <p className="text-xs text-neutral-400">{member.email || member.phone}</p>
                          </div>
                        </div>
                      </td>
                      <td className="px-5 py-[18px] hidden md:table-cell font-mono text-xs text-neutral-500">{member.member_no}</td>
                      <td className="px-5 py-[18px] hidden lg:table-cell text-neutral-500">{member.department || '—'}</td>
                      <td className="px-5 py-[18px]">
                        <span className={`px-2.5 py-1 rounded-full text-xs font-medium ${st.cls}`}>{st.label}</span>
                      </td>
                      <td className="px-5 py-[18px] text-right hidden sm:table-cell font-mono text-xs text-neutral-600">
                        KES {((member.account_balance || 0) / 100).toLocaleString()}
                      </td>
                      <td className="px-5 py-[18px] text-right">
                        <div className="flex items-center justify-end gap-1">
                          <Link to={`/members/${member.id}`} className="p-2 hover:bg-neutral-100 rounded-lg transition-colors" title="View">
                            <Eye size={15} className="text-neutral-400" />
                          </Link>
                          <Link to={`/members/${member.id}/edit`} className="p-2 hover:bg-neutral-100 rounded-lg transition-colors" title="Edit">
                            <Edit size={15} className="text-neutral-400" />
                          </Link>
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="px-5 py-[18px] border-t border-neutral-50 flex items-center justify-between">
            <p className="text-sm text-neutral-400">
              Showing {(page - 1) * limit + 1}–{Math.min(page * limit, total)} of {total}
            </p>
            <div className="flex gap-1">
              <button onClick={() => setPage(Math.max(1, page - 1))} disabled={page === 1} className="p-2 hover:bg-neutral-50 rounded-lg disabled:opacity-30 transition-colors">
                <ChevronLeft size={16} className="text-neutral-400" />
              </button>
              {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
                let pageNum;
                if (totalPages <= 5) pageNum = i + 1;
                else if (page <= 3) pageNum = i + 1;
                else if (page >= totalPages - 2) pageNum = totalPages - 4 + i;
                else pageNum = page - 2 + i;
                return (
                  <button key={pageNum} onClick={() => setPage(pageNum)} className={`w-8 h-8 rounded-lg text-sm font-medium transition-all ${page === pageNum ? 'bg-neutral-900 text-white' : 'hover:bg-neutral-50 text-neutral-500'}`}>
                    {pageNum}
                  </button>
                );
              })}
              <button onClick={() => setPage(Math.min(totalPages, page + 1))} disabled={page === totalPages} className="p-2 hover:bg-neutral-50 rounded-lg disabled:opacity-30 transition-colors">
                <ChevronRight size={16} className="text-neutral-400" />
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
