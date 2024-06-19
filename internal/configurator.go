package configurator

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
}

type Node struct {
}

type BMC struct {
}
