package controllers

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"

	"geomelody/components"
	"geomelody/utils"

	"github.com/beego/beego/v2/server/web"
	"github.com/gomodule/redigo/redis"
)

type Preparer interface {
	UpdateComponent(interface{})
}

type BaseController struct {
	web.Controller
	ReqCtx    context.Context
	RedisConn redis.Conn
}

// Prepare is called before the http action is processes, to initialize.
func (c *BaseController) Prepare() {
	requestCtx := c.Ctx.Request.Context()
	c.ReqCtx = requestCtx
	conn, err := utils.RedisConn()
	if err != nil {
		c.Error(err)
	} else {
		c.RedisConn = conn
	}

	if app, ok := c.AppController.(Preparer); !ok {
		// do nothing
	} else if component, err := c.InitComponent(); err != nil {
		c.Error(err)
	} else {
		app.UpdateComponent(component)
	}
}

// Finish is called after the http action is processed, to clean-up
func (c *BaseController) Finish() {
	defer func(RedisConn redis.Conn) {
		err := RedisConn.Close()
		if err != nil {
			log.Printf("error closing redis connection")
		}
	}(c.RedisConn)
}

// InitComponent initializes the component whose methods needs to be called.
// It returns component function and error.
func (c *BaseController) InitComponent() (interface{}, error) {
	controller, _ := c.GetControllerAndAction()
	componentKey := strings.Replace(controller, "Controller", "", -1)

	componentFn, ok := components.ComponentMap[componentKey]
	if !ok {
		err := fmt.Errorf("failed to initialize component: %s", componentKey)
		return nil, err
	}

	base := &components.BaseComponent{
		ReqCtx:    c.ReqCtx,
		AppError:  new(utils.AppError),
		RedisConn: c.RedisConn,
	}

	return componentFn(base), nil
}

// GetRequestBody fetches the body of the incoming request.
// It returns the body byte data.
func (c *BaseController) GetRequestBody() []byte {
	body := c.Ctx.Input.RequestBody
	if decodedBytes, err := base64.StdEncoding.DecodeString(string(body)); err == nil {
		body = decodedBytes
	}

	if string(body) == "" {
		body = []byte(`{}`)
	}

	return body
}

// Error is used to stop execution, if any fatal error has occurred.
func (c *BaseController) Error(err error) {
	log.Printf("Some error occurred: %v", err)
	c.Data["json"] = utils.PrepareResponse(nil, err, http.StatusInternalServerError)
	c.Ctx.Output.SetStatus(http.StatusInternalServerError)
	_ = c.ServeJSON()
	c.StopRun() // stop controller execution immediately
}

// AddHeaders adds additional response headers.
func (c *BaseController) AddHeaders(status int, opts map[string]bool) {
	if fl, ok := opts["no_cache"]; ok && fl {
		c.Ctx.Output.Header("Cache-Control", "no-store, max-age=0")
	}

	c.Ctx.Output.SetStatus(status)
}
