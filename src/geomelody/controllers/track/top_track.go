package track

import (
	"encoding/json"
	"log"
	"net/http"

	"geomelody/components/track"
	"geomelody/controllers"
	"geomelody/utils"
)

type TopTrackController struct {
	controllers.BaseController
	Component track.TopTrack
}

// UpdateComponent is used to update the component object.
func (c *TopTrackController) UpdateComponent(component interface{}) {
	c.Component, _ = component.(track.TopTrack)
}

// GetRegionalTopTrack is used to retrieve the details, lyrics, and artists of the top track based on the given country or city. It also provides suggestions based on the retrieved track and artist.
// @router	/ [post]
func (c *TopTrackController) GetRegionalTopTrack() {
	var d *track.RegionalTopTrackResponse
	var err error
	var status int

	form := c.Component.GetRegionalTopTrackForm()

	if err = json.Unmarshal(c.GetRequestBody(), form); err != nil {
		status = http.StatusInternalServerError
	} else if d, err = c.Component.GetRegionalTopTrack(form); err != nil {
		status = c.Component.GetComponentAppError().Status
	}

	if err != nil {
		log.Printf("Some error occurred: %v", err)
	} else {
		status = http.StatusOK
	}

	c.Data["json"] = utils.PrepareResponse(d, err, status)
	c.AddHeaders(status, map[string]bool{"no_cache": true})
	_ = c.ServeJSON()
}
