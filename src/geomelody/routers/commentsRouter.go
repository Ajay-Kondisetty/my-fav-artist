package routers

import (
	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context/param"
)

func init() {

	beego.GlobalControllerRouter["geomelody/controllers/track:TopTrackController"] = append(beego.GlobalControllerRouter["geomelody/controllers/track:TopTrackController"],
		beego.ControllerComments{
			Method:           "GetRegionalTopTrack",
			Router:           `/`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})
}
