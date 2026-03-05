import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { User, Role, Token } from '../types';

interface UserState {
  token: string | null;
  user: User | null;
  role: Role | null;
  setToken: (token: Token) => void;
  setUser: (user: User) => void;
  setRole: (role: Role) => void;
  logout: () => void;
}

export const useUserStore = create<UserState>()(
  persist(
    (set) => ({
      token: null,
      user: null,
      role: null,
      setToken: (token) => set({ token: token.text }),
      setUser: (user) => set({ user }),
      setRole: (role) => set({ role }),
      logout: () => {
        set({ token: null, user: null, role: null });
      },
    }),
    {
      name: 'user-storage',
      partialize: (state) => ({ token: state.token }),
    }
  )
);
