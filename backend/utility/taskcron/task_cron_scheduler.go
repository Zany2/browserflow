package taskcron

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/Zany2/browserflow/backend/api/tasks/v1"
	taskcontroller "github.com/Zany2/browserflow/backend/internal/controller/tasks"
	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/Zany2/browserflow/backend/internal/model/do"
	"github.com/Zany2/browserflow/backend/utility/cronexpr"
	"github.com/Zany2/browserflow/backend/utility/tasklock"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcron"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
)

const (
	taskCronSyncName       = "task-cron-sync"
	taskCronSyncPattern    = "*/30 * * * * *"
	taskRecordSweepName    = "task-record-stale-sweep"
	taskRecordSweepPattern = "0 * * * * *"
	taskCronSyncBatchSize  = 500
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
		sweepStaleTaskRecords(ctx)

		if _, err := gcron.AddSingleton(ctx, taskCronSyncPattern, func(ctx context.Context) {
			syncCronTasks(ctx)
		}, taskCronSyncName); err != nil {
			g.Log().Line().Errorf(ctx, "start task cron sync failed: %+v", err)
			return
		}
		if _, err := gcron.AddSingleton(ctx, taskRecordSweepPattern, func(ctx context.Context) {
			sweepStaleTaskRecords(ctx)
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

// syncCronTasks syncs database cron tasks into gcron.
func syncCronTasks(ctx context.Context) {
	nextTasks, err := loadCronTaskMap(ctx)
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

func loadCronTaskMap(ctx context.Context) (map[string]string, error) {
	columns := dao.Tasks.Columns()
	result := make(map[string]string)
	lastID := int64(0)

	for {
		records, err := dao.Tasks.Ctx(ctx).
			Fields(columns.Id, columns.CronExpression).
			Where(columns.Enabled, true).
			Where(columns.CronExpression+" IS NOT NULL").
			Where(columns.CronExpression+" <> ?", "").
			WhereGT(columns.Id, lastID).
			OrderAsc(columns.Id).
			Limit(taskCronSyncBatchSize).
			All()
		if err != nil {
			return nil, err
		}
		if len(records) == 0 {
			return result, nil
		}

		for _, record := range records {
			lastID = gconv.Int64(record[columns.Id])
			taskID := gconv.String(record[columns.Id])
			cronExpression := cronexpr.Normalize(gconv.String(record[columns.CronExpression]))
			if taskID == "" || cronExpression == "" {
				continue
			}
			result[cronTaskName(taskID)] = cronExpression
		}
		if len(records) < taskCronSyncBatchSize {
			return result, nil
		}
	}
}

// executeCronTask executes one due cron task.
func executeCronTask(ctx context.Context, taskID string) {
	_, err := (&taskcontroller.ControllerV1{}).TaskExecute(ctx, &v1.TaskExecuteReq{
		ID:          taskID,
		TriggerType: "cron",
	})
	if err != nil {
		g.Log().Line().Warningf(ctx, "execute cron task failed: task_id=%s err=%+v", taskID, err)
	}
}

// sweepStaleTaskRecords marks stuck pending/queued/running task records failed.
func sweepStaleTaskRecords(ctx context.Context) {
	columns := dao.TaskRecords.Columns()
	cutoff := gtime.New(time.Now().Add(-tasklock.StaleAfter))
	records, err := dao.TaskRecords.Ctx(ctx).
		WhereIn(columns.Status, []string{"pending", "queued", "running"}).
		Where("COALESCE("+columns.StartedAt+", "+columns.CreatedAt+") < ?", cutoff).
		Limit(100).
		All()
	if err != nil {
		g.Log().Line().Warningf(ctx, "scan stale task records failed: %+v", err)
		return
	}

	for _, record := range records {
		recordID := gconv.Int64(record[columns.Id])
		clientIP := strings.TrimSpace(gconv.String(record[columns.ClientIp]))
		if recordID <= 0 {
			continue
		}

		commandID := "task-record-" + gconv.String(recordID)
		lockInfo, hasLock, lockErr := tasklock.Get(ctx, clientIP)
		if lockErr != nil {
			g.Log().Line().Warningf(ctx, "read client task lock failed: record_id=%d client_ip=%s err=%+v", recordID, clientIP, lockErr)
			continue
		}
		if hasLock && lockInfo.CommandID == commandID {
			continue
		}

		_, err = dao.TaskRecords.Ctx(ctx).
			WherePri(recordID).
			WhereIn(columns.Status, []string{"pending", "queued", "running"}).
			Data(do.TaskRecords{
				Status:       "failed",
				ErrorMessage: "client task execution timed out and was automatically ended",
				FinishedAt:   gtime.Now(),
			}).
			Update()
		if err != nil {
			g.Log().Line().Warningf(ctx, "mark stale task record failed: record_id=%d err=%+v", recordID, err)
			continue
		}
		if err = tasklock.Release(ctx, clientIP, commandID); err != nil {
			g.Log().Line().Warningf(ctx, "release stale task lock failed: record_id=%d client_ip=%s err=%+v", recordID, clientIP, err)
		}
	}
}

// cronTaskName builds cron job name.
func cronTaskName(taskID string) string {
	return "task-cron-" + taskID
}
