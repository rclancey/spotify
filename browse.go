package spotify

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type SeedInfo struct {
	Type string `json:"type"`
	ID string `json:"id"`
	Href string `json:"href"`
	InitialPoolSize int `json:"initialPoolSize"`
	AfterFilteringSize int `json:"afterFilteringSize"`
	AfterRelinkingSize int `json:"afterRelinkingSize"`
}

type RecommendationResult struct {
	Tracks []*Track `json:"tracks"`
	Seeds []*SeedInfo `json:"seeds"`
}

func (c *SpotifyClient) Recommend(seeds ...interface{}) (*RecommendationResult, error) {
	seedArtists := []string{}
	seedAlbums := []string{}
	seedTracks := []string{}
	for _, obj := range seeds {
		switch seed := obj.(type) {
		case *Artist:
			if seed.ID == "" {
				artists, err := c.SearchArtist(seed.Name)
				if err == nil && len(artists) > 0 {
					seed = artists[0]
				} else {
					log.Println("no sporitfy artist for", seed.Name)
					continue
				}
			}
			seedArtists = append(seedArtists, seed.ID)
		/*
		case *Album:
			if seed.ID == "" {
				artist := ""
				if len(seed.Artists) > 0 {
					artist = seed.Artists[0].Name
				}
				albums, err := c.SearchAlbum(artist, seed.Name)
				if err == nil && len(albums) > 0 {
					seed = albums[0]
				} else {
					log.Println("no spotify album for", artist, seed.Name)
					continue
				}
			}
			seedAlbums = append(seedAlbums, seed.ID)
		*/
		case *Track:
			if seed.ID == "" {
				artist := ""
				album := ""
				if len(seed.Artists) > 0 {
					artist = seed.Artists[0].Name
				}
				if seed.Album != nil {
					album = seed.Album.Name
				}
				tracks, err := c.SearchTrack(album, artist, seed.Name)
				if len(tracks) == 0 {
					tracks, err = c.SearchTrack("", artist, seed.Name)
				}
				if err == nil && len(tracks) > 0 {
					seed = tracks[0]
				} else {
					log.Println("no spotify track for", album, artist, seed.Name)
					continue
				}
			}
			seedTracks = append(seedTracks, seed.ID)
		}
	}
	q := url.Values{}
	ok := false
	if len(seedArtists) > 0 {
		q.Set("seed_artists", strings.Join(seedArtists, ","))
		ok = true
	}
	if len(seedAlbums) > 0 {
		q.Set("seed_albums", strings.Join(seedAlbums, ","))
		ok = true
	}
	if len(seedTracks) > 0 {
		q.Set("seed_tracks", strings.Join(seedTracks, ","))
		ok = true
	}
	if !ok {
		return nil, errors.New("no seeds")
	}
	q.Set("limit", "100")
	rsrc := "recommendations"
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
		result := &RecommendationResult{}
		err = json.Unmarshal(data, result)
		if err != nil {
			return nil, errors.Wrap(err, "can't unmarshal spotify search response")
		}
		return result, nil
	}
	return nil, nil
}

type ArgRange struct {
	Min    float64 `json:"min,omitempty"`
	Max    float64 `json:"max,omitempty"`
	Target float64 `json:"target,omitempty"`
}

func (r ArgRange) AddQuery(prefix string, q url.Values) {
	if r.Min != 0 {
		q.Set(prefix + "_min", strconv.FormatFloat(r.Min, 'f', 3, 64))
	}
	if r.Max != 0 {
		q.Set(prefix + "_max", strconv.FormatFloat(r.Max, 'f', 3, 64))
	}
	if r.Target != 0 {
		q.Set(prefix + "_target", strconv.FormatFloat(r.Target, 'f', 3, 64))
	}
}

type MixArgs struct {
	Acousticness     ArgRange `json:"acousticness,omitempty"`
	Danceability     ArgRange `json:"danceability,omitempty"`
	DurationMS       ArgRange `json:"duration_ms,omitempty"`
	Energy           ArgRange `json:"energy,omitempty"`
	Instrumentalness ArgRange `json:"instrumentalness,omitempty"`
	Key              ArgRange `json:"key,omitempty"`
	Liveness         ArgRange `json:"liveness,omitempty"`
	Loudness         ArgRange `json:"loudness,omitempty"`
	Mode             ArgRange `json:"mode,omitempty"`
	Popularity       ArgRange `json:"popularity,omitempty"`
	Speechiness      ArgRange `json:"speechiness,omitempty"`
	Tempo            ArgRange `json:"tempo,omitempty"`
	TimeSignature    ArgRange `json:"time_signature,omitempty"`
	Valence          ArgRange `json:"valence,omitempty"`
}

func (a MixArgs) AddQuery(q url.Values) {
	ra := reflect.ValueOf(a)
	rt := reflect.TypeOf(a)
	n := rt.NumField()
	for i := 0; i < n; i += 1 {
		r, isa := ra.Field(i).Interface().(ArgRange)
		if isa {
			ft := rt.Field(i)
			prefix := strings.Split(ft.Tag.Get("json"), ",")[0]
			r.AddQuery(prefix, q)
		}
	}
}

type GenresResponse struct {
	Genres []string `json:"genres"`
}

func (c *SpotifyClient) RecommendationGenres() ([]string, error) {
	for {
		res, err := c.client.Get("recommendations/available-genre-seeds", url.Values{})
		if err != nil {
			return nil, errors.Wrap(err, "can't execute spotify request")
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
		result := &GenresResponse{}
		err = json.Unmarshal(data, result)
		if err != nil {
			return nil, errors.Wrap(err, "can't unmarshal spotify search response")
		}
		return result.Genres, nil
	}
	return nil, nil
}

func (c *SpotifyClient) Mix(genre string, args MixArgs) (*RecommendationResult, error) {
	q := url.Values{}
	q.Set("seed_genres", genre)
	q.Set("market", "us")
	q.Set("limit", "100")
	args.AddQuery(q)
	log.Println("mix:", q.Encode())
	rsrc := "recommendations"
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
		result := &RecommendationResult{}
		err = json.Unmarshal(data, result)
		if err != nil {
			return nil, errors.Wrap(err, "can't unmarshal spotify search response")
		}
		return result, nil
	}
	return nil, nil
}
