import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './context/AuthContext';
import DashboardLayout from './layouts/DashboardLayout';
import LoginPage from './pages/LoginPage';
import DashboardPage from './pages/DashboardPage';
import MembersPage from './pages/MembersPage';
import ContributionsPage from './pages/ContributionsPage';
import ClaimsPage from './pages/ClaimsPage';
import VotingPage from './pages/VotingPage';
import HospitalsPage from './pages/HospitalsPage';
import ReportsPage from './pages/ReportsPage';
import PlaceholderPage from './pages/PlaceholderPage';

function ProtectedRoute({ children }) {
  const { user, loading } = useAuth();
  if (loading) return <div className="min-h-screen flex items-center justify-center"><div className="animate-spin w-8 h-8 border-2 border-blue-600 border-t-transparent rounded-full" /></div>;
  if (!user) return <Navigate to="/login" />;
  return children;
}

function AppRoutes() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/" element={<ProtectedRoute><DashboardLayout /></ProtectedRoute>}>
        <Route index element={<DashboardPage />} />
        <Route path="members" element={<MembersPage />} />
        <Route path="members/new" element={<PlaceholderPage title="Add Member" description="Create a new member record" />} />
        <Route path="members/:id" element={<PlaceholderPage title="Member Details" description="View member information" />} />
        <Route path="members/:id/edit" element={<PlaceholderPage title="Edit Member" description="Update member information" />} />
        <Route path="contributions" element={<ContributionsPage />} />
        <Route path="contributions/new" element={<PlaceholderPage title="Record Contribution" description="Record a new contribution" />} />
        <Route path="claims" element={<ClaimsPage />} />
        <Route path="claims/new" element={<PlaceholderPage title="New Claim" description="Submit a new claim" />} />
        <Route path="claims/:id" element={<PlaceholderPage title="Claim Details" description="View claim details" />} />
        <Route path="voting" element={<VotingPage />} />
        <Route path="voting/new" element={<PlaceholderPage title="New Election" description="Create a new election" />} />
        <Route path="voting/:id" element={<PlaceholderPage title="Manage Election" description="Manage election settings" />} />
        <Route path="voting/:id/results" element={<PlaceholderPage title="Election Results" description="View election results" />} />
        <Route path="hospitals" element={<HospitalsPage />} />
        <Route path="hospitals/new" element={<PlaceholderPage title="Add Hospital" description="Add a new hospital" />} />
        <Route path="hospitals/:id" element={<PlaceholderPage title="Hospital Details" description="View hospital information" />} />
        <Route path="sponsors" element={<PlaceholderPage title="Sponsors" description="Manage sponsors" />} />
        <Route path="reports" element={<ReportsPage />} />
        <Route path="bulk" element={<PlaceholderPage title="Bulk Processing" description="Import members, batch statements, annual posting" />} />
        <Route path="bulk/import" element={<PlaceholderPage title="Import Members" description="Bulk import members from CSV" />} />
        <Route path="maker-checker" element={<PlaceholderPage title="Maker-Checker" description="Review pending changes" />} />
        <Route path="tax" element={<PlaceholderPage title="Tax Management" description="Tax computation and exemptions" />} />
        <Route path="sms" element={<PlaceholderPage title="SMS Gateway" description="Send bulk messages" />} />
        <Route path="news" element={<PlaceholderPage title="News" description="Kenya government news" />} />
        <Route path="security" element={<PlaceholderPage title="Security" description="IP blacklisting and access control" />} />
        <Route path="settings" element={<PlaceholderPage title="Settings" description="System settings" />} />
        {/* Member Portal */}
        <Route path="portal" element={<PlaceholderPage title="My Dashboard" description="Member portal dashboard" />} />
        <Route path="portal/profile" element={<PlaceholderPage title="My Profile" description="View and update profile" />} />
        <Route path="portal/contributions" element={<PlaceholderPage title="My Contributions" description="View contribution history" />} />
        <Route path="portal/claims" element={<PlaceholderPage title="My Claims" description="View claim status" />} />
        <Route path="portal/voting" element={<PlaceholderPage title="Vote" description="Cast your vote" />} />
        <Route path="portal/projections" element={<PlaceholderPage title="Benefit Projections" description="Project retirement benefits" />} />
        <Route path="portal/feedback" element={<PlaceholderPage title="Feedback" description="Submit feedback" />} />
        <Route path="portal/settings" element={<PlaceholderPage title="Settings" description="Account settings" />} />
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
