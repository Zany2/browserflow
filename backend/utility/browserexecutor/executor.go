package browserexecutor

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/Zany2/browserflow/backend/utility/state"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

const defaultTimeout = 10 * time.Second

// Executor controls one BrowserFlow browser runtime. Executor controls one BrowserFlow browser runtime.
type Executor struct {
	browserID string
	runtime   *state.BrowserRuntime
	ctx       context.Context

	refMu        sync.RWMutex
	refMap       map[string]*model.BrowserExecutorRefData
	refCounter   int
	refSnapshot  *model.BrowserExecutorAccessibilitySnapshot
	refPageURL   string
	refPageID    string
	refTimestamp time.Time
	refTTL       time.Duration
}

// NewExecutor creates a browser executor. NewExecutor creates browser executor.
func NewExecutor(browserID string, runtime *state.BrowserRuntime) *Executor {
	return &Executor{
		browserID: browserID,
		runtime:   runtime,
		ctx:       context.Background(),
		refMap:    map[string]*model.BrowserExecutorRefData{},
		refTTL:    5 * time.Minute,
	}
}

// WithContext updates request context. WithContext updates request context.
func (e *Executor) WithContext(ctx context.Context) *Executor {
	if ctx != nil {
		e.ctx = ctx
	}
	return e
}

// Status returns current browser page status. Status returns current browser page status.
func (e *Executor) Status(ctx context.Context) model.BrowserExecutorStatus {
	page, err := e.activePage()
	if err != nil {
		return model.BrowserExecutorStatus{Running: false, BrowserID: e.browserID, Message: err.Error()}
	}
	info, _ := page.Info()
	status := model.BrowserExecutorStatus{Running: true, BrowserID: e.browserID, HasBusinessPage: e.hasBusinessPage()}
	if e.runtime != nil && e.runtime.Instance != nil {
		status.BrowserName = e.runtime.Instance.Name
	}
	if info != nil {
		status.URL = info.URL
		status.Title = info.Title
	}
	return status
}

// activePage returns a controllable page, preferring non browser-agent pages. activePage returns controllable page.
func (e *Executor) activePage() (*rod.Page, error) {
	page, err := e.businessPage()
	if err == nil {
		return page, nil
	}
	if agentPage, agentErr := e.agentPage(); agentErr == nil {
		return agentPage, nil
	}
	return nil, err
}

// businessPage returns a controllable non-agent page. businessPage returns business page.
func (e *Executor) businessPage() (*rod.Page, error) {
	if e.runtime == nil || e.runtime.Browser == nil {
		return nil, fmt.Errorf("browser runtime is unavailable")
	}

	pages, err := e.runtime.Browser.Pages()
	if err != nil {
		return nil, fmt.Errorf("failed to read browser pages: %w", err)
	}
	for i := len(pages) - 1; i >= 0; i-- {
		page := pages[i]
		info, infoErr := page.Info()
		if infoErr != nil || info == nil || info.Type != "page" {
			continue
		}
		if strings.Contains(info.URL, "#/browser-agent") {
			continue
		}
		return page, nil
	}
	return nil, fmt.Errorf("no controllable business page")
}

// agentPage returns the BrowserFlow client page when present. agentPage returns browser-agent page.
func (e *Executor) agentPage() (*rod.Page, error) {
	if e.runtime == nil || e.runtime.Browser == nil {
		return nil, fmt.Errorf("browser runtime is unavailable")
	}

	pages, err := e.runtime.Browser.Pages()
	if err != nil {
		return nil, fmt.Errorf("failed to read browser pages: %w", err)
	}
	for _, page := range pages {
		info, infoErr := page.Info()
		if infoErr == nil && info != nil && info.Type == "page" && isBrowserAgentURL(info.URL) {
			return page, nil
		}
	}
	return nil, fmt.Errorf("no BrowserFlow browser-agent page")
}

// hasBusinessPage reports whether a task page is available. hasBusinessPage reports business page availability.
func (e *Executor) hasBusinessPage() bool {
	_, err := e.businessPage()
	return err == nil
}

// InvalidateSnapshot clears RefID cache. InvalidateSnapshot clears RefID cache.
func (e *Executor) InvalidateSnapshot() {
	e.refMu.Lock()
	defer e.refMu.Unlock()
	e.refMap = map[string]*model.BrowserExecutorRefData{}
	e.refCounter = 0
	e.refSnapshot = nil
	e.refPageURL = ""
	e.refPageID = ""
	e.refTimestamp = time.Time{}
}

// GetSnapshot returns cached or fresh accessibility snapshot. GetSnapshot returns cached or fresh snapshot.
func (e *Executor) GetSnapshot(ctx context.Context) (*model.BrowserExecutorAccessibilitySnapshot, error) {
	page, err := e.activePage()
	if err != nil {
		return nil, err
	}
	pageURL, pageID := pageIdentity(page)

	e.refMu.RLock()
	cacheValid := e.refSnapshot != nil &&
		time.Since(e.refTimestamp) < e.refTTL &&
		e.refPageURL == pageURL &&
		e.refPageID == pageID
	if cacheValid {
		snapshot := e.refSnapshot
		e.refMu.RUnlock()
		return snapshot, nil
	}
	e.refMu.RUnlock()

	snapshot, err := buildAccessibilitySnapshot(ctx, page)
	if err != nil {
		return nil, err
	}

	e.refMu.Lock()
	defer e.refMu.Unlock()
	e.refMap = map[string]*model.BrowserExecutorRefData{}
	e.refCounter = 0
	e.assignRefIDs(snapshot)
	e.refSnapshot = snapshot
	e.refPageURL = pageURL
	e.refPageID = pageID
	e.refTimestamp = time.Now()
	return snapshot, nil
}

// Snapshot builds a response for model consumption. Snapshot builds response for model use.
func (e *Executor) Snapshot(ctx context.Context) (*model.BrowserExecutorSnapshotResult, error) {
	snapshot, err := e.GetSnapshot(ctx)
	if err != nil {
		return &model.BrowserExecutorSnapshotResult{Success: false, Error: err.Error(), Timestamp: time.Now()}, err
	}
	return &model.BrowserExecutorSnapshotResult{
		Success:      true,
		SnapshotText: snapshotText(snapshot.Refs),
		Elements:     snapshot.Refs,
		Timestamp:    time.Now(),
		Data: map[string]interface{}{
			"browser_id": e.browserID,
			"count":      len(snapshot.Refs),
		},
	}, nil
}

// assignRefIDs assigns stable RefIDs to interactive AX nodes. assignRefIDs assigns stable RefIDs.
func (e *Executor) assignRefIDs(snapshot *model.BrowserExecutorAccessibilitySnapshot) {
	roleNameCounter := map[string]int{}
	refs := make([]model.BrowserExecutorElementRef, 0)

	for _, node := range snapshot.Elements {
		if !isInteractiveRole(node.Role) && !isUsefulInputRole(node.Role) {
			continue
		}
		if strings.TrimSpace(node.Name) == "" && strings.TrimSpace(node.Value) == "" && strings.TrimSpace(node.Placeholder) == "" {
			continue
		}

		key := strings.ToLower(node.Role + ":" + node.Name + ":" + node.Placeholder)
		nth := roleNameCounter[key]
		roleNameCounter[key]++

		e.refCounter++
		refID := fmt.Sprintf("e%d", e.refCounter)
		node.RefID = refID

		refData := &model.BrowserExecutorRefData{
			Role:        node.Role,
			Name:        node.Name,
			Nth:         nth,
			BackendID:   node.BackendNodeID,
			Attributes:  node.Attributes,
			Placeholder: node.Placeholder,
		}
		if node.Attributes != nil {
			refData.Tag = node.Attributes["tag"]
			refData.Href = node.Attributes["href"]
		}
		e.refMap[refID] = refData

		refs = append(refs, model.BrowserExecutorElementRef{
			RefID:       "@" + refID,
			Role:        node.Role,
			Name:        node.Name,
			Value:       node.Value,
			Placeholder: node.Placeholder,
			BackendID:   node.BackendNodeID,
			Attributes:  compactAttributes(node.Attributes),
		})
	}
	snapshot.Refs = refs
}

func pageIdentity(page *rod.Page) (string, string) {
	info, err := page.Info()
	if err != nil || info == nil {
		return "", string(page.TargetID)
	}
	return info.URL, string(info.TargetID)
}

func operationOK(message string, data map[string]any) *model.BrowserExecutorOperationResult {
	return &model.BrowserExecutorOperationResult{Success: true, Message: message, Data: data, Timestamp: time.Now()}
}

func operationFail(message string, err error) *model.BrowserExecutorOperationResult {
	result := &model.BrowserExecutorOperationResult{Success: false, Error: message, Timestamp: time.Now()}
	if err != nil {
		result.Error = err.Error()
	}
	return result
}

func screenshotFormat(format string) proto.PageCaptureScreenshotFormat {
	if strings.EqualFold(format, "jpeg") || strings.EqualFold(format, "jpg") {
		return proto.PageCaptureScreenshotFormatJpeg
	}
	return proto.PageCaptureScreenshotFormatPng
}

// ensureBusinessPage returns a task page, creating one when only agent tab exists. ensureBusinessPage returns task page.
func (e *Executor) ensureBusinessPage(url string) (*rod.Page, bool, error) {
	page, err := e.businessPage()
	if err == nil {
		return page, false, nil
	}
	if e.runtime == nil || e.runtime.Browser == nil {
		return nil, false, fmt.Errorf("browser runtime is unavailable")
	}
	page, err = e.runtime.Browser.Page(proto.TargetCreateTarget{URL: url})
	if err != nil {
		return nil, false, err
	}
	_, _ = page.Activate()
	return page, true, nil
}

// isBrowserAgentURL reports whether URL points to BrowserFlow client page. isBrowserAgentURL checks agent URL.
func isBrowserAgentURL(url string) bool {
	return strings.Contains(strings.TrimSpace(url), "#/browser-agent")
}
