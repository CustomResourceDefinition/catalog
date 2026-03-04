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
	"os/exec"
	"path"
	"strings"
	"testing"

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
	tag    string
	branch string
	paths  []gitPath
}

type gitPath struct {
	path, file string
}

func setupGit(t *testing.T, bundles []gitBundle) (*string, error) {
	tmpDir := t.TempDir()

	err := exec.Command("git", "init", tmpDir).Run()
	if err != nil {
		return nil, fmt.Errorf("unable to init: %w", err)
	}

	gitDir := path.Join(tmpDir, ".git")

	err = exec.Command("git", "--git-dir", gitDir, "--work-tree", tmpDir, "config", "user.name", "runner").Run()
	if err != nil {
		return nil, fmt.Errorf("unable to set user name: %w", err)
	}

	err = exec.Command("git", "--git-dir", gitDir, "--work-tree", tmpDir, "config", "user.email", "runner").Run()
	if err != nil {
		return nil, fmt.Errorf("unable to set user email: %w", err)
	}

	err = exec.Command("git", "--git-dir", gitDir, "--work-tree", tmpDir, "commit", "--allow-empty", "--message", "Empty initial commit").Run()
	if err != nil {
		return nil, fmt.Errorf("unable to commit: %w", err)
	}

	bytes, err := exec.Command("git", "--git-dir", gitDir, "--work-tree", tmpDir, "branch", "--show-current").Output()
	if err != nil {
		return nil, fmt.Errorf("unable to look up current branch: %w", err)
	}

	if strings.TrimSuffix(string(bytes), "\n") == "master" {
		err = exec.Command("git", "--git-dir", gitDir, "--work-tree", tmpDir, "branch", "-m", "master", "main").Run()
		if err != nil {
			return nil, fmt.Errorf("unable to rename default branch: %w", err)
		}
	}

	firstBranch := true
	for _, bundle := range bundles {
		ignoreFirstMain := firstBranch && bundle.branch == "main"
		if bundle.branch != "" && !ignoreFirstMain {
			err = exec.Command("git", "--git-dir", gitDir, "--work-tree", tmpDir, "checkout", "-b", bundle.branch).Run()
			if err != nil {
				return nil, fmt.Errorf("unable to create branch %s: %w", bundle.branch, err)
			}
		}

		firstBranch = false

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

			err = exec.Command("git", "--git-dir", gitDir, "--work-tree", tmpDir, "add", p.path).Run()
			if err != nil {
				return nil, fmt.Errorf("unable to add: %w", err)
			}
		}

		if len(bundle.paths) > 0 {
			err = exec.Command("git", "--git-dir", gitDir, "--work-tree", tmpDir, "commit", "--message", "Add bundle").Run()
			if err != nil {
				return nil, fmt.Errorf("unable to commit: %w", err)
			}
		}

		if bundle.tag != "" {
			err = exec.Command("git", "--git-dir", gitDir, "--work-tree", tmpDir, "tag", bundle.tag).Run()
			if err != nil {
				return nil, fmt.Errorf("unable to commit: %w", err)
			}
		}
	}

	return &gitDir, nil
}

type gitHubResponse struct {
	prefix      string
	refsPerPage int
	tags        []githubRef
	branches    []githubRef
}

type githubRef struct {
	name          string
	committedDate string
	taggerDate    string
	targetType    string     // "commit" or "tag"
	nested        *githubRef // for nested tags
}

func buildTarget(ref githubRef) map[string]any {
	target := map[string]any{}
	if ref.targetType == "tag" {
		var innerTarget map[string]any
		if ref.nested != nil {
			innerTarget = buildTarget(*ref.nested)
		} else if ref.committedDate != "" {
			innerTarget = map[string]any{"committedDate": ref.committedDate}
		}
		target["target"] = innerTarget
		if ref.taggerDate != "" {
			target["tagger"] = map[string]any{"date": ref.taggerDate}
		}
	} else {
		target["committedDate"] = ref.committedDate
	}
	return target
}

func setupGitHubServer(t *testing.T, responses []gitHubResponse) func() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)

		_ = body["query"].(string)
		variables := body["variables"].(map[string]any)
		prefix := variables["prefix"].(string)
		cursor := ""
		if c, ok := variables["cursor"].(string); ok {
			cursor = c
		}

		var allRefs []githubRef
		var pageSize int
		for _, resp := range responses {
			if resp.prefix == prefix {
				if prefix == "refs/tags/" {
					allRefs = resp.tags
				} else {
					allRefs = resp.branches
				}
				pageSize = resp.refsPerPage
				break
			}
		}

		if pageSize == 0 {
			pageSize = len(allRefs)
		}

		startIdx := 0
		if cursor != "" {
			fmt.Sscanf(cursor, "page%d", &startIdx)
		}

		endIdx := min(startIdx+pageSize, len(allRefs))

		var nodes []any
		for i := startIdx; i < endIdx; i++ {
			ref := allRefs[i]
			if prefix == "refs/tags/" {
				target := buildTarget(ref)
				nodes = append(nodes, map[string]any{
					"name":   ref.name,
					"target": target,
				})
			} else {
				nodes = append(nodes, map[string]any{
					"name": ref.name,
					"target": map[string]any{
						"committedDate": ref.committedDate,
					},
				})
			}
		}

		hasNextPage := endIdx < len(allRefs)
		endCursor := ""
		if hasNextPage {
			endCursor = fmt.Sprintf("page%d", endIdx)
		}

		response := map[string]any{
			"data": map[string]any{
				"repository": map[string]any{
					"refs": map[string]any{
						"nodes": nodes,
						"pageInfo": map[string]any{
							"hasNextPage": hasNextPage,
							"endCursor":   endCursor,
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))

	originalEndpoint := graphQLEndpoint
	graphQLEndpoint = server.URL

	return func() {
		server.Close()
		graphQLEndpoint = originalEndpoint
	}
}
