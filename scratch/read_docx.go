package main

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func dumpAllXmlTexts(filePath string) error {
	r, err := zip.OpenReader(filePath)
	if err != nil {
		return err
	}
	defer r.Close()

	fmt.Printf("\n==================================================\n")
	fmt.Printf("FULL DOCX EXPLORATION: %s\n", filepath.Base(filePath))
	fmt.Printf("==================================================\n")

	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".xml") {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			decoder := xml.NewDecoder(rc)
			var texts []string
			var inT bool
			for {
				t, err := decoder.Token()
				if err == io.EOF {
					break
				}
				if err != nil {
					break
				}
				switch se := t.(type) {
				case xml.StartElement:
					if se.Name.Local == "t" {
						inT = true
					}
				case xml.CharData:
					if inT {
						str := strings.TrimSpace(string(se))
						if str != "" {
							texts = append(texts, str)
						}
					}
				case xml.EndElement:
					if se.Name.Local == "t" {
						inT = false
					}
				}
			}
			rc.Close()

			if len(texts) > 0 {
				fmt.Printf("\n--- ZIP Entry: %s ---\n", f.Name)
				for _, txt := range texts {
					fmt.Println(txt)
				}
			}
		}
	}
	return nil
}

func main() {
	files, _ := filepath.Glob("Docs/*.docx")
	for _, file := range files {
		dumpAllXmlTexts(file)
	}
}
