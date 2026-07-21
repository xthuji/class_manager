import { create } from 'zustand';
import { api } from '@/api/client';
import { HourRecharge, HourRechargeCreateRequest, HourRechargeListRequest } from '@/types';

interface HourRechargeStore {
  recharges: HourRecharge[];
  loading: boolean;
  createRecharge: (req: HourRechargeCreateRequest) => Promise<HourRecharge>;
  getRechargesByStudent: (studentID: number) => Promise<void>;
  listRecharges: (req?: HourRechargeListRequest) => Promise<void>;
  deleteRecharge: (id: number) => Promise<void>;
}

export const useHourRechargeStore = create<HourRechargeStore>((set) => ({
  recharges: [],
  loading: false,

  createRecharge: async (req: HourRechargeCreateRequest) => {
    set({ loading: true });
    try {
      const result = await api.hourRecharge.create(req);
      return result as HourRecharge;
    } finally {
      set({ loading: false });
    }
  },

  getRechargesByStudent: async (studentID: number) => {
    set({ loading: true });
    try {
      const result = await api.hourRecharge.getByStudent(studentID);
      set({ recharges: result as HourRecharge[] });
    } finally {
      set({ loading: false });
    }
  },

  listRecharges: async (req?: HourRechargeListRequest) => {
    set({ loading: true });
    try {
      const result = await api.hourRecharge.list(req);
      set({ recharges: result as HourRecharge[] });
    } finally {
      set({ loading: false });
    }
  },

  deleteRecharge: async (id: number) => {
    set({ loading: true });
    try {
      await api.hourRecharge.delete(id);
    } finally {
      set({ loading: false });
    }
  },
}));