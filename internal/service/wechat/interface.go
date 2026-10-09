package wechat

import (
	"context"
	"slb-api/internal/common"
)

type WechatUserServiceIFace interface {
	// List 列表
	List(ctx context.Context) (common.ServiceResult, error)
}
