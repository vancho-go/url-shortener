// Модуль compress выполняет функцию сжатия http.ResponseWriter в gzip.
package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type gzipWriter struct {
	w          http.ResponseWriter
	gzipWriter *gzip.Writer
}

// Write реализация метода Write для gzipWriter.
func (g *gzipWriter) Write(p []byte) (int, error) {
	return g.gzipWriter.Write(p)
}

// WriteHeader реализация метода WriteHeader для gzipWriter.
func (g *gzipWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		g.w.Header().Set("Content-Encoding", "gzip")
	}
	g.w.WriteHeader(statusCode)
}

// Header реализация метода Header для gzipWriter.
func (g *gzipWriter) Header() http.Header {
	return g.w.Header()
}

// Close реализация метода Close для gzipWriter.
func (g *gzipWriter) Close() error {
	return g.gzipWriter.Close()
}

func newGzipWriter(w http.ResponseWriter) *gzipWriter {
	return &gzipWriter{
		w:          w,
		gzipWriter: gzip.NewWriter(w),
	}
}

type gzipReader struct {
	r          io.ReadCloser
	gzipReader *gzip.Reader
}

// Read реализация метода Read для gzipReader.
func (g *gzipReader) Read(p []byte) (int, error) {
	return g.gzipReader.Read(p)
}

// Close реализация метода Close для gzipReader.
func (g *gzipReader) Close() error {
	if err := g.r.Close(); err != nil {
		return err
	}
	return g.gzipReader.Close()
}

func newGzipReader(r io.ReadCloser) (*gzipReader, error) {
	gzReader, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &gzipReader{r: r, gzipReader: gzReader}, nil
}

// GzipMiddleware выполняет роль middleware, которая оборачиваем оригинальный http.ResponseWriter новым с поддержкой сжатия.
func GzipMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		originalWriter := w

		// проверяем, что клиент умеет принимать gzip
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			// оборачиваем оригинальный http.ResponseWriter новым с поддержкой сжатия
			gzWriter := newGzipWriter(w)
			// меняем оригинальный http.ResponseWriter на новый
			originalWriter = gzWriter
			// не забываем отправить клиенту все сжатые данные после завершения middleware
			defer gzWriter.Close()
		}

		// проверяем, что клиент отправил данные в gzip
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			gzReader, err := newGzipReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			//	меняем тело запроса на распакованное
			r.Body = gzReader
			defer gzReader.Close()
		}

		//	отдаем стандартный хэндлер
		h.ServeHTTP(originalWriter, r)
	}
}
