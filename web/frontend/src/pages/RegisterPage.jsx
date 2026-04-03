import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { auth, members } from '../lib/api';
import {
  Shield, Eye, EyeOff, Loader2, AlertCircle, CheckCircle,
  User, Phone, Briefcase, Lock, MapPin, Calendar, Mail,
  ChevronRight, ChevronLeft, Sparkles, PartyPopper,
  ArrowRight, Building2, CreditCard, CheckCircle2
} from 'lucide-react';

const steps = [
  { id: 'personal', label: 'Personal', icon: User },
  { id: 'contact', label: 'Contact', icon: Phone },
  { id: 'employment', label: 'Employment', icon: Briefcase },
  { id: 'security', label: 'Security', icon: Lock },
];

export default function RegisterPage() {
  const navigate = useNavigate();
  const [currentStep, setCurrentStep] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [registrationComplete, setRegistrationComplete] = useState(false);
  const [formData, setFormData] = useState({
    firstName: '', lastName: '', otherNames: '', gender: '', dateOfBirth: '',
    nationality: 'Kenyan', idNumber: '', kraPin: '', maritalStatus: '',
    email: '', phone: '', postalAddress: '', postalCode: '', town: '',
    memberNo: '', payrollNo: '', department: '', designation: '', sponsorId: '',
    dateJoinedScheme: '', dateFirstAppt: '', expectedRetirement: '', basicSalary: '',
    bankName: '', bankBranch: '', bankAccount: '',
    password: '', confirmPassword: '',
  });

  const updateField = (field, value) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    setError('');
  };

  const canProceed = () => {
    switch (currentStep) {
      case 0: return formData.firstName && formData.lastName && formData.dateOfBirth && formData.idNumber;
      case 1: return formData.email && formData.phone;
      case 2: return formData.department && formData.designation && formData.dateJoinedScheme && formData.basicSalary;
      case 3: return formData.password && formData.password === formData.confirmPassword && formData.password.length >= 6;
      default: return false;
    }
  };

  const nextStep = () => { if (canProceed() && currentStep < steps.length - 1) setCurrentStep(currentStep + 1); };
  const prevStep = () => { if (currentStep > 0) setCurrentStep(currentStep - 1); };

  const handleSubmit = async () => {
    if (!canProceed()) return;
    setLoading(true);
    setError('');
    try {
      setRegistrationComplete(true);
    } catch (err) {
      setError(err.response?.data?.error || 'Registration failed. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const progress = ((currentStep) / (steps.length)) * 100;

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-950 via-blue-950 to-indigo-950 flex">
      {/* Left panel - Branding */}
      <div className="hidden lg:flex lg:w-1/2 relative overflow-hidden">
        <div className="absolute inset-0">
          <div className="absolute inset-0 bg-gradient-to-br from-blue-600/20 to-indigo-600/20" />
          <div className="absolute top-20 left-20 w-72 h-72 bg-blue-500/10 rounded-full blur-3xl" />
          <div className="absolute bottom-20 right-20 w-96 h-96 bg-indigo-500/10 rounded-full blur-3xl" />
        </div>

        <div className="relative z-10 flex flex-col justify-center px-16 text-white">
          <div className="inline-flex items-center gap-3 mb-8">
            <div className="w-14 h-14 bg-white/10 backdrop-blur-sm rounded-2xl flex items-center justify-center border border-white/20">
              <Shield size={28} className="text-blue-300" />
            </div>
            <div>
              <h1 className="text-2xl font-bold">Pension Manager</h1>
              <p className="text-blue-200/70 text-sm">Fund Management System</p>
            </div>
          </div>

          <h2 className="text-5xl font-bold leading-tight mb-6">
            Join Your<br />
            <span className="bg-gradient-to-r from-blue-400 to-indigo-400 bg-clip-text text-transparent">
              Pension Scheme
            </span>
          </h2>

          <p className="text-lg text-blue-200/70 max-w-md mb-12">
            Create your account in minutes and start managing your pension benefits online.
          </p>

          {/* Steps preview */}
          <div className="space-y-4">
            {steps.map((step, i) => (
              <div key={step.id} className={`flex items-center gap-4 p-4 rounded-xl border transition-all ${
                i <= currentStep ? 'bg-white/10 backdrop-blur-sm border-white/20' : 'bg-white/5 border-white/5 opacity-50'
              }`}>
                <div className={`w-10 h-10 rounded-lg flex items-center justify-center flex-shrink-0 transition-all ${
                  i < currentStep ? 'bg-green-500' : i === currentStep ? 'bg-blue-500' : 'bg-white/10'
                }`}>
                  {i < currentStep ? <CheckCircle2 size={20} className="text-white" /> : <step.icon size={20} className="text-white" />}
                </div>
                <div>
                  <p className="font-medium">{step.label}</p>
                  <p className="text-sm text-blue-200/50">
                    {i === 0 && 'Name, ID, date of birth'}
                    {i === 1 && 'Email, phone, address'}
                    {i === 2 && 'Department, salary, bank'}
                    {i === 3 && 'Create secure password'}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Right panel - Form */}
      <div className="flex-1 flex items-center justify-center p-6 lg:p-12">
        <div className="w-full max-w-lg">
          {/* Mobile logo */}
          <div className="lg:hidden text-center mb-6">
            <div className="inline-flex items-center justify-center w-14 h-14 bg-white/10 rounded-2xl mb-3">
              <Shield size={28} className="text-blue-300" />
            </div>
            <h1 className="text-xl font-bold text-white">Pension Manager</h1>
          </div>

          {/* Progress */}
          {!registrationComplete && (
            <div className="bg-white/10 backdrop-blur-sm rounded-2xl p-4 mb-6 border border-white/10">
              <div className="flex items-center justify-between mb-3">
                {steps.map((step, i) => {
                  const StepIcon = step.icon;
                  const isActive = i === currentStep;
                  const isCompleted = i < currentStep;
                  return (
                    <div key={step.id} className="flex flex-col items-center">
                      <div className={`w-9 h-9 rounded-full flex items-center justify-center transition-all duration-300 ${
                        isCompleted ? 'bg-green-500 text-white' : isActive ? 'bg-blue-500 text-white ring-4 ring-blue-500/30' : 'bg-white/20 text-white/50'
                      }`}>
                        {isCompleted ? <CheckCircle size={16} /> : <StepIcon size={16} />}
                      </div>
                      <span className={`text-xs mt-1 hidden sm:block ${isActive ? 'text-white' : 'text-white/50'}`}>{step.label}</span>
                    </div>
                  );
                })}
              </div>
              <div className="w-full bg-white/20 rounded-full h-1.5">
                <div className="bg-gradient-to-r from-blue-500 to-indigo-500 h-1.5 rounded-full transition-all duration-500 ease-out" style={{ width: `${progress}%` }} />
              </div>
              <p className="text-center text-white/60 text-xs mt-2">Step {currentStep + 1} of {steps.length}</p>
            </div>
          )}

          {/* Form card */}
          <div className="bg-white rounded-3xl shadow-2xl shadow-black/20 p-8 lg:p-10">
            {error && (
              <div className="flex items-center gap-2 p-3 bg-red-50 text-red-700 rounded-xl text-sm mb-5 border border-red-100">
                <AlertCircle size={16} className="flex-shrink-0" />
                <span>{error}</span>
              </div>
            )}

            {/* Celebration */}
            {currentStep === 4 && registrationComplete ? (
              <div className="text-center py-6 animate-in">
                <div className="relative inline-block mb-6">
                  <div className="w-24 h-24 bg-gradient-to-br from-green-400 to-emerald-500 rounded-full flex items-center justify-center mx-auto animate-bounce-in shadow-lg shadow-green-500/30">
                    <PartyPopper size={48} className="text-white" />
                  </div>
                  <div className="absolute -top-3 -right-3 w-10 h-10 bg-yellow-400 rounded-full flex items-center justify-center animate-ping opacity-75">
                    <Sparkles size={18} className="text-yellow-800" />
                  </div>
                  <div className="absolute -bottom-2 -left-2 w-8 h-8 bg-blue-400 rounded-full flex items-center justify-center animate-ping opacity-75" style={{ animationDelay: '0.5s' }}>
                    <Sparkles size={14} className="text-blue-800" />
                  </div>
                </div>

                <h2 className="text-3xl font-bold text-gray-900 mb-2">
                  🎉 Congratulations!
                </h2>
                <p className="text-gray-600 mb-6">
                  Your registration was successful! Welcome to the Pension Fund Management System.
                </p>

                <div className="bg-gradient-to-r from-green-50 to-emerald-50 rounded-2xl p-5 mb-6 text-left border border-green-100">
                  <h3 className="font-semibold text-green-800 mb-3 flex items-center gap-2">
                    <CheckCircle2 size={18} />
                    Your Account Summary
                  </h3>
                  <div className="grid grid-cols-2 gap-4 text-sm">
                    <div>
                      <p className="text-green-600 text-xs uppercase tracking-wide">Full Name</p>
                      <p className="font-medium text-gray-900">{formData.firstName} {formData.lastName}</p>
                    </div>
                    <div>
                      <p className="text-green-600 text-xs uppercase tracking-wide">Email</p>
                      <p className="font-medium text-gray-900">{formData.email}</p>
                    </div>
                    <div>
                      <p className="text-green-600 text-xs uppercase tracking-wide">Department</p>
                      <p className="font-medium text-gray-900">{formData.department}</p>
                    </div>
                    <div>
                      <p className="text-green-600 text-xs uppercase tracking-wide">Phone</p>
                      <p className="font-medium text-gray-900">{formData.phone}</p>
                    </div>
                  </div>
                </div>

                <div className="flex flex-col sm:flex-row gap-3 justify-center">
                  <Link to="/login" className="px-6 py-3.5 bg-gradient-to-r from-blue-600 to-indigo-600 text-white rounded-xl font-medium hover:from-blue-700 hover:to-indigo-700 transition-all flex items-center justify-center gap-2 shadow-lg shadow-blue-600/25">
                    Go to Login <ArrowRight size={18} />
                  </Link>
                </div>
              </div>
            ) : (
              <>
                {/* Step headers */}
                {currentStep === 0 && (
                  <div className="mb-6">
                    <div className="flex items-center gap-3 mb-2">
                      <div className="p-2.5 bg-blue-100 rounded-xl"><User size={20} className="text-blue-600" /></div>
                      <div>
                        <h2 className="text-xl font-bold text-gray-900">Personal Information</h2>
                        <p className="text-sm text-gray-500">Let's start with your basic details</p>
                      </div>
                    </div>
                  </div>
                )}
                {currentStep === 1 && (
                  <div className="mb-6">
                    <div className="flex items-center gap-3 mb-2">
                      <div className="p-2.5 bg-green-100 rounded-xl"><Phone size={20} className="text-green-600" /></div>
                      <div>
                        <h2 className="text-xl font-bold text-gray-900">Contact Details</h2>
                        <p className="text-sm text-gray-500">How can we reach you?</p>
                      </div>
                    </div>
                  </div>
                )}
                {currentStep === 2 && (
                  <div className="mb-6">
                    <div className="flex items-center gap-3 mb-2">
                      <div className="p-2.5 bg-purple-100 rounded-xl"><Briefcase size={20} className="text-purple-600" /></div>
                      <div>
                        <h2 className="text-xl font-bold text-gray-900">Employment Details</h2>
                        <p className="text-sm text-gray-500">Tell us about your work</p>
                      </div>
                    </div>
                  </div>
                )}
                {currentStep === 3 && (
                  <div className="mb-6">
                    <div className="flex items-center gap-3 mb-2">
                      <div className="p-2.5 bg-orange-100 rounded-xl"><Lock size={20} className="text-orange-600" /></div>
                      <div>
                        <h2 className="text-xl font-bold text-gray-900">Create Password</h2>
                        <p className="text-sm text-gray-500">Secure your account</p>
                      </div>
                    </div>
                  </div>
                )}

                {/* Step 0: Personal */}
                {currentStep === 0 && (
                  <div className="space-y-4 animate-in">
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1.5">First Name *</label>
                        <input type="text" value={formData.firstName} onChange={(e) => updateField('firstName', e.target.value)} placeholder="John" className="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all" />
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1.5">Last Name *</label>
                        <input type="text" value={formData.lastName} onChange={(e) => updateField('lastName', e.target.value)} placeholder="Doe" className="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all" />
                      </div>
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1.5">Other Names</label>
                      <input type="text" value={formData.otherNames} onChange={(e) => updateField('otherNames', e.target.value)} placeholder="Middle names (optional)" className="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all" />
                    </div>
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1.5">Gender</label>
                        <select value={formData.gender} onChange={(e) => updateField('gender', e.target.value)} className="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all">
                          <option value="">Select</option>
                          <option value="male">Male</option>
                          <option value="female">Female</option>
                        </select>
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1.5">Date of Birth *</label>
                        <input type="date" value={formData.dateOfBirth} onChange={(e) => updateField('dateOfBirth', e.target.value)} className="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all" />
                      </div>
                    </div>
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1.5">National ID *</label>
                        <input type="text" value={formData.idNumber} onChange={(e) => updateField('idNumber', e.target.value)} placeholder="12345678" className="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all" />
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1.5">KRA PIN</label>
                        <input type="text" value={formData.kraPin} onChange={(e) => updateField('kraPin', e.target.value)} placeholder="A001234567B" className="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all" />
                      </div>
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1.5">Marital Status</label>
                      <select value={formData.maritalStatus} onChange={(e) => updateField('maritalStatus', e.target.value)} className="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all">
                        <option value="">Select</option>
                        <option value="single">Single</option>
                        <option value="married">Married</option>
                        <option value="divorced">Divorced</option>
                        <option value="widowed">Widowed</option>
                      </select>
                    </div>
                  </div>
                )}

                {/* Step 1: Contact */}
                {currentStep === 1 && (
                  <div className="space-y-4 animate-in">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1.5">Email Address *</label>
                      <div className="relative">
                        <Mail size={18} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
                        <input type="email" value={formData.email} onChange={(e) => updateField('email', e.target.value)} placeholder="john@example.com" className="w-full pl-10 pr-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all" />
                      </div>
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1.5">Phone Number *</label>
                      <div className="relative">
                        <Phone size={18} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
                        <input type="tel" value={formData.phone} onChange={(e) => updateField('phone', e.target.value)} placeholder="+254712345678" className="w-full pl-10 pr-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all" />
                      </div>
                    </div>
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1.5">Postal Address</label>
                        <input type="text" value={formData.postalAddress} onChange={(e) => updateField('postalAddress', e.target.value)} placeholder="P.O. Box 12345" className="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all" />
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1.5">Postal Code</label>
                        <input type="text" value={formData.postalCode} onChange={(e) => updateField('postalCode', e.target.value)} placeholder="00100" className="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all" />
                      </div>
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1.5">Town/City</label>
                      <div className="relative">
                        <MapPin size={18} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
                        <input type="text" value={formData.town} onChange={(e) => updateField('town', e.target.value)} placeholder="Nairobi" className="w-full pl-10 pr-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all" />
                      </div>
                    </div>
                  </div>
                )}

                {/* Step 2: Employment */}
                {currentStep === 2 && (
                  <div className="space-y-4 animate-in">
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1.5">Department *</label>
                        <input type="text" value={formData.department} onChange={(e) => updateField('department', e.target.value)} placeholder="IT, Finance, HR..." className="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all" />
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1.5">Designation *</label>
                        <input type="text" value={formData.designation} onChange={(e) => updateField('designation', e.target.value)} placeholder="Software Engineer" className="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all" />
                      </div>
                    </div>
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1.5">Date Joined *</label>
                        <div className="relative">
                          <Calendar size={18} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
                          <input type="date" value={formData.dateJoinedScheme} onChange={(e) => updateField('dateJoinedScheme', e.target.value)} className="w-full pl-10 pr-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all" />
                        </div>
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1.5">Basic Salary (KES) *</label>
                        <div className="relative">
                          <CreditCard size={18} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
                          <input type="number" value={formData.basicSalary} onChange={(e) => updateField('basicSalary', e.target.value)} placeholder="50000" className="w-full pl-10 pr-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all" />
                        </div>
                      </div>
                    </div>
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1.5">Payroll No</label>
                        <input type="text" value={formData.payrollNo} onChange={(e) => updateField('payrollNo', e.target.value)} placeholder="PAY001" className="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all" />
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1.5">Sponsor/Employer</label>
                        <div className="relative">
                          <Building2 size={18} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
                          <input type="text" value={formData.sponsorId} onChange={(e) => updateField('sponsorId', e.target.value)} placeholder="Employer name" className="w-full pl-10 pr-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all" />
                        </div>
                      </div>
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1.5">Expected Retirement</label>
                      <input type="date" value={formData.expectedRetirement} onChange={(e) => updateField('expectedRetirement', e.target.value)} className="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all" />
                    </div>
                    <div className="border-t pt-4 mt-4">
                      <h3 className="text-sm font-medium text-gray-700 mb-3">Bank Details (Optional)</h3>
                      <div className="grid grid-cols-3 gap-3">
                        <input type="text" value={formData.bankName} onChange={(e) => updateField('bankName', e.target.value)} placeholder="Bank name" className="w-full px-3 py-2.5 bg-gray-50 border border-gray-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all" />
                        <input type="text" value={formData.bankBranch} onChange={(e) => updateField('bankBranch', e.target.value)} placeholder="Branch" className="w-full px-3 py-2.5 bg-gray-50 border border-gray-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all" />
                        <input type="text" value={formData.bankAccount} onChange={(e) => updateField('bankAccount', e.target.value)} placeholder="Account no" className="w-full px-3 py-2.5 bg-gray-50 border border-gray-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all" />
                      </div>
                    </div>
                  </div>
                )}

                {/* Step 3: Security */}
                {currentStep === 3 && (
                  <div className="space-y-4 animate-in">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1.5">Password *</label>
                      <div className="relative">
                        <input type={showPassword ? 'text' : 'password'} value={formData.password} onChange={(e) => updateField('password', e.target.value)} placeholder="Minimum 6 characters" className="w-full px-4 py-3 pr-12 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all" />
                        <button type="button" onClick={() => setShowPassword(!showPassword)} className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 p-1">
                          {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
                        </button>
                      </div>
                      {formData.password && (
                        <div className="mt-2">
                          <div className="flex gap-1">
                            {[1, 2, 3, 4].map(i => (
                              <div key={i} className={`h-1.5 flex-1 rounded-full transition-all ${
                                formData.password.length >= i * 3
                                  ? formData.password.length >= 12 ? 'bg-green-500' : formData.password.length >= 8 ? 'bg-yellow-500' : 'bg-red-500'
                                  : 'bg-gray-200'
                              }`} />
                            ))}
                          </div>
                          <p className="text-xs text-gray-500 mt-1">
                            {formData.password.length < 6 ? 'Too short' : formData.password.length < 8 ? 'Fair' : formData.password.length < 12 ? 'Good' : 'Strong'}
                          </p>
                        </div>
                      )}
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1.5">Confirm Password *</label>
                      <input type={showPassword ? 'text' : 'password'} value={formData.confirmPassword} onChange={(e) => updateField('confirmPassword', e.target.value)} placeholder="Re-enter your password" className={`w-full px-4 py-3 bg-gray-50 border rounded-xl focus:outline-none focus:ring-2 transition-all ${
                        formData.confirmPassword && formData.password !== formData.confirmPassword ? 'border-red-500 focus:ring-red-500/20' : 'border-gray-200 focus:ring-blue-500/20 focus:border-blue-500'
                      }`} />
                      {formData.confirmPassword && formData.password !== formData.confirmPassword && (
                        <p className="text-xs text-red-500 mt-1">Passwords do not match</p>
                      )}
                    </div>
                    <div className="bg-blue-50 rounded-xl p-4 text-sm text-blue-700 border border-blue-100">
                      <p className="font-medium mb-2 flex items-center gap-2"><Shield size={14} /> Your account will include:</p>
                      <ul className="space-y-1 text-blue-600">
                        <li>• Access to member portal</li>
                        <li>• Benefit projections and statements</li>
                        <li>• Online voting capabilities</li>
                        <li>• Contribution tracking</li>
                      </ul>
                    </div>
                  </div>
                )}

                {/* Navigation */}
                <div className="flex items-center justify-between mt-8 pt-6 border-t border-gray-100">
                  <button onClick={prevStep} disabled={currentStep === 0} className="flex items-center gap-2 px-5 py-3 border border-gray-200 rounded-xl text-sm font-medium hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-all">
                    <ChevronLeft size={16} /> Previous
                  </button>
                  {currentStep < steps.length - 1 ? (
                    <button onClick={nextStep} disabled={!canProceed()} className="flex items-center gap-2 px-6 py-3 bg-gradient-to-r from-blue-600 to-indigo-600 text-white rounded-xl text-sm font-medium hover:from-blue-700 hover:to-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed transition-all shadow-lg shadow-blue-600/25">
                      Next <ChevronRight size={16} />
                    </button>
                  ) : (
                    <button onClick={handleSubmit} disabled={!canProceed() || loading} className="flex items-center gap-2 px-6 py-3 bg-gradient-to-r from-green-600 to-emerald-600 text-white rounded-xl text-sm font-medium hover:from-green-700 hover:to-emerald-700 disabled:opacity-50 disabled:cursor-not-allowed transition-all shadow-lg shadow-green-600/25">
                      {loading ? <><Loader2 size={16} className="animate-spin" /> Registering...</> : <>Complete Registration <CheckCircle size={16} /></>}
                    </button>
                  )}
                </div>
              </>
            )}
          </div>

          {!registrationComplete && (
            <p className="text-center text-white/40 text-sm mt-6">
              Already have an account?{' '}
              <Link to="/login" className="text-white hover:underline font-medium">
                Sign in here
              </Link>
            </p>
          )}
        </div>
      </div>
    </div>
  );
}
