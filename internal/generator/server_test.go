package generator

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
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

type ociChart struct {
	repoName, name, tag, path string
}

func setupOciServer(charts []ociChart) (mockServer, func()) {
	mock := mockServer{mux: http.NewServeMux()}

	configBytes := []byte(`{"mediaType":"application/vnd.cncf.helm.config.v1+json"}`)
	configHash := sha256.Sum256(configBytes)
	configDigest := "sha256:" + hex.EncodeToString(configHash[:])

	for _, chart := range charts {
		data, err := os.ReadFile(chart.path)
		if err != nil {
			continue // FIXME:
		}

		hash := sha256.Sum256(data)
		layerDigest := "sha256:" + hex.EncodeToString(hash[:])
		layerSize := len(data)
		namespace := fmt.Sprintf("%s/%s", chart.repoName, chart.name)

		mock.mux.HandleFunc(fmt.Sprintf("/v2/%s/manifests/%s", namespace, chart.tag), func(w http.ResponseWriter, r *http.Request) {
			manifest := map[string]interface{}{
				"schemaVersion": 2,
				"mediaType":     "application/vnd.oci.image.manifest.v1+json",
				"config": map[string]interface{}{
					"mediaType": "application/vnd.cncf.helm.config.v1+json",
					"digest":    configDigest,
					"size":      len(configBytes),
				},
				"layers": []map[string]interface{}{
					{
						"mediaType": "application/vnd.cncf.helm.chart.content.v1.tar+gzip",
						"digest":    layerDigest,
						"size":      layerSize,
					},
				},
			}

			w.Header().Set("Content-Type", "application/vnd.oci.image.manifest.v1+json")
			_ = json.NewEncoder(w).Encode(manifest)
		})

		mock.mux.HandleFunc(fmt.Sprintf("/v2/%s/blobs/%s", namespace, layerDigest), func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/vnd.cncf.helm.chart.content.v1.tar+gzip")
			w.Header().Set("Content-Length", fmt.Sprintf("%d", layerSize))
			w.WriteHeader(http.StatusOK)
			if r.Method == http.MethodHead {
				return
			}
			_, _ = w.Write(data)
		})

		mock.mux.HandleFunc(fmt.Sprintf("/v2/%s/blobs/%s", namespace, configDigest), func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(configBytes)))
			w.WriteHeader(http.StatusOK)
			if r.Method == http.MethodHead {
				return
			}
			_, _ = w.Write(configBytes)
		})

		mock.mux.HandleFunc(fmt.Sprintf("/v2/%s/tags/list", namespace), func(w http.ResponseWriter, r *http.Request) {
			token := strings.Trim(r.Header.Get("Authorization"), " ")
			if len(token) == 0 {
				w.WriteHeader(http.StatusUnauthorized)
			} else {
				manifest := map[string]interface{}{
					"name": namespace,
					"tags": []string{chart.tag},
				}
				_ = json.NewEncoder(w).Encode(manifest)
			}
		})

		mock.mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
			manifest := map[string]interface{}{
				"token": "bearer-token",
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(manifest)
		})

		mock.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			log.Printf("unable to serve path from oci server: %s\n", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		})
	}

	server := httptest.NewServer(mock.mux)
	mock.server = server
	mock.URL = strings.ReplaceAll(server.URL, "http://", "oci://") // FIXME: argh?

	return mock, func() {
		server.Close()
	}
}
