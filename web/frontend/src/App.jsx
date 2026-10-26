import { lazy, Suspense } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './context/AuthContext';
import DashboardLayout from './layouts/DashboardLayout';
const LoginPage = lazy(() => import('./pages/LoginPage'));
const RegisterPage = lazy(() => import('./pages/RegisterPage'));
const DashboardPage = lazy(() => import('./pages/DashboardPage'));
const MembersPage = lazy(() => import('./pages/MembersPage'));
const AddMemberPage = lazy(() => import('./pages/AddMemberPage'));
const MemberDetailsPage = lazy(() => import('./pages/MemberDetailsPage'));
const EditMemberPage = lazy(() => import('./pages/EditMemberPage'));
const ContributionsPage = lazy(() => import('./pages/ContributionsPage'));
const ClaimsPage = lazy(() => import('./pages/ClaimsPage'));
const ClaimDetailsPage = lazy(() => import('./pages/ClaimDetailsPage'));
const VotingPage = lazy(() => import('./pages/VotingPage'));
const HospitalsPage = lazy(() => import('./pages/HospitalsPage'));
const HospitalDetailsPage = lazy(() => import('./pages/HospitalDetailsPage'));
const ReportsPage = lazy(() => import('./pages/ReportsPage'));
const SponsorsPage = lazy(() => import('./pages/SponsorsPage'));
const BulkProcessingPage = lazy(() => import('./pages/BulkProcessingPage'));
const MakerCheckerPage = lazy(() => import('./pages/MakerCheckerPage'));
const TaxPage = lazy(() => import('./pages/TaxPage'));
const SMSPage = lazy(() => import('./pages/SMSPage'));
const NewsPage = lazy(() => import('./pages/NewsPage'));
const SecurityPage = lazy(() => import('./pages/SecurityPage'));
const SettingsPage = lazy(() => import('./pages/SettingsPage'));
const PortalDashboardPage = lazy(() => import('./pages/PortalDashboardPage'));
const PortalProfilePage = lazy(() => import('./pages/PortalProfilePage'));
const PortalContributionsPage = lazy(() => import('./pages/PortalContributionsPage'));
const PortalClaimsPage = lazy(() => import('./pages/PortalClaimsPage'));
const PortalVotingPage = lazy(() => import('./pages/PortalVotingPage'));
const PortalProjectionsPage = lazy(() => import('./pages/PortalProjectionsPage'));
const PortalFeedbackPage = lazy(() => import('./pages/PortalFeedbackPage'));
const PortalNewsPage = lazy(() => import('./pages/PortalNewsPage'));
const NewElectionPage = lazy(() => import('./pages/NewElectionPage'));
const ManageElectionPage = lazy(() => import('./pages/ManageElectionPage'));
const ElectionResultsPage = lazy(() => import('./pages/ElectionResultsPage'));
const AddHospitalPage = lazy(() => import('./pages/AddHospitalPage'));
const AddSponsorPage = lazy(() => import('./pages/AddSponsorPage'));
const NewClaimPage = lazy(() => import('./pages/NewClaimPage'));
const RecordContributionPage = lazy(() => import('./pages/RecordContributionPage'));
function ProtectedRoute({ children }) {
  const { user, loading } = useAuth();
  if (loading) return <div className="min-h-screen flex items-center justify-center"><div className="animate-spin w-8 h-8 border-2 border-neutral-900 border-t-transparent rounded-full" /></div>;
  if (!user) return <Navigate to="/login" />;
  return children;
}

function AppRoutes() {
  return (
    <Suspense fallback={<div className="min-h-screen flex items-center justify-center"><div className="animate-spin w-8 h-8 border-2 border-black border-t-transparent rounded-full"></div></div>}>
      <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />
      <Route path="/" element={<ProtectedRoute><DashboardLayout /></ProtectedRoute>}>
        <Route index element={<DashboardPage />} />
        <Route path="members" element={<MembersPage />} />
        <Route path="members/new" element={<AddMemberPage />} />
        <Route path="members/:id" element={<MemberDetailsPage />} />
        <Route path="members/:id/edit" element={<EditMemberPage />} />
        <Route path="contributions" element={<ContributionsPage />} />
        <Route path="contributions/new" element={<RecordContributionPage />} />
        <Route path="claims" element={<ClaimsPage />} />
        <Route path="claims/new" element={<NewClaimPage />} />
        <Route path="claims/:id" element={<ClaimDetailsPage />} />
        <Route path="voting" element={<VotingPage />} />
        <Route path="voting/new" element={<NewElectionPage />} />
        <Route path="voting/:id" element={<ManageElectionPage />} />
        <Route path="voting/:id/results" element={<ElectionResultsPage />} />
        <Route path="hospitals" element={<HospitalsPage />} />
        <Route path="hospitals/new" element={<AddHospitalPage />} />
        <Route path="hospitals/:id" element={<HospitalDetailsPage />} />
        <Route path="sponsors" element={<SponsorsPage />} />
        <Route path="sponsors/new" element={<AddSponsorPage />} />
        <Route path="reports" element={<ReportsPage />} />
        <Route path="bulk" element={<BulkProcessingPage />} />
        <Route path="bulk/import" element={<BulkProcessingPage />} />
        <Route path="maker-checker" element={<MakerCheckerPage />} />
        <Route path="tax" element={<TaxPage />} />
        <Route path="sms" element={<SMSPage />} />
        <Route path="news" element={<NewsPage />} />
        <Route path="security" element={<SecurityPage />} />
        <Route path="settings" element={<SettingsPage />} />
        {/* Member Portal */}
        <Route path="portal" element={<PortalDashboardPage />} />
        <Route path="portal/profile" element={<PortalProfilePage />} />
        <Route path="portal/contributions" element={<PortalContributionsPage />} />
        <Route path="portal/claims" element={<PortalClaimsPage />} />
        <Route path="portal/voting" element={<PortalVotingPage />} />
        <Route path="portal/projections" element={<PortalProjectionsPage />} />
        <Route path="portal/feedback" element={<PortalFeedbackPage />} />
        <Route path="portal/news" element={<PortalNewsPage />} />
        <Route path="portal/settings" element={<SettingsPage />} />
      </Route>
      <Route path="*" element={<Navigate to="/" />} />
    </Routes>
      </Suspense>
  );
}

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <AppRoutes />
      </AuthProvider>
    </BrowserRouter>
  );
}