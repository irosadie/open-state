package capability

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// EgressPolicy is shared by MCP HTTP transports and OAuth HTTP calls. The
// default production posture is HTTPS-only, port 443, no private addresses,
// and no redirects outside the same policy.
type EgressPolicy struct {
	Mode             string
	AllowedSchemes   map[string]struct{}
	AllowedPorts     map[int]struct{}
	AllowedHosts     []string
	AllowedCIDRs     []*net.IPNet
	AllowLocalDev    bool
	AllowPrivate     bool
	MaxResponseBytes int64
	ResponseTimeout  time.Duration
}

func NewEgressPolicy(mode string, schemes []string, ports []int, hosts []string, cidrs []string, allowLocalDev, allowPrivate bool) (*EgressPolicy, error) {
	p := &EgressPolicy{
		Mode:             strings.ToLower(strings.TrimSpace(mode)),
		AllowedSchemes:   make(map[string]struct{}),
		AllowedPorts:     make(map[int]struct{}),
		AllowedHosts:     normalizeHosts(hosts),
		AllowLocalDev:    allowLocalDev,
		AllowPrivate:     allowPrivate,
		MaxResponseBytes: 8 * 1024 * 1024,
		ResponseTimeout:  15 * time.Second,
	}
	if p.Mode == "" {
		p.Mode = "production"
	}
	for _, scheme := range schemes {
		scheme = strings.ToLower(strings.TrimSpace(scheme))
		if scheme != "" {
			p.AllowedSchemes[scheme] = struct{}{}
		}
	}
	if len(p.AllowedSchemes) == 0 {
		p.AllowedSchemes["https"] = struct{}{}
	}
	for _, port := range ports {
		if port > 0 && port <= 65535 {
			p.AllowedPorts[port] = struct{}{}
		}
	}
	if len(p.AllowedPorts) == 0 {
		p.AllowedPorts[443] = struct{}{}
	}
	for _, raw := range cidrs {
		_, network, err := net.ParseCIDR(strings.TrimSpace(raw))
		if err != nil {
			return nil, fmt.Errorf("invalid MCP egress CIDR")
		}
		p.AllowedCIDRs = append(p.AllowedCIDRs, network)
	}
	return p, nil
}

func (p *EgressPolicy) ValidateURL(ctx context.Context, raw string) error {
	if p == nil {
		return nil
	}
	u, err := url.ParseRequestURI(strings.TrimSpace(raw))
	if err != nil || u.Scheme == "" || u.Hostname() == "" || u.User != nil {
		return errors.New("MCP endpoint is not allowed by egress policy")
	}
	if _, ok := p.AllowedSchemes[strings.ToLower(u.Scheme)]; !ok {
		return errors.New("MCP endpoint scheme is not allowed")
	}
	port := 443
	if u.Port() != "" {
		port, err = strconv.Atoi(u.Port())
		if err != nil {
			return errors.New("MCP endpoint port is not allowed")
		}
	} else if strings.EqualFold(u.Scheme, "http") {
		port = 80
	}
	if _, ok := p.AllowedPorts[port]; !ok {
		return errors.New("MCP endpoint port is not allowed")
	}
	host := strings.ToLower(strings.TrimSuffix(u.Hostname(), "."))
	local := isLocalHost(host)
	if local && !(p.Mode == "development" && p.AllowLocalDev) {
		return errors.New("local MCP endpoints are disabled")
	}
	if len(p.AllowedHosts) > 0 && !matchesHostAllowlist(host, p.AllowedHosts) {
		return errors.New("MCP endpoint host is not allowlisted")
	}
	ips, err := resolveHost(ctx, host)
	if err != nil {
		return errors.New("MCP endpoint DNS resolution failed")
	}
	for _, ip := range ips {
		if !p.allowedIP(host, ip) {
			return errors.New("MCP endpoint resolves to a prohibited network")
		}
	}
	return nil
}

func (p *EgressPolicy) allowedIP(host string, ip net.IP) bool {
	if p.Mode == "development" && p.AllowLocalDev && isLocalHost(host) {
		return true
	}
	// Loopback, link-local, multicast, and unspecified addresses are never
	// reachable through the generic private-network switch. Local development
	// must be opted into explicitly and is restricted to a loopback hostname.
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() || ip.IsMulticast() {
		return false
	}
	if p.AllowPrivate {
		return true
	}
	for _, network := range p.AllowedCIDRs {
		if network.Contains(ip) {
			return true
		}
	}
	return !isProhibitedIP(ip)
}

// HTTPClient provides connect-time DNS revalidation and redirect enforcement.
// Dialing a validated IP instead of the hostname closes the common DNS
// rebinding gap between validation and socket creation.
func (p *EgressPolicy) HTTPClient(ctx context.Context) *http.Client {
	timeout := p.ResponseTimeout
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	transport := &http.Transport{
		Proxy:                 nil,
		ForceAttemptHTTP2:     true,
		DialContext:           p.dialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: timeout,
		MaxIdleConns:          32,
		MaxIdleConnsPerHost:   8,
		IdleConnTimeout:       30 * time.Second,
	}
	return &http.Client{
		Transport: &boundedResponseTransport{base: transport, maxBytes: p.MaxResponseBytes},
		Timeout:   timeout,
		CheckRedirect: func(req *http.Request, _ []*http.Request) error {
			if err := p.ValidateURL(req.Context(), req.URL.String()); err != nil {
				return errors.New("MCP redirect blocked by egress policy")
			}
			return nil
		},
	}
}

type boundedResponseTransport struct {
	base     http.RoundTripper
	maxBytes int64
}

func (t *boundedResponseTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	response, err := t.base.RoundTrip(req)
	if err != nil || response == nil || response.Body == nil || t.maxBytes <= 0 {
		return response, err
	}
	response.Body = &boundedResponseBody{ReadCloser: response.Body, remaining: t.maxBytes}
	return response, nil
}

type boundedResponseBody struct {
	io.ReadCloser
	remaining int64
}

func (b *boundedResponseBody) Read(p []byte) (int, error) {
	if b.remaining <= 0 {
		var probe [1]byte
		n, err := b.ReadCloser.Read(probe[:])
		if n > 0 {
			return n, errors.New("MCP provider response exceeded the configured size limit")
		}
		if err == io.EOF {
			return 0, io.EOF
		}
		if err != nil {
			return 0, err
		}
		return 0, errors.New("MCP provider response exceeded the configured size limit")
	}
	if int64(len(p)) > b.remaining+1 {
		p = p[:b.remaining+1]
	}
	n, err := b.ReadCloser.Read(p)
	b.remaining -= int64(n)
	if b.remaining < 0 {
		return n, errors.New("MCP provider response exceeded the configured size limit")
	}
	return n, err
}

func (p *EgressPolicy) dialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, errors.New("MCP dial address is invalid")
	}
	ips, err := resolveHost(ctx, host)
	if err != nil {
		return nil, errors.New("MCP endpoint DNS resolution failed")
	}
	for _, ip := range ips {
		if !p.allowedIP(strings.ToLower(strings.TrimSuffix(host, ".")), ip) {
			continue
		}
		dialer := &net.Dialer{Timeout: 10 * time.Second}
		conn, dialErr := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
		if dialErr == nil {
			return conn, nil
		}
	}
	return nil, errors.New("MCP endpoint connection blocked by egress policy")
}

func resolveHost(ctx context.Context, host string) ([]net.IP, error) {
	if ip := net.ParseIP(host); ip != nil {
		return []net.IP{ip}, nil
	}
	addresses, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
	if err != nil || len(addresses) == 0 {
		return nil, errors.New("DNS lookup failed")
	}
	return addresses, nil
}

func isProhibitedIP(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() || ip.IsMulticast() || isCarrierGradeNAT(ip)
}

func isCarrierGradeNAT(ip net.IP) bool {
	if v4 := ip.To4(); v4 != nil {
		return v4[0] == 100 && v4[1] >= 64 && v4[1] <= 127
	}
	return false
}

func isLocalHost(host string) bool {
	if host == "localhost" || host == "localhost.localdomain" {
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback()
	}
	return false
}

func normalizeHosts(hosts []string) []string {
	out := make([]string, 0, len(hosts))
	for _, host := range hosts {
		host = strings.ToLower(strings.TrimSpace(strings.TrimSuffix(host, ".")))
		if host != "" {
			out = append(out, host)
		}
	}
	return out
}

func matchesHostAllowlist(host string, allowlist []string) bool {
	for _, allowed := range allowlist {
		if allowed == host || (strings.HasPrefix(allowed, "*.") && strings.HasSuffix(host, strings.TrimPrefix(allowed, "*"))) {
			return true
		}
	}
	return false
}
