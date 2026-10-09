package handlers

import (
	"slb-api/server/http/handlers/wechat"
	"slb-api/utils"

	"github.com/gin-gonic/gin"
)

// Handler 根路由处理器
type Handler struct {
	router *gin.Engine
}

// NewHandler 创建根路由处理器
func NewHandler(router *gin.Engine) utils.HttpServerHandler {
	h := &Handler{router: router}
	h.RegisterRoutes()
	return h
}

// RegisterRoutes 注册所有路由
func (h *Handler) RegisterRoutes() {
	g := h.router.Group("/api")

	wechat.NewWechatUserHandler(h.router).RegisterRoutes(g)
}
