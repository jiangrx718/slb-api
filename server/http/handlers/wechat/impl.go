package wechat

import (
	wechatUserService "slb-api/internal/service/wechat"

	"github.com/gin-gonic/gin"
)

type HomeHandler struct{}

func NewWechatUserHandler(engine *gin.Engine) *WechatUserHandler {
	return &WechatUserHandler{
		engine:  engine,
		service: wechatUserService.NewWechatUserService(),
	}
}

type WechatUserHandler struct {
	engine  *gin.Engine
	service wechatUserService.WechatUserServiceIFace
}

func (h *WechatUserHandler) RegisterRoutes(routerGroup *gin.RouterGroup) {
	g := routerGroup.Group("/wechat-user")
	g.GET("/list", h.List)
}
