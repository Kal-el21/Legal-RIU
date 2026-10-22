package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"strings"

	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/repository"
	"legal-riu-portal/internal/storage"
	"legal-riu-portal/internal/utils"

	"github.com/google/uuid"
)

// ErrInvalidFile menandai lampiran yang ditolak (tipe atau ukurannya). Handler
// menjawabnya 400 dengan alasannya.
var ErrInvalidFile = errors.New("lampiran ditolak")

// validateFiles memeriksa semua lampiran sebelum ada yang disimpan, supaya satu
// file yang ditolak tidak meninggalkan pengajuan kosong atau lampiran separuh jadi.
func validateFiles(files []*multipart.FileHeader) error {
	for _, file := range files {
		if err := storage.ValidateUpload(file); err != nil {
			return fmt.Errorf("%w: %s (%s)", ErrInvalidFile, err.Error(), file.Filename)
		}
	}
	return nil
}

// ErrFileNotFound dijawab 404 oleh alamat unduh: file tidak terdaftar di
// modulnya, pemanggil tidak boleh melihat pemiliknya, atau file tercatat tetapi
// tidak ada di penyimpanan. Nilainya sama dengan storage.ErrNotFound supaya
// ketiganya tertangkap oleh satu pemeriksaan errors.Is.
var ErrFileNotFound = storage.ErrNotFound

// Galat alamat draf (PR Draf). ErrDrafTidakAda dan ErrLampiranTidakAda dijawab 404:
// draf tidak ada, bukan milik pemanggil, atau pengajuannya sudah dikirim.
// ErrIsianBelumLengkap dijawab 400 dengan pesan yang sama seperti saat membuat pengajuan.
var (
	ErrDrafTidakAda      = errors.New("draf tidak ditemukan")
	ErrLampiranTidakAda  = errors.New("lampiran tidak ditemukan")
	ErrIsianBelumLengkap = errors.New("Semua field wajib diisi")
)

// nomorDraf: nomor tiket sementara untuk draf (keputusan K1). Polanya tidak cocok dengan
// <prefix>-YYYYMM-NNNN, jadi urutan nomor asli tidak terganggu oleh draf yang dibuang.
func nomorDraf() string {
	return "DRAF-" + uuid.NewString()
}

// catatRepositori: sinkron Repositori Dokumen yang gagal hanya ditulis ke log server.
// Pengajuannya tetap tersimpan, dan backfill saat server menyala melengkapinya.
func catatRepositori(id uuid.UUID, err error) {
	if err != nil {
		log.Printf("Repositori Dokumen untuk pengajuan %s tidak tersinkron: %v", id, err)
	}
}

// hapusFileDraf menghapus file draf dari MinIO. Kegagalan diabaikan: barisnya sudah
// terhapus, jadi file yang tertinggal tidak bisa diunduh siapa pun.
func hapusFileDraf(store *storage.MinIOClient, paths []string) {
	for _, p := range paths {
		_ = store.DeleteFile(context.Background(), p)
	}
}

// denganNomorTiket memberi nomor tiket berikutnya, lalu menyimpan pengajuan lewat simpan (buat
// baru atau kirim draf). Nomor diambil dari yang terbesar bulan ini di luar transaksi, jadi dua
// pengajuan yang disimpan bersamaan bisa mendapat nomor sama, dan yang kalah ditolak indeks unik
// ticket_number. Kalau itu terjadi, nomor berikutnya diambil dan penyimpanan diulang, maksimal 3
// kali (PR A tahap A.3, keputusan A2). Galat lain langsung dikembalikan.
func denganNomorTiket(prefix utils.TicketPrefix, berikut func(string) (int, error), simpan func(nomor string) error) error {
	var err error
	for coba := 0; coba < 3; coba++ {
		var seq int
		if seq, err = berikut(string(prefix)); err != nil {
			return err
		}
		if err = simpan(utils.GenerateTicketNumber(prefix, seq)); !repository.NomorTiketBentrok(err) {
			return err
		}
	}
	return err
}

func parseUUID(id string) (uuid.UUID, error) {
	return uuid.Parse(id)
}

func canAccessAllSubmissions(role string) bool {
	return role == string(entity.RoleAdmin) || role == string(entity.RoleLegal)
}

func toUserResponse(user *entity.User) dto.UserResponse {
	division := user.Division
	divisionID := ""
	var divisionDetail *dto.DivisionResponse
	if user.DivisionRef.ID != uuid.Nil {
		division = user.DivisionRef.Name
		divisionID = user.DivisionRef.ID.String()
		resp := toDivisionResponse(&user.DivisionRef)
		divisionDetail = &resp
	} else if user.DivisionID != nil && *user.DivisionID != uuid.Nil {
		divisionID = user.DivisionID.String()
	}

	resp := dto.UserResponse{
		ID:                 user.ID.String(),
		FullName:           user.FullName,
		Email:              user.Email,
		Position:           user.Position,
		Division:           division,
		DivisionID:         divisionID,
		DivisionDetail:     divisionDetail,
		Role:               string(user.Role),
		Status:             string(user.Status),
		EmailNotifications: user.EmailNotifications,
		TwoFAEnabled:       user.TwoFAEnabled,
		CompanyID:          "",
		PurposeTypeID:      "",
	}
	if user.CompanyID != nil {
		resp.CompanyID = user.CompanyID.String()
	}
	if user.Company.ID != uuid.Nil {
		company := toCompanyResponse(&user.Company)
		resp.Company = &company
	}
	if user.PurposeTypeID != nil {
		resp.PurposeTypeID = user.PurposeTypeID.String()
	}
	if user.PurposeType.ID != uuid.Nil {
		pt := toPurposeTypeResponse(&user.PurposeType)
		resp.PurposeType = &pt
	}
	return resp
}

func resolveDivisionSelection(userRepo repository.UserRepository, value string) (string, *uuid.UUID, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil, errors.New("divisi wajib diisi")
	}

	if id, err := uuid.Parse(value); err == nil {
		division, err := userRepo.FindDivisionByID(id)
		if err != nil {
			return "", nil, errors.New("divisi tidak ditemukan")
		}
		divisionID := division.ID
		return division.Name, &divisionID, nil
	}

	division, err := userRepo.FindDivisionByName(value)
	if err != nil {
		return "", nil, errors.New("divisi tidak ditemukan")
	}
	divisionID := division.ID
	return division.Name, &divisionID, nil
}

// resolveCompanySelection menentukan perusahaan user dari perannya. LEGAL adalah
// legal induk dan LEGAL_AU legal anak usaha, jadi keduanya tidak bebas memilih:
// kalau tertukar, LEGAL ikut dibatasi ke perkara anak usaha, atau LEGAL_AU
// melihat perkara semua perusahaan. Aturannya ada di docs/case_company_access.md.
func resolveCompanySelection(userRepo repository.UserRepository, role string, companyIDStr string) (*uuid.UUID, error) {
	switch role {
	case "EXTERNAL":
		return nil, nil
	case "LEGAL":
		// Kiriman form diabaikan, supaya LEGAL_AU yang dijadikan LEGAL ikut pindah.
		company, err := userRepo.FindParentCompany()
		if err != nil {
			return nil, errors.New("perusahaan induk tidak ditemukan")
		}
		return &company.ID, nil
	case "LEGAL_AU":
		if companyIDStr == "" {
			return nil, errors.New("perusahaan wajib dipilih untuk role LEGAL_AU")
		}
		companyID, err := uuid.Parse(companyIDStr)
		if err != nil {
			return nil, errors.New("ID perusahaan tidak valid")
		}
		company, err := userRepo.FindCompanyByID(companyID)
		if err != nil {
			return nil, errors.New("perusahaan tidak ditemukan")
		}
		if company.IsParent {
			return nil, errors.New("role LEGAL_AU harus terdaftar di anak usaha, bukan di perusahaan induk")
		}
		return &company.ID, nil
	default:
		if companyIDStr != "" {
			companyID, err := uuid.Parse(companyIDStr)
			if err != nil {
				return nil, errors.New("ID perusahaan tidak valid")
			}
			company, err := userRepo.FindCompanyByID(companyID)
			if err != nil {
				return nil, errors.New("perusahaan tidak ditemukan")
			}
			return &company.ID, nil
		}
		company, err := userRepo.FindParentCompany()
		if err == nil {
			return &company.ID, nil
		}
		company, err = userRepo.FindFirstCompany()
		if err != nil {
			return nil, errors.New("tidak ada perusahaan yang tersedia")
		}
		return &company.ID, nil
	}
}

func toDivisionResponse(division *entity.Division) dto.DivisionResponse {
	return dto.DivisionResponse{
		ID:          division.ID.String(),
		Name:        division.Name,
		Description: division.Description,
		CreatedAt:   division.CreatedAt,
		UpdatedAt:   division.UpdatedAt,
	}
}

func toDivisionResponsePointer(division *entity.Division) *dto.DivisionResponse {
	if division == nil || division.ID == uuid.Nil {
		return nil
	}
	resp := toDivisionResponse(division)
	return &resp
}

func toCompanyResponse(company *entity.Company) dto.CompanyResponse {
	return dto.CompanyResponse{
		ID:          company.ID.String(),
		Name:        company.Name,
		EmailDomain: company.EmailDomain,
		IsInternal:  company.IsInternal,
		IsParent:    company.IsParent,
		CreatedAt:   company.CreatedAt,
		UpdatedAt:   company.UpdatedAt,
	}
}

func toPurposeTypeResponse(pt *entity.PurposeType) dto.PurposeTypeResponse {
	return dto.PurposeTypeResponse{
		ID:          pt.ID.String(),
		Name:        pt.Name,
		Description: pt.Description,
		IsActive:    pt.IsActive,
		CreatedAt:   pt.CreatedAt,
		UpdatedAt:   pt.UpdatedAt,
	}
}

func toCaseTypeResponse(caseType *entity.CaseType) dto.CaseTypeResponse {
	return dto.CaseTypeResponse{
		ID:        caseType.ID.String(),
		Code:      caseType.Code,
		Label:     caseType.Label,
		IsActive:  caseType.IsActive,
		CreatedAt: caseType.CreatedAt,
		UpdatedAt: caseType.UpdatedAt,
	}
}

func toCaseCategoryResponse(caseCategory *entity.CaseCategory) dto.CaseCategoryResponse {
	return dto.CaseCategoryResponse{
		ID:        caseCategory.ID.String(),
		Code:      caseCategory.Code,
		Label:     caseCategory.Label,
		IsActive:  caseCategory.IsActive,
		CreatedAt: caseCategory.CreatedAt,
		UpdatedAt: caseCategory.UpdatedAt,
	}
}
