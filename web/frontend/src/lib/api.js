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

// Portal (Member self-service)
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
  downloadStatementPDF: () => api.get('/portal/statement/pdf', { responseType: 'blob' }),
  projectBenefits: (data) => api.post('/portal/projection', data),
  getBenefitQuote: () => api.get('/portal/projection/quote'),
  uploadPhoto: (formData) => api.post('/portal/photo-upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  }),
  getDocuments: () => api.get('/portal/documents'),
  downloadDocument: (id) => api.get(`/portal/documents/${id}/download`, { responseType: 'blob' }),
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
  // Tax reminders
  getExpiring: (days) => api.get('/tax/reminders/expiring', { params: { days } }),
  getOverdue: () => api.get('/tax/reminders/overdue'),
  getPending: () => api.get('/tax/reminders/pending'),
  sendReminders: () => api.post('/tax/reminders/send'),
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
  get: (params) => api.get('/news', { params }),
  getCategories: () => api.get('/news/categories'),
  refresh: () => api.get('/news/refresh'),
  getPublic: (params) => api.get('/news/public', { params }),
};

// Users (Admin)
export const users = {
  list: () => api.get('/admin/users'),
  create: (data) => api.post('/admin/users', data),
  updateRole: (id, role) => api.put(`/admin/users/${id}/role`, { role }),
  disable: (id) => api.delete(`/admin/users/${id}`),
};

// Pending Changes (Maker-Checker)
export const pendingChanges = {
  list: (params) => api.get('/pending-changes', { params }),
  get: (id) => api.get(`/pending-changes/${id}`),
  approve: (id, notes) => api.post(`/pending-changes/${id}/approve`, { notes }),
  reject: (id, reason) => api.post(`/pending-changes/${id}/reject`, { reason }),
  getCount: () => api.get('/pending-changes/count'),
};

// Dashboard
export const dashboard = {
  get: () => api.get('/dashboard'),
};

// Ghost Mode (Fraud Detection)
export const ghost = {
  get: () => api.get('/ghost'),
};

export default api;
