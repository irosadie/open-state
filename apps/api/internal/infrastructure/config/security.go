package config

import "os"

// SecurityConfig controls HTTP security headers and CORS (PRD §139, §74).
type SecurityConfig struct {
	CSP        string
	EnableHSTS bool
	HSTSMaxAge int
	// AllowedOrigins is the CORS allow-list (comma-separated in env).
	AllowedOrigins []string
}

// loadSecurity reads security env vars.
func loadSecurity() SecurityConfig {
	return SecurityConfig{
		CSP:            os.Getenv("SECURITY_CSP"),
		EnableHSTS:     envBool("SECURITY_HSTS_ENABLED", false),
		HSTSMaxAge:     envInt("SECURITY_HSTS_MAX_AGE", 31536000),
		AllowedOrigins: parseCSV(os.Getenv("CORS_ALLOWED_ORIGINS")),
	}
}

func parseCSV(s string) []string {
	if s == "" {
		return nil
	}
	out := []string{}
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' {
			part := trimSpace(s[start:i])
			if part != "" {
				out = append(out, part)
			}
			start = i + 1
		}
	}
	return out
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}
