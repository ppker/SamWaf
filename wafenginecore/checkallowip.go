package wafenginecore

import (
	"SamWaf/global"
	"SamWaf/innerbean"
	"SamWaf/model/detection"
	"SamWaf/model/wafenginmodel"
	"SamWaf/utils"
	"net/http"
	"net/url"
)

/*
*
检测白名单 ip
*/
func (waf *WafEngine) CheckAllowIP(r *http.Request, weblogbean *innerbean.WebLog, formValue url.Values, hostTarget *wafenginmodel.HostSafe, globalHostTarget *wafenginmodel.HostSafe) detection.Result {
	result := detection.Result{
		JumpGuardResult: false,
		IsBlock:         false,
		Title:           "",
		Content:         "",
	}
	//ip白名单策略（局部）
	if hostTarget.IPWhiteLists != nil {
		for i := 0; i < len(hostTarget.IPWhiteLists); i++ {
			if utils.CheckIPInCIDR(weblogbean.SRC_IP, hostTarget.IPWhiteLists[i].Ip) {
				result.JumpGuardResult = true
				break
			}
		}
	}
	//ip白名单策略（全局）
	if waf.HostTarget[global.GWAF_GLOBAL_HOST_NAME].Host.GUARD_STATUS == 1 && waf.HostTarget[global.GWAF_GLOBAL_HOST_NAME].IPWhiteLists != nil {
		for i := 0; i < len(waf.HostTarget[global.GWAF_GLOBAL_HOST_NAME].IPWhiteLists); i++ {
			if utils.CheckIPInCIDR(weblogbean.SRC_IP, waf.HostTarget[global.GWAF_GLOBAL_HOST_NAME].IPWhiteLists[i].Ip) {
				result.JumpGuardResult = true
				break
			}
		}
	}
	return result
}
