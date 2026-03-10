package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/yourorg/mcp-marketplace/backend/internal/config"
	"github.com/yourorg/mcp-marketplace/backend/internal/models"
)

type dockerRuntime struct {
	socketPath string
	network    string
	publicHost string
	client     *http.Client
}

func newDockerRuntime(cfg config.Config) *dockerRuntime {
	socketPath := strings.TrimSpace(cfg.DockerRuntimeSocket)
	if socketPath == "" {
		return nil
	}
	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", socketPath)
		},
	}
	return &dockerRuntime{
		socketPath: socketPath,
		network:    strings.TrimSpace(cfg.DockerRuntimeNetwork),
		publicHost: strings.TrimSpace(cfg.DockerRuntimePublicHost),
		client: &http.Client{
			Transport: transport,
			Timeout:   60 * time.Second,
		},
	}
}

func (d *dockerRuntime) configured() bool {
	return d != nil && strings.TrimSpace(d.socketPath) != ""
}

func (a *App) currentDockerRuntime() *dockerRuntime {
	if a == nil {
		return nil
	}
	return a.dockerRuntime
}

func (d *dockerRuntime) deployContainer(ctx context.Context, server models.Server) (n8nDeployResult, error) {
	if !d.configured() {
		return n8nDeployResult{}, fmt.Errorf("docker runtime is not configured")
	}
	containerPort := server.ContainerPort
	if containerPort <= 0 {
		containerPort = 3000
	}
	name := "mcp-server-" + sanitizeN8NPath(server.ID)
	if err := d.removeContainer(ctx, name); err != nil {
		return n8nDeployResult{}, err
	}
	if !d.imageExists(ctx, server.DockerImage) {
		if err := d.pullImage(ctx, server.DockerImage); err != nil {
			return n8nDeployResult{}, err
		}
	}

	containerPortKey := strconv.Itoa(containerPort) + "/tcp"
	body := map[string]interface{}{
		"Image": server.DockerImage,
		"Env": []string{
			"PORT=" + strconv.Itoa(containerPort),
		},
		"ExposedPorts": map[string]interface{}{
			containerPortKey: map[string]interface{}{},
		},
		"HostConfig": map[string]interface{}{
			"PortBindings": map[string]interface{}{
				containerPortKey: []map[string]string{{"HostIp": "0.0.0.0", "HostPort": ""}},
			},
			"NetworkMode": d.network,
		},
	}
	created := map[string]interface{}{}
	if err := d.requestJSON(ctx, http.MethodPost, "/containers/create?name="+url.QueryEscape(name), body, &created); err != nil {
		return n8nDeployResult{}, err
	}
	containerID := strings.TrimSpace(stringFromAny(created["Id"]))
	if containerID == "" {
		return n8nDeployResult{}, fmt.Errorf("docker create did not return a container id")
	}
	if err := d.requestJSON(ctx, http.MethodPost, "/containers/"+url.PathEscape(containerID)+"/start", nil, nil); err != nil {
		return n8nDeployResult{}, err
	}
	inspected := map[string]interface{}{}
	if err := d.requestJSON(ctx, http.MethodGet, "/containers/"+url.PathEscape(containerID)+"/json", nil, &inspected); err != nil {
		return n8nDeployResult{}, err
	}
	hostPort := dockerHostPort(inspected, containerPortKey)
	if hostPort == "" {
		return n8nDeployResult{}, fmt.Errorf("docker runtime did not publish %s", containerPortKey)
	}
	publicHost := d.publicHost
	if publicHost == "" {
		publicHost = "localhost"
	}
	runtimeURL := "http://" + publicHost + ":" + hostPort
	if err := d.waitForRuntime(ctx, name, containerPort); err != nil {
		return n8nDeployResult{}, err
	}
	return n8nDeployResult{
		RuntimeContainerID: containerID,
		RuntimeURL:         runtimeURL,
	}, nil
}

func (d *dockerRuntime) imageExists(ctx context.Context, image string) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://docker/images/"+url.PathEscape(strings.TrimSpace(image))+"/json", nil)
	if err != nil {
		return false
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

func (d *dockerRuntime) waitForRuntime(ctx context.Context, containerName string, containerPort int) error {
	client := &http.Client{Timeout: 3 * time.Second}
	target := "http://" + containerName + ":" + strconv.Itoa(containerPort)
	var lastErr error
	for attempt := 0; attempt < 20; attempt++ {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
		resp, err := client.Do(req)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode > 0 {
				return nil
			}
		} else {
			lastErr = err
		}
		time.Sleep(1500 * time.Millisecond)
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("container runtime did not become reachable")
	}
	return lastErr
}

func dockerHostPort(inspected map[string]interface{}, key string) string {
	networkSettings, _ := inspected["NetworkSettings"].(map[string]interface{})
	ports, _ := networkSettings["Ports"].(map[string]interface{})
	rawList, ok := ports[key].([]interface{})
	if !ok || len(rawList) == 0 {
		return ""
	}
	binding, _ := rawList[0].(map[string]interface{})
	return strings.TrimSpace(stringFromAny(binding["HostPort"]))
}

func (d *dockerRuntime) pullImage(ctx context.Context, image string) error {
	endpoint := "/images/create?fromImage=" + url.QueryEscape(strings.TrimSpace(image))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://docker"+endpoint, nil)
	if err != nil {
		return err
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("docker image pull failed: %s", strings.TrimSpace(string(raw)))
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}

func (d *dockerRuntime) removeContainer(ctx context.Context, name string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, "http://docker/containers/"+url.PathEscape(name)+"?force=1", nil)
	if err != nil {
		return err
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("docker container cleanup failed: %s", strings.TrimSpace(string(raw)))
	}
	return nil
}

func (d *dockerRuntime) requestJSON(ctx context.Context, method, endpoint string, payload interface{}, out interface{}) error {
	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, "http://docker"+path.Clean("/"+endpoint), body)
	if err != nil {
		return err
	}
	if strings.Contains(endpoint, "?") {
		req.URL.RawQuery = strings.SplitN(endpoint, "?", 2)[1]
		req.URL.Path = path.Clean("/" + strings.SplitN(endpoint, "?", 2)[0])
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("docker runtime status %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	if out != nil && len(raw) > 0 {
		return json.Unmarshal(raw, out)
	}
	return nil
}
