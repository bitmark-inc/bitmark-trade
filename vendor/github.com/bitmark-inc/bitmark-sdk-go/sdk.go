package bitmarksdk

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Config struct {
	Network    Network
	HTTPClient *http.Client
	APIToken   string
}

var (
	config    *Config
	apiClient *BackendImplementation
)

type Network string

const (
	Livenet = Network("livenet")
	Testnet = Network("testnet")
)

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (ae *APIError) Error() string {
	return fmt.Sprintf("[%d] %s", ae.Code, ae.Message)
}

func Init(cfg *Config) {
	config = cfg
	switch cfg.Network {
	case Livenet:
		apiClient = &BackendImplementation{
			HTTPClient:   cfg.HTTPClient,
			URLAuthority: "https://api.bitmark.com",
			APIToken:     cfg.APIToken,
		}
	case Testnet:
		apiClient = &BackendImplementation{
			HTTPClient:   cfg.HTTPClient,
			URLAuthority: "https://api.test.bitmark.com",
			APIToken:     cfg.APIToken,
		}
	}
}

func GetNetwork() Network {
	return config.Network
}

func GetAPIClient() *BackendImplementation {
	return apiClient
}

type BackendImplementation struct {
	HTTPClient        *http.Client
	URLAuthority      string
	APIToken          string
	MaxNetworkRetries int
}

func (s *BackendImplementation) NewRequest(method, path string, body io.Reader) (*http.Request, error) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	url := s.URLAuthority + path

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("api-token", s.APIToken)
	// TODO: workaroun for gateway proxy
	req.Header.Add("Accept-Encoding", "*")

	return req, nil
}

func (s *BackendImplementation) Do(req *http.Request, v interface{}) error {
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		var aerr APIError
		if err := json.NewDecoder(resp.Body).Decode(&aerr); err != nil {
			return errors.New("unexpected api response")
		}

		return &aerr
	}

	if v != nil {
		return json.NewDecoder(resp.Body).Decode(v)
	}

	return nil
}
