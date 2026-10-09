package wechat

import (
	"context"
	"slb-api/internal/common"
	"slb-api/internal/dao"
	"slb-api/model"
	"slb-api/utils"

	"gorm.io/gen"
)

type ListResponseData struct {
	List []WechatUserItem `json:"list"`
}

type WechatUserItem struct {
	UserName string `json:"user_name"`
	DeviceId string `json:"device_id"`
	Score    int    `json:"score"`
}

func toWechatUserItem(m *model.SWechatUser) WechatUserItem {
	return WechatUserItem{
		UserName: m.UserName,
		DeviceId: m.DeviceId,
		Score:    m.Score,
	}
}

func (s *WechatUserService) List(ctx context.Context) (common.ServiceResult, error) {
	var (
		logger = utils.SugarContext(ctx)
		result = common.NewServiceResult()
	)

	wechatUserDao := dao.SWechatUser
	where := []gen.Condition{
		wechatUserDao.Score.Gt(0),
		wechatUserDao.DeviceId.Neq(""),
	}

	list, _, err := wechatUserDao.Where(where...).Order(wechatUserDao.Score.Desc()).Debug().FindByPage(0, 10)
	if err != nil {
		logger.Errorw("WechatUserService List FindByPage error", "error", err)
		return result, err
	}

	items := make([]WechatUserItem, 0, len(list))
	for _, m := range list {
		items = append(items, toWechatUserItem(m))
	}

	result.Data = ListResponseData{
		List: items,
	}
	result.SetMessage("操作成功")
	return result, nil
}
