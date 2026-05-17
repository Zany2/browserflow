package browserexecutor

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/proto"
)

// Navigate opens a URL in the active page. Navigate opens URL in current page.
func (e *Executor) Navigate(ctx context.Context, url string, timeoutSeconds int) (*model.BrowserExecutorOperationResult, error) {
	if !strings.Contains(url, "://") {
		url = "https://" + url
	}
	page, created, err := e.ensureBusinessPage(url)
	if err != nil {
		return operationFail("", err), err
	}
	timeout := secondsOrDefault(timeoutSeconds, 60)
	if !created {
		if err = page.Timeout(timeout).Navigate(url); err != nil {
			return operationFail("", err), err
		}
	}
	if err = page.Timeout(timeout).WaitLoad(); err != nil {
		return operationFail("", err), err
	}
	e.InvalidateSnapshot()
	return operationOK("Page opened", map[string]any{"url": url, "created_tab": created}), nil
}

// Click clicks an element. Click clicks element by identifier.
func (e *Executor) Click(ctx context.Context, identifier string) (*model.BrowserExecutorOperationResult, error) {
	page, err := e.activePage()
	if err != nil {
		return operationFail("", err), err
	}
	elem, err := e.findElement(ctx, page, identifier, defaultTimeout)
	if err != nil {
		return operationFail("", err), err
	}
	_ = elem.ScrollIntoView()
	if _, jsErr := elem.Eval(`() => { this.focus && this.focus(); this.click && this.click(); }`); jsErr != nil {
		if clickErr := elem.Click(proto.InputMouseButtonLeft, 1); clickErr != nil {
			return operationFail("", clickErr), clickErr
		}
	}
	e.InvalidateSnapshot()
	return operationOK("Element clicked", map[string]any{"identifier": identifier}), nil
}

// Type types text into an element. Type inputs text by identifier.
func (e *Executor) Type(ctx context.Context, identifier string, text string, clear bool) (*model.BrowserExecutorOperationResult, error) {
	page, err := e.activePage()
	if err != nil {
		return operationFail("", err), err
	}
	elem, err := e.findElement(ctx, page, identifier, defaultTimeout)
	if err != nil {
		return operationFail("", err), err
	}
	if err = elem.Focus(); err != nil {
		return operationFail("", err), err
	}
	if clear {
		if err = elem.SelectAllText(); err == nil {
			_ = page.Keyboard.Type(input.Backspace)
		}
	}
	if err = elem.Input(text); err != nil {
		return operationFail("", err), err
	}
	e.InvalidateSnapshot()
	return operationOK("Text typed", map[string]any{"identifier": identifier}), nil
}

// Select selects an option by value or visible text. 选择下拉选项。
func (e *Executor) Select(ctx context.Context, identifier string, value string) (*model.BrowserExecutorOperationResult, error) {
	page, err := e.activePage()
	if err != nil {
		return operationFail("", err), err
	}
	elem, err := e.findElement(ctx, page, identifier, defaultTimeout)
	if err != nil {
		return operationFail("", err), err
	}
	if err = elem.Select([]string{value}, true, rod.SelectorTypeRegex); err != nil {
		if fallbackErr := elem.Select([]string{value}, true, rod.SelectorTypeText); fallbackErr != nil {
			return operationFail("", fallbackErr), fallbackErr
		}
	}
	e.InvalidateSnapshot()
	return operationOK("Option selected", map[string]any{"identifier": identifier, "value": value}), nil
}

// PressKey presses one key or shortcut. PressKey sends a key or shortcut.
func (e *Executor) PressKey(ctx context.Context, key string, ctrl bool, shift bool, alt bool, meta bool) (*model.BrowserExecutorOperationResult, error) {
	page, err := e.activePage()
	if err != nil {
		return operationFail("", err), err
	}
	keyboard := page.Keyboard
	if ctrl {
		_ = keyboard.Press(input.ControlLeft)
		defer keyboard.Release(input.ControlLeft)
	}
	if shift {
		_ = keyboard.Press(input.ShiftLeft)
		defer keyboard.Release(input.ShiftLeft)
	}
	if alt {
		_ = keyboard.Press(input.AltLeft)
		defer keyboard.Release(input.AltLeft)
	}
	if meta {
		_ = keyboard.Press(input.MetaLeft)
		defer keyboard.Release(input.MetaLeft)
	}
	if err = keyboard.Type(normalizeKey(key)); err != nil {
		return operationFail("", err), err
	}
	e.InvalidateSnapshot()
	return operationOK("Key pressed", map[string]any{"key": key}), nil
}

// Wait waits for page, element, or time. Wait waits for page, element, or duration.
func (e *Executor) Wait(ctx context.Context, identifier string, state string, timeoutSeconds int, count int) (*model.BrowserExecutorOperationResult, error) {
	page, err := e.activePage()
	if err != nil {
		return operationFail("", err), err
	}
	timeout := secondsOrDefault(timeoutSeconds, 10)
	state = strings.ToLower(strings.TrimSpace(state))
	if state == "" {
		state = "load"
	}
	if state == "time" {
		time.Sleep(timeout)
		return operationOK("Wait completed", map[string]any{"state": state, "timeout": int(timeout.Seconds())}), nil
	}
	if state == "load" || strings.TrimSpace(identifier) == "" && isPageWaitState(state) {
		err = page.Timeout(timeout).WaitLoad()
	} else if state == "dom-stable" {
		err = page.Timeout(timeout).WaitDOMStable(500*time.Millisecond, 0)
	} else if state == "request-idle" {
		page.Timeout(timeout).WaitRequestIdle(500*time.Millisecond, nil, nil, nil)()
	} else if state == "page-stable" || state == "stable" && strings.TrimSpace(identifier) == "" {
		err = page.Timeout(timeout).WaitStable(500 * time.Millisecond)
	} else if state == "elements-more-than" {
		if strings.TrimSpace(identifier) == "" {
			err = fmt.Errorf("identifier css selector is required for elements-more-than")
		} else {
			err = page.Timeout(timeout).WaitElementsMoreThan(identifier, count)
		}
	} else {
		var elem *rod.Element
		elem, err = e.findElement(ctx, page, identifier, timeout)
		if err == nil {
			elem = elem.Timeout(timeout)
			if state == "hidden" {
				err = elem.Timeout(timeout).WaitInvisible()
			} else if state == "enabled" {
				err = elem.WaitEnabled()
			} else if state == "interactable" {
				_, err = elem.WaitInteractable()
			} else if state == "writable" {
				err = elem.WaitWritable()
			} else if state == "stable" {
				err = elem.WaitStable(300 * time.Millisecond)
			} else {
				err = elem.WaitVisible()
			}
		}
	}
	if err != nil {
		return operationFail("", err), err
	}
	return operationOK("Wait completed", map[string]any{"state": state, "timeout": int(timeout.Seconds()), "count": count}), nil
}

// Reload reloads the current page. 刷新当前页面。
func (e *Executor) Reload(ctx context.Context) (*model.BrowserExecutorOperationResult, error) {
	page, err := e.activePage()
	if err != nil {
		return operationFail("", err), err
	}
	if err = page.Reload(); err != nil {
		return operationFail("", err), err
	}
	e.InvalidateSnapshot()
	return operationOK("Page reloaded", nil), nil
}

// GoBack navigates browser history backward. 浏览器历史后退。
func (e *Executor) GoBack(ctx context.Context) (*model.BrowserExecutorOperationResult, error) {
	page, err := e.activePage()
	if err != nil {
		return operationFail("", err), err
	}
	if err = page.NavigateBack(); err != nil {
		return operationFail("", err), err
	}
	e.InvalidateSnapshot()
	return operationOK("Went back", nil), nil
}

// GoForward navigates browser history forward. 浏览器历史前进。
func (e *Executor) GoForward(ctx context.Context) (*model.BrowserExecutorOperationResult, error) {
	page, err := e.activePage()
	if err != nil {
		return operationFail("", err), err
	}
	if err = page.NavigateForward(); err != nil {
		return operationFail("", err), err
	}
	e.InvalidateSnapshot()
	return operationOK("Went forward", nil), nil
}

// Hover moves cursor over an element. 鼠标悬停到元素上。
func (e *Executor) Hover(ctx context.Context, identifier string) (*model.BrowserExecutorOperationResult, error) {
	page, err := e.activePage()
	if err != nil {
		return operationFail("", err), err
	}
	elem, err := e.findElement(ctx, page, identifier, defaultTimeout)
	if err != nil {
		return operationFail("", err), err
	}
	if err = elem.Hover(); err != nil {
		return operationFail("", err), err
	}
	return operationOK("Element hovered", map[string]any{"identifier": identifier}), nil
}

// Resize changes viewport size. 调整视口大小。
func (e *Executor) Resize(ctx context.Context, width int, height int) (*model.BrowserExecutorOperationResult, error) {
	page, err := e.activePage()
	if err != nil {
		return operationFail("", err), err
	}
	if width <= 0 {
		width = 1440
	}
	if height <= 0 {
		height = 900
	}
	if err = page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
		Width:             width,
		Height:            height,
		DeviceScaleFactor: 1,
		Mobile:            false,
	}); err != nil {
		return operationFail("", err), err
	}
	e.InvalidateSnapshot()
	return operationOK("Viewport resized", map[string]any{"width": width, "height": height}), nil
}

// ClosePage closes current page. 关闭当前页面。
func (e *Executor) ClosePage(ctx context.Context) (*model.BrowserExecutorOperationResult, error) {
	page, err := e.businessPage()
	if err != nil {
		return operationFail("", err), err
	}
	info, _ := page.Info()
	if info != nil && isBrowserAgentURL(info.URL) {
		err = fmt.Errorf("refusing to close BrowserFlow browser-agent tab")
		return operationFail("", err), err
	}
	if err = page.Close(); err != nil {
		return operationFail("", err), err
	}
	e.InvalidateSnapshot()
	return operationOK("Page closed", nil), nil
}

// FillForm fills multiple form fields in one operation. 一次性填写多个表单字段。
func (e *Executor) FillForm(ctx context.Context, fields []model.BrowserExecutorFormField, submit bool, timeoutSeconds int) (*model.BrowserExecutorOperationResult, error) {
	page, err := e.activePage()
	if err != nil {
		return operationFail("", err), err
	}
	if len(fields) == 0 {
		err = fmt.Errorf("form fields cannot be empty")
		return operationFail("", err), err
	}

	timeout := secondsOrDefault(timeoutSeconds, 10)
	filled := 0
	fieldErrors := make([]string, 0)
	for _, field := range fields {
		// Fill each field independently. 单独填写每个字段。
		if fillErr := e.fillSingleField(ctx, page, field, timeout); fillErr != nil {
			fieldErrors = append(fieldErrors, fmt.Sprintf("%s: %s", field.Name, fillErr.Error()))
			continue
		}
		filled++
	}

	submitted := false
	if submit {
		if submitErr := e.submitForm(ctx, page); submitErr != nil {
			fieldErrors = append(fieldErrors, "submit: "+submitErr.Error())
		} else {
			submitted = true
		}
	}

	e.InvalidateSnapshot()
	return operationOK("Form filled", map[string]any{
		"filled_count": filled,
		"total_fields": len(fields),
		"errors":       fieldErrors,
		"submitted":    submitted,
	}), nil
}

// Drag drags one element to another element. 拖拽一个元素到另一个元素。
func (e *Executor) Drag(ctx context.Context, fromIdentifier string, toIdentifier string) (*model.BrowserExecutorOperationResult, error) {
	page, err := e.activePage()
	if err != nil {
		return operationFail("", err), err
	}
	from, err := e.findElement(ctx, page, fromIdentifier, defaultTimeout)
	if err != nil {
		return operationFail("", err), err
	}
	to, err := e.findElement(ctx, page, toIdentifier, defaultTimeout)
	if err != nil {
		return operationFail("", err), err
	}
	_ = from.ScrollIntoView()
	fromPoint, err := elementCenter(from)
	if err != nil {
		return operationFail("", err), err
	}
	_ = to.ScrollIntoView()
	toPoint, err := elementCenter(to)
	if err != nil {
		return operationFail("", err), err
	}

	mouse := page.Mouse
	if err = mouse.MoveTo(fromPoint); err != nil {
		return operationFail("", err), err
	}
	if err = mouse.Down(proto.InputMouseButtonLeft, 1); err != nil {
		return operationFail("", err), err
	}
	if err = mouse.MoveLinear(toPoint, 15); err != nil {
		_ = mouse.Up(proto.InputMouseButtonLeft, 1)
		return operationFail("", err), err
	}
	if err = mouse.Up(proto.InputMouseButtonLeft, 1); err != nil {
		return operationFail("", err), err
	}

	e.InvalidateSnapshot()
	return operationOK("Element dragged", map[string]any{"from_identifier": fromIdentifier, "to_identifier": toIdentifier}), nil
}

// FileUpload sets local file paths on a file input. 设置文件输入框的本地文件路径。
func (e *Executor) FileUpload(ctx context.Context, identifier string, filePaths []string) (*model.BrowserExecutorOperationResult, error) {
	page, err := e.activePage()
	if err != nil {
		return operationFail("", err), err
	}
	if len(filePaths) == 0 {
		err = fmt.Errorf("file paths cannot be empty")
		return operationFail("", err), err
	}
	elem, err := e.findElement(ctx, page, identifier, defaultTimeout)
	if err != nil {
		return operationFail("", err), err
	}
	if err = elem.SetFiles(filePaths); err != nil {
		return operationFail("", err), err
	}
	e.InvalidateSnapshot()
	return operationOK("Files uploaded", map[string]any{"identifier": identifier, "file_count": len(filePaths)}), nil
}

// HandleDialog arms a handler for the next JavaScript dialog. 预置下一个 JavaScript 弹窗处理器。
func (e *Executor) HandleDialog(ctx context.Context, accept bool, text string, timeoutSeconds int) (*model.BrowserExecutorOperationResult, error) {
	page, err := e.activePage()
	if err != nil {
		return operationFail("", err), err
	}
	timeout := secondsOrDefault(timeoutSeconds, 10)
	wait, handle := page.Timeout(timeout).HandleDialog()

	go func() {
		// Handle asynchronously for REST and batch workflows. 异步处理以适配 REST 和批量流程。
		defer func() {
			_ = recover()
		}()
		_ = wait()
		_ = handle(&proto.PageHandleJavaScriptDialog{Accept: accept, PromptText: text})
	}()

	return operationOK("Dialog handler armed", map[string]any{
		"accept":  accept,
		"timeout": int(timeout.Seconds()),
	}), nil
}

// PageInfo returns active page information. PageInfo returns active page metadata.
func (e *Executor) PageInfo(ctx context.Context) (*model.BrowserExecutorOperationResult, error) {
	page, err := e.activePage()
	if err != nil {
		return operationFail("", err), err
	}
	info, err := page.Info()
	if err != nil {
		return operationFail("", err), err
	}
	data := map[string]any{"url": info.URL, "title": info.Title, "target_id": string(info.TargetID)}
	if detail, detailErr := pageDetail(page); detailErr == nil {
		for key, value := range detail {
			data[key] = value
		}
	}
	if title := strings.TrimSpace(info.Title); title != "" {
		data["page_title"] = title
	}
	return operationOK("Page info", data), nil
}

// Observe returns compact page context for LLM decisions. 返回紧凑页面上下文。
func (e *Executor) Observe(ctx context.Context, includeText bool, textLimit int) (*model.BrowserExecutorObserveResult, error) {
	status := e.Status(ctx)
	result := &model.BrowserExecutorObserveResult{
		Success:   status.Running,
		Status:    status,
		Timestamp: time.Now(),
	}
	if !status.Running {
		result.Error = status.Message
		return result, errors.New(status.Message)
	}

	if pageInfo, err := e.PageInfo(ctx); err == nil && pageInfo != nil {
		result.PageInfo = pageInfo.Data
	}
	if snapshot, err := e.Snapshot(ctx); err == nil && snapshot != nil {
		result.SnapshotText = snapshot.SnapshotText
		result.Elements = snapshot.Elements
	} else if err != nil {
		result.Error = err.Error()
	}
	if includeText {
		text, err := e.pageText(ctx, textLimit)
		if err == nil {
			result.PageText = text
		}
	}
	return result, nil
}

// ElementRefs returns compact clickable or input refs. ElementRefs returns filtered refs.
func (e *Executor) ElementRefs(ctx context.Context, category string, limit int) (*model.BrowserExecutorOperationResult, error) {
	snapshot, err := e.GetSnapshot(ctx)
	if err != nil {
		return operationFail("", err), err
	}
	items := FilterElementRefs(snapshot.Refs, category, limit)
	return operationOK("Element refs", map[string]any{"elements": items, "count": len(items), "category": category}), nil
}

// ElementInfo returns compact diagnostics for one element. ElementInfo 返回单个元素诊断信息。
func (e *Executor) ElementInfo(ctx context.Context, identifier string, attributes []string) (*model.BrowserExecutorOperationResult, error) {
	page, err := e.activePage()
	if err != nil {
		return operationFail("", err), err
	}
	elem, err := e.findElement(ctx, page, identifier, defaultTimeout)
	if err != nil {
		return operationFail("", err), err
	}
	if len(attributes) == 0 {
		attributes = []string{"id", "name", "class", "type", "role", "aria-label", "placeholder", "href", "src", "title", "alt"}
	}

	result, err := elem.Eval(`attrs => {
		const rect = this.getBoundingClientRect();
		const style = window.getComputedStyle(this);
		const attrData = {};
		(attrs || []).forEach(name => {
			const value = this.getAttribute(name);
			if (value !== null && value !== '') attrData[name] = value;
		});
		return {
			tag: this.tagName ? this.tagName.toLowerCase() : '',
			text: this.innerText || this.textContent || '',
			html: this.outerHTML || '',
			value: this.value ?? this.getAttribute('value') ?? '',
			attributes: attrData,
			visible: !!(rect.width || rect.height) && style.visibility !== 'hidden' && style.display !== 'none',
			disabled: !!this.disabled || this.getAttribute('aria-disabled') === 'true',
			readonly: !!this.readOnly || this.getAttribute('aria-readonly') === 'true',
			checked: !!this.checked,
			box: { x: rect.x, y: rect.y, width: rect.width, height: rect.height },
			selector_hint: this.id ? '#' + CSS.escape(this.id) : ''
		};
	}`, attributes)
	if err != nil {
		return operationFail("", err), err
	}

	data := map[string]any{"identifier": identifier}
	if raw := result.Value.String(); raw != "" {
		_ = json.Unmarshal([]byte(raw), &data)
		data["identifier"] = identifier
	}
	if xpath, xpathErr := elem.GetXPath(true); xpathErr == nil {
		data["xpath"] = xpath
	}
	return operationOK("Element info", data), nil
}

// PageStructure returns compact structured page data. PageStructure 返回紧凑页面结构。
func (e *Executor) PageStructure(ctx context.Context, options model.BrowserExecutorPageStructureOptions) (*model.BrowserExecutorOperationResult, error) {
	page, err := e.activePage()
	if err != nil {
		return operationFail("", err), err
	}
	if options.Limit <= 0 {
		options.Limit = 50
	}
	if !options.IncludeLinks && !options.IncludeForms && !options.IncludeTables && !options.IncludeImages && !options.IncludeButtons {
		options.IncludeLinks = true
		options.IncludeForms = true
		options.IncludeTables = true
		options.IncludeImages = true
		options.IncludeButtons = true
	}
	result, err := page.Eval(`opts => {
		const limit = Math.max(1, opts.limit || 50);
		const visible = el => {
			const rect = el.getBoundingClientRect();
			const style = window.getComputedStyle(el);
			return !!(rect.width || rect.height) && style.visibility !== 'hidden' && style.display !== 'none';
		};
		const text = el => (el.innerText || el.textContent || el.getAttribute('aria-label') || el.getAttribute('title') || '').trim().replace(/\s+/g, ' ').slice(0, 240);
		const take = list => Array.from(list).filter(visible).slice(0, limit);
		const data = {
			url: location.href,
			title: document.title,
			ready_state: document.readyState,
			headings: take(document.querySelectorAll('h1,h2,h3')).map(el => ({ level: el.tagName.toLowerCase(), text: text(el) })),
			counts: {
				links: document.querySelectorAll('a[href]').length,
				buttons: document.querySelectorAll('button,[role="button"],input[type="button"],input[type="submit"]').length,
				inputs: document.querySelectorAll('input,textarea,select,[contenteditable="true"]').length,
				forms: document.querySelectorAll('form').length,
				tables: document.querySelectorAll('table').length,
				images: document.querySelectorAll('img').length
			}
		};
		if (opts.include_links) data.links = take(document.querySelectorAll('a[href]')).map(el => ({ text: text(el), href: el.href }));
		if (opts.include_buttons) data.buttons = take(document.querySelectorAll('button,[role="button"],input[type="button"],input[type="submit"]')).map(el => ({ text: text(el), type: el.type || '', disabled: !!el.disabled }));
		if (opts.include_images) data.images = take(document.querySelectorAll('img')).map(el => ({ alt: el.alt || '', src: el.currentSrc || el.src || '', width: el.naturalWidth || 0, height: el.naturalHeight || 0 }));
		if (opts.include_forms) data.forms = take(document.querySelectorAll('form')).map(form => ({
			action: form.action || '',
			method: form.method || 'get',
			fields: Array.from(form.querySelectorAll('input,textarea,select,[contenteditable="true"]')).slice(0, limit).map(el => ({
				tag: el.tagName ? el.tagName.toLowerCase() : '',
				type: el.type || '',
				name: el.name || '',
				id: el.id || '',
				placeholder: el.getAttribute('placeholder') || '',
				label: el.getAttribute('aria-label') || '',
				value: el.value || ''
			}))
		}));
		if (opts.include_tables) data.tables = take(document.querySelectorAll('table')).map(table => ({
			caption: text(table.querySelector('caption') || table),
			headers: Array.from(table.querySelectorAll('th')).slice(0, limit).map(text),
			rows: Array.from(table.querySelectorAll('tr')).slice(0, Math.min(limit, 10)).map(row => Array.from(row.children).slice(0, 10).map(text))
		}));
		return data;
	}`, map[string]any{
		"include_links":   options.IncludeLinks,
		"include_forms":   options.IncludeForms,
		"include_tables":  options.IncludeTables,
		"include_images":  options.IncludeImages,
		"include_buttons": options.IncludeButtons,
		"limit":           options.Limit,
	})
	if err != nil {
		return operationFail("", err), err
	}
	data := map[string]any{}
	if raw := result.Value.String(); raw != "" {
		_ = json.Unmarshal([]byte(raw), &data)
	}
	return operationOK("Page structure", data), nil
}

// GetText returns element text. GetText returns element text by identifier.
func (e *Executor) GetText(ctx context.Context, identifier string) (*model.BrowserExecutorOperationResult, error) {
	page, err := e.activePage()
	if err != nil {
		return operationFail("", err), err
	}
	elem, err := e.findElement(ctx, page, identifier, defaultTimeout)
	if err != nil {
		return operationFail("", err), err
	}
	text, err := elem.Text()
	if err != nil {
		return operationFail("", err), err
	}
	return operationOK("Element text", map[string]any{"text": text}), nil
}

// GetValue returns element value. 返回元素值。
func (e *Executor) GetValue(ctx context.Context, identifier string) (*model.BrowserExecutorOperationResult, error) {
	page, err := e.activePage()
	if err != nil {
		return operationFail("", err), err
	}
	elem, err := e.findElement(ctx, page, identifier, defaultTimeout)
	if err != nil {
		return operationFail("", err), err
	}
	result, err := elem.Eval(`() => this.value ?? this.getAttribute('value') ?? this.textContent ?? ''`)
	if err != nil {
		return operationFail("", err), err
	}
	return operationOK("Element value", map[string]any{"value": result.Value.String()}), nil
}

// PageText returns visible page text with optional limit. 返回页面可见文本。
func (e *Executor) PageText(ctx context.Context, limit int) (*model.BrowserExecutorOperationResult, error) {
	text, err := e.pageText(ctx, limit)
	if err != nil {
		return operationFail("", err), err
	}
	return operationOK("Page text", map[string]any{"text": text, "length": len(text)}), nil
}

// PageContent returns current page HTML with optional limit. 返回页面 HTML。
func (e *Executor) PageContent(ctx context.Context, limit int) (*model.BrowserExecutorOperationResult, error) {
	page, err := e.activePage()
	if err != nil {
		return operationFail("", err), err
	}
	html, err := page.HTML()
	if err != nil {
		return operationFail("", err), err
	}
	html = limitString(html, limit)
	return operationOK("Page content", map[string]any{"html": html, "length": len(html)}), nil
}

// Extract extracts text/href/value from selected elements. Extract reads selected data from the page.
func (e *Executor) Extract(ctx context.Context, selector string, fields []string, multiple bool) (*model.BrowserExecutorOperationResult, error) {
	page, err := e.activePage()
	if err != nil {
		return operationFail("", err), err
	}
	if strings.TrimSpace(selector) == "" {
		result, evalErr := page.Eval(`() => document.body ? document.body.innerText : ''`)
		if evalErr != nil {
			return operationFail("", evalErr), evalErr
		}
		return operationOK("Page text", map[string]any{"text": result.Value.String()}), nil
	}
	if len(fields) == 0 {
		fields = []string{"text", "href", "value"}
	}
	script := `selector => Array.from(document.querySelectorAll(selector)).map(el => ({
		text: el.innerText || el.textContent || '',
		href: el.href || '',
		value: el.value || '',
		html: el.outerHTML || ''
	}))`
	result, err := page.Eval(script, selector)
	if err != nil {
		return operationFail("", err), err
	}
	var items []map[string]any
	_ = json.Unmarshal([]byte(result.Value.String()), &items)
	if !multiple && len(items) > 1 {
		items = items[:1]
	}
	return operationOK("Data extracted", map[string]any{"items": filterFields(items, fields)}), nil
}

// Screenshot captures a page screenshot as base64. Screenshot returns base64 image data.
func (e *Executor) Screenshot(ctx context.Context, fullPage bool, format string, quality int) (*model.BrowserExecutorOperationResult, error) {
	page, err := e.activePage()
	if err != nil {
		return operationFail("", err), err
	}
	req := &proto.PageCaptureScreenshot{Format: screenshotFormat(format)}
	if quality > 0 {
		req.Quality = &quality
	}
	data, err := page.Screenshot(fullPage, req)
	if err != nil {
		return operationFail("", err), err
	}
	return operationOK("Screenshot captured", map[string]any{
		"format":    string(req.Format),
		"size":      len(data),
		"base64":    base64.StdEncoding.EncodeToString(data),
		"full_page": fullPage,
		"mime_type": "image/" + string(req.Format),
	}), nil
}

// ElementScreenshot captures one element as base64 image data. ElementScreenshot 截取单个元素图片。
func (e *Executor) ElementScreenshot(ctx context.Context, identifier string, format string, quality int) (*model.BrowserExecutorOperationResult, error) {
	page, err := e.activePage()
	if err != nil {
		return operationFail("", err), err
	}
	elem, err := e.findElement(ctx, page, identifier, defaultTimeout)
	if err != nil {
		return operationFail("", err), err
	}
	formatValue := screenshotFormat(format)
	data, err := elem.Screenshot(formatValue, quality)
	if err != nil {
		return operationFail("", err), err
	}
	return operationOK("Element screenshot captured", map[string]any{
		"identifier": identifier,
		"format":     string(formatValue),
		"size":       len(data),
		"base64":     base64.StdEncoding.EncodeToString(data),
		"mime_type":  "image/" + string(formatValue),
	}), nil
}

// Evaluate executes JavaScript. Evaluate executes page JavaScript.
func (e *Executor) Evaluate(ctx context.Context, script string) (*model.BrowserExecutorOperationResult, error) {
	page, err := e.activePage()
	if err != nil {
		return operationFail("", err), err
	}
	result, err := page.Eval(wrapEvaluateScript(script))
	if err != nil {
		return operationFail("", err), err
	}
	return operationOK("Script evaluated", map[string]any{"result": result.Value.String()}), nil
}

// Tabs manages browser tabs. Tabs lists, creates, switches, or closes tabs.
func (e *Executor) Tabs(ctx context.Context, action string, url string, index int) (*model.BrowserExecutorOperationResult, error) {
	browser := e.runtime.Browser
	action = strings.ToLower(strings.TrimSpace(action))
	pages, err := browser.Pages()
	if err != nil {
		return operationFail("", err), err
	}
	pageTabs := make([]*rod.Page, 0)
	tabs := make([]map[string]any, 0)
	for _, page := range pages {
		info, infoErr := page.Info()
		if infoErr != nil || info == nil || info.Type != "page" {
			continue
		}
		isAgent := isBrowserAgentURL(info.URL)
		pageTabs = append(pageTabs, page)
		tabs = append(tabs, map[string]any{"index": len(pageTabs) - 1, "url": info.URL, "title": info.Title, "target_id": string(info.TargetID), "is_agent": isAgent, "is_business": !isAgent})
	}
	switch action {
	case "list":
		return operationOK("Tabs listed", map[string]any{"tabs": tabs, "count": len(tabs)}), nil
	case "new":
		if strings.TrimSpace(url) == "" {
			url = "about:blank"
		}
		page, newErr := browser.Page(proto.TargetCreateTarget{URL: url})
		if newErr != nil {
			return operationFail("", newErr), newErr
		}
		_, _ = page.Activate()
		e.InvalidateSnapshot()
		return operationOK("Tab created", map[string]any{"url": url}), nil
	case "switch":
		if index < 0 || index >= len(pageTabs) {
			err = fmt.Errorf("tab index out of range")
			return operationFail("", err), err
		}
		if tabInfo, ok := tabs[index]["url"].(string); ok && isBrowserAgentURL(tabInfo) {
			err = fmt.Errorf("refusing to switch to BrowserFlow browser-agent tab for task execution")
			return operationFail("", err), err
		}
		_, err = pageTabs[index].Activate()
		e.InvalidateSnapshot()
		return operationOK("Tab switched", map[string]any{"index": index}), err
	case "close":
		if index < 0 || index >= len(pageTabs) {
			err = fmt.Errorf("tab index out of range")
			return operationFail("", err), err
		}
		if tabInfo, ok := tabs[index]["url"].(string); ok && isBrowserAgentURL(tabInfo) {
			err = fmt.Errorf("refusing to close BrowserFlow browser-agent tab")
			return operationFail("", err), err
		}
		err = pageTabs[index].Close()
		e.InvalidateSnapshot()
		return operationOK("Tab closed", map[string]any{"index": index}), err
	default:
		err = fmt.Errorf("unknown tabs action: %s", action)
		return operationFail("", err), err
	}
}

// Scroll scrolls the page or a target element. 滚动页面或目标元素。
func (e *Executor) Scroll(ctx context.Context, direction string, pixels int, identifier string) (*model.BrowserExecutorOperationResult, error) {
	page, err := e.activePage()
	if err != nil {
		return operationFail("", err), err
	}
	direction = strings.ToLower(strings.TrimSpace(direction))
	if direction == "" {
		direction = "down"
	}
	if pixels <= 0 {
		pixels = 700
	}
	if strings.TrimSpace(identifier) != "" {
		elem, findErr := e.findElement(ctx, page, identifier, defaultTimeout)
		if findErr != nil {
			return operationFail("", findErr), findErr
		}
		if direction == "top" || direction == "up" {
			err = elem.ScrollIntoView()
		} else {
			_, err = elem.Eval(`(px) => this.scrollBy ? this.scrollBy(0, px) : this.scrollIntoView(false)`, pixels)
		}
	} else {
		script := `(direction, pixels) => {
			if (direction === 'top') window.scrollTo(0, 0);
			else if (direction === 'bottom') window.scrollTo(0, document.documentElement.scrollHeight);
			else window.scrollBy(0, direction === 'up' ? -pixels : pixels);
			return {
				scrollX: window.scrollX || window.pageXOffset || 0,
				scrollY: window.scrollY || window.pageYOffset || 0,
				scrollHeight: document.documentElement.scrollHeight,
				innerHeight: window.innerHeight
			};
		}`
		_, err = page.Eval(script, direction, pixels)
	}
	if err != nil {
		return operationFail("", err), err
	}
	e.InvalidateSnapshot()
	return operationOK("Page scrolled", map[string]any{"direction": direction, "pixels": pixels}), nil
}

// Mouse runs coordinate based mouse operations. Mouse 执行坐标鼠标操作。
func (e *Executor) Mouse(ctx context.Context, options model.BrowserExecutorMouseOptions) (*model.BrowserExecutorOperationResult, error) {
	page, err := e.activePage()
	if err != nil {
		return operationFail("", err), err
	}
	action := strings.ToLower(strings.TrimSpace(options.Action))
	button := mouseButton(options.Button)
	steps := options.Steps
	if steps <= 0 {
		steps = 10
	}
	point := proto.Point{X: options.X, Y: options.Y}
	mouse := page.Mouse
	switch action {
	case "move", "move-to":
		err = mouse.MoveLinear(point, steps)
	case "click":
		err = mouse.MoveTo(point)
		if err == nil {
			err = mouse.Click(button, 1)
		}
	case "double-click", "dblclick":
		err = mouse.MoveTo(point)
		if err == nil {
			err = mouse.Click(button, 2)
		}
	case "right-click", "context-menu":
		err = mouse.MoveTo(point)
		if err == nil {
			err = mouse.Click(proto.InputMouseButtonRight, 1)
		}
	case "down":
		err = mouse.MoveTo(point)
		if err == nil {
			err = mouse.Down(button, 1)
		}
	case "up":
		err = mouse.MoveTo(point)
		if err == nil {
			err = mouse.Up(button, 1)
		}
	case "scroll", "wheel":
		err = mouse.MoveTo(point)
		if err == nil {
			err = mouse.Scroll(options.DeltaX, options.DeltaY, steps)
		}
	default:
		err = fmt.Errorf("unknown mouse action: %s", options.Action)
	}
	if err != nil {
		return operationFail("", err), err
	}
	e.InvalidateSnapshot()
	return operationOK("Mouse operation completed", map[string]any{
		"action":  action,
		"x":       options.X,
		"y":       options.Y,
		"delta_x": options.DeltaX,
		"delta_y": options.DeltaY,
		"steps":   steps,
		"button":  string(button),
	}), nil
}

// Window manages browser window bounds and state. Window 执行浏览器窗口操作。
func (e *Executor) Window(ctx context.Context, options model.BrowserExecutorWindowOptions) (*model.BrowserExecutorOperationResult, error) {
	page, err := e.activePage()
	if err != nil {
		return operationFail("", err), err
	}
	action := strings.ToLower(strings.TrimSpace(options.Action))
	if action == "" {
		action = "info"
	}
	switch action {
	case "info", "get":
		bounds, infoErr := page.GetWindow()
		if infoErr != nil {
			return operationFail("", infoErr), infoErr
		}
		return operationOK("Window info", windowBoundsData(bounds)), nil
	case "set", "resize", "move":
		bounds := &proto.BrowserBounds{}
		if options.Left != 0 {
			bounds.Left = &options.Left
		}
		if options.Top != 0 {
			bounds.Top = &options.Top
		}
		if options.Width > 0 {
			bounds.Width = &options.Width
		}
		if options.Height > 0 {
			bounds.Height = &options.Height
		}
		err = page.SetWindow(bounds)
	case "maximize", "maximized":
		err = page.SetWindow(&proto.BrowserBounds{WindowState: proto.BrowserWindowStateMaximized})
	case "minimize", "minimized":
		err = page.SetWindow(&proto.BrowserBounds{WindowState: proto.BrowserWindowStateMinimized})
	case "fullscreen":
		err = page.SetWindow(&proto.BrowserBounds{WindowState: proto.BrowserWindowStateFullscreen})
	case "normal", "restore":
		err = page.SetWindow(&proto.BrowserBounds{WindowState: proto.BrowserWindowStateNormal})
	default:
		err = fmt.Errorf("unknown window action: %s", options.Action)
	}
	if err != nil {
		return operationFail("", err), err
	}
	bounds, _ := page.GetWindow()
	data := windowBoundsData(bounds)
	data["action"] = action
	return operationOK("Window operation completed", data), nil
}

// Act runs a smart browser action from intent. 根据意图执行智能浏览器动作。
func (e *Executor) Act(ctx context.Context, options model.BrowserExecutorActOptions) (*model.BrowserExecutorOperationResult, error) {
	intent := strings.ToLower(strings.TrimSpace(options.Intent))
	valueText := firstNonEmpty(options.Text, stringValue(options.Value))
	switch intent {
	case "navigate", "open", "goto":
		return e.Navigate(ctx, valueText, options.Timeout)
	case "click", "press":
		return e.Click(ctx, options.Identifier)
	case "type", "input", "fill":
		return e.Type(ctx, options.Identifier, valueText, options.Clear)
	case "select", "choose":
		return e.Select(ctx, options.Identifier, valueText)
	case "check":
		return e.setChecked(ctx, options.Identifier, true)
	case "uncheck":
		return e.setChecked(ctx, options.Identifier, false)
	case "fill-form", "form":
		return e.FillForm(ctx, options.Fields, options.Submit, options.Timeout)
	case "press-key", "key":
		return e.PressKey(ctx, valueText, false, false, false, false)
	case "scroll":
		return e.Scroll(ctx, valueText, 0, options.Identifier)
	case "hover":
		return e.Hover(ctx, options.Identifier)
	default:
		err := fmt.Errorf("unknown act intent: %s", options.Intent)
		return operationFail("", err), err
	}
}

// AppendObserve adds optional observe data to an operation result. 添加可选观察结果。
func (e *Executor) AppendObserve(ctx context.Context, result *model.BrowserExecutorOperationResult, enabled bool, includeText bool, textLimit int) (*model.BrowserExecutorOperationResult, error) {
	if !enabled || result == nil {
		return result, nil
	}
	observe, err := e.Observe(ctx, includeText, textLimit)
	if result.Data == nil {
		result.Data = map[string]interface{}{}
	}
	result.Data["observe"] = observe
	return result, err
}

func secondsOrDefault(seconds int, fallback int) time.Duration {
	if seconds <= 0 {
		seconds = fallback
	}
	return time.Duration(seconds) * time.Second
}

func isPageWaitState(state string) bool {
	switch state {
	case "load", "dom-stable", "request-idle", "page-stable":
		return true
	default:
		return false
	}
}

func mouseButton(button string) proto.InputMouseButton {
	switch strings.ToLower(strings.TrimSpace(button)) {
	case "right":
		return proto.InputMouseButtonRight
	case "middle":
		return proto.InputMouseButtonMiddle
	default:
		return proto.InputMouseButtonLeft
	}
}

func windowBoundsData(bounds *proto.BrowserBounds) map[string]any {
	data := map[string]any{}
	if bounds == nil {
		return data
	}
	if bounds.Left != nil {
		data["left"] = *bounds.Left
	}
	if bounds.Top != nil {
		data["top"] = *bounds.Top
	}
	if bounds.Width != nil {
		data["width"] = *bounds.Width
	}
	if bounds.Height != nil {
		data["height"] = *bounds.Height
	}
	if bounds.WindowState != "" {
		data["window_state"] = string(bounds.WindowState)
	}
	return data
}

func normalizeKey(key string) input.Key {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "enter", "return":
		return input.Enter
	case "tab":
		return input.Tab
	case "escape", "esc":
		return input.Escape
	case "backspace":
		return input.Backspace
	case "delete":
		return input.Delete
	case "up", "arrowup":
		return input.ArrowUp
	case "down", "arrowdown":
		return input.ArrowDown
	case "left", "arrowleft":
		return input.ArrowLeft
	case "right", "arrowright":
		return input.ArrowRight
	case "home":
		return input.Home
	case "end":
		return input.End
	case "pageup":
		return input.PageUp
	case "pagedown":
		return input.PageDown
	case "space":
		return input.Space
	default:
		runes := []rune(key)
		if len(runes) > 0 {
			return input.Key(runes[0])
		}
		return input.Enter
	}
}

func (e *Executor) pageText(ctx context.Context, limit int) (string, error) {
	page, err := e.activePage()
	if err != nil {
		return "", err
	}
	result, err := page.Eval(`() => document.body ? document.body.innerText : ''`)
	if err != nil {
		return "", err
	}
	return limitString(result.Value.String(), limit), nil
}

func (e *Executor) fillSingleField(ctx context.Context, page *rod.Page, field model.BrowserExecutorFormField, timeout time.Duration) error {
	name := strings.TrimSpace(field.Name)
	if name == "" {
		return fmt.Errorf("field name cannot be empty")
	}

	elem, err := e.findFormField(ctx, page, name, timeout)
	if err != nil {
		return err
	}
	tag, inputType := fieldElementType(elem)
	switch tag {
	case "select":
		return fillSelectField(elem, fmt.Sprintf("%v", field.Value))
	case "textarea":
		return fillTextField(page, elem, fmt.Sprintf("%v", field.Value))
	case "input":
		return fillInputField(page, elem, inputType, field.Value)
	default:
		return fillTextField(page, elem, fmt.Sprintf("%v", field.Value))
	}
}

func (e *Executor) findFormField(ctx context.Context, page *rod.Page, name string, timeout time.Duration) (*rod.Element, error) {
	escaped := cssQuote(name)
	selectors := []string{
		fmt.Sprintf(`input[name='%s']`, escaped),
		fmt.Sprintf(`textarea[name='%s']`, escaped),
		fmt.Sprintf(`select[name='%s']`, escaped),
		fmt.Sprintf(`input[id='%s']`, escaped),
		fmt.Sprintf(`textarea[id='%s']`, escaped),
		fmt.Sprintf(`select[id='%s']`, escaped),
		fmt.Sprintf(`input[placeholder*='%s']`, escaped),
		fmt.Sprintf(`textarea[placeholder*='%s']`, escaped),
		fmt.Sprintf(`input[aria-label*='%s']`, escaped),
		fmt.Sprintf(`textarea[aria-label*='%s']`, escaped),
	}
	for _, selector := range selectors {
		if elem, err := page.Timeout(timeout).Element(selector); err == nil {
			return elem, nil
		}
	}
	if elem, err := e.findElement(ctx, page, name, timeout); err == nil {
		return elem, nil
	}
	return findFormFieldByLabel(page, name, timeout)
}

func findFormFieldByLabel(page *rod.Page, labelText string, timeout time.Duration) (*rod.Element, error) {
	labels, err := page.Timeout(timeout).Elements("label")
	if err != nil {
		return nil, err
	}
	needle := strings.ToLower(labelText)
	for _, label := range labels {
		text, textErr := label.Text()
		if textErr != nil || !strings.Contains(strings.ToLower(text), needle) {
			continue
		}
		if forAttr, _ := label.Attribute("for"); forAttr != nil && *forAttr != "" {
			if elem, err := page.Element("#" + cssID(*forAttr)); err == nil {
				return elem, nil
			}
		}
		if elem, err := label.Element("input, textarea, select, [contenteditable='true']"); err == nil {
			return elem, nil
		}
	}
	return nil, fmt.Errorf("form field not found: %s", labelText)
}

func fieldElementType(elem *rod.Element) (string, string) {
	tag := ""
	inputType := ""
	if result, err := elem.Eval(`() => this.tagName ? this.tagName.toLowerCase() : ''`); err == nil {
		tag = strings.Trim(result.Value.String(), `"`)
	}
	if tag == "input" {
		if attr, _ := elem.Attribute("type"); attr != nil {
			inputType = strings.ToLower(strings.TrimSpace(*attr))
		}
	}
	return tag, inputType
}

func fillInputField(page *rod.Page, elem *rod.Element, inputType string, value any) error {
	switch inputType {
	case "checkbox", "radio":
		shouldCheck := truthy(value)
		current, err := elem.Property("checked")
		if err != nil {
			return err
		}
		if current.Bool() != shouldCheck {
			return elem.Click(proto.InputMouseButtonLeft, 1)
		}
		return nil
	default:
		return fillTextField(page, elem, fmt.Sprintf("%v", value))
	}
}

func fillTextField(page *rod.Page, elem *rod.Element, value string) error {
	_ = elem.ScrollIntoView()
	if err := elem.Focus(); err != nil {
		return err
	}
	if err := elem.SelectAllText(); err == nil {
		_ = page.Keyboard.Type(input.Backspace)
	}
	return elem.Input(value)
}

func fillSelectField(elem *rod.Element, value string) error {
	if err := elem.Select([]string{value}, true, rod.SelectorTypeText); err == nil {
		return nil
	}
	if err := elem.Select([]string{value}, true, rod.SelectorTypeRegex); err == nil {
		return nil
	}
	_, err := elem.Eval(`value => { this.value = value; this.dispatchEvent(new Event('input', { bubbles: true })); this.dispatchEvent(new Event('change', { bubbles: true })); }`, value)
	return err
}

func (e *Executor) submitForm(ctx context.Context, page *rod.Page) error {
	if elem, err := page.Element("button[type='submit'], input[type='submit']"); err == nil {
		return elem.Click(proto.InputMouseButtonLeft, 1)
	}
	if elem, err := page.ElementR("button", "提交|Submit|登录|Login|保存|Save|确定|OK"); err == nil {
		return elem.Click(proto.InputMouseButtonLeft, 1)
	}
	return page.Keyboard.Type(input.Enter)
}

func (e *Executor) setChecked(ctx context.Context, identifier string, checked bool) (*model.BrowserExecutorOperationResult, error) {
	page, err := e.activePage()
	if err != nil {
		return operationFail("", err), err
	}
	elem, err := e.findElement(ctx, page, identifier, defaultTimeout)
	if err != nil {
		return operationFail("", err), err
	}
	current, err := elem.Property("checked")
	if err != nil {
		return operationFail("", err), err
	}
	if current.Bool() != checked {
		if err = elem.Click(proto.InputMouseButtonLeft, 1); err != nil {
			return operationFail("", err), err
		}
	}
	e.InvalidateSnapshot()
	return operationOK("Checked state updated", map[string]any{"identifier": identifier, "checked": checked}), nil
}

func elementCenter(elem *rod.Element) (proto.Point, error) {
	shape, err := elem.Shape()
	if err != nil {
		return proto.Point{}, err
	}
	box := proto.Shape(shape.Quads).Box()
	if box == nil {
		return proto.Point{}, fmt.Errorf("element has no visible box")
	}
	return proto.Point{X: box.X + box.Width/2, Y: box.Y + box.Height/2}, nil
}

func truthy(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		switch strings.ToLower(strings.TrimSpace(typed)) {
		case "true", "1", "yes", "on", "checked":
			return true
		}
	case float64:
		return typed != 0
	case int:
		return typed != 0
	}
	return false
}

func stringValue(value any) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}

func wrapEvaluateScript(script string) string {
	script = strings.TrimSpace(script)
	if script == "" {
		return `() => null`
	}
	if strings.HasPrefix(script, "()") ||
		strings.HasPrefix(script, "async ") ||
		strings.HasPrefix(script, "function") ||
		strings.HasPrefix(script, "(") {
		return script
	}
	return "() => {\n" + script + "\n}"
}

func pageDetail(page *rod.Page) (map[string]any, error) {
	result, err := page.Eval(`() => ({
		ready_state: document.readyState,
		viewport: { width: window.innerWidth, height: window.innerHeight, device_pixel_ratio: window.devicePixelRatio },
		scroll: {
			x: window.scrollX || window.pageXOffset || 0,
			y: window.scrollY || window.pageYOffset || 0,
			width: document.documentElement.scrollWidth,
			height: document.documentElement.scrollHeight,
			can_scroll: document.documentElement.scrollHeight > window.innerHeight
		},
		counts: {
			links: document.querySelectorAll('a').length,
			buttons: document.querySelectorAll('button, [role="button"]').length,
			inputs: document.querySelectorAll('input, textarea, select, [contenteditable="true"]').length,
			forms: document.querySelectorAll('form').length,
			iframes: document.querySelectorAll('iframe').length
		},
		language: document.documentElement.lang || '',
		direction: document.documentElement.dir || 'ltr'
	})`)
	if err != nil {
		return nil, err
	}
	data := map[string]any{}
	if raw := result.Value.String(); raw != "" {
		_ = json.Unmarshal([]byte(raw), &data)
	}
	return data, nil
}

func limitString(value string, limit int) string {
	if limit <= 0 {
		limit = 8000
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}

func filterFields(items []map[string]any, fields []string) []map[string]any {
	fieldSet := map[string]bool{}
	for _, field := range fields {
		fieldSet[strings.TrimSpace(field)] = true
	}
	for _, item := range items {
		for key := range item {
			if !fieldSet[key] {
				delete(item, key)
			}
		}
	}
	return items
}
