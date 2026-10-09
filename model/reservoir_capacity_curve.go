package model

import "time"

// ReservoirCapacityCurve 库容曲线(水位-库容)
type ReservoirCapacityCurve struct {
	Id           uint64    `gorm:"column:id;type:bigint unsigned;primary_key;auto_increment;comment:主键" json:"id"`
	StationCode  string    `gorm:"column:station_code;type:varchar(32);comment:站点编号;NOT NULL" json:"station_code"`
	RegionCode   string    `gorm:"column:region_code;type:varchar(32);comment:区域编号;NOT NULL" json:"region_code"`
	FileName     string    `gorm:"column:file_name;type:varchar(512);comment:来源文件名称;NOT NULL" json:"file_name"`
	SourceFile   string    `gorm:"column:source_file;type:varchar(16);comment:来源json: aerial / capacity;NOT NULL" json:"source_file"`
	TableIndex   int16     `gorm:"column:table_index;type:smallint;comment:表格在json中的序号;NOT NULL;default:0" json:"table_index"`
	Seq          int       `gorm:"column:seq;type:int(11);comment:曲线点在表内序号;NOT NULL;default:0" json:"seq"`
	WaterLevel   float64   `gorm:"column:water_level;type:decimal(10,3);comment:水位(m);NOT NULL" json:"water_level"`
	Capacity     float64   `gorm:"column:capacity;type:decimal(14,3);comment:库容;NOT NULL" json:"capacity"`
	CapacityUnit string    `gorm:"column:capacity_unit;type:varchar(32);comment:库容单位;NOT NULL;default:'万m³'" json:"capacity_unit"`
	TableContext *string   `gorm:"column:table_context;type:varchar(512);comment:表格上文" json:"table_context"`
	CreatedAt    time.Time `gorm:"column:created_at;type:datetime;comment:入库时间;NOT NULL;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (m *ReservoirCapacityCurve) TableName() string {
	return "reservoir_capacity_curve"
}
