package spotify

import (
	//"log"
	"sort"

	"github.com/pkg/errors"
)


type Artist struct {
	ID string `json:"id"`
	Name string `json:"name"`
	URI string `json:"uri"`
	Type string `json:"type"`
	ExternalURLs map[string]string `json:"external_urls"`
	Followers *FollowerInfo `json:"followers"`
	Genres []string `json:"genres"`
	Href string `json:"href"`
	Images []*Image `json:"images"`
	Popularity int `json:"popularity"`
}

func (art *Artist) GetImage(c *SpotifyClient) (img []byte, ct string, err error) {
	if len(art.Images) == 0 {
		return nil, "", nil
	}
	sort.Sort(SortableImages(art.Images))
	for _, im := range art.Images {
		img, ct, err = im.Get(c)
		if err == nil {
			return img, ct, nil
		}
	}
	return nil, "", errors.Wrap(err, "can't get artist image")
}

func (c *SpotifyClient) SearchArtist(name string) ([]*Artist, error) {
	res, err := c.Search(name, "artist")
	if err != nil {
		return nil, errors.Wrap(err, "can't search spotify for artist " + name)
	}
	return res.Artists, nil
}

func (c *SpotifyClient) GetArtistImage(name string) (img []byte, ct string, err error) {
	arts, err := c.SearchArtist(name)
	if err != nil {
		return nil, "", errors.Wrap(err, "can't find artist " + name)
	}
	if len(arts) == 0 {
		return nil, "", errors.New("no such artist")
	}
	for _, art := range arts {
		img, ct, err = art.GetImage(c)
		if err == nil && img != nil {
			return img, ct, nil
		}
	}
	return nil, "", errors.Wrap(err, "can't get artist image")
}
