package components

import (
	"context"

	"geomelody/utils"
)

type BaseComponent struct {
	ReqCtx   context.Context
	AppError *utils.AppError
}

var ComponentMap = make(map[string]func(*BaseComponent) interface{})
