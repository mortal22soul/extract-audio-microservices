import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { User, LoginRequest, RegisterRequest } from '@/types';
import apiClient from '@/lib/api';
import toast from 'react-hot-toast';

interface AuthState {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;

  // Actions
  login: (credentials: LoginRequest) => Promise<void>;
  register: (userData: RegisterRequest) => Promise<void>;
  logout: () => Promise<void>;
  checkAuth: () => Promise<void>;
  setUser: (user: User | null) => void;
  setLoading: (loading: boolean) => void;
  getAuthToken: () => string | null;
}

export const useAuthStore = create<AuthState>()(
  persist(
    set => ({
      user: null,
      isLoading: false,
      isAuthenticated: false,

      login: async (credentials: LoginRequest) => {
        try {
          set({ isLoading: true });

          const response = await apiClient.login(credentials);

          // Store token in cookie for middleware
          if (typeof document !== 'undefined') {
            document.cookie = `auth_token=${response.token}; path=/; max-age=${7 * 24 * 60 * 60}; secure; samesite=strict`;
          }

          set({
            user: response.user,
            isAuthenticated: true,
            isLoading: false,
          });

          toast.success('Login successful!');
        } catch (error: any) {
          set({ isLoading: false });
          const message = error.response?.data?.error || 'Login failed';
          toast.error(message);
          throw error;
        }
      },

      register: async (userData: RegisterRequest) => {
        try {
          set({ isLoading: true });

          const response = await apiClient.register(userData);

          set({
            user: response.user,
            isAuthenticated: true,
            isLoading: false,
          });

          toast.success('Registration successful!');
        } catch (error: any) {
          set({ isLoading: false });
          const message = error.response?.data?.error || 'Registration failed';
          toast.error(message);
          throw error;
        }
      },

      logout: async () => {
        try {
          await apiClient.logout();

          // Clear token cookie
          if (typeof document !== 'undefined') {
            document.cookie =
              'auth_token=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT';
          }

          set({
            user: null,
            isAuthenticated: false,
            isLoading: false,
          });

          toast.success('Logged out successfully');
        } catch (error: any) {
          // Even if logout fails on server, clear local state
          set({
            user: null,
            isAuthenticated: false,
            isLoading: false,
          });

          console.error('Logout error:', error);
        }
      },

      checkAuth: async () => {
        try {
          if (!apiClient.isAuthenticated()) {
            set({ user: null, isAuthenticated: false });
            return;
          }

          set({ isLoading: true });

          const user = await apiClient.getCurrentUser();

          set({
            user,
            isAuthenticated: true,
            isLoading: false,
          });
        } catch (error: any) {
          set({
            user: null,
            isAuthenticated: false,
            isLoading: false,
          });

          console.error('Auth check failed:', error);
        }
      },

      setUser: (user: User | null) => {
        set({
          user,
          isAuthenticated: !!user,
        });
      },

      setLoading: (loading: boolean) => {
        set({ isLoading: loading });
      },

      getAuthToken: () => {
        return apiClient.getAuthToken();
      },
    }),
    {
      name: 'auth-storage',
      partialize: state => ({
        user: state.user,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);
