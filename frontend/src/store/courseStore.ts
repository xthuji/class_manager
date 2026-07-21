import { create } from 'zustand';
import { Course, CourseCreateRequest, CourseUpdateRequest, EnrollRequest, CourseHoursRequest, Student } from '@/types';
import { api } from '@/api/client';

interface CourseState {
  courses: Course[];
  currentCourse: Course | null;
  courseStudents: Student[];
  loading: boolean;
  createCourse: (req: CourseCreateRequest) => Promise<Course>;
  updateCourse: (req: CourseUpdateRequest) => Promise<Course>;
  deleteCourse: (id: number) => Promise<boolean>;
  getCourseByID: (id: number) => Promise<Course>;
  listCourses: () => Promise<void>;
  enrollStudent: (req: EnrollRequest) => Promise<boolean>;
  getCourseStudents: (courseID: number) => Promise<void>;
  addCourseHours: (req: CourseHoursRequest) => Promise<Course>;
}

export const useCourseStore = create<CourseState>((set) => ({
  courses: [],
  currentCourse: null,
  courseStudents: [],
  loading: false,

  createCourse: async (req) => {
    set({ loading: true });
    try {
      const result = await api.course.create(req) as Course;
      set((state) => ({ courses: [...state.courses, result], loading: false }));
      return result;
    } finally {
      set({ loading: false });
    }
  },

  updateCourse: async (req) => {
    set({ loading: true });
    try {
      const result = await api.course.update(req) as Course;
      set((state) => ({
        courses: state.courses.map((c) => (c.id === req.id ? result : c)),
        currentCourse: state.currentCourse?.id === req.id ? result : state.currentCourse,
        loading: false,
      }));
      return result;
    } finally {
      set({ loading: false });
    }
  },

  deleteCourse: async (id) => {
    set({ loading: true });
    try {
      const result = await api.course.delete(id) as boolean;
      if (result) {
        set((state) => ({
          courses: state.courses.filter((c) => c.id !== id),
          currentCourse: state.currentCourse?.id === id ? null : state.currentCourse,
        }));
      }
      return result;
    } finally {
      set({ loading: false });
    }
  },

  getCourseByID: async (id) => {
    set({ loading: true });
    try {
      const result = await api.course.getByID(id) as Course;
      set({ currentCourse: result, loading: false });
      return result;
    } finally {
      set({ loading: false });
    }
  },

  listCourses: async () => {
    set({ loading: true });
    try {
      const result = await api.course.list({ page: 1, page_size: 100 }) as Course[];
      set({ courses: result, loading: false });
    } finally {
      set({ loading: false });
    }
  },

  enrollStudent: async (req) => {
    set({ loading: true });
    try {
      const result = await api.course.enroll(req) as boolean;
      const state = useCourseStore.getState();
      if (result && state.currentCourse !== null) {
        await state.getCourseStudents(state.currentCourse.id);
      }
      return result;
    } finally {
      set({ loading: false });
    }
  },

  getCourseStudents: async (courseID) => {
    set({ loading: true });
    try {
      const result = await api.course.getStudents(courseID) as Student[];
      set({ courseStudents: result, loading: false });
    } finally {
      set({ loading: false });
    }
  },

  addCourseHours: async (req) => {
    set({ loading: true });
    try {
      const result = await api.course.addHours(req) as Course;
      set((state) => ({
        courses: state.courses.map((c) => (c.id === req.course_id ? result : c)),
        currentCourse: state.currentCourse?.id === req.course_id ? result : state.currentCourse,
        loading: false,
      }));
      return result;
    } finally {
      set({ loading: false });
    }
  },
}));
