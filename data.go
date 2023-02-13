package main

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"

	"github.com/ochronus/instapaper-go-client/instapaper"
)

type bookmarkData struct {
	Bookmark           *instapaper.Bookmark
	BookmarkExportMeta *bookmarkExportMeta
	Highlights         []instapaper.Highlight `json:"-"`
	FullText           string                 `json:"-"`
	ContainingFolder   string
}

type bookmarkExportMeta struct {
	URL       string
	Title     string
	Selection string
	Folder    string
	Timestamp string
	Hash      string
}

// Returns the Bookmark ID if it has one, otherwise the timestamp in the export meta. Otherwise, "NO_ID"
func (d *bookmarkData) GetID() string {
	if d.Bookmark != nil && d.Bookmark.ID > 0 {
		return strconv.Itoa(d.Bookmark.ID)
	}
	if d.BookmarkExportMeta != nil && d.BookmarkExportMeta.URL != "" {
		return d.GetHash()
	}
	return "NO_ID"
}

func (d *bookmarkData) GetTitle() string {
	if d.Bookmark != nil && d.Bookmark.Title != "" {
		return d.Bookmark.Title
	}
	if d.BookmarkExportMeta != nil && d.BookmarkExportMeta.Title != "" {
		return d.BookmarkExportMeta.Title
	}
	return "NO_HASH"
}

func (d *bookmarkData) GetHash() string {
	if d.Bookmark != nil && d.Bookmark.Hash != "" {
		return d.Bookmark.Hash
	}
	if d.BookmarkExportMeta != nil && d.BookmarkExportMeta.URL != "" {
		if d.BookmarkExportMeta.Hash == "" {
			hash := sha256.Sum256([]byte(d.BookmarkExportMeta.URL))
			encoded := hex.EncodeToString(hash[:])
			d.BookmarkExportMeta.Hash = "sha-" + string(encoded[0:10])
		}
		return d.BookmarkExportMeta.Hash
	}
	return "NO_HASH"
}

func (d bookmarkData) GetURL() string {
	if d.Bookmark != nil && d.Bookmark.URL != "" {
		return d.Bookmark.URL
	}
	if d.BookmarkExportMeta != nil && d.BookmarkExportMeta.URL != "" {
		return d.BookmarkExportMeta.URL
	}
	return "NO_URL"
}

func (d bookmarkData) GetYYYYMMDD() string {
	if d.Bookmark != nil && d.Bookmark.Time > 0 {
		return time.Unix(int64(d.Bookmark.Time), 0).Format("2006-01-02")
	}
	if d.BookmarkExportMeta != nil && d.BookmarkExportMeta.Timestamp != "" {
		unix, err := strconv.Atoi(d.BookmarkExportMeta.Timestamp)
		if err != nil {
			return "2001-10-10" // error, Instapaper didn't exist then
		}
		return time.Unix(int64(unix), 0).Format("2006-01-02")
	}
	return "2000-01-01" // error, Instapaper didn't exist then
}

func (d bookmarkData) String() string {
	return "{ID:" + d.GetID() + ", URL:" + d.GetURL() + "}"
}
