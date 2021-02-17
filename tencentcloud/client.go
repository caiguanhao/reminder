package tencentcloud

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type (
	Client struct {
		AccessKeyId     string
		AccessKeySecret string
	}
)

func (c *Client) buildRequest(service, action, version string) (*http.Request, error) {
	method := http.MethodPost
	canonicalURI := "/"
	canonicalQueryString := ""
	host := service + ".tencentcloudapi.com"
	now := time.Now().UTC()
	date := now.Format("2006-01-02")
	timestamp := strconv.FormatInt(now.Unix(), 10)
	contentType := "application/json; charset=utf-8"
	canonicalHeaders := fmt.Sprintf("content-type:%s\nhost:%s\n", contentType, host)
	requestPayload := ""
	hashedRequestPayload := sha256hex(requestPayload)
	url := "https://" + host + canonicalURI
	if canonicalQueryString != "" {
		url = url + "?" + canonicalQueryString
	}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	algorithm := "TC3-HMAC-SHA256"
	credentialScope := strings.Join([]string{date, service, "tc3_request"}, "/")
	signedHeaders := "content-type;host"
	canonicalRequest := strings.Join([]string{
		method,
		canonicalURI,
		canonicalQueryString,
		canonicalHeaders,
		signedHeaders,
		hashedRequestPayload,
	}, "\n")
	hashedCanonicalRequest := sha256hex(canonicalRequest)
	string2sign := strings.Join([]string{
		algorithm,
		timestamp,
		credentialScope,
		hashedCanonicalRequest,
	}, "\n")
	secretDate := hmacsha256(date, "TC3"+c.AccessKeySecret)
	secretService := hmacsha256(service, secretDate)
	signature := hex.EncodeToString([]byte(hmacsha256(string2sign, hmacsha256("tc3_request", secretService))))
	authorization := fmt.Sprintf(
		"%s Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		algorithm,
		c.AccessKeyId,
		credentialScope,
		signedHeaders,
		signature,
	)
	req.Header.Set("Authorization", authorization)
	req.Header.Set("X-TC-Action", action)
	req.Header.Set("X-TC-Version", version)
	req.Header.Set("X-TC-Timestamp", timestamp)
	req.Header.Set("Content-Type", contentType)
	return req, nil
}

func hmacsha256(s, key string) string {
	hashed := hmac.New(sha256.New, []byte(key))
	hashed.Write([]byte(s))
	return string(hashed.Sum(nil))
}

func sha256hex(s string) string {
	b := sha256.Sum256([]byte(s))
	return hex.EncodeToString(b[:])
}
