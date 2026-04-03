import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import {
  Shield, Eye, EyeOff, Loader2, AlertCircle, CheckCircle,
  User, Mail, Briefcase, Lock, ChevronRight, ChevronLeft,
  PartyPopper, ArrowRight, CheckCircle2, Sparkles,
  Calendar, CreditCard, Building2, MapPin, Phone
} from 'lucide-react';

const steps = [
  { id: 'personal', label: 'Personal', icon: User },
  { id: 'contact', label: 'Contact', icon: Mail },
  { id: 'employment', label: 'Work', icon: Briefcase },
  { id: 'security', label: 'Password', icon: Lock },
];

export default function RegisterPage() {
  const navigate = useNavigate();
  const [currentStep, setCurrentStep] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [formData, setFormData] = useState({
    firstName: '', lastName: '', otherNames: '', gender: '', dateOfBirth: '',
    idNumber: '', kraPin: '', maritalStatus: '',
    email: '', phone: '', postalAddress: '', postalCode: '', town: '',
    department: '', designation: '', dateJoinedScheme: '', basicSalary: '',
    payrollNo: '', sponsorId: '', expectedRetirement: '',
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

  const nextStep = () => {
    if (canProceed() && currentStep < steps.length - 1) {
      setCurrentStep(currentStep + 1);
      window.scrollTo({ top: 0, behavior: 'smooth' });
    }
  };

  const prevStep = () => {
    if (currentStep > 0) {
      setCurrentStep(currentStep - 1);
      window.scrollTo({ top: 0, behavior: 'smooth' });
    }
  };

  const handleSubmit = async () => {
    if (!canProceed() || loading) return;
    setLoading(true);
    setError('');
    try {
      await new Promise(resolve => setTimeout(resolve, 1500));
      setCurrentStep(4);
    } catch (err) {
      setError('Registration failed. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const progress = ((currentStep) / (steps.length)) * 100;

  // Celebration screen
  if (currentStep === 4) {
    return (
      <div className="min-h-screen bg-neutral-50 flex flex-col items-center justify-center px-6 py-12">
        <div className="w-full max-w-sm animate-scale-in">
          <div className="flex items-center gap-3 mb-12">
            <div className="w-10 h-10 bg-neutral-900 rounded-xl flex items-center justify-center">
              <Shield size={20} className="text-white" />
            </div>
            <div>
              <h1 className="text-lg font-semibold text-neutral-900 tracking-tight">Pension Manager</h1>
              <p className="text-xs text-neutral-400">Fund Management System</p>
            </div>
          </div>

          <div className="bg-white rounded-2xl border border-neutral-200 overflow-hidden">
            {/* Success header */}
            <div className="bg-neutral-900 px-8 py-10 text-center">
              <div className="relative inline-block animate-float">
                <div className="w-16 h-16 bg-white/10 rounded-2xl flex items-center justify-center">
                  <PartyPopper size={28} className="text-white" />
                </div>
                <div className="absolute -top-1.5 -right-1.5 w-6 h-6 bg-yellow-400 rounded-full flex items-center justify-center animate-pulse-soft">
                  <Sparkles size={12} className="text-yellow-900" />
                </div>
              </div>
              <h2 className="text-xl font-semibold text-white mt-5 tracking-tight">Welcome aboard</h2>
              <p className="text-neutral-400 text-sm mt-1.5">Your account has been created successfully</p>
            </div>

            {/* Summary */}
            <div className="px-8 py-8">
              <div className="bg-neutral-50 rounded-xl p-5 mb-6 border border-neutral-100">
                <h3 className="font-medium text-neutral-900 mb-4 flex items-center gap-2 text-sm">
                  <CheckCircle2 size={16} className="text-neutral-900" />
                  Account Summary
                </h3>
                <div className="grid grid-cols-2 gap-5 text-sm">
                  <div>
                    <p className="text-neutral-400 text-xs mb-0.5">Name</p>
                    <p className="font-medium text-neutral-900">{formData.firstName} {formData.lastName}</p>
                  </div>
                  <div>
                    <p className="text-neutral-400 text-xs mb-0.5">Email</p>
                    <p className="font-medium text-neutral-900 truncate">{formData.email}</p>
                  </div>
                  <div>
                    <p className="text-neutral-400 text-xs mb-0.5">Department</p>
                    <p className="font-medium text-neutral-900">{formData.department}</p>
                  </div>
                  <div>
                    <p className="text-neutral-400 text-xs mb-0.5">Phone</p>
                    <p className="font-medium text-neutral-900">{formData.phone}</p>
                  </div>
                </div>
              </div>

              <div className="space-y-3">
                <Link
                  to="/login"
                  className="btn-hover w-full flex items-center justify-center gap-2 px-6 py-3.5 bg-neutral-900 text-white rounded-xl text-sm font-medium hover:bg-neutral-800 transition-all"
                >
                  Go to Login <ArrowRight size={16} />
                </Link>
                <button
                  onClick={() => navigate('/')}
                  className="btn-hover w-full px-6 py-3.5 border border-neutral-200 rounded-xl text-sm font-medium hover:bg-neutral-50 transition-all text-neutral-700"
                >
                  Back to Home
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-neutral-50 flex flex-col items-center justify-center px-6 py-12">
      <div className="w-full max-w-md animate-fade-in-up">
        {/* Logo */}
        <div className="flex items-center gap-3 mb-12">
          <div className="w-10 h-10 bg-neutral-900 rounded-xl flex items-center justify-center">
            <Shield size={20} className="text-white" />
          </div>
          <div>
            <h1 className="text-lg font-semibold text-neutral-900 tracking-tight">Pension Manager</h1>
            <p className="text-xs text-neutral-400">Create your account</p>
          </div>
        </div>

        {/* Card */}
        <div className="bg-white rounded-2xl border border-neutral-200 overflow-hidden animate-fade-in-up stagger-1">
          {/* Progress */}
          <div className="px-8 py-6 bg-neutral-50 border-b border-neutral-100">
            <div className="flex items-center justify-between mb-4">
              {steps.map((step, i) => {
                const StepIcon = step.icon;
                const isActive = i === currentStep;
                const isCompleted = i < currentStep;
                return (
                  <div key={step.id} className="flex flex-col items-center">
                    <div className={`w-8 h-8 rounded-full flex items-center justify-center transition-all duration-300 ${
                      isCompleted ? 'bg-neutral-900 text-white' : isActive ? 'bg-neutral-900 text-white' : 'bg-neutral-200 text-neutral-400'
                    }`}>
                      {isCompleted ? <CheckCircle size={14} /> : <StepIcon size={14} />}
                    </div>
                  </div>
                );
              })}
            </div>
            <div className="w-full bg-neutral-200 rounded-full h-1">
              <div className="bg-neutral-900 h-1 rounded-full transition-all duration-500 ease-out" style={{ width: `${progress}%` }} />
            </div>
            <p className="text-xs text-neutral-400 mt-2 text-center">Step {currentStep + 1} of {steps.length}</p>
          </div>

          {/* Form */}
          <div className="px-8 py-8">
            {error && (
              <div className="flex items-center gap-2.5 p-3.5 bg-red-50 text-red-600 rounded-xl text-sm mb-6 animate-fade-in">
                <AlertCircle size={15} className="flex-shrink-0" />
                <span>{error}</span>
              </div>
            )}

            {/* Step 0: Personal */}
            {currentStep === 0 && (
              <div className="space-y-5 animate-slide-in">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-neutral-700 mb-2">First Name</label>
                    <input type="text" value={formData.firstName} onChange={(e) => updateField('firstName', e.target.value)} placeholder="John" className="input-focus w-full px-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all placeholder:text-neutral-400" />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-neutral-700 mb-2">Last Name</label>
                    <input type="text" value={formData.lastName} onChange={(e) => updateField('lastName', e.target.value)} placeholder="Doe" className="input-focus w-full px-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all placeholder:text-neutral-400" />
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium text-neutral-700 mb-2">Other Names</label>
                  <input type="text" value={formData.otherNames} onChange={(e) => updateField('otherNames', e.target.value)} placeholder="Optional" className="input-focus w-full px-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all placeholder:text-neutral-400" />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-neutral-700 mb-2">Gender</label>
                    <select value={formData.gender} onChange={(e) => updateField('gender', e.target.value)} className="input-focus w-full px-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all text-neutral-700">
                      <option value="">Select</option>
                      <option value="male">Male</option>
                      <option value="female">Female</option>
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-neutral-700 mb-2">Date of Birth</label>
                    <input type="date" value={formData.dateOfBirth} onChange={(e) => updateField('dateOfBirth', e.target.value)} className="input-focus w-full px-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all text-neutral-700" />
                  </div>
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-neutral-700 mb-2">National ID</label>
                    <input type="text" value={formData.idNumber} onChange={(e) => updateField('idNumber', e.target.value)} placeholder="12345678" className="input-focus w-full px-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all placeholder:text-neutral-400" />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-neutral-700 mb-2">KRA PIN</label>
                    <input type="text" value={formData.kraPin} onChange={(e) => updateField('kraPin', e.target.value)} placeholder="A001234567B" className="input-focus w-full px-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all placeholder:text-neutral-400" />
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium text-neutral-700 mb-2">Marital Status</label>
                  <select value={formData.maritalStatus} onChange={(e) => updateField('maritalStatus', e.target.value)} className="input-focus w-full px-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all text-neutral-700">
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
              <div className="space-y-5 animate-slide-in">
                <div>
                  <label className="block text-sm font-medium text-neutral-700 mb-2">Email</label>
                  <div className="relative">
                    <Mail size={16} className="absolute left-3.5 top-1/2 -translate-y-1/2 text-neutral-400" />
                    <input type="email" value={formData.email} onChange={(e) => updateField('email', e.target.value)} placeholder="you@example.com" className="input-focus w-full pl-10 pr-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all placeholder:text-neutral-400" />
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium text-neutral-700 mb-2">Phone</label>
                  <div className="relative">
                    <Phone size={16} className="absolute left-3.5 top-1/2 -translate-y-1/2 text-neutral-400" />
                    <input type="tel" value={formData.phone} onChange={(e) => updateField('phone', e.target.value)} placeholder="+254712345678" className="input-focus w-full pl-10 pr-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all placeholder:text-neutral-400" />
                  </div>
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-neutral-700 mb-2">Postal Address</label>
                    <input type="text" value={formData.postalAddress} onChange={(e) => updateField('postalAddress', e.target.value)} placeholder="P.O. Box 12345" className="input-focus w-full px-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all placeholder:text-neutral-400" />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-neutral-700 mb-2">Postal Code</label>
                    <input type="text" value={formData.postalCode} onChange={(e) => updateField('postalCode', e.target.value)} placeholder="00100" className="input-focus w-full px-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all placeholder:text-neutral-400" />
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium text-neutral-700 mb-2">Town / City</label>
                  <div className="relative">
                    <MapPin size={16} className="absolute left-3.5 top-1/2 -translate-y-1/2 text-neutral-400" />
                    <input type="text" value={formData.town} onChange={(e) => updateField('town', e.target.value)} placeholder="Nairobi" className="input-focus w-full pl-10 pr-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all placeholder:text-neutral-400" />
                  </div>
                </div>
              </div>
            )}

            {/* Step 2: Employment */}
            {currentStep === 2 && (
              <div className="space-y-5 animate-slide-in">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-neutral-700 mb-2">Department</label>
                    <input type="text" value={formData.department} onChange={(e) => updateField('department', e.target.value)} placeholder="IT, Finance..." className="input-focus w-full px-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all placeholder:text-neutral-400" />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-neutral-700 mb-2">Designation</label>
                    <input type="text" value={formData.designation} onChange={(e) => updateField('designation', e.target.value)} placeholder="Software Engineer" className="input-focus w-full px-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all placeholder:text-neutral-400" />
                  </div>
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-neutral-700 mb-2">Date Joined</label>
                    <div className="relative">
                      <Calendar size={16} className="absolute left-3.5 top-1/2 -translate-y-1/2 text-neutral-400" />
                      <input type="date" value={formData.dateJoinedScheme} onChange={(e) => updateField('dateJoinedScheme', e.target.value)} className="input-focus w-full pl-10 pr-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all text-neutral-700" />
                    </div>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-neutral-700 mb-2">Basic Salary (KES)</label>
                    <div className="relative">
                      <CreditCard size={16} className="absolute left-3.5 top-1/2 -translate-y-1/2 text-neutral-400" />
                      <input type="number" value={formData.basicSalary} onChange={(e) => updateField('basicSalary', e.target.value)} placeholder="50000" className="input-focus w-full pl-10 pr-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all placeholder:text-neutral-400" />
                    </div>
                  </div>
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-neutral-700 mb-2">Payroll No</label>
                    <input type="text" value={formData.payrollNo} onChange={(e) => updateField('payrollNo', e.target.value)} placeholder="PAY001" className="input-focus w-full px-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all placeholder:text-neutral-400" />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-neutral-700 mb-2">Sponsor</label>
                    <div className="relative">
                      <Building2 size={16} className="absolute left-3.5 top-1/2 -translate-y-1/2 text-neutral-400" />
                      <input type="text" value={formData.sponsorId} onChange={(e) => updateField('sponsorId', e.target.value)} placeholder="Employer" className="input-focus w-full pl-10 pr-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all placeholder:text-neutral-400" />
                    </div>
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium text-neutral-700 mb-2">Expected Retirement</label>
                  <input type="date" value={formData.expectedRetirement} onChange={(e) => updateField('expectedRetirement', e.target.value)} className="input-focus w-full px-4 py-3.5 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all text-neutral-700" />
                </div>
                <div className="border-t border-neutral-100 pt-5 mt-3">
                  <h3 className="text-sm font-medium text-neutral-700 mb-4">Bank Details</h3>
                  <div className="grid grid-cols-3 gap-3">
                    <input type="text" value={formData.bankName} onChange={(e) => updateField('bankName', e.target.value)} placeholder="Bank" className="input-focus w-full px-3.5 py-3 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all placeholder:text-neutral-400" />
                    <input type="text" value={formData.bankBranch} onChange={(e) => updateField('bankBranch', e.target.value)} placeholder="Branch" className="input-focus w-full px-3.5 py-3 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all placeholder:text-neutral-400" />
                    <input type="text" value={formData.bankAccount} onChange={(e) => updateField('bankAccount', e.target.value)} placeholder="Account" className="input-focus w-full px-3.5 py-3 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all placeholder:text-neutral-400" />
                  </div>
                </div>
              </div>
            )}

            {/* Step 3: Security */}
            {currentStep === 3 && (
              <div className="space-y-5 animate-slide-in">
                <div>
                  <label className="block text-sm font-medium text-neutral-700 mb-2">Password</label>
                  <div className="relative">
                    <input type={showPassword ? 'text' : 'password'} value={formData.password} onChange={(e) => updateField('password', e.target.value)} placeholder="Minimum 6 characters" className="input-focus w-full px-4 py-3.5 pr-12 bg-neutral-50 border border-neutral-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-neutral-900/10 focus:border-neutral-900 transition-all placeholder:text-neutral-400" />
                    <button type="button" onClick={() => setShowPassword(!showPassword)} className="absolute right-3 top-1/2 -translate-y-1/2 text-neutral-400 hover:text-neutral-600 p-1.5 transition-colors">
                      {showPassword ? <EyeOff size={16} /> : <Eye size={16} />}
                    </button>
                  </div>
                  {formData.password && (
                    <div className="mt-3">
                      <div className="flex gap-1">
                        {[1, 2, 3, 4].map(i => (
                          <div key={i} className={`h-1 flex-1 rounded-full transition-all duration-300 ${
                            formData.password.length >= i * 3
                              ? formData.password.length >= 12 ? 'bg-neutral-900' : formData.password.length >= 8 ? 'bg-neutral-500' : 'bg-neutral-300'
                              : 'bg-neutral-100'
                          }`} />
                        ))}
                      </div>
                      <p className="text-xs text-neutral-400 mt-1.5">
                        {formData.password.length < 6 ? 'Too short' : formData.password.length < 8 ? 'Fair' : formData.password.length < 12 ? 'Good' : 'Strong'}
                      </p>
                    </div>
                  )}
                </div>
                <div>
                  <label className="block text-sm font-medium text-neutral-700 mb-2">Confirm Password</label>
                  <input type={showPassword ? 'text' : 'password'} value={formData.confirmPassword} onChange={(e) => updateField('confirmPassword', e.target.value)} placeholder="Re-enter password" className={`input-focus w-full px-4 py-3.5 bg-neutral-50 border rounded-xl text-sm focus:outline-none focus:ring-2 transition-all placeholder:text-neutral-400 ${
                    formData.confirmPassword && formData.password !== formData.confirmPassword ? 'border-red-300 focus:ring-red-500/10 focus:border-red-400' : 'border-neutral-200 focus:ring-neutral-900/10 focus:border-neutral-900'
                  }`} />
                  {formData.confirmPassword && formData.password !== formData.confirmPassword && (
                    <p className="text-xs text-red-500 mt-1.5">Passwords do not match</p>
                  )}
                </div>
                <div className="bg-neutral-50 rounded-xl p-5 border border-neutral-100">
                  <p className="text-sm font-medium text-neutral-900 mb-2.5 flex items-center gap-2"><Shield size={14} /> Your account includes:</p>
                  <ul className="space-y-1.5 text-sm text-neutral-500">
                    <li>• Member portal access</li>
                    <li>• Benefit projections</li>
                    <li>• Online voting</li>
                    <li>• Contribution tracking</li>
                  </ul>
                </div>
              </div>
            )}

            {/* Navigation */}
            <div className="flex items-center justify-between mt-10 pt-6 border-t border-neutral-100">
              <button
                onClick={prevStep}
                disabled={currentStep === 0}
                className="btn-hover flex items-center gap-2 px-5 py-3.5 border border-neutral-200 rounded-xl text-sm font-medium hover:bg-neutral-50 disabled:opacity-40 disabled:cursor-not-allowed transition-all"
              >
                <ChevronLeft size={16} /> Back
              </button>
              {currentStep < steps.length - 1 ? (
                <button
                  onClick={nextStep}
                  disabled={!canProceed()}
                  className="btn-hover flex items-center gap-2 px-6 py-3.5 bg-neutral-900 text-white rounded-xl text-sm font-medium hover:bg-neutral-800 disabled:opacity-40 disabled:cursor-not-allowed transition-all"
                >
                  Continue <ChevronRight size={16} />
                </button>
              ) : (
                <button
                  onClick={handleSubmit}
                  disabled={!canProceed() || loading}
                  className="btn-hover flex items-center gap-2 px-6 py-3.5 bg-neutral-900 text-white rounded-xl text-sm font-medium hover:bg-neutral-800 disabled:opacity-40 disabled:cursor-not-allowed transition-all"
                >
                  {loading ? (
                    <>
                      <Loader2 size={16} className="animate-spin" />
                      Creating...
                    </>
                  ) : (
                    <>
                      Create Account
                      <CheckCircle2 size={16} />
                    </>
                  )}
                </button>
              )}
            </div>
          </div>
        </div>

        {/* Footer */}
        <p className="text-center text-sm text-neutral-400 mt-8 animate-fade-in-up stagger-5">
          Already have an account?{' '}
          <Link to="/login" className="text-neutral-900 hover:underline font-medium">
            Sign in
          </Link>
        </p>
      </div>
    </div>
  );
}
