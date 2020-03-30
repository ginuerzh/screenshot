package screenshot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// Screenshoter is a webpage screenshot interface.
type Screenshoter interface {
	Screenshot(ctx context.Context, url string, options ...ScreenshotOption) (io.Reader, error)
}

type chromeRemoteScreenshoter struct {
	url string
}

// NewChromeRemoteScreenshoter creates a Screenshoter backed by Chrome DevTools Protocol.
// The addr is the headless chrome websocket debugger endpoint, such as 127.0.0.1:9222.
func NewChromeRemoteScreenshoter(addr string) (Screenshoter, error) {
	// Due to issue#505 (https://github.com/chromedp/chromedp/issues/505),
	// chrome restricts the host must be IP or localhost, we should rewrite the url.
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/json/version", addr), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Host", "localhost")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &chromeRemoteScreenshoter{
		url: strings.Replace(result["webSocketDebuggerUrl"].(string), "localhost", addr, 1),
	}, nil
}

func (s *chromeRemoteScreenshoter) Screenshot(ctx context.Context, url string, options ...ScreenshotOption) (io.Reader, error) {
	allocatorCtx, cancel := chromedp.NewRemoteAllocator(ctx, s.url)
	defer cancel()

	ctxt, cancel := chromedp.NewContext(allocatorCtx)
	defer cancel()

	var opts ScreenshotOptions
	for _, o := range options {
		o(&opts)
	}

	var buf []byte
	err := chromedp.Run(ctxt,
		emulation.SetDeviceMetricsOverride(opts.Width, opts.Height, opts.ScaleFactor, opts.Mobile),
		chromedp.Navigate(url),
		s.captureAction(&buf, opts.Format, opts.Quality),
		s.closePageAction(),
	)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(buf), nil
}

func (s *chromeRemoteScreenshoter) captureAction(res *[]byte, format string, quality int64) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) (err error) {
		if res == nil {
			return
		}

		params := page.CaptureScreenshot()
		switch format {
		case "jpg", "jpeg":
			params.Format = page.CaptureScreenshotFormatJpeg
		default:
			params.Format = page.CaptureScreenshotFormatPng
		}
		params.Quality = quality
		*res, err = params.Do(ctx)
		return
	})
}

func (s *chromeRemoteScreenshoter) closePageAction() chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) (err error) {
		return page.Close().Do(ctx)
	})
}

// ScreenshotOptions is the options used by Screenshot.
type ScreenshotOptions struct {
	Width       int64
	Height      int64
	ScaleFactor float64
	Mobile      bool
	Format      string // png, jpg, default png.
	Quality     int64  // jpeg only
}

type ScreenshotOption func(*ScreenshotOptions)

func WidthScreenshotOption(width int64) ScreenshotOption {
	return func(opts *ScreenshotOptions) {
		opts.Width = width
	}
}

func HeightScreenshotOption(height int64) ScreenshotOption {
	return func(opts *ScreenshotOptions) {
		opts.Height = height
	}
}

func ScaleFactorScreenshotOption(factor float64) ScreenshotOption {
	return func(opts *ScreenshotOptions) {
		opts.ScaleFactor = factor
	}
}

func MobileScreenshotOption(b bool) ScreenshotOption {
	return func(opts *ScreenshotOptions) {
		opts.Mobile = b
	}
}

func FormatScreenshotOption(format string) ScreenshotOption {
	return func(opts *ScreenshotOptions) {
		opts.Format = format
	}
}

func QualityScreenshotOption(quality int64) ScreenshotOption {
	return func(opts *ScreenshotOptions) {
		opts.Quality = quality
	}
}
