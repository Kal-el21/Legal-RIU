// ─── Enums ───────────────────────────────────────────────────────────────────

export type UserRole = 'USER' | 'ADMIN' | 'LEGAL' | 'EXTERNAL' | 'LEGAL_AU'
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

export type DocumentTypeValue =
  | 'Surat Perintah Kerja'
  | 'Perjanjian Kerjasama Non Teknik'
  | 'Kontrak Treaty'
  | 'Kontrak Retro'
  | 'Pembatalan Perjanjian'
  | 'Nota Kesepahaman'
  | 'Surat'
  | 'Lain-Lain'

// ─── Entities ─────────────────────────────────────────────────────────────────

export interface Company {
  id: string
  name: string
  email_domain: string
  is_internal: boolean
  created_at: string
  updated_at: string
}

export interface PurposeType {
  id: string
  name: string
  description?: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface CaseType {
  id: string
  code: string
  label: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface CaseCategory {
  id: string
  code: string
  label: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface DocumentType {
  id: string
  name: string
  label: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface LegalMaterial {
  id: string
  title: string
  excerpt?: string
  content: string
  created_by: string
  updated_by: string
  created_at: string
  updated_at: string
}

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
  company_id: string
  company_detail?: Company
  purpose_type_id?: string
  purpose_type_detail?: PurposeType
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
  status_updated_at?: string
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
  document_type: DocumentTypeValue | string
  document_type_other?: string
  additional_note?: string
  status: SubmissionStatus
  admin_note?: string
  attachments?: Attachment[]
  results?: SubmissionResult[]
  created_at: string
  updated_at: string
  status_updated_at?: string
}

export interface CompanyMaster {
  id: string
  name: string
  address?: string
  npwp?: string
  phone?: string
  email?: string
  default_pejabat?: string
  default_jabatan?: string
  default_tempat_ttd?: string
  is_active: boolean
}

export interface CompanyMasterTemplate {
  version: string
  template_path: string
  base_pdf_path: string
  uploaded_at: string
}

export interface TemplateFieldPosition {
  id?: string
  template_version: string
  field_name: string
  x: number
  y: number
  font?: string
  style?: string
  size?: number
  align?: string
  page_number?: number
}

export interface AgreementAttachment {
  id: string
  agreement_id: string
  file_name: string
  file_path: string
  file_size: number
  upload_round: number
  created_at: string
}

export interface AgreementDocument {
  id: string
  ticket_number: string
  user_id: string
  user?: Pick<User, 'id' | 'full_name' | 'email'>
  pihak_pertama_id: string
  pihak_pertama?: CompanyMaster
  pihak_pertama_pejabat?: string
  pihak_pertama_jabatan?: string
  form_data: Record<string, string | number>
  generated_pdf_path?: string
  generated_file_name?: string
  status: SubmissionStatus
  admin_note?: string
  attachments?: AgreementAttachment[]
  created_at: string
  updated_at: string
  status_updated_at?: string
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

export interface ImportRowError {
  row: number
  field: string
  reason: string
}

export interface ImportResult {
  imported: number
  skipped: number
  errors: ImportRowError[]
}



export interface LegalCase {
  id: string
  case_name: string
  case_summary?: string
  ticket_number?: string
  related_party_id: string
  related_party?: Cedant
  category_id: string
  category?: CaseCategory
  specification?: string
  case_type_id: string
  case_type?: CaseType
  technical_reserve?: number
  case_value: number
  pic: string
  pic_division?: Division
  document_link?: string
  photo?: string
  current_status?: string
  case_date: string
  level: string
  additional_notes?: string
  location_regency_id: string
  location_regency?: Regency
  company_id: string
  company?: Company
  chronologies?: CaseChronology[]
  created_at: string
  updated_at: string
  status_updated_at?: string
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
  permissions?: string[]
}

export type PermissionEffect = 'ALLOW' | 'DENY'

export interface Permission {
  id: string
  code: string
  feature: string
  action: string
  scope: string
  label: string
  description?: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface UserPermissionOverride {
  code: string
  effect: PermissionEffect
  updated_at: string
}

export interface UserPermissionAccess {
  user_id: string
  role: UserRole
  permissions: Permission[]
  role_permissions: string[]
  overrides: UserPermissionOverride[]
  effective_permissions: string[]
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
  | 'PERMISSION_UPDATE'

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
  last_updated_at?: string | null
  days_since_submission: number
  days_since_last_update: number
  warning_level: WarningLevel
  warning_color: string
  is_read: boolean
  assigned_legal_name?: string
}

export interface RemindersResponse {
  yellow: ReminderItem[]
  red: ReminderItem[]
  none: ReminderItem[]
  items: ReminderItem[]
  total: number
  unread_total: number
  page: number
  limit: number
  total_pages: number
}

export interface ReportChartSeries {
  name: string
  data: number[]
}

export interface ReportChartResponse {
  labels: string[]
  series: ReportChartSeries[]
}

export type ReportFeature = 'legal-cases' | 'legal-opinions' | 'document-reviews'

export type ReportGroupBy =
  | 'company'
  | 'case_type'
  | 'category'
  | 'status'
  | 'level'
  | 'location'
  | 'pic'
  | 'legal_type'
  | 'division'
  | 'document_type'

export const REPORT_GROUP_BY_OPTIONS: Record<ReportFeature, { value: ReportGroupBy; label: string }[]> = {
  'legal-cases': [
    { value: 'company', label: 'Perusahaan' },
    { value: 'case_type', label: 'Jenis Kasus' },
    { value: 'category', label: 'Kategori' },
    { value: 'status', label: 'Status' },
    { value: 'level', label: 'Level' },
    { value: 'location', label: 'Lokasi' },
    { value: 'pic', label: 'PIC (Divisi)' },
  ],
  'legal-opinions': [
    { value: 'company', label: 'Perusahaan' },
    { value: 'legal_type', label: 'Jenis Hukum' },
    { value: 'status', label: 'Status' },
    { value: 'division', label: 'Divisi' },
  ],
  'document-reviews': [
    { value: 'company', label: 'Perusahaan' },
    { value: 'document_type', label: 'Jenis Dokumen' },
    { value: 'status', label: 'Status' },
    { value: 'division', label: 'Divisi' },
  ],
}
