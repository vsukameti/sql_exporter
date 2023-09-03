package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/burningalchemist/sql_exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"k8s.io/klog/v2"
)

const (
	contentTypeHeader     string = "Content-Type"
	contentLengthHeader   string = "Content-Length"
	contentEncodingHeader string = "Content-Encoding"
	acceptEncodingHeader  string = "Accept-Encoding"
)

// ExporterHandlerFor returns an http.Handler for the provided Exporter.
func ExporterHandlerFor(exporter sql_exporter.Exporter) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx, cancel := contextFor(req, exporter)
		defer cancel()

		// Go through prometheus.Gatherers to sanitize and sort metrics.
		gatherer := prometheus.Gatherers{exporter.WithContext(ctx), sql_exporter.SvcRegistry}
		mfs, err := gatherer.Gather()
		if err != nil {
			klog.Errorf("Error gathering metrics: %s", err)
			if errors.Is(err, context.DeadlineExceeded) {
				http.Error(w, "Timeout while collecting metrics", http.StatusRequestTimeout)
				return
			}
			if len(mfs) == 0 {
				http.Error(w, "No metrics gathered, "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

		contentType := expfmt.Negotiate(req.Header)
		buf := getBuf()
		defer giveBuf(buf)
		writer, encoding := decorateWriter(req, buf)
		enc := expfmt.NewEncoder(writer, contentType)
		var errs prometheus.MultiError
		for _, mf := range mfs {
			if err := enc.Encode(mf); err != nil {
				errs = append(errs, err)
				klog.Errorf("Error encoding metric family %q: %s", mf.GetName(), err)
			}
		}
		if closer, ok := writer.(io.Closer); ok {
			closer.Close()
		}
		if errs.MaybeUnwrap() != nil && buf.Len() == 0 {
			http.Error(w, "No metrics encoded, "+errs.Error(), http.StatusInternalServerError)
			return
		}
		header := w.Header()
		header.Set(contentTypeHeader, string(contentType))
		header.Set(contentLengthHeader, fmt.Sprint(buf.Len()))
		if encoding != "" {
			header.Set(contentEncodingHeader, encoding)
		}
		_, _ = w.Write(buf.Bytes())
	})
}

func contextFor(req *http.Request, exporter sql_exporter.Exporter) (context.Context, context.CancelFunc) {
	timeout := time.Duration(0)
	configTimeout := time.Duration(exporter.Config().Globals.ScrapeTimeout)
	// If a timeout is provided in the Prometheus header, use it.
	if v := req.Header.Get("X-Prometheus-Scrape-Timeout-Seconds"); v != "" {
		timeoutSeconds, err := strconv.ParseFloat(v, 64)
		if err != nil {
			parseError := errors.Unwrap(err)
			switch {
			case errors.Is(parseError, strconv.ErrSyntax):
				{
					klog.Errorf("Failed to parse timeout from Prometheus header: unsupported value")
				}
			case errors.Is(parseError, strconv.ErrRange):
				{
					klog.Errorf("Failed to parse timeout from Prometheus header: value is out of range")
				}
			}
		} else {
			timeout = time.Duration(timeoutSeconds * float64(time.Second))

			// Subtract the timeout offset, unless the result would be negative or zero.
			timeoutOffset := time.Duration(exporter.Config().Globals.TimeoutOffset)
			if timeoutOffset > timeout {
				klog.Errorf("global.scrape_timeout_offset (`%s`) is greater than Prometheus' scraping timeout (`%s`), ignoring",
					timeoutOffset, timeout)
			} else {
				timeout -= timeoutOffset
			}
		}
	}

	// If the configured scrape timeout is more restrictive, use that instead.
	if configTimeout > 0 && (timeout <= 0 || configTimeout < timeout) {
		timeout = configTimeout
	}

	if timeout <= 0 {
		return context.Background(), func() {}
	}
	return context.WithTimeout(context.Background(), timeout)
}
