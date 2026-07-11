package main

import (
	"bytes"

	"github.com/ledongthuc/pdf"
)

// extractText pulls raw text out of a PDF file on disk.
func extractText(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var buf bytes.Buffer
	totalPages := r.NumPage()
	for pageIndex := 1; pageIndex <= totalPages; pageIndex++ {
		page := r.Page(pageIndex)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			continue // skip pages that fail to parse rather than aborting the whole doc
		}
		buf.WriteString(text)
		buf.WriteString("\n")
	}
	return buf.String(), nil
}