package spotify

import (
	"fmt"
	"net/url"
	"path"

	"github.com/pkg/errors"
)

type Album struct {
	Type string `json:"type"`
	ID string `json:"id"`
	URI string `json:"uri"`
	Name string `json:"name"`
	Artists []*Artist `json:"artists"`
	Genres []string `json:"genres"`
	ReleaseDate string `json:"release_date"`
	ReleaseDatePrecision string `json:"release_date_precision"`
	Popularity int `json:"popularity"`
	Images []*Image `json:"images"`
	Href string `json:"href"`
	Tracks []*Track `json:"tracks"`
	c *SpotifyClient
}

func (c *SpotifyClient) SearchAlbum(albumArtist, name string) ([]*Album, error) {
	query := fmt.Sprintf("album:\"%s\"", name)
	if albumArtist != "" {
		query += fmt.Sprintf(" artist:\"%s\"", albumArtist)
	}
	res, err := c.Search(query, "album")
	if err != nil {
		return nil, errors.Wrap(err, "can't search spotify for album " + name)
	}
	return res.Albums, nil
}

func (c *SpotifyClient) addClientToAlbums(albums ...*Album) {
	for _, alb := range albums {
		if alb.c == nil {
			alb.c = c
			c.addClientToArtists(alb.Artists...)
			c.addClientToTracks(alb.Tracks...)
		}
	}
}

func (alb *Album) GetTracks() ([]*Track, error) {
	if alb.Tracks != nil && len(alb.Tracks) > 0 {
		return alb.Tracks, nil
	}
	rsrc := path.Join("albums", alb.ID, "tracks")
	q := url.Values{}
	q.Set("limit", "50")
	q.Set("offset", "0")
	sr, err := alb.c.GetPaged(rsrc, q)
	if err != nil {
		return nil, err
	}
	alb.c.addClientToTracks(sr.Tracks...)
	alb.Tracks = sr.Tracks
	return alb.Tracks, nil
}
