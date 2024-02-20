package tentou

// Containerlab topology config definition based on github.com/srl-labs/containerlab v0.34.0
type Config struct {
	Type     string    `json:"type,omitempty"`
    Version  string    `json:"version,omitempty"`
	Name     string    `json:"name,omitempty"`
	Prefix   *string   `json:"prefix,omitempty"`
	// Mgmt     *MgmtNet  `json:"mgmt,omitempty"`
	// Topology *Topology `json:"topology,omitempty"`
	// Infra    *Infra    `json:"infra,omitempty`
	Networks map[string]*NetworkDefinition `yaml:"networks,omitempty"`
	// Nodes    map[string]*NodeDefinition `yaml:"nodes,omitempty"`
    Nodes []*NodeDefinition `yaml:"nodes,omitempty"`
}

type NetworkDefinition struct {
	Name                 string            `yaml:name,omitempty`
}

type Files struct {
    Name    string `yaml:name,omitempty`
    Src     string `yaml:src,omitempty`
    Dst     string `yaml:dst,omitempty`
}

type Facility struct {
    Name    string `yaml:name,omitempty`
    Type    string `yaml:type,omitempty`
    Settings []*Setting `yaml:settings,omitempty`
}

type Setting struct {
    Nets *Net `yaml:"nets,omitempty"`
    Files []*Files `yaml:files:omitempty`
    Cmds []string `yaml:"cmds,omitempty"`
}

type Net struct {
    Name string `yaml:"name,omitempty"`
    BindIp string `yaml:"bindip,omitempty"`
}

type NodeDefinition struct {
    Name                 string            `yaml:"name,omitempty"`
    Type                 string            `yaml:"type,omitempty"`
    Os                   string            `yaml:"os,omitempty"`
    Facilities           []*Facility       `yaml:"facilities,omitempty`
}

type LinkConfig struct {
	Endpoints []string               `yaml:"endpoints,flow"`
	Labels    map[string]string      `yaml:"labels,omitempty"`
	Vars      map[string]interface{} `yaml:"vars,omitempty"`
}

type ConfigDispatcher struct {
	Vars map[string]interface{} `yaml:"vars,omitempty"`
}

type Extras struct {
	SRLAgents []string `yaml:"srl-agents,omitempty"`
	// Nokia SR Linux agents. As of now just the agents spec files can be provided here
	MysocketProxy string `yaml:"mysocket-proxy,omitempty"`
	// Proxy address that mysocketctl will use
	CeosCopyToFlash []string `yaml:"ceos-copy-to-flash,omitempty"`
	// paths to files which are to be copied to ceos flash dir
}
