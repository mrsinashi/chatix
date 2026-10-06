import { create } from 'zustand';
import { login as apiLogin, getMe, logout as apiLogout, UserInfo } from '../api/authApi';

interface AuthState {
  token: string | null;
  user: UserInfo | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
  login: (username: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  checkSession: () => Promise<void>;
}

export const useAuthStore = create<AuthState>((set) => ({
  token: localStorage.getItem('chatix_token'),
  user: null,
  isAuthenticated: false,
  isLoading: false,
  error: null,

  login: async (username, password) => {
    set({ isLoading: true, error: null });
    try {
      const response = await apiLogin({ username, password });
      localStorage.setItem('chatix_token', response.token);
      set({
        token: response.token,
        user: response.user,
        isAuthenticated: true,
        isLoading: false,
      });
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : 'Ошибка входа',
        isLoading: false,
      });
      throw error;
    }
  },

  logout: async () => {
    const token = useAuthStore.getState().token;
    if (token) {
      try {
        await apiLogout(token);
      } catch {
        // Игнорируем ошибки при выходе
      }
    }
    localStorage.removeItem('chatix_token');
    set({ token: null, user: null, isAuthenticated: false });
  },

  checkSession: async () => {
    const token = useAuthStore.getState().token;
    if (!token) {
      set({ isAuthenticated: false });
      return;
    }
    try {
      const response = await getMe(token);
      set({ user: response.user, isAuthenticated: true });
    } catch {
      localStorage.removeItem('chatix_token');
      set({ token: null, user: null, isAuthenticated: false });
    }
  },
}));