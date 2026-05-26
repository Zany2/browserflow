package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/Zany2/browserflow/backend/internal/model/do"
	websockets "github.com/Zany2/browserflow/backend/utility/websocket"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
)

const (
	taskRecordTablePreviewLimit  = 20
	taskRecordResultFilesBaseDir = "data/task-record-files"
)

type persistence struct{}

func init() {
	websockets.RegisterPersistence(persistence{})
}

func (p persistence) CanRecoverTask(ctx context.Context, recordID int64) (bool, error) {
	columns := dao.TaskRecords.Columns()
	record, err := dao.TaskRecords.Ctx(ctx).
		Fields(columns.Id, columns.Status).
		WherePri(recordID).
		WhereIn(columns.Status, []string{"pending", "queued", "running"}).
		One()
	if err != nil {
		return false, err
	}
	return !record.IsEmpty(), nil
}

func (p persistence) SaveClientRegister(ctx context.Context, data websockets.NodeRegisterData) error {
	if strings.TrimSpace(data.ClientIP) == "" {
		return nil
	}

	columns := dao.Clients.Columns()
	now := gtime.Now()
	capabilitiesJSON := "{}"
	if len(data.Capabilities) > 0 {
		if body, err := json.Marshal(data.Capabilities); err == nil {
			capabilitiesJSON = string(body)
		}
	}

	saveData := do.Clients{
		ClientIp:         strings.TrimSpace(data.ClientIP),
		NodeId:           strings.TrimSpace(data.NodeID),
		NodeIndex:        data.NodeIndex,
		WorkerVersion:    strings.TrimSpace(data.WorkerVersion),
		ProfileDir:       strings.TrimSpace(data.ProfileDir),
		ExtensionDir:     strings.TrimSpace(data.ExtensionDir),
		UserAgent:        strings.TrimSpace(data.UserAgent),
		Status:           "online",
		BusyStatus:       "idle",
		PluginStatus:     strings.TrimSpace(data.PluginStatus),
		AutomaVersion:    strings.TrimSpace(data.AutomaVersion),
		BrowserName:      strings.TrimSpace(data.BrowserName),
		BrowserVersion:   strings.TrimSpace(data.BrowserVersion),
		OsName:           strings.TrimSpace(data.OsName),
		OsVersion:        strings.TrimSpace(data.OsVersion),
		Hostname:         strings.TrimSpace(data.Hostname),
		CapabilitiesJson: capabilitiesJSON,
		LastSeenAt:       now,
		ConnectedAt:      now,
	}

	clientModel := dao.Clients.Ctx(ctx)
	if strings.TrimSpace(data.NodeID) != "" {
		clientModel = clientModel.Where(columns.ClientIp, data.ClientIP).Where(columns.NodeId, data.NodeID)
	} else {
		clientModel = clientModel.Where(columns.ClientIp, data.ClientIP)
	}
	record, err := clientModel.One()
	if err != nil {
		return err
	}
	if record.IsEmpty() {
		saveData.FirstSeenAt = now
		_, err = dao.Clients.Ctx(ctx).Data(saveData).Insert()
		return err
	}

	updateModel := dao.Clients.Ctx(ctx)
	if strings.TrimSpace(data.NodeID) != "" {
		updateModel = updateModel.Where(columns.ClientIp, data.ClientIP).Where(columns.NodeId, data.NodeID)
	} else {
		updateModel = updateModel.Where(columns.ClientIp, data.ClientIP)
	}
	_, err = updateModel.Data(saveData).Update()
	return err
}

func (p persistence) UpdateClientLastSeen(ctx context.Context, data websockets.NodeHeartbeatData) error {
	clientIP := strings.TrimSpace(data.ClientIP)
	if clientIP == "" {
		return nil
	}

	columns := dao.Clients.Columns()
	updateData := do.Clients{
		ClientIp:   clientIP,
		NodeId:     strings.TrimSpace(data.NodeID),
		Status:     "online",
		LastSeenAt: gtime.Now(),
	}
	if data.UpdatePluginInfo {
		updateData.PluginStatus = strings.TrimSpace(data.PluginStatus)
		updateData.AutomaVersion = strings.TrimSpace(data.AutomaVersion)
	}
	if commandID := strings.TrimSpace(data.CommandID); commandID != "" {
		updateData.BusyStatus = "busy"
		updateData.CurrentExecutionId = commandID
		updateData.CurrentTaskRecordId = data.RecordID
		updateData.LastLockRenewedAt = gtime.Now()
	}

	updateModel := dao.Clients.Ctx(ctx)
	if strings.TrimSpace(data.NodeID) != "" {
		updateModel = updateModel.Where(columns.ClientIp, clientIP).Where(columns.NodeId, data.NodeID)
	} else {
		updateModel = updateModel.Where(columns.ClientIp, clientIP)
	}
	_, err := updateModel.Data(updateData).Update()
	return err
}

func (p persistence) MarkClientOffline(ctx context.Context, clientIP string, nodeID string) error {
	clientIP = strings.TrimSpace(clientIP)
	nodeID = strings.TrimSpace(nodeID)
	if clientIP == "" && nodeID == "" {
		return nil
	}

	columns := dao.Clients.Columns()
	updateModel := dao.Clients.Ctx(ctx)
	if nodeID != "" {
		updateModel = updateModel.Where(columns.ClientIp, clientIP).Where(columns.NodeId, nodeID)
	} else {
		updateModel = updateModel.Where(columns.ClientIp, clientIP)
	}
	_, err := updateModel.
		Data(do.Clients{
			Status:              "offline",
			BusyStatus:          "idle",
			CurrentExecutionId:  "",
			CurrentTaskRecordId: 0,
			DisconnectedAt:      gtime.Now(),
		}).
		Update()
	return err
}

func (p persistence) UpdateTaskRecordResult(ctx context.Context, data websockets.TaskRecordResultData) error {
	resultJSON := "{}"
	if len(data.Result) > 0 {
		resultJSON = string(data.Result)
	}
	resultJSON = saveTaskRecordResultFiles(ctx, data.RecordID, data.ClientIP, resultJSON)
	status := websockets.ResolveTaskRecordResultStatus(data.Success, data.Result)
	errorMessage := strings.TrimSpace(data.ErrorText)
	if errorMessage == "" && status == "failed" {
		errorMessage = websockets.ResolveTaskRecordResultMessage(data.Result)
		if errorMessage == "" && websockets.IsLostTaskResultStatus(data.Result) {
			errorMessage = "client has no local execution state for this task"
		}
	}

	updateData := do.TaskRecords{
		Status:       status,
		ResultJson:   resultJSON,
		ErrorMessage: errorMessage,
	}
	if automaExecutionID := websockets.ResolveAutomaExecutionID(data.Result); automaExecutionID != "" {
		updateData.AutomaExecutionId = automaExecutionID
	}
	if status == "running" {
		updateData.StartedAt = gtime.Now()
	}
	if status == "success" || status == "failed" {
		updateData.FinishedAt = gtime.Now()
	}

	columns := dao.TaskRecords.Columns()
	_, err := dao.TaskRecords.Ctx(ctx).
		WherePri(data.RecordID).
		WhereIn(columns.Status, []string{"pending", "queued", "running"}).
		Data(updateData).
		Update()
	if err != nil {
		return err
	}
	if status == "success" || status == "failed" {
		return p.MarkNodeIdle(ctx, data.ClientIP, data.NodeID, data.CommandID, data.RecordID)
	}
	return err
}

func (p persistence) FailClientRunningTask(ctx context.Context, recordID int64, message string) error {
	columns := dao.TaskRecords.Columns()
	_, err := dao.TaskRecords.Ctx(ctx).
		WherePri(recordID).
		WhereIn(columns.Status, []string{"pending", "queued", "running"}).
		Data(do.TaskRecords{
			Status:       "failed",
			ErrorMessage: strings.TrimSpace(message),
			FinishedAt:   gtime.Now(),
		}).
		Update()
	return err
}

func (p persistence) MarkNodeIdle(ctx context.Context, clientIP string, nodeID string, commandID string, recordID int64) error {
	clientIP = strings.TrimSpace(clientIP)
	nodeID = strings.TrimSpace(nodeID)
	if clientIP == "" && nodeID == "" {
		return nil
	}
	columns := dao.Clients.Columns()
	updateModel := dao.Clients.Ctx(ctx)
	if clientIP != "" && nodeID != "" {
		updateModel = updateModel.Where(columns.ClientIp, clientIP).Where(columns.NodeId, nodeID)
	} else if nodeID != "" {
		updateModel = updateModel.Where(columns.NodeId, nodeID)
	} else {
		updateModel = updateModel.Where(columns.ClientIp, clientIP)
	}
	if strings.TrimSpace(commandID) != "" {
		updateModel = updateModel.Where(columns.CurrentExecutionId, strings.TrimSpace(commandID))
	} else if recordID > 0 {
		updateModel = updateModel.Where(columns.CurrentTaskRecordId, recordID)
	}
	_, err := updateModel.Data(do.Clients{
		BusyStatus:          "idle",
		CurrentExecutionId:  "",
		CurrentTaskRecordId: 0,
	}).Update()
	return err
}

func saveTaskRecordResultFiles(ctx context.Context, recordID int64, clientIP string, resultJSON string) string {
	var result map[string]any
	if recordID <= 0 || strings.TrimSpace(resultJSON) == "" || json.Unmarshal([]byte(resultJSON), &result) != nil {
		return resultJSON
	}

	data, ok := result["data"].(map[string]any)
	if !ok {
		return resultJSON
	}
	table, ok := data["table"].([]any)
	if !ok || len(table) == 0 {
		return resultJSON
	}

	recordColumns := dao.TaskRecords.Columns()
	record, err := dao.TaskRecords.Ctx(ctx).WherePri(recordID).One()
	if err != nil || record.IsEmpty() {
		if err != nil {
			g.Log().Line().Warningf(ctx, "query task record before saving result file failed: record_id=%d err=%+v", recordID, err)
		}
		return resultJSON
	}

	fileDir := filepath.Join(taskRecordResultFilesBaseDir, fmt.Sprintf("%d", recordID))
	if err = os.MkdirAll(fileDir, 0o755); err != nil {
		g.Log().Line().Warningf(ctx, "create task result file directory failed: record_id=%d err=%+v", recordID, err)
		return resultJSON
	}

	fileName := fmt.Sprintf("task-record-%d-table.json", recordID)
	filePath := filepath.Join(fileDir, fileName)
	fileBody, err := json.MarshalIndent(table, "", "  ")
	if err != nil {
		g.Log().Line().Warningf(ctx, "marshal task table result failed: record_id=%d err=%+v", recordID, err)
		return resultJSON
	}
	if err = os.WriteFile(filePath, fileBody, 0o644); err != nil {
		g.Log().Line().Warningf(ctx, "write task table result file failed: record_id=%d err=%+v", recordID, err)
		return resultJSON
	}

	preview := table
	if len(preview) > taskRecordTablePreviewLimit {
		preview = preview[:taskRecordTablePreviewLimit]
	}
	previewBody, err := json.Marshal(preview)
	if err != nil {
		previewBody = []byte("[]")
	}

	totalRows := gconv.Int(data["table_total"])
	if totalRows <= 0 {
		totalRows = len(table)
	}
	_, err = dao.TaskRecordFiles.Ctx(ctx).Data(do.TaskRecordFiles{
		RecordId:    recordID,
		TaskId:      gconv.Int64(record[recordColumns.TaskId]),
		WorkflowId:  strings.TrimSpace(gconv.String(record[recordColumns.WorkflowId])),
		ClientIp:    strings.TrimSpace(clientIP),
		NodeId:      strings.TrimSpace(gconv.String(record[recordColumns.NodeId])),
		ExecutionId: strings.TrimSpace(gconv.String(record[recordColumns.ExecutionId])),
		FileType:    "table_json",
		FileName:    fileName,
		FilePath:    filepath.ToSlash(filePath),
		MimeType:    "application/json",
		FileSize:    int64(len(fileBody)),
		RowCount:    totalRows,
		PreviewJson: string(previewBody),
		Remark:      buildTaskRecordTableFileRemark(data, len(table), totalRows),
	}).Insert()
	if err != nil {
		g.Log().Line().Warningf(ctx, "save task result file record failed: record_id=%d err=%+v", recordID, err)
		return resultJSON
	}

	data["table"] = preview
	data["table_preview_count"] = len(preview)
	data["table_file"] = map[string]any{
		"type":      "table_json",
		"file_name": fileName,
		"file_path": filepath.ToSlash(filePath),
		"file_size": len(fileBody),
		"row_count": totalRows,
		"preview":   len(preview),
		"truncated": gconv.Bool(data["table_truncated"]),
	}

	trimmedResult, err := json.Marshal(result)
	if err != nil {
		return resultJSON
	}
	return string(trimmedResult)
}

func buildTaskRecordTableFileRemark(data map[string]any, returnedRows int, totalRows int) string {
	limitRows := gconv.Int(data["table_limit"])
	if gconv.Bool(data["table_truncated"]) {
		return fmt.Sprintf("Automa returned %d/%d rows, limited by table_limit=%d", returnedRows, totalRows, limitRows)
	}
	return fmt.Sprintf("Automa returned %d rows", returnedRows)
}
