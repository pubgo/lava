// Package turncred 生成 coturn TURN REST API 临时凭证（HMAC-SHA1）。
//
// 与 coturn `use-auth-secret` + `static-auth-secret` 配合：
//   - username = "<expiry_unix>:<user_id>"
//   - password = base64(HMAC-SHA1(secret, username))
//
// 参见 https://github.com/coturn/coturn/blob/master/README.turnserver
package turncred

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"time"
)

const defaultUserID = "lava"

// Generate 生成临时 TURN 用户名与密码。
// userID 为空时使用 "lava"；ttl <= 0 时默认 24 小时。
func Generate(secret, userID string, ttl time.Duration, now time.Time) (username, password string, err error) {
	if secret == "" {
		return "", "", fmt.Errorf("turncred: empty secret")
	}
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	if userID == "" {
		userID = defaultUserID
	}
	if now.IsZero() {
		now = time.Now()
	}
	expiry := now.Add(ttl).Unix()
	username = fmt.Sprintf("%d:%s", expiry, userID)

	mac := hmac.New(sha1.New, []byte(secret))
	if _, err := mac.Write([]byte(username)); err != nil {
		return "", "", err
	}
	password = base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return username, password, nil
}
