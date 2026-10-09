package reservoir

import (
	reservoirService "slb-api/internal/service/reservoir"

	"github.com/gin-gonic/gin"
)

func NewReservoirHandler(engine *gin.Engine) *ReservoirHandler {
	return &ReservoirHandler{
		engine:  engine,
		service: reservoirService.NewReservoirService(),
	}
}

type ReservoirHandler struct {
	engine  *gin.Engine
	service reservoirService.ReservoirServiceIFace
}

func (h *ReservoirHandler) RegisterRoutes(routerGroup *gin.RouterGroup) {
	g := routerGroup.Group("/reservoir")
	g.POST("/query", h.Query)
}
