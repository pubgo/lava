package chi

import (
	"github.com/pubgo/lug/logger"

	"net/http"
)

var _ http.ResponseWriter = (*response)(nil)

type response struct {
	w     http.ResponseWriter
	code  int
	bytes []byte
}

func (r *response) Header() http.Header {
	return r.w.Header()
}

func (r *response) Write(bytes []byte) (int, error) {
	r.bytes = bytes
	return len(bytes), nil
}

func (r *response) WriteHeader(statusCode int) {
	r.code = statusCode
}

func (r *response) do() {
	r.w.WriteHeader(r.code)
	_, err := r.w.Write(r.bytes)
	logger.ErrLog(err)
}
