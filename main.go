package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	url       string
	logFile   string
	logger    *log.Logger
	interval  time.Duration
	silent    bool
	verbose   bool
	threshold float64
	retries   int
)

func main() {
	flag.StringVar(&url, "url", "", "URL to check")
	flag.StringVar(&logFile, "logfile", "healthcheck.log", "file to log output to")
	flag.DurationVar(&interval, "interval", 2*time.Second, "Interval betwween healthchecks")
	flag.BoolVar(&silent, "silent", false, "Run in silent mode without studout output")
	flag.BoolVar(&verbose, "verbose", false, "run in verbose mode . overrides silent mode")
	flag.Float64Var(&threshold, "threshold", 0.5, "threshold value for considering a response to be too slow (in seconds)")
	flag.IntVar(&retries, "retries", 3, "number of retries for a failed request")
	flag.Parse()

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating log file %s: %v", logFile, err)
		log.Fatal(err)
	}
	defer file.Close()

	if silent && !verbose {
		logger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		logger = log.New(io.MultiWriter(file, os.Stdout), "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		checkURL(url, threshold, retries)
	}

}

func checkURL(url string, threshold float64, retries int) {
	var resp *http.Response
	var err error
	var duration time.Duration
	for attempt := 1; attempt <= retries; attempt++ {
		start := time.Now()
		resp, err = http.Get(url)
		duration = time.Since(start)

		if err != nil {
			defer resp.Body.Close()
			break
		}

		if attempt < retries {
			fmt.Fprintf(os.Stderr, "attempt %d failed , retrying...\n", attempt+1)
			time.Sleep(time.Second * 2)
		}
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching %s after %d retries: %v\n", url, retries, err)
		return
	}
	if duration.Seconds() > threshold && verbose {
		fmt.Fprintf(os.Stderr, "Warning: %s response time (%v) exceeded threshold of %fs\n", url, duration, threshold)
	}
	logger.Printf("Checked %s, Status: %d\n", url, resp.StatusCode)
}
