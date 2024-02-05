package track

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

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
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"artists_info"`

		Lyrics string `json:"lyrics"`
	} `json:"track"`
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

	if data, err = fetchRegionalTopTrackData(ttc.ReqCtx, "india"); err != nil {
		ttc.SetComponentAppError(http.StatusInternalServerError, err)
	} else if err = processRegionalTrackData(data, resp); err != nil {
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
		tracks := tr.(map[string]interface{})
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

	}

	return nil
}

func fetchTrackLyrics(reqCtx context.Context, rttr *RegionalTopTrackResponse) error {
	var data utils.Data
	var err error
	if data, err = fetchTrackID(reqCtx, rttr.Track.ArtistsInfo.Name, rttr.Track.Name); err != nil {
		return err
	}

	fmt.Println(data)

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
