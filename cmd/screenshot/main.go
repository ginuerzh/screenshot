package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/ginuerzh/screenshot"
)

var (
	addr             string
	chromeRemoteAddr string
)

func init() {
	flag.StringVar(&addr, "l", ":8080", "server address")
	flag.StringVar(&chromeRemoteAddr, "chrome_remote_addr", "127.0.0.1:9222", "chrome websocket debugger endpoint address")
	flag.Parse()
}

func main() {
	http.HandleFunc("/screenshot", screenshotHandler)
	http.HandleFunc("/health", healthHandler)
	fmt.Println("listen on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	return
}

func screenshotHandler(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue("url")
	width, _ := strconv.ParseInt(r.FormValue("width"), 10, 64)
	height, _ := strconv.ParseInt(r.FormValue("height"), 10, 64)
	mobile, _ := strconv.ParseBool(r.FormValue("mobile"))
	format := r.FormValue("format")
	quality, _ := strconv.ParseInt(r.FormValue("quality"), 10, 64)
	timeout, _ := time.ParseDuration(r.FormValue("timeout"))

	if timeout <= time.Second {
		timeout = 30 * time.Second
	}

	defer func(start time.Time) {
		fmt.Printf("%s %s url: %s, width: %d, height: %d, mobile: %v, timeout: %v, duration: %v",
			start.Format("2006-01-02T15:04:05.999"), r.RemoteAddr,
			url, width, height, mobile, timeout, time.Since(start))
	}(time.Now())

	if url == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println("url is required")
		return
	}

	s, err := screenshot.NewChromeRemoteScreenshoter(chromeRemoteAddr)
	if err != nil {
		fmt.Println("screenshot:", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	rd, err := s.Screenshot(ctx, url,
		screenshot.WidthScreenshotOption(width),
		screenshot.HeightScreenshotOption(height),
		screenshot.MobileScreenshotOption(mobile),
		screenshot.ScaleFactorScreenshotOption(1),
		screenshot.FormatScreenshotOption(format),
		screenshot.QualityScreenshotOption(quality),
	)
	if err != nil {
		fmt.Println("screenshot:", err)
		if err == context.DeadlineExceeded {
			http.Error(w, err.Error(), http.StatusRequestTimeout)
			return
		}
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	io.Copy(w, rd)
}
