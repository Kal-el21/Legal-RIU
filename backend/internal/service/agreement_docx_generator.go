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
var wordParagraphRE = regexp.MustCompile(`(?s)<w:p(?:\s[^>]*)?>.*?</w:p>`)
var placeholderRE = regexp.MustCompile(`\{\{[A-Z0-9_]+\}\}`)
var scopePrefixRE = regexp.MustCompile(`^\s*(?:(?:[-*•])|(?:\d+|[A-Za-z])[.)])\s*`)

// scopeListPlaceholder digandakan satu paragraf per baris ruang lingkup, jadi
// penomoran a/b/c-nya harus berasal dari numbered list Word pada paragraf itu.
const scopeListPlaceholder = "{{RUANG_LINGKUP_LIST}}"

// GenerateOptions mengatur perlakuan khusus saat dokumen dibentuk. Template
// diisi hanya lewat placeholder {{...}}; template lama tanpa placeholder ditolak
// sebelum sampai sini (ErrTemplateLama).
//
// Draft untuk pratinjau: setiap halaman diberi watermark DRAFT lewat header, dan
// sorotan (highlight) dari template dibiarkan sebagai penanda kerja. Tanpa Draft
// (dokumen final), sorotan dibuang. Isi dokumen di luar sorotan sama persis,
// sehingga pratinjau dan final berpisah halaman di titik yang sama.
type GenerateOptions struct {
	Draft bool
}

// IsianDokumen: nilai placeholder dan paragraf berulang, dibentuk placeholderValues.
type IsianDokumen struct {
	Nilai map[string]string
	Ulang []ParagrafUlang
}

// ParagrafUlang: setiap paragraf yang memuat salah satu Token diulang sekali per
// Baris, dan token di tiap salinan diganti nilai barisnya. Tanpa baris, paragraf
// tercetak sekali dengan "-".
type ParagrafUlang struct {
	Token []string
	Baris []map[string]string
}

// watermarkDraft adalah watermark Word "DRAFT" dari template asli tim legal
// (VML PowerPlusWaterMarkObject), yang terbukti dirender LibreOffice. Dua %d
// membuat id bentuknya unik per header.
const watermarkDraft = `<w:r><w:rPr><w:rFonts w:eastAsia="Arial"/><w:color w:val="000000"/></w:rPr><w:pict><v:shapetype id="_x0000_t136" coordsize="21600,21600" o:spt="136" adj="10800" path="m@7,l@8,m@5,21600l@6,21600e"><v:formulas><v:f eqn="sum #0 0 10800"/><v:f eqn="prod #0 2 1"/><v:f eqn="sum 21600 0 @1"/><v:f eqn="sum 0 0 @2"/><v:f eqn="sum 21600 0 @3"/><v:f eqn="if @0 @3 0"/><v:f eqn="if @0 21600 @1"/><v:f eqn="if @0 0 @2"/><v:f eqn="if @0 @4 21600"/><v:f eqn="mid @5 @6"/><v:f eqn="mid @8 @5"/><v:f eqn="mid @7 @8"/><v:f eqn="mid @6 @7"/><v:f eqn="sum @6 0 @5"/></v:formulas><v:path textpathok="t" o:connecttype="custom" o:connectlocs="@9,0;@10,10800;@11,21600;@12,10800" o:connectangles="270,180,90,0"/><v:textpath on="t" fitshape="t"/><v:handles><v:h position="#0,bottomRight" xrange="6629,14971"/></v:handles><o:lock v:ext="edit" text="t" shapetype="t"/></v:shapetype><v:shape id="PowerPlusWaterMarkObject%d" o:spid="_x0000_s%d" type="#_x0000_t136" style="position:absolute;margin-left:0;margin-top:0;width:412.4pt;height:247.45pt;rotation:315;z-index:-251658752;visibility:visible;mso-position-horizontal:center;mso-position-horizontal-relative:margin;mso-position-vertical:center;mso-position-vertical-relative:margin" fillcolor="silver" stroked="f"><v:fill opacity=".5"/><v:textpath style="font-family:&quot;Calibri&quot;;font-size:1pt" string="DRAFT"/><w10:wrap anchorx="margin" anchory="margin"/></v:shape></w:pict></w:r>`

// Namespace yang dibutuhkan watermarkDraft di akar header.
var vmlNamespaces = []string{
	`xmlns:v="urn:schemas-microsoft-com:vml"`,
	`xmlns:o="urn:schemas-microsoft-com:office:office"`,
	`xmlns:w10="urn:schemas-microsoft-com:office:word"`,
}

// Header pengganti untuk section tanpa header (mis. template unggahan Admin),
// supaya watermark tetap muncul di setiap halaman pratinjau.
const (
	headerDraftPart = "word/header-alea-draft.xml"
	headerDraftRel  = "rIdAleaDraft"
)

var (
	highlightRE        = regexp.MustCompile(`<w:highlight\b[^>]*/>`)
	sectPrRE           = regexp.MustCompile(`(?s)<w:sectPr\b[^>]*?(?:/>|>.*?</w:sectPr>)`)
	defaultHeaderRefRE = regexp.MustCompile(`<w:headerReference\b[^>]*w:type="default"`)
	firstHeaderRefRE   = regexp.MustCompile(`<w:headerReference\b[^>]*w:type="first"`)
)

func (g *AgreementGenerator) Generate(template []byte, isi IsianDokumen, opt GenerateOptions) ([]byte, string, error) {
	if len(template) == 0 {
		return nil, "", errors.New("template dokumen tidak ditemukan")
	}
	r, err := zip.NewReader(bytes.NewReader(template), int64(len(template)))
	if err != nil {
		return nil, "", errors.New("template bukan DOCX yang valid")
	}
	// Section tanpa header dibuatkan header untuk watermark. Relasi dan tipe isinya
	// ditulis sebelum document.xml di dalam zip, jadi kebutuhannya dicek dulu.
	headerBaru := false
	if opt.Draft {
		for _, f := range r.File {
			if f.Name != "word/document.xml" {
				continue
			}
			src, e := f.Open()
			if e != nil {
				return nil, "", e
			}
			doc, e := io.ReadAll(src)
			src.Close()
			if e != nil {
				return nil, "", e
			}
			headerBaru = lengkapiHeaderDraft(string(doc)) != string(doc)
		}
	}
	nomorHeader := 0

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
			if f.Name == "word/document.xml" {
				text = expandScopeParagraphs(text, isi.Nilai["RUANG_LINGKUP"])
				for _, u := range isi.Ulang {
					text = ulangParagraf(text, u)
				}
			}
			for key, val := range isi.Nilai {
				// Dibatasi sebanyak kemunculannya, jadi isian yang memuat teks
				// placeholder-nya sendiri tidak diganti berulang tanpa henti.
				needle := "{{" + key + "}}"
				for range strings.Count(paragraphText(text), needle) {
					text = replaceAcrossWordRuns(text, needle, wordValue(val))
				}
			}
			if opt.Draft && strings.HasPrefix(f.Name, "word/header") {
				nomorHeader++
				text = pasangWatermark(text, nomorHeader)
			}
			if headerBaru && f.Name == "word/document.xml" {
				text = lengkapiHeaderDraft(text)
			}
			if !opt.Draft {
				text = highlightRE.ReplaceAllString(text, "")
			}
			data = []byte(text)
		}
		if headerBaru {
			switch f.Name {
			case "[Content_Types].xml":
				data = []byte(strings.Replace(string(data), "</Types>", `<Override PartName="/`+headerDraftPart+`" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.header+xml"/></Types>`, 1))
			case "word/_rels/document.xml.rels":
				data = []byte(strings.Replace(string(data), "</Relationships>", `<Relationship Id="`+headerDraftRel+`" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/header" Target="`+strings.TrimPrefix(headerDraftPart, "word/")+`"/></Relationships>`, 1))
			}
		}
		if _, e = dst.Write(data); e != nil {
			return nil, "", e
		}
	}
	if headerBaru {
		dst, e := w.CreateHeader(&zip.FileHeader{Name: headerDraftPart, Method: zip.Deflate})
		if e != nil {
			return nil, "", e
		}
		header := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:hdr xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"></w:hdr>`
		nomorHeader++
		if _, e = dst.Write([]byte(pasangWatermark(header, nomorHeader))); e != nil {
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

// pasangWatermark menaruh watermark DRAFT di paragraf pertama header dan
// melengkapi namespace VML di akarnya.
func pasangWatermark(header string, nomor int) string {
	akar := strings.Index(header, "<w:hdr")
	if akar < 0 {
		return header
	}
	tutup := akar + strings.Index(header[akar:], ">")
	tag := strings.TrimSuffix(header[akar:tutup], "/")
	kosong := strings.HasSuffix(header[akar:tutup], "/")
	for _, ns := range vmlNamespaces {
		if !strings.Contains(tag, ns[:strings.Index(ns, "=")+1]) {
			tag += " " + ns
		}
	}
	run := fmt.Sprintf(watermarkDraft, nomor, 2048+nomor)
	sisa := header[tutup+1:]
	switch {
	case kosong:
		sisa = "<w:p>" + run + "</w:p></w:hdr>" + sisa
	case strings.Contains(sisa, "</w:p>"):
		sisa = strings.Replace(sisa, "</w:p>", run+"</w:p>", 1)
	default:
		sisa = strings.Replace(sisa, "</w:hdr>", "<w:p>"+run+"</w:p></w:hdr>", 1)
	}
	return header[:akar] + tag + ">" + sisa
}

// lengkapiHeaderDraft menambah rujukan ke header pengganti pada section yang
// tidak punya header utama (dan header halaman pertama, kalau halaman pertama
// dibedakan). Section yang sudah punya header tidak diubah.
func lengkapiHeaderDraft(doc string) string {
	return sectPrRE.ReplaceAllStringFunc(doc, func(sect string) string {
		ref := ""
		if !defaultHeaderRefRE.MatchString(sect) {
			ref += `<w:headerReference w:type="default" r:id="` + headerDraftRel + `"/>`
		}
		if strings.Contains(sect, "<w:titlePg") && !firstHeaderRefRE.MatchString(sect) {
			ref += `<w:headerReference w:type="first" r:id="` + headerDraftRel + `"/>`
		}
		if ref == "" {
			return sect
		}
		if strings.HasSuffix(sect, "/>") && !strings.Contains(sect, "</w:sectPr>") {
			return strings.TrimSuffix(sect, "/>") + ">" + ref + "</w:sectPr>"
		}
		i := strings.Index(sect, ">") + 1
		return sect[:i] + ref + sect[i:]
	})
}

func paragraphText(paragraph string) string {
	var b strings.Builder
	for _, match := range wordTextRE.FindAllStringSubmatch(paragraph, -1) {
		b.WriteString(html.UnescapeString(match[2]))
	}
	return b.String()
}

var (
	paraIDRE   = regexp.MustCompile(`\s+w14:(?:paraId|textId)="[^"]*"`)
	bookmarkRE = regexp.MustCompile(`<w:bookmark(?:Start|End)\b[^>]*/>`)
)

// paragrafBerisi: paragraf yang teksnya memuat salah satu token u.
func paragrafBerisi(paragraph string, u ParagrafUlang) bool {
	teks := paragraphText(paragraph)
	for _, t := range u.Token {
		if strings.Contains(teks, "{{"+t+"}}") {
			return true
		}
	}
	return false
}

// ulangParagraf mengulang setiap paragraf yang memuat token u sekali per baris.
// Penggantian tiap token dibatasi sebanyak kemunculannya di paragraf asli, jadi
// isian yang kebetulan berisi teks placeholder tidak diganti berulang tanpa henti.
func ulangParagraf(xmlText string, u ParagrafUlang) string {
	baris := u.Baris
	if len(baris) == 0 {
		kosong := map[string]string{}
		for _, t := range u.Token {
			kosong[t] = "-"
		}
		baris = []map[string]string{kosong}
	}
	var out strings.Builder
	akhir := 0
	for _, loc := range wordParagraphRE.FindAllStringIndex(xmlText, -1) {
		paragraph := xmlText[loc[0]:loc[1]]
		if !paragrafBerisi(paragraph, u) {
			continue
		}
		teks := paragraphText(paragraph)
		// Salinan tidak boleh berbagi id paragraf Word.
		paragraph = paraIDRE.ReplaceAllString(paragraph, "")
		out.WriteString(xmlText[akhir:loc[0]])
		for i, b := range baris {
			salinan := paragraph
			if i > 0 {
				// Id bookmark Word harus unik; bookmark tetap di salinan pertama.
				salinan = bookmarkRE.ReplaceAllString(salinan, "")
			}
			for t, v := range b {
				for range strings.Count(teks, "{{"+t+"}}") {
					salinan = replaceAcrossWordRuns(salinan, "{{"+t+"}}", wordValue(v))
				}
			}
			out.WriteString(salinan)
		}
		akhir = loc[1]
	}
	out.WriteString(xmlText[akhir:])
	return out.String()
}

// expandScopeParagraphs: {{RUANG_LINGKUP_LIST}} versi template sebelum PR 10, satu
// paragraf per baris isian ruang lingkup (textarea).
func expandScopeParagraphs(xmlText, raw string) string {
	const token = "RUANG_LINGKUP_LIST"
	u := ParagrafUlang{Token: []string{token}}
	for _, line := range strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(scopePrefixRE.ReplaceAllString(line, ""))
		if line != "" && line != "-" {
			u.Baris = append(u.Baris, map[string]string{token: line})
		}
	}
	return ulangParagraf(xmlText, u)
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

// InspectTemplate membaca placeholder yang dipakai sebuah template dan
// memeriksa apakah paragraf daftar ruang lingkup sudah diformat sebagai
// numbered list Word. Dipakai saat validasi unggahan template baru.
func InspectTemplate(docx []byte) (placeholders []string, scopeListNumbered bool, err error) {
	r, e := zip.NewReader(bytes.NewReader(docx), int64(len(docx)))
	if e != nil {
		return nil, false, errors.New("file bukan DOCX yang valid")
	}
	set := map[string]bool{}
	scopeListNumbered = true
	for _, f := range r.File {
		if !strings.HasPrefix(f.Name, "word/") || !strings.HasSuffix(f.Name, ".xml") {
			continue
		}
		s, e := f.Open()
		if e != nil {
			return nil, false, e
		}
		b, e := io.ReadAll(s)
		s.Close()
		if e != nil {
			return nil, false, e
		}
		text := string(b)
		for _, p := range placeholderRE.FindAllString(text, -1) {
			set[p] = true
		}
		if f.Name == "word/document.xml" && strings.Contains(text, scopeListPlaceholder) {
			scopeListNumbered = scopeParagraphIsNumbered(text)
		}
	}
	placeholders = make([]string, 0, len(set))
	for p := range set {
		placeholders = append(placeholders, p)
	}
	sort.Strings(placeholders)
	return placeholders, scopeListNumbered, nil
}

var wordTableRE = regexp.MustCompile(`(?s)<w:tbl>.*?</w:tbl>`)

// paragrafKosongTerpanjang menghitung deretan paragraf kosong terpanjang di isi
// dokumen, beserta teks paragraf sesudahnya. Deretan Enter ikut terdorong saat
// isian di atasnya panjang, sehingga muncul ruang atau halaman kosong. Ruang
// kosong di dalam tabel (mis. tempat tanda tangan basah) disengaja dan tidak
// dihitung.
func paragrafKosongTerpanjang(docx []byte) (int, string, error) {
	doc, e := isiDokumenXML(docx)
	if e != nil {
		return 0, "", e
	}
	run, longest, before := 0, 0, ""
	for _, paragraph := range wordParagraphRE.FindAllString(wordTableRE.ReplaceAllString(doc, ""), -1) {
		text := strings.TrimSpace(paragraphText(paragraph))
		if text == "" {
			run++
			continue
		}
		if run > longest {
			longest, before = run, text
		}
		run = 0
	}
	return longest, before, nil
}

// isiDokumenXML membaca word/document.xml.
func isiDokumenXML(docx []byte) (string, error) {
	r, e := zip.NewReader(bytes.NewReader(docx), int64(len(docx)))
	if e != nil {
		return "", errors.New("file bukan DOCX yang valid")
	}
	for _, f := range r.File {
		if f.Name != "word/document.xml" {
			continue
		}
		src, e := f.Open()
		if e != nil {
			return "", e
		}
		b, e := io.ReadAll(src)
		src.Close()
		return string(b), e
	}
	return "", errors.New("isi dokumen (word/document.xml) tidak ditemukan")
}

// periksaParagrafUlang: paragraf field daftar harus numbered list Word supaya
// poinnya bernomor, dan placeholder per baris satu grup harus berada di satu
// paragraf. Paragraf itulah yang diulang; placeholder di paragraf lain akan
// tercetak sebagai blok tersendiri, tidak berselang-seling per baris.
func periksaParagrafUlang(docx []byte, sections []AgreementSectionDefinition) error {
	doc, e := isiDokumenXML(docx)
	if e != nil {
		return e
	}
	paragraphs := wordParagraphRE.FindAllString(doc, -1)
	for _, s := range sections {
		for _, f := range s.Fields {
			if f.Type != "list" && f.Type != "group" {
				continue
			}
			u := paragrafUlang(f, nil)
			kena := 0
			for _, p := range paragraphs {
				if !paragrafBerisi(p, u) {
					continue
				}
				kena++
				if f.Type == "list" && !strings.Contains(p, "<w:numPr>") {
					return fmt.Errorf("paragraf {{%s}} harus diformat sebagai numbered list Word agar poin %s bernomor", u.Token[0], f.Label)
				}
			}
			if f.Type == "group" && kena > 1 {
				return fmt.Errorf("placeholder per baris %s tersebar di %d paragraf; tulis semuanya dalam satu paragraf, yang diulang sekali per baris", f.Label, kena)
			}
		}
	}
	return nil
}

func scopeParagraphIsNumbered(xmlText string) bool {
	for _, loc := range wordParagraphRE.FindAllStringIndex(xmlText, -1) {
		paragraph := xmlText[loc[0]:loc[1]]
		if strings.Contains(paragraphText(paragraph), scopeListPlaceholder) {
			return strings.Contains(paragraph, "<w:numPr>")
		}
	}
	return false
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
