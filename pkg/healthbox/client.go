package healthbox

import (
	"encoding/json"
	"fmt"
	"github.com/hotid/healthbox-hass-gw/internal/pkg/config"
	"io/ioutil"
	"log"
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
		log.Fatal(err)
	}

	c := &Client{httpClient: &http.Client{Timeout: time.Second * 5}, baseUrl: u}

	return c
}

func (c *Client) SetMaxTries(maxTries int) {
	c.maxTries = maxTries
}

func (c *Client) getUrl(path string) string {
	return fmt.Sprintf("%s%s", c.baseUrl.String(), path)
}

func (c *Client) GetCurrentData() (*CurrentData, error) {
	req, err := http.NewRequest(http.MethodGet, c.getUrl("/v1/api/data/current"), nil)
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

	currentData := CurrentData{}
	err = json.Unmarshal(body, &currentData)
	if err != nil {
		return nil, err
	}

	return &currentData, nil
}
