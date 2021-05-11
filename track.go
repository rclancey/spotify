package spotify

import (
	"fmt"

	"github.com/pkg/errors"
)

type Track struct {
	Type string `json:"type"`
	ID string `json:"id"`
	URI string `json:"uri"`
	Name string `json:"name"`
	Album *Album `json:"album"`
	Artists []*Artist `json:"artist"`
	TrackNumber int `json:"track_number"`
	DiscNumber int `json:"disc_number"`
	DurationMS int `json:"duration_ms"`
	Popularity int `json:"popularity"`
	Explicit bool `json:"explicit"`
	Href string `json:"href"`
	PreviewURL string `json:"preview_url"`
	c *SpotifyClient
}

func (c *SpotifyClient) SearchTrack(album, artist, name string) ([]*Track, error) {
	query := fmt.Sprintf("track:\"%s\"", name)
	if album != "" {
		query += fmt.Sprintf(" album:\"%s\"", album)
	}
	if artist != "" {
		query += fmt.Sprintf(" artist:\"%s\"", artist)
	}
	res, err := c.Search(query, "track")
	if err != nil {
		return nil, errors.Wrap(err, "can't search spotify for track " + name)
	}
	return res.Tracks, nil
}

func (c *SpotifyClient) addClientToTracks(tracks ...*Track) {
	for _, tr := range tracks {
		if tr.c == nil {
			tr.c = c
			if tr.Album != nil {
				c.addClientToAlbums(tr.Album)
			}
			c.addClientToArtists(tr.Artists...)
		}
	}
}
