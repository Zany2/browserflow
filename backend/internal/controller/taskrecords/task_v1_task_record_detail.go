package taskrecords

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/Zany2/browserflow/backend/api/taskrecords/v1"
	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/rr"
	"github.com/Zany2/browserflow/backend/utility/taskdata"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

// TaskRecordDetail returns task record detail 获取任务记录详情
func (c *ControllerV1) TaskRecordDetail(ctx context.Context, req *v1.TaskRecordDetailReq) (res *v1.TaskRecordDetailRes, err error) {
	record, err := dao.TaskRecords.Ctx(ctx).
		WherePri(gconv.Int64(req.ID)).
		One()
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		rr.FailedJsonWithMessageExitAll(g.RequestFromCtx(ctx), "执行记录不存在")
		return nil, nil
	}

	recordMap, err := taskdata.BuildTaskRecordMap(ctx, record)
	if err != nil {
		return nil, err
	}

	fileColumns := dao.TaskRecordFiles.Columns()
	fileRecords, err := dao.TaskRecordFiles.Ctx(ctx).
		Where(fileColumns.RecordId, gconv.Int64(req.ID)).
		Where(fileColumns.DeletedAt + " IS NULL").
		OrderDesc(fileColumns.CreatedAt).
		OrderDesc(fileColumns.Id).
		All()
	if err != nil {
		return nil, err
	}
	files := make([]*model.TaskRecordFileResModel, 0, len(fileRecords))
	for _, fileRecord := range fileRecords {
		preview := model.JSONMap{}
		if previewJSON := strings.TrimSpace(gconv.String(fileRecord[fileColumns.PreviewJson])); previewJSON != "" {
			if err := json.Unmarshal([]byte(previewJSON), &preview); err != nil {
				rawPreview, _ := json.Marshal(previewJSON)
				preview = model.JSONMap{"raw": rawPreview}
			}
		}
		files = append(files, &model.TaskRecordFileResModel{
			ID:         gconv.Int64(fileRecord[fileColumns.Id]),
			RecordID:   gconv.Int64(fileRecord[fileColumns.RecordId]),
			TaskID:     gconv.Int64(fileRecord[fileColumns.TaskId]),
			WorkflowID: strings.TrimSpace(gconv.String(fileRecord[fileColumns.WorkflowId])),
			ClientIP:   strings.TrimSpace(gconv.String(fileRecord[fileColumns.ClientIp])),
			FileType:   strings.TrimSpace(gconv.String(fileRecord[fileColumns.FileType])),
			FileName:   strings.TrimSpace(gconv.String(fileRecord[fileColumns.FileName])),
			FilePath:   strings.TrimSpace(gconv.String(fileRecord[fileColumns.FilePath])),
			MimeType:   strings.TrimSpace(gconv.String(fileRecord[fileColumns.MimeType])),
			FileSize:   gconv.Int64(fileRecord[fileColumns.FileSize]),
			RowCount:   gconv.Int(fileRecord[fileColumns.RowCount]),
			Preview:    preview,
			Remark:     strings.TrimSpace(gconv.String(fileRecord[fileColumns.Remark])),
			CreatedAt:  fileRecord[fileColumns.CreatedAt].GTime(),
			UpdatedAt:  fileRecord[fileColumns.UpdatedAt].GTime(),
			DeletedAt:  fileRecord[fileColumns.DeletedAt].GTime(),
		})
	}
	return &v1.TaskRecordDetailRes{Record: recordMap, Files: files}, nil
}
