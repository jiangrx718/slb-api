package reservoir

import (
	"context"

	"slb-api/internal/common"
)

type ReservoirServiceIFace interface {
	// Query 按站点编号与数据类型查询水库数据
	Query(ctx context.Context, stcd []string, types []string) (common.ServiceResult, error)
}
