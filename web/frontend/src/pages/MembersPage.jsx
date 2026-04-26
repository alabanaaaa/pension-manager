import { useState, useEffect, useCallback } from 'react';
import { Link } from 'react-router-dom';
import { members } from '../lib/api';
import {
  Search, Plus, ChevronLeft, ChevronRight,
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
      const data = Array.isArray(res.data) ? res.data : (res.data?.members || []);
      setData(data);
      setTotal(res.data?.total || data.length);
    } catch { setData([]); setTotal(0); }
    finally { setLoading(false); }
  }, [search, page]);

  useEffect(() => { fetchMembers(); }, [fetchMembers]);

  const totalPages = Math.ceil(total / limit);

  const statusMap = {
    active: { label: 'Active', cls: 'badge-success' },
    retired: { label: 'Retired', cls: 'badge-info' },
    deceased: { label: 'Deceased', cls: 'badge-neutral' },
    deferred: { label: 'Deferred', cls: 'badge-warning' },
    withdrawn: { label: 'Withdrawn', cls: 'badge-error' },
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-black">Members</h1>
          <p className="text-sm text-gray-500 mt-1">{total} total members</p>
        </div>
        <div className="flex gap-2">
          <Link to="/bulk/import" className="btn btn-secondary flex items-center gap-2">
            <Upload size={14} /> Import
          </Link>
          <Link to="/members/new" className="btn btn-primary flex items-center gap-2">
            <Plus size={14} /> Add Member
          </Link>
        </div>
      </div>

      {/* Search */}
      <div className="filter-bar">
        <div className="table-search flex-1 max-w-md">
          <Search size={14} className="table-search-icon" />
          <input
            type="text"
            placeholder="Search by name, member number, ID..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1); }}
            className="input pl-9"
          />
        </div>
      </div>

      {/* Table - Uber Style */}
      <div className="card">
        {loading ? (
          <div className="p-16 text-center">
            <Loader2 size={24} className="animate-spin mx-auto text-gray-300" />
            <p className="text-sm text-gray-400 mt-3">Loading members...</p>
          </div>
        ) : data.length === 0 ? (
          <div className="empty-state">
            <p className="text-sm text-gray-400">No members found</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="table">
              <thead>
                <tr>
                  <th className="text-left">Member</th>
                  <th className="text-left hidden md:table-cell">Member No</th>
                  <th className="text-left hidden lg:table-cell">Department</th>
                  <th className="text-left">Status</th>
                  <th className="text-right hidden sm:table-cell">Balance</th>
                  <th className="text-right">Actions</th>
                </tr>
              </thead>
              <tbody>
                {data.map((member) => {
                  const st = statusMap[member.membership_status] || { label: member.membership_status, cls: 'badge-neutral' };
                  return (
                    <tr key={member.id}>
                      <td>
                        <div className="flex items-center gap-3">
                          <div className="w-9 h-9 bg-black text-white rounded-full flex items-center justify-center text-xs font-semibold flex-shrink-0">
                            {member.first_name?.[0]}{member.last_name?.[0]}
                          </div>
                          <div>
                            <p className="font-medium text-black">{member.first_name} {member.last_name}</p>
                            <p className="text-xs text-gray-400">{member.email || member.phone || '—'}</p>
                          </div>
                        </div>
                      </td>
                      <td className="hidden md:table-cell font-mono text-xs text-gray-500">{member.member_no}</td>
                      <td className="hidden lg:table-cell text-gray-500">{member.department || '—'}</td>
                      <td>
                        <span className={`badge ${st.cls}`}>{st.label}</span>
                      </td>
                      <td className="text-right hidden sm:table-cell font-mono text-xs">
                        KES {((member.account_balance || 0) / 100).toLocaleString()}
                      </td>
                      <td className="text-right">
                        <div className="flex items-center justify-end gap-1">
                          <Link to={`/members/${member.id}`} className="action-menu" title="View">
                            <Eye size={15} />
                          </Link>
                          <Link to={`/members/${member.id}/edit`} className="action-menu" title="Edit">
                            <Edit size={15} />
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
          <div className="card-footer flex items-center justify-between">
            <p className="text-sm text-gray-500">
              Showing {(page - 1) * limit + 1}–{Math.min(page * limit, total)} of {total}
            </p>
            <div className="flex gap-1">
              <button 
                onClick={() => setPage(Math.max(1, page - 1))} 
                disabled={page === 1} 
                className="action-menu"
              >
                <ChevronLeft size={16} />
              </button>
              {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
                let pageNum;
                if (totalPages <= 5) pageNum = i + 1;
                else if (page <= 3) pageNum = i + 1;
                else if (page >= totalPages - 2) pageNum = totalPages - 4 + i;
                else pageNum = page - 2 + i;
                return (
                  <button 
                    key={pageNum} 
                    onClick={() => setPage(pageNum)} 
                    className={`w-8 h-8 rounded text-sm font-medium transition-all ${
                      page === pageNum ? 'bg-black text-white' : 'hover:bg-gray-100 text-gray-500'
                    }`}
                  >
                    {pageNum}
                  </button>
                );
              })}
              <button 
                onClick={() => setPage(Math.min(totalPages, page + 1))} 
                disabled={page === totalPages} 
                className="action-menu"
              >
                <ChevronRight size={16} />
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
