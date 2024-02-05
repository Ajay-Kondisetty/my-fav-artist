package track

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"geomelody/components"
	"geomelody/constants"
	"geomelody/utils"

	"github.com/microcosm-cc/bluemonday"
)

type TopTrackComponent struct {
	components.BaseComponent
}

type TopTrack interface {
	GetRegionalTopTrack(*RegionalTopTrackForm) (*RegionalTopTrackResponse, error)
	GetRegionalTopTrackForm() *RegionalTopTrackForm
	GetComponentAppError() *utils.AppError
}

type RegionalTopTrackForm struct {
	Country  string `json:"country"`
	UseCache bool   `json:"use_cache"`
}

type RegionalTopTrackResponse struct {
	Meta struct {
		Country string `json:"country"`
	} `json:"meta"`

	Track struct {
		Rank      int    `json:"rank"`
		Name      string `json:"name"`
		Duration  string `json:"duration"`
		Listeners int    `json:"listeners"`
		URL       string `json:"url"`

		ArtistsInfo struct {
			Name    string        `json:"name"`
			URL     string        `json:"url"`
			Images  []interface{} `json:"images"`
			Summary string        `json:"summary"`
			Stats   struct {
				Listeners int `json:"listeners"`
				PlayCount int `json:"play_count"`
			}
		} `json:"artists_info"`

		Lyrics string `json:"lyrics"`
	} `json:"track"`
}

type MusicMixSearchResponse struct {
	HasTranslation bool   `json:"has_translation"`
	HasLyrics      bool   `json:"has_lyrics"`
	TrackID        int    `json:"track_id"`
	CommonTrackID  int    `json:"commontrack_id"`
	TrackName      string `json:"track_name"`
	ArtistID       int    `json:"artist_id"`
	ArtistName     string `json:"artist_name"`
	AlbumID        int    `json:"album_id"`
	AlbumName      string `json:"album_name"`
}

// GetRegionalTopTrack is used to call required external APIs to fetch top track, artists info of the track, lyrics of the track, and suggestions based on the artists and track of the given country.
// It returns top track data and error.
func (ttc *TopTrackComponent) GetRegionalTopTrack(form *RegionalTopTrackForm) (*RegionalTopTrackResponse, error) {
	resp := new(RegionalTopTrackResponse)
	var err error
	var data utils.Data
	if err := form.Valid(); err != nil {
		ttc.AppError = &utils.AppError{
			Status: http.StatusBadRequest,
			Error:  err,
		}
		return nil, err
	}

	if data, err = fetchRegionalTopTrackData(ttc.ReqCtx, form.Country); err != nil {
		ttc.SetComponentAppError(http.StatusInternalServerError, err)
	} else if err = processRegionalTrackData(data, resp); err != nil {
		ttc.SetComponentAppError(http.StatusInternalServerError, err)
	} else if data, err = fetchArtistInfo(ttc.ReqCtx, resp.Track.ArtistsInfo.Name); err != nil {
		ttc.SetComponentAppError(http.StatusInternalServerError, err)
	} else if err = processArtistInfo(data, resp); err != nil {
		ttc.SetComponentAppError(http.StatusInternalServerError, err)
	} else if err = fetchTrackLyrics(ttc.ReqCtx, resp); err != nil {
		ttc.SetComponentAppError(http.StatusInternalServerError, err)
	}

	return resp, err
}

// fetchRegionalTopTrackData is used to fetch top track of the given country from LAST API.
// It returns API response and error.
func fetchRegionalTopTrackData(reqCtx context.Context, country string) (utils.Data, error) {
	url := fmt.Sprintf("%v", constants.LAST_API_URL)
	reqHeaders := map[string]string{"Content-Type": "application/json"}
	params := map[string]string{
		"method":  "geo.gettoptracks",
		"country": country,
		"api_key": constants.LAST_API_KEY,
		"format":  "json",
		"limit":   "1",
	}
	var data interface{}
	var err error
	if data, err = utils.GetAPIResponse(reqCtx, "GetTopTrackByCountry", url, http.MethodGet, nil, params, reqHeaders); err != nil {
		return nil, err
	}
	caMap, _ := data.(map[string]interface{})

	return caMap, nil
}

func processRegionalTrackData(data utils.Data, rttr *RegionalTopTrackResponse) error {
	if tr, ok := data["tracks"]; ok {
		if tracks, ok := tr.(map[string]interface{}); ok {
			attr := tracks["@attr"].(map[string]interface{})
			rttr.Meta.Country = attr["country"].(string)

			track := tracks["track"].([]interface{})[0].(map[string]interface{})
			rttr.Track.Name = track["name"].(string)
			rttr.Track.Duration = track["duration"].(string)

			listeners, _ := strconv.Atoi(track["listeners"].(string))
			rttr.Track.Listeners = listeners

			rttr.Track.URL = track["url"].(string)

			rank, _ := strconv.Atoi(track["@attr"].(map[string]interface{})["rank"].(string))
			rttr.Track.Rank = rank + 1

			artist := track["artist"].(map[string]interface{})
			rttr.Track.ArtistsInfo.Name = artist["name"].(string)
			rttr.Track.ArtistsInfo.URL = artist["url"].(string)
		} else {
			return errors.New("error while processing track vendor API data")
		}

		return nil
	}

	return errors.New("received empty tracks data from vendor API. Please check input params")
}

func fetchArtistInfo(reqCtx context.Context, artist string) (utils.Data, error) {
	url := fmt.Sprintf("%v", constants.LAST_API_URL)
	reqHeaders := map[string]string{"Content-Type": "application/json"}
	params := map[string]string{
		"method":  "artist.getinfo",
		"artist":  artist,
		"api_key": constants.LAST_API_KEY,
		"format":  "json",
		"limit":   "1",
	}
	var data interface{}
	var err error
	if data, err = utils.GetAPIResponse(reqCtx, "GetArtistInfo", url, http.MethodGet, nil, params, reqHeaders); err != nil {
		return nil, err
	}
	caMap, _ := data.(map[string]interface{})

	return caMap, nil
}

func processArtistInfo(data utils.Data, rttr *RegionalTopTrackResponse) error {
	if artist, ok := data["artist"]; ok {
		if ar, ok := artist.(map[string]interface{}); ok {
			rttr.Track.ArtistsInfo.Images = ar["image"].([]interface{})

			stats := ar["stats"].(map[string]interface{})
			playCount, _ := strconv.Atoi(stats["playcount"].(string))
			rttr.Track.ArtistsInfo.Stats.PlayCount = playCount

			listeners, _ := strconv.Atoi(stats["listeners"].(string))
			rttr.Track.ArtistsInfo.Stats.Listeners = listeners

			summary := strings.Replace(ar["bio"].(map[string]interface{})["summary"].(string), "\n", ". ", -1)
			rttr.Track.ArtistsInfo.Summary = summary
		} else {
			return errors.New("error while processing artist vendor API data")
		}
	}

	return nil
}

func fetchTrackLyrics(reqCtx context.Context, rttr *RegionalTopTrackResponse) error {
	var data utils.Data
	var err error
	musicMixResp := new(MusicMixSearchResponse)
	if data, err = fetchTrackID(reqCtx, rttr.Track.ArtistsInfo.Name, rttr.Track.Name); err != nil {
		return err
	} else if err = processTrackIDData(data, musicMixResp); err != nil {
		return err
	}

	if musicMixResp.HasLyrics {
		if data, err = fetchLyrics(reqCtx, strconv.Itoa(musicMixResp.TrackID), strconv.Itoa(musicMixResp.CommonTrackID)); err != nil {
			return err
		} else if err = processLyricsData(data, rttr); err != nil {
			return err
		}
	}

	if musicMixResp.HasTranslation {
		rttr.Track.Name = musicMixResp.TrackName
		rttr.Track.ArtistsInfo.Name = musicMixResp.ArtistName
	}

	return nil
}

func fetchTrackID(reqCtx context.Context, artist, track string) (utils.Data, error) {
	url := fmt.Sprintf("%vtrack.search", constants.MUSIC_MIX_URL)
	reqHeaders := map[string]string{"Content-Type": "application/json"}
	params := map[string]string{
		"q_track":   track,
		"q_artist":  artist,
		"apikey":    constants.MUSIC_MIX_API_KEY,
		"page_size": "1",
	}
	var data interface{}
	var err error
	if data, err = utils.GetAPIResponse(reqCtx, "GetTrackID", url, http.MethodGet, nil, params, reqHeaders); err != nil {
		return nil, err
	}
	caMap, _ := data.(map[string]interface{})

	return caMap, nil
}

func processTrackIDData(data utils.Data, mms *MusicMixSearchResponse) error {
	if out, ok := data["message"].(map[string]interface{}); ok {
		trackList := out["body"].(map[string]interface{})["track_list"].([]interface{})

		if len(trackList) == 1 {
			track := trackList[0].(map[string]interface{})["track"].(map[string]interface{})
			hasLyrics := int(track["has_lyrics"].(float64))
			if hasLyrics == 0 {
				mms.HasLyrics = false
				return nil
			} else {
				mms.HasLyrics = true
			}

			mms.TrackID = int(track["track_id"].(float64))
			mms.CommonTrackID = int(track["commontrack_id"].(float64))
			mms.ArtistID = int(track["artist_id"].(float64))
			mms.AlbumID = int(track["album_id"].(float64))
			mms.AlbumName = track["album_name"].(string)

			checkForTranslation(track, mms)
		}
	} else {
		return errors.New("error while processing track ID vendor API data")
	}

	return nil
}

func checkForTranslation(track map[string]interface{}, mms *MusicMixSearchResponse) {
	if hasTranslation, ok := track["track_name_translation_list"]; ok {
		translations := hasTranslation.([]interface{})
		for _, translation := range translations {
			val := translation.(map[string]interface{})
			transVal := val["track_name_translation"].(map[string]interface{})
			if transVal["language"].(string) == "EN" {
				mms.HasTranslation = true
				mms.TrackName = transVal["translation"].(string)
				mms.ArtistName = track["artist_name"].(string)
				break
			}
		}
	}
}

func fetchLyrics(reqCtx context.Context, trackID, commomTrackID string) (utils.Data, error) {
	url := fmt.Sprintf("%vtrack.lyrics.get", constants.MUSIC_MIX_URL)
	reqHeaders := map[string]string{"Content-Type": "application/json"}
	params := map[string]string{
		"track_id":       trackID,
		"commontrack_id": commomTrackID,
		"apikey":         constants.MUSIC_MIX_API_KEY,
		"page_size":      "1",
	}
	var data interface{}
	var err error
	if data, err = utils.GetAPIResponse(reqCtx, "GetTrackLyrics", url, http.MethodGet, nil, params, reqHeaders); err != nil {
		return nil, err
	}
	caMap, _ := data.(map[string]interface{})

	return caMap, nil
}

func processLyricsData(data utils.Data, rttr *RegionalTopTrackResponse) error {
	if out, ok := data["message"].(map[string]interface{}); ok {
		lyrics := out["body"].(map[string]interface{})["lyrics"].(map[string]interface{})

		lyric := lyrics["lyrics_body"].(string)
		lyric = strings.Replace(lyric, "******* This Lyrics is NOT for Commercial use *******", "", -1)
		lyric = strings.Replace(lyric, "\n", ". ", -1)
		rttr.Track.Lyrics = lyric
	} else {
		return errors.New("error while processing lyrics vendor API data")
	}

	return nil
}

// GetRegionalTopTrackForm is used to return new regional top track form instance.
// It returns regional top track form instance.
func (ttc *TopTrackComponent) GetRegionalTopTrackForm() *RegionalTopTrackForm {
	return new(RegionalTopTrackForm)
}

// GetComponentAppError is used to retrieve app error from the component struct.
// It returns app error of the component.
func (ttc *TopTrackComponent) GetComponentAppError() *utils.AppError {
	return ttc.AppError
}

func (ttc *TopTrackComponent) SetComponentAppError(status int, err error) {
	ttc.AppError = &utils.AppError{
		Status: status,
		Error:  err,
	}
}

// Valid validates and sanitizes the top regional track form.
func (f *RegionalTopTrackForm) Valid() error {
	errMsg := ""
	if f.Country == "" {
		errMsg += "`country` parameter is invalid"
	}

	if f.UseCache != true && f.UseCache != false {
		if errMsg != "" {
			errMsg += "\n"
		}
		errMsg += "`use_cache parameter is invalid"
	}

	p := bluemonday.UGCPolicy()
	f.Country = p.Sanitize(f.Country)

	if errMsg != "" {
		return errors.New(errMsg)
	}

	return nil
}

func init() {
	components.ComponentMap["TopTrack"] = func(bc *components.BaseComponent) interface{} {
		c := &TopTrackComponent{BaseComponent: *bc}

		return TopTrack(c)
	}
}
