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
	describeRegionsResponse struct {
		Message string
		Regions struct {
			Region []struct {
				RegionID       string `json:"RegionId"`
				RegionEndpoint string `json:"RegionEndpoint"`
				LocalName      string `json:"LocalName"`
			} `json:"Region"`
		} `json:"Regions"`
	}

	Region struct {
		ID       string
		Name     string
		Endpoint string
	}

	describeInstancesResponse struct {
		Message   string
		Instances struct {
			Instance []struct {
				InstanceName string `json:"InstanceName"`
				ExpiredTime  string `json:"ExpiredTime"`
			} `json:"Instance"`
		} `json:"Instances"`
	}

	Instance struct {
		Name      string
		ExpiredAt time.Time
	}

	ByInstanceExpiredAtAsc []Instance
)

func (c *Client) GetRegionList(filter ...string) (regions []Region, err error) {
	v := url.Values{}
	v.Set("Action", "DescribeRegions")
	v.Set("Version", "2014-05-26")
	var req *http.Request
	req, err = c.buildRequest("ecs", &v)
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
	var resp describeRegionsResponse
	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		return
	}
	if resp.Message != "" {
		err = errors.New(resp.Message)
		return
	}
	for _, r := range resp.Regions.Region {
		if len(filter) > 0 && !contains(filter, r.RegionID) {
			continue
		}
		regions = append(regions, Region{
			ID:       r.RegionID,
			Name:     r.LocalName,
			Endpoint: r.RegionEndpoint,
		})
	}
	return
}

func (c *Client) GetInstanceList(region Region) (instances []Instance, err error) {
	v := url.Values{}
	v.Set("Action", "DescribeInstances")
	v.Set("RegionId", region.ID)
	v.Set("PageSize", "100")
	v.Set("Version", "2014-05-26")
	var req *http.Request
	req, err = c.buildRequest(region.Endpoint, &v)
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
	var resp describeInstancesResponse
	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		return
	}
	if resp.Message != "" {
		err = errors.New(resp.Message)
		return
	}
	loc := time.FixedZone("UTC+8", 8*60*60)
	for _, i := range resp.Instances.Instance {
		t, _ := time.Parse("2006-01-02T15:04Z", i.ExpiredTime)
		instances = append(instances, Instance{
			Name:      i.InstanceName,
			ExpiredAt: t.In(loc),
		})
	}
	return
}

func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func (a ByInstanceExpiredAtAsc) Len() int           { return len(a) }
func (a ByInstanceExpiredAtAsc) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByInstanceExpiredAtAsc) Less(i, j int) bool { return a[i].ExpiredAt.Before(a[j].ExpiredAt) }
