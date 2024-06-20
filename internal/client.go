package configurator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/OpenCHAMI/configurator/internal/util"
)

type SmdClient struct {
	http.Client `json:"-"`
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	AccessToken string `yaml:"access-token"`
}

type Params = map[string]any
type Option func(Params)

func WithVerbosity() Option {
	return func(p util.Params) {
		p["verbose"] = true
	}
}

func NewParams() Params {
	return Params{
		"verbose": false,
	}
}

// Fetch the ethernet interfaces from SMD service using its API. An access token may be required if the SMD
// service SMD_JWKS_URL envirnoment variable is set.
func (client *SmdClient) FetchEthernetInterfaces(opts ...util.Option) ([]EthernetInterface, error) {
	var (
		params  = util.GetParams(opts...)
		verbose = util.Get[bool](params, "verbose")
		eths    = []EthernetInterface{}
	)
	// make request to SMD endpoint
	b, err := client.makeRequest("/Inventory/EthernetInterfaces")
	if err != nil {
		return nil, fmt.Errorf("failed to read HTTP response: %v", err)
	}

	// unmarshal response body JSON and extract in object
	err = json.Unmarshal(b, &eths)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	// print what we got if verbose is set
	if verbose != nil {
		if *verbose {
			fmt.Printf("Ethernet Interfaces: %v\n", string(b))
		}
	}

	return eths, nil
}

// Fetch the components from SMD using its API. An access token may be required if the SMD
// service SMD_JWKS_URL envirnoment variable is set.
func (client *SmdClient) FetchComponents(opts ...util.Option) ([]Component, error) {
	var (
		params  = util.GetParams(opts...)
		verbose = util.Get[bool](params, "verbose")
		comps   = []Component{}
	)
	// make request to SMD endpoint
	b, err := client.makeRequest("/State/Components")
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %v", err)
	}

	// make sure our response is actually JSON
	if !json.Valid(b) {
		return nil, fmt.Errorf("expected valid JSON response: %v", string(b))
	}

	// unmarshal response body JSON and extract in object
	var tmp map[string]any
	err = json.Unmarshal(b, &tmp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}
	b, err = json.Marshal(tmp["RedfishEndpoints"].([]any))
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %v", err)
	}
	err = json.Unmarshal(b, &comps)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	// print what we got if verbose is set
	if verbose != nil {
		if *verbose {
			fmt.Printf("Components: %v\n", string(b))
		}
	}

	return comps, nil
}

func (client *SmdClient) FetchRedfishEndpoints(opts ...util.Option) ([]RedfishEndpoint, error) {
	var (
		params  = util.GetParams(opts...)
		verbose = util.Get[bool](params, "verbose")
		eps     = []RedfishEndpoint{}
	)

	b, err := client.makeRequest("/Inventory/RedfishEndpoints")
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP resquest: %v", err)
	}
	if !json.Valid(b) {
		return nil, fmt.Errorf("expected valid JSON response: %v", string(b))
	}
	var tmp map[string]any
	err = json.Unmarshal(b, &tmp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	b, err = json.Marshal(tmp["RedfishEndpoints"].([]any))
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %v", err)
	}
	err = json.Unmarshal(b, &eps)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if verbose != nil {
		if *verbose {
			fmt.Printf("Redfish endpoints: %v\n", string(b))
		}
	}

	return eps, nil
}

func (client *SmdClient) makeRequest(endpoint string) ([]byte, error) {
	if client == nil {
		return nil, fmt.Errorf("client is nil")
	}

	// fetch DHCP related information from SMD's endpoint:
	url := fmt.Sprintf("%s:%d/hsm/v2%s", client.Host, client.Port, endpoint)
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
