package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/browserexecutor"
)

// BrowserExecutorAct runs a smart browser action. 执行智能浏览器动作。
func (c *ControllerV1) BrowserExecutorAct(ctx context.Context, req *v1.BrowserExecutorActReq) (res *v1.BrowserExecutorActRes, err error) {
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
	clear := true
	if req.Clear != nil {
		clear = *req.Clear
	}
	result, opErr := executor.Act(ctx, model.BrowserExecutorActOptions{
		Intent:     req.Intent,
		Identifier: req.Identifier,
		Value:      req.Value,
		Text:       req.Text,
		Fields:     fields,
		Submit:     req.Submit,
		Clear:      clear,
		Timeout:    req.Timeout,
	})
	result, observeErr := executor.AppendObserve(ctx, result, req.ReturnObserve, req.IncludeText, req.TextLimit)
	if opErr == nil {
		opErr = observeErr
	}
	return &v1.BrowserExecutorActRes{Result: result}, opErr
}
