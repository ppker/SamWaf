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

type WafLdpUrlService struct{}

var WafLdpUrlServiceApp = new(WafLdpUrlService)

func (receiver *WafLdpUrlService) AddApi(req request.WafLdpUrlAddReq) error {
	var bean = &model.LDPUrl{
		BaseOrm: baseorm.BaseOrm{
			Id:          uuid.GenUUID(),
			USER_CODE:   global.GWAF_USER_CODE,
			Tenant_ID:   global.GWAF_TENANT_ID,
			CREATE_TIME: customtype.JsonTime(time.Now()),
			UPDATE_TIME: customtype.JsonTime(time.Now()),
		},
		HostCode:    req.HostCode,
		CompareType: req.CompareType,
		Url:         req.Url,
		Remarks:     req.Remarks,
	}
	global.GWAF_LOCAL_DB.Create(bean)
	return nil
}

func (receiver *WafLdpUrlService) CheckIsExistApi(req request.WafLdpUrlAddReq) error {
	return global.GWAF_LOCAL_DB.First(&model.LDPUrl{}, "host_code = ? and url= ?", req.HostCode,
		req.Url).Error
}
func (receiver *WafLdpUrlService) ModifyApi(req request.WafLdpUrlEditReq) error {
	var ipWhite model.LDPUrl
	global.GWAF_LOCAL_DB.Where("host_code = ? and url= ?", req.HostCode,
		req.Url).Find(&ipWhite)
	if ipWhite.Id != "" && ipWhite.Url != req.Url {
		return errors.New("当前网站和url已经存在")
	}
	ipWhiteMap := map[string]interface{}{
		"Host_Code":    req.HostCode,
		"Url":          req.Url,
		"Remarks":      req.Remarks,
		"Compare_Type": req.CompareType,
		"UPDATE_TIME":  customtype.JsonTime(time.Now()),
	}
	err := global.GWAF_LOCAL_DB.Model(model.LDPUrl{}).Where("id = ?", req.Id).Updates(ipWhiteMap).Error

	return err
}
func (receiver *WafLdpUrlService) GetDetailApi(req request.WafLdpUrlDetailReq) model.LDPUrl {
	var bean model.LDPUrl
	global.GWAF_LOCAL_DB.Where("id=?", req.Id).Find(&bean)
	return bean
}
func (receiver *WafLdpUrlService) GetDetailByIdApi(id string) model.LDPUrl {
	var bean model.LDPUrl
	global.GWAF_LOCAL_DB.Where("id=?", id).Find(&bean)
	return bean
}
func (receiver *WafLdpUrlService) GetListApi(req request.WafLdpUrlSearchReq) ([]model.LDPUrl, int64, error) {
	var list []model.LDPUrl
	var total int64 = 0

	/*where条件*/
	var whereField = ""
	var whereValues []interface{}
	//where字段
	whereField = ""
	if len(req.HostCode) > 0 {
		if len(whereField) > 0 {
			whereField = whereField + " and "
		}
		whereField = whereField + " host_code=? "
	}
	if len(req.Url) > 0 {
		if len(whereField) > 0 {
			whereField = whereField + " and "
		}
		whereField = whereField + " url =? "
	}
	//where字段赋值
	if len(req.HostCode) > 0 {
		whereValues = append(whereValues, req.HostCode)
	}
	if len(req.Url) > 0 {
		whereValues = append(whereValues, req.Url)
	}

	global.GWAF_LOCAL_DB.Model(&model.LDPUrl{}).Where(whereField, whereValues...).Limit(req.PageSize).Offset(req.PageSize * (req.PageIndex - 1)).Find(&list)
	global.GWAF_LOCAL_DB.Model(&model.LDPUrl{}).Where(whereField, whereValues...).Count(&total)

	return list, total, nil
}
func (receiver *WafLdpUrlService) DelApi(req request.WafLdpUrlDelReq) error {
	var ipWhite model.LDPUrl
	err := global.GWAF_LOCAL_DB.Where("id = ?", req.Id).First(&ipWhite).Error
	if err != nil {
		return err
	}
	err = global.GWAF_LOCAL_DB.Where("id = ?", req.Id).Delete(model.LDPUrl{}).Error
	return err
}
