package tentou

import (
    "encoding/json"
    "fmt"
    // "net/netip"
    //"path/filepath"

    "github.com/goccy/go-yaml"

    "github.com/danna2019/dot2net/pkg/model"
)

const (
    DEFAULT_NAME = "tentou"
    DEFAULT_TYPE = "infrastructure"
    DEFAULT_VERSION = "0.1"
    DEFAULT_HOST_TYPE = "Physical"
    DEFAULT_HOST_OS = "ubuntu-server-20.04-std"
    DEFAULT_FACILITY_TYPE = "FRR"
    DEFAULT_NET_NAME = "net0"
    DEFAULT_VLAN_ID = "auto"
    DEFAULT_URL = "http://vmuser190.pub.starbed.org:8080/frr"
)

func getTentouNode(cfg *model.Config, n *model.Node) (*NodeDefinition, error) {
    // tentou node attributes
    ndef := &NodeDefinition{}
    if n.TentouAttr == nil {
        return ndef, nil
    }
    mapper := n.TentouAttr

    bytes, err := json.Marshal(mapper)
    if err != nil {
        return nil, err
    }
    err = json.Unmarshal(bytes, ndef)
    if err != nil {
        return nil, err
    }
    return ndef, nil
}

func getTentouLink(cfg *model.Config, conn *model.Connection) *LinkConfig {
    src := conn.Src.Node.Name + ":" + conn.Src.Name
    dst := conn.Dst.Node.Name + ":" + conn.Dst.Name
    link := LinkConfig{
        Endpoints: []string{src, dst},
    }
    return &link
}

func GetTentouInfra(cfg *model.Config, nm *model.NetworkModel) ([]byte, error) {

    config := &Config{
        Type: "",
        Version: "",
        Name: "",
        Networks: make([]*NetworkDefinition, 0),
        Nodes: make([]*NodeDefinition, 0),
    }

    // tentou global attributes
    gattr := cfg.GlobalSettings.TentouAttr
    for k, v := range gattr {
        switch k {
        case "type":
            typee, ok := v.(string)
            if !ok {
                return nil, fmt.Errorf("global.tentou.type must be string")
            }
            config.Type = typee
        case "version":
            version, ok := v.(string)
            if !ok {
                return nil, fmt.Errorf("global.tentou.type must be string")
            }
            config.Version = version
        case "name":
            name, ok := v.(string)
            if !ok {
                return nil, fmt.Errorf("global.tentou.name must be string")
            }
            config.Name = name
        default:
            return nil, fmt.Errorf("invalid field in global.tentou")
        }
    }

    // global settings
    if config.Name == "" {
        if cfg.Name != "" {
            config.Name = cfg.Name
        } else {
            config.Name = DEFAULT_NAME
        }
    }
    if config.Type == "" {
        if cfg.Type != "" {
            config.Type = cfg.Type
        } else {
            config.Type = DEFAULT_TYPE
        }
    }
    if config.Version == "" {
        if cfg.Version != "" {
            config.Version = cfg.Version
        } else {
            config.Version = DEFAULT_VERSION
        }
    }

    netdef := &NetworkDefinition{
        Name: DEFAULT_NET_NAME,
        Vlan: DEFAULT_VLAN_ID,
        Nodes: make([]*NodeOnNetwork, 0),
    }
    for _, node := range nm.Nodes {
        if node.Virtual {
            continue
        }
        n := &NodeOnNetwork{
            Name: node.Name,
            IPAddr: "192.168.0.1/24",
        }
        // nodeのloのIPが取れたら、それをnのIPAddrに入れる
        netdef.Nodes = append(netdef.Nodes, n)
    }
    config.Networks = append(config.Networks, netdef)

    for _, node := range nm.Nodes {
        // skip virtual nodes
        if node.Virtual {
            continue
        }

        // node settings
        ndef, err := getTentouNode(cfg, node)
        if err != nil {
            return nil, err
        }
        ndef.Name = node.Name
        ndef.Type = DEFAULT_HOST_TYPE
        ndef.Os = DEFAULT_HOST_OS

        fa := &Facility{
            Name: node.Name + "-" + DEFAULT_FACILITY_TYPE,
            Type: DEFAULT_FACILITY_TYPE,
            Settings: make([]*Setting, 0),
        }

        n := &Net{
            Name: DEFAULT_NET_NAME,
            BindIp: "{{ip." + node.Name + "." + DEFAULT_NET_NAME + "}}",
        }

        s := &Setting{
            Nets: n,
        }

        embed := node.Files.GetEmbeddedConfig()
        if embed != nil {
            // add inline configuration commands
            s.Cmds = append(s.Cmds, node.Files.GetEmbeddedConfig().Content...)
        }

        for _, filename := range node.Files.FileNames() {
            file := node.Files.GetFile(filename)
            if file.FileDefinition.Path == "" {
                continue
            }
            cfgpath := DEFAULT_URL + "/" + node.Name + "/" + file.FileDefinition.Name
            targetpath := file.FileDefinition.Path
            f := &Files{
                Name: file.FileDefinition.Name,
                Src: cfgpath,
                Dst: targetpath,
            }
            s.Files = append(s.Files, f)
        }

        fa.Settings = append(fa.Settings, s)
        ndef.Facilities = append(ndef.Facilities, fa)
        config.Nodes = append(config.Nodes, ndef)
    }

    bytes, err := yaml.Marshal(config)
    if err != nil {
        return nil, err
    }
    return bytes, nil
}
