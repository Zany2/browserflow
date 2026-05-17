package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorFillForm fills multiple form fields. 批量填写表单字段。
func (c *ControllerV1) BrowserExecutorFillForm(ctx context.Context, req *v1.BrowserExecutorFillFormReq) (res *v1.BrowserExecutorFillFormRes, err error) {
	executor, err := browserexecutor.Current(ctx)
	if err != nil {
		return nil, err
	}
	fields := make([]model.BrowserExecutorFormField, 0, len(req.Fields))
	for _, field := range req.Fields {
		// Convert API field. 转换接口字段。
		fields = append(fields, model.BrowserExecutorFormField{
			Name:  field.Name,
			Value: field.Value,
			Type:  field.Type,
		})
	}
	result, opErr := executor.FillForm(ctx, fields, req.Submit, req.Timeout)
	result, observeErr := executor.AppendObserve(ctx, result, req.ReturnObserve, req.IncludeText, req.TextLimit)
	if opErr == nil {
		opErr = observeErr
	}
	return &v1.BrowserExecutorFillFormRes{Result: result}, opErr
}
