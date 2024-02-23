package track

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"geomelody/components"
	"geomelody/constants"
	"geomelody/utils"

	"github.com/gomodule/redigo/redis"
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

type TrackSuggestion struct {
	Name      string  `json:"name"`
	Match     float64 `json:"match"`
	Duration  float64 `json:"duration"`
	PlayCount float64 `json:"playCount"`
	URL       string  `json:"url"`

	ArtistInfo struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"artist_info"`
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

	TrackSuggestion []TrackSuggestion
}

var countriesMap map[string]string

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

	if isRespInCache(form, ttc.RedisConn, resp) {
		return resp, nil
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
	} else if data, err = fetchTrackSuggestions(ttc.ReqCtx, resp.Track.Name, resp.Track.ArtistsInfo.Name); err != nil {
		ttc.SetComponentAppError(http.StatusInternalServerError, err)
	} else if err = processTrackSuggestionsData(data, resp); err != nil {
		ttc.SetComponentAppError(http.StatusInternalServerError, err)
	} else {
		checkAndCacheResp(form, ttc.RedisConn, resp)
	}

	return resp, err
}

func isRespInCache(form *RegionalTopTrackForm, redisConn redis.Conn, resp *RegionalTopTrackResponse) bool {
	if form.UseCache {
		dataStr, err := utils.RedisGetData(redisConn, form.Country)
		if err != nil {
			log.Printf("data not found in cache")
		} else if dataBytes, err := base64.StdEncoding.DecodeString(dataStr); err != nil {
			log.Printf("error while decoding base64 cache data")
		} else if err := json.Unmarshal(dataBytes, resp); err != nil {
			log.Printf("error unmarshaling cache data")
		} else {
			log.Printf("data found in cache")
			return true
		}
	}

	return false
}

func checkAndCacheResp(form *RegionalTopTrackForm, redisConn redis.Conn, resp *RegionalTopTrackResponse) {
	if form.UseCache {
		if respBytes, err := json.Marshal(resp); err != nil {
			log.Printf("error marshaling data to store in cache")
		} else if respStr := base64.StdEncoding.EncodeToString(respBytes); respStr != "" {
			ttl, _ := strconv.Atoi(constants.REDIS_DEFAULT_EXPIRY)
			if status, err := utils.RedisSetData(redisConn, form.Country, respStr, ttl); err != nil || status {
				log.Printf("error setting data in cache")
			} else {
				log.Printf("data succesfully stored in cache")
			}
		}
	}
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

	log.Printf("fetched regional track data")

	return caMap, nil
}

func processRegionalTrackData(data utils.Data, rttr *RegionalTopTrackResponse) error {
	if tr, ok := data["tracks"]; ok {
		if tempTracks, ok := tr.(map[string]interface{}); ok {
			if attr, ok := tempTracks["@attr"].(map[string]interface{}); ok {
				rttr.Meta.Country, _ = attr["country"].(string)
			}

			tracks := tempTracks["track"].([]interface{})

			if len(tracks) > 0 {
				track, ok := tracks[0].(map[string]interface{})
				if !ok {
					return errors.New("error while processing track vendor API data")
				}
				rttr.Track.Name, _ = track["name"].(string)
				rttr.Track.Duration, _ = track["duration"].(string)

				tempListeners, _ := track["listeners"].(string)
				listeners, _ := strconv.Atoi(tempListeners)
				rttr.Track.Listeners = listeners

				rttr.Track.URL, _ = track["url"].(string)

				tempRank, _ := track["@attr"].(map[string]interface{})["rank"].(string)
				rank, _ := strconv.Atoi(tempRank)
				rttr.Track.Rank = rank + 1

				if artist, ok := track["artist"].(map[string]interface{}); ok {
					rttr.Track.ArtistsInfo.Name, _ = artist["name"].(string)
					rttr.Track.ArtistsInfo.URL, _ = artist["url"].(string)
				}
			} else {
				return errors.New("received empty track data from vendor API. Please check input params")
			}
		} else {
			return errors.New("error while processing track vendor API data")
		}

		log.Printf("processed regional track data")

		return nil
	}

	return errors.New("received empty track data from vendor API. Please check input params")
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

	log.Printf("fetched artist data")

	return caMap, nil
}

func processArtistInfo(data utils.Data, rttr *RegionalTopTrackResponse) error {
	if artist, ok := data["artist"]; ok {
		if ar, ok := artist.(map[string]interface{}); ok {
			rttr.Track.ArtistsInfo.Images, _ = ar["image"].([]interface{})

			stats, ok := ar["stats"].(map[string]interface{})
			if !ok {
				return errors.New("error while processing artist vendor API data")
			}

			tempPlayCount, _ := stats["playcount"].(string)
			playCount, _ := strconv.Atoi(tempPlayCount)
			rttr.Track.ArtistsInfo.Stats.PlayCount = playCount

			tempListeners, _ := stats["listeners"].(string)
			listeners, _ := strconv.Atoi(tempListeners)
			rttr.Track.ArtistsInfo.Stats.Listeners = listeners

			bio, ok := ar["bio"].(map[string]interface{})
			if !ok {
				return errors.New("error while processing artist vendor API data")
			}

			tempSummary, _ := bio["summary"].(string)

			summary := strings.Replace(tempSummary, "\n", ". ", -1)
			rttr.Track.ArtistsInfo.Summary = summary

			log.Printf("processed artist data")
		} else {
			return errors.New("error while processing artist vendor API data")
		}
	} else {
		log.Printf("received empty artist data")
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

	log.Printf("fetched track ID data")

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
				log.Printf("lyrics not present for the track")
				return nil
			} else {
				mms.HasLyrics = true
			}

			trackID, _ := track["track_id"].(float64)
			mms.TrackID = int(trackID)

			commonTrackId, _ := track["commontrack_id"].(float64)
			mms.CommonTrackID = int(commonTrackId)

			artistID, _ := track["artist_id"].(float64)
			mms.ArtistID = int(artistID)

			albumID, _ := track["album_id"].(float64)
			mms.AlbumID = int(albumID)

			mms.AlbumName, _ = track["album_name"].(string)

			checkForTranslation(track, mms)

			log.Printf("processed track ID data")
		} else {
			log.Printf("received more than one track ID data, hence not processed")
		}
	} else {
		return errors.New("error while processing track ID vendor API data")
	}

	return nil
}

func checkForTranslation(track map[string]interface{}, mms *MusicMixSearchResponse) {
	if hasTranslation, ok := track["track_name_translation_list"]; ok {
		translations, _ := hasTranslation.([]interface{})
		for _, translation := range translations {
			val, ok := translation.(map[string]interface{})
			if !ok {
				return
			}

			transVal, ok := val["track_name_translation"].(map[string]interface{})
			if !ok {
				return
			}

			if transVal["language"].(string) == "EN" {
				mms.HasTranslation = true
				mms.TrackName = transVal["translation"].(string)
				mms.ArtistName = track["artist_name"].(string)
				log.Printf("found and processed translated data")
				break
			}
		}
	}
}

func fetchLyrics(reqCtx context.Context, trackID, commonTrackID string) (utils.Data, error) {
	url := fmt.Sprintf("%vtrack.lyrics.get", constants.MUSIC_MIX_URL)
	reqHeaders := map[string]string{"Content-Type": "application/json"}
	params := map[string]string{
		"track_id":       trackID,
		"commontrack_id": commonTrackID,
		"apikey":         constants.MUSIC_MIX_API_KEY,
		"page_size":      "1",
	}
	var data interface{}
	var err error
	if data, err = utils.GetAPIResponse(reqCtx, "GetTrackLyrics", url, http.MethodGet, nil, params, reqHeaders); err != nil {
		return nil, err
	}
	caMap, _ := data.(map[string]interface{})

	log.Printf("fetched lyrics data")

	return caMap, nil
}

func processLyricsData(data utils.Data, rttr *RegionalTopTrackResponse) error {
	if out, ok := data["message"].(map[string]interface{}); ok {
		if body, _ := out["body"].(map[string]interface{}); ok {
			lyrics, ok := body["lyrics"].(map[string]interface{})
			if !ok {
				return errors.New("error while processing lyrics vendor API data")
			}

			lyric, _ := lyrics["lyrics_body"].(string)
			lyric = strings.Replace(lyric, "******* This Lyrics is NOT for Commercial use *******", "", -1)
			lyric = strings.Replace(lyric, "\n", ". ", -1)
			rttr.Track.Lyrics = lyric

			log.Printf("processed lyrics data")
		}
	} else {
		return errors.New("error while processing lyrics vendor API data")
	}

	return nil
}

func fetchTrackSuggestions(reqCtx context.Context, track, artist string) (utils.Data, error) {
	url := fmt.Sprintf("%v", constants.LAST_API_URL)
	reqHeaders := map[string]string{"Content-Type": "application/json"}
	params := map[string]string{
		"method":  "track.getsimilar",
		"artist":  artist,
		"track":   track,
		"api_key": constants.LAST_API_KEY,
		"format":  "json",
		"limit":   "5",
	}
	var data interface{}
	var err error
	if data, err = utils.GetAPIResponse(reqCtx, "GetTrackSuggestions", url, http.MethodGet, nil, params, reqHeaders); err != nil {
		return nil, err
	}
	caMap, _ := data.(map[string]interface{})

	log.Printf("fetched track suggestions data")

	return caMap, nil
}

func processTrackSuggestionsData(data utils.Data, rttr *RegionalTopTrackResponse) error {
	if st, ok := data["similartracks"]; ok {
		if similarTracks, ok := st.(map[string]interface{}); ok {
			tracks, _ := similarTracks["track"].([]interface{})
			rttr.TrackSuggestion = make([]TrackSuggestion, 0)
			if len(tracks) == 0 {
				log.Printf("received empty track suggestions data")

				return nil
			}
			for _, track := range tracks {
				val, ok := track.(map[string]interface{})
				if !ok {
					return errors.New("error while processing track vendor API data")
				}
				trackSuggestion := new(TrackSuggestion)

				trackSuggestion.Name, _ = val["name"].(string)
				trackSuggestion.URL, _ = val["url"].(string)
				trackSuggestion.Match, _ = val["match"].(float64)
				trackSuggestion.Duration, _ = val["duration"].(float64)

				trackSuggestion.PlayCount, _ = val["playcount"].(float64)

				if artist, ok := val["artist"].(map[string]interface{}); ok {
					trackSuggestion.ArtistInfo.Name, _ = artist["name"].(string)
					trackSuggestion.ArtistInfo.URL, _ = artist["url"].(string)
				}

				rttr.TrackSuggestion = append(rttr.TrackSuggestion, *trackSuggestion)
			}

			log.Printf("processed track suggestions data")
		} else {
			return errors.New("error while processing track vendor API data")
		}
	} else {
		log.Printf("received empty track suggestions data")
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
	} else {
		countriesStr, err := os.ReadFile(constants.COUNTRIES_JSON_FILE_NAME)
		if err != nil {
			return err
		}

		err = json.Unmarshal(countriesStr, &countriesMap)
		if err != nil {
			return err
		}

		if country, ok := countriesMap[strings.ToLower(f.Country)]; ok {
			f.Country = country
		} else {
			errMsg += "`country` not found in our database. Please check the country input param, it should follow the ISO 3166-1-Alpha-2 code format"
		}
	}

	if f.UseCache != true && f.UseCache != false {
		if errMsg != "" {
			errMsg += "\n"
		}
		errMsg += "`use_cache` parameter is invalid"
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
