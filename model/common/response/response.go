package response

import (
	"SamWaf/global"
	"SamWaf/wafsec"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Response struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}

const (
	ERROR             = -1
	SUCCESS           = 0
	INPUT_SECRET_CODE = -2
	NEED_BIND_2FA     = -3
	AUTHFAIL          = -999
)

func Result(code int, data interface{}, msg string, c *gin.Context) {
	result, _ := json.Marshal(data) //将数据转换为json
	encryptStr, _ := wafsec.AesEncrypt(result, global.GWAF_COMMUNICATION_KEY)
	// 开始时间
	c.JSON(http.StatusOK, Response{
		code,
		encryptStr,
		msg,
	})
}

func Ok(c *gin.Context) {
	Result(SUCCESS, map[string]interface{}{}, "操作成功", c)
}

func OkWithMessage(message string, c *gin.Context) {
	Result(SUCCESS, map[string]interface{}{}, message, c)
}

func OkWithData(data interface{}, c *gin.Context) {
	Result(SUCCESS, data, "查询成功", c)
}

func OkWithDetailed(data interface{}, message string, c *gin.Context) {
	Result(SUCCESS, data, message, c)
}

func Fail(c *gin.Context) {
	Result(ERROR, map[string]interface{}{}, "操作失败", c)
}

func FailWithMessage(message string, c *gin.Context) {
	Result(ERROR, map[string]interface{}{}, message, c)
}

func FailWithDetailed(data interface{}, message string, c *gin.Context) {
	Result(ERROR, data, message, c)
}
func AuthFailWithMessage(message string, c *gin.Context) {
	Result(AUTHFAIL, map[string]interface{}{}, message, c)
}
func SecretCodeFailWithMessage(message string, c *gin.Context) {
	Result(INPUT_SECRET_CODE, map[string]interface{}{}, message, c)
}
func NeedBind2FAWithMessage(message string, c *gin.Context) {
	Result(NEED_BIND_2FA, map[string]interface{}{}, message, c)
}
