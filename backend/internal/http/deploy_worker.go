package http

import (
	"context"
	"time"

	"github.com/yourorg/mcp-marketplace/backend/internal/models"
)

func (a *App) startDeployWorker() {
	if a.n8n == nil || !a.n8n.configured() {
		return
	}

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		a.triggerDeployWorker()
		for {
			select {
			case <-a.deployTrigger:
				a.processDeployQueue(context.Background(), 20)
			case <-ticker.C:
				a.processDeployQueue(context.Background(), 20)
			}
		}
	}()
}

func (a *App) triggerDeployWorker() {
	if a.deployTrigger == nil {
		return
	}
	select {
	case a.deployTrigger <- struct{}{}:
	default:
	}
}

func (a *App) processDeployQueue(ctx context.Context, limit int) {
	if a.n8n == nil || !a.n8n.configured() {
		return
	}

	now := time.Now().UTC()
	tasks := a.store.ListDueDeployTasks(now, limit)
	for _, task := range tasks {
		a.processDeployTask(ctx, task)
	}
}

func (a *App) processDeployTask(ctx context.Context, task models.DeployTask) {
	if task.Status != models.DeployTaskStatusPending {
		return
	}
	now := time.Now().UTC()

	task.Status = models.DeployTaskStatusProcessing
	task.AttemptCount++
	task.UpdatedAt = now
	if !a.store.UpdateDeployTask(task) {
		return
	}

	server, ok := a.store.GetServerByID(task.ServerID)
	if !ok {
		task.Status = models.DeployTaskStatusFailed
		task.CompletedAt = now
		task.LastError = "server not found"
		task.UpdatedAt = now
		_ = a.store.UpdateDeployTask(task)
		return
	}

	result, err := a.n8n.deployWorkflow(ctx, server, task.PreferredWorkflowID)
	if err != nil {
		a.handleDeployTaskFailure(task, server, err)
		return
	}

	a.normalizeServerLifecycleForView(&server)
	server.DeploymentStatus = models.ServerDeploymentDeployed
	server.DeployedAt = now
	if task.RequestedBy != "" {
		server.DeployedBy = task.RequestedBy
	}
	if task.DeploymentTarget != "" {
		server.DeploymentTarget = task.DeploymentTarget
	}
	if result.WorkflowID != "" {
		server.N8nWorkflowID = result.WorkflowID
	}
	if result.WorkflowURL != "" {
		server.N8nWorkflowURL = result.WorkflowURL
	}
	if server.Status != models.ServerStatusPublished {
		server.Status = models.ServerStatusDraft
		server.PublishedAt = time.Time{}
	}
	server.UpdatedAt = now
	_ = a.store.UpdateServer(server)

	task.Status = models.DeployTaskStatusCompleted
	task.LastError = ""
	task.NextAttemptAt = time.Time{}
	task.CompletedAt = now
	task.UpdatedAt = now
	_ = a.store.UpdateDeployTask(task)

	a.store.AddAuditLog(models.AuditLog{
		TenantID:   server.TenantID,
		ActorID:    task.RequestedBy,
		Action:     "server.deploy.completed",
		TargetType: "server",
		TargetID:   server.ID,
		Outcome:    "success",
		Metadata: map[string]interface{}{
			"workflowId":  server.N8nWorkflowID,
			"workflowUrl": server.N8nWorkflowURL,
			"attempts":    task.AttemptCount,
		},
	})
}

func (a *App) handleDeployTaskFailure(task models.DeployTask, server models.Server, err error) {
	now := time.Now().UTC()
	task.LastError = err.Error()
	task.UpdatedAt = now

	if task.AttemptCount >= task.MaxAttempts {
		task.Status = models.DeployTaskStatusFailed
		task.CompletedAt = now
		server.DeploymentStatus = models.ServerDeploymentPending
		server.UpdatedAt = now
		_ = a.store.UpdateServer(server)
	} else {
		task.Status = models.DeployTaskStatusPending
		task.NextAttemptAt = now.Add(deployRetryDelay(task.AttemptCount))
		server.DeploymentStatus = models.ServerDeploymentQueued
		server.UpdatedAt = now
		_ = a.store.UpdateServer(server)
		a.triggerDeployWorker()
	}
	_ = a.store.UpdateDeployTask(task)

	a.store.AddAuditLog(models.AuditLog{
		TenantID:   server.TenantID,
		ActorID:    task.RequestedBy,
		Action:     "server.deploy.failed",
		TargetType: "server",
		TargetID:   server.ID,
		Outcome:    "failure",
		Metadata: map[string]interface{}{
			"error":       err.Error(),
			"attempts":    task.AttemptCount,
			"maxAttempts": task.MaxAttempts,
			"status":      task.Status,
		},
	})
}

func deployRetryDelay(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	delay := time.Duration(attempt*attempt) * 5 * time.Second
	if delay > 5*time.Minute {
		return 5 * time.Minute
	}
	return delay
}
