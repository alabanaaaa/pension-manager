import { useState, useEffect } from 'react';
import { security } from '../lib/api';
import { Loader2, Shield, Ban, Trash2, Plus, AlertTriangle } from 'lucide-react';

export default function SecurityPage() {
  const [ips, setIPs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [newIP, setNewIP] = useState('');
  const [reason, setReason] = useState('');
  const [processing, setProcessing] = useState(false);

  useEffect(() => { fetchIPs(); }, []);

  const fetchIPs = async () => {
    setLoading(true);
    try {
      const res = await security.listBlacklisted();
      setIPs(Array.isArray(res.data) ? res.data : []);
    } catch { setIPs([]); }
    finally { setLoading(false); }
  };

  const handleBlacklist = async () => {
    if (!newIP || !reason) return;
    setProcessing(true);
    try {
      await security.blacklistIP({ ip_address: newIP, reason });
      setNewIP('');
      setReason('');
      fetchIPs();
    } catch (err) {
      console.error(err);
    }
    finally { setProcessing(false); }
  };

  const handleRemove = async (ip) => {
    try {
      await security.removeIP(ip);
      fetchIPs();
    } catch (err) {
      console.error(err);
    }
  };

  return (
    <div className="space-y-8 animate-fade-in-up">
      <div>
        <h1 className="text-3xl font-semibold tracking-tight text-neutral-900">Security</h1>
        <p className="text-neutral-500 mt-2 text-base">IP blacklisting and access control</p>
      </div>

      {/* Add IP */}
      <div className="bg-white rounded-2xl border border-neutral-50 overflow-hidden">
        <div className="flex items-center justify-between px-6 py-5 border-b border-neutral-50">
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-red-50"><Ban size={20} className="text-red-600" /></div>
            <h2 className="text-lg font-semibold tracking-tight text-neutral-900">Blacklist IP Address</h2>
          </div>
        </div>
        <div className="p-6">
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <input
              type="text"
              value={newIP}
              onChange={e => setNewIP(e.target.value)}
              placeholder="IP address (e.g. 192.168.1.1)"
              className="px-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all placeholder:text-neutral-300"
            />
            <input
              type="text"
              value={reason}
              onChange={e => setReason(e.target.value)}
              placeholder="Reason for blacklisting"
              className="px-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all placeholder:text-neutral-300"
            />
            <button
              onClick={handleBlacklist}
              disabled={processing || !newIP || !reason}
              className="btn-hover flex items-center justify-center gap-2 px-4 py-3.5 bg-neutral-900 text-white rounded-xl text-sm font-medium hover:bg-neutral-800 disabled:opacity-50 transition-all"
            >
              {processing ? <Loader2 size={15} className="animate-spin" /> : <><Plus size={15} /> Blacklist</>}
            </button>
          </div>
        </div>
      </div>

      {/* Blacklisted IPs */}
      <div className="bg-white rounded-2xl border border-neutral-50 overflow-hidden">
        <div className="flex items-center justify-between px-6 py-5 border-b border-neutral-50">
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-amber-50"><Shield size={20} className="text-amber-600" /></div>
            <h2 className="text-lg font-semibold tracking-tight text-neutral-900">Blacklisted IPs ({ips.length})</h2>
          </div>
        </div>
        {loading ? (
          <div className="p-16 text-center"><Loader2 size={24} className="animate-spin mx-auto text-neutral-300" /><p className="text-sm text-neutral-400 mt-3">Loading...</p></div>
        ) : ips.length === 0 ? (
          <div className="p-16 text-center"><p className="text-neutral-500">No blacklisted IPs</p></div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-neutral-50">
                  <th className="text-left px-6 py-4 font-medium text-neutral-400 text-xs uppercase tracking-wider">IP Address</th>
                  <th className="text-left px-6 py-4 font-medium text-neutral-400 text-xs uppercase tracking-wider">Reason</th>
                  <th className="text-left px-6 py-4 font-medium text-neutral-400 text-xs uppercase tracking-wider">Date</th>
                  <th className="text-right px-6 py-4 font-medium text-neutral-400 text-xs uppercase tracking-wider">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-neutral-50">
                {ips.map((ip, i) => (
                  <tr key={ip.id || i} className="hover:bg-neutral-50/50 transition-colors">
                    <td className="px-6 py-4 font-mono text-sm">{ip.ip_address}</td>
                    <td className="px-6 py-4 text-neutral-500">{ip.reason}</td>
                    <td className="px-6 py-4 text-neutral-400 text-xs">{new Date(ip.created_at).toLocaleDateString()}</td>
                    <td className="px-6 py-4 text-right">
                      <button onClick={() => handleRemove(ip.ip_address)} className="p-2 hover:bg-red-50 rounded-lg transition-colors text-red-500">
                        <Trash2 size={15} />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
