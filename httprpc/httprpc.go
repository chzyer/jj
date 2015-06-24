package httprpc

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"gopkg.in/logex.v1"
)

type ErrorResp struct {
	Result int    `json:"result"`
	Reason string `json:"reason"`
}

type Client struct {
	host   string
	client *http.Client
}

func NewClient(host string, timeout time.Duration) (*Client, error) {
	c := &Client{
		host: host,
		client: &http.Client{
			Timeout: timeout,
		},
	}
	return c, nil
}

func (c *Client) Call(method string, data url.Values, v interface{}) error {
	r := bytes.NewReader([]byte(data.Encode()))
	req, err := http.NewRequest("POST", c.getPath(method), r)
	if err != nil {
		return logex.Trace(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.client.Do(req)
	if err != nil {
		return logex.Trace(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return logex.Trace(err)
	}

	if err := json.Unmarshal(body, &v); err != nil {
		return logex.Trace(err)
	}
	return nil
}

func (c *Client) getPath(method string) string {
	return c.host + method
}
