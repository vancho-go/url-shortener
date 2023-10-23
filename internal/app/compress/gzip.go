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

func (g *gzipWriter) Write(p []byte) (int, error) {
	return g.gzipWriter.Write(p)
}

func (g *gzipWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		g.w.Header().Set("Content-Encoding", "gzip")
	}
	g.w.WriteHeader(statusCode)
}

func (g *gzipWriter) Header() http.Header {
	return g.w.Header()
}

func (g *gzipWriter) Close() error {
	return g.gzipWriter.Close()
}

func newGzipWriter(w http.ResponseWriter) *gzipWriter {
	return &gzipWriter{
		w:          w,
		gzipWriter: gzip.NewWriter(w)}
}

type gzipReader struct {
	r          io.ReadCloser
	gzipReader *gzip.Reader
}

func (g *gzipReader) Read(p []byte) (int, error) {
	return g.gzipReader.Read(p)
}

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

func GzipMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		originalWriter := w

		// проверяем, что content-type, который нам нужен
		//contentType := r.Header.Get("Content-Type")
		//if contentType != "application/json" && contentType != "text/html" {
		//	h.ServeHTTP(originalWriter, r)
		//	return
		//}

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
