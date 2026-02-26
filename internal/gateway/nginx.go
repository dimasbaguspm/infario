package gateway

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GatewayDeployment represents deployment data needed for nginx configuration.
type GatewayDeployment struct {
	ID          string
	Hash        string
	ProjectID   string
	ProjectName string
	EntryPath   *string
}

// NginxGateway manages dynamic nginx configuration generation.
type NginxGateway struct {
	configDir  string
	domain     string
	storageDir string
}

// NewNginxGateway creates a new nginx gateway.
func NewNginxGateway(configDir, domain, storageDir string) *NginxGateway {
	return &NginxGateway{
		configDir:  configDir,
		domain:     domain,
		storageDir: storageDir,
	}
}

// WriteProjectConfig generates and writes nginx configuration for a project's deployments.
// If deployments list is empty, removes the config file instead.
func (ng *NginxGateway) WriteProjectConfig(projectID, projectName string, deployments []GatewayDeployment) error {
	// If no deployments, remove the config file
	if len(deployments) == 0 {
		return ng.RemoveProjectConfig(projectID)
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(ng.configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Build nginx server block
	config := fmt.Sprintf("# Auto-generated nginx config for project: %s\n", projectName)
	config += fmt.Sprintf("# Generated for deployments: %d\n\n", len(deployments))

	for _, dep := range deployments {
		// Server block
		serverName := fmt.Sprintf("%s.%s.%s", dep.Hash, projectName, ng.domain)
		config += fmt.Sprintln("server {\n")
		config += fmt.Sprintf("    server_name %s;\n", serverName)
		config += fmt.Sprintf("    listen 80;\n\n")

		// Deployment directory
		deploymentDir := fmt.Sprintf("/storage/deployments/%s/%s", dep.ProjectID, dep.ID)

		// Location block - handle both file and directory entry paths
		if dep.EntryPath != nil && *dep.EntryPath != "" && *dep.EntryPath != "/" {
			entryPath := *dep.EntryPath
			// Check if entry path is a file (has a file extension or looks like a file path)
			isFile := strings.Contains(entryPath, ".") && !strings.HasSuffix(entryPath, "/")

			if isFile {
				// Entry path is a file - extract directory and filename
				dirPart := filepath.Dir(entryPath)
				filePart := filepath.Base(entryPath)

				// Normalize "/" directory
				if dirPart == "." {
					dirPart = ""
				}

				rootPath := deploymentDir
				if dirPart != "" && dirPart != "/" {
					rootPath = deploymentDir + dirPart + "/"
				} else if dirPart == "/" {
					rootPath = deploymentDir + "/"
				}

				config += fmt.Sprintf("    location / {\n")
				config += fmt.Sprintf("        alias %s;\n", rootPath)
				config += fmt.Sprintf("        try_files $uri /%s =404;\n", filePart)
				config += fmt.Sprintf("    }\n\n")
			} else {
				// Entry path is a directory - serve with path prefix
				config += fmt.Sprintf("    location %s {\n", entryPath)
				config += fmt.Sprintf("        alias %s;\n", deploymentDir+entryPath)
				config += fmt.Sprintf("        try_files $uri =404;\n")
				config += fmt.Sprintf("    }\n\n")
			}
		} else {
			// No entry path or root entry path - serve from deployment root
			config += fmt.Sprintf("    location / {\n")
			config += fmt.Sprintf("        root %s;\n", deploymentDir)
			config += fmt.Sprintf("        index index.html;\n")
			config += fmt.Sprintf("    }\n\n")
		}

		config += fmt.Sprintf("}\n\n")
	}

	// Write to file
	configPath := filepath.Join(ng.configDir, projectID+".conf")
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// RemoveProjectConfig deletes the configuration file for a project.
// Silently succeeds if the file does not exist.
func (ng *NginxGateway) RemoveProjectConfig(projectID string) error {
	configPath := filepath.Join(ng.configDir, projectID+".conf")
	err := os.Remove(configPath)

	// Ignore "file not found" errors
	if os.IsNotExist(err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("failed to remove config file: %w", err)
	}

	return nil
}
