package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ochronus/instapaper-go-client/instapaper"
)

var jekyllOutputWriterTestDir = filepath.Join("tmp", "jekyllOutputWriter")

func cleanupTestTmpDir(dir string) {
	if !strings.HasPrefix(dir, filepath.Join("tmp")) {
		log.Printf("dir not in tmp: %q", dir)
		return
	}
	_ = os.RemoveAll(dir)
}

func TestJekyllOutputWriter_Preflight(t *testing.T) {
	w := jekyllOutputWriter{Directory: jekyllOutputWriterTestDir}
	if err := w.Preflight(); err != nil {
		t.Fatalf("preflight failed: %v", err)
	}
}

func TestJekyllOutputWriter_Write(t *testing.T) {
	w := jekyllOutputWriter{Directory: jekyllOutputWriterTestDir}
	if err := w.Preflight(); err != nil {
		t.Fatalf("preflight failed: %v", err)
	}
	defer cleanupTestTmpDir(jekyllOutputWriterTestDir)
	bookmark := bookmarkData{
		Bookmark: &instapaper.Bookmark{
			Hash:              "hash1234",
			Description:       "A description",
			ID:                1234,
			Title:             "Title for the bookmark",
			URL:               "https://example.com/bookmark1234",
			ProgressTimestamp: 5678,
			Time:              1288584076,
			Progress:          0.5,
			Starred:           "starred1234",
		},
		BookmarkExportMeta: &bookmarkExportMeta{
			URL:       "https://example.com/bookmark1234",
			Title:     "Title for the bookmark",
			Selection: "selection1234",
			Folder:    instapaper.FolderIDUnread,
			Timestamp: "1288584076",
			Hash:      "hash1234",
		},
		FullText: "full text\n\nof an article",
		Highlights: []instapaper.Highlight{
			{
				ID:         92841,
				BookmarkID: 1234,
				Text:       "Text of the highlight",
				Note:       "Note for highlight",
				Time:       json.Number(strconv.FormatInt(time.Now().UnixMilli(), 10)),
				Position:   10,
			},
		},
	}
	if err := w.Write(bookmark); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	// Now, verify.
	fileContentsMatch(t, filepath.Join(w.Directory, "_data", "1234.json"), `"Title": "Title for the bookmark",`)
	fileContentsMatch(t, filepath.Join(w.Directory, "_data", "1234.highlights.json"), `"Text": "Text of the highlight",`)
	fileContentsMatch(t, filepath.Join(w.Directory, "_mirror", "1234.html"), "full text\n\nof an article")
	fileContentsMatch(t, filepath.Join(w.Directory, "_posts", "2010-10-31-1234.html"), `archive_id: "1234"`)
	fileContentsMatch(t, filepath.Join(w.Directory, "_posts", "2010-10-31-1234.html"), "{% raw %}\nfull text\n\nof an article")
}
