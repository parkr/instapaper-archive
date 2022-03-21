# instapaper-archive

Archive your Instapaper bookmarks. Currently creates a Jekyll site in
addition to placing JSON data in `archive/_data`, and HTML content in
`archive/_mirror` where text data exists.

```text
Usage of ./instapaper-archive:
  -directory string
    	The directory in which to write the archive (default "archive")
  -email string
    	The email address for the login credentials
  -export-csv-file string
    	The path to the instapaper export CSV (default "instapaper-export.csv")
  -password string
    	The password associated with the given email
  -password-file string
    	The file containing the password (defaults to stdin) (default "-")
  -workers int
    	Number of workers (default 10)
```

## Installing

```text
go install github.com/parkr/instapaper-archive@latest
```

## Running

0. Create a dedicated directory – do not use a checkout of this repository.

1. Create a `.env` file:

```text
export INSTAPAPER_CLIENT_ID=...
export INSTAPAPER_CLIENT_SECRET=...
```

2. Download your Instapaper export archive as CSV.

3. Put your Instapaper password somewhere else so it can be fed through via
stdin.

4. Then:

```text
cat instapaper-password | instapaper-archive -email=instapaper-email
```
