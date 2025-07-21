package generator

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type realmClient struct {
	client        http.Client
	defaultScheme string
}

type versionsResponse struct {
	Tags []string `json:"tags"`
}

type tokenResponse struct {
	Token string `json:"token"`
}

func newRealmClient(plainHttp bool) realmClient {
	defaultScheme := "https"
	if plainHttp {
		defaultScheme = "http"
	}

	return realmClient{
		client: http.Client{
			Timeout: 15 * time.Second,
		},
		defaultScheme: defaultScheme,
	}
}

// ListOciTags produces a list of tags from a OCI repository. If needed, it will handle authentication.
func (realm *realmClient) ListOciTags(uri string) ([]string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	if !strings.EqualFold(u.Scheme, "oci") {
		return nil, fmt.Errorf("not a valid oci repository uri: %s", uri)
	}
	scheme := strings.ReplaceAll(u.Scheme, "oci", realm.defaultScheme)

	if strings.EqualFold(u.Host, "docker.io") {
		u.Host = "registry-1.docker.io"
	}

	path := strings.Trim(u.Path, "/")
	return realm.listOciTags(fmt.Sprintf("%s://%s", scheme, u.Host), fmt.Sprintf("/v2/%s/tags/list", path), "")
}

func (realm *realmClient) listOciTags(host, path, token string) ([]string, error) {
	request := fmt.Sprintf("%s%s", host, path)
	req, err := http.NewRequest(http.MethodGet, request, nil)
	if err != nil {
		return nil, err
	}

	if len(token) != 0 {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	resp, err := realm.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		token, err := realm.authenticate(resp)
		if err != nil {
			return nil, err
		}

		return realm.listOciTags(host, path, *token)
	}

	bytes, err := realm.read(resp)
	if err != nil {
		return nil, err
	}

	r := versionsResponse{}
	err = json.Unmarshal(bytes, &r)
	if err != nil {
		return nil, err
	}

	tags, err := realm.next(resp.Header, host, token)
	if err != nil {
		return nil, err
	}

	return append(r.Tags, tags...), nil
}

func (realm *realmClient) fetchToken(req string) (*string, error) {
	resp, err := realm.client.Get(req)
	if err != nil {
		return nil, err
	}

	bytes, err := realm.read(resp)
	if err != nil {
		return nil, err
	}

	r := tokenResponse{}
	err = json.Unmarshal(bytes, &r)
	if err != nil {
		return nil, err
	}

	return &r.Token, nil
}

func (*realmClient) read(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("response had status: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (realm *realmClient) authenticate(resp *http.Response) (*string, error) {
	defer resp.Body.Close()

	challenge, err := realm.parseChallenge(resp.Header.Get("www-authenticate"))
	if err != nil {
		return nil, err
	}

	tokenRequest := fmt.Sprintf("%s?service=%s&scope=%s", challenge["realm"], challenge["service"], challenge["scope"])
	return realm.fetchToken(tokenRequest)
}

func (*realmClient) parseChallenge(challenge string) (map[string]string, error) {
	parts := strings.SplitN(challenge, " ", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[0]) != "Bearer" {
		return nil, fmt.Errorf("encountered a non-bearer challenge: %s", challenge)
	}

	cfg := make(map[string]string, 0)
	parameters := strings.SplitSeq(parts[1], ",")
	for parameter := range parameters {
		parameter = strings.TrimSpace(parameter)
		kv := strings.SplitN(parameter, "=", 2)
		if len(kv) != 2 {
			continue
		}
		cfg[strings.ToLower(kv[0])] = strings.Trim(kv[1], `"`)
	}

	return cfg, nil
}

func (realm *realmClient) next(header http.Header, host, token string) ([]string, error) {
	result := make([]string, 0)
	links := header.Get("link")

	for link := range strings.SplitSeq(links, ",") {
		if len(link) == 0 || !strings.Contains(strings.ToLower(link), `; rel="next"`) {
			continue
		}

		req := strings.SplitN(link, ";", 2)[0]
		path := strings.Trim(req, "<>")

		return realm.listOciTags(host, path, token)
	}

	return result, nil
}
