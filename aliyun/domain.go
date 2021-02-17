package aliyun

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"
)

type (
	queryDomainListResponse struct {
		Message string
		Data    struct {
			Domain []struct {
				DomainName     string `json:"DomainName"`
				ExpirationDate string `json:"ExpirationDate"`
			} `json:"Domain"`
		} `json:"Data"`
	}

	Domain struct {
		Name      string
		ExpiredAt time.Time
	}

	ByDomainExpiredAtAsc []Domain
)

func (c *Client) GetDomainList() (domains []Domain, err error) {
	v := url.Values{}
	v.Set("Action", "QueryDomainList")
	v.Set("PageNum", "1")
	v.Set("PageSize", "100")
	v.Set("Version", "2018-01-29")
	var req *http.Request
	req, err = c.buildRequest("domain", &v)
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5*time.Second))
	defer cancel()
	req = req.WithContext(ctx)
	var res *http.Response
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	var resp queryDomainListResponse
	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		return
	}
	if resp.Message != "" {
		err = errors.New(resp.Message)
		return
	}
	loc := time.FixedZone("UTC+8", 8*60*60)
	for _, d := range resp.Data.Domain {
		t, _ := time.ParseInLocation("2006-01-02 15:04:05", d.ExpirationDate, loc)
		domains = append(domains, Domain{
			Name:      d.DomainName,
			ExpiredAt: t,
		})
	}
	return
}

func (a ByDomainExpiredAtAsc) Len() int           { return len(a) }
func (a ByDomainExpiredAtAsc) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByDomainExpiredAtAsc) Less(i, j int) bool { return a[i].ExpiredAt.Before(a[j].ExpiredAt) }
