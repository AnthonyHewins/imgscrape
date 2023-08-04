package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/AnthonyHewins/imgscrape/internal/cmdline"
	_ "github.com/lib/pq"
	"github.com/namsral/flag"
)

const appName = "backend"

// CLI vars
var (
	envStr = flag.String("env", "local", "which env: local | dev | stage | prod")

	// logging
	logExporter = flag.String("log-exporter", "", "File to log to. Blank for stdout")
	logLevel    = flag.String("log-level", "INFO", "Log level to use: DEBUG | INFO | WARN | ERROR")
	logFmt      = flag.String("log-format", "json", "Log format to use: json | logfmt")

	// db plaintext config
	dbHost = flag.String("db-host", "localhost", "the database host to connect to. If localhost, sslmode=disable; for any other host, sslmode=require")
	dbPort = flag.Uint("db-port", 5432, "what port to connect to the DB on")
	dbName = flag.String("db-name", "aq", "what database to connect to")

	// timeouts
	httpTimeout    = flag.Duration("http-client-timeout", time.Second*5, "Timeout for HTTP client")
	processTimeout = flag.Duration("process-timeout", time.Hour*3, "Time before the entire process times out")

	// db reader user
	dbReaderUser     = flag.String("db-reader-user", "dbreader", "The database reader username")
	dbReaderPassword = flag.String("db-reader-password", "", "database reader's password")

	// db writer user
	dbWriterUser     = flag.String("db-writer-user", "dbwriter", "The database writer username")
	dbWriterPassword = flag.String("db-writer-password", "", "database writer's password")

	// input
	fileSrc = flag.String("file", "", "File to read from")

	// output
	outDir = flag.String("out-dir", "images", "Directory to place files")
)

func main() {
	app, err := cmdline.NewApp(appName, *logLevel, *logFmt, *logExporter, true)
	if err != nil {
		log.Fatal(err)
	}

	logger := app.Logger()
	ctx, cancel := context.WithTimeout(context.Background(), *processTimeout)
	defer cancel()
	logger.InfoContext(ctx, "Starting process", "times out in", *processTimeout)

	var rows []row
	switch {
	case *fileSrc == "":
		rows, err = csv(ctx, logger, *fileSrc)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	default:
		logger.Error("no source specified; exiting")
		fmt.Println("no source specified; exiting")
		os.Exit(1)
	}

	httpClient := &http.Client{Timeout: *httpTimeout}
	for i, v := range rows {
		l := logger.With("index", i, "row", v)
		if v.ImageURL == "" || v.ID == "" {
			l.WarnContext(ctx, "empty URI/UUID")
			continue
		}

		dir := fmt.Sprintf("%s/%s", *outDir, v.ID)
		info, err := os.Stat(dir)
		switch {
		case err == nil:
			if !info.IsDir() {
				l.InfoContext(ctx, "output already exists")
			} else {
				l.ErrorContext(ctx, "output directory for this ID exists as a file already")
			}

			continue
		case os.IsNotExist(err):
			// doesn't exist; proceed
		case err != nil:
			l.ErrorContext(ctx, "stat error for file trying to save this ID", "err", err)
			continue
		}

		path := v.ImageURL + "/full/full/0/default.jpg"
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
		if err != nil {
			l.ErrorContext(ctx, "failed creating request", "err", err)
			continue
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			l.ErrorContext(ctx, "failed making request", "err", err)
			continue
		}

		if code := resp.StatusCode; code >= 300 || code < 200 {
			l.ErrorContext(ctx, "request failed", "code", code, "resp", resp)
			continue
		}

		buf, err := io.ReadAll(resp.Body)
		if err != nil {
			l.ErrorContext(ctx, "failed reading body", "err", err)
			continue
		}

		err = os.WriteFile(dir+"/image.jpg", buf, 0700)
		if err != nil {
			l.ErrorContext(ctx, "failed writing image")
		}
	}
}

/*
require 'net/http'
require 'json'
require 'csv'

lines.each do |i|
	directory = "images/#{i[0]}"
	if Dir.exists?(directory)
		puts "Directory for ID #{i[0]} already exists, skipping"
		next
	end

	path = i[1]+"/full/full/0/default.jpg"
	puts "Fetching #{path}..."

	begin
		resp = Net::HTTP.get_response URI(path)
	rescue Exception => e
		puts "Failed request to #{path}: #{e}. Skipping"
		next
	end

	if resp.is_a?(Net::HTTPNotFound)
		puts "Failed: #{i[0]}"
		next
	end

	if resp.body.length == 0
		puts "Length of file is 0, skipping"
		next
	end

	puts "Creating dir #{directory}"
	Dir.mkdir(directory)

	meta = directory+"/metadata.json"
	puts "Writing metadata #{meta}"
	File.write(meta,header.zip(i).to_h.to_json)

	img = directory+"/image.jpg"
	puts "Writing image #{img}"
	File.write(img,resp.body)

	sleep 5
end
*/
