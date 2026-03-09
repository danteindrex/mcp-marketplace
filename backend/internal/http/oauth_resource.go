package http

import (
	"net/url"
	"path"
	"strings"
)

func canonicalHubResource(baseURL, resource string) (canonical, tenantID, userID string, ok bool) {
	base, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil || base.Scheme == "" || base.Host == "" {
		return "", "", "", false
	}
	requested, err := url.Parse(strings.TrimSpace(resource))
	if err != nil || requested.Scheme == "" || requested.Host == "" {
		return "", "", "", false
	}
	if !strings.EqualFold(base.Scheme, requested.Scheme) || !strings.EqualFold(base.Host, requested.Host) {
		return "", "", "", false
	}
	basePath := strings.TrimSuffix(path.Clean("/"+strings.TrimSpace(base.Path)), "/")
	if basePath == "/." {
		basePath = ""
	}
	hubPath := strings.Trim(path.Clean("/"+strings.TrimSpace(requested.Path)), "/")
	parts := strings.Split(hubPath, "/")
	baseParts := []string{}
	if strings.Trim(basePath, "/") != "" {
		baseParts = strings.Split(strings.Trim(basePath, "/"), "/")
	}
	if len(parts) != len(baseParts)+4 {
		return "", "", "", false
	}
	for i := range baseParts {
		if parts[i] != baseParts[i] {
			return "", "", "", false
		}
	}
	if parts[len(baseParts)] != "mcp" || parts[len(baseParts)+1] != "hub" {
		return "", "", "", false
	}
	tenantID = strings.TrimSpace(parts[len(baseParts)+2])
	userID = strings.TrimSpace(parts[len(baseParts)+3])
	if tenantID == "" || userID == "" {
		return "", "", "", false
	}
	return strings.TrimRight(base.String(), "/") + "/mcp/hub/" + tenantID + "/" + userID, tenantID, userID, true
}

func (a *App) hubResourceForSubject(tenantID, userID string) string {
	return strings.TrimRight(a.cfg.BaseURL, "/") + "/mcp/hub/" + tenantID + "/" + userID
}
