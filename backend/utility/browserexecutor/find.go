package browserexecutor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// findElement resolves RefID/CSS/XPath/text to a Rod element. findElement resolves RefID/CSS/XPath/text.
func (e *Executor) findElement(ctx context.Context, page *rod.Page, identifier string, timeout time.Duration) (*rod.Element, error) {
	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		return nil, fmt.Errorf("element identifier cannot be empty")
	}

	if strings.HasPrefix(strings.ToLower(identifier), "css:") {
		identifier = strings.TrimSpace(identifier[4:])
	}
	if strings.HasPrefix(strings.ToLower(identifier), "xpath:") {
		identifier = strings.TrimSpace(identifier[6:])
	}

	if isRefID(identifier) {
		if elem, err := e.findByRefID(ctx, page, strings.TrimPrefix(identifier, "@")); err == nil {
			return elem, nil
		}
	}

	timeoutPage := page.Timeout(timeout)
	if elem, err := timeoutPage.Element(identifier); err == nil {
		return elem, nil
	}
	if strings.HasPrefix(identifier, "/") || strings.HasPrefix(identifier, "(") {
		if elem, err := timeoutPage.ElementX(identifier); err == nil {
			return elem, nil
		}
	}
	if elem, err := timeoutPage.ElementR("button", identifier); err == nil {
		return elem, nil
	}
	if elem, err := timeoutPage.ElementR("a", identifier); err == nil {
		return elem, nil
	}
	if elem, err := timeoutPage.Element(fmt.Sprintf("[aria-label*='%s']", cssQuote(identifier))); err == nil {
		return elem, nil
	}
	if elem, err := timeoutPage.Element(fmt.Sprintf("[placeholder*='%s']", cssQuote(identifier))); err == nil {
		return elem, nil
	}
	return nil, fmt.Errorf("element not found: %s", identifier)
}

func (e *Executor) findByRefID(ctx context.Context, page *rod.Page, refID string) (*rod.Element, error) {
	e.refMu.RLock()
	refData := e.refMap[refID]
	e.refMu.RUnlock()
	if refData == nil {
		_, _ = e.GetSnapshot(ctx)
		e.refMu.RLock()
		refData = e.refMap[refID]
		e.refMu.RUnlock()
	}
	if refData == nil {
		return nil, fmt.Errorf("RefID %s not found, call snapshot again", refID)
	}

	if refData.BackendID > 0 {
		if elem, err := e.findByBackendNodeID(page, refData.BackendID); err == nil {
			return elem, nil
		}
	}
	if refData.Href != "" {
		if elem, err := page.ElementX(fmt.Sprintf("//a[@href='%s']", xpathQuote(refData.Href))); err == nil {
			return elem, nil
		}
	}
	if refData.Attributes != nil && refData.Attributes["id"] != "" {
		if elem, err := page.Element("#" + cssID(refData.Attributes["id"])); err == nil {
			return elem, nil
		}
	}

	xpath := xpathForRef(refData)
	elements, err := page.ElementsX(xpath)
	if err != nil {
		return nil, err
	}
	if len(elements) == 0 {
		return nil, fmt.Errorf("element for RefID %s no longer exists, call snapshot again", refID)
	}
	if refData.Nth >= len(elements) {
		return elements[0], nil
	}
	return elements[refData.Nth], nil
}

func (e *Executor) findByBackendNodeID(page *rod.Page, backendID int) (*rod.Element, error) {
	res, err := (proto.DOMResolveNode{BackendNodeID: proto.DOMBackendNodeID(backendID)}).Call(page)
	if err != nil {
		return nil, err
	}
	if res == nil || res.Object == nil || res.Object.ObjectID == "" {
		return nil, fmt.Errorf("BackendNodeID %d cannot be resolved", backendID)
	}
	return page.ElementFromObject(res.Object)
}

func isRefID(value string) bool {
	value = strings.TrimPrefix(strings.TrimSpace(value), "@")
	if len(value) < 2 || len(value) > 12 || value[0] != 'e' {
		return false
	}
	for _, ch := range value[1:] {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

func xpathForRef(ref *model.BrowserExecutorRefData) string {
	name := xpathQuote(firstNonEmpty(ref.Name, ref.Placeholder))
	switch strings.ToLower(ref.Role) {
	case "button":
		return fmt.Sprintf(`//button[contains(normalize-space(.), '%s') or @aria-label='%s'] | //*[@role='button' and contains(normalize-space(.), '%s')]`, name, name, name)
	case "link":
		return fmt.Sprintf(`//a[contains(normalize-space(.), '%s') or @aria-label='%s'] | //*[@role='link' and contains(normalize-space(.), '%s')]`, name, name, name)
	case "textbox", "searchbox":
		return fmt.Sprintf(`//input[contains(@placeholder, '%s') or @aria-label='%s' or @name='%s'] | //textarea[contains(@placeholder, '%s') or @aria-label='%s']`, name, name, name, name, name)
	default:
		return fmt.Sprintf(`//*[contains(normalize-space(.), '%s') or @aria-label='%s' or contains(@placeholder, '%s')]`, name, name, name)
	}
}

func cssID(value string) string {
	return strings.ReplaceAll(value, `"`, `\"`)
}

func cssQuote(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `'`, `\'`)
	return value
}

func xpathQuote(value string) string {
	return strings.ReplaceAll(value, `'`, `&apos;`)
}
