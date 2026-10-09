package model

type SWechatUser struct {
	BaseModelFieldId
	DeviceId string `gorm:"column:device_id;type:varchar(64);comment:设备ID;NOT NULL;default:''" json:"device_id"`
	UserName string `gorm:"column:user_name;type:varchar(128);comment:用户名;NOT NULL;default:'';index:idx_s_wechat_user_user_name" json:"user_name"`
	Score    int    `gorm:"column:score;type:int(11);default:0;comment:分值;NOT NULL" json:"score"`
	BaseModelFieldTime
}

func (m *SWechatUser) TableName() string {
	return "s_wechat_user"
}
