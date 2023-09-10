package commands

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/killbill/kbcli/v3/kbclient/nodes_info"
	"github.com/killbill/kbcli/v3/kbcmd/cmdlib"
	"github.com/killbill/kbcli/v3/kbcommon"
	"github.com/killbill/kbcli/v3/kbmodel"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
)

var nodesInfoFormatter = cmdlib.Formatter{
	Columns: []cmdlib.Column{
		{
			Name: "Node Name",
			Path: "$.nodeName",
		},
		{
			Name: "Boot Time",
			Path: "$.bootTime",
		},
		{
			Name: "Last Updated Date",
			Path: "$.lastUpdatedDate",
		},
		{
			Name: "Killbill Version",
			Path: "$.kbVersion",
		},
		{
			Name: "API Version",
			Path: "$.apiVersion",
		},
		{
			Name: "Plugin API Version",
			Path: "$.pluginApiVersion",
		},
		{
			Name: "Common Version",
			Path: "$.commonVersion",
		},
		{
			Name: "Platform Version",
			Path: "$.platformVersion",
		},
	},
	SubItems: []cmdlib.SubItem{
		{
			Name:      "Plugins",
			FieldName: "PluginsInfo",
			Formatter: &pluginInfoFormatter,
		},
	},
}

var pluginInfoFormatter = cmdlib.Formatter{
	Columns: []cmdlib.Column{
		{
			Name: "Plugin Name",
			Path: "$.pluginName",
		},
		{
			Name: "Bundle Symbolic Name",
			Path: "$.bundleSymbolicName",
		},
		{
			Name: "Is configured to Start",
			Path: "$.isSelectedToStart",
		},
		{
			Name: "Plugin State",
			Path: "$.state",
		},
		{
			Name: "Plugin Services",
			Path: "$.services",
		},
	},
}

type PluginData struct {
	Type       string            `yaml:"type"`
	ArtifactId string            `yaml:"artifact_id"`
	Versions   map[string]string `yaml:"versions"`
}

func installPlugin(ctx context.Context, o *cmdlib.Options) error {
	if len(o.Args) == 0 {
		return cmdlib.ErrorInvalidArgs
	}

	// Fetch the YAML file
	respYaml, err := http.Get("https://github.com/killbill/killbill-cloud/raw/master/kpm/lib/kpm/plugins_directory.yml")
	if err != nil {
		fmt.Println("Error fetching the YAML file:", err)
		return err
	}
	defer respYaml.Body.Close()

	// Read the response body
	data, err := io.ReadAll(respYaml.Body)
	if err != nil {
		fmt.Println("Error reading the response body:", err)
		return err
	}

	// Parse the YAML content
	pluginsMap := make(map[string]PluginData)
	err = yaml.Unmarshal(data, &pluginsMap)
	if err != nil {
		fmt.Println("Error parsing the YAML content:", err)
		return err
	}

	// Extract the required positional parameters
	pluginKey := o.Args[0]
	pluginType := ""
	pluginVersion := "" // default value
	artifactId := ""
	groupId := ""

	// Placeholder for optional named parameters, in a real scenario you'd use a flag parsing library like "flag"
	pluginPackaging := ""
	pluginClassifier := "null"

	// Check if pluginKey is standard
	if pluginData, exists := pluginsMap[":"+pluginKey]; exists {
		pluginType = pluginData.Type
		artifactId = pluginData.ArtifactId
		// If a Kill Bill version is provided, try to map it to a plugin version
		if len(o.Args) > 1 && o.Args[1] != "" {
			if version, ok := pluginData.Versions[o.Args[1]]; ok {
				pluginVersion = version
			}
		}
	} else {
		// If not a standard plugin, artifactId and groupId are required
		if len(o.Args) < 3 {
			fmt.Println("Missing required arguments for non-standard plugin")
			return err
		}
		pluginVersion = o.Args[1]
		artifactId = o.Args[2]
		groupId = o.Args[3]
		if len(o.Args) > 4 {
			pluginType = o.Args[4]
		}
	}

	// Set default values based on pluginType
	if pluginType == "java" {
		pluginPackaging = "jar"
	} else if pluginType == "ruby" {
		pluginPackaging = "tar.gz"
	}

	pluginProperties := []*kbmodel.NodeCommandProperty{
		{
			Key:   "pluginKey",
			Value: pluginKey,
		},
		{
			Key:   "pluginType",
			Value: pluginType,
		},
	}

	if artifactId != "" {
		pluginProperties = append(pluginProperties, &kbmodel.NodeCommandProperty{
			Key:   "pluginArtifactId",
			Value: artifactId,
		})
	}

	if groupId != "" {
		pluginProperties = append(pluginProperties, &kbmodel.NodeCommandProperty{
			Key:   "pluginGroupId",
			Value: groupId,
		})
	}

	if pluginVersion != "" {
		pluginProperties = append(pluginProperties, &kbmodel.NodeCommandProperty{
			Key:   "pluginVersion",
			Value: pluginVersion,
		})
	}

	if pluginPackaging != "" {
		pluginProperties = append(pluginProperties, &kbmodel.NodeCommandProperty{
			Key:   "pluginPackaging",
			Value: pluginPackaging,
		})
	}

	if pluginClassifier != "null" {
		pluginProperties = append(pluginProperties, &kbmodel.NodeCommandProperty{
			Key:   "pluginClassifier",
			Value: pluginClassifier,
		})
	}

	command := &kbmodel.NodeCommand{
		IsSystemCommandType:   true,
		NodeCommandType:       "INSTALL_PLUGIN",
		NodeCommandProperties: pluginProperties,
	}

	params := &nodes_info.TriggerNodeCommandParams{
		Body: command,
	}

	resp, err := o.Client().NodesInfo.TriggerNodeCommand(ctx, params)
	if err != nil {
		return err
	}
	if resp.IsSuccess() {
		o.Print(resp)
	}
	return nil
}

func managePlugin(ctx context.Context, o *cmdlib.Options, commandType, version string) error {
	if len(o.Args) < 1 {
		return cmdlib.ErrorInvalidArgs
	}

	// Extract the required positional parameters
	pluginKey := o.Args[0]

	pluginProperties := []*kbmodel.NodeCommandProperty{
		{
			Key:   "pluginKey",
			Value: pluginKey,
		},
	}

	// Add the plugin version if provided
	if version != "" {
		pluginProperties = append(pluginProperties, &kbmodel.NodeCommandProperty{
			Key:   "pluginVersion",
			Value: version,
		})
	}

	command := &kbmodel.NodeCommand{
		IsSystemCommandType:   true,
		NodeCommandType:       commandType,
		NodeCommandProperties: pluginProperties,
	}

	params := &nodes_info.TriggerNodeCommandParams{
		Body: command,
	}

	resp, err := o.Client().NodesInfo.TriggerNodeCommand(ctx, params)
	if err != nil {
		o.Print(err)
		return err
	}
	if resp.IsSuccess() {
		o.Print(resp)
	}
	return nil
}

func startPlugin(ctx context.Context, o *cmdlib.Options) error {
	return managePlugin(ctx, o, "START_PLUGIN", "")
}

func stopPlugin(ctx context.Context, o *cmdlib.Options) error {
	return managePlugin(ctx, o, "STOP_PLUGIN", "")
}

func restartPlugin(ctx context.Context, o *cmdlib.Options) error {
	return managePlugin(ctx, o, "RESTART_PLUGIN", "")
}

func uninstallPlugin(ctx context.Context, o *cmdlib.Options) error {
	if len(o.Args) != 2 {
		return cmdlib.ErrorInvalidArgs
	}
	pluginVersion := o.Args[1]
	return managePlugin(ctx, o, "UNINSTALL_PLUGIN", pluginVersion)
}

func getNodesInfo(ctx context.Context, o *cmdlib.Options) error {

	if len(o.Args) != 0 {
		return cmdlib.ErrorInvalidArgs
	}

	params := &nodes_info.GetNodesInfoParams{}

	resp, err := o.Client().NodesInfo.GetNodesInfo(ctx, params)
	if err != nil {
		return err
	}
	if resp.IsSuccess() {

		o.Print(resp.Payload)
		plugins := resp.GetPayload()[0].PluginsInfo
		o.Print(plugins)
		bodyBytes, err := io.ReadAll(resp.HttpResponse.Body())
		if err != nil {
			return err
		}
		o.Print(string(bodyBytes))
	}

	return nil
}

func registerNodesInfoCommands(r *cmdlib.App) {
	cmdlib.AddFormatter(reflect.TypeOf(&kbmodel.NodeInfo{}), nodesInfoFormatter)
	cmdlib.AddFormatter(reflect.TypeOf(&kbmodel.PluginInfo{}), pluginInfoFormatter)
	cmdlib.AddFormatter(reflect.TypeOf(&kbcommon.KillbillError{}), cmdlib.Formatter{
		Columns: []cmdlib.Column{
			{
				Name: "Node Command Failed to be Triggered",
				Path: "$.HTTPCode",
			},
		},
		CustomFn: cmdlib.CustomFormatter(func(v interface{}, fo cmdlib.FormatOptions) cmdlib.Output {
			if errorResponse, ok := v.(*kbcommon.KillbillError); ok {
				return cmdlib.Output{
					Title:   "Node Command Failed to be Triggered",
					Columns: []string{"Formatted Value"},
					Rows: []cmdlib.OutputRow{
						{
							Values:   []string{fmt.Sprintf("HTTP CODE: %d - Likely a bad param", errorResponse.HTTPCode)},
							Children: nil, // No child outputs in this example
						},
					},
				}
			}
			return cmdlib.Output{
				Title:   "Node Command Failed to be Triggered",
				Columns: []string{"Formatted Value"},
				Rows: []cmdlib.OutputRow{
					{
						Values:   []string{fmt.Sprintf("Likely a bad param")},
						Children: nil, // No child outputs in this example
					},
				},
			}
		}),
	})

	cmdlib.AddFormatter(reflect.TypeOf(&nodes_info.TriggerNodeCommandAccepted{}), cmdlib.Formatter{
		Columns: []cmdlib.Column{
			{
				Name: "Node Command Successfully Triggered",
				Path: "",
			},
		},
		CustomFn: cmdlib.CustomFormatter(func(v interface{}, fo cmdlib.FormatOptions) cmdlib.Output {
			//if i, ok := v.(int); ok {
			return cmdlib.Output{
				Title:   "Node Command Successfully Triggered",
				Columns: []string{"Formatted Value"},
				Rows: []cmdlib.OutputRow{
					{
						Values:   []string{"True - Remember to check the node logs for the result"},
						Children: nil, // No child outputs in this example
					},
				},
			}
			//}
			/*
				return cmdlib.Output{
					Title: "Error", Columns: []string{"Node Command Successfully Triggered"}, Rows: []cmdlib.OutputRow{{Values: []string{"Invalid response"}}},
				}
			*/
		}),
	})

	// Register top level command
	r.Register("", cli.Command{
		Name:  "nodes-info",
		Usage: "Node information and pluging Management",
	}, nil)

	r.Register("nodes-info", cli.Command{
		Name: "get-nodes-info",
		Usage: `Retrieves all the nodes infos
		`,
	}, getNodesInfo)

	r.Register("nodes-info", cli.Command{
		Name: "start-plugin",
		Usage: `
		Usage: [command] <pluginKey>
			`,
	}, startPlugin)

	r.Register("nodes-info", cli.Command{
		Name: "stop-plugin",
		Usage: `
		Usage: [command] <pluginKey>
			`,
	}, stopPlugin)

	r.Register("nodes-info", cli.Command{
		Name: "restart-plugin",
		Usage: `
		Usage: [command] <pluginKey>
			`,
	}, restartPlugin)

	r.Register("nodes-info", cli.Command{
		Name: "uninstall-plugin",
		Usage: `
		Usage: [command] <pluginKey>
			`,
	}, uninstallPlugin)

	r.Register("nodes-info", cli.Command{
		Name: "install-plugin",
		UsageText: `
		Usage: [command] <pluginKey> [secondParam] [thirdParam] [fourthParam]

Positional Arguments:
  pluginKey          The key identifier for the plugin. (Required)

  secondParam        Behavior depends on whether the pluginKey refers to a standard or custom plugin:
                     - Standard Plugin: Specifies the version of the Kill Bill instance, allowing the system to map to the appropriate plugin version. (Optional)
                     - Custom Plugin: Represents the Maven artifactId for the plugin. (Required)

  thirdParam         Only relevant for custom plugins:
                     - Represents the Maven groupId for the plugin. (Required if secondParam is the artifactId)

  fourthParam        Only relevant for custom plugins:
                     - Represents the type of the plugin - either 'ruby' or 'java'. Determines default packaging values. (Required if thirdParam is the groupId)

Note: 
  - For plugins hosted on the Kill Bill GitHub organization, the system will attempt to map the Kill Bill version with the appropriate plugin version.
  - The default packaging for a 'java' pluginType is 'jar', and for 'ruby', it's 'tar.gz'.
  - If you're unsure about the pluginKey being standard or custom, refer to the official plugins directory: https://github.com/killbill/killbill-cloud/blob/master/kpm/lib/kpm/plugins_directory.yml
		`,
	}, installPlugin)

}
