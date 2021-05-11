package spotify

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

type SearchResultPage struct {
	Artists PagingObject `json:"artists"`
	Albums PagingObject `json:"albums"`
	Tracks PagingObject `json:"tracks"`
}

type SearchResult struct {
	Artists []*Artist
	Albums []*Album
	Tracks []*Track
}

func (c *SpotifyClient) Search(name, kind string) (*SearchResult, error) {
	q := url.Values{}
	q.Set("q", name)
	q.Set("type", kind)
	rsrc := "search"
	result := &SearchResult{}
	for {
		res, err := c.client.Get(rsrc, q)
		if err != nil {
			return nil, errors.Wrap(err, "can't execute spotify search")
		}
		if res.StatusCode != http.StatusOK {
			res.Body.Close()
			if res.StatusCode == http.StatusTooManyRequests {
				wait, err := strconv.Atoi(res.Header.Get("Retry-After"))
				if err == nil {
					log.Printf("API ratelimit; waiting %d seconds", wait)
					time.Sleep(time.Duration(wait + 1) * time.Second)
					continue
				}
			}
			return nil, errors.New(res.Status)
		}
		data, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			return nil, errors.Wrap(err, "can't read spotify search response")
		}
		sr := &SearchResultPage{}
		err = json.Unmarshal(data, sr)
		if err != nil {
			return nil, errors.Wrap(err, "can't unmarshal spotify search response")
		}
		itemsets := []TypedItems{
			sr.Artists.Items,
			sr.Albums.Items,
			sr.Tracks.Items,
		}
		for _, items := range itemsets {
			if items == nil {
				continue
			}
			for _, item := range items {
				switch it := item.(type) {
				case *Artist:
					if result.Artists == nil {
						result.Artists = []*Artist{it}
					} else {
						result.Artists = append(result.Artists, it)
					}
				case *Album:
					if result.Albums == nil {
						result.Albums = []*Album{it}
					} else {
						result.Albums = append(result.Albums, it)
					}
				case *Track:
					if result.Tracks == nil {
						result.Tracks = []*Track{it}
					} else {
						result.Tracks = append(result.Tracks, it)
					}
				}
			}
		}
		// TODO
		if sr.Artists.NextHref == nil || *sr.Artists.NextHref == "" {
			break
		}
		nu, err := url.Parse(*sr.Artists.NextHref)
		if err != nil {
			break
		}
		rsrc = nu.Path
		q = nu.Query()
		break
	}
	c.addClientToArtists(result.Artists...)
	c.addClientToAlbums(result.Albums...)
	c.addClientToTracks(result.Tracks...)
	return result, nil
}

