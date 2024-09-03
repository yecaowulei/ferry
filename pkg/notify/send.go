package notify

import (
	"bytes"
	"ferry/models/system"
	"ferry/pkg/logger"
	"ferry/pkg/notify/email"
	"ferry/pkg/notify/wecom"
	"fmt"
	"text/template"

	"github.com/spf13/viper"
)

/*
  @Author : lanyulei
  @同时发送多种通知方式
*/

type BodyData struct {
	SendTo         interface{} // 接受人
	EmailCcTo      []string    // 抄送人邮箱列表
	Subject        string      // 标题
	Classify       []int       // 通知类型
	Id             int         // 工单ID
	Title          string      // 工单标题
	Creator        string      // 工单创建人
	Priority       int         // 工单优先级
	PriorityValue  string      // 工单优先级
	CreatedAt      string      // 工单创建时间
	Content        string      // 通知的内容
	Description    string      // 表格上面的描述信息
	ProcessId      int         // 流程ID
	Domain         string      // 域名地址
	CurrentProcess string      // 当前流程
	Env            string      // 上线环境
	PlanTime       string      // 计划上线时间
}

func (b *BodyData) ParsingTemplate() (err error) {
	// 读取模版数据
	var (
		buf bytes.Buffer
	)

	tmpl, err := template.ParseFiles("./static/template/email.html")
	if err != nil {
		return
	}

	b.Domain = viper.GetString("settings.domain.url")
	err = tmpl.Execute(&buf, b)
	if err != nil {
		return
	}

	b.Content = buf.String()

	return
}

func (b *BodyData) SendNotify() (err error) {
	var (
		emailList    []string
		usernameList string
		message      string
	)

	switch b.Priority {
	case 1:
		b.PriorityValue = "正常"
	case 2:
		b.PriorityValue = "紧急"
	case 3:
		b.PriorityValue = "非常紧急"
	}

	users := b.SendTo.(map[string]interface{})["userList"].([]system.SysUser)

	for _, c := range b.Classify {
		switch c {
		case 1: // 邮件
			if len(users) > 0 {
				for _, user := range users {
					emailList = append(emailList, user.Email)
				}
				err = b.ParsingTemplate()
				if err != nil {
					logger.Errorf("模版内容解析失败，%v", err.Error())
					return
				}
				go email.SendMail(emailList, b.EmailCcTo, b.Subject, b.Content)
			}
		case 2: // 企微
			if len(users) > 0 {
				for _, user := range users {
					usernameList = usernameList + fmt.Sprintf("<@%s>", user.Username)
				}

				orderUrl := fmt.Sprintf("%s/#/process/handle-ticket?workOrderId=%d&processId=%d", viper.GetString("settings.domain.url"), b.Id, b.ProcessId)
				if b.Env != "" && b.PlanTime != "" {
					message = fmt.Sprintf("## %s\n标题：%s\n申请人：%s\n优先级：%s\n申请时间：%s\n当前流程：%s\n上线环境：%s\n计划上线时间：%s\n%s\n[点击此处跳转到工单详情](%s)", b.Subject, b.Title, b.Creator, b.PriorityValue, b.CreatedAt, b.CurrentProcess, b.Env, b.PlanTime, usernameList, orderUrl)
				} else {
					message = fmt.Sprintf("## %s\n标题：%s\n申请人：%s\n优先级：%s\n申请时间：%s\n当前流程：%s\n%s\n[点击此处跳转到工单详情](%s)", b.Subject, b.Title, b.Creator, b.PriorityValue, b.CreatedAt, b.CurrentProcess, usernameList, orderUrl)
				}

				go wecom.SendWeCom(message)
			}
		}
	}
	return
}
