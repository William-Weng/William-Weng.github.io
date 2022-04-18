package utility

import (
	"io/ioutil"

	"github.com/gin-gonic/gin"
)

// 輸出結果JSON
func ContextJSON(context *gin.Context, httpStatus int, result interface{}, error error) {

	context.JSON(httpStatus, gin.H{
		"error":  error,
		"result": result,
	})
}

// 讀取Request上的Body值
func RequestBody(context *gin.Context) ([]byte, error) {
	return ioutil.ReadAll(context.Request.Body)
}

// 讀取Request上的Body值 => MAP
func RequestBodyToMap(context *gin.Context) map[string]interface{} {

	rawJSON, _ := RequestBody(context)
	dictionary := RawBodyToMap(rawJSON)

	return dictionary
}
