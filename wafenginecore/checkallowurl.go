package wafenginecore

import (
	"SamWaf/global"
	"SamWaf/innerbean"
	"SamWaf/model/detection"
	"SamWaf/model/wafenginmodel"
	"net/http"
	"net/url"
	"strings"
)

/*
*
检测允许的URL
返回是否满足条件
*/
func (waf *WafEngine) CheckAllowURL(r *http.Request, weblogbean innerbean.WebLog, formValue url.Values, hostTarget *wafenginmodel.HostSafe, globalHostTarget *wafenginmodel.HostSafe) detection.Result {
	result := detection.Result{
		JumpGuardResult: false,
		IsBlock:         false,
		Title:           "",
		Content:         "",
	}
	//url白名单策略（局部）
	if hostTarget.UrlWhiteLists != nil {
		for i := 0; i < len(hostTarget.UrlWhiteLists); i++ {
			if (hostTarget.UrlWhiteLists[i].CompareType == "等于" && hostTarget.UrlWhiteLists[i].Url == weblogbean.URL) ||
				(hostTarget.UrlWhiteLists[i].CompareType == "前缀匹配" && strings.HasPrefix(weblogbean.URL, hostTarget.UrlWhiteLists[i].Url)) ||
				(hostTarget.UrlWhiteLists[i].CompareType == "后缀匹配" && strings.HasSuffix(weblogbean.URL, hostTarget.UrlWhiteLists[i].Url)) ||
				(hostTarget.UrlWhiteLists[i].CompareType == "包含匹配" && strings.Contains(weblogbean.URL, hostTarget.UrlWhiteLists[i].Url)) {
				result.JumpGuardResult = true
				break
			}
		}
	}
	//url白名单策略（全局）
	if waf.HostTarget[global.GWAF_GLOBAL_HOST_NAME].Host.GUARD_STATUS == 1 && waf.HostTarget[global.GWAF_GLOBAL_HOST_NAME].UrlWhiteLists != nil {
		for i := 0; i < len(waf.HostTarget[global.GWAF_GLOBAL_HOST_NAME].UrlWhiteLists); i++ {
			if (waf.HostTarget[global.GWAF_GLOBAL_HOST_NAME].UrlWhiteLists[i].CompareType == "等于" && waf.HostTarget[global.GWAF_GLOBAL_HOST_NAME].UrlWhiteLists[i].Url == weblogbean.URL) ||
				(waf.HostTarget[global.GWAF_GLOBAL_HOST_NAME].UrlWhiteLists[i].CompareType == "前缀匹配" && strings.HasPrefix(weblogbean.URL, waf.HostTarget[global.GWAF_GLOBAL_HOST_NAME].UrlWhiteLists[i].Url)) ||
				(waf.HostTarget[global.GWAF_GLOBAL_HOST_NAME].UrlWhiteLists[i].CompareType == "后缀匹配" && strings.HasSuffix(weblogbean.URL, waf.HostTarget[global.GWAF_GLOBAL_HOST_NAME].UrlWhiteLists[i].Url)) ||
				(waf.HostTarget[global.GWAF_GLOBAL_HOST_NAME].UrlWhiteLists[i].CompareType == "包含匹配" && strings.Contains(weblogbean.URL, waf.HostTarget[global.GWAF_GLOBAL_HOST_NAME].UrlWhiteLists[i].Url)) {
				result.JumpGuardResult = true
				break
			}
		}
	}
	return result
}
