package main

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/caiguanhao/reminder/aliyun"
)

func getServersInfo() (retRegions []string, retServers map[string][]item, retServersError string) {
	regions, err := ac.GetRegionList(regionIds...)
	if err != nil {
		retServersError = err.Error()
		return
	}
	var wg sync.WaitGroup
	wg.Add(len(regions))
	var s sync.Map
	for _, r := range regions {
		retRegions = append(retRegions, r.ID)
		go func(r aliyun.Region) {
			i, err := ac.GetInstanceList(r)
			if err == nil {
				sort.Sort(aliyun.ByInstanceExpiredAtAsc(i))
				var servers []item
				for _, a := range i {
					days := int(time.Until(a.ExpiredAt).Hours() / 24)
					date := fmt.Sprintf("%s (%d days)", a.ExpiredAt.Format("2006-01-02"), days)
					class := "text-danger"
					if days > 1000 {
						date = "∞"
					}
					if days >= 14 {
						class = "text-success"
					} else if days >= 7 {
						class = "text-warning"
					}
					servers = append(servers, item{
						Name:  a.Name,
						Date:  date,
						Class: class,
					})
				}
				s.Store(r.ID, servers)
			} else {
				s.Store(r.ID, []item{
					{
						Name:  "Error",
						Date:  err.Error(),
						Class: "text-danger",
					},
				})
			}
			wg.Done()
		}(r)
	}
	wg.Wait()
	servers := map[string][]item{}
	s.Range(func(key, value interface{}) bool {
		servers[key.(string)] = value.([]item)
		return true
	})
	retServers = servers
	return
}
