package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/ranjib/gypsy/structs"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"
)

// Config represent configuration for API endpoints
type Config struct {
	Address    string
	WaitTime   time.Duration
	HttpClient *http.Client
}

// DefaultConfig Obtain API config struct with default values
func DefaultConfig() *Config {
	config := &Config{
		Address:    "http://localhost:5678",
		HttpClient: cleanhttp.DefaultClient(),
	}
	if addr := os.Getenv("GYPSY_ADDR"); addr != "" {
		config.Address = addr
	}
	return config
}

type Client struct {
	config Config
}

func NewClient(config *Config) (*Client, error) {
	defConfig := DefaultConfig()
	if config.Address == "" {
		config.Address = defConfig.Address
	} else if _, err := url.Parse(config.Address); err != nil {
		return nil, fmt.Errorf("Invalid address '%s': %v", config.Address, err)
	}

	if config.HttpClient == nil {
		config.HttpClient = defConfig.HttpClient
	}

	client := &Client{
		config: *config,
	}
	return client, nil
}

func (c *Client) Request(method, endpoint string, body io.Reader) (*http.Response, error) {
	req, err := c.newRequest(method, endpoint, body)
	if err != nil {
		return nil, err
	}
	return c.config.HttpClient.Do(req)
}

func (c *Client) newRequest(method, path string, body io.Reader) (*http.Request, error) {
	base, _ := url.Parse(c.config.Address)
	u, _ := url.Parse(path)
	target := &url.URL{
		Scheme: base.Scheme,
		Host:   base.Host,
		Path:   u.Path,
	}
	log.Debugf("Request URL: %#v\n", target)
	var params url.Values
	params = make(map[string][]string)
	// Add in the query parameters, if any
	for key, values := range u.Query() {
		for _, value := range values {
			params.Add(key, value)
		}
	}
	target.RawQuery = params.Encode()
	req, err := http.NewRequest(method, target.RequestURI(), body)
	if err != nil {
		log.Errorf("Failed to build http request. Error: %s\n", err)
		return nil, err
	}
	req.URL.Host = target.Host
	req.URL.Scheme = target.Scheme
	req.Host = target.Host
	return req, nil
}

func (c *Client) ListPipelines() ([]string, error) {
	resp, err := c.Request("GET", "/pipelines", bytes.NewBuffer(nil))
	if err != nil {
		log.Errorf("Failed to obtain pipeline list. Error:%s\n", err)
		return nil, err

	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	out := []string{}
	return out, dec.Decode(&out)
}

func (c *Client) GetPipeline(name string) (string, error) {
	resp, err := c.Request("GET", "/pipelines/"+name, bytes.NewBuffer(nil))
	if err != nil {
		log.Errorf("Failed to obtain pipeline details. Error:%s\n", err)
		return "", err
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Failed to read response body. Error:%s\n", err)
		return "", err
	}
	return string(content), nil
}

func (c *Client) CreatePipeline(pipeline *structs.Pipeline) error {
	data, err := yaml.Marshal(pipeline)
	if err != nil {
		log.Errorf("Failed to convert pipeline into yaml. Error:%s\n", err)
		return err
	}
	resp, err := c.Request("POST", "/pipelines", bytes.NewBuffer(data))
	if err != nil {
		log.Errorf("Failed to create pipeline. Error:%s\n", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed to create pipeline. HTTP Status code: %d", resp.StatusCode)
	}
	return nil
}
func (c *Client) DeletePipeline(pipeline string) error {
	resp, err := c.Request("DELETE", "/pipelines/"+pipeline, bytes.NewBuffer(nil))
	if err != nil {
		log.Errorf("Failed to delete pipeline. Error:%s\n", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed to delete pipeline. HTTP Status code: %d", resp.StatusCode)
	}
	return nil
}
