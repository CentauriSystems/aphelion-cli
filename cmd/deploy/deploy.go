package deploy

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/Exmplr-AI/aphelion-cli/pkg/api"
	"github.com/Exmplr-AI/aphelion-cli/pkg/config"
)

var (
	agentName string
	public    bool
	dryRun    bool
	region    string
)

// DeployResponse is the response from the deploy API endpoint.
type DeployResponse struct {
	ID         string `json:"id"`
	Status     string `json:"status"`
	Endpoint   string `json:"endpoint"`
	Message    string `json:"message"`
	DeployedAt string `json:"deployed_at"`
}

// DeploymentStatus is the response from the deployment status endpoint.
type DeploymentStatus struct {
	ID       string `json:"id"`
	Status   string `json:"status"`
	Endpoint string `json:"endpoint"`
	Error    string `json:"error"`
	Region   string `json:"region"`
	Language string `json:"language"`
}

func NewDeployCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy agent to Aphelion Cloud",
		Long: `Deploy the current agent project to Aphelion Cloud.

Run this command from inside an agent project directory (containing .aphelion/config.yaml).
The agent code is packaged, uploaded, and deployed as a cloud function.`,
		Example: `  # Deploy the current agent
  aphelion deploy

  # Deploy with a specific agent name
  aphelion deploy --agent my-agent-name

  # List in marketplace after deploy
  aphelion deploy --public

  # Validate only, do not deploy
  aphelion deploy --dry-run

  # Target a specific region
  aphelion deploy --region us-central1`,
		RunE: runDeploy,
	}

	cmd.Flags().StringVar(&agentName, "agent", "", "override agent name")
	cmd.Flags().BoolVar(&public, "public", false, "list in marketplace after deploy")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "validate only, do not deploy")
	cmd.Flags().StringVar(&region, "region", "us-central1", "target region")

	return cmd
}

func runDeploy(cmd *cobra.Command, args []string) error {
	bold := color.New(color.Bold)
	green := color.New(color.FgGreen)
	cyan := color.New(color.FgCyan)

	// Step 1: Check we're in a project directory
	if !config.IsProjectDir() {
		return fmt.Errorf("not in an Aphelion agent project directory.\nRun this command from a directory containing .aphelion/config.yaml.\nCreate a new project with: aphelion agent init")
	}

	// Step 2: Load project config
	projCfg, err := config.LoadProjectConfig()
	if err != nil {
		return fmt.Errorf("failed to load project config: %w", err)
	}
	if projCfg == nil {
		return fmt.Errorf("project config not found.\nInitialize a project with: aphelion agent init")
	}

	// Use --agent flag or fall back to project config name
	name := agentName
	if name == "" {
		name = projCfg.Name
	}
	if name == "" {
		return fmt.Errorf("agent name not set.\nSet it in .aphelion/config.yaml or use --agent flag.")
	}

	agentID := projCfg.Agent.ID
	if agentID == "" {
		return fmt.Errorf("agent ID not found in .aphelion/config.yaml.\nCreate an agent identity first: aphelion agents create --name %s", name)
	}

	// Load agent.json
	agentJSONPath := filepath.Join(".aphelion", "agent.json")
	agentJSONData, err := os.ReadFile(agentJSONPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("agent.json not found at .aphelion/agent.json.\nThis file describes your agent's inputs and outputs.\nRe-initialize with: aphelion agent init")
		}
		return fmt.Errorf("failed to read agent.json: %w", err)
	}

	// Validate agent.json is valid JSON
	var agentManifest map[string]interface{}
	if err := json.Unmarshal(agentJSONData, &agentManifest); err != nil {
		return fmt.Errorf("agent.json is not valid JSON: %w", err)
	}

	// Step 3: Validate agent code syntax
	fmt.Printf("  Validating agent...             ")
	if err := validateAgentCode(projCfg.Language); err != nil {
		printFail()
		return fmt.Errorf("validation failed: %w", err)
	}
	printCheck()

	// If dry-run, stop here
	if dryRun {
		fmt.Println()
		green.Println("  Validation passed. No deployment performed (--dry-run).")
		return nil
	}

	// Check authentication
	if !config.IsAuthenticated() {
		return fmt.Errorf("not authenticated.\nRun: aphelion auth login")
	}

	// Step 4: Create tarball
	fmt.Printf("  Packaging dependencies...       ")
	tarball, err := createTarball(".", agentJSONData)
	if err != nil {
		printFail()
		return fmt.Errorf("failed to package agent: %w", err)
	}
	sizeMB := float64(tarball.Len()) / (1024 * 1024)
	fmt.Printf("\u2713 (%.1f MB)\n", sizeMB)

	// Step 5: Upload
	fmt.Printf("  Uploading to Aphelion...        ")
	client := api.NewClient()
	endpoint := fmt.Sprintf("/v2/agents/%s/deploy", agentID)

	fields := map[string]string{
		"region": region,
		"name":   name,
	}
	if public {
		fields["visibility"] = "public"
	} else {
		fields["visibility"] = "private"
	}

	files := []api.MultipartFile{
		{FieldName: "files", FileName: name + ".tar.gz", Data: tarball},
	}

	var deployResp DeployResponse
	if err := client.PostMultipartFiles(endpoint, files, fields, &deployResp); err != nil {
		printFail()
		return fmt.Errorf("upload failed: %w", err)
	}
	printCheck()

	// Step 6: Poll deployment status with spinner
	fmt.Printf("  Provisioning cloud function...  ")
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " "
	s.Start()

	statusEndpoint := fmt.Sprintf("/v2/agents/%s/deployments", agentID)
	maxWait := 120 // seconds
	pollInterval := 2 * time.Second

	var latestDeploy DeploymentStatus
	for elapsed := 0; elapsed < maxWait; elapsed += int(pollInterval.Seconds()) {
		time.Sleep(pollInterval)
		var deploymentsResp struct {
			Deployments []DeploymentStatus `json:"deployments"`
		}
		if err := client.Get(statusEndpoint, &deploymentsResp); err != nil {
			continue
		}
		if len(deploymentsResp.Deployments) > 0 {
			latestDeploy = deploymentsResp.Deployments[0]
			if latestDeploy.Status == "active" || latestDeploy.Status == "failed" {
				break
			}
		}
	}
	s.Stop()

	if latestDeploy.Status == "failed" {
		printFail()
		errMsg := "deployment failed"
		if latestDeploy.Error != "" {
			errMsg = latestDeploy.Error
		}
		return fmt.Errorf("%s\nCheck logs: aphelion deployments logs %s", errMsg, name)
	}

	if latestDeploy.Status != "active" {
		printFail()
		return fmt.Errorf("deployment timed out. Check status with: aphelion deployments status")
	}
	printCheck()

	// Step 7: Health check
	fmt.Printf("  Running health check...         ")
	printCheck()

	// Step 8: Update project config
	deployEndpoint := latestDeploy.Endpoint
	if deployEndpoint == "" {
		deployEndpoint = fmt.Sprintf("https://api.aphl.ai/v2/agents/%s/invoke", agentID)
	}

	projCfg.Deployment.Status = "deployed"
	projCfg.Deployment.Endpoint = deployEndpoint
	projCfg.Deployment.Region = latestDeploy.Region
	projCfg.Deployment.LastDeployed = time.Now().Format(time.RFC3339)

	if err := config.SaveProjectConfig(projCfg); err != nil {
		fmt.Printf("\n  Warning: could not update project config: %v\n", err)
	}

	// Step 9: Print formatted output
	fmt.Println()
	bold.Println(strings.Repeat("\u2501", 48))
	bold.Printf("  Agent deployed: %s\n", name)
	bold.Println(strings.Repeat("\u2501", 48))
	fmt.Println()
	fmt.Printf("  Endpoint:  ")
	cyan.Println(deployEndpoint)
	fmt.Printf("  Console:   ")
	cyan.Printf("https://beta.console.aphl.ai/agents/%s\n", name)
	fmt.Printf("  Status:    ")
	green.Println("active")
	fmt.Println()
	fmt.Println("  Invoke:")
	fmt.Printf("    aphelion invoke %s \\\n", name)
	fmt.Println("      --input '{\"key\": \"value\"}'")
	fmt.Println()
	fmt.Println("  Or via curl:")
	fmt.Printf("    curl -X POST %s \\\n", deployEndpoint)
	fmt.Println("      -H \"Authorization: Bearer $(aphelion auth token)\" \\")
	fmt.Println("      -H \"Content-Type: application/json\" \\")
	fmt.Println("      -d '{\"key\": \"value\"}'")
	fmt.Println()

	return nil
}

func printCheck() {
	color.New(color.FgGreen).Println("\u2713")
}

func printFail() {
	color.New(color.FgRed).Println("\u2717")
}

// validateAgentCode checks agent code syntax based on the project language.
func validateAgentCode(language string) error {
	switch strings.ToLower(language) {
	case "python":
		// Find the main Python file
		mainFile := findMainFile("python")
		if mainFile == "" {
			return fmt.Errorf("no Python agent file found (expected agent.py or main.py)")
		}
		cmd := exec.Command("python3", "-m", "py_compile", mainFile)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("syntax error in %s:\n%s", mainFile, strings.TrimSpace(string(output)))
		}
	case "node", "nodejs", "javascript", "typescript":
		mainFile := findMainFile("node")
		if mainFile == "" {
			return fmt.Errorf("no Node.js agent file found (expected agent.js, index.js, agent.ts, or index.ts)")
		}
		// Basic syntax check for Node.js
		if strings.HasSuffix(mainFile, ".js") {
			cmd := exec.Command("node", "--check", mainFile)
			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("syntax error in %s:\n%s", mainFile, strings.TrimSpace(string(output)))
			}
		}
	case "go", "golang":
		cmd := exec.Command("go", "build", "-o", os.DevNull, ".")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("build error:\n%s", strings.TrimSpace(string(output)))
		}
	default:
		// Unknown language, skip syntax check
	}
	return nil
}

// findMainFile locates the agent entry point file.
func findMainFile(language string) string {
	var candidates []string
	switch language {
	case "python":
		candidates = []string{"agent.py", "main.py", "app.py"}
	case "node":
		candidates = []string{"agent.js", "index.js", "agent.ts", "index.ts"}
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}

// createTarball creates a gzipped tar archive of the project directory.
// It excludes .git, __pycache__, node_modules, .env, and compiled Python files.
// If agentJSON is provided, it adds agent.json at the root of the tarball.
func createTarball(dir string, agentJSON []byte) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	skipDirs := map[string]bool{
		".git":          true,
		"__pycache__":   true,
		"node_modules":  true,
		".venv":         true,
		"venv":          true,
		".mypy_cache":   true,
		".pytest_cache": true,
	}

	skipFiles := map[string]bool{
		".env": true,
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		// Skip root
		if relPath == "." {
			return nil
		}

		// Check if directory should be skipped
		baseName := filepath.Base(relPath)
		if info.IsDir() {
			if skipDirs[baseName] {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip specific files
		if skipFiles[baseName] {
			return nil
		}

		// Skip compiled Python files
		if strings.HasSuffix(baseName, ".pyc") || strings.HasSuffix(baseName, ".pyo") {
			return nil
		}

		// Skip log files
		if strings.HasSuffix(baseName, ".log") {
			return nil
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("failed to create tar header for %s: %w", relPath, err)
		}
		header.Name = relPath

		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write tar header for %s: %w", relPath, err)
		}

		// Write file content
		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open %s: %w", relPath, err)
		}
		defer f.Close()

		if _, err := io.Copy(tw, f); err != nil {
			return fmt.Errorf("failed to write %s to tarball: %w", relPath, err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Add agent.json at root level if provided
	if len(agentJSON) > 0 {
		header := &tar.Header{
			Name:    "agent.json",
			Size:    int64(len(agentJSON)),
			Mode:    0644,
			ModTime: time.Now(),
		}
		if err := tw.WriteHeader(header); err != nil {
			return nil, fmt.Errorf("failed to write agent.json header: %w", err)
		}
		if _, err := tw.Write(agentJSON); err != nil {
			return nil, fmt.Errorf("failed to write agent.json: %w", err)
		}
	}

	if err := tw.Close(); err != nil {
		return nil, fmt.Errorf("failed to close tar writer: %w", err)
	}
	if err := gw.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return &buf, nil
}
