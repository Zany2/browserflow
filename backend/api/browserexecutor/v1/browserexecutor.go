package v1

import (
	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/gogf/gf/v2/frame/g"
)

// BrowserExecutorStatusReq checks browser executor status. 检查浏览器执行器状态。
type BrowserExecutorStatusReq struct {
	g.Meta `path:"/status" method:"get" tags:"浏览器执行器" summary:"检查浏览器执行器状态"`
}

// BrowserExecutorStatusRes is the status response. 状态响应。
type BrowserExecutorStatusRes struct {
	Status *model.BrowserExecutorStatus `json:"status,omitempty" dc:"状态"`
}

// BrowserExecutorHelpReq returns command help. 返回命令帮助。
type BrowserExecutorHelpReq struct {
	g.Meta  `path:"/help" method:"get,post" tags:"浏览器执行器" summary:"获取浏览器执行器帮助"`
	Command string `json:"command" in:"query" dc:"命令名称"`
}

// BrowserExecutorHelpRes is the help response. 帮助响应。
type BrowserExecutorHelpRes struct {
	BasePath string                             `json:"base_path" dc:"接口路径"`
	Commands []model.BrowserExecutorHelpCommand `json:"commands" dc:"命令列表"`
}

// BrowserExecutorExportSkillReq exports executor Skill. 导出执行器 Skill。
type BrowserExecutorExportSkillReq struct {
	g.Meta `path:"/export/skill" method:"get" tags:"浏览器执行器" summary:"导出浏览器控制 Skill"`
}

// BrowserExecutorExportSkillRes is the export response. 导出响应。
type BrowserExecutorExportSkillRes struct{}

// BrowserExecutorNavigateReq navigates page. 页面导航。
type BrowserExecutorNavigateReq struct {
	g.Meta        `path:"/navigate" method:"post" tags:"浏览器执行器" summary:"打开 URL"`
	URL           string `json:"url" v:"required#URL不能为空" dc:"URL"`
	WaitUntil     string `json:"wait_until" dc:"等待条件"`
	Timeout       int    `json:"timeout" dc:"超时秒数"`
	ReturnObserve bool   `json:"return_observe" dc:"是否返回操作后的页面观察"`
	IncludeText   bool   `json:"include_text" dc:"是否包含页面文本"`
	TextLimit     int    `json:"text_limit" dc:"页面文本最大长度"`
}

// BrowserExecutorNavigateRes is the navigate response. 打开页面响应。
type BrowserExecutorNavigateRes struct {
	Result *model.BrowserExecutorOperationResult `json:"result,omitempty" dc:"操作结果"`
}

// BrowserExecutorSnapshotReq reads accessibility snapshot. 读取可访问性快照。
type BrowserExecutorSnapshotReq struct {
	g.Meta `path:"/snapshot" method:"get,post" tags:"浏览器执行器" summary:"获取页面快照"`
}

// BrowserExecutorSnapshotRes is the snapshot response. 快照响应。
type BrowserExecutorSnapshotRes struct {
	Result *model.BrowserExecutorSnapshotResult `json:"result,omitempty" dc:"快照结果"`
}

// BrowserExecutorClickableElementsReq returns clickable elements. 返回可点击元素。
type BrowserExecutorClickableElementsReq struct {
	g.Meta `path:"/clickable-elements" method:"get,post" tags:"浏览器执行器" summary:"获取可点击元素"`
	Limit  int `json:"limit" dc:"最大数量"`
}

// BrowserExecutorClickableElementsRes is the clickable elements response. 可点击元素响应。
type BrowserExecutorClickableElementsRes struct {
	Result *model.BrowserExecutorOperationResult `json:"result,omitempty" dc:"操作结果"`
}

// BrowserExecutorInputElementsReq returns input elements. 返回输入元素。
type BrowserExecutorInputElementsReq struct {
	g.Meta `path:"/input-elements" method:"get,post" tags:"浏览器执行器" summary:"获取输入元素"`
	Limit  int `json:"limit" dc:"最大数量"`
}

// BrowserExecutorInputElementsRes is the input elements response. 输入元素响应。
type BrowserExecutorInputElementsRes struct {
	Result *model.BrowserExecutorOperationResult `json:"result,omitempty" dc:"操作结果"`
}

// BrowserExecutorClickReq clicks element. 点击元素。
type BrowserExecutorClickReq struct {
	g.Meta        `path:"/click" method:"post" tags:"浏览器执行器" summary:"点击元素"`
	Identifier    string `json:"identifier" v:"required#元素标识不能为空" dc:"RefID/CSS/XPath/文本"`
	ReturnObserve bool   `json:"return_observe" dc:"是否返回操作后的页面观察"`
	IncludeText   bool   `json:"include_text" dc:"是否包含页面文本"`
	TextLimit     int    `json:"text_limit" dc:"页面文本最大长度"`
}

// BrowserExecutorClickRes is the click response. 点击响应。
type BrowserExecutorClickRes struct {
	Result *model.BrowserExecutorOperationResult `json:"result,omitempty" dc:"操作结果"`
}

// BrowserExecutorTypeReq types text. 输入文本。
type BrowserExecutorTypeReq struct {
	g.Meta        `path:"/type" method:"post" tags:"浏览器执行器" summary:"输入文本"`
	Identifier    string `json:"identifier" v:"required#元素标识不能为空" dc:"RefID/CSS/XPath/文本"`
	Text          string `json:"text" dc:"输入内容"`
	Clear         *bool  `json:"clear" dc:"是否清空原值"`
	ReturnObserve bool   `json:"return_observe" dc:"是否返回操作后的页面观察"`
	IncludeText   bool   `json:"include_text" dc:"是否包含页面文本"`
	TextLimit     int    `json:"text_limit" dc:"页面文本最大长度"`
}

// BrowserExecutorTypeRes is the type response. 输入响应。
type BrowserExecutorTypeRes struct {
	Result *model.BrowserExecutorOperationResult `json:"result,omitempty" dc:"操作结果"`
}

// BrowserExecutorSelectReq selects an option. 选择下拉选项。
type BrowserExecutorSelectReq struct {
	g.Meta        `path:"/select" method:"post" tags:"浏览器执行器" summary:"选择下拉选项"`
	Identifier    string `json:"identifier" v:"required#元素标识不能为空" dc:"RefID/CSS/XPath/文本"`
	Value         string `json:"value" v:"required#选项值不能为空" dc:"选项文本或 value"`
	ReturnObserve bool   `json:"return_observe" dc:"是否返回操作后的页面观察"`
	IncludeText   bool   `json:"include_text" dc:"是否包含页面文本"`
	TextLimit     int    `json:"text_limit" dc:"页面文本最大长度"`
}

// BrowserExecutorSelectRes is the select response. 选择响应。
type BrowserExecutorSelectRes struct {
	Result *model.BrowserExecutorOperationResult `json:"result,omitempty" dc:"操作结果"`
}

// BrowserExecutorPressKeyReq presses keyboard key. 按键。
type BrowserExecutorPressKeyReq struct {
	g.Meta  `path:"/press-key" method:"post" tags:"浏览器执行器" summary:"按键"`
	Key     string `json:"key" v:"required#按键不能为空" dc:"按键"`
	Ctrl    bool   `json:"ctrl" dc:"Ctrl"`
	Shift   bool   `json:"shift" dc:"Shift"`
	Alt     bool   `json:"alt" dc:"Alt"`
	MetaKey bool   `json:"meta" dc:"Meta"`
}

// BrowserExecutorPressKeyRes is the key response. 按键响应。
type BrowserExecutorPressKeyRes struct {
	Result *model.BrowserExecutorOperationResult `json:"result,omitempty" dc:"操作结果"`
}

// BrowserExecutorWaitReq waits for state. 等待状态。
type BrowserExecutorWaitReq struct {
	g.Meta     `path:"/wait" method:"post" tags:"浏览器执行器" summary:"等待"`
	Identifier string `json:"identifier" dc:"元素标识"`
	State      string `json:"state" dc:"visible/hidden/enabled/interactable/writable/stable/load/time/dom-stable/request-idle/elements-more-than"`
	Timeout    int    `json:"timeout" dc:"超时秒数"`
	Count      int    `json:"count" dc:"元素数量"`
}

// BrowserExecutorWaitRes is the wait response. 等待响应。
type BrowserExecutorWaitRes struct {
	Result *model.BrowserExecutorOperationResult `json:"result,omitempty" dc:"操作结果"`
}

// BrowserExecutorReloadReq reloads current page. 刷新当前页面。
type BrowserExecutorReloadReq struct {
	g.Meta `path:"/reload" method:"post" tags:"浏览器执行器" summary:"刷新页面"`
}

// BrowserExecutorReloadRes is the reload response. 刷新响应。
type BrowserExecutorReloadRes struct {
	Result *model.BrowserExecutorOperationResult `json:"result,omitempty" dc:"操作结果"`
}

// BrowserExecutorGoBackReq goes back in history. 后退。
type BrowserExecutorGoBackReq struct {
	g.Meta `path:"/go-back" method:"post" tags:"浏览器执行器" summary:"后退"`
}

// BrowserExecutorGoBackRes is the go back response. 后退响应。
type BrowserExecutorGoBackRes struct {
	Result *model.BrowserExecutorOperationResult `json:"result,omitempty" dc:"操作结果"`
}

// BrowserExecutorGoForwardReq goes forward in history. 前进。
type BrowserExecutorGoForwardReq struct {
	g.Meta `path:"/go-forward" method:"post" tags:"浏览器执行器" summary:"前进"`
}

// BrowserExecutorGoForwardRes is the go forward response. 前进响应。
type BrowserExecutorGoForwardRes struct {
	Result *model.BrowserExecutorOperationResult `json:"result,omitempty" dc:"操作结果"`
}

// BrowserExecutorHoverReq hovers an element. 悬停元素。
type BrowserExecutorHoverReq struct {
	g.Meta     `path:"/hover" method:"post" tags:"浏览器执行器" summary:"悬停元素"`
	Identifier string `json:"identifier" v:"required#元素标识不能为空" dc:"RefID/CSS/XPath/文本"`
}

// BrowserExecutorHoverRes is the hover response. 悬停响应。
type BrowserExecutorHoverRes struct {
	Result *model.BrowserExecutorOperationResult `json:"result,omitempty" dc:"操作结果"`
}

// BrowserExecutorResizeReq resizes viewport. 调整视口。
type BrowserExecutorResizeReq struct {
	g.Meta `path:"/resize" method:"post" tags:"浏览器执行器" summary:"调整视口"`
	Width  int `json:"width" dc:"宽度"`
	Height int `json:"height" dc:"高度"`
}

// BrowserExecutorResizeRes is the resize response. 调整视口响应。
type BrowserExecutorResizeRes struct {
	Result *model.BrowserExecutorOperationResult `json:"result,omitempty" dc:"操作结果"`
}

// BrowserExecutorPageInfoReq gets page info. 获取页面信息。
type BrowserExecutorPageInfoReq struct {
	g.Meta `path:"/page-info" method:"get,post" tags:"浏览器执行器" summary:"获取页面信息"`
}

// BrowserExecutorPageInfoRes is the page info response. 页面信息响应。
type BrowserExecutorPageInfoRes struct {
	Result *model.BrowserExecutorOperationResult `json:"result,omitempty" dc:"操作结果"`
}

// BrowserExecutorObserveReq returns compact page context. 返回紧凑页面上下文。
type BrowserExecutorObserveReq struct {
	g.Meta      `path:"/observe" method:"get,post" tags:"浏览器执行器" summary:"观察当前页面"`
	IncludeText bool `json:"include_text" dc:"是否包含页面文本"`
	TextLimit   int  `json:"text_limit" dc:"页面文本最大长度"`
}

// BrowserExecutorObserveRes is the observe response. 页面观察响应。
type BrowserExecutorObserveRes struct {
	Result *model.BrowserExecutorObserveResult `json:"result,omitempty" dc:"观察结果"`
}

// BrowserExecutorGetTextReq gets element text. 获取元素文本。
type BrowserExecutorGetTextReq struct {
	g.Meta     `path:"/get-text" method:"post" tags:"浏览器执行器" summary:"获取元素文本"`
	Identifier string `json:"identifier" v:"required#元素标识不能为空" dc:"RefID/CSS/XPath/文本"`
}

// BrowserExecutorGetTextRes is the get text response. 获取文本响应。
type BrowserExecutorGetTextRes struct {
	Result *model.BrowserExecutorOperationResult `json:"result,omitempty" dc:"操作结果"`
}

// BrowserExecutorGetValueReq gets element value. 获取元素值。
type BrowserExecutorGetValueReq struct {
	g.Meta     `path:"/get-value" method:"post" tags:"浏览器执行器" summary:"获取元素值"`
	Identifier string `json:"identifier" v:"required#元素标识不能为空" dc:"RefID/CSS/XPath/文本"`
}

// BrowserExecutorGetValueRes is the get value response. 获取值响应。
type BrowserExecutorGetValueRes struct {
	Result *model.BrowserExecutorOperationResult `json:"result,omitempty" dc:"操作结果"`
}

// BrowserExecutorElementInfoReq reads element diagnostics. Element diagnostics.
type BrowserExecutorElementInfoReq struct {
	g.Meta     `path:"/element-info" method:"post" tags:"浏览器执行器" summary:"获取元素详情"`
	Identifier string   `json:"identifier" v:"required#元素标识不能为空" dc:"RefID/CSS/XPath/文本"`
	Attributes []string `json:"attributes" dc:"要读取的属性列表"`
}

// BrowserExecutorElementInfoRes is the element diagnostics response. Element info response.
type BrowserExecutorElementInfoRes struct {
	Result *model.BrowserExecutorOperationResult `json:"result,omitempty" dc:"操作结果"`
}

// BrowserExecutorPageTextReq gets page visible text. 获取页面可见文本。
type BrowserExecutorPageTextReq struct {
	g.Meta `path:"/page-text" method:"get,post" tags:"浏览器执行器" summary:"获取页面文本"`
	Limit  int `json:"limit" dc:"最大长度"`
}

// BrowserExecutorPageTextRes is the page text response. 页面文本响应。
type BrowserExecutorPageTextRes struct {
	Result *model.BrowserExecutorOperationResult `json:"result,omitempty" dc:"操作结果"`
}

// BrowserExecutorPageContentReq gets page HTML. 获取页面 HTML。
type BrowserExecutorPageContentReq struct {
	g.Meta `path:"/page-content" method:"get,post" tags:"浏览器执行器" summary:"获取页面 HTML"`
	Limit  int `json:"limit" dc:"最大长度"`
}

// BrowserExecutorPageContentRes is the page content response. 页面内容响应。
type BrowserExecutorPageContentRes struct {
	Result *model.BrowserExecutorOperationResult `json:"result,omitempty" dc:"操作结果"`
}

// BrowserExecutorPageStructureReq gets compact structured page data. Page structure request.
type BrowserExecutorPageStructureReq struct {
	g.Meta         `path:"/page-structure" method:"get,post" tags:"浏览器执行器" summary:"获取页面结构"`
	IncludeLinks   bool `json:"include_links" dc:"是否包含链接"`
	IncludeForms   bool `json:"include_forms" dc:"是否包含表单"`
	IncludeTables  bool `json:"include_tables" dc:"是否包含表格"`
	IncludeImages  bool `json:"include_images" dc:"是否包含图片"`
	IncludeButtons bool `json:"include_buttons" dc:"是否包含按钮"`
	Limit          int  `json:"limit" dc:"每类最大数量"`
}

// BrowserExecutorPageStructureRes is the page structure response. Page structure response.
type BrowserExecutorPageStructureRes struct {
	Result *model.BrowserExecutorOperationResult `json:"result,omitempty" dc:"操作结果"`
}

// BrowserExecutorExtractReq extracts page data. 提取页面数据。
type BrowserExecutorExtractReq struct {
	g.Meta   `path:"/extract" method:"post" tags:"浏览器执行器" summary:"提取页面数据"`
	Selector string   `json:"selector" dc:"CSS 选择器"`
	Fields   []string `json:"fields" dc:"字段列表"`
	Multiple bool     `json:"multiple" dc:"是否多元素"`
}

// BrowserExecutorExtractRes is the extract response. 提取响应。
type BrowserExecutorExtractRes struct {
	Result *model.BrowserExecutorOperationResult `json:"result,omitempty" dc:"操作结果"`
}

// BrowserExecutorScreenshotReq captures screenshot. 截图。
type BrowserExecutorScreenshotReq struct {
	g.Meta   `path:"/screenshot" method:"post" tags:"浏览器执行器" summary:"截图"`
	FullPage bool   `json:"full_page" dc:"是否全页"`
	Format   string `json:"format" dc:"png/jpeg"`
	Quality  int    `json:"quality" dc:"质量"`
}

// BrowserExecutorScreenshotRes is the screenshot response. 截图响应。
type BrowserExecutorScreenshotRes struct {
	Result *model.BrowserExecutorOperationResult `json:"result,omitempty" dc:"操作结果"`
}

// BrowserExecutorElementScreenshotReq captures element screenshot. Element screenshot request.
type BrowserExecutorElementScreenshotReq struct {
	g.Meta     `path:"/element-screenshot" method:"post" tags:"浏览器执行器" summary:"截取元素图片"`
	Identifier string `json:"identifier" v:"required#元素标识不能为空" dc:"RefID/CSS/XPath/文本"`
	Format     string `json:"format" dc:"png/jpeg"`
	Quality    int    `json:"quality" dc:"图片质量"`
}

// BrowserExecutorElementScreenshotRes is the element screenshot response. Element screenshot response.
type BrowserExecutorElementScreenshotRes struct {
	Result *model.BrowserExecutorOperationResult `json:"result,omitempty" dc:"操作结果"`
}

// BrowserExecutorEvaluateReq evaluates JavaScript. 执行 JavaScript。
type BrowserExecutorEvaluateReq struct {
	g.Meta `path:"/evaluate" method:"post" tags:"浏览器执行器" summary:"执行 JavaScript"`
	Script string `json:"script" v:"required#脚本不能为空" dc:"JavaScript 脚本"`
}

// BrowserExecutorEvaluateRes is the evaluate response. 执行脚本响应。
type BrowserExecutorEvaluateRes struct {
	Result *model.BrowserExecutorOperationResult `json:"result,omitempty" dc:"操作结果"`
}

// BrowserExecutorTabsReq manages tabs. 管理标签页。
type BrowserExecutorTabsReq struct {
	g.Meta `path:"/tabs" method:"post" tags:"浏览器执行器" summary:"管理标签页"`
	Action string `json:"action" v:"required#操作不能为空" dc:"list/new/switch/close"`
	URL    string `json:"url" dc:"新标签 URL"`
	Index  int    `json:"index" dc:"标签索引"`
}

// BrowserExecutorTabsRes is the tabs response. 标签页响应。
type BrowserExecutorTabsRes struct {
	Result *model.BrowserExecutorOperationResult `json:"result,omitempty" dc:"操作结果"`
}

// BrowserExecutorScrollReq scrolls the page or an element. 滚动页面或元素。
type BrowserExecutorScrollReq struct {
	g.Meta        `path:"/scroll" method:"post" tags:"浏览器执行器" summary:"滚动页面"`
	Direction     string `json:"direction" dc:"down/up/top/bottom"`
	Pixels        int    `json:"pixels" dc:"滚动像素"`
	Identifier    string `json:"identifier" dc:"可选元素标识"`
	ReturnObserve bool   `json:"return_observe" dc:"是否返回操作后的页面观察"`
	IncludeText   bool   `json:"include_text" dc:"是否包含页面文本"`
	TextLimit     int    `json:"text_limit" dc:"页面文本最大长度"`
}

// BrowserExecutorScrollRes is the scroll response. 滚动响应。
type BrowserExecutorScrollRes struct {
	Result *model.BrowserExecutorOperationResult `json:"result,omitempty" dc:"操作结果"`
}

// BrowserExecutorMouseReq runs coordinate mouse operations. Mouse operation request.
type BrowserExecutorMouseReq struct {
	g.Meta `path:"/mouse" method:"post" tags:"浏览器执行器" summary:"鼠标操作"`
	Action string  `json:"action" v:"required#鼠标动作不能为空" dc:"move/click/double-click/right-click/down/up/scroll"`
	X      float64 `json:"x" dc:"X 坐标"`
	Y      float64 `json:"y" dc:"Y 坐标"`
	DeltaX float64 `json:"delta_x" dc:"横向滚动距离"`
	DeltaY float64 `json:"delta_y" dc:"纵向滚动距离"`
	Steps  int     `json:"steps" dc:"移动或滚动步数"`
	Button string  `json:"button" dc:"left/right/middle"`
}

// BrowserExecutorMouseRes is the mouse operation response. Mouse operation response.
type BrowserExecutorMouseRes struct {
	Result *model.BrowserExecutorOperationResult `json:"result,omitempty" dc:"操作结果"`
}

// BrowserExecutorWindowReq manages browser window bounds and state. Window operation request.
type BrowserExecutorWindowReq struct {
	g.Meta `path:"/window" method:"post" tags:"浏览器执行器" summary:"窗口操作"`
	Action string `json:"action" v:"required#窗口动作不能为空" dc:"info/set/maximize/minimize/fullscreen/normal"`
	Left   int    `json:"left" dc:"窗口左侧位置"`
	Top    int    `json:"top" dc:"窗口顶部位置"`
	Width  int    `json:"width" dc:"窗口宽度"`
	Height int    `json:"height" dc:"窗口高度"`
}

// BrowserExecutorWindowRes is the window operation response. Window operation response.
type BrowserExecutorWindowRes struct {
	Result *model.BrowserExecutorOperationResult `json:"result,omitempty" dc:"操作结果"`
}

// BrowserExecutorClosePageReq closes current page. 关闭当前页面。
type BrowserExecutorClosePageReq struct {
	g.Meta `path:"/close-page" method:"post" tags:"浏览器执行器" summary:"关闭当前页面"`
}

// BrowserExecutorClosePageRes is the close page response. 关闭页面响应。
type BrowserExecutorClosePageRes struct {
	Result *model.BrowserExecutorOperationResult `json:"result,omitempty" dc:"操作结果"`
}

// BrowserExecutorFillFormReq fills multiple form fields. 批量填写表单字段。
type BrowserExecutorFillFormReq struct {
	g.Meta `path:"/fill-form" method:"post" tags:"浏览器执行器" summary:"批量填写表单"`
	Fields []struct {
		Name  string `json:"name" v:"required#字段名称不能为空" dc:"字段名称"`
		Value any    `json:"value" dc:"字段值"`
		Type  string `json:"type" dc:"字段类型"`
	} `json:"fields" v:"required|min-length:1#字段列表不能为空|至少需要一个字段" dc:"字段列表"`
	Submit        bool `json:"submit" dc:"是否提交表单"`
	Timeout       int  `json:"timeout" dc:"单字段超时秒数"`
	ReturnObserve bool `json:"return_observe" dc:"是否返回操作后的页面观察"`
	IncludeText   bool `json:"include_text" dc:"是否包含页面文本"`
	TextLimit     int  `json:"text_limit" dc:"页面文本最大长度"`
}

// BrowserExecutorFillFormRes is the fill form response. 表单填写响应。
type BrowserExecutorFillFormRes struct {
	Result *model.BrowserExecutorOperationResult `json:"result,omitempty" dc:"操作结果"`
}

// BrowserExecutorDragReq drags one element to another. 拖拽元素到目标元素。
type BrowserExecutorDragReq struct {
	g.Meta         `path:"/drag" method:"post" tags:"浏览器执行器" summary:"拖拽元素"`
	FromIdentifier string `json:"from_identifier" v:"required#源元素不能为空" dc:"源元素标识"`
	ToIdentifier   string `json:"to_identifier" v:"required#目标元素不能为空" dc:"目标元素标识"`
}

// BrowserExecutorDragRes is the drag response. 拖拽响应。
type BrowserExecutorDragRes struct {
	Result *model.BrowserExecutorOperationResult `json:"result,omitempty" dc:"操作结果"`
}

// BrowserExecutorFileUploadReq uploads files to file input. 上传文件到文件输入框。
type BrowserExecutorFileUploadReq struct {
	g.Meta     `path:"/file-upload" method:"post" tags:"浏览器执行器" summary:"上传文件"`
	Identifier string   `json:"identifier" v:"required#文件输入元素不能为空" dc:"文件输入元素标识"`
	FilePaths  []string `json:"file_paths" v:"required|min-length:1#文件路径不能为空|至少需要一个文件" dc:"文件路径列表"`
}

// BrowserExecutorFileUploadRes is the file upload response. 文件上传响应。
type BrowserExecutorFileUploadRes struct {
	Result *model.BrowserExecutorOperationResult `json:"result,omitempty" dc:"操作结果"`
}

// BrowserExecutorHandleDialogReq handles next JavaScript dialog. 处理下一个 JavaScript 弹窗。
type BrowserExecutorHandleDialogReq struct {
	g.Meta  `path:"/handle-dialog" method:"post" tags:"浏览器执行器" summary:"处理 JavaScript 弹窗"`
	Accept  bool   `json:"accept" dc:"是否接受弹窗"`
	Text    string `json:"text" dc:"Prompt 输入文本"`
	Timeout int    `json:"timeout" dc:"等待弹窗超时秒数"`
}

// BrowserExecutorHandleDialogRes is the dialog response. 弹窗处理响应。
type BrowserExecutorHandleDialogRes struct {
	Result *model.BrowserExecutorOperationResult `json:"result,omitempty" dc:"操作结果"`
}

// BrowserExecutorActReq runs a smart browser action. 执行智能浏览器动作。
type BrowserExecutorActReq struct {
	g.Meta     `path:"/act" method:"post" tags:"浏览器执行器" summary:"执行智能浏览器动作"`
	Intent     string `json:"intent" v:"required#动作意图不能为空" dc:"click/type/select/check/uncheck/fill-form/navigate/press-key/scroll"`
	Identifier string `json:"identifier" dc:"元素标识"`
	Value      any    `json:"value" dc:"动作值"`
	Text       string `json:"text" dc:"输入文本"`
	Fields     []struct {
		Name  string `json:"name" v:"required#字段名称不能为空" dc:"字段名称"`
		Value any    `json:"value" dc:"字段值"`
		Type  string `json:"type" dc:"字段类型"`
	} `json:"fields" dc:"表单字段"`
	Submit        bool  `json:"submit" dc:"是否提交表单"`
	Clear         *bool `json:"clear" dc:"是否清空原值"`
	Timeout       int   `json:"timeout" dc:"超时秒数"`
	ReturnObserve bool  `json:"return_observe" dc:"是否返回操作后的页面观察"`
	IncludeText   bool  `json:"include_text" dc:"是否包含页面文本"`
	TextLimit     int   `json:"text_limit" dc:"页面文本最大长度"`
}

// BrowserExecutorActRes is the smart action response. 智能动作响应。
type BrowserExecutorActRes struct {
	Result *model.BrowserExecutorOperationResult `json:"result,omitempty" dc:"操作结果"`
}

// BrowserExecutorBatchReq runs batch operations. 执行批量操作。
type BrowserExecutorBatchReq struct {
	g.Meta     `path:"/batch" method:"post" tags:"浏览器执行器" summary:"批量执行操作"`
	Operations []struct {
		Type        string         `json:"type" dc:"操作类型"`
		Params      map[string]any `json:"params" dc:"操作参数"`
		StopOnError bool           `json:"stop_on_error" dc:"失败后停止"`
	} `json:"operations" v:"required|min-length:1#操作列表不能为空|至少需要一个操作" dc:"操作列表"`
}

// BrowserExecutorBatchRes is the batch response. 批量响应。
type BrowserExecutorBatchRes struct {
	Result *model.BrowserExecutorOperationResult `json:"result,omitempty" dc:"操作结果"`
}
