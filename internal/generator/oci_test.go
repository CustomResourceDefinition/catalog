package generator

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOciGeneratorVersions(t *testing.T) {
	chartPath := "testdata/connect-2.0.0.tgz"

	reg, err := NewMockOCIRegistry(chartPath, "helm/connect", "2.0.0")
	if err != nil {
		t.Fatal(err)
	}

	err = reg.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer reg.Stop()

	out, err := pullOCIChart("oci://"+reg.Addr+"/helm/connect", "2.0.0")
	assert.Nil(t, err)
	assert.NotNil(t, out)
}

type MockOCIRegistry struct {
	Addr        string
	listener    net.Listener
	server      *http.Server
	chartData   []byte
	layerSize   int
	layerDigest string
	namespace   string
	tag         string
	mu          sync.Mutex
}

func NewMockOCIRegistry(chartPath, namespace, tag string) (*MockOCIRegistry, error) {
	data, err := os.ReadFile(chartPath)
	if err != nil {
		return nil, fmt.Errorf("read chart: %w", err)
	}

	hash := sha256.Sum256(data)
	layerDigest := "sha256:" + hex.EncodeToString(hash[:])

	return &MockOCIRegistry{
		chartData:   data,
		layerSize:   len(data),
		layerDigest: layerDigest,
		namespace:   namespace,
		tag:         tag,
	}, nil
}

func (m *MockOCIRegistry) Start() error {
	mux := http.NewServeMux()

	configBytes := []byte(`{"mediaType":"application/vnd.cncf.helm.config.v1+json"}`)
	configHash := sha256.Sum256(configBytes)
	configDigest := "sha256:" + hex.EncodeToString(configHash[:])

	// manifest endpoint
	manifestPath := fmt.Sprintf("/v2/%s/manifests/%s", m.namespace, m.tag)
	mux.HandleFunc(manifestPath, func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[Registry] %s %s", r.Method, r.URL.Path)

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
					"digest":    m.layerDigest,
					"size":      m.layerSize,
				},
			},
		}

		w.Header().Set("Content-Type", "application/vnd.oci.image.manifest.v1+json")
		_ = json.NewEncoder(w).Encode(manifest)
	})

	// blob endpoint
	blobPath := fmt.Sprintf("/v2/%s/blobs/%s", m.namespace, m.layerDigest)
	mux.HandleFunc(blobPath, func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[Registry] %s %s", r.Method, r.URL.Path)
		w.Header().Set("Content-Type", "application/vnd.cncf.helm.chart.content.v1.tar+gzip")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", m.layerSize))
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusOK)
			return
		}
		_, _ = w.Write(m.chartData)
	})

	mux.HandleFunc("/v2/helm/connect/blobs/"+configDigest, func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[Registry] %s %s", r.Method, r.URL.Path)

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Length", strconv.Itoa(len(configBytes)))
		w.WriteHeader(http.StatusOK)
		if r.Method == http.MethodHead {
			return
		}
		n, _ := w.Write(configBytes)
		if n != len(configBytes) {
			log.Printf(" - failed length")
		}
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	m.listener = ln
	m.Addr = ln.Addr().String()

	m.server = &http.Server{
		Handler: mux,
	}

	go func() {
		err := m.server.Serve(m.listener)
		if err != nil && err != http.ErrServerClosed {
			log.Printf("[Registry] Server error: %v", err)
		}
	}()

	time.Sleep(50 * time.Millisecond)
	log.Printf("[Registry] Mock OCI registry started at %s", m.Addr)

	return nil
}

func (m *MockOCIRegistry) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.server != nil {
		return m.server.Close()
	}
	return nil
}
