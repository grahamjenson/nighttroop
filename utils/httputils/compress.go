package httputils

import (
	"bufio"
	"bytes"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

// compressLevel is gzip's fastest level. Each gzip writer holds about 0.8 MB, where level 5's holds 1 MB, and a day of
// readings is only about 6% larger.
const compressLevel = 1

// compressMinSize is the smallest response that is gzipped. A gzip writer holds its memory however short the response
// is, and gzip makes a reply this short longer, not shorter, so the small replies most phones and followers get are
// sent as they are.
const compressMinSize = 1024

// compress gzips text responses of compressMinSize bytes or more, like the pages, styles, scripts and JSON, for clients
// that accept it
func compress(next http.Handler) http.Handler {
	return middleware.Compress(compressLevel)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		small := &smallWriter{compressing: w, status: http.StatusOK}
		// Deferred, so a response written before a later panic, as when recording that its identity was seen, is sent
		defer small.finish()
		next.ServeHTTP(small, r)
	}))
}

// smallWriter holds a response until it is compressMinSize bytes. One that ends shorter is written past the
// compression, as it is. One that is longer, or flushed, or hijacked, goes through the compression.
type smallWriter struct {
	compressing http.ResponseWriter
	held        bytes.Buffer
	status      int
	wroteHeader bool
	passed      bool
}

func (s *smallWriter) Header() http.Header {
	return s.compressing.Header()
}

func (s *smallWriter) WriteHeader(status int) {
	// An informational status, like a 103, is sent at once, before the response's own
	if s.passed || status < http.StatusOK {
		s.compressing.WriteHeader(status)
		return
	}
	if !s.wroteHeader {
		s.wroteHeader, s.status = true, status
	}
}

func (s *smallWriter) Write(p []byte) (int, error) {
	if !s.passed && s.held.Len()+len(p) >= compressMinSize {
		if err := s.pass(); err != nil {
			return 0, err
		}
	}
	if s.passed {
		return s.compressing.Write(p)
	}
	s.wroteHeader = true
	return s.held.Write(p)
}

// pass sends what is held, and all after it, through the compression
func (s *smallWriter) pass() error {
	if s.passed {
		return nil
	}
	s.passed = true
	if s.wroteHeader {
		s.compressing.WriteHeader(s.status)
	}
	if s.held.Len() == 0 {
		return nil
	}
	_, err := s.compressing.Write(s.held.Bytes())
	s.held.Reset()
	return err
}

// FlushError streams the response, like a live update, so it goes through the compression
func (s *smallWriter) FlushError() error {
	if err := s.pass(); err != nil {
		return err
	}
	return http.NewResponseController(s.compressing).Flush()
}

func (s *smallWriter) Flush() {
	_ = s.FlushError()
}

// Hijack is for a WebSocket, which writes to the connection itself, so nothing is written for it after
func (s *smallWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	s.passed = true
	return http.NewResponseController(s.compressing).Hijack()
}

func (s *smallWriter) Unwrap() http.ResponseWriter {
	return s.compressing
}

// finish writes a response that ended shorter than compressMinSize as it is, past the compression
func (s *smallWriter) finish() {
	if s.passed {
		return
	}
	plain := s.compressing
	if unwrapper, ok := plain.(interface{ Unwrap() http.ResponseWriter }); ok {
		plain = unwrapper.Unwrap()
	}
	if s.wroteHeader {
		plain.WriteHeader(s.status)
	}
	if s.held.Len() > 0 {
		_, _ = plain.Write(s.held.Bytes())
	}
}
