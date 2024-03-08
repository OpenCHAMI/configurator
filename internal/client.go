package configurator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type SmdClient struct {
	http.Client
	Host        string
	Port        int
	AccessToken string
}

func (client *SmdClient) FetchDNS(config *Config) error {
	// fetch DNS related information from SMD's endpoint:
	return nil
}

func (client *SmdClient) FetchEthernetInterfaces() ([]EthernetInterface, error) {
	if client == nil {
		return nil, fmt.Errorf("client is nil")
	}
	// fetch DHCP related information from SMD's endpoint:
	url := fmt.Sprintf("%s:%d/hsm/v2/Inventory/EthernetInterfaces", client.Host, client.Port)
	req, err := http.NewRequest(http.MethodGet, url, bytes.NewBuffer([]byte{}))

	// include access token in authorzation header if found
	if client.AccessToken != "" {
		req.Header.Add("Authorization", "Bearer "+client.AccessToken)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create new HTTP request: %v", err)
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read HTTP response: %v", err)
	}

	// unmarshal JSON and extract
	eths := []EthernetInterface{} // []map[string]any{}
	json.Unmarshal(b, &eths)
	fmt.Printf("ethernet interfaces: %v\n", string(b))

	return eths, nil
}
