package generator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRealmClient(t *testing.T) {
	server, finish := setupOciServer(t, ociChart{
		repoName: "helm",
		name:     "connect",
		tag:      "2.0.0",
		path:     "testdata/connect-2.0.0.tgz",
	})
	defer finish()

	client := newRealmClient(true)
	versions, err := client.ListOciTags(fmt.Sprintf("%s%s", server.URL, "/helm/connect"))

	assert.Nil(t, err)
	assert.Equal(t, []string{"2.0.0"}, versions)
}
