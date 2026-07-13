package dto

// CreateAgreementRequest is the payload submitted by a requester (USER/EXTERNAL).
// Only Pihak Kedua and the agreement details are filled; Pihak Pertama master
// data is injected by the system and the pejabat/jabatan are filled by the approver.
type CreateAgreementRequest struct {
	// Informasi umum
	NomorPihakKedua   string `json:"nomor_pihak_kedua"`
	TempatTtd         string `json:"tempat_ttd"`
	TanggalTtd        string `json:"tanggal_ttd"`

	// Pihak Kedua
	PihakKeduaNama     string `json:"pihak_kedua_nama" binding:"required"`
	PihakKeduaBidang   string `json:"pihak_kedua_bidang"`
	PihakKeduaAlamat   string `json:"pihak_kedua_alamat"`
	PihakKeduaTelepon  string `json:"pihak_kedua_telepon"`
	PihakKeduaEmail    string `json:"pihak_kedua_email"`
	PihakKeduaPic      string `json:"pihak_kedua_pic"`
	PihakKeduaPejabat  string `json:"pihak_kedua_pejabat" binding:"required"`
	PihakKeduaJabatan  string `json:"pihak_kedua_jabatan" binding:"required"`

	// Dasar hukum
	JenisPekerjaan         string `json:"jenis_pekerjaan" binding:"required"`
	SuratPenawaranNomor    string `json:"surat_penawaran_nomor"`
	SuratPenawaranPerihal  string `json:"surat_penawaran_perihal"`
	SuratPenawaranTanggal  string `json:"surat_penawaran_tanggal"`
	SuratPenunjukanNomor   string `json:"surat_penunjukan_nomor"`
	SuratPenunjukanPerihal string `json:"surat_penunjukan_perihal"`
	SuratPenunjukanTanggal string `json:"surat_penunjukan_tanggal"`

	// Ketentuan khusus
	RuangLingkup      string  `json:"ruang_lingkup" binding:"required"`
	JangkaWaktuMulai  string  `json:"jangka_waktu_mulai"`
	JangkaWaktuSelesai string `json:"jangka_waktu_selesai"`
	NilaiKontrak      float64 `json:"nilai_kontrak"`
	Termin1Persen     float64 `json:"termin1_persen"`
	Termin1Nilai      float64 `json:"termin1_nilai"`
	Termin2Persen     float64 `json:"termin2_persen"`
	Termin2Nilai      float64 `json:"termin2_nilai"`
	Bank              string  `json:"bank"`
	NomorRekening     string  `json:"nomor_rekening"`
	AtasNama          string  `json:"atas_nama"`
}

type UpdateAgreementRequest = CreateAgreementRequest

// UpdatePihakPertamaRequest is used by the approver to fill Pihak Pertama details.
type UpdatePihakPertamaRequest struct {
	PihakPertamaPejabat string `json:"pihak_pertama_pejabat"`
	PihakPertamaJabatan string `json:"pihak_pertama_jabatan"`
}

type AgreementListQuery struct {
	Page   int    `form:"page,default=1"`
	Limit  int    `form:"limit,default=10"`
	Status string `form:"status"`
}

type AgreementDecisionRequest struct {
	AdminNote string `json:"admin_note"`
}
