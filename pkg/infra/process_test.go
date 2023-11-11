package infra_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

type Message struct {
	PW   string   `json:"PassWord"`
	Args []string `json:"Args"`
}

func TestServer(t *testing.T) {

	r := gin.Default()
	gin.SetMode(gin.ReleaseMode)
	data := map[string]interface{}{
		"PassWord": "123",
		"Args":     []string{"TransferAsset"},
	}
	PostJson("http://localhost:8080/invoke", data, r)

}

func PostJson(uri string, data map[string]interface{}, router *gin.Engine) {

	// 将JSON数据转换为字节切片
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}

	// 设置HTTP请求
	url := "http://localhost:8080/invoke"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发起HTTP请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()

	// 处理响应
	fmt.Println("Response Status:", resp)
	// 这里你可以根据需要读取和处理响应的内容
}
