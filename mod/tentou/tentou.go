package tentou

import (
	"embed"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cpflat/dot2net/pkg/types"
)

// special parameter ".virtual" for nodes
// -> config files for the virtual nodes are not generated and just ignored

const VirtualNodeClassName = "virtual"

const TentouOutputFile = "infra.yaml"

const TentouNetworkNameParamName = "_ten_networkName"
const TentouImageParamName = "image"
const TentouFacilityParamName = "facility"
const TentouBaseURLParamName = "baseurl"
const TentouBindMountsParamName = "_ten_bindMounts"

const TentouYamlFormatName = "_tentouYaml"
const InfraCmdFormatName = "tentouInfraCmd"

// const TentouVtyshCLIFormatName = "tentouVtyshCLI"

const NetworkClassName = "_tentouNetwork"
const NodeClassName = "_tentouNode"
const InterfaceClassName = "_tentouInterface"

//go:embed templates/*
var templates embed.FS

type TentouModule struct {
	*types.StandardModule
}

func NewModule() types.Module {
	return &TentouModule{
		StandardModule: types.NewStandardModule(),
	}
}

func (m *TentouModule) UpdateConfig(cfg *types.Config) error {
	// add file format
	fileFormat := &types.FileFormat{
		Name:           TentouYamlFormatName,
		BlockSeparator: ", ",
	}
	cfg.AddFileFormat(fileFormat)
	fileFormat = &types.FileFormat{
		Name:           InfraCmdFormatName,
		LinePrefix:     "        - ",
		BlockSeparator: "\n",
	}
	cfg.AddFileFormat(fileFormat)
	// 	fileFormat = &types.FileFormat{
	// 		Name:          TinetVtyshCLIFormatName,
	// 		LineSeparator: "\" -c \"",
	// 		BlockPrefix:   "vtysh -c \"conf t\" -c \"",
	// 		BlockSuffix:   "\"",
	// 	}
	// 	cfg.AddFileFormat(fileFormat)

	// add file definition
	fileDef := &types.FileDefinition{
		Name:  TentouOutputFile,
		Path:  "",
		Scope: types.ClassTypeNetwork,
	}
	cfg.AddFileDefinition(fileDef)

	// add network class
	ct1 := &types.ConfigTemplate{File: TentouOutputFile}
	bytes, err := templates.ReadFile("templates/infra.yaml.network")
	if err != nil {
		return err
	}
	ct1.Template = []string{string(bytes)}

	networkClass := &types.NetworkClass{
		Name:            NetworkClassName,
		ConfigTemplates: []*types.ConfigTemplate{ct1},
	}
	cfg.AddNetworkClass(networkClass)

	// add node class
	ct1 = &types.ConfigTemplate{Name: "ten_cmds", Format: InfraCmdFormatName, Depends: []string{"startup"}}
	bytes, err = templates.ReadFile("templates/infra.yaml.node_ten_cmd")
	if err != nil {
		return err
	}
	ct1.Template = []string{string(bytes)}

	ct2 := &types.ConfigTemplate{Name: "ten_spec"}
	bytes, err = templates.ReadFile("templates/infra.yaml.node_ten_spec")
	if err != nil {
		return err
	}
	ct2.Template = []string{string(bytes)}

	//	ct3 := &types.ConfigTemplate{Name: "ten_config", Depends: []string{"ten_cmds"}}
	//	bytes, err = templates.ReadFile("templates/infra.yaml.node_ten_config")
	//	if err != nil {
	//		return err
	//	}
	//	ct3.Template = []string{string(bytes)}

	nodeClass := &types.NodeClass{
		Name: NodeClassName,
		//		ConfigTemplates: []*types.ConfigTemplate{ct1, ct2, ct3},
		ConfigTemplates: []*types.ConfigTemplate{ct1, ct2},
	}
	cfg.AddNodeClass(nodeClass)
	m.AddModuleNodeClassLabel(NodeClassName)

	// add interface class
	ct1 = &types.ConfigTemplate{Name: "ten_spec", Format: TentouYamlFormatName}
	bytes, err = templates.ReadFile("templates/infra.yaml.interface_spec")
	if err != nil {
		return err
	}
	ct1.Template = []string{string(bytes)}
	interfaceClass := &types.InterfaceClass{
		Name:            InterfaceClassName,
		ConfigTemplates: []*types.ConfigTemplate{ct1},
	}
	cfg.AddInterfaceClass(interfaceClass)
	m.AddModuleInterfaceClassLabel(InterfaceClassName)

	return nil
}

func (m TentouModule) GenerateParameters(cfg *types.Config, nm *types.NetworkModel) error {

	// set network name
	nm.AddParam(TentouNetworkNameParamName, cfg.Name)

	for _, node := range nm.Nodes {
		// generate file mount point descriptions
		bindItems := []string{}
		urlPath := "http://vmuser190.lab.starbed.org:8080/frr/"
		for _, fileDef := range cfg.FileDefinitions {
			if fileDef.Path == "" {
				continue
			}

			//dirpath, err := os.Getwd()
			//if err != nil {
			//	return fmt.Errorf("failed to obtain currrent directory")
			//}
			nameTag := "        - name: " + fileDef.Name
			srcPath := "\n          src: " + filepath.Join(urlPath, node.Name, fileDef.Name)
			dstPath := "\n          dst: " + fileDef.Path
			bindItems = append(bindItems, nameTag+srcPath+dstPath)
		}
		node.AddParam(TentouBindMountsParamName, strings.Join(bindItems, "\n"))
	}

	return nil
}

func (m TentouModule) CheckModuleRequirements(cfg *types.Config, nm *types.NetworkModel) error {
	// node config templates named startup
	flag := false
	for _, nc := range cfg.NodeClasses {
		for _, ct := range nc.ConfigTemplates {
			if ct.Name == "startup" {
				flag = true
			}
		}
	}
	if !flag {
		return fmt.Errorf("node config templates named startup is required")
	}

	// parameter {{ .image }}
	for _, node := range nm.Nodes {
		if node.IsVirtual() {
			break
		} else {
			_, err := node.GetParamValue(TentouImageParamName)
			if err != nil {
				return fmt.Errorf("every (non-virtual) node must have {{ .image }} parameter (none for %s)", node.Name)
			}
			_, err = node.GetParamValue(TentouBaseURLParamName)
			if err != nil {
				return fmt.Errorf("every (non-virtual) node must have {{ .base_url }} parameter (none for %s)", node.Name)
			}
		}
	}
	return nil
}
