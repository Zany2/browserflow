// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package browserexecutor

import (
	"context"

	"github.com/Zany2/browserflow/backend/api/browserexecutor/v1"
)

type IBrowserexecutorV1 interface {
	BrowserExecutorStatus(ctx context.Context, req *v1.BrowserExecutorStatusReq) (res *v1.BrowserExecutorStatusRes, err error)
	BrowserExecutorHelp(ctx context.Context, req *v1.BrowserExecutorHelpReq) (res *v1.BrowserExecutorHelpRes, err error)
	BrowserExecutorExportSkill(ctx context.Context, req *v1.BrowserExecutorExportSkillReq) (res *v1.BrowserExecutorExportSkillRes, err error)
	BrowserExecutorNavigate(ctx context.Context, req *v1.BrowserExecutorNavigateReq) (res *v1.BrowserExecutorNavigateRes, err error)
	BrowserExecutorSnapshot(ctx context.Context, req *v1.BrowserExecutorSnapshotReq) (res *v1.BrowserExecutorSnapshotRes, err error)
	BrowserExecutorClickableElements(ctx context.Context, req *v1.BrowserExecutorClickableElementsReq) (res *v1.BrowserExecutorClickableElementsRes, err error)
	BrowserExecutorInputElements(ctx context.Context, req *v1.BrowserExecutorInputElementsReq) (res *v1.BrowserExecutorInputElementsRes, err error)
	BrowserExecutorClick(ctx context.Context, req *v1.BrowserExecutorClickReq) (res *v1.BrowserExecutorClickRes, err error)
	BrowserExecutorType(ctx context.Context, req *v1.BrowserExecutorTypeReq) (res *v1.BrowserExecutorTypeRes, err error)
	BrowserExecutorSelect(ctx context.Context, req *v1.BrowserExecutorSelectReq) (res *v1.BrowserExecutorSelectRes, err error)
	BrowserExecutorPressKey(ctx context.Context, req *v1.BrowserExecutorPressKeyReq) (res *v1.BrowserExecutorPressKeyRes, err error)
	BrowserExecutorWait(ctx context.Context, req *v1.BrowserExecutorWaitReq) (res *v1.BrowserExecutorWaitRes, err error)
	BrowserExecutorReload(ctx context.Context, req *v1.BrowserExecutorReloadReq) (res *v1.BrowserExecutorReloadRes, err error)
	BrowserExecutorGoBack(ctx context.Context, req *v1.BrowserExecutorGoBackReq) (res *v1.BrowserExecutorGoBackRes, err error)
	BrowserExecutorGoForward(ctx context.Context, req *v1.BrowserExecutorGoForwardReq) (res *v1.BrowserExecutorGoForwardRes, err error)
	BrowserExecutorHover(ctx context.Context, req *v1.BrowserExecutorHoverReq) (res *v1.BrowserExecutorHoverRes, err error)
	BrowserExecutorResize(ctx context.Context, req *v1.BrowserExecutorResizeReq) (res *v1.BrowserExecutorResizeRes, err error)
	BrowserExecutorPageInfo(ctx context.Context, req *v1.BrowserExecutorPageInfoReq) (res *v1.BrowserExecutorPageInfoRes, err error)
	BrowserExecutorObserve(ctx context.Context, req *v1.BrowserExecutorObserveReq) (res *v1.BrowserExecutorObserveRes, err error)
	BrowserExecutorGetText(ctx context.Context, req *v1.BrowserExecutorGetTextReq) (res *v1.BrowserExecutorGetTextRes, err error)
	BrowserExecutorGetValue(ctx context.Context, req *v1.BrowserExecutorGetValueReq) (res *v1.BrowserExecutorGetValueRes, err error)
	BrowserExecutorElementInfo(ctx context.Context, req *v1.BrowserExecutorElementInfoReq) (res *v1.BrowserExecutorElementInfoRes, err error)
	BrowserExecutorPageText(ctx context.Context, req *v1.BrowserExecutorPageTextReq) (res *v1.BrowserExecutorPageTextRes, err error)
	BrowserExecutorPageContent(ctx context.Context, req *v1.BrowserExecutorPageContentReq) (res *v1.BrowserExecutorPageContentRes, err error)
	BrowserExecutorPageStructure(ctx context.Context, req *v1.BrowserExecutorPageStructureReq) (res *v1.BrowserExecutorPageStructureRes, err error)
	BrowserExecutorExtract(ctx context.Context, req *v1.BrowserExecutorExtractReq) (res *v1.BrowserExecutorExtractRes, err error)
	BrowserExecutorScreenshot(ctx context.Context, req *v1.BrowserExecutorScreenshotReq) (res *v1.BrowserExecutorScreenshotRes, err error)
	BrowserExecutorElementScreenshot(ctx context.Context, req *v1.BrowserExecutorElementScreenshotReq) (res *v1.BrowserExecutorElementScreenshotRes, err error)
	BrowserExecutorEvaluate(ctx context.Context, req *v1.BrowserExecutorEvaluateReq) (res *v1.BrowserExecutorEvaluateRes, err error)
	BrowserExecutorTabs(ctx context.Context, req *v1.BrowserExecutorTabsReq) (res *v1.BrowserExecutorTabsRes, err error)
	BrowserExecutorScroll(ctx context.Context, req *v1.BrowserExecutorScrollReq) (res *v1.BrowserExecutorScrollRes, err error)
	BrowserExecutorMouse(ctx context.Context, req *v1.BrowserExecutorMouseReq) (res *v1.BrowserExecutorMouseRes, err error)
	BrowserExecutorWindow(ctx context.Context, req *v1.BrowserExecutorWindowReq) (res *v1.BrowserExecutorWindowRes, err error)
	BrowserExecutorClosePage(ctx context.Context, req *v1.BrowserExecutorClosePageReq) (res *v1.BrowserExecutorClosePageRes, err error)
	BrowserExecutorFillForm(ctx context.Context, req *v1.BrowserExecutorFillFormReq) (res *v1.BrowserExecutorFillFormRes, err error)
	BrowserExecutorDrag(ctx context.Context, req *v1.BrowserExecutorDragReq) (res *v1.BrowserExecutorDragRes, err error)
	BrowserExecutorFileUpload(ctx context.Context, req *v1.BrowserExecutorFileUploadReq) (res *v1.BrowserExecutorFileUploadRes, err error)
	BrowserExecutorHandleDialog(ctx context.Context, req *v1.BrowserExecutorHandleDialogReq) (res *v1.BrowserExecutorHandleDialogRes, err error)
	BrowserExecutorAct(ctx context.Context, req *v1.BrowserExecutorActReq) (res *v1.BrowserExecutorActRes, err error)
	BrowserExecutorBatch(ctx context.Context, req *v1.BrowserExecutorBatchReq) (res *v1.BrowserExecutorBatchRes, err error)
}
