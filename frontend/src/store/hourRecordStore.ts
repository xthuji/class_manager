import { create } from 'zustand';
import { HourRecord, HourRecordCreateRequest, BatchHourRecordRequest, HoursSummaryResponse, ExportRequest, HourRecordWithDetails, HourRecordListRequest } from '@/types';
import { api } from '@/api/client';

interface HourRecordState {
  records: HourRecordWithDetails[];
  summary: HoursSummaryResponse | null;
  loading: boolean;
  createHourRecord: (req: HourRecordCreateRequest) => Promise<HourRecord>;
  batchCreateHourRecord: (req: BatchHourRecordRequest) => Promise<HourRecord[]>;
  getHourRecordsByStudent: (studentID: number) => Promise<void>;
  listHourRecords: (req: HourRecordListRequest) => Promise<void>;
  getDistinctDates: () => Promise<string[]>;
  exportHourRecords: (req: ExportRequest) => Promise<string>;
  getStudentHoursSummary: (studentID: number) => Promise<HoursSummaryResponse>;
}

export const useHourRecordStore = create<HourRecordState>((set) => ({
  records: [],
  summary: null,
  loading: false,

  createHourRecord: async (req) => {
    set({ loading: true });
    try {
      const result = await api.hourRecord.create(req) as HourRecord;
      return result;
    } finally {
      set({ loading: false });
    }
  },

  batchCreateHourRecord: async (req) => {
    set({ loading: true });
    try {
      const result = await api.hourRecord.batchCreate(req) as HourRecord[];
      return result;
    } finally {
      set({ loading: false });
    }
  },

  getHourRecordsByStudent: async (studentID) => {
    set({ loading: true });
    try {
      const result = await api.hourRecord.getByStudent(studentID) as HourRecord[];
      set({ records: result as unknown as HourRecordWithDetails[], loading: false });
    } finally {
      set({ loading: false });
    }
  },

  listHourRecords: async (req) => {
    set({ loading: true });
    try {
      const result = await api.hourRecord.list(req) as HourRecordWithDetails[];
      set({ records: result, loading: false });
    } finally {
      set({ loading: false });
    }
  },

  getDistinctDates: async () => {
    const result = await api.hourRecord.getDistinctDates();
    return result as string[];
  },

  exportHourRecords: async (req) => {
    set({ loading: true });
    try {
      const result = await api.hourRecord.export(req) as string;
      return result;
    } finally {
      set({ loading: false });
    }
  },

  getStudentHoursSummary: async (studentID) => {
    set({ loading: true });
    try {
      const result = await api.hourRecord.getSummary(studentID) as HoursSummaryResponse;
      set({ summary: result, loading: false });
      return result;
    } finally {
      set({ loading: false });
    }
  },
}));