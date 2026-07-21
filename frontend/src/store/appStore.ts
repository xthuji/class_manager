import { create } from 'zustand';
import { DashboardStats } from '@/types';
import { api } from '@/api/client';

interface AppState {
  theme: 'light' | 'dark';
  isLoading: boolean;
  currentPage: string;
  stats: DashboardStats | null;
  toggleTheme: () => void;
  setCurrentPage: (page: string) => void;
  loadStats: () => Promise<void>;
}

export const useAppStore = create<AppState>((set) => ({
  theme: 'light',
  isLoading: false,
  currentPage: '/',
  stats: null,

  toggleTheme: () => {
    set((state) => {
      const newTheme = state.theme === 'light' ? 'dark' : 'light';
      document.documentElement.classList.toggle('dark', newTheme === 'dark');
      return { theme: newTheme };
    });
  },

  setCurrentPage: (page) => {
    set({ currentPage: page });
  },

  loadStats: async () => {
    try {
      const [students, courses, unreadCount] = await Promise.all([
        api.student.list({ page: 1, page_size: 1 }) as Promise<unknown[]>,
        api.course.list({ page: 1, page_size: 1 }) as Promise<unknown[]>,
        api.notification.getUnreadCount() as Promise<number>,
      ]);
      set({
        stats: {
          total_students: students.length,
          total_courses: courses.length,
          today_hours: 0,
          unread_notifications: unreadCount,
        },
      });
    } catch {
      set({
        stats: {
          total_students: 0,
          total_courses: 0,
          today_hours: 0,
          unread_notifications: 0,
        },
      });
    }
  },
}));