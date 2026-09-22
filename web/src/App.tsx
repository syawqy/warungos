import { useState } from 'react';
import { useAuth } from './hooks/useAuth';
import Sidebar from './components/Layout/Sidebar';
import POSPage from './pages/POSPage';
import OrdersPage from './pages/OrdersPage';
import DashboardPage from './pages/DashboardPage';
import InventoryPage from './pages/InventoryPage';
import LoginPage from './components/Auth/LoginPage';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

type Page = 'pos' | 'orders' | 'dashboard' | 'inventory';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: 1, refetchOnWindowFocus: false },
  },
});

function AppContent() {
  const { user, isAuthenticated, isLoading, login, logout } = useAuth();
  const [currentPage, setCurrentPage] = useState<Page>('pos');

  if (isLoading) {
    return (
      <div className="h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <div className="w-12 h-12 border-4 border-warung-orange border-t-transparent rounded-full animate-spin mx-auto mb-4" />
          <p className="text-gray-500">Memuat...</p>
        </div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <LoginPage onLogin={login} />;
  }

  const renderPage = () => {
    switch (currentPage) {
      case 'pos': return <POSPage />;
      case 'orders': return <OrdersPage />;
      case 'dashboard': return <DashboardPage />;
      case 'inventory': return <InventoryPage />;
    }
  };

  return (
    <div className="h-screen flex overflow-hidden">
      <Sidebar currentPage={currentPage} onNavigate={setCurrentPage} />
      <main className="flex-1 overflow-hidden bg-gray-50">
        {renderPage()}
      </main>
    </div>
  );
}

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AppContent />
    </QueryClientProvider>
  );
}
