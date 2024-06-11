package main

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/dslipak/pdf"
)

func TestDslipakPDF(t *testing.T) {
	content, err := readPdf("./cerita-singkat.pdf") // Read local pdf file
	if err != nil {
		t.Error(err)
	}

	fmt.Println(content)
}

func readPdf(path string) (string, error) {
	r, err := pdf.Open(path)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		return "", err
	}
	buf.ReadFrom(b)
	return buf.String(), nil
}