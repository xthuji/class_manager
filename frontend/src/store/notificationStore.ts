import { create } from 'zustand';
import { Notification } from '@/types';
import { api } from '@/api/client';

interface NotificationState {
  notifications: Notification[];
  unreadCount: number;
  loading: boolean;
  getNotifications: (status?: string) => Promise<void>;
  getUnreadCount: () => Promise<void>;
  markAsRead: (id: number) => Promise<boolean>;
  batchMarkAsRead: (ids: number[]) => Promise<boolean>;
  markAsProcessed: (id: number) => Promise<boolean>;
  batchMarkAsProcessed: (ids: number[]) => Promise<boolean>;
  checkThresholds: () => Promise<void>;
}

export const useNotificationStore = create<NotificationState>((set) => ({
  notifications: [],
  unreadCount: 0,
  loading: false,

  getNotifications: async (status = '') => {
    set({ loading: true });
    try {
      const result = await api.notification.list(status) as Notification[];
      set({ notifications: result, loading: false });
    } finally {
      set({ loading: false });
    }
  },

  getUnreadCount: async () => {
    try {
      const result = await api.notification.getUnreadCount() as number;
      set({ unreadCount: result });
    } catch {
      set({ unreadCount: 0 });
    }
  },

  markAsRead: async (id) => {
    try {
      const result = await api.notification.markAsRead(id) as boolean;
      if (result) {
        set((state) => ({
          notifications: state.notifications.map((n) =>
            n.id === id ? { ...n, status: 'read' as const } : n
          ),
          unreadCount: Math.max(0, state.unreadCount - 1),
        }));
      }
      return result;
    } catch {
      return false;
    }
  },

  batchMarkAsRead: async (ids) => {
    try {
      const result = await api.notification.batchMarkAsRead(ids) as boolean;
      if (result) {
        set((state) => ({
          notifications: state.notifications.map((n) =>
            ids.includes(n.id) ? { ...n, status: 'read' as const } : n
          ),
          unreadCount: Math.max(0, state.unreadCount - ids.length),
        }));
      }
      return result;
    } catch {
      return false;
    }
  },

  markAsProcessed: async (id) => {
    try {
      const result = await api.notification.markAsProcessed(id) as boolean;
      if (result) {
        set((state) => ({
          notifications: state.notifications.map((n) =>
            n.id === id ? { ...n, status: 'processed' as const } : n
          ),
          unreadCount: Math.max(0, state.unreadCount - (state.notifications.find((n) => n.id === id)?.status === 'unread' ? 1 : 0)),
        }));
      }
      return result;
    } catch {
      return false;
    }
  },

  batchMarkAsProcessed: async (ids) => {
    try {
      const result = await api.notification.batchMarkAsProcessed(ids) as boolean;
      if (result) {
        set((state) => {
          const unreadToSubtract = state.notifications.filter(
            (n) => ids.includes(n.id) && n.status === 'unread'
          ).length;
          return {
            notifications: state.notifications.map((n) =>
              ids.includes(n.id) ? { ...n, status: 'processed' as const } : n
            ),
            unreadCount: Math.max(0, state.unreadCount - unreadToSubtract),
          };
        });
      }
      return result;
    } catch {
      return false;
    }
  },

  checkThresholds: async () => {
    try {
      await api.notification.checkThresholds();
      await useNotificationStore.getState().getUnreadCount();
    } catch {
    }
  },
}));
