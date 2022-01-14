package main

import (
	"crypto/rand"
	"encoding/hex"
	"io/ioutil"
	"log"
	"os"

	"github.com/gopsql/goconf"
)

type (
	config struct {
		AliyunAccessKeyId           string
		AliyunAccessKeySecret       string
		AliyunRegionIds             []string
		TencentCloudAccessKeyId     string
		TencentCloudAccessKeySecret string
		OSSAccessKeyId              string
		OSSAccessKeySecret          string
		OSSPrefix                   string
		OSSBucket                   string
		OSSCertsDir                 string
		OSSCertsEncryptionKey       key
	}

	key []byte
)

func (k *key) SetString(input string) (err error) {
	*k, err = hex.DecodeString(input)
	return
}

func (k key) String() string {
	return hex.EncodeToString(k)
}

func readConf(file string, toUpdate bool) (conf config) {
	created := false
	content, err := ioutil.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) && toUpdate {
			conf = config{
				OSSPrefix: "https://example.oss-cn-hongkong.aliyuncs.com",
				OSSBucket: "example",
			}
			created = true
		} else {
			log.Fatalln(err)
		}
	} else {
		err = goconf.Unmarshal([]byte(content), &conf)
		if err != nil {
			log.Fatalln(err)
		}
	}
	if toUpdate {
		if len(conf.OSSCertsEncryptionKey) == 0 {
			k := make(key, 32)
			rand.Read(k)
			conf.OSSCertsEncryptionKey = k
		}
		content, err := goconf.Marshal(conf)
		if err != nil {
			log.Fatalln(err)
		}
		err = ioutil.WriteFile(file, content, 0600)
		if err != nil {
			log.Fatalln(err)
		}
		if created {
			log.Println("Config file created:", file)
		} else {
			log.Println("Config file updated:", file)
		}
		os.Exit(0)
	}
	return
}
