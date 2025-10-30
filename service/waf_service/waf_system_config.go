package waf_service

import (
	"SamWaf/common/uuid"
	"SamWaf/customtype"
	"SamWaf/global"
	"SamWaf/model"
	"SamWaf/model/baseorm"
	"SamWaf/model/request"
	"errors"
	"time"
)

type WafSystemConfigService struct{}

var WafSystemConfigServiceApp = new(WafSystemConfigService)

func (receiver *WafSystemConfigService) AddApi(wafSystemConfigAddReq request.WafSystemConfigAddReq) error {
	var bean = &model.SystemConfig{
		BaseOrm: baseorm.BaseOrm{
			Id:          uuid.GenUUID(),
			USER_CODE:   global.GWAF_USER_CODE,
			Tenant_ID:   global.GWAF_TENANT_ID,
			CREATE_TIME: customtype.JsonTime(time.Now()),
			UPDATE_TIME: customtype.JsonTime(time.Now()),
		},
		ItemClass: wafSystemConfigAddReq.ItemClass,
		Item:      wafSystemConfigAddReq.Item,
		Value:     wafSystemConfigAddReq.Value,
		IsSystem:  "0",
		Remarks:   wafSystemConfigAddReq.Remarks,
		HashInfo:  "",
	}
	global.GWAF_LOCAL_DB.Create(bean)
	return nil
}

func (receiver *WafSystemConfigService) CheckIsExistApi(wafSystemConfigAddReq request.WafSystemConfigAddReq) error {
	return global.GWAF_LOCAL_DB.First(&model.SystemConfig{}, "item = ? ", wafSystemConfigAddReq.Item).Error
}
func (receiver *WafSystemConfigService) ModifyApi(req request.WafSystemConfigEditReq) error {
	var sysConfig model.SystemConfig
	global.GWAF_LOCAL_DB.Where("id = ?", req.Id).Find(&sysConfig)
	if req.Id != "" && req.Item != req.Item {
		return errors.New("当前配置已经存在")
	}
	editMap := map[string]interface{}{
		"Item":        req.Item,
		"ItemClass":   req.ItemClass,
		"Value":       req.Value,
		"Remarks":     req.Remarks,
		"ItemType":    req.ItemType,
		"Options":     req.Options,
		"UPDATE_TIME": customtype.JsonTime(time.Now()),
	}

	err := global.GWAF_LOCAL_DB.Model(model.SystemConfig{}).Where("id = ?", req.Id).Updates(editMap).Error

	return err
}
func (receiver *WafSystemConfigService) GetDetailApi(req request.WafSystemConfigDetailReq) model.SystemConfig {
	var bean model.SystemConfig
	global.GWAF_LOCAL_DB.Where("id=?", req.Id).Find(&bean)
	return bean
}
func (receiver *WafSystemConfigService) GetDetailByItemApi(req request.WafSystemConfigDetailByItemReq) model.SystemConfig {
	var bean model.SystemConfig
	global.GWAF_LOCAL_DB.Where("Item=?", req.Item).Find(&bean)
	return bean
}
func (receiver *WafSystemConfigService) GetDetailByIdApi(id string) model.SystemConfig {
	var bean model.SystemConfig
	global.GWAF_LOCAL_DB.Where("id=?", id).Find(&bean)
	return bean
}
func (receiver *WafSystemConfigService) GetDetailByItem(item string) model.SystemConfig {
	var bean model.SystemConfig
	global.GWAF_LOCAL_DB.Where("Item=?", item).Find(&bean)
	return bean
}
func (receiver *WafSystemConfigService) GetListApi(req request.WafSystemConfigSearchReq) ([]model.SystemConfig, int64, error) {
	var list []model.SystemConfig
	var total int64 = 0
	/*where条件*/
	var whereField = ""
	var whereValues []interface{}
	//where字段
	whereField = ""
	if len(req.Item) > 0 {
		if len(whereField) > 0 {
			whereField = whereField + " and "
		}
		whereField = whereField + " item=? "
	}
	if len(req.Remarks) > 0 {
		if len(whereField) > 0 {
			whereField = whereField + " and "
		}
		whereField = whereField + " remarks like ? "
	}
	//where字段赋值
	if len(req.Item) > 0 {
		whereValues = append(whereValues, req.Item)
	}
	if len(req.Remarks) > 0 {
		whereValues = append(whereValues, "%"+req.Remarks+"%")
	}
	global.GWAF_LOCAL_DB.Model(&model.SystemConfig{}).Where(whereField, whereValues...).Limit(req.PageSize).Offset(req.PageSize * (req.PageIndex - 1)).Find(&list)
	global.GWAF_LOCAL_DB.Model(&model.SystemConfig{}).Where(whereField, whereValues...).Count(&total)

	return list, total, nil
}
func (receiver *WafSystemConfigService) DelApi(req request.WafSystemConfigDelReq) error {
	var bean model.SystemConfig
	err := global.GWAF_LOCAL_DB.Where("id = ? and is_system=0", req.Id).First(&bean).Error
	if err != nil {
		return err
	}
	err = global.GWAF_LOCAL_DB.Where("id = ? and is_system=0", req.Id).Delete(model.SystemConfig{}).Error
	return err
}

// GetAllConfigs 批量获取所有配置项，返回以item为key的map
func (receiver *WafSystemConfigService) GetAllConfigs() map[string]model.SystemConfig {
	var configs []model.SystemConfig
	configMap := make(map[string]model.SystemConfig)

	// 一次性查询所有配置
	global.GWAF_LOCAL_DB.Find(&configs)

	// 构建以item为key的map
	for _, config := range configs {
		configMap[config.Item] = config
	}

	return configMap
}
