import { useEffect } from 'react';
import { useAuthStore } from './stores/authStore';
import LoginPage from './pages/LoginPage';

export default function App() {
  const { isAuthenticated, checkSession, user, logout } = useAuthStore();

  useEffect(() => {
    checkSession();
  }, [checkSession]);

  if (!isAuthenticated) {
    return <LoginPage />;
  }

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center">
      <div className="text-center">
        <h1 className="text-2xl font-semibold text-gray-900">Chatix</h1>
        <p className="mt-2 text-gray-500">
          Добро пожаловать, {user?.display_name}!
        </p>
        <button
          onClick={logout}
          className="mt-4 text-sm text-blue-600 hover:text-blue-700"
        >
          Выйти
        </button>
      </div>
    </div>
  );
}