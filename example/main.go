package main

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/ginuerzh/screenshot"
)

func main() {
	sc, err := screenshot.NewChromeRemoteScreenshoter("127.0.0.1:9222")
	if err != nil {
		log.Fatal(err)
	}

	format := "png"
	r, err := sc.Screenshot(
		context.Background(),
		"https://bing.com",
		screenshot.WidthScreenshotOption(1080),
		screenshot.HeightScreenshotOption(1920),
		screenshot.MobileScreenshotOption(true),
		screenshot.ScaleFactorScreenshotOption(1),
		screenshot.FormatScreenshotOption(format),
		screenshot.QualityScreenshotOption(100),
	)
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.Create("chrome-screenshot." + format)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	io.Copy(f, r)
}
