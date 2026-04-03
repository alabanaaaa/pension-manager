import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { auth, members } from '../lib/api';
import {
  Shield, Eye, EyeOff, Loader2, AlertCircle, CheckCircle,
  User, Phone, Briefcase, Lock, MapPin, Calendar, Mail,
  ChevronRight, ChevronLeft, Sparkles, PartyPopper,
  ArrowRight, Building2, CreditCard
} from 'lucide-react';

const steps = [
  { id: 'personal', label: 'Personal Info', icon: User },
  { id: 'contact', label: 'Contact Details', icon: Phone },
  { id: 'employment', label: 'Employment', icon: Briefcase },
  { id: 'security', label: 'Security', icon: Lock },
  { id: 'complete', label: 'Complete', icon: PartyPopper },
];

export default function RegisterPage() {
  const navigate = useNavigate();
  const [currentStep, setCurrentStep] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [registrationComplete, setRegistrationComplete] = useState(false);
  const [formData, setFormData] = useState({
    // Personal
    firstName: '',
    lastName: '',
    otherNames: '',
    gender: '',
    dateOfBirth: '',
    nationality: 'Kenyan',
    idNumber: '',
    kraPin: '',
    maritalStatus: '',
    // Contact
    email: '',
    phone: '',
    postalAddress: '',
    postalCode: '',
    town: '',
    // Employment
    memberNo: '',
    payrollNo: '',
    department: '',
    designation: '',
    sponsorId: '',
    dateJoinedScheme: '',
    dateFirstAppt: '',
    expectedRetirement: '',
    basicSalary: '',
    bankName: '',
    bankBranch: '',
    bankAccount: '',
    // Security
    password: '',
    confirmPassword: '',
  });

  const updateField = (field, value) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    setError('');
  };

  const canProceed = () => {
    switch (currentStep) {
      case 0: // Personal
        return formData.firstName && formData.lastName && formData.dateOfBirth && formData.idNumber;
      case 1: // Contact
        return formData.email && formData.phone;
      case 2: // Employment
        return formData.department && formData.designation && formData.dateJoinedScheme && formData.basicSalary;
      case 3: // Security
        return formData.password && formData.password === formData.confirmPassword && formData.password.length >= 6;
      default:
        return false;
    }
  };

  const nextStep = () => {
    if (canProceed()) {
      if (currentStep < steps.length - 1) {
        setCurrentStep(currentStep + 1);
      }
    }
  };

  const prevStep = () => {
    if (currentStep > 0) {
      setCurrentStep(currentStep - 1);
    }
  };

  const handleSubmit = async () => {
    if (!canProceed()) return;
    setLoading(true);
    setError('');

    try {
      // Generate member number if not provided
      const memberNo = formData.memberNo || `MEM-${Date.now().toString(36).toUpperCase()}`;

      // Create member account
      const memberData = {
        member_no: memberNo,
        first_name: formData.firstName,
        last_name: formData.lastName,
        other_names: formData.otherNames,
        gender: formData.gender,
        date_of_birth: formData.dateOfBirth,
        nationality: formData.nationality,
        id_number: formData.idNumber,
        kra_pin: formData.kraPin,
        marital_status: formData.maritalStatus,
        email: formData.email,
        phone: formData.phone,
        postal_address: formData.postalAddress,
        postal_code: formData.postalCode,
        town: formData.town,
        payroll_no: formData.payrollNo,
        department: formData.department,
        designation: formData.designation,
        sponsor_id: formData.sponsorId,
        date_joined_scheme: formData.dateJoinedScheme,
        date_first_appt: formData.dateFirstAppt,
        expected_retirement: formData.expectedRetirement,
        basic_salary: parseInt(formData.basicSalary) * 100, // Convert to cents
        bank_name: formData.bankName,
        bank_branch: formData.bankBranch,
        bank_account: formData.bankAccount,
        membership_status: 'active',
      };

      // Note: In production, you'd create a system user account here
      // For now, we'll just show success
      setRegistrationComplete(true);
      setCurrentStep(4);
    } catch (err) {
      setError(err.response?.data?.error || 'Registration failed. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const progress = ((currentStep) / (steps.length - 1)) * 100;

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-900 via-blue-800 to-slate-900 flex items-center justify-center p-4">
      <div className="w-full max-w-2xl">
        {/* Header */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-16 h-16 bg-white/10 rounded-2xl mb-4">
            <Shield size={32} className="text-blue-300" />
          </div>
          <h1 className="text-3xl font-bold text-white">Pension Manager</h1>
          <p className="text-blue-200 mt-2">Member Registration</p>
        </div>

        {/* Progress bar */}
        {!registrationComplete && (
          <div className="bg-white/10 backdrop-blur-sm rounded-2xl p-4 mb-6">
            <div className="flex items-center justify-between mb-3">
              {steps.slice(0, -1).map((step, i) => {
                const StepIcon = step.icon;
                const isActive = i === currentStep;
                const isCompleted = i < currentStep;
                return (
                  <div key={step.id} className="flex flex-col items-center">
                    <div className={`
                      w-10 h-10 rounded-full flex items-center justify-center transition-all duration-300
                      ${isCompleted ? 'bg-green-500 text-white' : isActive ? 'bg-blue-500 text-white ring-4 ring-blue-500/30' : 'bg-white/20 text-white/50'}
                    `}>
                      {isCompleted ? <CheckCircle size={18} /> : <StepIcon size={18} />}
                    </div>
                    <span className={`text-xs mt-1 ${isActive ? 'text-white' : 'text-white/50'}`}>
                      {step.label}
                    </span>
                  </div>
                );
              })}
            </div>
            <div className="w-full bg-white/20 rounded-full h-2">
              <div
                className="bg-blue-500 h-2 rounded-full transition-all duration-500 ease-out"
                style={{ width: `${progress}%` }}
              />
            </div>
            <p className="text-center text-white/60 text-sm mt-2">
              Step {currentStep + 1} of {steps.length - 1}
            </p>
          </div>
        )}

        {/* Form card */}
        <div className="bg-white rounded-2xl shadow-xl p-8">
          {error && (
            <div className="flex items-center gap-2 p-3 bg-red-50 text-red-700 rounded-lg text-sm mb-5">
              <AlertCircle size={16} />
              <span>{error}</span>
            </div>
          )}

          {/* Step 0: Personal Info */}
          {currentStep === 0 && (
            <div className="space-y-5 animate-in">
              <div className="flex items-center gap-3 mb-6">
                <div className="p-2 bg-blue-100 rounded-lg"><User size={20} className="text-blue-600" /></div>
                <div>
                  <h2 className="text-lg font-semibold">Personal Information</h2>
                  <p className="text-sm text-gray-500">Let's start with your basic details</p>
                </div>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1.5">First Name *</label>
                  <input
                    type="text"
                    value={formData.firstName}
                    onChange={(e) => updateField('firstName', e.target.value)}
                    placeholder="John"
                    className="w-full px-4 py-2.5 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1.5">Last Name *</label>
                  <input
                    type="text"
                    value={formData.lastName}
                    onChange={(e) => updateField('lastName', e.target.value)}
                    placeholder="Doe"
                    className="w-full px-4 py-2.5 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1.5">Other Names</label>
                <input
                  type="text"
                  value={formData.otherNames}
                  onChange={(e) => updateField('otherNames', e.target.value)}
                  placeholder="Middle names (optional)"
                  className="w-full px-4 py-2.5 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1.5">Gender</label>
                  <select
                    value={formData.gender}
                    onChange={(e) => updateField('gender', e.target.value)}
                    className="w-full px-4 py-2.5 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  >
                    <option value="">Select gender</option>
                    <option value="male">Male</option>
                    <option value="female">Female</option>
                    <option value="other">Other</option>
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1.5">Date of Birth *</label>
                  <input
                    type="date"
                    value={formData.dateOfBirth}
                    onChange={(e) => updateField('dateOfBirth', e.target.value)}
                    className="w-full px-4 py-2.5 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1.5">National ID *</label>
                  <input
                    type="text"
                    value={formData.idNumber}
                    onChange={(e) => updateField('idNumber', e.target.value)}
                    placeholder="12345678"
                    className="w-full px-4 py-2.5 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1.5">KRA PIN</label>
                  <input
                    type="text"
                    value={formData.kraPin}
                    onChange={(e) => updateField('kraPin', e.target.value)}
                    placeholder="A001234567B"
                    className="w-full px-4 py-2.5 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1.5">Marital Status</label>
                <select
                  value={formData.maritalStatus}
                  onChange={(e) => updateField('maritalStatus', e.target.value)}
                  className="w-full px-4 py-2.5 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="">Select status</option>
                  <option value="single">Single</option>
                  <option value="married">Married</option>
                  <option value="divorced">Divorced</option>
                  <option value="widowed">Widowed</option>
                </select>
              </div>
            </div>
          )}

          {/* Step 1: Contact Details */}
          {currentStep === 1 && (
            <div className="space-y-5 animate-in">
              <div className="flex items-center gap-3 mb-6">
                <div className="p-2 bg-green-100 rounded-lg"><Phone size={20} className="text-green-600" /></div>
                <div>
                  <h2 className="text-lg font-semibold">Contact Details</h2>
                  <p className="text-sm text-gray-500">How can we reach you?</p>
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1.5">Email Address *</label>
                <div className="relative">
                  <Mail size={18} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
                  <input
                    type="email"
                    value={formData.email}
                    onChange={(e) => updateField('email', e.target.value)}
                    placeholder="john@example.com"
                    className="w-full pl-10 pr-4 py-2.5 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1.5">Phone Number *</label>
                <div className="relative">
                  <Phone size={18} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
                  <input
                    type="tel"
                    value={formData.phone}
                    onChange={(e) => updateField('phone', e.target.value)}
                    placeholder="+254712345678"
                    className="w-full pl-10 pr-4 py-2.5 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1.5">Postal Address</label>
                  <input
                    type="text"
                    value={formData.postalAddress}
                    onChange={(e) => updateField('postalAddress', e.target.value)}
                    placeholder="P.O. Box 12345"
                    className="w-full px-4 py-2.5 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1.5">Postal Code</label>
                  <input
                    type="text"
                    value={formData.postalCode}
                    onChange={(e) => updateField('postalCode', e.target.value)}
                    placeholder="00100"
                    className="w-full px-4 py-2.5 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1.5">Town/City</label>
                <div className="relative">
                  <MapPin size={18} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
                  <input
                    type="text"
                    value={formData.town}
                    onChange={(e) => updateField('town', e.target.value)}
                    placeholder="Nairobi"
                    className="w-full pl-10 pr-4 py-2.5 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
              </div>
            </div>
          )}

          {/* Step 2: Employment */}
          {currentStep === 2 && (
            <div className="space-y-5 animate-in">
              <div className="flex items-center gap-3 mb-6">
                <div className="p-2 bg-purple-100 rounded-lg"><Briefcase size={20} className="text-purple-600" /></div>
                <div>
                  <h2 className="text-lg font-semibold">Employment Details</h2>
                  <p className="text-sm text-gray-500">Tell us about your work</p>
                </div>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1.5">Department *</label>
                  <input
                    type="text"
                    value={formData.department}
                    onChange={(e) => updateField('department', e.target.value)}
                    placeholder="IT, Finance, HR..."
                    className="w-full px-4 py-2.5 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1.5">Designation *</label>
                  <input
                    type="text"
                    value={formData.designation}
                    onChange={(e) => updateField('designation', e.target.value)}
                    placeholder="Software Engineer"
                    className="w-full px-4 py-2.5 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1.5">Date Joined Scheme *</label>
                  <div className="relative">
                    <Calendar size={18} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
                    <input
                      type="date"
                      value={formData.dateJoinedScheme}
                      onChange={(e) => updateField('dateJoinedScheme', e.target.value)}
                      className="w-full pl-10 pr-4 py-2.5 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1.5">Basic Salary (KES) *</label>
                  <div className="relative">
                    <CreditCard size={18} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
                    <input
                      type="number"
                      value={formData.basicSalary}
                      onChange={(e) => updateField('basicSalary', e.target.value)}
                      placeholder="50000"
                      className="w-full pl-10 pr-4 py-2.5 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                  </div>
                </div>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1.5">Payroll Number</label>
                  <input
                    type="text"
                    value={formData.payrollNo}
                    onChange={(e) => updateField('payrollNo', e.target.value)}
                    placeholder="PAY001"
                    className="w-full px-4 py-2.5 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1.5">Sponsor/Employer</label>
                  <div className="relative">
                    <Building2 size={18} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
                    <input
                      type="text"
                      value={formData.sponsorId}
                      onChange={(e) => updateField('sponsorId', e.target.value)}
                      placeholder="Employer name"
                      className="w-full pl-10 pr-4 py-2.5 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                  </div>
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1.5">Expected Retirement Date</label>
                <input
                  type="date"
                  value={formData.expectedRetirement}
                  onChange={(e) => updateField('expectedRetirement', e.target.value)}
                  className="w-full px-4 py-2.5 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>

              <div className="border-t pt-5 mt-5">
                <h3 className="text-sm font-medium text-gray-700 mb-3">Bank Details (Optional)</h3>
                <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                  <input
                    type="text"
                    value={formData.bankName}
                    onChange={(e) => updateField('bankName', e.target.value)}
                    placeholder="Bank name"
                    className="w-full px-4 py-2.5 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                  <input
                    type="text"
                    value={formData.bankBranch}
                    onChange={(e) => updateField('bankBranch', e.target.value)}
                    placeholder="Branch"
                    className="w-full px-4 py-2.5 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                  <input
                    type="text"
                    value={formData.bankAccount}
                    onChange={(e) => updateField('bankAccount', e.target.value)}
                    placeholder="Account number"
                    className="w-full px-4 py-2.5 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
              </div>
            </div>
          )}

          {/* Step 3: Security */}
          {currentStep === 3 && (
            <div className="space-y-5 animate-in">
              <div className="flex items-center gap-3 mb-6">
                <div className="p-2 bg-orange-100 rounded-lg"><Lock size={20} className="text-orange-600" /></div>
                <div>
                  <h2 className="text-lg font-semibold">Create Password</h2>
                  <p className="text-sm text-gray-500">Secure your account</p>
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1.5">Password *</label>
                <div className="relative">
                  <input
                    type={showPassword ? 'text' : 'password'}
                    value={formData.password}
                    onChange={(e) => updateField('password', e.target.value)}
                    placeholder="Minimum 6 characters"
                    className="w-full px-4 py-2.5 pr-10 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600"
                  >
                    {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
                  </button>
                </div>
                {formData.password && (
                  <div className="mt-2">
                    <div className="flex gap-1">
                      {[1, 2, 3, 4].map(i => (
                        <div
                          key={i}
                          className={`h-1.5 flex-1 rounded-full ${
                            formData.password.length >= i * 3
                              ? formData.password.length >= 12 ? 'bg-green-500' : formData.password.length >= 8 ? 'bg-yellow-500' : 'bg-red-500'
                              : 'bg-gray-200'
                          }`}
                        />
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
                <input
                  type={showPassword ? 'text' : 'password'}
                  value={formData.confirmPassword}
                  onChange={(e) => updateField('confirmPassword', e.target.value)}
                  placeholder="Re-enter your password"
                  className={`w-full px-4 py-2.5 border rounded-lg focus:outline-none focus:ring-2 ${
                    formData.confirmPassword && formData.password !== formData.confirmPassword
                      ? 'border-red-500 focus:ring-red-500'
                      : 'focus:ring-blue-500'
                  }`}
                />
                {formData.confirmPassword && formData.password !== formData.confirmPassword && (
                  <p className="text-xs text-red-500 mt-1">Passwords do not match</p>
                )}
              </div>

              <div className="bg-blue-50 rounded-lg p-4 text-sm text-blue-700">
                <p className="font-medium mb-1">Your account will include:</p>
                <ul className="space-y-1">
                  <li>• Access to member portal</li>
                  <li>• Benefit projections and statements</li>
                  <li>• Online voting capabilities</li>
                  <li>• Contribution tracking</li>
                </ul>
              </div>
            </div>
          )}

          {/* Step 4: Complete (Celebration) */}
          {currentStep === 4 && registrationComplete && (
            <div className="text-center py-8 animate-in">
              <div className="relative inline-block">
                <div className="w-24 h-24 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-6 animate-bounce">
                  <PartyPopper size={48} className="text-green-600" />
                </div>
                <div className="absolute -top-2 -right-2 w-8 h-8 bg-yellow-400 rounded-full flex items-center justify-center animate-ping">
                  <Sparkles size={16} className="text-yellow-800" />
                </div>
              </div>

              <h2 className="text-2xl font-bold text-gray-900 mb-2">
                🎉 Congratulations!
              </h2>
              <p className="text-gray-600 mb-6">
                Your registration was successful! Welcome to the Pension Fund Management System.
              </p>

              <div className="bg-green-50 rounded-xl p-5 mb-6 text-left">
                <h3 className="font-semibold text-green-800 mb-3">Your Account Summary</h3>
                <div className="grid grid-cols-2 gap-3 text-sm">
                  <div>
                    <p className="text-green-600">Name</p>
                    <p className="font-medium">{formData.firstName} {formData.lastName}</p>
                  </div>
                  <div>
                    <p className="text-green-600">Email</p>
                    <p className="font-medium">{formData.email}</p>
                  </div>
                  <div>
                    <p className="text-green-600">Department</p>
                    <p className="font-medium">{formData.department}</p>
                  </div>
                  <div>
                    <p className="text-green-600">Phone</p>
                    <p className="font-medium">{formData.phone}</p>
                  </div>
                </div>
              </div>

              <div className="flex flex-col sm:flex-row gap-3 justify-center">
                <Link
                  to="/login"
                  className="px-6 py-3 bg-blue-600 text-white rounded-lg font-medium hover:bg-blue-700 transition-colors flex items-center justify-center gap-2"
                >
                  <ArrowRight size={18} />
                  Go to Login
                </Link>
                <button
                  onClick={() => navigate('/')}
                  className="px-6 py-3 border rounded-lg font-medium hover:bg-gray-50 transition-colors"
                >
                  Back to Home
                </button>
              </div>
            </div>
          )}

          {/* Navigation buttons */}
          {!registrationComplete && (
            <div className="flex items-center justify-between mt-8 pt-6 border-t">
              <button
                onClick={prevStep}
                disabled={currentStep === 0}
                className="flex items-center gap-2 px-4 py-2.5 border rounded-lg text-sm font-medium hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <ChevronLeft size={16} />
                Previous
              </button>

              {currentStep < steps.length - 2 ? (
                <button
                  onClick={nextStep}
                  disabled={!canProceed()}
                  className="flex items-center gap-2 px-6 py-2.5 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Next
                  <ChevronRight size={16} />
                </button>
              ) : (
                <button
                  onClick={handleSubmit}
                  disabled={!canProceed() || loading}
                  className="flex items-center gap-2 px-6 py-2.5 bg-green-600 text-white rounded-lg text-sm font-medium hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {loading ? (
                    <>
                      <Loader2 size={16} className="animate-spin" />
                      Registering...
                    </>
                  ) : (
                    <>
                      Complete Registration
                      <CheckCircle size={16} />
                    </>
                  )}
                </button>
              )}
            </div>
          )}
        </div>

        {/* Login link */}
        {!registrationComplete && (
          <p className="text-center text-blue-200/60 text-sm mt-6">
            Already have an account?{' '}
            <Link to="/login" className="text-white hover:underline font-medium">
              Sign in here
            </Link>
          </p>
        )}
      </div>
    </div>
  );
}
