import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './context/AuthContext';
import DashboardLayout from './layouts/DashboardLayout';
import LoginPage from './pages/LoginPage';
import RegisterPage from './pages/RegisterPage';
import DashboardPage from './pages/DashboardPage';
import MembersPage from './pages/MembersPage';
import AddMemberPage from './pages/AddMemberPage';
import MemberDetailsPage from './pages/MemberDetailsPage';
import EditMemberPage from './pages/EditMemberPage';
import ContributionsPage from './pages/ContributionsPage';
import ClaimsPage from './pages/ClaimsPage';
import ClaimDetailsPage from './pages/ClaimDetailsPage';
import VotingPage from './pages/VotingPage';
import HospitalsPage from './pages/HospitalsPage';
import HospitalDetailsPage from './pages/HospitalDetailsPage';
import ReportsPage from './pages/ReportsPage';
import SponsorsPage from './pages/SponsorsPage';
import BulkProcessingPage from './pages/BulkProcessingPage';
import MakerCheckerPage from './pages/MakerCheckerPage';
import TaxPage from './pages/TaxPage';
import SMSPage from './pages/SMSPage';
import NewsPage from './pages/NewsPage';
import SecurityPage from './pages/SecurityPage';
import SettingsPage from './pages/SettingsPage';
import PortalDashboardPage from './pages/PortalDashboardPage';
import PortalProfilePage from './pages/PortalProfilePage';
import PortalContributionsPage from './pages/PortalContributionsPage';
import PortalClaimsPage from './pages/PortalClaimsPage';
import PortalVotingPage from './pages/PortalVotingPage';
import PortalProjectionsPage from './pages/PortalProjectionsPage';
import PortalFeedbackPage from './pages/PortalFeedbackPage';
import PortalNewsPage from './pages/PortalNewsPage';
import NewElectionPage from './pages/NewElectionPage';
import ManageElectionPage from './pages/ManageElectionPage';
import ElectionResultsPage from './pages/ElectionResultsPage';
import AddHospitalPage from './pages/AddHospitalPage';
import AddSponsorPage from './pages/AddSponsorPage';
import NewClaimPage from './pages/NewClaimPage';
import RecordContributionPage from './pages/RecordContributionPage';
function ProtectedRoute({ children }) {
  const { user, loading } = useAuth();
  if (loading) return <div className="min-h-screen flex items-center justify-center"><div className="animate-spin w-8 h-8 border-2 border-neutral-900 border-t-transparent rounded-full" /></div>;
  if (!user) return <Navigate to="/login" />;
  return children;
}

function AppRoutes() {
  return (
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
