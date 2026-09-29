package engine

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/mail"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func readDocumentMetadata(path string) map[string]string {
	out := map[string]string{}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".pdf":
		readPDFMetadata(path, out)
	case ".docx", ".xlsx", ".pptx":
		readOfficeOpenXMLMetadata(path, out)
	case ".eml":
		readEmailMetadata(path, out)
	case ".epub":
		readEPUBMetadata(path, out)
	}
	return out
}

var pdfPageRE = regexp.MustCompile("/Type\\s*/Page\\b")
var pdfTitleRE = regexp.MustCompile("/Title\\s*\\(([^()]*)\\)")
var pdfAuthorRE = regexp.MustCompile("/Author\\s*\\(([^()]*)\\)")
var pdfCreatorRE = regexp.MustCompile("/Creator\\s*\\(([^()]*)\\)")

func readPDFMetadata(path string, out map[string]string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return
	}
	limit := st.Size()
	if limit > 64*1024*1024 {
		limit = 64 * 1024 * 1024
	}
	data, err := io.ReadAll(io.LimitReader(f, limit))
	if err != nil {
		return
	}
	setMeta(out, "pages", strconv.Itoa(len(pdfPageRE.FindAll(data, -1))))
	if m := pdfTitleRE.FindSubmatch(data); len(m) > 1 {
		setMeta(out, "title", decodePDFLiteral(string(m[1])))
	}
	if m := pdfAuthorRE.FindSubmatch(data); len(m) > 1 {
		setMeta(out, "author", decodePDFLiteral(string(m[1])))
	}
	if m := pdfCreatorRE.FindSubmatch(data); len(m) > 1 {
		setMeta(out, "creator", decodePDFLiteral(string(m[1])))
	}
}

func decodePDFLiteral(s string) string {
	r := strings.NewReplacer("\\(", "(", "\\)", ")", "\\\\", "\\")
	return strings.TrimSpace(r.Replace(s))
}

func readOfficeOpenXMLMetadata(path string, out map[string]string) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return
	}
	defer zr.Close()

	slideCount := 0
	for _, f := range zr.File {
		name := strings.ToLower(f.Name)
		switch name {
		case "docprops/core.xml":
			data := readZipFile(f, 2*1024*1024)
			parseCoreProperties(data, out)
		case "docprops/app.xml":
			data := readZipFile(f, 2*1024*1024)
			parseAppProperties(data, out)
		default:
			if strings.HasPrefix(name, "ppt/slides/slide") && strings.HasSuffix(name, ".xml") && !strings.Contains(name, "_rels") {
				slideCount++
			}
		}
	}
	if out["pages"] == "" && slideCount > 0 {
		setMeta(out, "pages", strconv.Itoa(slideCount))
	}
}

func parseCoreProperties(data []byte, out map[string]string) {
	if len(data) == 0 {
		return
	}
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := dec.Token()
		if err != nil {
			return
		}
		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		name := strings.ToLower(start.Name.Local)
		switch name {
		case "title", "subject", "creator":
			var value string
			if err := dec.DecodeElement(&value, &start); err == nil {
				setMeta(out, name, value)
				if name == "creator" {
					setMeta(out, "author", value)
				}
			}
		case "created":
			var value string
			if err := dec.DecodeElement(&value, &start); err == nil {
				setMeta(out, "date created", value)
			}
		case "modified":
			var value string
			if err := dec.DecodeElement(&value, &start); err == nil {
				setMeta(out, "date modified", value)
			}
		}
	}
}

func parseAppProperties(data []byte, out map[string]string) {
	if len(data) == 0 {
		return
	}
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := dec.Token()
		if err != nil {
			return
		}
		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		name := strings.ToLower(start.Name.Local)
		if name == "pages" || name == "slides" {
			var value string
			if err := dec.DecodeElement(&value, &start); err == nil && strings.TrimSpace(value) != "" {
				setMeta(out, "pages", value)
				return
			}
		}
	}
}

func readZipFile(f *zip.File, max int64) []byte {
	rc, err := f.Open()
	if err != nil {
		return nil
	}
	defer rc.Close()
	data, _ := io.ReadAll(io.LimitReader(rc, max))
	return data
}

func readEmailMetadata(path string, out map[string]string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	msg, err := mail.ReadMessage(bufio.NewReader(f))
	if err != nil {
		return
	}
	h := msg.Header
	setMeta(out, "subject", h.Get("Subject"))
	if dateText := h.Get("Date"); dateText != "" {
		setMeta(out, "date", dateText)
		if t, err := mail.ParseDate(dateText); err == nil {
			setMeta(out, "email date parsed", t.Format(time.RFC3339))
		}
	}
	storeAddressHeader(out, "from", h.Get("From"))
	storeAddressHeader(out, "to", h.Get("To"))
	storeAddressHeader(out, "cc", h.Get("Cc"))
	storeAddressHeader(out, "bcc", h.Get("Bcc"))
}

func storeAddressHeader(out map[string]string, key, raw string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return
	}
	setMeta(out, key, raw)
	addresses, err := mail.ParseAddressList(raw)
	if err != nil {
		return
	}
	var names, emails, full []string
	for i, addr := range addresses {
		names = append(names, addr.Name)
		emails = append(emails, addr.Address)
		fullValue := addr.Address
		if addr.Name != "" {
			fullValue = fmt.Sprintf("%s <%s>", addr.Name, addr.Address)
		}
		full = append(full, fullValue)
		setMeta(out, fmt.Sprintf("%s %d", key, i+1), fullValue)
		setMeta(out, fmt.Sprintf("%sname %d", key, i+1), addr.Name)
		setMeta(out, fmt.Sprintf("%semail %d", key, i+1), addr.Address)
	}
	setMeta(out, key+"name", strings.Join(names, ", "))
	setMeta(out, key+"email", strings.Join(emails, ", "))
	setMeta(out, key, strings.Join(full, ", "))
}

func readEPUBMetadata(path string, out map[string]string) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return
	}
	defer zr.Close()
	for _, f := range zr.File {
		if !strings.HasSuffix(strings.ToLower(f.Name), ".opf") {
			continue
		}
		data := readZipFile(f, 4*1024*1024)
		parseEPUBOPF(data, out)
		if out["title"] != "" || out["creator"] != "" {
			return
		}
	}
}

func parseEPUBOPF(data []byte, out map[string]string) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := dec.Token()
		if err != nil {
			return
		}
		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		name := strings.ToLower(start.Name.Local)
		switch name {
		case "title", "creator", "subject", "date":
			var value string
			if err := dec.DecodeElement(&value, &start); err == nil {
				setMetaIfEmpty(out, name, value)
				if name == "creator" {
					setMetaIfEmpty(out, "author", value)
				}
			}
		}
	}
}
