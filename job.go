package main

import (
	"errors"
	"log"
	"os"

	"github.com/ochronus/instapaper-go-client/instapaper"
)

type OutputWriter interface {
	Preflight() error
	Write(bookmarkData) error
}

type InstapaperBookmarkDownloadJob struct {
	APIClient        *instapaper.Client
	BookmarkService  *instapaper.BookmarkService
	HighlightService *instapaper.HighlightService
	Directory        string
	BookmarkData     *bookmarkData
	OutputWriter     OutputWriter
}

func (j *InstapaperBookmarkDownloadJob) Process() error {
	log.Printf("[%s] data: %s", j.BookmarkData.GetID(), j.BookmarkData)
	if j.BookmarkData.Bookmark != nil && j.BookmarkData.Bookmark.ID > 0 {
		// Fill out what we can.
		var err error
		j.BookmarkData.FullText, err = j.BookmarkService.GetText(j.BookmarkData.Bookmark.ID)
		if err != nil {
			log.Printf("[%s] error fetching full text: %v", j.BookmarkData.GetID(), err)
		}
		j.BookmarkData.Highlights, err = j.HighlightService.List(j.BookmarkData.Bookmark.ID)
		if err != nil {
			log.Printf("[%s] error fetching highlights: %v", j.BookmarkData.GetID(), err)
		}

	}
	if err := j.OutputWriter.Write(*j.BookmarkData); err != nil {
		log.Printf("[%s] error writing: %v", j.BookmarkData.GetID(), err)
		return err
	}
	log.Printf("[%s] archived bookmark", j.BookmarkData.GetID())
	return nil
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !errors.Is(err, os.ErrNotExist)
}
