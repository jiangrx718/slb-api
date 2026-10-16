package reservoir

import (
	"context"
	"fmt"
	"strings"

	"slb-api/internal/common"
	"slb-api/internal/dao"
	"slb-api/model"
	"slb-api/utils"
)

// 接口 slb-type 取值，对应库表
const (
	TypeAttr      = "attr"      // 特征属性 reservoir_feature_attr
	TypeCapacity  = "capacity"  // 库容曲线 reservoir_capacity_curve
	TypeDischarge = "discharge" // 泄流曲线 reservoir_discharge_curve
)

var validTypes = map[string]struct{}{
	TypeAttr:      {},
	TypeCapacity:  {},
	TypeDischarge: {},
}

type QueryResponseData struct {
	List []StationData `json:"list"`
}

// StationData 单个站点的查询结果，attr/capacity/discharge 三个字段固定返回，无数据时为空数组
type StationData struct {
	Stcd      string               `json:"stcd"`
	Attr      []FeatureAttrItem    `json:"attr"`
	Capacity  []CapacityCurveItem  `json:"capacity"`
	Discharge []DischargeCurveItem `json:"discharge"`
}

// FeatureAttrItem 字段固定返回，无内容时以零值（空字符串/0）兜底
type FeatureAttrItem struct {
	RegionCode string `json:"region_code"`
	FileName   string `json:"file_name"`
	SourceFile string `json:"source_file"`
	TableIndex int16  `json:"table_index"`
	Category   string `json:"category"`
	AttrName   string `json:"attr_name"`
	AttrValue  string `json:"attr_value"`
	Unit       string `json:"unit"`
	Remark     string `json:"remark"`
}

// CapacityCurveItem 字段固定返回，无内容时以零值（空字符串/0）兜底
type CapacityCurveItem struct {
	RegionCode   string  `json:"region_code"`
	FileName     string  `json:"file_name"`
	SourceFile   string  `json:"source_file"`
	TableIndex   int16   `json:"table_index"`
	Seq          int     `json:"seq"`
	WaterLevel   float64 `json:"water_level"`
	Capacity     float64 `json:"capacity"`
	CapacityUnit string  `json:"capacity_unit"`
}

// DischargeCurveItem 字段固定返回，无内容时以零值（空字符串/0）兜底
type DischargeCurveItem struct {
	RegionCode    string  `json:"region_code"`
	FileName      string  `json:"file_name"`
	SourceFile    string  `json:"source_file"`
	TableIndex    int16   `json:"table_index"`
	Seq           int     `json:"seq"`
	WaterLevel    float64 `json:"water_level"`
	Discharge     float64 `json:"discharge"`
	DischargeUnit string  `json:"discharge_unit"`
}

func toFeatureAttrItem(m *model.ReservoirFeatureAttr) FeatureAttrItem {
	return FeatureAttrItem{
		RegionCode: m.RegionCode,
		FileName:   m.FileName,
		SourceFile: m.SourceFile,
		TableIndex: m.TableIndex,
		Category:   derefString(m.Category),
		AttrName:   m.AttrName,
		AttrValue:  derefString(m.AttrValue),
		Unit:       derefString(m.Unit),
		Remark:     derefString(m.Remark),
	}
}

func toCapacityCurveItem(m *model.ReservoirCapacityCurve) CapacityCurveItem {
	return CapacityCurveItem{
		RegionCode:   m.RegionCode,
		FileName:     m.FileName,
		SourceFile:   m.SourceFile,
		TableIndex:   m.TableIndex,
		Seq:          m.Seq,
		WaterLevel:   m.WaterLevel,
		Capacity:     m.Capacity,
		CapacityUnit: m.CapacityUnit,
	}
}

func toDischargeCurveItem(m *model.ReservoirDischargeCurve) DischargeCurveItem {
	return DischargeCurveItem{
		RegionCode:    m.RegionCode,
		FileName:      m.FileName,
		SourceFile:    m.SourceFile,
		TableIndex:    m.TableIndex,
		Seq:           m.Seq,
		WaterLevel:    m.WaterLevel,
		Discharge:     m.Discharge,
		DischargeUnit: m.DischargeUnit,
	}
}

func derefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// normalizeRequest 去重、去空白并校验请求参数
// types 为空（未传或空数组）时默认查询全部三类数据
func normalizeRequest(stcd []string, types []string) ([]string, []string, error) {
	codes := make([]string, 0, len(stcd))
	seenCode := make(map[string]struct{}, len(stcd))
	for _, code := range stcd {
		code = strings.TrimSpace(code)
		if code == "" {
			continue
		}
		if _, ok := seenCode[code]; ok {
			continue
		}
		seenCode[code] = struct{}{}
		codes = append(codes, code)
	}
	if len(codes) == 0 {
		return nil, nil, fmt.Errorf("站点编号 stcd 不能为空")
	}

	if len(types) == 0 {
		return codes, []string{TypeAttr, TypeCapacity, TypeDischarge}, nil
	}

	wanted := make([]string, 0, len(types))
	seenType := make(map[string]struct{}, len(types))
	for _, t := range types {
		t = strings.TrimSpace(t)
		if _, ok := validTypes[t]; !ok {
			return nil, nil, fmt.Errorf("无效的数据类型: %s，可选值: attr、capacity、discharge", t)
		}
		if _, ok := seenType[t]; ok {
			continue
		}
		seenType[t] = struct{}{}
		wanted = append(wanted, t)
	}

	return codes, wanted, nil
}

func (s *ReservoirService) Query(ctx context.Context, stcd []string, types []string) (common.ServiceResult, error) {
	var (
		logger = utils.SugarContext(ctx)
		result = common.NewServiceResult()
	)

	codes, wanted, err := normalizeRequest(stcd, types)
	if err != nil {
		result.SetError(&common.ServiceError{Code: 400, Message: err.Error()})
		return result, nil
	}

	// 按请求顺序组织站点结果，三个字段初始化为空数组保证固定返回
	stations := make([]*StationData, 0, len(codes))
	index := make(map[string]*StationData, len(codes))
	for _, code := range codes {
		station := &StationData{
			Stcd:      code,
			Attr:      make([]FeatureAttrItem, 0),
			Capacity:  make([]CapacityCurveItem, 0),
			Discharge: make([]DischargeCurveItem, 0),
		}
		stations = append(stations, station)
		index[code] = station
	}

	for _, t := range wanted {
		switch t {
		case TypeAttr:
			featureDao := dao.ReservoirFeatureAttr
			rows, err := featureDao.
				Where(featureDao.StationCode.In(codes...)).
				Order(featureDao.StationCode, featureDao.TableIndex, featureDao.Id).
				Find()
			if err != nil {
				logger.Errorw("ReservoirService Query featureAttr Find error", "error", err)
				return result, err
			}
			for _, row := range rows {
				if station, ok := index[row.StationCode]; ok {
					station.Attr = append(station.Attr, toFeatureAttrItem(row))
				}
			}
		case TypeCapacity:
			capacityDao := dao.ReservoirCapacityCurve
			rows, err := capacityDao.
				Where(capacityDao.StationCode.In(codes...)).
				Order(capacityDao.StationCode, capacityDao.TableIndex, capacityDao.Seq).
				Find()
			if err != nil {
				logger.Errorw("ReservoirService Query capacityCurve Find error", "error", err)
				return result, err
			}
			for _, row := range rows {
				if station, ok := index[row.StationCode]; ok {
					station.Capacity = append(station.Capacity, toCapacityCurveItem(row))
				}
			}
		case TypeDischarge:
			dischargeDao := dao.ReservoirDischargeCurve
			rows, err := dischargeDao.
				Where(dischargeDao.StationCode.In(codes...)).
				Order(dischargeDao.StationCode, dischargeDao.TableIndex, dischargeDao.Seq).
				Find()
			if err != nil {
				logger.Errorw("ReservoirService Query dischargeCurve Find error", "error", err)
				return result, err
			}
			for _, row := range rows {
				if station, ok := index[row.StationCode]; ok {
					station.Discharge = append(station.Discharge, toDischargeCurveItem(row))
				}
			}
		}
	}

	list := make([]StationData, 0, len(stations))
	for _, station := range stations {
		list = append(list, *station)
	}

	result.Data = QueryResponseData{
		List: list,
	}
	result.SetMessage("操作成功")
	return result, nil
}
