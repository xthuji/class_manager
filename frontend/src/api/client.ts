declare global {
  interface Window {
    go: {
      GenerateStudentID: () => Promise<unknown>;
      CreateStudent: (req: unknown) => Promise<unknown>;
      UpdateStudent: (req: unknown) => Promise<unknown>;
      DeleteStudent: (id: number) => Promise<unknown>;
      GetStudentByID: (id: number) => Promise<unknown>;
      ListStudents: (req: unknown) => Promise<unknown>;
      SearchStudents: (req: unknown) => Promise<unknown>;
      BatchImportStudents: (csvContent: string) => Promise<unknown>;
      ExportStudents: (ids: number[]) => Promise<unknown>;
      CreateCourse: (req: unknown) => Promise<unknown>;
      UpdateCourse: (req: unknown) => Promise<unknown>;
      DeleteCourse: (id: number) => Promise<unknown>;
      GetCourseByID: (id: number) => Promise<unknown>;
      ListCourses: (req: unknown) => Promise<unknown>;
      EnrollStudent: (req: unknown) => Promise<unknown>;
      GetCourseStudents: (courseID: number) => Promise<unknown>;
      AddCourseHours: (req: unknown) => Promise<unknown>;
      CreateHourRecord: (req: unknown) => Promise<unknown>;
      BatchCreateHourRecord: (req: unknown) => Promise<unknown>;
      GetHourRecordsByStudent: (studentID: number) => Promise<unknown>;
      ListHourRecords: (req: unknown) => Promise<unknown>;
      GetDistinctDates: () => Promise<unknown>;
      ExportHourRecords: (req: unknown) => Promise<unknown>;
      GetStudentHoursSummary: (studentID: number) => Promise<unknown>;
      CreateRecharge: (req: unknown) => Promise<unknown>;
      GetRechargesByStudent: (studentID: number) => Promise<unknown>;
      ListRecharges: (req?: unknown) => Promise<unknown>;
      DeleteRecharge: (id: number) => Promise<unknown>;
      CreateSchedule: (req: unknown) => Promise<unknown>;
      UpdateSchedule: (req: unknown) => Promise<unknown>;
      DeleteSchedule: (id: number) => Promise<unknown>;
      GetCourseSchedule: (courseID: number) => Promise<unknown>;
      SetThreshold: (req: unknown) => Promise<unknown>;
      GetThreshold: (subject: string) => Promise<unknown>;
      ListThresholds: () => Promise<unknown>;
      SaveThresholdTemplate: (req: unknown) => Promise<unknown>;
      ApplyThresholdTemplate: (templateName: string) => Promise<unknown>;
      PreviewThreshold: (subject: string, value: number) => Promise<unknown>;
      GetNotifications: (status: string) => Promise<unknown>;
      GetUnreadCount: () => Promise<unknown>;
      MarkAsRead: (id: number) => Promise<unknown>;
      BatchMarkAsRead: (ids: number[]) => Promise<unknown>;
      MarkAsProcessed: (id: number) => Promise<unknown>;
      BatchMarkAsProcessed: (ids: number[]) => Promise<unknown>;
      SendSystemNotification: (req: unknown) => Promise<unknown>;
      CheckThresholds: () => Promise<unknown>;
      ExportAllData: () => Promise<unknown>;
      ImportAllData: (jsonData: string) => Promise<unknown>;
      GetAppInfo: () => Promise<unknown>;
    };
  }
}

export const api = {
  student: {
    generateID: () => window.go.GenerateStudentID(),
    create: (req: unknown) => window.go.CreateStudent(req),
    update: (req: unknown) => window.go.UpdateStudent(req),
    delete: (id: number) => window.go.DeleteStudent(id),
    getByID: (id: number) => window.go.GetStudentByID(id),
    list: (req: unknown) => window.go.ListStudents(req),
    search: (req: unknown) => window.go.SearchStudents(req),
    batchImport: (csvContent: string) => window.go.BatchImportStudents(csvContent),
    export: (ids: number[]) => window.go.ExportStudents(ids),
  },
  course: {
    create: (req: unknown) => window.go.CreateCourse(req),
    update: (req: unknown) => window.go.UpdateCourse(req),
    delete: (id: number) => window.go.DeleteCourse(id),
    getByID: (id: number) => window.go.GetCourseByID(id),
    list: (req: unknown) => window.go.ListCourses(req),
    enroll: (req: unknown) => window.go.EnrollStudent(req),
    getStudents: (courseID: number) => window.go.GetCourseStudents(courseID),
    addHours: (req: unknown) => window.go.AddCourseHours(req),
  },
  hourRecord: {
    create: (req: unknown) => window.go.CreateHourRecord(req),
    batchCreate: (req: unknown) => window.go.BatchCreateHourRecord(req),
    getByStudent: (studentID: number) => window.go.GetHourRecordsByStudent(studentID),
    list: (req: unknown) => window.go.ListHourRecords(req),
    getDistinctDates: () => window.go.GetDistinctDates(),
    export: (req: unknown) => window.go.ExportHourRecords(req),
    getSummary: (studentID: number) => window.go.GetStudentHoursSummary(studentID),
  },
  hourRecharge: {
    create: (req: unknown) => window.go.CreateRecharge(req),
    getByStudent: (studentID: number) => window.go.GetRechargesByStudent(studentID),
    list: (req?: unknown) => req ? window.go.ListRecharges(req) : window.go.ListRecharges(),
    delete: (id: number) => window.go.DeleteRecharge(id),
  },
  schedule: {
    create: (req: unknown) => window.go.CreateSchedule(req),
    update: (req: unknown) => window.go.UpdateSchedule(req),
    delete: (id: number) => window.go.DeleteSchedule(id),
    getByCourse: (courseID: number) => window.go.GetCourseSchedule(courseID),
  },
  threshold: {
    set: (req: unknown) => window.go.SetThreshold(req),
    get: (subject: string) => window.go.GetThreshold(subject),
    list: () => window.go.ListThresholds(),
    saveTemplate: (req: unknown) => window.go.SaveThresholdTemplate(req),
    applyTemplate: (templateName: string) => window.go.ApplyThresholdTemplate(templateName),
    preview: (subject: string, value: number) => window.go.PreviewThreshold(subject, value),
  },
  notification: {
    list: (status: string) => window.go.GetNotifications(status),
    getUnreadCount: () => window.go.GetUnreadCount(),
    markAsRead: (id: number) => window.go.MarkAsRead(id),
    batchMarkAsRead: (ids: number[]) => window.go.BatchMarkAsRead(ids),
    markAsProcessed: (id: number) => window.go.MarkAsProcessed(id),
    batchMarkAsProcessed: (ids: number[]) => window.go.BatchMarkAsProcessed(ids),
    sendSystem: (req: unknown) => window.go.SendSystemNotification(req),
    checkThresholds: () => window.go.CheckThresholds(),
  },
  data: {
    exportAll: () => window.go.ExportAllData(),
    importAll: (jsonData: string) => window.go.ImportAllData(jsonData),
  },
};