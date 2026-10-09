package reservoir

import (
	"slb-api/internal/dao"
	"slb-api/utils"

	"gorm.io/gorm"
)

type ReservoirService struct {
	db *gorm.DB
}

func NewReservoirService() *ReservoirService {
	s := &ReservoirService{db: utils.DB()}
	dao.SetDefault(utils.DB())
	return s
}
