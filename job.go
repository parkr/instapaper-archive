package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ochronus/instapaper-go-client/instapaper"
)

type InstapaperBookmarkDownloadJob struct {
	APIClient        *instapaper.Client
	BookmarkService  *instapaper.BookmarkService
	HighlightService *instapaper.HighlightService
	Directory        string
	BookmarkData     *bookmarkData
}

func (j *InstapaperBookmarkDownloadJob) Process() {
	log.Printf("[%s] data: %s", j.BookmarkData.GetID(), j.BookmarkData)
	var text string
	if j.BookmarkData.Bookmark != nil && j.BookmarkData.Bookmark.ID > 0 {
		text, _ = j.BookmarkService.GetText(j.BookmarkData.Bookmark.ID)
	}
	if err := j.writeJSONFile(); err != nil {
		log.Printf("[%s] error writing JSON: %v", j.BookmarkData.GetID(), err)
	}
	if err := j.writeJekyllPost(text); err != nil {
		log.Printf("[%s] error writing jekyll post: %v", j.BookmarkData.GetID(), err)
	}
	if err := j.writeTextFile(text); err != nil {
		log.Printf("[%s] error writing text: %v", j.BookmarkData.GetID(), err)
	}
	if err := j.writeHighlightsFile(); err != nil {
		log.Printf("[%s] error writing highlights: %v", j.BookmarkData.GetID(), err)
	}
	log.Printf("[%s] archived bookmark", j.BookmarkData.GetID())
}

func (j *InstapaperBookmarkDownloadJob) writeJSONFile() error {
	outputFilePath := filepath.Join(j.Directory, "_data", fmt.Sprintf("%s.json", j.BookmarkData.GetID()))
	if fileExists(outputFilePath) {
		return nil
	}
	data, err := json.MarshalIndent(j.BookmarkData, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(outputFilePath, data, 0644)
}

func (j *InstapaperBookmarkDownloadJob) writeJekyllPost(text string) error {
	outputFilePath := filepath.Join(j.Directory, "_posts", fmt.Sprintf("%s-%s.html", j.BookmarkData.GetYYYYMMDD(), j.BookmarkData.GetID()))
	if fileExists(outputFilePath) {
		return nil
	}
	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.WriteString("archive_id: \"" + j.BookmarkData.GetID() + "\"\n")
	buf.WriteString("title: \"" + strings.ReplaceAll(j.BookmarkData.GetTitle(), `"`, `\"`) + "\"\n")
	buf.WriteString("---\n\n")
	if len(text) > 0 {
		buf.WriteString("{% raw %}\n")
		buf.WriteString(text)
		buf.WriteString("\n")
		buf.WriteString("{% endraw %}\n")
	}
	return ioutil.WriteFile(outputFilePath, buf.Bytes(), 0644)
}

func (j *InstapaperBookmarkDownloadJob) writeTextFile(text string) error {
	if len(text) == 0 {
		return nil
	}

	outputFilePath := filepath.Join(j.Directory, "_mirror", fmt.Sprintf("%s.html", j.BookmarkData.GetID()))
	if fileExists(outputFilePath) {
		return nil
	}
	return ioutil.WriteFile(outputFilePath, []byte(text), 0644)
}

func (j *InstapaperBookmarkDownloadJob) writeHighlightsFile() error {
	if j.BookmarkData.Bookmark == nil || j.BookmarkData.Bookmark.ID == 0 {
		return fmt.Errorf("no real bookmark ID")
	}

	outputFilePath := filepath.Join(j.Directory, "_data", fmt.Sprintf("%s.highlights.json", j.BookmarkData.GetID()))
	if fileExists(outputFilePath) {
		return nil
	}
	highlights, err := j.HighlightService.List(j.BookmarkData.Bookmark.ID)
	if err != nil {
		return err
	}
	if len(highlights) == 0 {
		return nil // no highlights, no file
	}
	data, err := json.MarshalIndent(highlights, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(outputFilePath, data, 0644)
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !errors.Is(err, os.ErrNotExist)
}
