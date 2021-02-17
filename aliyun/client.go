package aliyun

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type (
	Client struct {
		AccessKeyId     string
		AccessKeySecret string
	}
)

func (c *Client) buildRequest(service string, v *url.Values) (*http.Request, error) {
	randomBytes := make([]byte, 10)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, err
	}
	v.Set("Format", "JSON")
	v.Set("AccessKeyId", c.AccessKeyId)
	v.Set("SignatureMethod", "HMAC-SHA1")
	v.Set("Timestamp", time.Now().UTC().Format("2006-01-02T15:04:05Z"))
	v.Set("SignatureVersion", "1.0")
	v.Set("SignatureNonce", hex.EncodeToString(randomBytes))
	c.sign(v)
	if !strings.HasSuffix(service, ".aliyuncs.com") {
		service += ".aliyuncs.com"
	}
	url := "https://" + service + "/?" + v.Encode()
	return http.NewRequest(http.MethodGet, url, nil)
}

func (c *Client) sign(v *url.Values) {
	stringToSign := v.Encode()
	stringToSign = strings.Replace(stringToSign, "+", "%20", -1)
	stringToSign = strings.Replace(stringToSign, "*", "%2A", -1)
	stringToSign = strings.Replace(stringToSign, "%7E", "~", -1)
	stringToSign = url.QueryEscape(stringToSign)
	stringToSign = http.MethodGet + "&%2F&" + stringToSign
	hmac := hmac.New(sha1.New, []byte(c.AccessKeySecret+"&"))
	hmac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(hmac.Sum(nil))
	v.Set("Signature", signature)
}
