package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/ochronus/instapaper-go-client/instapaper"
)

func fatal(format string, args ...interface{}) {
	fmt.Printf("fatal: "+format+"\n", args...)
	os.Exit(1)
}

// readPassword reads the password from stdin.
func readPassword(passwordFile string) (string, error) {
	var err error
	var f io.Reader = os.Stdin
	if passwordFile != "-" {
		f, err = os.Open(passwordFile)
		if err != nil {
			return "", fmt.Errorf("unable to read password file %q: %v", passwordFile, err)
		}
	}
	data, err := ioutil.ReadAll(f)
	return strings.TrimSpace(string(data)), err
}

func newInstapaperClient(emailAddress, password string) (*instapaper.Client, error) {
	apiClient, err := instapaper.NewClient(
		os.Getenv("INSTAPAPER_CLIENT_ID"),
		os.Getenv("INSTAPAPER_CLIENT_SECRET"),
		emailAddress,
		password,
	)
	if err != nil {
		return nil, fmt.Errorf("error initializing the client: %v", err)
	}
	authErr := apiClient.Authenticate()
	if authErr != nil {
		return nil, fmt.Errorf("error authenticating: %v", authErr)
	}
	return &apiClient, nil
}

func createInstapaperArchive(client instapaper.Client, directory string, exportCSVFileName string, outputWriter OutputWriter, queue *JobQueue) error {
	// 0. Create directories
	if err := outputWriter.Preflight(); err != nil {
		return err
	}

	bookmarkService := instapaper.BookmarkService{Client: client}
	highlightService := instapaper.HighlightService{Client: client}
	folderService := instapaper.FolderService{Client: client}

	// 1. Read in instapaper-export.csv, which has URLs but no IDs
	// Download from https://www.instapaper.com/user -> Download .CSV file
	allBookmarks, err := readBookmarksFromCSVExport(exportCSVFileName)
	if err != nil {
		return err
	}

	// 2. List folders and get all the bookmarks we can from them.
	// I can't get pagination to work with the 'have' parameter.
	folders, err := folderService.List()
	if err != nil {
		return err
	}
	folders = append(folders,
		instapaper.Folder{ID: instapaper.FolderIDUnread, Title: "Unread", Slug: "unread"},
		instapaper.Folder{ID: instapaper.FolderIDStarred, Title: "Starred", Slug: "starred"},
		instapaper.Folder{ID: instapaper.FolderIDArchive, Title: "Archive", Slug: "archive"},
	)
	err = listBookmarksFromFolders(bookmarkService, folders, allBookmarks)
	if err != nil {
		return err
	}

	// 2. Enqueue bookmarks to be archived.
	log.Printf("Bookmarks to archive: %d", len(allBookmarks))
	for _, bookmarkDatum := range allBookmarks {
		queue.Submit(&InstapaperBookmarkDownloadJob{
			BookmarkData:     bookmarkDatum,
			Directory:        directory,
			APIClient:        &client,
			BookmarkService:  &bookmarkService,
			HighlightService: &highlightService,
			OutputWriter:     outputWriter,
		})
	}

	return nil
}

func listBookmarksFromFolders(bookmarkService instapaper.BookmarkService, folders []instapaper.Folder, bookmarks map[string]*bookmarkData) error {
	for _, folder := range folders {
		resp, err := bookmarkService.List(instapaper.BookmarkListRequestParams{
			// this is limited to 500 by the API, and pagination doesn't work,
			// so only the latest 500 bookmarks are returned.
			Limit:  100000,
			Folder: folder.ID.String(),
		})
		if err != nil {
			return err
		}
		for _, bookmark := range resp.Bookmarks {
			bookmark := bookmark
			data, ok := bookmarks[bookmark.URL]
			if ok {
				data.Bookmark = &bookmark
			} else {
				bookmarks[bookmark.URL] = &bookmarkData{Bookmark: &bookmark}
			}
		}
	}
	return nil
}

func readBookmarksFromCSVExport(exportCSVFileName string) (map[string]*bookmarkData, error) {
	csvFile, err := os.Open(exportCSVFileName)
	if err != nil {
		return nil, err
	}
	records, err := csv.NewReader(csvFile).ReadAll()
	if err != nil {
		return nil, err
	}
	bookmarks := map[string]*bookmarkData{}
	for i := range records[1:] {
		row := records[i]
		url := row[0]
		if url == "URL" {
			// CSV title row
			continue
		}
		bookmarks[url] = &bookmarkData{
			BookmarkExportMeta: &bookmarkExportMeta{
				// URL,Title,Selection,Folder,Timestamp
				URL:       url,
				Title:     row[1],
				Selection: row[2],
				Folder:    row[3],
				Timestamp: row[4],
			},
		}
	}
	return bookmarks, nil
}

func main() {
	var emailAddress string
	flag.StringVar(&emailAddress, "email", "", "The email address for the login credentials")
	var passwordFile string
	flag.StringVar(&passwordFile, "password-file", "-", "The file containing the password (defaults to stdin)")
	var password string
	flag.StringVar(&password, "password", "", "The password associated with the given email")
	var directory string
	flag.StringVar(&directory, "directory", "archive", "The directory in which to write the archive")
	var exportCSVFileName string
	flag.StringVar(&exportCSVFileName, "export-csv-file", "instapaper-export.csv", "The path to the instapaper export CSV")
	var numWorkers int
	flag.IntVar(&numWorkers, "workers", 10, "Number of workers")
	var outputFormat string
	flag.StringVar(&outputFormat, "format", "jekyll", "Archive format")
	flag.Parse()

	if password == "" {
		var err error
		password, err = readPassword(passwordFile)
		if err != nil {
			fatal("error reading password: %v", err)
		}
	}
	if len(password) == 0 {
		fatal("must supply password from stdin, via -password flag, or via -password-file flag")
	}

	apiClient, err := newInstapaperClient(emailAddress, password)
	if err != nil {
		fatal("error creating instapaper client: %v", err)
	}

	var outputWriter OutputWriter
	switch strings.ToLower(outputFormat) {
	case "jekyll":
		outputWriter = jekyllOutputWriter{Directory: directory}
	default:
		log.Fatalf("unsupported output format: %q", outputFormat)
	}

	queue := NewJobQueue(runtime.NumCPU())
	queue.Start()
	defer queue.Stop()

	err = createInstapaperArchive(*apiClient, directory, exportCSVFileName, outputWriter, queue)
	if err != nil {
		fatal("error creating instapaper archive: %v", err)
	}
}
