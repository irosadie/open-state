package config

import (
	"os"
	"strconv"
	"strings"
)

// MCPConfig contains deployment-level security defaults. Per-connection
// timeout and resilience values are persisted on the project connection.
type MCPConfig struct {
	SecretStore       string
	Egress            MCPEgressConfig
	STDIOProfilesJSON string
}

type MCPEgressConfig struct {
	Mode          string
	Schemes       []string
	Ports         []int
	AllowedHosts  []string
	AllowedCIDRs  []string
	AllowLocalDev bool
	AllowPrivate  bool
}

func loadMCP() MCPConfig {
	return MCPConfig{
		SecretStore:       envDefault("MCP_SECRET_STORE", "composite"),
		STDIOProfilesJSON: os.Getenv("MCP_STDIO_PROFILES_JSON"),
		Egress: MCPEgressConfig{
			Mode:          envDefault("MCP_EGRESS_MODE", "production"),
			Schemes:       parseCSV(envDefault("MCP_EGRESS_SCHEMES", "https")),
			Ports:         parsePorts(os.Getenv("MCP_EGRESS_PORTS")),
			AllowedHosts:  parseCSV(os.Getenv("MCP_EGRESS_ALLOWED_HOSTS")),
			AllowedCIDRs:  parseCSV(os.Getenv("MCP_EGRESS_ALLOWED_CIDRS")),
			AllowLocalDev: envBool("MCP_EGRESS_ALLOW_LOCAL_DEV", false),
			AllowPrivate:  envBool("MCP_EGRESS_ALLOW_PRIVATE", false),
		},
	}
}

func parsePorts(raw string) []int {
	parts := parseCSV(raw)
	ports := make([]int, 0, len(parts))
	for _, part := range parts {
		value, err := strconv.Atoi(strings.TrimSpace(part))
		if err == nil && value > 0 && value <= 65535 {
			ports = append(ports, value)
		}
	}
	return ports
}
