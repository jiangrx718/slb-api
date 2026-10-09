package model

import "time"

// ReservoirFeatureAttr 水库特征属性(key-value)
type ReservoirFeatureAttr struct {
	Id           uint64    `gorm:"column:id;type:bigint unsigned;primary_key;auto_increment;comment:主键" json:"id"`
	StationCode  string    `gorm:"column:station_code;type:varchar(32);comment:站点编号;NOT NULL" json:"station_code"`
	RegionCode   string    `gorm:"column:region_code;type:varchar(32);comment:区域编号;NOT NULL" json:"region_code"`
	FileName     string    `gorm:"column:file_name;type:varchar(512);comment:来源文件名称;NOT NULL" json:"file_name"`
	SourceFile   string    `gorm:"column:source_file;type:varchar(16);comment:来源json: aerial / capacity;NOT NULL" json:"source_file"`
	TableIndex   int16     `gorm:"column:table_index;type:smallint;comment:表格在json中的序号;NOT NULL;default:0" json:"table_index"`
	Category     *string   `gorm:"column:category;type:varchar(255);comment:属性分类" json:"category"`
	AttrName     string    `gorm:"column:attr_name;type:varchar(255);comment:属性名(键);NOT NULL" json:"attr_name"`
	AttrValue    *string   `gorm:"column:attr_value;type:text;comment:属性值" json:"attr_value"`
	Unit         *string   `gorm:"column:unit;type:varchar(64);comment:单位" json:"unit"`
	Remark       *string   `gorm:"column:remark;type:varchar(512);comment:备注" json:"remark"`
	TableContext *string   `gorm:"column:table_context;type:varchar(512);comment:表格上文" json:"table_context"`
	CreatedAt    time.Time `gorm:"column:created_at;type:datetime;comment:入库时间;NOT NULL;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (m *ReservoirFeatureAttr) TableName() string {
	return "reservoir_feature_attr"
}
