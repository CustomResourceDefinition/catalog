package generator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRealmClient(t *testing.T) {
	t.Setenv(HELM_OCI_PLAIN_HTTP, "true")

	server, finish := setupOciServer(t, []ociChart{
		{
			repoName: "helm",
			name:     "connect",
			tag:      "2.0.0",
			path:     "testdata/connect-2.0.0.tgz",
		},
	})
	defer finish()

	client := newRealmClient()
	versions, err := client.ListOciTags(fmt.Sprintf("%s%s", server.URL, "/helm/connect"))

	assert.Nil(t, err)
	assert.Equal(t, []string{"2.0.0"}, versions)
}
