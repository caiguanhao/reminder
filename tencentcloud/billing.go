package tencentcloud

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type (
	describeAccountBalanceResponse struct {
		Response struct {
			Balance *int `json:"Balance"`
		} `json:"Response"`
	}
)

func (c *Client) GetAccountBalance() (float64, error) {
	req, err := c.buildRequest("billing", "DescribeAccountBalance", "2018-07-09")
	if err != nil {
		return 0, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5*time.Second))
	defer cancel()
	req = req.WithContext(ctx)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	var resp describeAccountBalanceResponse
	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		return 0, err
	}
	if resp.Response.Balance == nil {
		return 0, errors.New("no data")
	}
	return float64(*resp.Response.Balance) / 100, nil
}
