package tools

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Exmplr-AI/aphelion-cli/internal/utils"
	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
)

type ToolDescription struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Parameters  map[string]interface{} `json:"parameters"`
	Examples    []ToolExample          `json:"examples,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
}

type ToolExample struct {
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

func newDescribeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <tool-name>",
		Short: "Describe a tool's parameters and usage",
		Long:  "Show detailed information about a tool including parameters, schema, and examples",
		Args:  cobra.ExactArgs(1),
		RunE:  runDescribe,
	}

	return cmd
}

func runDescribe(cmd *cobra.Command, args []string) error {
	toolName := args[0]

	client := api.NewClient()

	// First try the direct describe endpoint
	endpoint := fmt.Sprintf("/v2/tools/%s/describe", toolName)
	var toolDesc ToolDescription
	if err := client.Get(endpoint, &toolDesc); err == nil {
		return outputToolDescription(toolDesc)
	}

	// Fallback: search for the service by name, then get its manifest
	var searchResp struct {
		Tools []struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"tools"`
	}
	if err := client.GetWithQuery("/search/tools", map[string]string{"q": toolName}, &searchResp); err != nil {
		return fmt.Errorf("failed to describe tool %s: %w", toolName, err)
	}

	if len(searchResp.Tools) == 0 {
		return fmt.Errorf("tool %q not found.\nSearch available tools: aphelion tools search <query>", toolName)
	}

	// Get the manifest for the first match
	serviceID := searchResp.Tools[0].ID
	manifestEndpoint := fmt.Sprintf("/services/%s/manifest", serviceID)
	if err := client.Get(manifestEndpoint, &toolDesc); err != nil {
		return fmt.Errorf("failed to get manifest for %s: %w", toolName, err)
	}

	return outputToolDescription(toolDesc)
}

func outputToolDescription(tool ToolDescription) error {
	outputFormat := viper.GetString("output")
	
	switch outputFormat {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(tool)
		
	case "yaml":
		return utils.OutputYAML(tool)
		
	default: // table format
		return outputToolDescriptionTable(tool)
	}
}

func outputToolDescriptionTable(tool ToolDescription) error {
	fmt.Printf("🔧 Tool: %s\n", tool.Name)
	fmt.Printf("📝 Description: %s\n", tool.Description)
	
	if tool.Version != "" {
		fmt.Printf("📦 Version: %s\n", tool.Version)
	}
	
	if len(tool.Tags) > 0 {
		fmt.Printf("🏷️  Tags: %v\n", tool.Tags)
	}
	
	fmt.Println("\n📋 Parameters:")
	if len(tool.Parameters) == 0 {
		fmt.Println("  No parameters required")
	} else {
		for name, schema := range tool.Parameters {
			fmt.Printf("  • %s\n", name)
			if schemaMap, ok := schema.(map[string]interface{}); ok {
				if desc, ok := schemaMap["description"].(string); ok {
					fmt.Printf("    Description: %s\n", desc)
				}
				if paramType, ok := schemaMap["type"].(string); ok {
					fmt.Printf("    Type: %s\n", paramType)
				}
				if required, ok := schemaMap["required"].(bool); ok && required {
					fmt.Printf("    Required: ✅\n")
				}
				if defaultVal, ok := schemaMap["default"]; ok {
					fmt.Printf("    Default: %v\n", defaultVal)
				}
			}
			fmt.Println()
		}
	}
	
	if len(tool.Examples) > 0 {
		fmt.Println("💡 Examples:")
		for i, example := range tool.Examples {
			fmt.Printf("  %d. %s\n", i+1, example.Description)
			if len(example.Parameters) > 0 {
				fmt.Printf("     Parameters: ")
				paramsJSON, _ := json.Marshal(example.Parameters)
				fmt.Println(string(paramsJSON))
			}
			fmt.Println()
		}
	}
	
	return nil
}