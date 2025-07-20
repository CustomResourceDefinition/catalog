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
	"path"
	"strings"
	"testing"

	"github.com/go-git/go-billy/v6/osfs"
	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/go-git/go-git/v6/storage/filesystem"
	"github.com/stretchr/testify/assert"
)

func setupLogger() io.Writer {
	return bytes.NewBuffer([]byte{})
}

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

func setupOciServer(t *testing.T, chart ociChart) (mockServer, func()) {
	mock := mockServer{mux: http.NewServeMux()}

	configBytes := []byte(`{"mediaType":"application/vnd.cncf.helm.config.v1+json"}`)
	configHash := sha256.Sum256(configBytes)
	configDigest := "sha256:" + hex.EncodeToString(configHash[:])

	server := httptest.NewServer(mock.mux)
	mock.server = server
	mock.URL = strings.ReplaceAll(server.URL, "http://", "oci://")

	data, err := os.ReadFile(chart.path)
	if err != nil {
		assert.Fail(t, "unable to read chart at %s", chart.path)
		data = []byte(fmt.Sprintf("unable to read chart at %s", chart.path))
	}

	hash := sha256.Sum256(data)
	layerDigest := "sha256:" + hex.EncodeToString(hash[:])
	layerSize := len(data)
	namespace := fmt.Sprintf("%s/%s", chart.repoName, chart.name)
	token := hex.EncodeToString(hash[:])

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
		bearer := strings.Trim(r.Header.Get("Authorization"), " ")
		parts := strings.Split(bearer, " ")
		if len(parts) != 2 || token != parts[1] {
			w.Header().Set("www-authenticate", fmt.Sprintf(`Bearer realm="%s/token",service="registry",scope="repository:%s:pull",charset="UTF-8"`, server.URL, namespace))
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
		q := r.URL.Query()

		if v, ok := q["service"]; !ok || len(v) == 0 || v[0] != "registry" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if v, ok := q["scope"]; !ok || len(v) == 0 || v[0] != fmt.Sprintf("repository:%s:pull", namespace) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		manifest := map[string]interface{}{
			"token": token,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(manifest)
	})

	mock.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("unable to serve path from oci server: %s\n", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	})

	return mock, func() {
		server.Close()
	}
}

type gitBundle struct {
	tag   string
	paths []gitPath
}

type gitPath struct {
	path, file string
}

func setupGit(bundles []gitBundle) (*string, error) {
	tmpDir, err := os.MkdirTemp("", "generator.config.Name")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	dotgitdir := path.Join(tmpDir, ".git")
	store := filesystem.NewStorage(osfs.New(dotgitdir), nil)

	repo, err := git.Init(store, git.WithDefaultBranch("refs/heads/main"), git.WithWorkTree(osfs.New(tmpDir)))
	if err != nil {
		return nil, err
	}

	cfg, err := repo.Config()
	if err != nil {
		return nil, err
	}

	cfg.User.Name = "runner"
	// set config to create file _required_ by work trees
	err = repo.Storer.SetConfig(cfg)
	if err != nil {
		return nil, err
	}

	tree, err := repo.Worktree()
	if err != nil {
		return nil, err
	}

	_, err = tree.Commit("Empty initial commit", &git.CommitOptions{Author: &object.Signature{Name: "runner", Email: "runner"}, AllowEmptyCommits: true})
	if err != nil {
		return nil, err
	}

	for _, bundle := range bundles {
		for _, p := range bundle.paths {
			directory := path.Join(tmpDir, path.Dir(p.path))

			bytes, err := os.ReadFile(p.file)
			if err != nil {
				return nil, err
			}

			os.MkdirAll(directory, 0775)
			err = os.WriteFile(path.Join(tmpDir, p.path), bytes, 0644)
			if err != nil {
				return nil, err
			}

			_, err = tree.Add(p.path)
			if err != nil {
				return nil, err
			}
		}

		if len(bundle.paths) > 0 {
			commit, err := tree.Commit("Add bundle", &git.CommitOptions{Author: &object.Signature{Name: "runner", Email: "runner"}})
			if err != nil {
				return nil, err
			}

			_, err = repo.CommitObject(commit)
			if err != nil {
				return nil, err
			}

			_, err = repo.CreateTag(bundle.tag, commit, nil)
			if err != nil {
				return nil, err
			}
		} else {
			commit, _ := repo.Head()

			_, err = repo.CreateTag(bundle.tag, commit.Hash(), nil)
			if err != nil {
				return nil, err
			}
		}
	}

	return &dotgitdir, nil
}
