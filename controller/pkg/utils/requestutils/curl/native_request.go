package curl

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
)

// ExecuteRequest accepts a set of Option and executes a native Go HTTP request
// If multiple Option modify the same parameter, the last defined one will win
//
// Example:
//
//	resp, err := ExecuteRequest(WithMethod("GET"), WithMethod("POST"))
//	will executeNative a POST request
//
// A notable exception is the WithHeader option, which accumulates headers
func ExecuteRequest(options ...Option) (*http.Response, error) {
	config := &requestConfig{
		host:    "127.0.0.1",
		port:    80,
		headers: make(map[string][]string),
		scheme:  "http",
	}

	for _, opt := range options {
		opt(config)
	}

	return config.executeNative()
}

func (c *requestConfig) executeNative() (*http.Response, error) {
	fullURL := c.buildURL()
	if err := validateNativeRequestTarget(fullURL); err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: c.timeout,
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	if c.tlsConfig != nil {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.TLSClientConfig = c.tlsConfig
		client.Transport = transport
	}

	method := c.method

	var bodyReader io.Reader
	if c.body != "" {
		bodyReader = bytes.NewBufferString(c.body)
		if method == "" {
			method = "POST"
		}
	}

	if method == "" {
		method = "GET"
	}

	req, err := http.NewRequest(method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for key, values := range c.headers {
		for _, value := range values {
			if strings.EqualFold(key, "Host") {
				req.Host = value
			} else {
				req.Header.Add(key, value)
			}
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *requestConfig) buildURL() string {
	path := c.path
	if path != "" && !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	baseURL := fmt.Sprintf("%s://%s:%d%s", c.scheme, c.host, c.port, path)
	return baseURL
}

func validateNativeRequestTarget(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid request URL: %w", err)
	}

	switch parsed.Scheme {
	case "http", "https":
	default:
		return fmt.Errorf("unsupported request scheme %q", parsed.Scheme)
	}

	host := strings.TrimSuffix(strings.ToLower(parsed.Hostname()), ".")
	if host == "" {
		return fmt.Errorf("request host is required")
	}

	if host == "metadata" || strings.HasSuffix(host, ".metadata") || host == "metadata.google.internal" {
		return fmt.Errorf("disallowed metadata service target %q", host)
	}

	addr, err := netip.ParseAddr(host)
	if err != nil {
		return nil
	}

	if isMetadataServiceAddress(addr) {
		return fmt.Errorf("disallowed metadata service target %q", host)
	}

	return nil
}

func isMetadataServiceAddress(addr netip.Addr) bool {
	if addr.Is4() {
		metadataPrefix := netip.MustParsePrefix("169.254.169.254/32")
		return metadataPrefix.Contains(addr)
	}

	return addr == netip.MustParseAddr("fd00:ec2::254")
}
