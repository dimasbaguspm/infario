package gateway

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// GatewayDeployment represents deployment data needed for Traefik configuration.
type GatewayDeployment struct {
	ID          string
	Hash        string
	ProjectID   string
	ProjectName string
}

// Traefik dynamic config YAML structure types
type TraefikConfig struct {
	HTTP httpConfig `yaml:"http"`
}

type httpConfig struct {
	Routers  map[string]traefikRouter  `yaml:"routers"`
	Services map[string]traefikService `yaml:"services"`
}

type traefikRouter struct {
	Rule    string `yaml:"rule"`
	Service string `yaml:"service"`
}

type traefikService struct {
	LoadBalancer loadBalancer `yaml:"loadBalancer"`
}

type loadBalancer struct {
	Servers []server `yaml:"servers"`
}

type server struct {
	URL string `yaml:"url"`
}

// TraefikGateway manages dynamic Traefik configuration generation.
type TraefikGateway struct {
	configDir string
	domain    string
}

// NewTraefikGateway creates a new Traefik gateway.
func NewTraefikGateway(configDir, domain string) *TraefikGateway {
	return &TraefikGateway{
		configDir: configDir,
		domain:    domain,
	}
}

// WriteProjectConfig generates and writes Traefik configuration for a project's deployments.
// If deployments list is empty, removes the config file instead.
func (tg *TraefikGateway) WriteProjectConfig(projectID, projectName string, deployments []GatewayDeployment) error {
	// If no deployments, remove the config file
	if len(deployments) == 0 {
		return tg.RemoveProjectConfig(projectID)
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(tg.configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Build routers and services maps
	routers := make(map[string]traefikRouter)
	services := make(map[string]traefikService)

	for _, dep := range deployments {
		routerKey := fmt.Sprintf("deploy-%s", dep.ID)
		rule := fmt.Sprintf("Host(`%s.%s.%s`)", dep.Hash, projectName, tg.domain)

		routers[routerKey] = traefikRouter{
			Rule:    rule,
			Service: routerKey,
		}

		services[routerKey] = traefikService{
			LoadBalancer: loadBalancer{
				Servers: []server{
					{URL: "http://localhost:8081"},
				},
			},
		}
	}

	// Create config structure
	config := TraefikConfig{
		HTTP: httpConfig{
			Routers:  routers,
			Services: services,
		},
	}

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	configPath := fmt.Sprintf("%s/%s.yaml", tg.configDir, projectID)
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// RemoveProjectConfig deletes the configuration file for a project.
// Silently succeeds if the file does not exist.
func (tg *TraefikGateway) RemoveProjectConfig(projectID string) error {
	configPath := fmt.Sprintf("%s/%s.yaml", tg.configDir, projectID)
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
