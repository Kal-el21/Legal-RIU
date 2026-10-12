package service

import (
	"encoding/json"
	"errors"
	"slices"
	"sort"
	"time"

	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/repository"

	"github.com/google/uuid"
)

type NotificationSettingService interface {
	GetAll() ([]dto.NotificationSettingResponse, error)
	GetByID(id string) (*dto.NotificationSettingResponse, error)
	Update(id string, req dto.UpdateNotificationSettingRequest) (*dto.NotificationSettingResponse, error)
	GetActiveSettings() (map[string]map[string]*entity.NotificationSetting, error)
	GetThresholdFor(submissionType, warningLevel string) (int, error)
	GetReminders(userID string, access ReminderAccess, filter ReminderFilter, page int, limit int) (*dto.RemindersResponse, error)
	MarkReminderRead(userID string, submissionType string, submissionID string) error
	MarkAllRemindersRead(userID string, access ReminderAccess) error
}

type notificationSettingService struct {
	repo          repository.NotificationSettingRepository
	dashboardRepo repository.DashboardRepository
}

func NewNotificationSettingService(repo repository.NotificationSettingRepository, dashboardRepo repository.DashboardRepository) NotificationSettingService {
	return &notificationSettingService{
		repo:          repo,
		dashboardRepo: dashboardRepo,
	}
}

func (s *notificationSettingService) GetAll() ([]dto.NotificationSettingResponse, error) {
	settings, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	resp := make([]dto.NotificationSettingResponse, 0, len(settings))
	for _, setting := range settings {
		resp = append(resp, dto.NotificationSettingResponse{
			ID:             setting.ID.String(),
			SubmissionType: setting.SubmissionType,
			WarningLevel:   setting.WarningLevel,
			DaysThreshold:  setting.DaysThreshold,
			IsActive:       setting.IsActive,
			CreatedAt:      setting.CreatedAt.Format(time.RFC3339),
			UpdatedAt:      setting.UpdatedAt.Format(time.RFC3339),
		})
	}
	return resp, nil
}

func (s *notificationSettingService) GetByID(id string) (*dto.NotificationSettingResponse, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("ID tidak valid")
	}

	setting, err := s.repo.GetByID(uid)
	if err != nil {
		return nil, errors.New("Setting tidak ditemukan")
	}

	return &dto.NotificationSettingResponse{
		ID:             setting.ID.String(),
		SubmissionType: setting.SubmissionType,
		WarningLevel:   setting.WarningLevel,
		DaysThreshold:  setting.DaysThreshold,
		IsActive:       setting.IsActive,
		CreatedAt:      setting.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      setting.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *notificationSettingService) Update(id string, req dto.UpdateNotificationSettingRequest) (*dto.NotificationSettingResponse, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("ID tidak valid")
	}

	setting, err := s.repo.GetByID(uid)
	if err != nil {
		return nil, errors.New("Setting tidak ditemukan")
	}

	setting.DaysThreshold = req.DaysThreshold
	if req.IsActive != nil {
		setting.IsActive = *req.IsActive
	}

	if err := s.repo.Update(setting); err != nil {
		return nil, errors.New("Gagal memperbarui setting")
	}

	return &dto.NotificationSettingResponse{
		ID:             setting.ID.String(),
		SubmissionType: setting.SubmissionType,
		WarningLevel:   setting.WarningLevel,
		DaysThreshold:  setting.DaysThreshold,
		IsActive:       setting.IsActive,
		CreatedAt:      setting.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      setting.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *notificationSettingService) GetActiveSettings() (map[string]map[string]*entity.NotificationSetting, error) {
	settings, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	result := make(map[string]map[string]*entity.NotificationSetting)
	for i := range settings {
		setting := &settings[i]
		if !setting.IsActive {
			continue
		}
		if _, ok := result[setting.SubmissionType]; !ok {
			result[setting.SubmissionType] = make(map[string]*entity.NotificationSetting)
		}
		result[setting.SubmissionType][setting.WarningLevel] = setting
	}
	return result, nil
}

func (s *notificationSettingService) GetThresholdFor(submissionType, warningLevel string) (int, error) {
	settings, err := s.repo.GetByTypeAndLevel(submissionType, warningLevel)
	if err != nil {
		return 0, err
	}
	if len(settings) == 0 {
		return 0, errors.New("setting tidak ditemukan")
	}
	if !settings[0].IsActive {
		return 0, errors.New("setting tidak aktif")
	}
	return settings[0].DaysThreshold, nil
}

// Reminders
type WarningLevel string

const (
	WarningLevelNone   WarningLevel = "NONE"
	WarningLevelYellow WarningLevel = "YELLOW"
	WarningLevelRed    WarningLevel = "RED"
)

// ReminderScope adalah hak seorang user atas satu modul pengajuan, dihitung
// handler dari permission efektifnya.
type ReminderScope struct {
	Owner     bool // boleh melihat pengajuan: diberi kabar saat status pengajuannya berubah
	Processor bool // boleh mengubah status: diberi kabar pengajuan masuk dan peringatan
}

// ReminderAccess memetakan submission_type ke hak user di modul itu.
type ReminderAccess map[string]ReminderScope

// ReminderTypes adalah modul pengajuan yang masuk lonceng. Namanya dipakai
// sebagai submission_type dan sama dengan nama fitur di kode permission.
var ReminderTypes = []string{"legal_opinion", "document_review", "agreement_document"}

// ReminderFilter mempersempit daftar lonceng sebelum dipotong per halaman.
type ReminderFilter struct {
	WarningOnly bool // hanya yang sudah melewati ambang kuning atau merah
	UnreadOnly  bool // hanya yang belum dibaca
}

// Kabar selesai atau ditolak yang lebih tua dari ini tidak lagi ditampilkan ke
// pemiliknya: sudah bukan kabar, dan tanpa batas ini daftarnya tumbuh terus.
const finishedNewsMaxAge = 30 * 24 * time.Hour

// Status hasil tindakan pemroses, jadi kabar bagi pemilik pengajuan.
var ownerStatuses = []entity.SubmissionStatus{
	entity.StatusUnderReview, entity.StatusNeedRevision, entity.StatusRejected, entity.StatusCompleted,
}

// Status yang masih menunggu pemroses.
var processorStatuses = []entity.SubmissionStatus{
	entity.StatusSubmitted, entity.StatusResubmitted, entity.StatusUnderReview,
}

// reminderSource adalah bagian pengajuan yang dibutuhkan lonceng, sama untuk
// semua modul.
type reminderSource struct {
	submissionType  string
	id              uuid.UUID
	ownerID         uuid.UUID
	requesterName   string
	ticketNumber    string
	title           string
	status          entity.SubmissionStatus
	note            string
	createdAt       time.Time
	statusUpdatedAt *time.Time
}

// reminderRow adalah satu baris lonceng berikut waktu terakhir ia berubah
// (masuk status sekarang atau naik level), yang dipakai untuk mengurutkan.
type reminderRow struct {
	item      dto.ReminderItem
	changedAt time.Time
}

func (s *notificationSettingService) GetReminders(userID string, access ReminderAccess, filter ReminderFilter, page int, limit int) (*dto.RemindersResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	rows, err := s.loadReminders(uid, access)
	if err != nil {
		return nil, err
	}
	if err := s.applyReminderReadState(rows, uid); err != nil {
		return nil, err
	}
	rows = slices.DeleteFunc(rows, func(row reminderRow) bool {
		return (filter.WarningOnly && row.item.WarningLevel == string(WarningLevelNone)) ||
			(filter.UnreadOnly && row.item.IsRead)
	})
	return paginateReminders(rows, page, limit), nil
}

func (s *notificationSettingService) MarkReminderRead(userID string, submissionType string, submissionID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("user tidak valid")
	}
	submissionUUID, err := uuid.Parse(submissionID)
	if err != nil {
		return errors.New("submission tidak valid")
	}
	if !slices.Contains(ReminderTypes, submissionType) {
		return errors.New("tipe submission tidak valid")
	}

	return s.repo.MarkNotificationRead(uid, submissionType, submissionUUID)
}

func (s *notificationSettingService) MarkAllRemindersRead(userID string, access ReminderAccess) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("user tidak valid")
	}

	rows, err := s.loadReminders(uid, access)
	if err != nil {
		return err
	}

	reads := make([]entity.NotificationRead, 0, len(rows))
	for _, row := range rows {
		submissionID, err := uuid.Parse(row.item.ID)
		if err != nil {
			return errors.New("submission tidak valid")
		}
		reads = append(reads, entity.NotificationRead{
			SubmissionType: row.item.SubmissionType,
			SubmissionID:   submissionID,
		})
	}

	return s.repo.MarkNotificationsRead(uid, reads)
}

// loadReminders mengumpulkan semua baris lonceng seorang user dari tiap modul
// pengajuan yang boleh ia lihat atau proses.
func (s *notificationSettingService) loadReminders(userID uuid.UUID, access ReminderAccess) ([]reminderRow, error) {
	settings, err := s.GetActiveSettings()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	rows := []reminderRow{}
	add := func(src reminderSource) {
		thresholds := settings[src.submissionType]
		row, ok := buildReminderRow(src, access[src.submissionType], userID,
			thresholdDays(thresholds["YELLOW"]), thresholdDays(thresholds["RED"]), now)
		if ok {
			rows = append(rows, row)
		}
	}

	if scope := access["legal_opinion"]; scope.Owner || scope.Processor {
		items, err := s.dashboardRepo.ReminderLegalOpinions(userID, statusesIf(scope.Owner, ownerStatuses), statusesIf(scope.Processor, processorStatuses))
		if err != nil {
			return nil, err
		}
		for i := range items {
			lo := &items[i]
			add(reminderSource{
				submissionType: "legal_opinion", id: lo.ID, ownerID: lo.UserID, requesterName: lo.User.FullName,
				ticketNumber: lo.TicketNumber, title: lo.Title, status: lo.Status, note: lo.AdminNote,
				createdAt: lo.CreatedAt, statusUpdatedAt: lo.StatusUpdatedAt,
			})
		}
	}

	if scope := access["document_review"]; scope.Owner || scope.Processor {
		items, err := s.dashboardRepo.ReminderDocumentReviews(userID, statusesIf(scope.Owner, ownerStatuses), statusesIf(scope.Processor, processorStatuses))
		if err != nil {
			return nil, err
		}
		for i := range items {
			dr := &items[i]
			add(reminderSource{
				submissionType: "document_review", id: dr.ID, ownerID: dr.UserID, requesterName: dr.User.FullName,
				ticketNumber: dr.TicketNumber, title: dr.DocumentName, status: dr.Status, note: dr.AdminNote,
				createdAt: dr.CreatedAt, statusUpdatedAt: dr.StatusUpdatedAt,
			})
		}
	}

	// Perjanjian tidak punya ambang peringatan, jadi pemroses hanya menerima
	// yang baru masuk dan yang diajukan ulang
	if scope := access["agreement_document"]; scope.Owner || scope.Processor {
		items, err := s.dashboardRepo.ReminderAgreementDocuments(userID, statusesIf(scope.Owner, ownerStatuses), statusesIf(scope.Processor, processorStatuses))
		if err != nil {
			return nil, err
		}
		for i := range items {
			doc := &items[i]
			add(reminderSource{
				submissionType: "agreement_document", id: doc.ID, ownerID: doc.UserID, requesterName: doc.User.FullName,
				ticketNumber: doc.TicketNumber, title: agreementTitle(doc), status: doc.Status, note: doc.ApproverNote,
				createdAt: doc.CreatedAt, statusUpdatedAt: doc.StatusUpdatedAt,
			})
		}
	}

	return rows, nil
}

// agreementTitle menyusun judul baris lonceng untuk perjanjian, yang tidak punya
// kolom judul: jenis dokumennya, ditambah nama pihak kedua kalau formulirnya
// memuatnya.
func agreementTitle(doc *entity.AgreementDocument) string {
	var form struct {
		PihakKedua string `json:"pihak_kedua_nama"`
	}
	_ = json.Unmarshal(doc.FormData, &form)
	if form.PihakKedua == "" {
		return doc.DocumentTypeCode
	}
	return doc.DocumentTypeCode + " - " + form.PihakKedua
}

func statusesIf(ok bool, statuses []entity.SubmissionStatus) []entity.SubmissionStatus {
	if ok {
		return statuses
	}
	return nil
}

// thresholdDays mengembalikan 0 untuk ambang yang tidak ada atau tidak aktif.
func thresholdDays(setting *entity.NotificationSetting) int {
	if setting == nil {
		return 0
	}
	return setting.DaysThreshold
}

// buildReminderRow menentukan apakah satu pengajuan masuk lonceng seorang user.
// Pemilik diberi kabar status yang diubah pemroses. Pemroses diberi kabar
// pengajuan yang baru masuk atau diajukan ulang, dan pengajuan yang sedang
// direview hanya kalau sudah melewati ambang peringatan.
func buildReminderRow(src reminderSource, scope ReminderScope, viewerID uuid.UUID, yellowDays, redDays int, now time.Time) (reminderRow, bool) {
	eventAt := src.createdAt
	if src.statusUpdatedAt != nil {
		eventAt = *src.statusUpdatedAt
	}
	daysWaiting := int(now.Sub(eventAt).Hours() / 24)

	// Peringatan hanya untuk pemroses, dan hanya selama pengajuan masih menunggunya
	level, changedAt := WarningLevelNone, eventAt
	if scope.Processor && slices.Contains(processorStatuses, src.status) {
		level, changedAt = warningLevel(eventAt, daysWaiting, yellowDays, redDays)
	}

	finished := src.status == entity.StatusCompleted || src.status == entity.StatusRejected
	asOwner := scope.Owner && src.ownerID == viewerID && slices.Contains(ownerStatuses, src.status) &&
		!(finished && now.Sub(eventAt) > finishedNewsMaxAge)
	asProcessor := scope.Processor &&
		(src.status == entity.StatusSubmitted || src.status == entity.StatusResubmitted || level != WarningLevelNone)
	if !asOwner && !asProcessor {
		return reminderRow{}, false
	}

	// Catatan lama tetap tersimpan setelah status berganti, jadi hanya dikirim
	// selagi masih berlaku
	note := ""
	if src.status == entity.StatusNeedRevision || src.status == entity.StatusRejected {
		note = src.note
	}

	return reminderRow{
		item: dto.ReminderItem{
			ID:             src.id.String(),
			SubmissionType: src.submissionType,
			TicketNumber:   src.ticketNumber,
			Title:          src.title,
			Status:         string(src.status),
			RequesterName:  src.requesterName,
			Note:           note,
			EventAt:        eventAt.Format(time.RFC3339),
			DaysWaiting:    daysWaiting,
			WarningLevel:   string(level),
		},
		changedAt: changedAt,
	}, true
}

// warningLevel mengembalikan level peringatan dan kapan pengajuan masuk ke
// level itu. Ambang 0 berarti tidak aktif.
func warningLevel(eventAt time.Time, daysWaiting, yellowDays, redDays int) (WarningLevel, time.Time) {
	const day = 24 * time.Hour
	if redDays > 0 && daysWaiting >= redDays {
		return WarningLevelRed, eventAt.Add(time.Duration(redDays) * day)
	}
	if yellowDays > 0 && daysWaiting >= yellowDays {
		return WarningLevelYellow, eventAt.Add(time.Duration(yellowDays) * day)
	}
	return WarningLevelNone, eventAt
}

// applyReminderReadState menandai baris yang sudah dibaca. Tanda terbaca gugur
// kalau baris berubah lagi setelah dibaca: statusnya berganti, atau levelnya
// naik.
func (s *notificationSettingService) applyReminderReadState(rows []reminderRow, userID uuid.UUID) error {
	reads, err := s.repo.GetNotificationReadsByUser(userID)
	if err != nil {
		return err
	}

	readAt := make(map[string]time.Time, len(reads))
	for _, read := range reads {
		at := read.UpdatedAt
		if read.ReadAt != nil {
			at = *read.ReadAt
		}
		readAt[read.SubmissionType+":"+read.SubmissionID.String()] = at
	}
	for i := range rows {
		at, ok := readAt[rows[i].item.SubmissionType+":"+rows[i].item.ID]
		rows[i].item.IsRead = ok && !at.Before(rows[i].changedAt)
	}
	return nil
}

// paginateReminders mengurutkan dari yang terakhir berubah, lalu memotong
// sesuai halaman.
func paginateReminders(rows []reminderRow, page int, limit int) *dto.RemindersResponse {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i].changedAt.After(rows[j].changedAt)
	})

	unreadTotal := 0
	for _, row := range rows {
		if !row.item.IsRead {
			unreadTotal++
		}
	}

	total := len(rows)
	start := min((page-1)*limit, total)
	end := min(start+limit, total)
	items := make([]dto.ReminderItem, 0, end-start)
	for _, row := range rows[start:end] {
		items = append(items, row.item)
	}

	return &dto.RemindersResponse{
		Items:       items,
		Total:       total,
		UnreadTotal: unreadTotal,
		Page:        page,
		Limit:       limit,
		TotalPages:  (total + limit - 1) / limit,
	}
}
