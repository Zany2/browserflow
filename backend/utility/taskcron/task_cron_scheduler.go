package taskcron

import (
	"context"
	"strings"
	"sync"

	"github.com/Zany2/browserflow/backend/api/tasks/v1"
	taskcontroller "github.com/Zany2/browserflow/backend/internal/controller/tasks"
	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcron"
	"github.com/gogf/gf/v2/util/gconv"
)

const (
	taskCronSyncName    = "task-cron-sync" // taskCronSyncName scheduler sync job name 调度同步任务名称
	taskCronSyncPattern = "*/30 * * * * *" // taskCronSyncPattern scheduler sync interval 调度同步间隔
)

var cronScheduler = struct {
	once  sync.Once         // once starts scheduler once 确保调度器只启动一次
	mutex sync.Mutex        // mutex protects task registry 保护任务注册表
	tasks map[string]string // tasks registered job expressions 已注册任务表达式
}{
	tasks: make(map[string]string),
}

// StartCronScheduler starts task cron scheduler 启动任务定时调度器
func StartCronScheduler(ctx context.Context) {
	cronScheduler.once.Do(func() {
		syncCronTasks(ctx)

		if _, err := gcron.AddSingleton(ctx, taskCronSyncPattern, func(ctx context.Context) {
			syncCronTasks(ctx)
		}, taskCronSyncName); err != nil {
			g.Log().Line().Errorf(ctx, "启动任务定时调度同步器失败: %+v", err)
			return
		}

		g.Log().Line().Info(ctx, "任务定时调度器已启动")
	})
}

// StopCronScheduler stops task cron scheduler 停止任务定时调度器
func StopCronScheduler() {
	cronScheduler.mutex.Lock()
	defer cronScheduler.mutex.Unlock()

	for jobName := range cronScheduler.tasks {
		gcron.Remove(jobName)
	}
	gcron.Remove(taskCronSyncName)
	cronScheduler.tasks = make(map[string]string)
}

// syncCronTasks syncs database cron tasks into gcron 查询数据库并同步定时任务
func syncCronTasks(ctx context.Context) {
	columns := dao.Tasks.Columns()
	records, err := dao.Tasks.Ctx(ctx).
		Where(columns.Enabled, true).
		Where(columns.CronExpression+" IS NOT NULL").
		Where(columns.CronExpression+" <> ?", "").
		All()
	if err != nil {
		g.Log().Line().Errorf(ctx, "查询定时任务失败: %+v", err)
		return
	}

	nextTasks := make(map[string]string, len(records))
	for _, record := range records {
		taskID := gconv.String(record[columns.Id])
		cronExpression := normalizeCronExpression(gconv.String(record[columns.CronExpression]))
		if taskID == "" || cronExpression == "" {
			continue
		}
		nextTasks[cronTaskName(taskID)] = cronExpression
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
			g.Log().Line().Warningf(ctx, "注册任务定时调度失败: task_id=%s cron=%s err=%+v", taskID, cronExpression, err)
			continue
		}
		cronScheduler.tasks[jobName] = cronExpression
	}
}

// executeCronTask executes one due cron task 执行命中的定时任务
func executeCronTask(ctx context.Context, taskID string) {
	_, err := (&taskcontroller.ControllerV1{}).TaskExecute(ctx, &v1.TaskExecuteReq{
		ID:          taskID,
		TriggerType: "cron",
	})
	if err != nil {
		g.Log().Line().Warningf(ctx, "执行定时任务失败: task_id=%s err=%+v", taskID, err)
	}
}

// cronTaskName builds cron job name 构建定时任务名称
func cronTaskName(taskID string) string {
	return "task-cron-" + taskID
}

// normalizeCronExpression adapts five-field cron to gcron 适配五段式 Cron 表达式
func normalizeCronExpression(cronExpression string) string {
	cronExpression = strings.TrimSpace(cronExpression)
	parts := strings.Fields(cronExpression)
	if len(parts) == 5 {
		return "0 " + cronExpression
	}
	return cronExpression
}
