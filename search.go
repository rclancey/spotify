package spotify

import (
	"encoding/json"
	"io/ioutil"
	//"log"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

type PagingObject struct {
	Href string `json:"href"`
	PreviousHref *string `json:"previous"`
	NextHref *string `json:"next"`
	Offset int `json:"offset"`
	Limit int `json:"limit"`
	Total int `json:"total"`
	Items TypedItems `json:"items"`
}

type TypedItems []interface{}

type TypedItem struct {
	Type string `json:"type"`
}

func (tis *TypedItems) UnmarshalJSON(data []byte) error {
	rawItems := []json.RawMessage{}
	err := json.Unmarshal(data, &rawItems)
	if err != nil {
		return errors.Wrap(err, "can't unmarshal typed items into raw message")
	}
	items := make([]interface{}, len(rawItems))
	for i, rawItem := range rawItems {
		ti := &TypedItem{}
		err := json.Unmarshal(rawItem, ti)
		if err != nil {
			return errors.Wrap(err, "can't unmarshal typed item")
		}
		switch ti.Type {
		case "artist":
			items[i] = &Artist{}
		case "album":
			items[i] = &Album{}
		case "track":
			items[i] = &Track{}
		default:
			return errors.Errorf("unknown item type: %s", ti.Type)
		}
		err = json.Unmarshal(rawItem, items[i])
		if err != nil {
			return errors.Wrapf(err, "can't unmarshal item into %T", items[i])
		}
	}
	*tis = items
	return nil
}

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
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			return nil, errors.New(res.Status)
		}
		data, err := ioutil.ReadAll(res.Body)
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
	return result, nil
}

