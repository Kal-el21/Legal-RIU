package service

import (
	"errors"
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
	GetReminders(userID string, role string, page int, limit int) (*dto.RemindersResponse, error)
	MarkReminderRead(userID string, submissionType string, submissionID string) error
	MarkAllRemindersRead(userID string, role string) error
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

func (s *notificationSettingService) GetReminders(userID string, role string, page int, limit int) (*dto.RemindersResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	result, err := s.buildReminderResponse(uid, role)
	if err != nil {
		return nil, err
	}

	if err := s.applyReminderReadState(result, uid); err != nil {
		return nil, err
	}
	s.applyReminderPagination(result, page, limit)
	return result, nil
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
	if submissionType != "legal_opinion" && submissionType != "document_review" {
		return errors.New("tipe submission tidak valid")
	}

	return s.repo.MarkNotificationRead(uid, submissionType, submissionUUID)
}

func (s *notificationSettingService) MarkAllRemindersRead(userID string, role string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("user tidak valid")
	}

	result, err := s.buildReminderResponse(uid, role)
	if err != nil {
		return err
	}

	items := make([]dto.ReminderItem, 0, len(result.Red)+len(result.Yellow))
	items = append(items, result.Red...)
	items = append(items, result.Yellow...)

	reads := make([]entity.NotificationRead, 0, len(items))
	for _, item := range items {
		submissionID, err := uuid.Parse(item.ID)
		if err != nil {
			return errors.New("submission tidak valid")
		}
		reads = append(reads, entity.NotificationRead{
			SubmissionType: item.SubmissionType,
			SubmissionID:   submissionID,
		})
	}

	return s.repo.MarkNotificationsRead(uid, reads)
}

func (s *notificationSettingService) buildReminderResponse(userID uuid.UUID, role string) (*dto.RemindersResponse, error) {
	settings, err := s.GetActiveSettings()
	if err != nil {
		return nil, err
	}

	yellowThresholds := make(map[string]int)
	redThresholds := make(map[string]int)

	if val, ok := settings["legal_opinion"]; ok {
		if setting, ok := val["YELLOW"]; ok {
			yellowThresholds["legal_opinion"] = setting.DaysThreshold
		}
		if setting, ok := val["RED"]; ok {
			redThresholds["legal_opinion"] = setting.DaysThreshold
		}
	}
	if val, ok := settings["document_review"]; ok {
		if setting, ok := val["YELLOW"]; ok {
			yellowThresholds["document_review"] = setting.DaysThreshold
		}
		if setting, ok := val["RED"]; ok {
			redThresholds["document_review"] = setting.DaysThreshold
		}
	}

	result := &dto.RemindersResponse{
		Yellow: []dto.ReminderItem{},
		Red:    []dto.ReminderItem{},
		None:   []dto.ReminderItem{},
	}

	if role == "ADMIN" || role == "LEGAL" {
		if err := s.loadAllReminders(result, yellowThresholds, redThresholds); err != nil {
			return nil, err
		}
		return result, nil
	}

	if err := s.loadUserReminders(result, userID, yellowThresholds, redThresholds); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *notificationSettingService) loadAllReminders(result *dto.RemindersResponse, yellowThresholds, redThresholds map[string]int) error {
	los, err := s.dashboardRepo.RecentAllLegalOpinions(1000)
	if err != nil {
		return err
	}
	for i := range los {
		item := s.mapLOToReminder(&los[i], yellowThresholds["legal_opinion"], redThresholds["legal_opinion"])
		s.appendReminder(result, &item)
	}

	drs, err := s.dashboardRepo.RecentAllDocumentReviews(1000)
	if err != nil {
		return err
	}
	for i := range drs {
		item := s.mapDRToReminder(&drs[i], yellowThresholds["document_review"], redThresholds["document_review"])
		s.appendReminder(result, &item)
	}
	return nil
}

func (s *notificationSettingService) loadUserReminders(result *dto.RemindersResponse, userID uuid.UUID, yellowThresholds, redThresholds map[string]int) error {
	los, err := s.dashboardRepo.RecentLegalOpinionsByUser(userID, 1000)
	if err != nil {
		return err
	}
	for i := range los {
		item := s.mapLOToReminder(&los[i], yellowThresholds["legal_opinion"], redThresholds["legal_opinion"])
		s.appendReminder(result, &item)
	}

	drs, err := s.dashboardRepo.RecentDocumentReviewsByUser(userID, 1000)
	if err != nil {
		return err
	}
	for i := range drs {
		item := s.mapDRToReminder(&drs[i], yellowThresholds["document_review"], redThresholds["document_review"])
		s.appendReminder(result, &item)
	}
	return nil
}

func (s *notificationSettingService) mapLOToReminder(lo *entity.LegalOpinion, yellowDays, redDays int) dto.ReminderItem {
	submittedAt := lo.CreatedAt
	lastUpdatedAt := &lo.UpdatedAt
	if lo.UpdatedAt.Equal(lo.CreatedAt) {
		lastUpdatedAt = nil
	}

	daysSinceSubmission := int(time.Since(submittedAt).Hours() / 24)
	daysSinceLastUpdate := 0
	if lastUpdatedAt != nil {
		daysSinceLastUpdate = int(time.Since(*lastUpdatedAt).Hours() / 24)
	}

	level, color := s.calculateWarning(daysSinceLastUpdate, daysSinceSubmission, yellowDays, redDays)

	name := ""
	if lo.User.FullName != "" {
		name = lo.User.FullName
	}

	var lastUpdatedStr *string
	if lastUpdatedAt != nil {
		value := lastUpdatedAt.Format(time.RFC3339)
		lastUpdatedStr = &value
	}

	return dto.ReminderItem{
		ID:                  lo.ID.String(),
		SubmissionType:      "legal_opinion",
		TicketNumber:        lo.TicketNumber,
		Title:               lo.Title,
		Status:              string(lo.Status),
		SubmittedAt:         submittedAt.Format(time.RFC3339),
		LastUpdatedAt:       lastUpdatedStr,
		DaysSinceSubmission: daysSinceSubmission,
		DaysSinceLastUpdate: daysSinceLastUpdate,
		WarningLevel:        string(level),
		WarningColor:        color,
		AssignedLegalName:   name,
	}
}

func (s *notificationSettingService) mapDRToReminder(dr *entity.DocumentReview, yellowDays, redDays int) dto.ReminderItem {
	submittedAt := dr.CreatedAt
	lastUpdatedAt := &dr.UpdatedAt
	if dr.UpdatedAt.Equal(dr.CreatedAt) {
		lastUpdatedAt = nil
	}

	daysSinceSubmission := int(time.Since(submittedAt).Hours() / 24)
	daysSinceLastUpdate := 0
	if lastUpdatedAt != nil {
		daysSinceLastUpdate = int(time.Since(*lastUpdatedAt).Hours() / 24)
	}

	level, color := s.calculateWarning(daysSinceLastUpdate, daysSinceSubmission, yellowDays, redDays)

	name := ""
	if dr.User.FullName != "" {
		name = dr.User.FullName
	}

	var lastUpdatedStr *string
	if lastUpdatedAt != nil {
		value := lastUpdatedAt.Format(time.RFC3339)
		lastUpdatedStr = &value
	}

	return dto.ReminderItem{
		ID:                  dr.ID.String(),
		SubmissionType:      "document_review",
		TicketNumber:        dr.TicketNumber,
		Title:               dr.DocumentName,
		Status:              string(dr.Status),
		SubmittedAt:         submittedAt.Format(time.RFC3339),
		LastUpdatedAt:       lastUpdatedStr,
		DaysSinceSubmission: daysSinceSubmission,
		DaysSinceLastUpdate: daysSinceLastUpdate,
		WarningLevel:        string(level),
		WarningColor:        color,
		AssignedLegalName:   name,
	}
}

func (s *notificationSettingService) calculateWarning(daysSinceLastUpdate, daysSinceSubmission, yellowDays, redDays int) (WarningLevel, string) {
	if yellowDays <= 0 && redDays > 0 {
		if daysSinceSubmission >= redDays || daysSinceLastUpdate >= redDays {
			return WarningLevelRed, "#DC2626"
		}
		return WarningLevelNone, ""
	}
	if redDays <= 0 && yellowDays > 0 {
		if daysSinceSubmission >= yellowDays || daysSinceLastUpdate >= yellowDays {
			return WarningLevelYellow, "#F59E0B"
		}
		return WarningLevelNone, ""
	}

	if daysSinceSubmission >= redDays || daysSinceLastUpdate >= redDays {
		return WarningLevelRed, "#DC2626"
	}
	if daysSinceSubmission >= yellowDays || daysSinceLastUpdate >= yellowDays {
		return WarningLevelYellow, "#F59E0B"
	}
	return WarningLevelNone, ""
}

func (s *notificationSettingService) appendReminder(result *dto.RemindersResponse, item *dto.ReminderItem) {
	switch WarningLevel(item.WarningLevel) {
	case WarningLevelRed:
		result.Red = append(result.Red, *item)
	case WarningLevelYellow:
		result.Yellow = append(result.Yellow, *item)
	default:
		result.None = append(result.None, *item)
	}
}

func (s *notificationSettingService) applyReminderReadState(result *dto.RemindersResponse, userID uuid.UUID) error {
	reads, err := s.repo.GetNotificationReadsByUser(userID)
	if err != nil {
		return err
	}

	readMap := make(map[string]bool, len(reads))
	for _, read := range reads {
		readMap[read.SubmissionType+":"+read.SubmissionID.String()] = read.IsRead
	}

	apply := func(items []dto.ReminderItem) {
		for i := range items {
			items[i].IsRead = readMap[items[i].SubmissionType+":"+items[i].ID]
		}
	}

	apply(result.Red)
	apply(result.Yellow)
	apply(result.None)
	return nil
}

func (s *notificationSettingService) applyReminderPagination(result *dto.RemindersResponse, page int, limit int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	items := make([]dto.ReminderItem, 0, len(result.Red)+len(result.Yellow))
	items = append(items, result.Red...)
	items = append(items, result.Yellow...)

	sort.SliceStable(items, func(i, j int) bool {
		return items[i].SubmittedAt > items[j].SubmittedAt
	})

	total := len(items)
	unreadTotal := 0
	for _, item := range items {
		if !item.IsRead {
			unreadTotal++
		}
	}
	totalPages := 0
	if total > 0 {
		totalPages = (total + limit - 1) / limit
	}

	start := (page - 1) * limit
	if start >= total {
		result.Items = []dto.ReminderItem{}
	} else {
		end := start + limit
		if end > total {
			end = total
		}
		result.Items = items[start:end]
	}

	result.Total = total
	result.UnreadTotal = unreadTotal
	result.Page = page
	result.Limit = limit
	result.TotalPages = totalPages
}
