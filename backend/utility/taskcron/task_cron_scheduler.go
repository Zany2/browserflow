package taskcron

import (
	"context"
	"strings"
	"sync"

	"github.com/Zany2/browserflow/backend/api/tasks/v1"
	taskcontroller "github.com/Zany2/browserflow/backend/internal/controller/tasks"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcron"
)

const (
	taskCronSyncName       = "task-cron-sync"
	taskCronSyncPattern    = "*/30 * * * * *"
	taskRecordSweepName    = "task-record-stale-sweep"
	taskRecordSweepPattern = "0 * * * * *"
)

var cronScheduler = struct {
	once  sync.Once
	mutex sync.Mutex
	tasks map[string]string
}{
	tasks: make(map[string]string),
}

// StartCronScheduler starts task cron scheduler.
func StartCronScheduler(ctx context.Context) {
	cronScheduler.once.Do(func() {
		syncCronTasks(ctx)
		taskcontroller.SweepStaleTaskRecords(ctx)

		if _, err := gcron.AddSingleton(ctx, taskCronSyncPattern, func(ctx context.Context) {
			syncCronTasks(ctx)
		}, taskCronSyncName); err != nil {
			g.Log().Line().Errorf(ctx, "start task cron sync failed: %+v", err)
			return
		}
		if _, err := gcron.AddSingleton(ctx, taskRecordSweepPattern, func(ctx context.Context) {
			taskcontroller.SweepStaleTaskRecords(ctx)
		}, taskRecordSweepName); err != nil {
			g.Log().Line().Errorf(ctx, "start stale task record sweep failed: %+v", err)
			return
		}

		g.Log().Line().Info(ctx, "task cron scheduler started")
	})
}

// StopCronScheduler stops task cron scheduler.
func StopCronScheduler() {
	cronScheduler.mutex.Lock()
	defer cronScheduler.mutex.Unlock()

	for jobName := range cronScheduler.tasks {
		gcron.Remove(jobName)
	}
	gcron.Remove(taskCronSyncName)
	gcron.Remove(taskRecordSweepName)
	cronScheduler.tasks = make(map[string]string)
}

func syncCronTasks(ctx context.Context) {
	nextTasks, err := taskcontroller.LoadCronTaskMap(ctx)
	if err != nil {
		g.Log().Line().Errorf(ctx, "query cron tasks failed: %+v", err)
		return
	}

	cronScheduler.mutex.Lock()
	defer cronScheduler.mutex.Unlock()

	for jobName, oldExpression := range cronScheduler.tasks {
		newExpression, exists := nextTasks[jobName]
		if exists && oldExpression == newExpression {
			delete(nextTasks, jobName)
			continue
		}

		gcron.Remove(jobName)
		delete(cronScheduler.tasks, jobName)
	}

	for jobName, cronExpression := range nextTasks {
		taskID := strings.TrimPrefix(jobName, "task-cron-")
		jobTaskID := taskID
		if _, err := gcron.AddSingleton(ctx, cronExpression, func(ctx context.Context) {
			executeCronTask(ctx, jobTaskID)
		}, jobName); err != nil {
			g.Log().Line().Warningf(ctx, "register cron task failed: task_id=%s cron=%s err=%+v", taskID, cronExpression, err)
			continue
		}
		cronScheduler.tasks[jobName] = cronExpression
		g.Log().Line().Infof(ctx, "registered cron task: task_id=%s cron=%s", taskID, cronExpression)
	}
}

func executeCronTask(ctx context.Context, taskID string) {
	_, err := (&taskcontroller.ControllerV1{}).TaskExecute(ctx, &v1.TaskExecuteReq{
		ID:          taskID,
		TriggerType: "cron",
	})
	if err != nil {
		g.Log().Line().Warningf(ctx, "execute cron task failed: task_id=%s err=%+v", taskID, err)
	}
}
