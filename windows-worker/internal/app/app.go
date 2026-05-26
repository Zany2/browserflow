//go:build windows

package app

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"github.com/Zany2/browserflow/windows-worker/internal/automa"
	"github.com/Zany2/browserflow/windows-worker/internal/chrome"
	"github.com/Zany2/browserflow/windows-worker/internal/config"
	"github.com/Zany2/browserflow/windows-worker/internal/worker"
)

const (
	windowClassName = "BrowserFlowWorkerWindow"
	windowTitle     = "BrowserFlowWorker"

	idServerEdit   = 101
	idChromeEdit   = 102
	idNodeEdit     = 103
	idStatusText   = 104
	idAutomaText   = 105
	idDownloadEdit = 106

	idDetectButton         = 201
	idBrowseButton         = 202
	idSaveButton           = 203
	idStartButton          = 204
	idCloseButton          = 205
	idAutomaButton         = 206
	idBrowseDownloadButton = 207
)

var (
	user32    = syscall.NewLazyDLL("user32.dll")
	kernel32  = syscall.NewLazyDLL("kernel32.dll")
	gdi32     = syscall.NewLazyDLL("gdi32.dll")
	comdlg32  = syscall.NewLazyDLL("comdlg32.dll")
	shell32   = syscall.NewLazyDLL("shell32.dll")
	ole32     = syscall.NewLazyDLL("ole32.dll")
	procCache = map[string]*syscall.LazyProc{}
)

type windowState struct {
	hwnd         uintptr
	serverEdit   uintptr
	chromeEdit   uintptr
	nodeEdit     uintptr
	downloadEdit uintptr
	statusText   uintptr
	automaText   uintptr
	font         uintptr
	icon         uintptr
	exeDir       string
	configPath   string
	config       config.Config
	nodes        []worker.Node
	stopping     bool
}

// Run starts the Windows GUI.
func Run() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if hr, _, _ := ole32.NewProc("OleInitialize").Call(0); int32(hr) >= 0 {
		defer ole32.NewProc("OleUninitialize").Call()
	}

	exeDir, err := executableDir()
	if err != nil {
		return err
	}
	configPath, err := config.DefaultConfigPath()
	if err != nil {
		return err
	}
	cfg, _, err := config.LoadOrCreate(configPath)
	if err != nil {
		return err
	}

	state := &windowState{
		exeDir:     exeDir,
		configPath: configPath,
		config:     cfg,
	}
	return state.run()
}

func (s *windowState) run() error {
	hInstance := getModuleHandle()
	className := utf16Ptr(windowClassName)
	icon := loadResourceIcon(32)
	if icon == 0 {
		icon = loadIcon(filepath.Join(s.exeDir, "assets", "favicon.ico"))
	}
	if icon == 0 {
		icon = loadDefaultIcon()
	}
	s.icon = icon

	wndClass := wndclassex{
		cbSize:        uint32(unsafe.Sizeof(wndclassex{})),
		style:         3,
		lpfnWndProc:   syscall.NewCallback(wndProc),
		hInstance:     hInstance,
		hIcon:         icon,
		hCursor:       loadDefaultCursor(),
		hbrBackground: uintptr(16),
		lpszClassName: uintptr(unsafe.Pointer(className)),
		hIconSm:       icon,
	}
	if ret, _, err := proc("RegisterClassExW").Call(uintptr(unsafe.Pointer(&wndClass))); ret == 0 {
		return fmt.Errorf("RegisterClassExW failed: %v", err)
	}

	hwnd, _, err := proc("CreateWindowExW").Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(utf16Ptr(windowTitle))),
		uintptr(wsOverlappedWindow&^wsMaximizeBox&^wsSizeBox),
		uintptr(cwUseDefault),
		uintptr(cwUseDefault),
		840,
		570,
		0,
		0,
		hInstance,
		uintptr(unsafe.Pointer(s)),
	)
	if hwnd == 0 {
		return fmt.Errorf("CreateWindowExW failed: %v", err)
	}
	setWindowText(hwnd, windowTitle)

	proc("ShowWindow").Call(hwnd, 1)
	proc("UpdateWindow").Call(hwnd)

	var msg msg
	for {
		ret, _, _ := proc("GetMessageW").Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(ret) == -1 {
			return fmt.Errorf("GetMessageW failed")
		}
		if ret == 0 {
			return nil
		}
		proc("TranslateMessage").Call(uintptr(unsafe.Pointer(&msg)))
		proc("DispatchMessageW").Call(uintptr(unsafe.Pointer(&msg)))
	}
}

func (s *windowState) initControls() {
	setWindowText(s.hwnd, windowTitle)
	s.font = createUIFont()
	if s.font == 0 {
		s.font = getDefaultFont()
	}
	smallIcon := loadResourceIcon(16)
	if smallIcon == 0 {
		smallIcon = s.icon
	}
	sendMessage(s.hwnd, wmSetIcon, iconSmall, smallIcon)
	sendMessage(s.hwnd, wmSetIcon, iconBig, s.icon)

	createLabel(s.hwnd, "BrowserFlow Windows Worker", 32, 26, 420, 30)
	createLabel(s.hwnd, "配置这台电脑要启动的独立 Chrome 执行节点。", 32, 58, 720, 24)

	createLabel(s.hwnd, "服务端地址", 32, 106, 110, 26)
	s.serverEdit = createEdit(s.hwnd, idServerEdit, s.config.ServerURL, 150, 102, 620, 32)

	createLabel(s.hwnd, "Chrome 路径", 32, 154, 110, 26)
	s.chromeEdit = createEdit(s.hwnd, idChromeEdit, s.config.ChromePath, 150, 150, 460, 32)
	createButton(s.hwnd, idDetectButton, "自动检测", 624, 149, 92, 34)
	createButton(s.hwnd, idBrowseButton, "浏览...", 728, 149, 72, 34)

	createLabel(s.hwnd, "执行节点数", 32, 202, 110, 26)
	nodeText := strconv.Itoa(s.config.NodeCount)
	s.nodeEdit = createEdit(s.hwnd, idNodeEdit, nodeText, 150, 198, 86, 32)
	createLabel(s.hwnd, "每个节点会启动一个独立 Chrome 窗口，建议按电脑性能逐步增加。", 252, 202, 540, 26)

	createLabel(s.hwnd, "下载目录", 32, 250, 110, 26)
	s.downloadEdit = createEdit(s.hwnd, idDownloadEdit, s.config.DownloadDir, 150, 246, 566, 32)
	createButton(s.hwnd, idBrowseDownloadButton, "浏览...", 728, 245, 72, 34)

	createButton(s.hwnd, idSaveButton, "保存配置", 150, 312, 112, 36)
	createButton(s.hwnd, idStartButton, "启动执行节点", 278, 312, 148, 36)
	createButton(s.hwnd, idCloseButton, "关闭所有执行节点", 442, 312, 160, 36)
	createButton(s.hwnd, idAutomaButton, "安装/更新 Automa", 618, 312, 160, 36)

	createLabel(s.hwnd, "状态", 32, 390, 110, 26)
	s.statusText = createLabel(s.hwnd, "就绪", 150, 390, 620, 60)
	createLabel(s.hwnd, "Automa 插件", 32, 475, 110, 26)
	s.automaText = createLabel(s.hwnd, s.automaStatus(), 150, 475, 620, 44)
	s.applyFont(s.hwnd)
}

func (s *windowState) applyFont(hwnd uintptr) {
	children := []uintptr{
		s.serverEdit,
		s.chromeEdit,
		s.nodeEdit,
		s.downloadEdit,
		s.statusText,
		s.automaText,
	}
	for _, child := range children {
		sendMessage(child, wmSetFont, s.font, 1)
	}
	proc("EnumChildWindows").Call(hwnd, syscall.NewCallback(func(child uintptr, lparam uintptr) uintptr {
		sendMessage(child, wmSetFont, s.font, 1)
		return 1
	}), 0)
}

func (s *windowState) handleCommand(id int) {
	switch id {
	case idDetectButton:
		s.detectChrome()
	case idBrowseButton:
		s.browseChrome()
	case idBrowseDownloadButton:
		s.browseDownloadDir()
	case idSaveButton:
		s.save()
	case idStartButton:
		s.startNodes()
	case idCloseButton:
		s.stopNodes()
	case idAutomaButton:
		s.installAutoma()
	}
}

func (s *windowState) detectChrome() {
	path, err := chrome.Detect()
	if err != nil {
		s.setStatus("未检测到 Chrome，请点击“浏览...”手动选择 Chrome 可执行文件")
		messageBox(s.hwnd, "未检测到 Chrome，请手动选择 Chrome 可执行文件。")
		return
	}
	setWindowText(s.chromeEdit, path)
	s.setStatus("已检测到 Chrome：" + path)
}

func (s *windowState) browseChrome() {
	path, ok := openFileDialog(s.hwnd, "选择 Chrome 可执行文件", []fileFilter{
		{Name: "可执行文件", Pattern: "*.exe"},
		{Name: "所有文件", Pattern: "*.*"},
	}, "exe")
	if !ok {
		return
	}
	if err := chrome.ValidatePath(path); err != nil {
		s.setStatus("选择的 Chrome 不可用：" + err.Error())
		messageBox(s.hwnd, "选择的 Chrome 不可用：\n"+err.Error())
		return
	}
	setWindowText(s.chromeEdit, path)
	s.setStatus("已选择 Chrome：" + path)
}

func (s *windowState) browseDownloadDir() {
	path, ok := openFolderDialog(s.hwnd, "选择下载目录")
	if !ok {
		return
	}
	setWindowText(s.downloadEdit, path)
	s.setStatus("已选择下载目录：" + path)
}

func (s *windowState) installAutoma() {
	zipPath, ok := openFileDialog(s.hwnd, "选择 Automa 安装包", []fileFilter{
		{Name: "Automa 安装包", Pattern: "*.zip"},
		{Name: "所有文件", Pattern: "*.*"},
	}, "zip")
	if !ok {
		return
	}
	targetDir := config.DefaultAutomaExtensionDir()
	s.setStatus("正在安装 Automa，请稍候...")
	if err := automa.InstallFromZip(zipPath, targetDir); err != nil {
		s.setStatus("Automa 安装失败：" + err.Error())
		messageBox(s.hwnd, "Automa 安装失败：\n"+err.Error())
		return
	}
	s.config.AutomaExtensionDir = targetDir
	_ = config.Save(s.configPath, s.config)
	setWindowText(s.automaText, s.automaStatus())
	s.setStatus("Automa 已安装/更新：" + targetDir)
	messageBox(s.hwnd, "Automa 已安装/更新完成。\n\n下次启动执行节点时会自动加载。")
}

func (s *windowState) save() bool {
	cfg, err := s.readConfigFromForm()
	if err != nil {
		s.setStatus(err.Error())
		messageBox(s.hwnd, err.Error())
		return false
	}
	if err = config.Save(s.configPath, cfg); err != nil {
		s.setStatus("保存失败：" + err.Error())
		messageBox(s.hwnd, "保存失败：\n"+err.Error())
		return false
	}
	s.config = cfg
	setWindowText(s.automaText, s.automaStatus())
	s.setStatus("配置已保存：" + s.configPath)
	return true
}

func (s *windowState) startNodes() {
	cfg, err := s.readConfigFromForm()
	if err != nil {
		s.setStatus(err.Error())
		messageBox(s.hwnd, err.Error())
		return
	}
	if err = cfg.Validate(); err != nil {
		s.setStatus(err.Error())
		messageBox(s.hwnd, err.Error())
		return
	}

	chromePath := strings.TrimSpace(cfg.ChromePath)
	if chromePath == "" {
		chromePath, err = chrome.Detect()
		if err != nil {
			s.setStatus("未检测到 Chrome，请手动选择 Chrome 可执行文件")
			messageBox(s.hwnd, "未检测到 Chrome，请手动选择 Chrome 可执行文件。")
			return
		}
		setWindowText(s.chromeEdit, chromePath)
		cfg.ChromePath = chromePath
	}
	if err = chrome.Validate(chromePath); err != nil {
		s.setStatus("Chrome 路径不可用：" + err.Error())
		messageBox(s.hwnd, "Chrome 路径不可用：\n"+err.Error())
		return
	}

	dataDir := resolvePath(s.exeDir, cfg.DataDir, config.DefaultDataDir())
	downloadDir := resolvePath(s.exeDir, cfg.DownloadDir, config.DefaultDownloadDir())
	extensionDir := resolvePath(s.exeDir, cfg.AutomaExtensionDir, config.DefaultAutomaExtensionDir())
	nodes, err := worker.Launcher{
		ChromePath:          chromePath,
		ServerURL:           cfg.ServerURL,
		MachineID:           cfg.MachineID,
		MachineName:         cfg.MachineName,
		NodeCount:           cfg.NodeCount,
		DataDir:             dataDir,
		DownloadDir:         downloadDir,
		AutomaExtensionDir:  extensionDir,
		RequireAutomaFolder: cfg.RequireAutomaFolder,
	}.Start()
	if err != nil {
		s.setStatus("启动失败：" + err.Error())
		messageBox(s.hwnd, "启动失败：\n"+err.Error())
		return
	}

	_ = config.Save(s.configPath, cfg)
	s.config = cfg
	s.nodes = append(s.nodes, nodes...)
	s.setStatus(fmt.Sprintf("已启动 %d 个执行节点。服务端：%s", len(nodes), cfg.ServerURL))
}

func (s *windowState) stopNodes() {
	if len(s.nodes) == 0 {
		s.setStatus("没有需要关闭的执行节点")
		return
	}
	if s.stopping {
		s.setStatus("正在关闭执行节点，请稍候...")
		return
	}
	nodes := append([]worker.Node(nil), s.nodes...)
	s.nodes = nil
	s.stopping = true
	s.setStatus("正在关闭执行节点，请稍候...")
	go func(hwnd uintptr, nodes []worker.Node) {
		stopped := worker.Stop(nodes)
		postMessage(hwnd, wmWorkerStopped, uintptr(stopped), 0)
	}(s.hwnd, nodes)
}

func (s *windowState) confirmClose() bool {
	if len(s.nodes) == 0 {
		return true
	}
	if confirmBox(s.hwnd, "当前已经启动了执行节点。\n\n是否同时关闭所有执行节点？") {
		worker.Stop(s.nodes)
		s.nodes = nil
	}
	return true
}

func (s *windowState) readConfigFromForm() (config.Config, error) {
	cfg := s.config
	cfg.ServerURL = strings.TrimSpace(getWindowText(s.serverEdit))
	cfg.ChromePath = strings.TrimSpace(getWindowText(s.chromeEdit))
	cfg.DownloadDir = strings.TrimSpace(getWindowText(s.downloadEdit))
	nodeCount, err := strconv.Atoi(strings.TrimSpace(getWindowText(s.nodeEdit)))
	if err != nil || nodeCount <= 0 {
		return cfg, fmt.Errorf("执行节点数必须是大于 0 的数字")
	}
	if nodeCount > 8 {
		return cfg, fmt.Errorf("执行节点数不能超过 8")
	}
	cfg.NodeCount = nodeCount
	return cfg, nil
}

func (s *windowState) setStatus(text string) {
	setWindowText(s.statusText, text)
}

func (s *windowState) automaStatus() string {
	extensionDir := resolvePath(s.exeDir, s.config.AutomaExtensionDir, config.DefaultAutomaExtensionDir())
	manifestPath := filepath.Join(extensionDir, "manifest.json")
	info, err := os.Stat(manifestPath)
	if err != nil {
		return "未安装（点击“安装/更新 Automa”选择 zip 安装包）"
	}
	return "已安装，更新时间：" + info.ModTime().Format("2006-01-02 15:04:05")
}

func createLabel(parent uintptr, text string, x, y, width, height int) uintptr {
	return createControl("STATIC", text, wsChild|wsVisible, parent, 0, x, y, width, height)
}

func createEdit(parent uintptr, id int, text string, x, y, width, height int) uintptr {
	return createControl("EDIT", text, wsChild|wsVisible|wsBorder|esAutoHScroll, parent, id, x, y, width, height)
}

func createButton(parent uintptr, id int, text string, x, y, width, height int) uintptr {
	return createControl("BUTTON", text, wsChild|wsVisible|bsPushButton, parent, id, x, y, width, height)
}

func createControl(className, text string, style uintptr, parent uintptr, id int, x, y, width, height int) uintptr {
	hwnd, _, _ := proc("CreateWindowExW").Call(
		0,
		uintptr(unsafe.Pointer(utf16Ptr(className))),
		uintptr(unsafe.Pointer(utf16Ptr(text))),
		style,
		uintptr(x),
		uintptr(y),
		uintptr(width),
		uintptr(height),
		parent,
		uintptr(id),
		getModuleHandle(),
		0,
	)
	return hwnd
}

func wndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case wmNccreate:
		create := (*createstruct)(unsafe.Pointer(lParam))
		state := (*windowState)(unsafe.Pointer(create.lpCreateParams))
		setWindowLongPtr(hwnd, gwlpUserData, uintptr(unsafe.Pointer(state)))
		state.hwnd = hwnd
		return 1
	case wmCreate:
		if state := getState(hwnd); state != nil {
			state.initControls()
		}
		return 0
	case wmCommand:
		if state := getState(hwnd); state != nil {
			state.handleCommand(int(loword(uint32(wParam))))
		}
		return 0
	case wmClose:
		if state := getState(hwnd); state != nil {
			if state.confirmClose() {
				proc("DestroyWindow").Call(hwnd)
			}
			return 0
		}
	case wmWorkerStopped:
		if state := getState(hwnd); state != nil {
			state.stopping = false
			state.setStatus(fmt.Sprintf("已关闭 %d 个执行节点", int(wParam)))
		}
		return 0
	case wmDestroy:
		proc("PostQuitMessage").Call(0)
		return 0
	}
	ret, _, _ := proc("DefWindowProcW").Call(hwnd, uintptr(msg), wParam, lParam)
	return ret
}

func getState(hwnd uintptr) *windowState {
	ptr := getWindowLongPtr(hwnd, gwlpUserData)
	if ptr == 0 {
		return nil
	}
	return (*windowState)(unsafe.Pointer(ptr))
}

type fileFilter struct {
	Name    string
	Pattern string
}

func openFileDialog(owner uintptr, title string, filters []fileFilter, defaultExt string) (string, bool) {
	buffer := make([]uint16, 260)
	filter := buildFileFilter(filters)
	ofn := openfilename{
		lStructSize: uint32(unsafe.Sizeof(openfilename{})),
		hwndOwner:   owner,
		lpstrFilter: uintptr(unsafe.Pointer(&filter[0])),
		lpstrFile:   uintptr(unsafe.Pointer(&buffer[0])),
		nMaxFile:    uint32(len(buffer)),
		lpstrTitle:  uintptr(unsafe.Pointer(utf16Ptr(title))),
		flags:       ofnFileMustExist | ofnPathMustExist,
		lpstrDefExt: uintptr(unsafe.Pointer(utf16Ptr(defaultExt))),
	}
	ret, _, _ := comdlg32.NewProc("GetOpenFileNameW").Call(uintptr(unsafe.Pointer(&ofn)))
	if ret == 0 {
		return "", false
	}
	return syscall.UTF16ToString(buffer), true
}

func openFolderDialog(owner uintptr, title string) (string, bool) {
	displayName := make([]uint16, 260)
	dialog := browseinfo{
		hwndOwner:      owner,
		pszDisplayName: uintptr(unsafe.Pointer(&displayName[0])),
		lpszTitle:      uintptr(unsafe.Pointer(utf16Ptr(title))),
		ulFlags:        bifReturnOnlyFSDirs | bifNewDialogStyle,
	}
	pidl, _, _ := proc("SHBrowseForFolderW").Call(uintptr(unsafe.Pointer(&dialog)))
	if pidl == 0 {
		return "", false
	}
	defer ole32.NewProc("CoTaskMemFree").Call(pidl)

	buffer := make([]uint16, 260)
	ret, _, _ := proc("SHGetPathFromIDListW").Call(pidl, uintptr(unsafe.Pointer(&buffer[0])))
	if ret == 0 {
		return "", false
	}
	return syscall.UTF16ToString(buffer), true
}

func buildFileFilter(filters []fileFilter) []uint16 {
	var parts []string
	for _, filter := range filters {
		parts = append(parts, filter.Name, filter.Pattern)
	}
	return utf16Filter(parts...)
}

func utf16Filter(parts ...string) []uint16 {
	var filter []uint16
	for _, part := range parts {
		filter = append(filter, syscall.StringToUTF16(part)...)
	}
	return append(filter, 0)
}

func resolvePath(baseDir, value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		value = fallback
	}
	if filepath.IsAbs(value) {
		return value
	}
	return filepath.Join(baseDir, value)
}

func executableDir() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(exePath), nil
}

func messageBox(owner uintptr, message string) {
	proc("MessageBoxW").Call(owner, uintptr(unsafe.Pointer(utf16Ptr(message))), uintptr(unsafe.Pointer(utf16Ptr(windowTitle))), 0)
}

func confirmBox(owner uintptr, message string) bool {
	ret, _, _ := proc("MessageBoxW").Call(
		owner,
		uintptr(unsafe.Pointer(utf16Ptr(message))),
		uintptr(unsafe.Pointer(utf16Ptr(windowTitle))),
		mbYesNo|mbIconQuestion,
	)
	return ret == idYes
}

func setWindowText(hwnd uintptr, text string) {
	proc("SetWindowTextW").Call(hwnd, uintptr(unsafe.Pointer(utf16Ptr(text))))
}

func getWindowText(hwnd uintptr) string {
	length, _, _ := proc("GetWindowTextLengthW").Call(hwnd)
	buffer := make([]uint16, int(length)+1)
	proc("GetWindowTextW").Call(hwnd, uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)))
	return syscall.UTF16ToString(buffer)
}

func sendMessage(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	ret, _, _ := proc("SendMessageW").Call(hwnd, uintptr(msg), wParam, lParam)
	return ret
}

func postMessage(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	ret, _, _ := proc("PostMessageW").Call(hwnd, uintptr(msg), wParam, lParam)
	return ret
}

func proc(name string) *syscall.LazyProc {
	if p, ok := procCache[name]; ok {
		return p
	}
	for _, dll := range []*syscall.LazyDLL{user32, kernel32, gdi32, shell32} {
		p := dll.NewProc(name)
		if err := p.Find(); err == nil {
			procCache[name] = p
			return p
		}
	}
	p := user32.NewProc(name)
	procCache[name] = p
	return p
}

func getModuleHandle() uintptr {
	ret, _, _ := kernel32.NewProc("GetModuleHandleW").Call(0)
	return ret
}

func getDefaultFont() uintptr {
	ret, _, _ := gdi32.NewProc("GetStockObject").Call(defaultGUIFont)
	return ret
}

func createUIFont() uintptr {
	logFont := logfont{
		lfHeight: -16,
		lfWeight: 400,
	}
	copy(logFont.lfFaceName[:], syscall.StringToUTF16("Segoe UI"))
	ret, _, _ := gdi32.NewProc("CreateFontIndirectW").Call(uintptr(unsafe.Pointer(&logFont)))
	return ret
}

func loadDefaultCursor() uintptr {
	ret, _, _ := proc("LoadCursorW").Call(0, uintptr(idcArrow))
	return ret
}

func loadDefaultIcon() uintptr {
	ret, _, _ := proc("LoadIconW").Call(0, uintptr(idiApplication))
	return ret
}

func loadResourceIcon(size int) uintptr {
	if icon := loadResourceIconByName("APP", size); icon != 0 {
		return icon
	}
	return loadResourceIconByID(1, size)
}

func loadResourceIconByName(name string, size int) uintptr {
	ret, _, _ := proc("LoadImageW").Call(
		getModuleHandle(),
		uintptr(unsafe.Pointer(utf16Ptr(name))),
		imageIcon,
		uintptr(size),
		uintptr(size),
		0,
	)
	return ret
}

func loadResourceIconByID(id uintptr, size int) uintptr {
	ret, _, _ := proc("LoadImageW").Call(
		getModuleHandle(),
		id,
		imageIcon,
		uintptr(size),
		uintptr(size),
		0,
	)
	return ret
}

func loadIcon(path string) uintptr {
	if _, err := os.Stat(path); err != nil {
		return 0
	}
	ret, _, _ := proc("LoadImageW").Call(
		0,
		uintptr(unsafe.Pointer(utf16Ptr(path))),
		imageIcon,
		32,
		32,
		lrLoadFromFile,
	)
	return ret
}

func setWindowLongPtr(hwnd uintptr, index int, value uintptr) uintptr {
	ret, _, _ := proc("SetWindowLongPtrW").Call(hwnd, uintptr(index), value)
	return ret
}

func getWindowLongPtr(hwnd uintptr, index int) uintptr {
	ret, _, _ := proc("GetWindowLongPtrW").Call(hwnd, uintptr(index))
	return ret
}

func utf16Ptr(value string) *uint16 {
	return syscall.StringToUTF16Ptr(value)
}

func loword(value uint32) uint16 {
	return uint16(value & 0xffff)
}

type wndclassex struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     uintptr
	hIcon         uintptr
	hCursor       uintptr
	hbrBackground uintptr
	lpszMenuName  uintptr
	lpszClassName uintptr
	hIconSm       uintptr
}

type msg struct {
	hwnd    uintptr
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      point
}

type point struct {
	x int32
	y int32
}

type createstruct struct {
	lpCreateParams uintptr
	hInstance      uintptr
	hMenu          uintptr
	hwndParent     uintptr
	cy             int32
	cx             int32
	y              int32
	x              int32
	style          int32
	lpszName       uintptr
	lpszClass      uintptr
	exStyle        uint32
}

type openfilename struct {
	lStructSize       uint32
	hwndOwner         uintptr
	hInstance         uintptr
	lpstrFilter       uintptr
	lpstrCustomFilter uintptr
	nMaxCustFilter    uint32
	nFilterIndex      uint32
	lpstrFile         uintptr
	nMaxFile          uint32
	lpstrFileTitle    uintptr
	nMaxFileTitle     uint32
	lpstrInitialDir   uintptr
	lpstrTitle        uintptr
	flags             uint32
	nFileOffset       uint16
	nFileExtension    uint16
	lpstrDefExt       uintptr
	lCustData         uintptr
	lpfnHook          uintptr
	lpTemplateName    uintptr
	pvReserved        uintptr
	dwReserved        uint32
	flagsEx           uint32
}

type browseinfo struct {
	hwndOwner      uintptr
	pidlRoot       uintptr
	pszDisplayName uintptr
	lpszTitle      uintptr
	ulFlags        uint32
	lpfn           uintptr
	lParam         uintptr
	iImage         int32
}

type logfont struct {
	lfHeight         int32
	lfWidth          int32
	lfEscapement     int32
	lfOrientation    int32
	lfWeight         int32
	lfItalic         byte
	lfUnderline      byte
	lfStrikeOut      byte
	lfCharSet        byte
	lfOutPrecision   byte
	lfClipPrecision  byte
	lfQuality        byte
	lfPitchAndFamily byte
	lfFaceName       [32]uint16
}

const (
	cwUseDefault = ^uintptr(0)

	wsOverlappedWindow = 0x00cf0000
	wsMaximizeBox      = 0x00010000
	wsSizeBox          = 0x00040000
	wsChild            = 0x40000000
	wsVisible          = 0x10000000
	wsBorder           = 0x00800000
	esAutoHScroll      = 0x00000080
	bsPushButton       = 0x00000000

	wmNccreate = 0x0081
	wmCreate   = 0x0001
	wmClose    = 0x0010
	wmCommand  = 0x0111
	wmDestroy  = 0x0002
	wmSetFont  = 0x0030
	wmSetIcon  = 0x0080

	wmWorkerStopped = 0x8001

	iconSmall = 0
	iconBig   = 1

	gwlpUserData = -21

	defaultGUIFont = 17
	imageIcon      = 1
	lrLoadFromFile = 0x00000010
	idcArrow       = 32512
	idiApplication = 32512
	mbYesNo        = 0x00000004
	mbIconQuestion = 0x00000020
	idYes          = 6

	ofnFileMustExist = 0x00001000
	ofnPathMustExist = 0x00000800

	bifReturnOnlyFSDirs = 0x00000001
	bifNewDialogStyle   = 0x00000040
)
