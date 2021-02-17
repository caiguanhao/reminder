# reminder

```
reminder | html2png -full | lark-upload-image -send ou_0123456789abcdef0123456789abcdef
```

See [html2png](https://github.com/caiguanhao/html2png) and
[lark-upload-image](https://github.com/caiguanhao/lark-slim/tree/master/cmd/lark-upload-image).

Create `./cmd/reminder/key.go` and `go build -v ./cmd/reminder`:

```go
func init() {
	ac = aliyun.Client{
		AccessKeyId:     "",
		AccessKeySecret: "",
	}

	tc = tencentcloud.Client{
		AccessKeyId:     "",
		AccessKeySecret: "",
	}

	oc = ossslim.Client{
		AccessKeyId:     ac.AccessKeyId,
		AccessKeySecret: ac.AccessKeySecret,
		Prefix:          "https://bucket.oss-cn-hongkong.aliyuncs.com",
		Bucket:          "bucket",
	}

	regionIds = []string{"cn-hangzhou", "cn-hongkong"}

	certsDir = "certs/"

	certsEncryptionKey = strings.Join([]string{
		"\x00", "\x00", "\x00", "\x00", "\x00", "\x00", "\x00", "\x00",
		"\x00", "\x00", "\x00", "\x00", "\x00", "\x00", "\x00", "\x00",
		"\x00", "\x00", "\x00", "\x00", "\x00", "\x00", "\x00", "\x00",
		"\x00", "\x00", "\x00", "\x00", "\x00", "\x00", "\x00", "\x00",
	}, "")

}
```
