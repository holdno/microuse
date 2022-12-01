package wechat_test

import (
	"context"
	"testing"
	"time"

	"github.com/holdno/microuse/srv/wechat"
	"github.com/holdno/microuse/srv/wechat/protobuf/micro/wechatpb"
)

const (
	endpoint = "micro.jihe.pro:443"
	appid    = ""
	secret   = ""
	openid   = ""
)

func Test_SendTplMessage(t *testing.T) {
	client, err := wechat.NewUserClient(endpoint, appid, secret, time.Second*3)
	if err != nil {
		t.Fatal(err)
	}

	pushData := map[string]*wechatpb.TplMessageItem{
		"first": &wechatpb.TplMessageItem{
			Value: "服务提醒",
			Color: "",
		},
		"keyword1": {
			Value: "测试",
		},
		"keyword2": {
			Value: "测试",
		},
		"remark": {
			Value: "请尽快查看。",
		},
	}

	if _, err = client.SendTplMessage(context.Background(), &wechatpb.SendTplMessageRequest{
		Openid: openid,
		Tpl:    "skabs_XNXbESLAzS1m62Fw_UEIpDLtfBlqf-agG8ljg",
		Data:   pushData,
	}); err != nil {
		t.Fatal(err)
	}
	t.Log("success")
}
