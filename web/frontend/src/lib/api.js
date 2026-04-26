import axios from 'axios';

const API_BASE = import.meta.env.VITE_API_URL || '/api';

const api = axios.create({
  baseURL: API_BASE,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add auth token and scheme_id to requests
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('access_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  // Add scheme_id from stored user if not already in params
  try {
    const user = JSON.parse(localStorage.getItem('user') || '{}');
    const schemeId = typeof user.scheme_id === 'string' ? user.scheme_id : (user.scheme_id?.String || '');
    if (schemeId && !config.params?.scheme_id) {
      config.params = { ...config.params, scheme_id: schemeId };
    }
  } catch {
    // ignore parse errors
  }
  return config;
});

// Handle 401 responses
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Only redirect if not already on login page
      if (!window.location.pathname.includes('/login')) {
        localStorage.removeItem('access_token');
        localStorage.removeItem('refresh_token');
        localStorage.removeItem('user');
        window.location.href = '/login';
      }
    }
    return Promise.reject(error);
  }
);

// Auth
export const auth = {
  login: (email, password) => api.post('/auth/login', { email, password }),
  memberLogin: (memberNo, pin) => api.post('/auth/member-login', { member_no: memberNo, pin }),
  refreshToken: (refreshToken) => api.post('/auth/refresh', { refresh_token: refreshToken }),
  requestOTP: (email) => api.get('/auth/otp', { params: { email } }),
  unlockUser: (userId) => api.post(`/auth/unlock/${userId}`),
};

// Members
export const members = {
  list: (params) => api.get('/members', { params }),
  get: (id) => api.get(`/members/${id}`),
  create: (data) => api.post('/members', data),
  update: (id, data) => api.put(`/members/${id}`, data),
  delete: (id) => api.delete(`/members/${id}`),
  getBeneficiaries: (id) => api.get(`/members/${id}/beneficiaries`),
  addBeneficiary: (id, data) => api.post(`/members/${id}/beneficiaries`, data),
};

// Contributions
export const contributions = {
  list: (params) => api.get('/contributions', { params }),
  getByMember: (memberId) => api.get(`/contributions/${memberId}`),
  create: (data) => api.post('/contributions', data),
  mpesa: (data) => api.post('/contributions/mpesa', data),
  reconcile: (data) => api.post('/contributions/reconcile', data),
};

// Claims
export const claims = {
  list: (params) => api.get('/claims', { params }),
  get: (id) => api.get(`/claims/${id}`),
  create: (data) => api.post('/claims', data),
  approve: (id, notes) => api.put(`/claims/${id}/approve`, { notes }),
  reject: (id, reason) => api.put(`/claims/${id}/reject`, { reason }),
  pay: (id, data) => api.put(`/claims/${id}/pay`, data),
  partialPayment: (id, data) => api.put(`/claims/${id}/partial-payment`, data),
  getDocuments: (id) => api.get(`/claims/${id}/documents`),
};

// Hospitals
export const hospitals = {
  list: (params) => api.get('/hospitals', { params }),
  get: (id) => api.get(`/hospitals/${id}`),
  create: (data) => api.post('/hospitals', data),
  update: (id, data) => api.put(`/hospitals/${id}`, data),
  getPendingBills: (params) => api.get('/medical-expenditures/pending', { params }),
  getAlerts: (params) => api.get('/medical-expenditures/alerts', { params }),
  recordExpenditure: (data) => api.post('/medical-expenditures', data),
  exportExcel: (params) => api.get('/medical-expenditures/export/excel', { params, responseType: 'blob' }),
};

// Voting
export const voting = {
  listElections: () => api.get('/voting/admin/elections'),
  getElection: (id) => api.get(`/voting/admin/elections/${id}`),
  createElection: (data) => api.post('/voting/admin/elections', data),
  updateStatus: (id, status) => api.put(`/voting/admin/elections/${id}/status`, { status }),
  listCandidates: (electionId) => api.get(`/voting/admin/elections/${electionId}/candidates`),
  addCandidate: (electionId, data) => api.post(`/voting/admin/elections/${electionId}/candidates`, data),
  addVoter: (electionId, data) => api.post(`/voting/admin/elections/${electionId}/voters`, data),
  bulkAddVoters: (electionId, data) => api.post(`/voting/admin/elections/${electionId}/voters/bulk`, data),
  getResults: (electionId) => api.get(`/voting/admin/elections/${electionId}/results`),
  getStats: (electionId) => api.get(`/voting/admin/elections/${electionId}/stats`),
  getVotedMembers: (electionId) => api.get(`/voting/admin/elections/${electionId}/voted-members`),
  getNotVotedMembers: (electionId) => api.get(`/voting/admin/elections/${electionId}/not-voted-members`),
  // Member voting
  memberListElections: () => api.get('/voting/elections'),
  memberGetElection: (id) => api.get(`/voting/elections/${id}`),
  memberListCandidates: (id) => api.get(`/voting/elections/${id}/candidates`),
  memberCastVote: (id, data) => api.post(`/voting/elections/${id}/vote`, data),
  memberGetMyVotes: (id) => api.get(`/voting/elections/${id}/my-votes`),
  memberLiveResults: (id) => api.get(`/voting/elections/${id}/live-results`),
};

// Reports
export const reports = {
  quarterly: (params) => api.get('/reports/quarterly', { params }),
  contributions: (params) => api.get('/reports/contributions', { params }),
  exportCSV: (params) => api.get('/reports/export', { params }),
  // Contribution reports
  breakdown: (params) => api.get('/reports/contributions/breakdown', { params }),
  ytd: (params) => api.get('/reports/contributions/ytd', { params }),
  cumulative: (params) => api.get('/reports/contributions/cumulative', { params }),
  trends: (params) => api.get('/reports/contributions/trends', { params }),
  avcSummary: (params) => api.get('/reports/contributions/avc-summary', { params }),
  exportReport: (params) => api.get('/reports/contributions/export', { params, responseType: 'blob' }),
};

// Bulk Processing
export const bulk = {
  importMembers: (formData) => api.post('/bulk/import/members', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  }),
  validate: (data) => api.post('/bulk/validate', data),
  processRetirements: (data) => api.post('/bulk/process/retirements', data),
  processEarlyLeavers: (data) => api.post('/bulk/process/early-leavers', data),
  annualPosting: (year) => api.post('/bulk/process/annual-posting', {}, { params: { year } }),
  batchStatements: (params) => api.get('/bulk/statements/batch', { params }),
  batchStatementsExport: (params) => api.get('/bulk/statements/batch/export', { params }),
};

// Sponsors
export const sponsor = {
  list: () => api.get('/sponsors'),
  get: (id) => api.get(`/sponsors/${id}`),
  create: (data) => api.post('/sponsors', data),
  update: (id, data) => api.put(`/sponsors/${id}`, data),
  getStats: (id) => api.get(`/sponsors/${id}/stats`),
};

// Pending Changes (Maker-Checker)
export const pendingChanges = {
  list: (params) => api.get('/pending-changes', { params }),
  get: (id) => api.get(`/pending-changes/${id}`),
  approve: (id, notes) => api.post(`/pending-changes/${id}/approve`, { notes }),
  reject: (id, reason) => api.post(`/pending-changes/${id}/reject`, { reason }),
  getCount: () => api.get('/pending-changes/count'),
};

// Pending Member Registrations (Maker-Checker)
export const pendingMembers = {
  list: (params) => api.get('/members/pending', { params }),
  get: (id) => api.get(`/members/pending/${id}`),
  approve: (id) => api.post(`/members/pending/${id}/approve`, {}),
  reject: (id, reason) => api.post(`/members/pending/${id}/reject`, { reason }),
};

// Pending Claims (Maker-Checker)
export const pendingClaims = {
  list: (params) => api.get('/claims/pending', { params }),
  get: (id) => api.get(`/claims/pending/${id}`),
  approve: (id) => api.post(`/claims/pending/${id}/approve`, {}),
  reject: (id, reason) => api.post(`/claims/pending/${id}/reject`, { reason }),
};

// Tax
export const tax = {
  compute: (data) => api.post('/tax/compute', data),
  computeMonthly: (data) => api.post('/tax/compute/monthly', data),
  computeWithdrawal: (data) => api.post('/tax/compute/withdrawal', data),
  computeMultiScheme: (data) => api.post('/tax/compute/multi-scheme', data),
  getBrackets: () => api.get('/tax/brackets'),
  getReliefs: () => api.get('/tax/reliefs'),
  getMemberTaxStatus: (memberId) => api.get(`/tax/member/${memberId}`),
  getExpiring: (days) => api.get('/tax/reminders/expiring', { params: { days } }),
  getOverdue: () => api.get('/tax/reminders/overdue'),
  getPending: () => api.get('/tax/reminders/pending'),
  sendReminders: () => api.post('/tax/reminders/send'),
};

// Tax Exemptions
export const taxExemptions = {
  list: (params) => api.get('/tax-exemptions', { params }),
  get: (id) => api.get(`/tax-exemptions/${id}`),
  create: (data) => api.post('/tax-exemptions', data),
  update: (id, data) => api.put(`/tax-exemptions/${id}`, data),
  delete: (id) => api.delete(`/tax-exemptions/${id}`),
  approve: (id) => api.post(`/tax-exemptions/${id}/approve`),
  reject: (id, reason) => api.post(`/tax-exemptions/${id}/reject`, { reason }),
  getByMember: (memberId) => api.get(`/members/${memberId}/tax-exemptions`),
};

// Annual Statements
export const annualStatements = {
  generate: (data) => api.post('/annual-statements/generate', data),
  list: (params) => api.get('/annual-statements', { params }),
  get: (id) => api.get(`/annual-statements/${id}`),
  downloadPDF: (id) => api.get(`/annual-statements/${id}/pdf`, { responseType: 'blob' }),
  email: (id) => api.post(`/annual-statements/${id}/email`),
  bulkEmail: (data) => api.post('/annual-statements/bulk-email', data),
  hold: (id, reason) => api.put(`/annual-statements/${id}/hold`, { reason }),
  release: (id) => api.put(`/annual-statements/${id}/release`),
  delete: (id) => api.delete(`/annual-statements/${id}`),
  getByMember: (memberId) => api.get(`/members/${memberId}/statements`),
};

// Beneficiary Drawdowns
export const drawdowns = {
  list: (params) => api.get('/drawdowns', { params }),
  get: (id) => api.get(`/drawdowns/${id}`),
  create: (deathId, data) => api.post(`/death-benefits/${deathId}/drawdowns`, data),
  listByDeath: (deathId, params) => api.get(`/death-benefits/${deathId}/drawdowns`, { params }),
  update: (id, data) => api.put(`/drawdowns/${id}`, data),
  approve: (id) => api.post(`/drawdowns/${id}/approve`),
  reject: (id, reason) => api.post(`/drawdowns/${id}/reject`, { reason }),
  process: (id, paymentRef) => api.post(`/drawdowns/${id}/process`, { payment_reference: paymentRef }),
};

// Signatures
export const signatures = {
  sign: (data) => api.post('/signatures/sign', data),
  verify: (entityType, entityId, signerId, signature) => api.get(`/signatures/verify/${entityType}/${entityId}`, { params: { signer_id: signerId, signature } }),
  getByEntity: (entityType, entityId) => api.get(`/signatures/${entityType}/${entityId}`),
  generateMerkle: (startTime, endTime) => api.post('/signatures/merkle/generate', { start_time: startTime, end_time: endTime }),
  getPublicKey: () => api.get('/signatures/public-key'),
  createMultiSigConfig: (data) => api.post('/signatures/multisig/config', data),
  getMultiSigConfig: (entityType) => api.get(`/signatures/multisig/config/${entityType}`),
};

// Medical Expenditures
export const medicalExpenditures = {
  list: (params) => api.get('/medical-expenditures', { params }),
  get: (id) => api.get(`/medical-expenditures/${id}`),
  create: (data) => api.post('/medical-expenditures', data),
  getPendingBills: (params) => api.get('/medical-expenditures/pending', { params }),
  getAlerts: (params) => api.get('/medical-expenditures/alerts', { params }),
};

// SMS
export const sms = {
  send: (data) => api.post('/sms/send', data),
  sendBulk: (data) => api.post('/sms/send/bulk', data),
  sendOTP: (data) => api.post('/sms/send/otp', data),
  sendNotification: (data) => api.post('/sms/send/member-notification', data),
  sendContributionAlert: (data) => api.post('/sms/send/contribution-alert', data),
  sendClaimUpdate: (data) => api.post('/sms/send/claim-update', data),
  sendElectionReminder: (data) => api.post('/sms/send/election-reminder', data),
  getBalance: () => api.get('/sms/balance'),
  getProvider: () => api.get('/sms/provider'),
};

// Security
export const security = {
  blacklistIP: (data) => api.post('/security/ip-blacklist', data),
  removeIP: (ip) => api.delete(`/security/ip-blacklist/${ip}`),
  listBlacklisted: () => api.get('/security/ip-blacklist'),
  checkIP: (ip) => api.get(`/security/ip-blacklist/check/${ip}`),
};

// News
export const news = {
  get: (params) => {
    // Add cache-busting timestamp
    const cacheBust = { _t: Date.now() };
    return api.get('/news', { params: { ...params, ...cacheBust } });
  },
  getCategories: () => api.get('/news/categories'),
  refresh: () => {
    // Add cache-busting
    return api.get('/news/refresh', { params: { _t: Date.now() } });
  },
  getPublic: (params) => {
    // Add cache-busting timestamp
    const cacheBust = { _t: Date.now() };
    return api.get('/news/public', { params: { ...params, ...cacheBust } });
  },
};

// Dashboard
export const dashboard = {
  get: () => api.get('/dashboard'),
};

// Ghost Mode (Fraud Detection)
export const ghost = {
  get: () => api.get('/ghost'),
};

// Portal
export const portal = {
  getProfile: () => api.get('/portal/profile'),
  getBeneficiaries: () => api.get('/portal/beneficiaries'),
  getContributions: (params) => api.get('/portal/contributions', { params }),
  getAnnualContributions: () => api.get('/portal/contributions/annual'),
  getChangeRequests: () => api.get('/portal/change-requests'),
  requestContactChange: (data) => api.post('/portal/change-requests/contact', data),
  requestBeneficiaryChange: (data) => api.post('/portal/change-requests/beneficiary', data),
  submitFeedback: (data) => api.post('/portal/feedback', data),
  getFeedback: () => api.get('/portal/feedback'),
  getLoginStats: () => api.get('/portal/login-stats'),
  getStatement: () => api.get('/portal/statement'),
  downloadStatementPDF: () => api.get('/portal/statement/pdf'),
  projectBenefits: (data) => api.post('/portal/projection', data),
  getBenefitQuote: () => api.get('/portal/projection/quote'),
  uploadPassportPhoto: (data) => {
    const formData = new FormData();
    Object.keys(data).forEach(key => formData.append(key, data[key]));
    return api.post('/portal/photo-upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
  },
  getSchemeDocuments: () => api.get('/portal/documents'),
  downloadSchemeDocument: (id) => api.get(`/portal/documents/${id}/download`),
};

export default api;

