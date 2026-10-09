package reservoir

import (
	reservoirService "slb-api/internal/service/reservoir"
	"slb-api/utils"
	"slb-api/utils/tool/response"

	"github.com/gin-gonic/gin"
)

// QueryRequest 查询请求
type QueryRequest struct {
	Stcd    []string `json:"stcd" binding:"required,min=1,dive,required"` // 站点编号数组
	SlbType []string `json:"slb-type"`                                    // 数据类型数组: attr / capacity / discharge，为空时默认查询全部三类
}

func (h *ReservoirHandler) Query(ctx *gin.Context) {
	var logger = utils.SugarContext(ctx)

	var req QueryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Errorw("Handler Reservoir Query ShouldBindJSON error", "error", err)
		response.ParameterError(ctx, err)
		return
	}

	result, err := h.service.Query(ctx, req.Stcd, req.SlbType)
	if err != nil {
		logger.Errorw("Handler Reservoir Query service.Query error", "error", err)
		response.InternalError(ctx)
		return
	}

	if result.GetCode() != 0 {
		response.Failed(ctx, result.GetCode(), result.GetMessage(), result.GetData())
		return
	}

	data, ok := result.GetData().(reservoirService.QueryResponseData)
	if !ok {
		response.InternalError(ctx)
		return
	}

	response.Successful(ctx, data)
}
