package model

import "time"

// ReservoirDischargeCurve 泄流曲线(水位-泄量)
type ReservoirDischargeCurve struct {
	Id            uint64    `gorm:"column:id;type:bigint unsigned;primary_key;auto_increment;comment:主键" json:"id"`
	StationCode   string    `gorm:"column:station_code;type:varchar(32);comment:站点编号;NOT NULL" json:"station_code"`
	RegionCode    string    `gorm:"column:region_code;type:varchar(32);comment:区域编号;NOT NULL" json:"region_code"`
	FileName      string    `gorm:"column:file_name;type:varchar(512);comment:来源文件名称;NOT NULL" json:"file_name"`
	SourceFile    string    `gorm:"column:source_file;type:varchar(16);comment:来源json: aerial / capacity;NOT NULL" json:"source_file"`
	TableIndex    int16     `gorm:"column:table_index;type:smallint;comment:表格在json中的序号;NOT NULL;default:0" json:"table_index"`
	Seq           int       `gorm:"column:seq;type:int(11);comment:曲线点在表内序号;NOT NULL;default:0" json:"seq"`
	WaterLevel    float64   `gorm:"column:water_level;type:decimal(10,3);comment:水位(m);NOT NULL" json:"water_level"`
	Discharge     float64   `gorm:"column:discharge;type:decimal(14,3);comment:泄量;NOT NULL" json:"discharge"`
	DischargeUnit string    `gorm:"column:discharge_unit;type:varchar(32);comment:泄量单位;NOT NULL;default:'m³/s'" json:"discharge_unit"`
	TableContext  *string   `gorm:"column:table_context;type:varchar(512);comment:表格上文" json:"table_context"`
	CreatedAt     time.Time `gorm:"column:created_at;type:datetime;comment:入库时间;NOT NULL;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (m *ReservoirDischargeCurve) TableName() string {
	return "reservoir_discharge_curve"
}
