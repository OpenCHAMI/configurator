package configurator

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/OpenCHAMI/configurator/pkg/util"
)

type ClientOption func(*SmdClient)

// An struct that's meant to extend functionality of the base HTTP client by
// adding commonly made requests to SMD. The implemented functions are can be
// used in generator plugins to fetch data when it is needed to substitute
// values for the Jinja templates used.
type SmdClient struct {
	http.Client `json:"-"`
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	AccessToken string `yaml:"access-token"`
}

// Constructor function that allows supplying ClientOption arguments to set
// things like the host, port, access token, etc.
func NewSmdClient(opts ...ClientOption) SmdClient {
	client := SmdClient{}
	for _, opt := range opts {
		opt(&client)
	}
	return client
}

func WithClientHost(host string) ClientOption {
	return func(c *SmdClient) {
		c.Host = host
	}
}

func WithClientPort(port int) ClientOption {
	return func(c *SmdClient) {
		c.Port = port
	}
}

func WithClientAccessToken(token string) ClientOption {
	return func(c *SmdClient) {
		c.AccessToken = token
	}
}

func WithClientCertPool(certPool *x509.CertPool) ClientOption {
	return func(c *SmdClient) {
		c.Client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:            certPool,
				InsecureSkipVerify: true,
			},
			DisableKeepAlives: true,
			Dial: (&net.Dialer{
				Timeout:   120 * time.Second,
				KeepAlive: 120 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   120 * time.Second,
			ResponseHeaderTimeout: 120 * time.Second,
		}
	}
}

// FIXME: Need to check for errors when reading from a file
func WithCertPoolFile(certPath string) ClientOption {
	if certPath == "" {
		return func(sc *SmdClient) {}
	}
	cacert, _ := os.ReadFile(certPath)
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(cacert)
	return WithClientCertPool(certPool)
}

// Create a set of params with all default values.
func NewParams() util.Params {
	return util.Params{
		"verbose": false,
	}
}

func WithNetwork(network string) util.Option {
	return func(p util.Params) {
		p["network"] = network
	}
}

func WithIPAddress(ipAddr string) util.Option {
	return func(p util.Params) {
		p["ip"] = ipAddr
	}
}

func GetNetwork(p util.Params) string {
	if network, ok := p["network"].(string); ok {
		return network
	}

	// default value
	return ""
}

func GetIPAddress(p util.Params) string {
	if ip, ok := p["ip"].(string); ok {
		return ip
	}

	// default value
	return ""
}

// Fetch the ethernet interfaces from SMD service using its API. An access token
// may be required if the SMD service SMD_JWKS_URL envirnoment variable is set.
//
// TODO: Change the `Option` type being used here to reduce it's scope.
func (client *SmdClient) FetchEthernetInterfaces(opts ...util.Option) ([]EthernetInterface, error) {
	var (
		params   = util.ToDict(opts...)
		verbose  = util.GetVerbose(params)
		network  = GetNetwork(params)
		eths     = []EthernetInterface{}
		endpoint = "/Inventory/EthernetInterfaces"
	)

	// add network to endpoint fetch filter by network type
	if network != "" {
		endpoint = fmt.Sprintf("%s?Network=%s", endpoint, network)
	}

	// make request to SMD endpoint
	b, err := client.makeRequest(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to read HTTP response: %v", err)
	}

	// unmarshal response body JSON and extract in object
	err = json.Unmarshal(b, &eths)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	// print what we got if verbose is set
	if verbose {
		fmt.Printf("Ethernet Interfaces: %v\n", string(b))
	}

	return eths, nil
}

// Fetch the components from SMD using its API. An access token may be required if the SMD
// service SMD_JWKS_URL envirnoment variable is set.
func (client *SmdClient) FetchComponents(opts ...util.Option) ([]Component, error) {
	var (
		params  = util.ToDict(opts...)
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
		params  = util.ToDict(opts...)
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
