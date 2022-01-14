package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"text/template"
	"time"

	"github.com/caiguanhao/ossslim"
	"github.com/caiguanhao/reminder/aliyun"
	"github.com/caiguanhao/reminder/tencentcloud"
)

type (
	item struct {
		Name  string
		Date  string
		Class string
	}
)

var (
	html string

	ac aliyun.Client
	tc tencentcloud.Client
	oc ossslim.Client

	regionIds []string

	certsDir           string
	certsEncryptionKey []byte
)

func main() {
	defaultConfigFile := ".ossenc.go"
	if home, _ := os.UserHomeDir(); home != "" {
		defaultConfigFile = filepath.Join(home, defaultConfigFile)
	}
	configFile := flag.String("c", defaultConfigFile, "location of the config file")
	createConfig := flag.Bool("C", false, "create (update if exists) config file and exit")
	flag.Parse()

	conf := readConf(*configFile, *createConfig)

	ac = aliyun.Client{
		AccessKeyId:     conf.AliyunAccessKeyId,
		AccessKeySecret: conf.AliyunAccessKeySecret,
	}

	regionIds = conf.AliyunRegionIds

	tc = tencentcloud.Client{
		AccessKeyId:     conf.TencentCloudAccessKeyId,
		AccessKeySecret: conf.TencentCloudAccessKeySecret,
	}

	oc = ossslim.Client{
		AccessKeyId:     conf.OSSAccessKeyId,
		AccessKeySecret: conf.OSSAccessKeySecret,
		Prefix:          conf.OSSPrefix,
		Bucket:          conf.OSSBucket,
	}

	certsDir = conf.OSSCertsDir

	certsEncryptionKey = conf.OSSCertsEncryptionKey

	var tplData struct {
		Now string

		AliyunBalance            string
		AliyunBalanceClass       string
		TencentCloudBalance      string
		TencentCloudBalanceClass string

		Certs   []item
		Domains []item
		Regions []string
		Servers map[string][]item

		ServersError string
	}

	tplData.Now = time.Now().Format("2006-01-02 15:04:05")

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		tplData.AliyunBalanceClass = "text-danger"
		b, err := ac.GetAccountBalance()
		if err == nil {
			tplData.AliyunBalance = fmt.Sprintf("%.2f", b)
			if b > 1000 {
				tplData.AliyunBalanceClass = "text-success"
			} else if b > 500 {
				tplData.AliyunBalanceClass = "text-warning"
			}
		} else {
			tplData.AliyunBalance = err.Error()
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		tplData.TencentCloudBalanceClass = "text-danger"
		b, err := tc.GetAccountBalance()
		if err == nil {
			tplData.TencentCloudBalance = fmt.Sprintf("%.2f", b)
			if b > 1000 {
				tplData.TencentCloudBalanceClass = "text-success"
			} else if b > 500 {
				tplData.TencentCloudBalanceClass = "text-warning"
			}
		} else {
			tplData.TencentCloudBalance = err.Error()
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		domains, err := ac.GetDomainList()
		if err == nil {
			sort.Sort(aliyun.ByDomainExpiredAtAsc(domains))
			for _, d := range domains {
				days := int(time.Until(d.ExpiredAt).Hours() / 24)
				class := "text-danger"
				if days > 180 {
					class = "text-success"
				} else if days > 30 {
					class = "text-warning"
				}
				tplData.Domains = append(tplData.Domains, item{
					Name:  d.Name,
					Date:  fmt.Sprintf("%s (%d days)", d.ExpiredAt.Format("2006-01-02"), days),
					Class: class,
				})
			}
		} else {
			tplData.Domains = append(tplData.Domains, item{
				Name:  "Error",
				Date:  err.Error(),
				Class: "text-danger",
			})
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		tplData.Regions, tplData.Servers, tplData.ServersError = getServersInfo()
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		tplData.Certs = getCertsInfo()
		wg.Done()
	}()

	wg.Wait()

	t, err := template.New("").Parse(html)
	if err != nil {
		panic(err)
	}
	err = t.Execute(os.Stdout, tplData)
	if err != nil {
		panic(err)
	}
	return
}
