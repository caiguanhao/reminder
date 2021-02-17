package aliyun

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type (
	queryAccountBalanceResponse struct {
		Message string `json:"Message"`
		Data    struct {
			AvailableCashAmount string `json:"AvailableCashAmount"`
			MybankCreditAmount  string `json:"MybankCreditAmount"`
			Currency            string `json:"Currency"`
			AvailableAmount     string `json:"AvailableAmount"`
			CreditAmount        string `json:"CreditAmount"`
		} `json:"Data"`
		Success bool `json:"Success"`
	}
)

func (c *Client) GetAccountBalance() (float64, error) {
	v := url.Values{}
	v.Set("Action", "QueryAccountBalance")
	v.Set("Version", "2017-12-14")
	req, err := c.buildRequest("business", &v)
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
	var resp queryAccountBalanceResponse
	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		return 0, err
	}
	if !resp.Success {
		return 0, errors.New(resp.Message)
	}
	value, err := strconv.ParseFloat(strings.Replace(resp.Data.AvailableAmount, ",", "", -1), 64)
	if err != nil {
		return 0, err
	}
	return value, nil
}
