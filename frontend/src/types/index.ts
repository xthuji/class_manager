export interface Student {
  id: number;
  name: string;
  student_id: string;
  contact: string;
  subjects: string[];
  total_hours: number;
  completed_hours: number;
  remaining_hours: number;
  created_at: string;
  updated_at: string;
}

export interface StudentCreateRequest {
  name: string;
  student_id: string;
  contact: string;
  subjects: string[];
  total_hours: number;
}

export interface StudentUpdateRequest {
  id: number;
  name: string;
  contact: string;
  subjects: string[];
  total_hours: number;
}

export interface StudentSearchRequest {
  name: string;
  student_id: string;
  subjects: string[];
}

export interface BatchImportResult {
  success_count: number;
  failed_count: number;
  errors: string[];
}

export interface Course {
  id: number;
  name: string;
  subject: string;
  description: string;
  student_count: number;
  created_at: string;
  updated_at: string;
}

export interface CourseCreateRequest {
  name: string;
  subject: string;
  description: string;
}

export interface CourseUpdateRequest {
  id: number;
  name: string;
  subject: string;
  description: string;
}

export interface EnrollRequest {
  student_id: number;
  course_id: number;
}

export interface CourseHoursRequest {
  course_id: number;
  hours: number;
}

export interface HourRecord {
  id: number;
  student_id: number;
  course_id: number;
  hours: number;
  record_date: string;
  description: string;
  created_at: string;
}

export interface HourRecordCreateRequest {
  student_id: number;
  course_id: number;
  hours: number;
  record_date: string;
  description: string;
}

export interface BatchHourRecordRequest {
  student_ids: number[];
  course_id: number;
  hours: number;
  record_date: string;
  description: string;
}

export interface HoursSummaryResponse {
  total_hours: number;
  completed_hours: number;
  remaining_hours: number;
}

export interface ExportRequest {
  student_id: number;
  start_date: string;
  end_date: string;
}

export interface HourRecordListRequest {
  student_ids: number[];
  course_ids: number[];
  start_date: string;
  end_date: string;
}

export interface HourRecordWithDetails {
  id: number;
  student_id: number;
  course_id: number;
  hours: number;
  record_date: string;
  description: string;
  created_at: string;
  student_name: string;
  student_no: string;
  course_name: string;
  course_subject: string;
}

export interface HourRecharge {
  id: number;
  student_id: number;
  student_name: string;
  student_no: string;
  hours: number;
  recharge_date: string;
  description: string;
  created_at: string;
}

export interface HourRechargeCreateRequest {
  student_id: number;
  hours: number;
  recharge_date: string;
  description: string;
}

export interface HourRechargeListRequest {
  page: number;
  page_size: number;
  student_id?: number;
  start_date?: string;
  end_date?: string;
}

export interface Schedule {
  id: number;
  course_id: number;
  day_of_week: number;
  period: string;
  start_time: string;
  end_time: string;
  hours_consumed: number;
  created_at: string;
  updated_at: string;
}

export interface ScheduleCreateRequest {
  course_id: number;
  day_of_week: number;
  period: string;
  start_time: string;
  end_time: string;
  hours_consumed: number;
}

export interface ScheduleUpdateRequest {
  id: number;
  day_of_week: number;
  period: string;
  start_time: string;
  end_time: string;
  hours_consumed: number;
}

export interface Threshold {
  id: number;
  subject: string;
  value: number;
  template_name: string;
  created_at: string;
  updated_at: string;
}

export interface ThresholdRequest {
  subject: string;
  value: number;
}

export interface ThresholdTemplateRequest {
  template_name: string;
  thresholds: ThresholdRequest[];
}

export interface ThresholdPreviewResponse {
  subject: string;
  threshold_value: number;
  trigger_count: number;
  total_student_count: number;
}

export interface Notification {
  id: number;
  student_id: number;
  student_name: string;
  subject: string;
  current_hours: number;
  threshold: number;
  status: 'unread' | 'read' | 'processed';
  created_at: string;
  processed_at: string;
}

export interface NotificationRequest {
  title: string;
  subtitle: string;
  body: string;
}

export interface BackupFile {
  filename: string;
  path: string;
  size: number;
  created_at: string;
  checksum: string;
}

export interface DashboardStats {
  total_students: number;
  total_courses: number;
  today_hours: number;
  unread_notifications: number;
}

export interface ExportData {
  students: any[];
  courses: any[];
  hour_records: any[];
  hour_recharges: any[];
  schedules: any[];
  threshold: any[];
  notifications: any[];
}