package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	configurator "github.com/OpenCHAMI/configurator/pkg"
	"github.com/rs/zerolog/log"
)

// An struct that's meant to extend functionality of the base HTTP client by
// adding commonly made requests to SMD. The implemented functions are can be
// used in generator plugins to fetch data when it is needed to substitute
// values for the Jinja templates used.
type SmdClient struct {
	http.Client `json:"-" yaml:"-"`
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	AccessToken string `yaml:"access-token"`
}

// Constructor function that allows supplying Option arguments to set
// things like the host, port, access token, etc.
func NewSmdClient(opts ...Option) SmdClient {
	var (
		params = ToParams(opts...)
		client = SmdClient{
			Host:        params.Host,
			AccessToken: params.AccessToken,
		}
	)

	return client
}

// Fetch the ethernet interfaces from SMD service using its API. An access token may be required if the SMD
// service SMD_JWKS_URL envirnoment variable is set.
func (client *SmdClient) FetchEthernetInterfaces(verbose bool) ([]configurator.EthernetInterface, error) {
	var (
		eths  = []configurator.EthernetInterface{}
		bytes []byte
		err   error
	)
	// make request to SMD endpoint
	bytes, err = client.makeRequest("/Inventory/EthernetInterfaces")
	if err != nil {
		return nil, fmt.Errorf("failed to read HTTP response: %v", err)
	}

	// unmarshal response body JSON and extract in object
	err = json.Unmarshal(bytes, &eths)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	// print what we got if verbose is set
	if verbose {
		log.Info().Str("ethernet_interfaces", string(bytes)).Msg("found interfaces")
	}

	return eths, nil
}

// Fetch the components from SMD using its API. An access token may be required if the SMD
// service SMD_JWKS_URL envirnoment variable is set.
func (client *SmdClient) FetchComponents(verbose bool) ([]configurator.Component, error) {
	var (
		comps = []configurator.Component{}
		bytes []byte
		err   error
	)
	// make request to SMD endpoint
	bytes, err = client.makeRequest("/State/Components")
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %v", err)
	}

	// make sure our response is actually JSON
	if !json.Valid(bytes) {
		return nil, fmt.Errorf("expected valid JSON response: %v", string(bytes))
	}

	// unmarshal response body JSON and extract in object
	var tmp map[string]any
	err = json.Unmarshal(bytes, &tmp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}
	bytes, err = json.Marshal(tmp["RedfishEndpoints"].([]any))
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %v", err)
	}
	err = json.Unmarshal(bytes, &comps)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	// print what we got if verbose is set
	if verbose {
		log.Info().Str("components", string(bytes)).Msg("found components")
	}

	return comps, nil
}

// TODO: improve implementation of this function
func (client *SmdClient) FetchRedfishEndpoints(verbose bool) ([]configurator.RedfishEndpoint, error) {
	var (
		eps = []configurator.RedfishEndpoint{}
		tmp map[string]any
	)

	// make initial request to get JSON with 'RedfishEndpoints' as property
	b, err := client.makeRequest("/Inventory/RedfishEndpoints")
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP resquest: %v", err)
	}
	// make sure response is in JSON
	if !json.Valid(b) {
		return nil, fmt.Errorf("expected valid JSON response: %v", string(b))
	}
	err = json.Unmarshal(b, &tmp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	// marshal RedfishEndpoint JSON back to configurator.RedfishEndpoint
	b, err = json.Marshal(tmp["RedfishEndpoints"].([]any))
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %v", err)
	}
	err = json.Unmarshal(b, &eps)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	// show the final result
	if verbose {
		log.Info().Str("redfish_endpoints", string(b)).Msg("found redfish endpoints")
	}

	return eps, nil
}

func (client *SmdClient) makeRequest(endpoint string) ([]byte, error) {
	if client == nil {
		return nil, fmt.Errorf("client is nil")
	}

	// fetch DHCP related information from SMD's endpoint:
	url := fmt.Sprintf("%s/hsm/v2%s", client.Host, endpoint)
	req, err := http.NewRequest(http.MethodGet, url, bytes.NewBuffer([]byte{}))
	if err != nil {
		return nil, fmt.Errorf("failed to create new HTTP request: %v", err)
	}

	// include access token in authorzation header if found
	// NOTE: This shouldn't be needed for this endpoint since it's public
	if client.AccessToken != "" {
		req.Header.Add("Authorization", "Bearer "+client.AccessToken)
	}

	// make the request to SMD
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}

	// read the contents of the response body
	return io.ReadAll(res.Body)
}
