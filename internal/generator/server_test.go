package generator

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
)

func setupServer(url, file string) (*httptest.Server, func()) {
	req := serverRequest{
		urlPath: url,
		status:  http.StatusOK,
		file:    file,
	}

	return setupRequests([]serverRequest{req})
}

type serverRequest struct {
	urlPath, file string
	status        int
	response      io.Reader
}

func setupRequests(requests []serverRequest) (*httptest.Server, func()) {
	mux := http.NewServeMux()

	for _, request := range requests {
		mux.HandleFunc(request.urlPath, func(w http.ResponseWriter, r *http.Request) {
			if len(request.file) != 0 {
				b, _ := os.ReadFile(request.file)
				request.response = bytes.NewReader(b)
			}
			w.WriteHeader(request.status)
			var bytes []byte
			request.response.Read(bytes)
			w.Write(bytes)
		})
	}

	server := httptest.NewServer(mux)

	return server, func() {
		server.Close()
	}
}
