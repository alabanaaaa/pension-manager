import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import {
  Eye, EyeOff, Loader2, AlertCircle, CheckCircle,
  User, Mail, Briefcase, Lock, ChevronRight, ChevronLeft,
  PartyPopper, ArrowRight, CheckCircle2, Sparkles,
  Calendar, CreditCard, Building2, MapPin, Phone
} from 'lucide-react';
import bankLogo from '/bank-logo.svg';

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

  if (currentStep === 4) {
    return (
      <div className="min-h-screen bg-white flex flex-col items-center justify-center px-10 py-16">
        <div className="w-full max-w-sm animate-scale-in text-center">
          <div className="mb-10">
            <img src={bankLogo} alt="Logo" className="w-14 h-14 mx-auto" />
          </div>
          <div className="relative inline-block mb-10">
            <div className="w-24 h-24 bg-neutral-900 rounded-3xl flex items-center justify-center mx-auto animate-float">
              <PartyPopper size={36} className="text-white" />
            </div>
            <div className="absolute -top-2 -right-2 w-9 h-9 bg-yellow-400 rounded-full flex items-center justify-center animate-pulse-soft">
              <Sparkles size={16} className="text-yellow-900" />
            </div>
          </div>
          <h1 className="text-4xl font-semibold tracking-tight text-neutral-900 mb-3">Welcome aboard</h1>
          <p className="text-neutral-500 mb-12 text-base">Your account has been created successfully</p>

          <div className="bg-neutral-50 rounded-2xl p-7 mb-10 text-left">
            <h3 className="font-medium text-neutral-900 mb-5 flex items-center gap-2.5 text-sm">
              <CheckCircle2 size={16} /> Account Summary
            </h3>
            <div className="grid grid-cols-2 gap-6 text-sm">
              <div>
                <p className="text-neutral-400 text-xs mb-1">Name</p>
                <p className="font-medium text-neutral-900">{formData.firstName} {formData.lastName}</p>
              </div>
              <div>
                <p className="text-neutral-400 text-xs mb-1">Email</p>
                <p className="font-medium text-neutral-900 truncate">{formData.email}</p>
              </div>
              <div>
                <p className="text-neutral-400 text-xs mb-1">Department</p>
                <p className="font-medium text-neutral-900">{formData.department}</p>
              </div>
              <div>
                <p className="text-neutral-400 text-xs mb-1">Phone</p>
                <p className="font-medium text-neutral-900">{formData.phone}</p>
              </div>
            </div>
          </div>

          <div className="space-y-4">
            <Link to="/login" className="btn-hover w-full flex items-center justify-center gap-2.5 px-6 py-4.5 bg-neutral-900 text-white rounded-xl text-sm font-medium hover:bg-neutral-800 transition-all">
              Go to Login <ArrowRight size={16} />
            </Link>
            <button onClick={() => navigate('/')} className="btn-hover w-full px-6 py-4.5 border border-neutral-200 rounded-xl text-sm font-medium hover:bg-neutral-50 transition-all text-neutral-700">
              Back to Home
            </button>
          </div>
          <p className="text-center text-xs text-neutral-300 mt-10">
            Powered by <span className="font-medium text-neutral-400">minidb</span>
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-white flex">
      {/* Left - Branding */}
      <div className="hidden lg:flex lg:w-1/2 bg-neutral-950 relative overflow-hidden">
        <div className="absolute inset-0">
          <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-blue-600/10 rounded-full blur-3xl" />
          <div className="absolute bottom-1/4 right-1/4 w-80 h-80 bg-indigo-600/10 rounded-full blur-3xl" />
        </div>
        <div className="relative z-10 flex flex-col justify-center px-20 text-white">
          <div className="mb-16">
            <img src={bankLogo} alt="Logo" className="w-16 h-16 opacity-90" />
          </div>
          <h2 className="text-6xl font-light leading-[1.15] tracking-tight mb-8">
            Join Your<br />
            <span className="font-semibold">Pension Scheme</span>
          </h2>
          <p className="text-neutral-400 text-lg max-w-md leading-relaxed mb-20">
            Create your account in minutes and start managing your pension benefits online.
          </p>

          {/* Steps */}
          <div className="space-y-7">
            {steps.map((step, i) => {
              const StepIcon = step.icon;
              const isActive = i === currentStep;
              const isCompleted = i < currentStep;
              return (
                <div key={step.id} className={`flex items-center gap-5 transition-all duration-300 ${isActive ? 'opacity-100' : isCompleted ? 'opacity-60' : 'opacity-30'}`}>
                  <div className={`w-10 h-10 rounded-full flex items-center justify-center flex-shrink-0 transition-all ${
                    isCompleted ? 'bg-white text-neutral-900' : isActive ? 'bg-blue-500 text-white' : 'bg-white/10 text-white/50'
                  }`}>
                    {isCompleted ? <CheckCircle size={18} /> : <StepIcon size={18} />}
                  </div>
                  <div>
                    <p className="text-sm font-medium">{step.label}</p>
                    <p className="text-xs text-neutral-500 mt-1">
                      {i === 0 && 'Name, ID, date of birth'}
                      {i === 1 && 'Email, phone, address'}
                      {i === 2 && 'Department, salary, bank'}
                      {i === 3 && 'Create secure password'}
                    </p>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      </div>

      {/* Right - Form */}
      <div className="flex-1 flex flex-col items-center justify-center px-10 py-16 lg:px-20">
        <div className="w-full max-w-md">
          {/* Mobile logo */}
          <div className="lg:hidden mb-12">
            <img src={bankLogo} alt="Logo" className="w-14 h-14" />
          </div>

          {/* Progress */}
          <div className="mb-12 animate-fade-in-up">
            <div className="flex items-center justify-between mb-4">
              {steps.map((step, i) => {
                const StepIcon = step.icon;
                const isActive = i === currentStep;
                const isCompleted = i < currentStep;
                return (
                  <div key={step.id} className="flex flex-col items-center">
                    <div className={`w-9 h-9 rounded-full flex items-center justify-center transition-all duration-300 ${
                      isCompleted ? 'bg-neutral-900 text-white' : isActive ? 'bg-neutral-900 text-white' : 'bg-neutral-200 text-neutral-400'
                    }`}>
                      {isCompleted ? <CheckCircle size={16} /> : <StepIcon size={16} />}
                    </div>
                  </div>
                );
              })}
            </div>
            <div className="w-full bg-neutral-100 rounded-full h-1">
              <div className="bg-neutral-900 h-1 rounded-full transition-all duration-500 ease-out" style={{ width: `${progress}%` }} />
            </div>
          </div>

          <div className="mb-10 animate-fade-in-up stagger-1">
            <h1 className="text-4xl font-semibold tracking-tight text-neutral-900">Create account</h1>
            <p className="text-neutral-500 mt-3 text-base">Fill in your details to get started</p>
          </div>

          {error && (
            <div className="flex items-center gap-2.5 p-4 bg-red-50 text-red-600 rounded-xl text-sm mb-8 animate-fade-in">
              <AlertCircle size={16} className="flex-shrink-0" />
              <span>{error}</span>
            </div>
          )}

          {/* Step 0: Personal */}
          {currentStep === 0 && (
            <div className="space-y-8 animate-slide-in">
              <div className="grid grid-cols-2 gap-8">
                <div>
                  <label className="block text-xs font-medium text-neutral-500 uppercase tracking-widest mb-3">First Name</label>
                  <input type="text" value={formData.firstName} onChange={(e) => updateField('firstName', e.target.value)} placeholder="John" className="input-focus w-full px-0 py-4 bg-transparent border-b border-neutral-200 text-base focus:outline-none focus:border-neutral-900 transition-colors placeholder:text-neutral-300" />
                </div>
                <div>
                  <label className="block text-xs font-medium text-neutral-500 uppercase tracking-widest mb-3">Last Name</label>
                  <input type="text" value={formData.lastName} onChange={(e) => updateField('lastName', e.target.value)} placeholder="Doe" className="input-focus w-full px-0 py-4 bg-transparent border-b border-neutral-200 text-base focus:outline-none focus:border-neutral-900 transition-colors placeholder:text-neutral-300" />
                </div>
              </div>
              <div>
                <label className="block text-xs font-medium text-neutral-500 uppercase tracking-widest mb-3">Other Names</label>
                <input type="text" value={formData.otherNames} onChange={(e) => updateField('otherNames', e.target.value)} placeholder="Optional" className="input-focus w-full px-0 py-4 bg-transparent border-b border-neutral-200 text-base focus:outline-none focus:border-neutral-900 transition-colors placeholder:text-neutral-300" />
              </div>
              <div className="grid grid-cols-2 gap-8">
                <div>
                  <label className="block text-xs font-medium text-neutral-500 uppercase tracking-widest mb-3">Gender</label>
                  <select value={formData.gender} onChange={(e) => updateField('gender', e.target.value)} className="input-focus w-full px-0 py-4 bg-transparent border-b border-neutral-200 text-base focus:outline-none focus:border-neutral-900 transition-colors text-neutral-700">
                    <option value="">Select</option>
                    <option value="male">Male</option>
                    <option value="female">Female</option>
                  </select>
                </div>
                <div>
                  <label className="block text-xs font-medium text-neutral-500 uppercase tracking-widest mb-3">Date of Birth</label>
                  <input type="date" value={formData.dateOfBirth} onChange={(e) => updateField('dateOfBirth', e.target.value)} className="input-focus w-full px-0 py-4 bg-transparent border-b border-neutral-200 text-base focus:outline-none focus:border-neutral-900 transition-colors text-neutral-700" />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-8">
                <div>
                  <label className="block text-xs font-medium text-neutral-500 uppercase tracking-widest mb-3">National ID</label>
                  <input type="text" value={formData.idNumber} onChange={(e) => updateField('idNumber', e.target.value)} placeholder="12345678" className="input-focus w-full px-0 py-4 bg-transparent border-b border-neutral-200 text-base focus:outline-none focus:border-neutral-900 transition-colors placeholder:text-neutral-300" />
                </div>
                <div>
                  <label className="block text-xs font-medium text-neutral-500 uppercase tracking-widest mb-3">KRA PIN</label>
                  <input type="text" value={formData.kraPin} onChange={(e) => updateField('kraPin', e.target.value)} placeholder="A001234567B" className="input-focus w-full px-0 py-4 bg-transparent border-b border-neutral-200 text-base focus:outline-none focus:border-neutral-900 transition-colors placeholder:text-neutral-300" />
                </div>
              </div>
              <div>
                <label className="block text-xs font-medium text-neutral-500 uppercase tracking-widest mb-3">Marital Status</label>
                <select value={formData.maritalStatus} onChange={(e) => updateField('maritalStatus', e.target.value)} className="input-focus w-full px-0 py-4 bg-transparent border-b border-neutral-200 text-base focus:outline-none focus:border-neutral-900 transition-colors text-neutral-700">
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
            <div className="space-y-8 animate-slide-in">
              <div>
                <label className="block text-xs font-medium text-neutral-500 uppercase tracking-widest mb-3">Email</label>
                <div className="relative">
                  <Mail size={16} className="absolute left-0 top-1/2 -translate-y-1/2 text-neutral-300" />
                  <input type="email" value={formData.email} onChange={(e) => updateField('email', e.target.value)} placeholder="you@example.com" className="input-focus w-full pl-7 pr-0 py-4 bg-transparent border-b border-neutral-200 text-base focus:outline-none focus:border-neutral-900 transition-colors placeholder:text-neutral-300" />
                </div>
              </div>
              <div>
                <label className="block text-xs font-medium text-neutral-500 uppercase tracking-widest mb-3">Phone</label>
                <div className="relative">
                  <Phone size={16} className="absolute left-0 top-1/2 -translate-y-1/2 text-neutral-300" />
                  <input type="tel" value={formData.phone} onChange={(e) => updateField('phone', e.target.value)} placeholder="+254712345678" className="input-focus w-full pl-7 pr-0 py-4 bg-transparent border-b border-neutral-200 text-base focus:outline-none focus:border-neutral-900 transition-colors placeholder:text-neutral-300" />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-8">
                <div>
                  <label className="block text-xs font-medium text-neutral-500 uppercase tracking-widest mb-3">Postal Address</label>
                  <input type="text" value={formData.postalAddress} onChange={(e) => updateField('postalAddress', e.target.value)} placeholder="P.O. Box 12345" className="input-focus w-full px-0 py-4 bg-transparent border-b border-neutral-200 text-base focus:outline-none focus:border-neutral-900 transition-colors placeholder:text-neutral-300" />
                </div>
                <div>
                  <label className="block text-xs font-medium text-neutral-500 uppercase tracking-widest mb-3">Postal Code</label>
                  <input type="text" value={formData.postalCode} onChange={(e) => updateField('postalCode', e.target.value)} placeholder="00100" className="input-focus w-full px-0 py-4 bg-transparent border-b border-neutral-200 text-base focus:outline-none focus:border-neutral-900 transition-colors placeholder:text-neutral-300" />
                </div>
              </div>
              <div>
                <label className="block text-xs font-medium text-neutral-500 uppercase tracking-widest mb-3">Town / City</label>
                <div className="relative">
                  <MapPin size={16} className="absolute left-0 top-1/2 -translate-y-1/2 text-neutral-300" />
                  <input type="text" value={formData.town} onChange={(e) => updateField('town', e.target.value)} placeholder="Nairobi" className="input-focus w-full pl-7 pr-0 py-4 bg-transparent border-b border-neutral-200 text-base focus:outline-none focus:border-neutral-900 transition-colors placeholder:text-neutral-300" />
                </div>
              </div>
            </div>
          )}

          {/* Step 2: Employment */}
          {currentStep === 2 && (
            <div className="space-y-8 animate-slide-in">
              <div className="grid grid-cols-2 gap-8">
                <div>
                  <label className="block text-xs font-medium text-neutral-500 uppercase tracking-widest mb-3">Department</label>
                  <input type="text" value={formData.department} onChange={(e) => updateField('department', e.target.value)} placeholder="IT, Finance..." className="input-focus w-full px-0 py-4 bg-transparent border-b border-neutral-200 text-base focus:outline-none focus:border-neutral-900 transition-colors placeholder:text-neutral-300" />
                </div>
                <div>
                  <label className="block text-xs font-medium text-neutral-500 uppercase tracking-widest mb-3">Designation</label>
                  <input type="text" value={formData.designation} onChange={(e) => updateField('designation', e.target.value)} placeholder="Software Engineer" className="input-focus w-full px-0 py-4 bg-transparent border-b border-neutral-200 text-base focus:outline-none focus:border-neutral-900 transition-colors placeholder:text-neutral-300" />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-8">
                <div>
                  <label className="block text-xs font-medium text-neutral-500 uppercase tracking-widest mb-3">Date Joined</label>
                  <div className="relative">
                    <Calendar size={16} className="absolute left-0 top-1/2 -translate-y-1/2 text-neutral-300" />
                    <input type="date" value={formData.dateJoinedScheme} onChange={(e) => updateField('dateJoinedScheme', e.target.value)} className="input-focus w-full pl-7 pr-0 py-4 bg-transparent border-b border-neutral-200 text-base focus:outline-none focus:border-neutral-900 transition-colors text-neutral-700" />
                  </div>
                </div>
                <div>
                  <label className="block text-xs font-medium text-neutral-500 uppercase tracking-widest mb-3">Basic Salary (KES)</label>
                  <div className="relative">
                    <CreditCard size={16} className="absolute left-0 top-1/2 -translate-y-1/2 text-neutral-300" />
                    <input type="number" value={formData.basicSalary} onChange={(e) => updateField('basicSalary', e.target.value)} placeholder="50000" className="input-focus w-full pl-7 pr-0 py-4 bg-transparent border-b border-neutral-200 text-base focus:outline-none focus:border-neutral-900 transition-colors placeholder:text-neutral-300" />
                  </div>
                </div>
              </div>
              <div className="grid grid-cols-2 gap-8">
                <div>
                  <label className="block text-xs font-medium text-neutral-500 uppercase tracking-widest mb-3">Payroll No</label>
                  <input type="text" value={formData.payrollNo} onChange={(e) => updateField('payrollNo', e.target.value)} placeholder="PAY001" className="input-focus w-full px-0 py-4 bg-transparent border-b border-neutral-200 text-base focus:outline-none focus:border-neutral-900 transition-colors placeholder:text-neutral-300" />
                </div>
                <div>
                  <label className="block text-xs font-medium text-neutral-500 uppercase tracking-widest mb-3">Sponsor</label>
                  <div className="relative">
                    <Building2 size={16} className="absolute left-0 top-1/2 -translate-y-1/2 text-neutral-300" />
                    <input type="text" value={formData.sponsorId} onChange={(e) => updateField('sponsorId', e.target.value)} placeholder="Employer" className="input-focus w-full pl-7 pr-0 py-4 bg-transparent border-b border-neutral-200 text-base focus:outline-none focus:border-neutral-900 transition-colors placeholder:text-neutral-300" />
                  </div>
                </div>
              </div>
              <div>
                <label className="block text-xs font-medium text-neutral-500 uppercase tracking-widest mb-3">Expected Retirement</label>
                <input type="date" value={formData.expectedRetirement} onChange={(e) => updateField('expectedRetirement', e.target.value)} className="input-focus w-full px-0 py-4 bg-transparent border-b border-neutral-200 text-base focus:outline-none focus:border-neutral-900 transition-colors text-neutral-700" />
              </div>
              <div className="border-t border-neutral-100 pt-8 mt-6">
                <h3 className="text-xs font-medium text-neutral-500 uppercase tracking-widest mb-6">Bank Details</h3>
                <div className="grid grid-cols-3 gap-8">
                  <input type="text" value={formData.bankName} onChange={(e) => updateField('bankName', e.target.value)} placeholder="Bank" className="input-focus w-full px-0 py-4 bg-transparent border-b border-neutral-200 text-sm focus:outline-none focus:border-neutral-900 transition-colors placeholder:text-neutral-300" />
                  <input type="text" value={formData.bankBranch} onChange={(e) => updateField('bankBranch', e.target.value)} placeholder="Branch" className="input-focus w-full px-0 py-4 bg-transparent border-b border-neutral-200 text-sm focus:outline-none focus:border-neutral-900 transition-colors placeholder:text-neutral-300" />
                  <input type="text" value={formData.bankAccount} onChange={(e) => updateField('bankAccount', e.target.value)} placeholder="Account" className="input-focus w-full px-0 py-4 bg-transparent border-b border-neutral-200 text-sm focus:outline-none focus:border-neutral-900 transition-colors placeholder:text-neutral-300" />
                </div>
              </div>
            </div>
          )}

          {/* Step 3: Security */}
          {currentStep === 3 && (
            <div className="space-y-8 animate-slide-in">
              <div>
                <label className="block text-xs font-medium text-neutral-500 uppercase tracking-widest mb-3">Password</label>
                <div className="relative">
                  <input type={showPassword ? 'text' : 'password'} value={formData.password} onChange={(e) => updateField('password', e.target.value)} placeholder="Minimum 6 characters" className="input-focus w-full px-0 py-4 pr-12 bg-transparent border-b border-neutral-200 text-base focus:outline-none focus:border-neutral-900 transition-colors placeholder:text-neutral-300" />
                  <button type="button" onClick={() => setShowPassword(!showPassword)} className="absolute right-0 top-1/2 -translate-y-1/2 text-neutral-300 hover:text-neutral-600 p-2 transition-colors">
                    {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
                  </button>
                </div>
                {formData.password && (
                  <div className="mt-4">
                    <div className="flex gap-1.5">
                      {[1, 2, 3, 4].map(i => (
                        <div key={i} className={`h-1 flex-1 rounded-full transition-all duration-300 ${
                          formData.password.length >= i * 3
                            ? formData.password.length >= 12 ? 'bg-neutral-900' : formData.password.length >= 8 ? 'bg-neutral-500' : 'bg-neutral-300'
                            : 'bg-neutral-100'
                        }`} />
                      ))}
                    </div>
                    <p className="text-xs text-neutral-400 mt-2">
                      {formData.password.length < 6 ? 'Too short' : formData.password.length < 8 ? 'Fair' : formData.password.length < 12 ? 'Good' : 'Strong'}
                    </p>
                  </div>
                )}
              </div>
              <div>
                <label className="block text-xs font-medium text-neutral-500 uppercase tracking-widest mb-3">Confirm Password</label>
                <input type={showPassword ? 'text' : 'password'} value={formData.confirmPassword} onChange={(e) => updateField('confirmPassword', e.target.value)} placeholder="Re-enter password" className={`input-focus w-full px-0 py-4 bg-transparent border-b text-base focus:outline-none transition-all placeholder:text-neutral-300 ${
                  formData.confirmPassword && formData.password !== formData.confirmPassword ? 'border-red-300 focus:border-red-500' : 'border-neutral-200 focus:border-neutral-900'
                }`} />
                {formData.confirmPassword && formData.password !== formData.confirmPassword && (
                  <p className="text-xs text-red-500 mt-2">Passwords do not match</p>
                )}
              </div>
              <div className="bg-neutral-50 rounded-2xl p-6">
                <p className="text-sm font-medium text-neutral-900 mb-3 flex items-center gap-2.5"><Lock size={14} /> Your account includes:</p>
                <ul className="space-y-2 text-sm text-neutral-500">
                  <li>• Member portal access</li>
                  <li>• Benefit projections</li>
                  <li>• Online voting</li>
                  <li>• Contribution tracking</li>
                </ul>
              </div>
            </div>
          )}

          {/* Navigation */}
          <div className="flex items-center justify-between mt-12 pt-8 border-t border-neutral-100">
            <button onClick={prevStep} disabled={currentStep === 0} className="btn-hover flex items-center gap-2.5 px-6 py-4 text-neutral-500 rounded-xl text-sm font-medium hover:bg-neutral-50 disabled:opacity-40 disabled:cursor-not-allowed transition-all">
              <ChevronLeft size={16} /> Back
            </button>
            {currentStep < steps.length - 1 ? (
              <button onClick={nextStep} disabled={!canProceed()} className="btn-hover flex items-center gap-2.5 px-7 py-4 bg-neutral-900 text-white rounded-xl text-sm font-medium hover:bg-neutral-800 disabled:opacity-40 disabled:cursor-not-allowed transition-all">
                Continue <ChevronRight size={16} />
              </button>
            ) : (
              <button onClick={handleSubmit} disabled={!canProceed() || loading} className="btn-hover flex items-center gap-2.5 px-7 py-4 bg-neutral-900 text-white rounded-xl text-sm font-medium hover:bg-neutral-800 disabled:opacity-40 disabled:cursor-not-allowed transition-all">
                {loading ? (
                  <><Loader2 size={16} className="animate-spin" /> Creating...</>
                ) : (
                  <>Create Account <CheckCircle2 size={16} /></>
                )}
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
