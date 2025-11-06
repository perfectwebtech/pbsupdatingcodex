import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { Toaster } from 'react-hot-toast';
import { useAuthStore } from './stores/authStore';
import Login from './pages/Login';
import Dashboard from './pages/Dashboard';
import Users from './pages/Users';
import Streams from './pages/Streams';
import Categories from './pages/Categories';
import Analytics from './pages/Analytics';
import Sessions from './pages/Sessions';
import Transcoding from './pages/Transcoding';
import Settings from './pages/Settings';
import Reports from './pages/Reports';
import Resellers from './pages/Resellers';
import Billing from './pages/Billing';
import Packages from './pages/Packages';
import Series from './pages/Series';
import EPG from './pages/EPG';
import Devices from './pages/Devices';
import Layout from './components/Layout';
import { useEffect } from 'react';

// Protected Route Component
const ProtectedRoute = ({ children }: { children: React.ReactNode }) => {
  const { isAuthenticated } = useAuthStore();

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
};

function App() {
  const { isAuthenticated, getCurrentUser } = useAuthStore();

  useEffect(() => {
    if (isAuthenticated) {
      getCurrentUser();
    }
  }, [isAuthenticated]);

  return (
    <>
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<Login />} />

          <Route
            path="/*"
            element={
              <ProtectedRoute>
                <Layout>
                  <Routes>
                    <Route path="/" element={<Dashboard />} />
                    <Route path="/users" element={<Users />} />
                    <Route path="/resellers" element={<Resellers />} />
                    <Route path="/streams" element={<Streams />} />
                    <Route path="/categories" element={<Categories />} />
                    <Route path="/series" element={<Series />} />
                    <Route path="/epg" element={<EPG />} />
                    <Route path="/devices" element={<Devices />} />
                    <Route path="/packages" element={<Packages />} />
                    <Route path="/billing" element={<Billing />} />
                    <Route path="/analytics" element={<Analytics />} />
                    <Route path="/reports" element={<Reports />} />
                    <Route path="/sessions" element={<Sessions />} />
                    <Route path="/transcoding" element={<Transcoding />} />
                    <Route path="/settings" element={<Settings />} />
                    <Route path="*" element={<Navigate to="/" replace />} />
                  </Routes>
                </Layout>
              </ProtectedRoute>
            }
          />
        </Routes>
      </BrowserRouter>

      <Toaster
        position="top-right"
        toastOptions={{
          duration: 4000,
          style: {
            background: '#1e293b',
            color: '#fff',
            borderRadius: '12px',
            padding: '16px',
          },
          success: {
            iconTheme: {
              primary: '#10b981',
              secondary: '#fff',
            },
          },
          error: {
            iconTheme: {
              primary: '#ef4444',
              secondary: '#fff',
            },
          },
        }}
      />
    </>
  );
}

export default App;
