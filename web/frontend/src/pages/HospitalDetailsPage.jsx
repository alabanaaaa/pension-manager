import { useState, useEffect } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { hospitals } from '../lib/api';
import { ArrowLeft, Loader2, MapPin, Phone, Mail, DollarSign, Edit, Building2 } from 'lucide-react';

export default function HospitalDetailsPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [hospital, setHospital] = useState(null);

  useEffect(() => {
    if (!id) { setLoading(false); return; }
    hospitals.get(id)
      .then(res => setHospital(res.data))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [id]);

  if (loading) return <div className="p-8 text-center"><Loader2 className="animate-spin mx-auto" /></div>;
  if (!hospital) return <div className="p-8 text-center">Hospital not found</div>;

  return (
    <div className="space-y-8">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Link to="/hospitals" className="p-2 hover:bg-neutral-100 rounded-lg"><ArrowLeft size={20} /></Link>
          <div className="flex items-center gap-4">
            <div className="w-14 h-14 bg-neutral-100 rounded-xl flex items-center justify-center">
              <Building2 size={24} className="text-neutral-500" />
            </div>
            <div>
              <h1 className="text-2xl font-bold tracking-tight text-black">{hospital.name}</h1>
              <p className="text-neutral-500 mt-1">Medical Facility</p>
            </div>
          </div>
        </div>
        <span className={`px-3 py-1.5 rounded-full text-sm font-medium ${hospital.status === 'active' ? 'bg-emerald-50 text-emerald-700' : 'bg-neutral-100 text-neutral-600'}`}>
          {hospital.status}
        </span>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2 space-y-6">
          <div className="bg-white rounded-2xl border border-neutral-100 p-6">
            <h2 className="text-lg font-semibold text-neutral-900 mb-4">Facility Information</h2>
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div className="flex items-center gap-3">
                <MapPin size={16} className="text-neutral-400" />
                <div>
                  <p className="text-neutral-500">Address</p>
                  <p className="font-medium">{hospital.address || '-'}</p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <Phone size={16} className="text-neutral-400" />
                <div>
                  <p className="text-neutral-500">Phone</p>
                  <p className="font-medium">{hospital.phone || '-'}</p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <Mail size={16} className="text-neutral-400" />
                <div>
                  <p className="text-neutral-500">Email</p>
                  <p className="font-medium">{hospital.email || '-'}</p>
                </div>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-2xl border border-neutral-100 p-6">
            <h2 className="text-lg font-semibold text-neutral-900 mb-4">Medical Limits</h2>
            <div className="grid grid-cols-2 gap-4">
              <div className="p-4 bg-neutral-50 rounded-xl">
                <p className="text-sm text-neutral-500">Inpatient Limit</p>
                <p className="text-xl font-semibold text-neutral-900 mt-1">KES {(hospital.inpatient_limit || 0).toLocaleString()}</p>
              </div>
              <div className="p-4 bg-neutral-50 rounded-xl">
                <p className="text-sm text-neutral-500">Outpatient Limit</p>
                <p className="text-xl font-semibold text-neutral-900 mt-1">KES {(hospital.outpatient_limit || 0).toLocaleString()}</p>
              </div>
              <div className="p-4 bg-neutral-50 rounded-xl">
                <p className="text-sm text-neutral-500">Total Limit</p>
                <p className="text-xl font-semibold text-neutral-900 mt-1">KES {((hospital.inpatient_limit || 0) + (hospital.outpatient_limit || 0)).toLocaleString()}</p>
              </div>
              <div className="p-4 bg-neutral-50 rounded-xl">
                <p className="text-sm text-neutral-500">Current Balance</p>
                <p className="text-xl font-semibold text-neutral-900 mt-1">KES {(hospital.account_balance || 0).toLocaleString()}</p>
              </div>
            </div>
          </div>
        </div>

        <div className="space-y-6">
          <div className="bg-white rounded-2xl border border-neutral-100 p-6">
            <h2 className="text-lg font-semibold text-neutral-900 mb-4">Quick Actions</h2>
            <div className="space-y-3">
              <button className="w-full flex items-center justify-center gap-2 px-4 py-2.5 border border-neutral-200 text-neutral-700 rounded-xl text-sm font-medium hover:bg-neutral-50">
                <Edit size={16} /> Edit Facility
              </button>
              <button className="w-full flex items-center justify-center gap-2 px-4 py-2.5 border border-neutral-200 text-neutral-700 rounded-xl text-sm font-medium hover:bg-neutral-50">
                <DollarSign size={16} /> Update Limits
              </button>
            </div>
          </div>

          <div className="bg-neutral-50 rounded-2xl border border-neutral-100 p-6">
            <h2 className="text-lg font-semibold text-neutral-900 mb-4">Details</h2>
            <div className="space-y-3 text-sm">
              <div className="flex items-center justify-between">
                <span className="text-neutral-500">Facility ID</span>
                <span className="font-medium font-mono">{hospital.id?.slice(0, 8)}</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-neutral-500">Created</span>
                <span className="font-medium">{hospital.created_at ? new Date(hospital.created_at).toLocaleDateString() : '-'}</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-neutral-500">Last Updated</span>
                <span className="font-medium">{hospital.updated_at ? new Date(hospital.updated_at).toLocaleDateString() : '-'}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
