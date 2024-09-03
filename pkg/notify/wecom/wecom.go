package wecom

/*
  @Author : linanfei
  @Desc : 发送企微通知
*/

import (
	"bytes"
	"encoding/json"
	"errors"
	"ferry/pkg/logger"
	"fmt"
	"github.com/spf13/viper"
	"net/http"
)

type MarkdownMessage struct {
	MsgType  string `json:"msgtype"`
	Markdown struct {
		Content string `json:"content"`
	} `json:"markdown"`
}

func server(message string) error {
	mdMessage := MarkdownMessage{
		MsgType: "markdown",
	}
	mdMessage.Markdown.Content = message

	body, err := json.Marshal(mdMessage)
	if err != nil {
		return errors.New(fmt.Sprintf("marshaling message: %v", err))
	}

	webhookURL := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=%s", viper.GetString("settings.wecom.wehookkey"))
	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(body))
	if err != nil {
		return errors.New(fmt.Sprintf("creating http request: %v", err))
	}

	req.Header.Set("Content-Type", "application/json")

	// 发送HTTP请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.New(fmt.Sprintf("sending http request: %v", err))
	}
	defer resp.Body.Close()

	// 检查HTTP响应状态
	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("request failed with status: %d", resp.StatusCode))
	}
	return err

}

func SendWeCom(message string) {
	err := server(message)
	if err != nil {
		logger.Info(err)
		return
	}
	logger.Info("send wecom successfully")
}
