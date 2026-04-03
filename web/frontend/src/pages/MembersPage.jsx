import { useState, useEffect, useCallback } from 'react';
import { Link } from 'react-router-dom';
import { members } from '../lib/api';
import {
  Search, Plus, Filter, ChevronLeft, ChevronRight,
  Edit, Trash2, Eye, Download, Upload, Loader2
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
      const res = await members.list({
        search,
        limit,
        offset: (page - 1) * limit,
      });
      setData(res.data.members || []);
      setTotal(res.data.total || 0);
    } catch {
      setData([]);
    } finally {
      setLoading(false);
    }
  }, [search, page]);

  useEffect(() => {
    fetchMembers();
  }, [fetchMembers]);

  const totalPages = Math.ceil(total / limit);

  const statusColors = {
    active: 'bg-green-100 text-green-800',
    retired: 'bg-blue-100 text-blue-800',
    deceased: 'bg-gray-100 text-gray-800',
    deferred: 'bg-yellow-100 text-yellow-800',
    withdrawn: 'bg-red-100 text-red-800',
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Members</h1>
          <p className="text-gray-500 mt-1">{total} total members</p>
        </div>
        <div className="flex gap-2">
          <Link
            to="/bulk/import"
            className="flex items-center gap-2 px-4 py-2 border rounded-lg text-sm font-medium hover:bg-gray-50"
          >
            <Upload size={16} />
            Import
          </Link>
          <Link
            to="/members/new"
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700"
          >
            <Plus size={16} />
            Add Member
          </Link>
        </div>
      </div>

      {/* Filters */}
      <div className="bg-white rounded-xl border p-4 flex flex-col sm:flex-row gap-3">
        <div className="relative flex-1">
          <Search size={18} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
          <input
            type="text"
            placeholder="Search by name, member number, ID..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1); }}
            className="w-full pl-10 pr-4 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
        <button className="flex items-center gap-2 px-4 py-2 border rounded-lg text-sm hover:bg-gray-50">
          <Filter size={16} />
          Filters
        </button>
      </div>

      {/* Table */}
      <div className="bg-white rounded-xl border overflow-hidden">
        {loading ? (
          <div className="p-12 text-center">
            <Loader2 size={32} className="animate-spin mx-auto text-gray-400" />
            <p className="text-gray-500 mt-3">Loading members...</p>
          </div>
        ) : data.length === 0 ? (
          <div className="p-12 text-center">
            <p className="text-gray-500">No members found</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="bg-gray-50 border-b">
                <tr>
                  <th className="text-left px-4 py-3 font-medium text-gray-500">Member</th>
                  <th className="text-left px-4 py-3 font-medium text-gray-500 hidden md:table-cell">Member No</th>
                  <th className="text-left px-4 py-3 font-medium text-gray-500 hidden lg:table-cell">Department</th>
                  <th className="text-left px-4 py-3 font-medium text-gray-500">Status</th>
                  <th className="text-right px-4 py-3 font-medium text-gray-500 hidden sm:table-cell">Balance</th>
                  <th className="text-right px-4 py-3 font-medium text-gray-500">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y">
                {data.map((member) => (
                  <tr key={member.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-3">
                        <div className="w-8 h-8 bg-blue-100 rounded-full flex items-center justify-center text-blue-700 font-medium text-xs">
                          {member.first_name?.[0]}{member.last_name?.[0]}
                        </div>
                        <div>
                          <p className="font-medium">{member.first_name} {member.last_name}</p>
                          <p className="text-gray-500 text-xs">{member.email || member.phone}</p>
                        </div>
                      </div>
                    </td>
                    <td className="px-4 py-3 hidden md:table-cell font-mono text-xs">{member.member_no}</td>
                    <td className="px-4 py-3 hidden lg:table-cell text-gray-500">{member.department || '—'}</td>
                    <td className="px-4 py-3">
                      <span className={`px-2 py-1 rounded-full text-xs font-medium ${statusColors[member.membership_status] || 'bg-gray-100 text-gray-800'}`}>
                        {member.membership_status || 'unknown'}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-right hidden sm:table-cell font-mono text-xs">
                      KES {((member.account_balance || 0) / 100).toLocaleString()}
                    </td>
                    <td className="px-4 py-3 text-right">
                      <div className="flex items-center justify-end gap-1">
                        <Link to={`/members/${member.id}`} className="p-1.5 hover:bg-gray-100 rounded" title="View">
                          <Eye size={16} className="text-gray-500" />
                        </Link>
                        <Link to={`/members/${member.id}/edit`} className="p-1.5 hover:bg-gray-100 rounded" title="Edit">
                          <Edit size={16} className="text-gray-500" />
                        </Link>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="px-4 py-3 border-t flex items-center justify-between">
            <p className="text-sm text-gray-500">
              Showing {(page - 1) * limit + 1}–{Math.min(page * limit, total)} of {total}
            </p>
            <div className="flex gap-1">
              <button
                onClick={() => setPage(Math.max(1, page - 1))}
                disabled={page === 1}
                className="p-2 border rounded hover:bg-gray-50 disabled:opacity-50"
              >
                <ChevronLeft size={16} />
              </button>
              {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
                let pageNum;
                if (totalPages <= 5) {
                  pageNum = i + 1;
                } else if (page <= 3) {
                  pageNum = i + 1;
                } else if (page >= totalPages - 2) {
                  pageNum = totalPages - 4 + i;
                } else {
                  pageNum = page - 2 + i;
                }
                return (
                  <button
                    key={pageNum}
                    onClick={() => setPage(pageNum)}
                    className={`w-8 h-8 rounded text-sm font-medium ${page === pageNum ? 'bg-blue-600 text-white' : 'hover:bg-gray-50'}`}
                  >
                    {pageNum}
                  </button>
                );
              })}
              <button
                onClick={() => setPage(Math.min(totalPages, page + 1))}
                disabled={page === totalPages}
                className="p-2 border rounded hover:bg-gray-50 disabled:opacity-50"
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
