package main

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

const (
	bridgeScheme       = "mcp-marketplace"
	logFileName        = "mcp-local-bridge.log"
	registrySchemePath = `HKCU\Software\Classes\mcp-marketplace`
)

var slugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{1,63}$`)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(2)
	}

	arg := strings.TrimSpace(os.Args[1])
	if strings.HasPrefix(strings.ToLower(arg), bridgeScheme+"://") {
		if err := executeFromURI(arg); err != nil {
			fatal(err)
		}
		return
	}

	switch arg {
	case "install":
		if err := installProtocolHandler(); err != nil {
			fatal(err)
		}
	case "uninstall":
		if err := uninstallProtocolHandler(); err != nil {
			fatal(err)
		}
	case "open-uri":
		if len(os.Args) < 3 {
			fatal(fmt.Errorf("missing URI argument"))
		}
		if err := executeFromURI(strings.TrimSpace(os.Args[2])); err != nil {
			fatal(err)
		}
	default:
		printUsage()
		os.Exit(2)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  local-bridge install")
	fmt.Fprintln(os.Stderr, "  local-bridge uninstall")
	fmt.Fprintln(os.Stderr, "  local-bridge open-uri mcp-marketplace://install?...")
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}

func installProtocolHandler() error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("protocol auto-install is currently supported on Windows only")
	}

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable path: %w", err)
	}
	exePath, err = filepath.Abs(exePath)
	if err != nil {
		return fmt.Errorf("resolve absolute executable path: %w", err)
	}

	openCommand := fmt.Sprintf("\"%s\" open-uri \"%%1\"", exePath)

	steps := [][]string{
		{"add", registrySchemePath, "/ve", "/d", "URL:MCP Marketplace Protocol", "/f"},
		{"add", registrySchemePath, "/v", "URL Protocol", "/d", "", "/f"},
		{"add", registrySchemePath + `\shell\open\command`, "/ve", "/d", openCommand, "/f"},
	}
	for _, args := range steps {
		if err := runReg(args...); err != nil {
			return err
		}
	}
	fmt.Println("registered protocol handler for mcp-marketplace://")
	return nil
}

func uninstallProtocolHandler() error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("protocol auto-uninstall is currently supported on Windows only")
	}
	out, err := exec.Command("reg", "delete", registrySchemePath, "/f").CombinedOutput()
	if err != nil {
		text := strings.ToLower(strings.TrimSpace(string(out)))
		if strings.Contains(text, "unable to find") {
			fmt.Println("protocol handler was not registered")
			return nil
		}
		return fmt.Errorf("remove protocol handler: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	fmt.Println("unregistered protocol handler for mcp-marketplace://")
	return nil
}

func runReg(args ...string) error {
	out, err := exec.Command("reg", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("reg %s failed: %w (%s)", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return nil
}

func executeFromURI(rawURI string) error {
	u, err := url.Parse(rawURI)
	if err != nil {
		return fmt.Errorf("invalid URI: %w", err)
	}
	if strings.ToLower(u.Scheme) != bridgeScheme {
		return fmt.Errorf("unsupported URI scheme %q", u.Scheme)
	}

	action := strings.Trim(strings.ToLower(u.Host), "/")
	if action == "" {
		action = strings.Trim(strings.ToLower(u.Path), "/")
	}
	if action != "install" {
		return fmt.Errorf("unsupported action %q", action)
	}

	query := u.Query()
	client := strings.ToLower(strings.TrimSpace(query.Get("client")))
	slug := strings.TrimSpace(query.Get("slug"))
	resource := strings.TrimSpace(query.Get("resource"))

	if !slugPattern.MatchString(slug) {
		return fmt.Errorf("invalid slug %q", slug)
	}
	resourceURL, err := validateResource(resource)
	if err != nil {
		return err
	}

	executable, args, err := installCommand(client, slug, resourceURL.String())
	if err != nil {
		return err
	}
	return runCommandWithLog(executable, args)
}

func validateResource(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid resource URL: %w", err)
	}
	scheme := strings.ToLower(strings.TrimSpace(u.Scheme))
	if scheme != "http" && scheme != "https" {
		return nil, fmt.Errorf("resource URL must use http or https")
	}
	hostname := strings.ToLower(strings.TrimSpace(u.Hostname()))
	if hostname == "" {
		return nil, fmt.Errorf("resource URL host is required")
	}
	if hostname != "localhost" && hostname != "127.0.0.1" && scheme != "https" {
		return nil, fmt.Errorf("resource URL must use https for non-local hosts")
	}
	if !isAllowedHost(hostname) {
		return nil, fmt.Errorf("resource host %q is not allowed by MCP_LOCAL_BRIDGE_ALLOWED_HOSTS", hostname)
	}
	return u, nil
}

func isAllowedHost(host string) bool {
	allowedRaw := strings.TrimSpace(os.Getenv("MCP_LOCAL_BRIDGE_ALLOWED_HOSTS"))
	if allowedRaw == "" {
		return true
	}
	allowed := splitCSVLower(allowedRaw)
	for _, entry := range allowed {
		if entry == "" {
			continue
		}
		if strings.HasPrefix(entry, "*.") {
			suffix := strings.TrimPrefix(entry, "*.")
			if host == suffix || strings.HasSuffix(host, "."+suffix) {
				return true
			}
			continue
		}
		if host == entry {
			return true
		}
	}
	return false
}

func splitCSVLower(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		out = append(out, strings.ToLower(strings.TrimSpace(part)))
	}
	return out
}

func installCommand(client string, slug string, resource string) (string, []string, error) {
	switch client {
	case "codex":
		return "codex", []string{"mcp", "add", slug, "--url", resource}, nil
	case "claude":
		return "claude", []string{"mcp", "add", "--transport", "http", slug, resource}, nil
	case "cursor":
		return "cursor", []string{"mcp", "add", "--url", resource, "--name", slug}, nil
	default:
		return "", nil, fmt.Errorf("unsupported client %q", client)
	}
}

func runCommandWithLog(executable string, args []string) error {
	logPath := filepath.Join(os.TempDir(), logFileName)
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open log file %s: %w", logPath, err)
	}
	defer logFile.Close()

	stdoutWriter := io.MultiWriter(os.Stdout, logFile)
	stderrWriter := io.MultiWriter(os.Stderr, logFile)
	_, _ = fmt.Fprintf(stdoutWriter, "\n[%s] running: %s %s\n", time.Now().UTC().Format(time.RFC3339), executable, strings.Join(args, " "))

	cmd := exec.Command(executable, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = stdoutWriter
	cmd.Stderr = stderrWriter
	if err := cmd.Run(); err != nil {
		var execErr *exec.Error
		if errors.As(err, &execErr) && errors.Is(execErr.Err, exec.ErrNotFound) {
			return fmt.Errorf("%s executable not found in PATH (see %s)", executable, logPath)
		}
		return fmt.Errorf("install command failed: %w (see %s)", err, logPath)
	}

	_, _ = fmt.Fprintf(stdoutWriter, "[%s] install command completed successfully\n", time.Now().UTC().Format(time.RFC3339))
	return nil
}
