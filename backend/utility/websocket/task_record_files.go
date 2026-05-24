package websockets

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/Zany2/browserflow/backend/internal/model/do"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
)

const (
	taskRecordTablePreviewLimit  = 20
	taskRecordResultFilesBaseDir = "data/task-record-files"
)

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
	now := gtime.Now()
	_, err = dao.TaskRecordFiles.Ctx(ctx).Data(do.TaskRecordFiles{
		RecordId:    recordID,
		TaskId:      gconv.Int64(record[recordColumns.TaskId]),
		WorkflowId:  strings.TrimSpace(gconv.String(record[recordColumns.WorkflowId])),
		ClientIp:    strings.TrimSpace(clientIP),
		MachineId:   strings.TrimSpace(gconv.String(record[recordColumns.MachineId])),
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
		CreatedAt:   now,
		UpdatedAt:   now,
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
