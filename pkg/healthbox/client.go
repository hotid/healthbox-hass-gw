package healthbox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hotid/healthbox-hass-gw/internal/pkg/config"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	baseUrl    *url.URL
	httpClient *http.Client
	maxTries   int
}

func NewClient(config *config.Config) *Client {
	baseUrl := fmt.Sprintf("http://%s:%s", config.Healthbox.Host, config.Healthbox.Port)
	u, err := url.Parse(baseUrl)
	if err != nil {
		panic(err)
	}

	c := &Client{httpClient: &http.Client{Timeout: time.Second * 5}, baseUrl: u}

	return c
}

func (c *Client) SetMaxTries(maxTries int) {
	c.maxTries = maxTries
}

func (c *Client) GetCurrentData() (*CurrentData, error) {
	body, err := c.getData("/v1/api/data/current")
	if err != nil {
		return nil, err
	}

	currentData := CurrentData{}
	err = json.Unmarshal(body, &currentData)
	if err != nil {
		return nil, err
	}

	return &currentData, nil
}

func (c *Client) GetBoostInfo(currentData *CurrentData) (*map[int]BoostInfo, error) {
	currentBoostInfo := make(map[int]BoostInfo)

	for _, room := range currentData.Room {
		body, err := c.getData(fmt.Sprintf("/v1/api/boost/%d", room.Id))
		boostInfo := BoostInfo{}
		err = json.Unmarshal(body, &boostInfo)
		if err != nil {
			return nil, err
		}
		currentBoostInfo[room.Id] = boostInfo
		time.Sleep(1 * time.Second)
	}
	return &currentBoostInfo, nil
}

func (c *Client) Put(endpoint string, payload []byte) error {
	req, err := http.NewRequest(http.MethodPut, c.getFullUrl(endpoint), bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	_, err = c.httpClient.Do(req)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) getData(path string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, c.getFullUrl(path), nil)
	if err != nil {
		return nil, err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	} else {
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (c *Client) getFullUrl(path string) string {
	return fmt.Sprintf("%s%s", c.baseUrl.String(), path)
}
