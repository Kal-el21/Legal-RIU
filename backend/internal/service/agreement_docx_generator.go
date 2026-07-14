package service

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"html"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type AgreementGenerator struct{}

func NewAgreementGenerator() *AgreementGenerator { return &AgreementGenerator{} }

var wordTextRE = regexp.MustCompile(`(?s)(<w:t(?:\s[^>]*)?>)(.*?)(</w:t>)`)
var placeholderRE = regexp.MustCompile(`\{\{[A-Z0-9_]+\}\}`)

func (g *AgreementGenerator) Generate(template []byte, values map[string]string, draft bool) ([]byte, string, error) {
	if len(template) == 0 {
		return nil, "", errors.New("template dokumen tidak ditemukan")
	}
	r, err := zip.NewReader(bytes.NewReader(template), int64(len(template)))
	if err != nil {
		return nil, "", errors.New("template bukan DOCX yang valid")
	}
	var out bytes.Buffer
	w := zip.NewWriter(&out)
	for _, f := range r.File {
		header := f.FileHeader
		header.CRC32 = 0
		header.CompressedSize = 0
		header.CompressedSize64 = 0
		header.UncompressedSize = 0
		header.UncompressedSize64 = 0
		header.Method = zip.Deflate
		dst, e := w.CreateHeader(&header)
		if e != nil {
			return nil, "", e
		}
		src, e := f.Open()
		if e != nil {
			return nil, "", e
		}
		data, e := io.ReadAll(src)
		src.Close()
		if e != nil {
			return nil, "", e
		}
		if strings.HasPrefix(f.Name, "word/") && strings.HasSuffix(f.Name, ".xml") {
			text := string(data)
			if f.Name == "word/document.xml" && !strings.Contains(text, "{{NOMOR_PIHAK_PERTAMA}}") {
				text = prepareLegacyPKS(text)
			}
			for key, val := range values {
				for {
					next := replaceAcrossWordRuns(text, "{{"+key+"}}", wordValue(val))
					if next == text {
						break
					}
					text = next
				}
			}
			if draft && f.Name == "word/document.xml" {
				text = strings.Replace(text, "<w:body>", `<w:body><w:p><w:r><w:rPr><w:b/><w:color w:val="C8102E"/></w:rPr><w:t>DRAFT - PREVIEW</w:t></w:r></w:p>`, 1)
			}
			data = []byte(text)
		}
		if _, e = dst.Write(data); e != nil {
			return nil, "", e
		}
	}
	if err = w.Close(); err != nil {
		return nil, "", err
	}
	result := out.Bytes()
	if left := findDOCXPlaceholders(result); len(left) > 0 {
		return nil, "", fmt.Errorf("placeholder belum terisi: %s", strings.Join(left, ", "))
	}
	sum := sha256.Sum256(template)
	return result, hex.EncodeToString(sum[:]), nil
}

func prepareLegacyPKS(text string) string {
	replacements := [][2]string{
		{"____________", "{{PIHAK_KEDUA_NAMA}}"}, {"____________", "{{JENIS_PEKERJAAN}}"},
		{"***/RM.01.01/**/IndonesiaRe/**/2026", "{{NOMOR_PIHAK_PERTAMA}}"},
		{"NOMOR PIHAK KEDUA", "NOMOR PIHAK KEDUA{{NOMOR_PIHAK_KEDUA}}"},
		{"Pada hari ini, [*], tanggal [*] bulan [*] tahun [*] (dd/mm/yyyy) bertempat di [*]", "Pada hari ini, {{HARI_TTD}}, tanggal {{TANGGAL_TTD}} bulan {{BULAN_TTD}} tahun {{TAHUN_TTD}} ({{TANGGAL_TTD_LENGKAP}}) bertempat di {{TEMPAT_TTD}}"},
		{"diwakili oleh ____________, selaku ____________", "diwakili oleh {{PIHAK_PERTAMA_PEJABAT}}, selaku {{PIHAK_PERTAMA_JABATAN}}"},
		{"____________, _________, selanjutnya disebut “PIHAK KEDUA”", "{{PIHAK_KEDUA_NAMA}}, beralamat di {{PIHAK_KEDUA_ALAMAT}}, dalam hal ini diwakili oleh {{PIHAK_KEDUA_PEJABAT}} selaku {{PIHAK_KEDUA_JABATAN}}, selanjutnya disebut “PIHAK KEDUA”"},
		{"bergerak di bidang ____________", "bergerak di bidang {{PIHAK_KEDUA_BIDANG}}"},
		{"pengadaan _____________", "pengadaan {{JENIS_PEKERJAAN}}"},
		{"Surat Penawaran No. ____________ perihal ____________pada tanggal ____________", "Surat Penawaran No. {{SURAT_PENAWARAN_NOMOR}} perihal {{SURAT_PENAWARAN_PERIHAL}} pada tanggal {{SURAT_PENAWARAN_TANGGAL}}"},
		{"Surat Penunjukan No. ____________ perihal ____________ pada tanggal ____________", "Surat Penunjukan No. {{SURAT_PENUNJUKAN_NOMOR}} perihal {{SURAT_PENUNJUKAN_PERIHAL}} pada tanggal {{SURAT_PENUNJUKAN_TANGGAL}}"},
		{"dengan _____ tentang ____________", "dengan {{PIHAK_KEDUA_NAMA}} tentang {{JENIS_PEKERJAAN}}"},
		{"Nama: ____________", "Nama: {{PIHAK_KEDUA_PEJABAT}}"}, {"Jabatan: ____________", "Jabatan: {{PIHAK_KEDUA_JABATAN}}"},
		{"Nama: ____________", "Nama: {{PIHAK_PERTAMA_PEJABAT}}"}, {"Jabatan: ____________", "Jabatan: {{PIHAK_PERTAMA_JABATAN}}"},
		{"____________", "{{PIHAK_KEDUA_NAMA}}"},
		{"Pekerjaan yang akan diserahkan kepada PIHAK KEDUA, yaitu ______.", "Pekerjaan yang akan diserahkan kepada PIHAK KEDUA, yaitu {{JENIS_PEKERJAAN}}."},
		{"RUANG LINGKUP PEKERJAAN", "RUANG LINGKUP PEKERJAAN{{RUANG_LINGKUP}}"},
		{"berlaku sejak tanggal _____ sampai dengan tanggal _______", "berlaku sejak tanggal {{JANGKA_WAKTU_MULAI}} sampai dengan tanggal {{JANGKA_WAKTU_SELESAI}}"},
		{"sebesar Rp. _____ (____________ rupiah)", "sebesar {{NILAI_KONTRAK}} ({{NILAI_KONTRAK_TERBILANG}} rupiah)"},
		{"sebesar **% (*** persen) dari Nilai Kontrak atau sebesar  Rp _______ ( _______ rupiah)", "sebesar {{TERMIN_1_PERSEN}}% ({{TERMIN_1_PERSEN_TERBILANG}} persen) dari Nilai Kontrak atau sebesar {{TERMIN_1_NILAI}} ({{TERMIN_1_NILAI_TERBILANG}} rupiah)"},
		{"sebesar **% (*** persen) dari Nilai Kontrak atau sebesar Rp. ______ (_______ rupiah)", "sebesar {{TERMIN_2_PERSEN}}% ({{TERMIN_2_PERSEN_TERBILANG}} persen) dari Nilai Kontrak atau sebesar {{TERMIN_2_NILAI}} ({{TERMIN_2_NILAI_TERBILANG}} rupiah)"},
		{"Bank:", "Bank: {{BANK}}"}, {"Nomor Rekening:", "Nomor Rekening: {{NOMOR_REKENING}}"}, {"Atas Nama:", "Atas Nama: {{ATAS_NAMA}}"},
		{"Alamat: Jl. Salemba Raya No. 30 Kenari Selatan, Jakarta Pusat", "Alamat: {{PIHAK_PERTAMA_ALAMAT}}"},
		{"Telepon: 021 3920101", "Telepon: {{PIHAK_PERTAMA_TELEPON}}"},
		{"________________________", "{{PIHAK_KEDUA_NAMA}}"},
		{"Alamat: ", "Alamat: {{PIHAK_KEDUA_ALAMAT}}"}, {"Telepon: ", "Telepon: {{PIHAK_KEDUA_TELEPON}}"},
		{"e-mail: ", "e-mail: {{PIHAK_PERTAMA_EMAIL}}"}, {"PIC:", "PIC: {{PIHAK_PERTAMA_PIC}}"},
		{"e-mail: ", "e-mail: {{PIHAK_KEDUA_EMAIL}}"}, {"PIC:", "PIC: {{PIHAK_KEDUA_PIC}}"},
		{"BAGIAN III", "BAGIAN III{{DAFTAR_LAMPIRAN}}"},
	}
	for _, r := range replacements {
		text = replaceAcrossWordRuns(text, r[0], r[1])
	}
	return text
}
func replaceAcrossWordRuns(xmlText, needle, replacement string) string {
	matches := wordTextRE.FindAllStringSubmatchIndex(xmlText, -1)
	if len(matches) == 0 {
		return xmlText
	}
	var logical strings.Builder
	type node struct {
		contentStart, contentEnd, logicalStart, logicalEnd int
		text                                               string
	}
	nodes := make([]node, 0, len(matches))
	pos := 0
	for _, m := range matches {
		t := html.UnescapeString(xmlText[m[4]:m[5]])
		nodes = append(nodes, node{m[4], m[5], pos, pos + len(t), t})
		pos += len(t)
		logical.WriteString(t)
	}
	idx := strings.Index(logical.String(), needle)
	if idx < 0 {
		return xmlText
	}
	end := idx + len(needle)
	parts := make([]string, len(nodes))
	first := -1
	last := -1
	for i, n := range nodes {
		parts[i] = n.text
		if idx < n.logicalEnd && end > n.logicalStart {
			if first < 0 {
				first = i
			}
			last = i
		}
	}
	if first < 0 {
		return xmlText
	}
	startOff := idx - nodes[first].logicalStart
	endOff := end - nodes[last].logicalStart
	if first == last {
		parts[first] = parts[first][:startOff] + replacement + parts[first][endOff:]
	} else {
		parts[first] = parts[first][:startOff] + replacement
		for i := first + 1; i < last; i++ {
			parts[i] = ""
		}
		parts[last] = parts[last][endOff:]
	}
	var b strings.Builder
	cursor := 0
	for i, n := range nodes {
		b.WriteString(xmlText[cursor:n.contentStart])
		b.WriteString(parts[i])
		cursor = n.contentEnd
	}
	b.WriteString(xmlText[cursor:])
	return b.String()
}
func wordValue(v string) string {
	escaped := html.EscapeString(v)
	return strings.ReplaceAll(strings.ReplaceAll(escaped, "\r\n", `</w:t><w:br/><w:t>`), "\n", `</w:t><w:br/><w:t>`)
}
func findDOCXPlaceholders(data []byte) []string {
	r, e := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if e != nil {
		return []string{"DOCX_INVALID"}
	}
	set := map[string]bool{}
	for _, f := range r.File {
		if !strings.HasPrefix(f.Name, "word/") || !strings.HasSuffix(f.Name, ".xml") {
			continue
		}
		s, _ := f.Open()
		b, _ := io.ReadAll(s)
		s.Close()
		for _, p := range placeholderRE.FindAllString(string(b), -1) {
			set[p] = true
		}
	}
	out := make([]string, 0, len(set))
	for p := range set {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

type DOCXConverter struct{}

func (DOCXConverter) ToPDF(ctx context.Context, docx []byte) ([]byte, error) {
	bin, err := exec.LookPath("libreoffice")
	if err != nil {
		bin, err = exec.LookPath("soffice")
	}
	if err != nil {
		return nil, errors.New("LibreOffice tidak tersedia")
	}
	dir, err := os.MkdirTemp("", "agreement-convert-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)
	input := filepath.Join(dir, "agreement.docx")
	if err = os.WriteFile(input, docx, 0600); err != nil {
		return nil, err
	}
	profileURI := "file://" + filepath.ToSlash(filepath.Join(dir, "libreoffice-profile"))
	cmd := exec.CommandContext(ctx, bin,
		"--headless",
		"--nologo",
		"--nodefault",
		"--nolockcheck",
		"--nofirststartwizard",
		"-env:UserInstallation="+profileURI,
		"--convert-to", "pdf",
		"--outdir", dir,
		input,
	)
	if out, e := cmd.CombinedOutput(); e != nil {
		return nil, fmt.Errorf("konversi PDF gagal: %s", strings.TrimSpace(string(out)))
	}
	pdf, err := os.ReadFile(filepath.Join(dir, "agreement.pdf"))
	if err != nil {
		return nil, fmt.Errorf("hasil konversi PDF tidak ditemukan: %w", err)
	}
	if len(pdf) < 5 || !bytes.Equal(pdf[:5], []byte("%PDF-")) {
		return nil, errors.New("hasil konversi bukan PDF yang valid")
	}
	return pdf, nil
}

var indoMonths = []string{"", "Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}
var indoDays = []string{"Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"}

func parseDate(s string) (time.Time, error) { return time.Parse("2006-01-02", s) }
func indoDate(s string) string {
	d, e := parseDate(s)
	if e != nil {
		return s
	}
	return fmt.Sprintf("%d %s %d", d.Day(), indoMonths[d.Month()], d.Year())
}
func rupiah(n int64) string {
	raw := fmt.Sprintf("%d", n)
	for i := len(raw) - 3; i > 0; i -= 3 {
		raw = raw[:i] + "." + raw[i:]
	}
	return "Rp " + raw
}
func terbilang(n int64) string {
	if n < 0 {
		return "minus " + terbilang(-n)
	}
	words := []string{"", "satu", "dua", "tiga", "empat", "lima", "enam", "tujuh", "delapan", "sembilan", "sepuluh", "sebelas"}
	switch {
	case n < 12:
		return words[n]
	case n < 20:
		return terbilang(n-10) + " belas"
	case n < 100:
		return terbilang(n/10) + " puluh" + suffixWords(n%10)
	case n < 200:
		return "seratus" + suffixWords(n-100)
	case n < 1000:
		return terbilang(n/100) + " ratus" + suffixWords(n%100)
	case n < 2000:
		return "seribu" + suffixWords(n-1000)
	case n < 1000000:
		return terbilang(n/1000) + " ribu" + suffixWords(n%1000)
	case n < 1000000000:
		return terbilang(n/1000000) + " juta" + suffixWords(n%1000000)
	case n < 1000000000000:
		return terbilang(n/1000000000) + " miliar" + suffixWords(n%1000000000)
	default:
		return terbilang(n/1000000000000) + " triliun" + suffixWords(n%1000000000000)
	}
}
func suffixWords(n int64) string {
	if n == 0 {
		return ""
	}
	return " " + terbilang(n)
}
