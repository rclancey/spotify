package spotify

import (
	"io/ioutil"
	//"log"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/rclancey/apiclient"
	"github.com/rclancey/cache/fs"
)

type SpotifyClient struct {
	client *apiclient.APIClient
}

func NewSpotifyClient(clientId, clientSecret, cacheDir string, cacheTime time.Duration) (*SpotifyClient, error) {
	auth, err := NewClientAuth(clientId, clientSecret)
	if err != nil {
		return nil, errors.Wrap(err, "can't create spotify auth")
	}
	opts := apiclient.APIClientOptions{
		BaseURL: "https://api.spotify.com/v1/",
		RequestTimeout: 0,
		CacheStore: fscache.NewFSCacheStore(cacheDir),
		MaxCacheTime: cacheTime,
		MaxRequestsPerSecond: 4.0,
		Auth: auth,
	}
	api, err := apiclient.NewAPIClient("https://api.spotify.com/v1/", cacheDir, cacheTime, 4.0, auth)
	if err != nil {
		return nil, errors.Wrap(err, "can't create spotify api client")
	}
	client := &SpotifyClient{
		client: api,
	}
	return client, nil
}

type FollowerInfo struct {
	Total int `json:"total"`
	Href *string `json:"href"`
}

type Image struct {
	URL string `json:"url"`
	Width *int `json:"width"`
	Height *int `json:"height"`
}

func (img *Image) Get(c *SpotifyClient) ([]byte, string, error) {
	res, err := c.client.Client().Get(img.URL)
	if err != nil {
		return nil, "", errors.Wrap(err, "can't get spotify image")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, "", errors.New(res.Status)
	}
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, "", errors.Wrap(err, "can't read spotify image response")
	}
	ct := res.Header.Get("Content-Type")
	return data, ct, nil
}

type SortableImages []*Image
func (si SortableImages) Len() int { return len(si) }
func (si SortableImages) Swap(i, j int) { si[i], si[j] = si[j], si[i] }
func (si SortableImages) Less(i, j int) bool {
	if si[i].Width != nil && si[j].Width != nil {
		return *si[i].Width > *si[j].Width
	}
	if si[i].Height != nil && si[j].Height != nil {
		return *si[i].Height > *si[j].Height
	}
	return i < j
}
