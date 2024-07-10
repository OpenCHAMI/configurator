package configurator

import "encoding/json"

type IPAddr struct {
	IpAddress string `json:"IPAddress"`
	Network   string `json:"Network"`
}

type EthernetInterface struct {
	Id          string
	Description string
	MacAddress  string
	LastUpdate  string
	ComponentId string
	Type        string
	IpAddresses []IPAddr
}

type Component struct {
	ID                  string      `json:"ID"`
	Type                string      `json:"Type"`
	State               string      `json:"State,omitempty"`
	Flag                string      `json:"Flag,omitempty"`
	Enabled             *bool       `json:"Enabled,omitempty"`
	SwStatus            string      `json:"SoftwareStatus,omitempty"`
	Role                string      `json:"Role,omitempty"`
	SubRole             string      `json:"SubRole,omitempty"`
	NID                 json.Number `json:"NID,omitempty"`
	Subtype             string      `json:"Subtype,omitempty"`
	NetType             string      `json:"NetType,omitempty"`
	Arch                string      `json:"Arch,omitempty"`
	Class               string      `json:"Class,omitempty"`
	ReservationDisabled bool        `json:"ReservationDisabled,omitempty"`
	Locked              bool        `json:"Locked,omitempty"`
}

type RedfishEndpoint struct {
	ID          string `json:"ID"`
	Type        string `json:"Type"`
	Name        string `json:"Name,omitempty"` // user supplied descriptive name
	Hostname    string `json:"Hostname"`
	Domain      string `json:"Domain"`
	FQDN        string `json:"FQDN"`
	Enabled     bool   `json:"Enabled"`
	UUID        string `json:"UUID,omitempty"`
	User        string `json:"User"`
	Password    string `json:"Password"` // Temporary until more secure method
	UseSSDP     bool   `json:"UseSSDP,omitempty"`
	MACRequired bool   `json:"MACRequired,omitempty"`
	MACAddr     string `json:"MACAddr,omitempty"`
	IPAddr      string `json:"IPAddress,omitempty"`
}

type Node struct {
}

type BMC struct {
}
