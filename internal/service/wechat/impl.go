package wechat

import (
	"slb-api/internal/dao"
	"slb-api/utils"

	"gorm.io/gorm"
)

type WechatUserService struct {
	db *gorm.DB
}

func NewWechatUserService() *WechatUserService {
	s := &WechatUserService{db: utils.DB()}
	dao.SetDefault(utils.DB())
	return s
}
