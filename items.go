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

func (c *SpotifyClient) GetPaged(rsrc string, q url.Values) (*SearchResult, error) {
	result := &SearchResult{
		Artists: []*Artist{},
		Albums: []*Album{},
		Tracks: []*Track{},
	}
	for {
		res, err := c.client.Get(rsrc, q)
		if err != nil {
			return nil, errors.Wrap(err, "can't execute spotify paged request")
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
		page := &PagingObject{}
		err = json.Unmarshal(data, page)
		if err != nil {
			return nil, errors.Wrap(err, "can't unmarshal spotify search response")
		}
		for _, item := range page.Items {
			switch it := item.(type) {
			case *Artist:
				result.Artists = append(result.Artists, it)
			case *Album:
				result.Albums = append(result.Albums, it)
			case *Track:
				result.Tracks = append(result.Tracks, it)
			}
		}
		if page.NextHref == nil || *page.NextHref == "" {
			break
		}
		nu, err := url.Parse(*page.NextHref)
		if err != nil {
			break
		}
		rsrc = nu.Path
		q = nu.Query()
	}
	return result, nil
}
