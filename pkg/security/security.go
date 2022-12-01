package security

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/holdno/microuse/etc/spacegrower/meta"
	"github.com/holdno/microuse/utils"

	"github.com/spacegrower/watermelon/infra/wlog"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const (
	signAppidKey = "wm-appid"
	signSignKey  = "wm-sign"
	signTimeKey  = "wm-time"
	TOKEN_KEY    = "sg-token"
)

func GenSign(appid, secret string, signTime int64) string {
	signString := fmt.Sprintf("appid=%s&secret=%s&time=%d", appid, secret, signTime)
	signString += fmt.Sprintf("%s&key=%s", signString, utils.MD5(signString))
	return utils.MD5(signString)
}

func NewPerRPCCredentialForToken(cli interface {
	GenerateJWT(ctx context.Context, in *meta.GenerateJWTRequest, opts ...grpc.CallOption) (*meta.GenerateJWTReply, error)
}, args *meta.GenerateJWTRequest) *perRPCCredentialForToken {
	return &perRPCCredentialForToken{
		cli: cli,
	}
}

// perRPCCredential implements "grpccredentials.PerRPCCredentials" interface.
type perRPCCredentialForToken struct {
	cli interface {
		GenerateJWT(ctx context.Context, in *meta.GenerateJWTRequest, opts ...grpc.CallOption) (*meta.GenerateJWTReply, error)
	}
	args       *meta.GenerateJWTRequest
	token      string
	expireTime time.Time
}

func (rc *perRPCCredentialForToken) RequireTransportSecurity() bool { return false }

func (rc *perRPCCredentialForToken) GetRequestMetadata(ctx context.Context, s ...string) (map[string]string, error) {
	if rc.expireTime.Before(time.Now()) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		resp, err := rc.cli.GenerateJWT(ctx, rc.args)
		if err != nil {
			wlog.Error("failed to generate jwt token", zap.String("components", "pkg/micro"))
			return nil, nil
		}
		rc.token = resp.JWT
		rc.expireTime = time.Unix(resp.ExpireTime, 0)
	}
	return map[string]string{TOKEN_KEY: rc.token}, nil
}

func NewPerRPCCredentialForSign(appid, secret string) *perRPCCredentialForSign {
	return &perRPCCredentialForSign{
		appid:  appid,
		secret: secret,
	}
}

// perRPCCredential implements "grpccredentials.PerRPCCredentials" interface.
type perRPCCredentialForSign struct {
	appid  string
	secret string
}

func (rc *perRPCCredentialForSign) RequireTransportSecurity() bool { return false }

func (rc *perRPCCredentialForSign) GetRequestMetadata(ctx context.Context, s ...string) (map[string]string, error) {
	now := time.Now().Unix()
	return map[string]string{signAppidKey: rc.appid, signSignKey: GenSign(rc.appid, rc.secret, now), signTimeKey: strconv.FormatInt(now, 10)}, nil
}
