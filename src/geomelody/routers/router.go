package routers

import (
	"fmt"

	"geomelody/constants"
	"geomelody/controllers/track"

	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
)

func InitRoutes() {
	ns := web.NewNamespace(fmt.Sprintf("/%v", constants.API_PATH),
		web.NSGet("/healthcheck", func(ctx *context.Context) {
			_ = ctx.Output.Body([]byte("i am alive"))
		}),

		web.NSNamespace("/track",
			web.NSNamespace(
				"/top-track",
				web.NSInclude(
					&track.TopTrackController{},
				),
			),
		),
	)

	web.AddNamespace(ns)
}
