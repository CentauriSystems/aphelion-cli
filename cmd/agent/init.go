package agent

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

var (
	initName        string
	initDescription string
	initLanguage    string
)

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new agent project",
		Long: `Scaffold a new agent project with all required files and configuration.

Interactively prompts for agent name, description, language, and tool subscriptions.
Creates a complete, runnable project directory with agent code, tests, and config.

Use flags for non-interactive mode:
  aphelion agent init --name my-agent --description "Does something" --language python`,
		RunE: runInit,
	}

	cmd.Flags().StringVar(&initName, "name", "", "Agent name (kebab-case)")
	cmd.Flags().StringVar(&initDescription, "description", "", "Agent description")
	cmd.Flags().StringVar(&initLanguage, "language", "", "Language: python, node, or go")

	return cmd
}

var kebabCaseRegex = regexp.MustCompile(`^[a-z][a-z0-9-]*[a-z0-9]$`)

func promptString(reader *bufio.Reader, prompt, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultVal)
	} else {
		fmt.Printf("%s: ", prompt)
	}
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return defaultVal
	}
	return line
}

func promptConfirm(reader *bufio.Reader, prompt string, defaultYes bool) bool {
	hint := "[Y/n]"
	if !defaultYes {
		hint = "[y/N]"
	}
	fmt.Printf("%s %s: ", prompt, hint)
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	if line == "" {
		return defaultYes
	}
	return line == "y" || line == "yes"
}

func runInit(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	// 1. Agent name
	name := initName
	if name == "" {
		for {
			name = promptString(reader, "Agent name (kebab-case)", "")
			if name == "" {
				fmt.Println("Agent name is required.")
				continue
			}
			if !kebabCaseRegex.MatchString(name) {
				fmt.Println("Name must be kebab-case: lowercase letters, numbers, and hyphens (e.g. review-management-agent)")
				continue
			}
			break
		}
	} else if !kebabCaseRegex.MatchString(name) {
		return fmt.Errorf("name must be kebab-case: lowercase letters, numbers, and hyphens (e.g. review-management-agent)")
	}

	// Check if directory already exists
	if _, err := os.Stat(name); err == nil {
		return fmt.Errorf("directory %q already exists", name)
	}

	// 2. Description
	description := initDescription
	if description == "" {
		description = promptString(reader, "Description", "An Aphelion agent")
	}

	// 3. Language
	language := initLanguage
	if language == "" {
		language = promptString(reader, "Language [python/node/go]", "python")
	}
	language = strings.ToLower(strings.TrimSpace(language))
	switch language {
	case "python", "node", "go":
		// valid
	case "javascript", "js":
		language = "node"
	case "":
		language = "python"
	default:
		return fmt.Errorf("unsupported language %q — choose python, node, or go", language)
	}

	// Detect non-interactive mode: all three flags provided
	nonInteractive := initName != "" && initDescription != "" && initLanguage != ""

	// 4. Tool subscriptions (skip in non-interactive mode)
	var subscribedTools []string
	if !nonInteractive {
		fmt.Print("Subscribe to tools (comma-separated, e.g. twilio, sendgrid): ")
		toolsLine, _ := reader.ReadString('\n')
		toolsLine = strings.TrimSpace(toolsLine)
		if toolsLine != "" {
			for _, t := range strings.Split(toolsLine, ",") {
				t = strings.TrimSpace(t)
				if t != "" {
					subscribedTools = append(subscribedTools, t)
				}
			}
		}
	}

	// 5. Create agent identity (auto-create in non-interactive mode)
	agentID := ""
	clientID := ""
	clientSecret := ""
	createIdentity := true
	if !nonInteractive {
		createIdentity = promptConfirm(reader, "Create agent identity now?", true)
	}
	if createIdentity {
		id, cid, csec, err := createAgentIdentity(name, description)
		if err != nil {
			fmt.Printf("\nWarning: Could not create agent identity: %v\n", err)
			fmt.Println("You can create one later with: aphelion agents create")
		} else {
			agentID = id
			clientID = cid
			clientSecret = csec
			fmt.Printf("\nAgent identity created.\n")
			fmt.Printf("  Agent ID:      %s\n", agentID)
			fmt.Printf("  Client ID:     %s\n", clientID)
			fmt.Printf("  Client Secret: %s (shown once — save it now)\n", csec)
		}
	}

	// 6. Create directory structure
	fmt.Printf("\nCreating project %s...\n", name)

	if err := createProjectScaffold(name, description, language, subscribedTools, agentID, clientID, clientSecret); err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	// 7. Print success
	entryFile := "agent.py"
	installCmd := "pip install -r requirements.txt"
	if language == "node" {
		entryFile = "index.js"
		installCmd = "npm install"
	}

	fmt.Println()
	fmt.Println("Agent project initialized successfully!")
	fmt.Println()
	fmt.Println("Created files:")
	fmt.Printf("  %s/%s\n", name, entryFile)
	if language == "python" {
		fmt.Printf("  %s/requirements.txt\n", name)
	} else if language == "node" {
		fmt.Printf("  %s/package.json\n", name)
	}
	fmt.Printf("  %s/.aphelion/config.yaml\n", name)
	fmt.Printf("  %s/.aphelion/agent.json\n", name)
	fmt.Printf("  %s/.aphelion/.gitignore\n", name)
	fmt.Printf("  %s/.env.example\n", name)
	fmt.Printf("  %s/README.md\n", name)
	if language == "python" {
		fmt.Printf("  %s/tests/test_agent.py\n", name)
	} else if language == "node" {
		fmt.Printf("  %s/tests/test_agent.js\n", name)
	}
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  cd %s\n", name)
	fmt.Printf("  %s\n", installCmd)
	fmt.Println("  cp .env.example .env    # fill in your values")
	fmt.Printf("  aphelion agent run %s --input '{\"key\": \"value\"}'\n", entryFile)
	fmt.Println()
	if len(subscribedTools) > 0 {
		fmt.Println("Subscribe to tools:")
		for _, t := range subscribedTools {
			fmt.Printf("  aphelion tools subscribe %s\n", t)
		}
		fmt.Println()
	}
	fmt.Println("Deploy when ready:")
	fmt.Println("  aphelion deploy")

	return nil
}

func createProjectScaffold(name, description, language string, tools []string, agentID, clientID, clientSecret string) error {
	// Create directories
	dirs := []string{
		name,
		filepath.Join(name, ".aphelion"),
		filepath.Join(name, "tests"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", d, err)
		}
	}

	// Helper to apply template replacements
	replace := func(tmpl string) string {
		s := tmpl
		s = strings.ReplaceAll(s, "{{.Name}}", name)
		s = strings.ReplaceAll(s, "{{.Description}}", description)
		return s
	}

	// Determine language-specific file details
	entryFile := "agent.py"
	depsFile := "requirements.txt"
	testFile := "test_agent.py"
	installCmd := "pip install -r requirements.txt"
	var agentContent, depsContent, testContent string

	switch language {
	case "python":
		agentContent = replace(pythonAgentTemplate)
		depsContent = requirementsTemplate
		testContent = replace(testPythonTemplate)
	case "node":
		entryFile = "index.js"
		depsFile = "package.json"
		testFile = "test_agent.js"
		installCmd = "npm install"
		agentContent = replace(nodeAgentTemplate)
		depsContent = replace(packageJSONTemplate)
		testContent = replace(testNodeTemplate)
	case "go":
		// For Go, use Python template as placeholder — Go agent SDK is future work
		agentContent = replace(pythonAgentTemplate)
		depsContent = requirementsTemplate
		testContent = replace(testPythonTemplate)
	}

	// Build tools YAML list
	toolsYAML := ""
	if len(tools) > 0 {
		for _, t := range tools {
			toolsYAML += fmt.Sprintf("    - %s\n", t)
		}
	} else {
		toolsYAML = "    []\n"
	}

	// Build project config
	projectConfig := projectConfigTemplate
	projectConfig = strings.ReplaceAll(projectConfig, "{{.Name}}", name)
	projectConfig = strings.ReplaceAll(projectConfig, "{{.Description}}", description)
	projectConfig = strings.ReplaceAll(projectConfig, "{{.Language}}", language)
	projectConfig = strings.ReplaceAll(projectConfig, "{{.AgentID}}", valueOrPlaceholder(agentID, ""))
	projectConfig = strings.ReplaceAll(projectConfig, "{{.ClientID}}", valueOrPlaceholder(clientID, ""))
	projectConfig = strings.ReplaceAll(projectConfig, "{{.ClientSecret}}", valueOrPlaceholder(clientSecret, ""))
	projectConfig = strings.ReplaceAll(projectConfig, "{{.ToolsList}}", toolsYAML)

	// Build README
	readmeContent := readmeTemplate
	readmeContent = strings.ReplaceAll(readmeContent, "{{.Name}}", name)
	readmeContent = strings.ReplaceAll(readmeContent, "{{.Description}}", description)
	readmeContent = strings.ReplaceAll(readmeContent, "{{.EntryFile}}", entryFile)
	readmeContent = strings.ReplaceAll(readmeContent, "{{.DepsFile}}", depsFile)
	readmeContent = strings.ReplaceAll(readmeContent, "{{.TestFile}}", testFile)
	readmeContent = strings.ReplaceAll(readmeContent, "{{.InstallCmd}}", installCmd)

	// Write all files
	type fileEntry struct {
		path    string
		content string
		perm    os.FileMode
	}

	files := []fileEntry{
		{filepath.Join(name, entryFile), agentContent, 0644},
		{filepath.Join(name, depsFile), depsContent, 0644},
		{filepath.Join(name, ".aphelion", "agent.json"), replace(agentJSONTemplate), 0644},
		{filepath.Join(name, ".aphelion", ".gitignore"), gitignoreTemplate, 0644},
		{filepath.Join(name, ".aphelion", "config.yaml"), projectConfig, 0600},
		{filepath.Join(name, ".env.example"), envExampleTemplate, 0644},
		{filepath.Join(name, "README.md"), readmeContent, 0644},
		{filepath.Join(name, "tests", testFile), testContent, 0644},
	}

	for _, f := range files {
		if err := os.WriteFile(f.path, []byte(f.content), f.perm); err != nil {
			return fmt.Errorf("failed to write %s: %w", f.path, err)
		}
	}

	return nil
}

func valueOrPlaceholder(val, placeholder string) string {
	if val != "" {
		return val
	}
	return placeholder
}

// createAgentIdentity calls POST /v2/agents to create an agent identity on the platform.
func createAgentIdentity(name, description string) (agentID, clientID, clientSecret string, err error) {
	token := config.GetAccessToken()
	if token == "" {
		return "", "", "", fmt.Errorf("not authenticated — run: aphelion auth login")
	}

	apiURL := config.GetAPIUrl()
	if apiURL == "" {
		apiURL = "https://api.aphl.ai"
	}

	payload := map[string]string{
		"name":        name,
		"description": description,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", apiURL+"/v2/agents", strings.NewReader(string(body)))
	if err != nil {
		return "", "", "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", "", "", fmt.Errorf("API returned status %d — you may need to re-authenticate: aphelion auth login", resp.StatusCode)
	}

	var result struct {
		ID           string `json:"id"`
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", "", fmt.Errorf("failed to parse API response: %w", err)
	}

	return result.ID, result.ClientID, result.ClientSecret, nil
}
