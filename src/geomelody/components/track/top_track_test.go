package track

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"geomelody/components"
	"geomelody/constants"
	"geomelody/utils"

	"github.com/stretchr/testify/assert"
)

func TestRegionalTopTrackForm_Valid(t *testing.T) {
	constants.COUNTRIES_JSON_FILE_NAME = "../../countries.json"

	type vars struct {
		form *RegionalTopTrackForm
	}

	testCases := []struct {
		name string

		vars vars

		hasErr bool
		err    string
	}{
		{
			name: "should fail when country is empty",
			vars: vars{
				form: &RegionalTopTrackForm{
					Country: "",
				},
			},
			hasErr: true,
			err:    "`country` parameter is invalid",
		},
		{
			name: "should fail when country format is not ISO 3166-1-Alpha-2 code format",
			vars: vars{
				form: &RegionalTopTrackForm{
					Country: "india",
				},
			},
			hasErr: true,
			err:    "`country` not found in our database. Please check the country input param, it should follow the ISO 3166-1-Alpha-2 code format",
		},
		{
			name: "should success to retrieve the ISO 3166-1 format country from the given ISO 3166-1-Alpha-2 country format",
			vars: vars{
				form: &RegionalTopTrackForm{
					Country: "in",
				},
			},
		},
	}

	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			// Setup
			form := tCase.vars.form

			// Run test
			err := form.Valid()

			// Assert
			if tCase.hasErr {
				if assert.Errorf(t, err, "case: %v", tCase) {
					assert.Containsf(t, err.Error(), tCase.err, "case: %v", tCase)
				}
			} else {
				assert.NoErrorf(t, err, "case: %v", tCase)
			}
		})
	}
}

func TestTopTrackComponent_GetComponentAppError(t *testing.T) {
	type vars struct {
		component components.BaseComponent
	}

	testCases := []struct {
		name string

		vars vars

		want *utils.AppError
	}{
		{
			name: "should success to fetch App Error from the component",
			vars: vars{
				component: components.BaseComponent{
					ReqCtx: context.Background(),
					AppError: &utils.AppError{
						Error:  errors.New("some error"),
						Status: http.StatusBadRequest,
					},
				},
			},
			want: &utils.AppError{
				Error:  errors.New("some error"),
				Status: http.StatusBadRequest,
			},
		},
	}

	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			// Setup
			ttc := &TopTrackComponent{
				BaseComponent: tCase.vars.component,
			}

			// Run test
			got := ttc.GetComponentAppError()

			// Assert
			assert.Exactlyf(t, tCase.want, got, "case: %v", tCase)
		})
	}
}

func TestTopTrackComponent_GetRegionalTopTrackForm(t *testing.T) {
	type vars struct {
		component components.BaseComponent
	}

	testCases := []struct {
		name string

		vars vars

		want *RegionalTopTrackForm
	}{
		{
			name: "should success to fetch the form from the component",
			vars: vars{
				component: components.BaseComponent{
					ReqCtx: context.Background(),
				},
			},
			want: new(RegionalTopTrackForm),
		},
	}

	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			// Setup
			ttc := &TopTrackComponent{
				BaseComponent: tCase.vars.component,
			}

			// Run test
			got := ttc.GetRegionalTopTrackForm()

			// Assert
			assert.Exactlyf(t, tCase.want, got, "case: %v", tCase)
		})
	}
}

func TestTopTrackComponent_SetComponentAppError(t *testing.T) {
	type vars struct {
		component components.BaseComponent
		status    int
		err       error
	}

	testCases := []struct {
		name string

		vars vars

		want *utils.AppError
	}{
		{
			name: "should success to set App Error to the component",
			vars: vars{
				component: components.BaseComponent{
					ReqCtx: context.Background(),
				},
				status: http.StatusInternalServerError,
				err:    errors.New("some error"),
			},
			want: &utils.AppError{
				Error:  errors.New("some error"),
				Status: http.StatusInternalServerError,
			},
		},
	}

	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			// Setup
			ttc := &TopTrackComponent{
				BaseComponent: tCase.vars.component,
			}

			// Run test
			ttc.SetComponentAppError(tCase.vars.status, tCase.vars.err)

			// Assert
			assert.Exactlyf(t, tCase.want, ttc.GetComponentAppError(), "case: %v", tCase)
		})
	}
}

func TestTopTrackComponent_GetRegionalTopTrack(t *testing.T) {
	constants.COUNTRIES_JSON_FILE_NAME = "../../countries.json"

	type vars struct {
		component components.BaseComponent

		form *RegionalTopTrackForm

		headers map[string]string
	}

	testCases := []struct {
		name string

		vars vars

		want   string
		hasErr bool
		err    string
	}{
		{
			name: "should success to fetch the top track of the region",
			vars: vars{
				component: components.BaseComponent{
					ReqCtx: context.Background(),
				},
				form: &RegionalTopTrackForm{
					Country: "in",
				},
				headers: map[string]string{
					"x-mock-api": "default",
				},
			},
			want: `{ "meta": { "country": "India" }, "track": { "rank": 1, "name": "Yellow", "duration": "267", "listeners": 2531979, "url": "https://www.last.fm/music/Coldplay/_/Yellow", "artists_info": { "name": "Coldplay", "url": "https://www.last.fm/music/Coldplay", "images": [ { "#text": "https://lastfm.freetls.fastly.net/i/u/34s/2a96cbd8b46e442fc41c2b86b821562f.png", "size": "small" }, { "#text": "https://lastfm.freetls.fastly.net/i/u/64s/2a96cbd8b46e442fc41c2b86b821562f.png", "size": "medium" }, { "#text": "https://lastfm.freetls.fastly.net/i/u/174s/2a96cbd8b46e442fc41c2b86b821562f.png", "size": "large" }, { "#text": "https://lastfm.freetls.fastly.net/i/u/300x300/2a96cbd8b46e442fc41c2b86b821562f.png", "size": "extralarge" }, { "#text": "https://lastfm.freetls.fastly.net/i/u/300x300/2a96cbd8b46e442fc41c2b86b821562f.png", "size": "mega" }, { "#text": "https://lastfm.freetls.fastly.net/i/u/300x300/2a96cbd8b46e442fc41c2b86b821562f.png", "size": "" } ], "summary": "Coldplay is a British alternative rock and britpop band formed in London in 1997. They consist of vocalist and pianist Chris Martin, guitarist Jonny Buckland, bassist Guy Berryman, drummer Will Champion and creative director Phil Harvey. They met at University College London and began playing music together from 1996 to 1998, initially calling themselves Starfish. Coldplay's music incorporates elements of soft rock, pop rock, piano rock, and post-britpop. \u003ca href=\"https://www.last.fm/music/Coldplay\"\u003eRead more on Last.fm\u003c/a\u003e", "Stats": { "listeners": 7296979, "play_count": 554561574 } }, "lyrics": "" }, "TrackSuggestion": [ { "name": "The Scientist", "match": 1, "duration": 309, "playCount": 20959436, "url": "https://www.last.fm/music/Coldplay/_/The+Scientist", "artist_info": { "name": "Coldplay", "url": "https://www.last.fm/music/Coldplay" } }, { "name": "Sparks", "match": 0.962653, "duration": 269, "playCount": 16814271, "url": "https://www.last.fm/music/Coldplay/_/Sparks", "artist_info": { "name": "Coldplay", "url": "https://www.last.fm/music/Coldplay" } }, { "name": "Somewhere Only We Know", "match": 0.567258, "duration": 234, "playCount": 15265085, "url": "https://www.last.fm/music/Keane/_/Somewhere+Only+We+Know", "artist_info": { "name": "Keane", "url": "https://www.last.fm/music/Keane" } }, { "name": "Chasing Cars", "match": 0.396943, "duration": 0, "playCount": 15387887, "url": "https://www.last.fm/music/Snow+Patrol/_/Chasing+Cars", "artist_info": { "name": "Snow Patrol", "url": "https://www.last.fm/music/Snow+Patrol" } }, { "name": "Iris", "match": 0.394981, "duration": 289, "playCount": 10365653, "url": "https://www.last.fm/music/Goo+Goo+Dolls/_/Iris", "artist_info": { "name": "Goo Goo Dolls", "url": "https://www.last.fm/music/Goo+Goo+Dolls" } } ] }`,
		},
		{
			name: "should fail to fetch the top track of the region",
			vars: vars{
				component: components.BaseComponent{
					ReqCtx: context.Background(),
				},
				form: &RegionalTopTrackForm{
					Country: "in",
				},
				headers: map[string]string{
					"x-mock-api": "error_response",
				},
			},
			hasErr: true,
			err:    "error",
		},
	}

	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			// Setup
			form := tCase.vars.form
			ttc := &TopTrackComponent{
				BaseComponent: tCase.vars.component,
			}
			ctx := ttc.ReqCtx
			ctx = context.WithValue(ctx, "x-mock-headers", tCase.vars.headers)
			ttc.ReqCtx = ctx

			// Run test
			got, err := ttc.GetRegionalTopTrack(form)

			// Assert
			if tCase.hasErr {
				if assert.Errorf(t, err, "case: %v", tCase) {
					assert.Containsf(t, err.Error(), tCase.err, "case: %v", tCase)
				}
			} else {
				assert.NoErrorf(t, err, "case: %v", tCase)
				tempWant := new(RegionalTopTrackResponse)
				_ = json.Unmarshal([]byte(tCase.want), tempWant)
				assert.Equal(t, tempWant, got, "case: %v", tCase)
			}
		})
	}

}
