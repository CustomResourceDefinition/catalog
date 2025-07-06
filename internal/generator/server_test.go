package generator

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
)

func setupServer(url, file string) (mockServer, func()) {
	req := serverRequest{
		urlPath: url,
		status:  http.StatusOK,
		file:    file,
	}

	return setupRequests([]serverRequest{req})
}

type mockServer struct {
	mux    *http.ServeMux
	server *httptest.Server
	URL    string
}

type serverRequest struct {
	urlPath, file string
	status        int
	response      io.Reader
}

func setupRequests(requests []serverRequest) (mockServer, func()) {
	mock := mockServer{mux: http.NewServeMux()}

	for _, request := range requests {
		mock.addRequest(request)
	}

	server := httptest.NewServer(mock.mux)
	mock.server = server
	mock.URL = server.URL

	return mock, func() {
		server.Close()
	}
}

func (mock mockServer) addRequest(request serverRequest) {
	path := request.urlPath
	if !strings.HasPrefix(path, "/") {
		path = fmt.Sprintf("/%s", path)
	}
	mock.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if len(request.file) != 0 {
			b, _ := os.ReadFile(request.file)
			request.response = bytes.NewReader(b)
		}
		w.WriteHeader(request.status)
		buf := bytes.Buffer{}
		io.Copy(&buf, request.response)
		w.Write(buf.Bytes())
	})
}
