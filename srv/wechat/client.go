package wechat

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/holdno/microuse/etc/spacegrower/meta"
	"github.com/holdno/microuse/srv/wechat/protobuf/micro/wechatpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	"github.com/holdno/microuse/pkg/security"
)

type WechatClient struct {
	appid  string
	secret string

	meta   meta.MetaClient
	wechat wechatpb.SrvClient

	mode    string
	timeout time.Duration

	token token
}

type token struct {
	expireTime time.Time
	token      string
}

const (
	userMode   = "user"
	systemMode = "system"
)

// "micro.jihe.pro:443"
func NewUserClient(endpoint, appid, secret string, dialTimeout time.Duration) (*WechatClient, error) {
	wc, err := newClient(endpoint, appid, secret, dialTimeout)
	if err != nil {
		return nil, err
	}
	wc.mode = userMode
	return wc, nil
}

func NewSystemClient(endpoint, appid, secret string, dialTimeout time.Duration) (*WechatClient, error) {
	wc, err := newClient(endpoint, appid, secret, dialTimeout)
	if err != nil {
		return nil, err
	}
	wc.mode = systemMode
	return wc, nil
}

func newClient(endpoint, appid, secret string, timeout time.Duration) (*WechatClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	proxy, err := grpc.DialContext(ctx, endpoint,
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
		grpc.WithPerRPCCredentials(security.NewPerRPCCredentialForSign(appid, secret)))
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s, %w", endpoint, err)
	}

	wc := &WechatClient{
		meta:    meta.NewMetaClient(proxy),
		wechat:  wechatpb.NewSrvClient(proxy),
		timeout: timeout,
	}

	return wc, nil
}

func (c *WechatClient) genJWT() (string, error) {
	if c.token.expireTime.Before(time.Now()) {
		ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
		defer cancel()

		dailCtx := metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{"namespace": "spacegrower"}))

		resp, err := c.meta.GenerateJWT(dailCtx, &meta.GenerateJWTRequest{
			User: "system",
			Fields: map[string]string{
				"type": c.mode,
			},
		})

		if err != nil {
			return "", fmt.Errorf("failed to generate jwt token, %w", err)
		}
		c.token.token = resp.JWT
		c.token.expireTime = time.Unix(resp.ExpireTime, 0)
	}
	return c.token.token, nil
}

func (c *WechatClient) SendTplMessage(ctx context.Context, in *wechatpb.SendTplMessageRequest) (*wechatpb.SendTplMessageResponse, error) {
	jwt, err := c.genJWT()
	if err != nil {
		return nil, err
	}

	dailCtx := metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{"namespace": "tools", security.TOKEN_KEY: jwt}))
	return c.wechat.SendTplMessage(dailCtx, in)
}
