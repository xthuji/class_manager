import { create } from 'zustand';
import { Student, StudentCreateRequest, StudentUpdateRequest, StudentSearchRequest, BatchImportResult } from '@/types';
import { api } from '@/api/client';

interface StudentState {
  students: Student[];
  currentStudent: Student | null;
  loading: boolean;
  searchFilter: StudentSearchRequest;
  createStudent: (req: StudentCreateRequest) => Promise<Student>;
  updateStudent: (req: StudentUpdateRequest) => Promise<Student>;
  deleteStudent: (id: number) => Promise<boolean>;
  getStudentByID: (id: number) => Promise<Student>;
  listStudents: () => Promise<void>;
  searchStudents: (filter: StudentSearchRequest) => Promise<void>;
  batchImportStudents: (csvContent: string) => Promise<BatchImportResult>;
  exportStudents: (ids: number[]) => Promise<string>;
}

export const useStudentStore = create<StudentState>((set) => ({
  students: [],
  currentStudent: null,
  loading: false,
  searchFilter: { name: '', student_id: '', subjects: [] },

  createStudent: async (req) => {
    set({ loading: true });
    try {
      const result = await api.student.create(req) as Student;
      set((state) => ({ students: [...state.students, result], loading: false }));
      return result;
    } finally {
      set({ loading: false });
    }
  },

  updateStudent: async (req) => {
    set({ loading: true });
    try {
      const result = await api.student.update(req) as Student;
      set((state) => ({
        students: state.students.map((s) => (s.id === req.id ? result : s)),
        currentStudent: state.currentStudent?.id === req.id ? result : state.currentStudent,
        loading: false,
      }));
      return result;
    } finally {
      set({ loading: false });
    }
  },

  deleteStudent: async (id) => {
    set({ loading: true });
    try {
      const result = await api.student.delete(id) as boolean;
      if (result) {
        set((state) => ({
          students: state.students.filter((s) => s.id !== id),
          currentStudent: state.currentStudent?.id === id ? null : state.currentStudent,
        }));
      }
      return result;
    } finally {
      set({ loading: false });
    }
  },

  getStudentByID: async (id) => {
    set({ loading: true });
    try {
      const result = await api.student.getByID(id) as Student;
      set({ currentStudent: result, loading: false });
      return result;
    } finally {
      set({ loading: false });
    }
  },

  listStudents: async () => {
    set({ loading: true });
    try {
      const result = await api.student.list({ page: 1, page_size: 100 }) as Student[];
      set({ students: result, loading: false });
    } finally {
      set({ loading: false });
    }
  },

  searchStudents: async (filter) => {
    set({ loading: true, searchFilter: filter });
    try {
      const result = await api.student.search(filter) as Student[];
      set({ students: result, loading: false });
    } finally {
      set({ loading: false });
    }
  },

  batchImportStudents: async (csvContent) => {
    set({ loading: true });
    try {
      const result = await api.student.batchImport(csvContent) as BatchImportResult;
      await useStudentStore.getState().listStudents();
      return result;
    } finally {
      set({ loading: false });
    }
  },

  exportStudents: async (ids) => {
    set({ loading: true });
    try {
      const result = await api.student.export(ids) as string;
      return result;
    } finally {
      set({ loading: false });
    }
  },
}));
