// ─── Enums ───────────────────────────────────────────────────────────────────

export type UserRole = 'USER' | 'ADMIN' | 'LEGAL' | 'EXTERNAL'
export type UserStatus = 'ACTIVE' | 'INACTIVE'

export type SubmissionStatus =
  | 'SUBMITTED'
  | 'UNDER_REVIEW'
  | 'NEED_REVISION'
  | 'REJECTED'
  | 'RESUBMITTED'
  | 'COMPLETED'

export type LegalType =
  | 'Permasalahan Hukum'
  | 'Bisnis Teknik'
  | 'Bisnis Penunjang'
  | 'Perjanjian Reasuransi (Treaty/Fakultatif)'
  | 'Lain-Lain'

export type DocumentType =
  | 'Surat Perintah Kerja'
  | 'Perjanjian Kerjasama Non Teknik'
  | 'Kontrak Treaty'
  | 'Kontrak Retro'
  | 'Pembatalan Perjanjian'
  | 'Nota Kesepahaman'
  | 'Surat'
  | 'Lain-Lain'

export type LegalCaseType =
  | 'NON_LITIGASI'
  | 'PERDATA'
  | 'PIDANA'
  | 'TIPEKOR'
  | 'ARBITRASE'
  | 'TUN'

export type CaseCategory =
  | 'Life'
  | 'BPPDAN'
  | 'Property'
  | 'COB'

// ─── Entities ─────────────────────────────────────────────────────────────────

export interface User {
  id: string
  full_name: string
  email: string
  position: string
  division: string
  division_id?: string
  division_detail?: Division
  role: UserRole
  status: UserStatus
  email_notifications: boolean
  two_fa_enabled: boolean
  created_at: string
  updated_at: string
}

export interface Attachment {
  id: string
  file_name: string
  file_path: string
  file_size: number
  upload_round: number
  created_at: string
}

export interface SubmissionResult {
  id: string
  file_name: string
  file_path: string
  notes: string
  uploaded_by: string
  uploader?: Pick<User, 'id' | 'full_name'>
  created_at: string
}

export interface LegalOpinion {
  id: string
  ticket_number: string
  user_id: string
  user?: Pick<User, 'id' | 'full_name' | 'email'>
  requestor_name: string
  requestor_position: string
  requestor_division: string
  requestor_email: string
  requestor_phone: string
  legal_type: LegalType | string
  legal_type_other?: string
  title: string
  chronology: string
  question: string
  status: SubmissionStatus
  admin_note?: string
  attachments?: Attachment[]
  results?: SubmissionResult[]
  created_at: string
  updated_at: string
}

export interface DocumentReview {
  id: string
  ticket_number: string
  user_id: string
  user?: Pick<User, 'id' | 'full_name' | 'email'>
  requestor_name: string
  requestor_position: string
  requestor_division: string
  requestor_email: string
  requestor_phone: string
  document_name: string
  second_party: string
  third_party?: string
  document_type: DocumentType | string
  document_type_other?: string
  additional_note?: string
  status: SubmissionStatus
  admin_note?: string
  attachments?: Attachment[]
  results?: SubmissionResult[]
  created_at: string
  updated_at: string
}

export interface Regency {
  id: string
  name: string
  province: string
  type: string
  label: string
}

export interface Division {
  id: string
  name: string
  description?: string
  created_at: string
  updated_at: string
}

export interface Cedant {
  id: string
  name: string
  description?: string
  created_at: string
  updated_at: string
}

export interface CaseChronology {
  id: string
  case_id: string
  agenda_date: string
  agenda: string
  description?: string
  documents: string[]
  created_at: string
  updated_at: string
}

export interface LegalCase {
  id: string
  case_name: string
  case_summary?: string
  related_party_id: string
  related_party?: Cedant
  category: CaseCategory
  specification?: string
  case_type: LegalCaseType | string
  technical_reserve?: string
  case_value: number
  pic: string
  pic_division?: Division
  document_link?: string
  current_status?: string
  case_date: string
  level: string
  additional_notes?: string
  location_regency_id: string
  location_regency?: Regency
  chronologies?: CaseChronology[]
  created_at: string
  updated_at: string
}

// ─── API Response wrappers ────────────────────────────────────────────────────

export interface ApiResponse<T> {
  success: boolean
  message: string
  data?: T
  errors?: Record<string, string>
}

export interface PaginatedData<T> {
  items: T[]
  total: number
  page: number
  limit: number
  total_pages: number
}

// ─── Auth ─────────────────────────────────────────────────────────────────────

export interface LoginRequest {
  email: string
  password: string
}

export interface AuthResponse {
  token: string
  access_token?: string
  refresh_token: string
  user: User
}

// ─── Dashboard ────────────────────────────────────────────────────────────────

export interface UserDashboardStats {
  total_legal_opinions: number
  total_document_reviews: number
  pending: number
  need_revision: number
  completed: number
}

export interface AdminDashboardStats {
  total_users: number
  total_legal_opinions: number
  total_document_reviews: number
  pending_review: number
  need_revision: number
  resubmitted: number
}

export type AuditAction =
  | 'STATUS_CHANGE'
  | 'FILE_UPLOAD'
  | 'USER_UPDATE'
  | 'LOGIN'
  | 'LOGOUT'
  | 'DELETE'
  | 'FILE_DELETE'

export interface AuditLog {
  id: string
  user_id: string
  user?: Pick<User, 'id' | 'full_name' | 'email' | 'role'>
  action: AuditAction
  entity_type: string
  entity_id: string
  old_value?: string
  new_value?: string
  description?: string
  ip_address: string
  user_agent: string
  created_at: string
}

export interface AuditLogFilters {
  action?: AuditAction
  entity_type?: string
  user_id?: string
  date_from?: string
  date_to?: string
  search?: string
  page: number
  limit: number
}

export type WarningLevel = 'NONE' | 'YELLOW' | 'RED'

export interface NotificationSetting {
  id: string
  submission_type: 'legal_opinion' | 'document_review' | 'ALL'
  warning_level: WarningLevel
  days_threshold: number
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface ReminderItem {
  id: string
  submission_type: string
  ticket_number: string
  title: string
  status: string
  submitted_at: string
  last_updated_at?: string
  days_since_submission: number
  days_since_last_update: number
  warning_level: WarningLevel
  warning_color: string
  assigned_legal_name?: string
}

export interface RemindersResponse {
  yellow: ReminderItem[]
  red: ReminderItem[]
  none: ReminderItem[]
}
