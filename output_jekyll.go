package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type jekyllOutputWriter struct {
	Directory string
}

func (w jekyllOutputWriter) Preflight() error {
	if err := os.MkdirAll(w.Directory, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(w.Directory+"/_posts", 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(w.Directory+"/_data", 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(w.Directory+"/_mirror", 0755); err != nil {
		return err
	}
	return nil
}

func (w jekyllOutputWriter) Write(bookmark bookmarkData) error {
	if err := w.writeJSONFile(bookmark); err != nil {
		log.Printf("[%s] error writing JSON: %v", bookmark.GetID(), err)
		return err
	}
	if err := w.writeJekyllPost(bookmark); err != nil {
		log.Printf("[%s] error writing jekyll post: %v", bookmark.GetID(), err)
		return err
	}
	if err := w.writeTextFile(bookmark); err != nil {
		log.Printf("[%s] error writing text: %v", bookmark.GetID(), err)
		return err
	}
	if err := w.writeHighlightsFile(bookmark); err != nil {
		log.Printf("[%s] error writing highlights: %v", bookmark.GetID(), err)
		return err
	}
	return nil
}

func (w jekyllOutputWriter) writeJSONFile(bookmark bookmarkData) error {
	outputFilePath := filepath.Join(w.Directory, "_data", fmt.Sprintf("%s.json", bookmark.GetID()))
	if fileExists(outputFilePath) {
		return nil
	}
	data, err := json.MarshalIndent(bookmark, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(outputFilePath, data, 0644)
}

func (w jekyllOutputWriter) writeJekyllPost(bookmark bookmarkData) error {
	outputFilePath := filepath.Join(w.Directory, "_posts", fmt.Sprintf("%s-%s.html", bookmark.GetYYYYMMDD(), bookmark.GetID()))
	if fileExists(outputFilePath) {
		return nil
	}
	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.WriteString("archive_id: \"" + bookmark.GetID() + "\"\n")
	buf.WriteString("title: \"" + strings.ReplaceAll(bookmark.GetTitle(), `"`, `\"`) + "\"\n")
	buf.WriteString("category: \"" + bookmark.ContainingFolder + "\"\n")
	buf.WriteString("---\n\n")
	if len(bookmark.FullText) > 0 {
		buf.WriteString("{% raw %}\n")
		buf.WriteString(bookmark.FullText)
		buf.WriteString("\n")
		buf.WriteString("{% endraw %}\n")
	}
	return ioutil.WriteFile(outputFilePath, buf.Bytes(), 0644)
}

func (w jekyllOutputWriter) writeTextFile(bookmark bookmarkData) error {
	if len(bookmark.FullText) == 0 {
		return nil
	}

	outputFilePath := filepath.Join(w.Directory, "_mirror", fmt.Sprintf("%s.html", bookmark.GetID()))
	if fileExists(outputFilePath) {
		return nil
	}
	return ioutil.WriteFile(outputFilePath, []byte(bookmark.FullText), 0644)
}

func (w jekyllOutputWriter) writeHighlightsFile(bookmark bookmarkData) error {
	if len(bookmark.Highlights) <= 0 {
		return nil // no highlights
	}

	outputFilePath := filepath.Join(w.Directory, "_data", fmt.Sprintf("%s.highlights.json", bookmark.GetID()))
	if fileExists(outputFilePath) {
		return nil
	}
	data, err := json.MarshalIndent(bookmark.Highlights, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(outputFilePath, data, 0644)
}
