//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

const (
	wsOverlappedWindow = 0x00CF0000
	wsChild            = 0x40000000
	wsVisible          = 0x10000000
	wsClipChildren     = 0x02000000
	exTopmost          = 0x00000008
	exToolwindow       = 0x00000080

	wmDestroy        = 0x0002
	wmPaint          = 0x000F
	wmClose          = 0x0010
	wmSetFont        = 0x0030
	wmCommand        = 0x0111
	wmTimer          = 0x0113
	wmHotkey         = 0x0312
	wmCtlColorStatic = 0x0138
	wmCtlColorBtn    = 0x0135
	wmEraseBkgnd     = 0x0014

	bsAutoCheckbox = 0x0003
	bmGetCheck     = 0x00F0
	bmSetCheck     = 0x00F1
	bstChecked     = 0x0001

	swShow          = 5
	hwndTopmost     = ^uintptr(0)
	swpNomove       = 0x0002
	swpNosize       = 0x0001
	swpNoactivate   = 0x0010
	swpShowWindow   = 0x0040
	srcCopy         = 0x00CC0020
	transparent     = 1
	modNorepeat     = 0x4000
	vkF8            = 0x77
	keyeventfKeyup  = 0x0002
	keyeventfScan   = 0x0008
	smCxScreen      = 0
	smCyScreen      = 1
	processQuery    = 0x1000
	colorWindow     = 5
	idCheck         = 101
)

var (
	user32 = syscall.NewLazyDLL("user32.dll")
	gdi32  = syscall.NewLazyDLL("gdi32.dll")
	kernel = syscall.NewLazyDLL("kernel32.dll")
	winmm  = syscall.NewLazyDLL("winmm.dll")

	pSetProcessDPIAware      = user32.NewProc("SetProcessDPIAware")
	pRegisterClassExW        = user32.NewProc("RegisterClassExW")
	pCreateWindowExW         = user32.NewProc("CreateWindowExW")
	pDefWindowProcW          = user32.NewProc("DefWindowProcW")
	pGetMessageW             = user32.NewProc("GetMessageW")
	pPeekMessageW            = user32.NewProc("PeekMessageW")
	pTranslateMessage        = user32.NewProc("TranslateMessage")
	pDispatchMessageW        = user32.NewProc("DispatchMessageW")
	pPostQuitMessage         = user32.NewProc("PostQuitMessage")
	pDestroyWindow           = user32.NewProc("DestroyWindow")
	pShowWindow              = user32.NewProc("ShowWindow")
	pMessageBoxW             = user32.NewProc("MessageBoxW")
	pSetForegroundWindow     = user32.NewProc("SetForegroundWindow")
	pBringWindowToTop       = user32.NewProc("BringWindowToTop")
	pUpdateWindow            = user32.NewProc("UpdateWindow")
	pSetTimer                = user32.NewProc("SetTimer")
	pRegisterHotKey          = user32.NewProc("RegisterHotKey")
	pSendMessageW            = user32.NewProc("SendMessageW")
	pInvalidateRect          = user32.NewProc("InvalidateRect")
	pBeginPaint              = user32.NewProc("BeginPaint")
	pEndPaint                = user32.NewProc("EndPaint")
	pFillRect                   = user32.NewProc("FillRect")
	pSetBkMode                  = gdi32.NewProc("SetBkMode")
	pSetTextColor               = gdi32.NewProc("SetTextColor")
	pTextOutW                   = gdi32.NewProc("TextOutW")
	pGetSystemMetrics        = user32.NewProc("GetSystemMetrics")
	pSetWindowPos            = user32.NewProc("SetWindowPos")
	pGetForegroundWindow     = user32.NewProc("GetForegroundWindow")
	pGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	pGetWindowRect           = user32.NewProc("GetWindowRect")
	pGetWindowTextW          = user32.NewProc("GetWindowTextW")
	pEnumWindows             = user32.NewProc("EnumWindows")
	pIsWindowVisible         = user32.NewProc("IsWindowVisible")
	pGetDC                   = user32.NewProc("GetDC")
	pReleaseDC               = user32.NewProc("ReleaseDC")
	pLoadCursorW             = user32.NewProc("LoadCursorW")
	pSendInput               = user32.NewProc("SendInput")
	pVkKeyScanExW            = user32.NewProc("VkKeyScanExW")
	pGetKeyboardLayout       = user32.NewProc("GetKeyboardLayout")
	pMapVirtualKeyW          = user32.NewProc("MapVirtualKeyW")
	pKeybdEvent              = user32.NewProc("keybd_event")
	pAttachThreadInput       = user32.NewProc("AttachThreadInput")
	pGetCurrentThreadId      = kernel.NewProc("GetCurrentThreadId")
	pGetModuleHandleW        = kernel.NewProc("GetModuleHandleW")
	pOpenProcess             = kernel.NewProc("OpenProcess")
	pCloseHandle                = kernel.NewProc("CloseHandle")
	pSleep                      = kernel.NewProc("Sleep")
	pTimeBeginPeriod            = winmm.NewProc("timeBeginPeriod")
	pTimeEndPeriod              = winmm.NewProc("timeEndPeriod")
	pQueryFullProcessImageNameW = kernel.NewProc("QueryFullProcessImageNameW")
	pCreateCompatibleDC      = gdi32.NewProc("CreateCompatibleDC")
	pCreateCompatibleBitmap  = gdi32.NewProc("CreateCompatibleBitmap")
	pSelectObject            = gdi32.NewProc("SelectObject")
	pBitBlt                  = gdi32.NewProc("BitBlt")
	pGetDIBits               = gdi32.NewProc("GetDIBits")
	pDeleteObject            = gdi32.NewProc("DeleteObject")
	pDeleteDC                = gdi32.NewProc("DeleteDC")
	pCreateSolidBrush        = gdi32.NewProc("CreateSolidBrush")
	pCreateFontW             = gdi32.NewProc("CreateFontW")
	pStretchDIBits           = gdi32.NewProc("StretchDIBits")
	pRoundRect               = gdi32.NewProc("RoundRect")
	pEllipse                 = gdi32.NewProc("Ellipse")
	pGetStockObject          = gdi32.NewProc("GetStockObject")
	pGetClientRect           = user32.NewProc("GetClientRect")
)

var (
	mainHwnd  uintptr
	checkHwnd uintptr
	bgBrush   uintptr
	cardBrush uintptr
	wellBrush uintptr
	onBrush   uintptr
	offBrush  uintptr
	knobBrush uintptr
	nullPen   uintptr
	font      uintptr
	fontBig   uintptr
	fontSmall uintptr
	fontTitle uintptr
	autoQTE   bool
	status    = "En attente de FiveM"
	detail    = "L'aide est coupée. Aucune touche n'est envoyée."
	wasInZone bool
	arm       Track
	miss       int
	heldKey    byte
	heldKeyAt  time.Time
	qteDone    bool
	pendingAt  time.Time
	voteRing   [9]byte
	voteN      int
	downKey    byte
	fpsN       int
	fpsAt      time.Time
	fps        int
	cachedGame uintptr
	cachedWhen time.Time
	releaseAt  time.Time
	lastTap   time.Time
	tickN      int
	enumCB     uintptr
	gameHWND   uintptr
	gameArea   int
	frame      []byte
	frameW     int
	frameH     int
)

func init() {
	runtime.LockOSThread()
}

func exeDir() string {
	path, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(path)
}

func logLine(text string) {
	path := filepath.Join(exeDir(), "Portee-log.txt")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.WriteString(text)
}

func alert(text string) {
	logLine(text)
	body, _ := syscall.UTF16PtrFromString(text)
	title, _ := syscall.UTF16PtrFromString("Portée")
	pMessageBoxW.Call(0, uintptr(unsafe.Pointer(body)), uintptr(unsafe.Pointer(title)), 0x00000030)
}

func main() {
	runtime.LockOSThread()
	defer func() {
		if rec := recover(); rec != nil {
			alert(fmt.Sprintf("Portée s'est arrêté.\n%v\n\n%s", rec, debug.Stack()))
		}
	}()

	pSetProcessDPIAware.Call()
	autoQTE = loadAuto()
	inst, _, err := pGetModuleHandleW.Call(0)
	if inst == 0 {
		alert("GetModuleHandle a échoué: " + err.Error())
		return
	}
	bgBrush, _, _ = pCreateSolidBrush.Call(0x00110F0E)
	cardBrush, _, _ = pCreateSolidBrush.Call(0x001C1917)
	wellBrush, _, _ = pCreateSolidBrush.Call(0x00140F0E)
	onBrush, _, _ = pCreateSolidBrush.Call(0x0098E0B4)
	offBrush, _, _ = pCreateSolidBrush.Call(0x00302826)
	knobBrush, _, _ = pCreateSolidBrush.Call(0x00F6F3EE)
	nullPen, _, _ = pGetStockObject.Call(8)
	cursor, _, _ := pLoadCursorW.Call(0, 32512)
	fontName, _ := syscall.UTF16PtrFromString("Segoe UI")
	bodyH, bigH, smallH, titleH := int32(-15), int32(-36), int32(-12), int32(-22)
	font, _, _ = pCreateFontW.Call(uintptr(bodyH), 0, 0, 0, 400, 0, 0, 0, 1, 0, 0, 5, 0, uintptr(unsafe.Pointer(fontName)))
	fontBig, _, _ = pCreateFontW.Call(uintptr(bigH), 0, 0, 0, 700, 0, 0, 0, 1, 0, 0, 5, 0, uintptr(unsafe.Pointer(fontName)))
	fontSmall, _, _ = pCreateFontW.Call(uintptr(smallH), 0, 0, 0, 400, 0, 0, 0, 1, 0, 0, 5, 0, uintptr(unsafe.Pointer(fontName)))
	fontTitle, _, _ = pCreateFontW.Call(uintptr(titleH), 0, 0, 0, 600, 0, 0, 0, 1, 0, 0, 5, 0, uintptr(unsafe.Pointer(fontName)))

	className, _ := syscall.UTF16PtrFromString("PorteeQTE")
	var wc struct {
		size       uint32
		style      uint32
		wndProc    uintptr
		clsExtra   int32
		wndExtra   int32
		instance   uintptr
		icon       uintptr
		cursor     uintptr
		background uintptr
		menuName   *uint16
		className  *uint16
		iconSm     uintptr
	}
	wc.size = uint32(unsafe.Sizeof(wc))
	wc.style = 3
	wc.wndProc = syscall.NewCallback(wndProc)
	wc.instance = inst
	wc.cursor = cursor
	if bgBrush != 0 {
		wc.background = bgBrush
	} else {
		wc.background = 6
	}
	wc.className = className
	atom, _, regErr := pRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	sw, _, _ := pGetSystemMetrics.Call(smCxScreen)
	const winW, winH = 440, 640
	x := int(sw) - winW - 16
	y := 16
	if x < 16 {
		x = 16
	}
	title, _ := syscall.UTF16PtrFromString("Portée")
	hwnd, _, createErr := pCreateWindowExW.Call(
		exTopmost|0x02000000,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(title)),
		wsOverlappedWindow|wsVisible|wsClipChildren,
		uintptr(x), uintptr(y), winW, winH,
		0, 0, inst, 0,
	)
	mainHwnd = hwnd
	if mainHwnd == 0 {
		alert(fmt.Sprintf("La fenêtre n'a pas pu s'ouvrir.\nClasse: %v (%s)\nCréation: %s\n\nEssaie un clic droit sur Portee.exe, Propriétés, Débloquer.", atom, regErr, createErr))
		return
	}
	refreshStatus()
	pShowWindow.Call(mainHwnd, swShow)
	pUpdateWindow.Call(mainHwnd)
	pSetWindowPos.Call(mainHwnd, hwndTopmost, uintptr(x), uintptr(y), winW, winH, swpShowWindow)
	pSetForegroundWindow.Call(mainHwnd)
	pRegisterHotKey.Call(mainHwnd, 1, modNorepeat, vkF8)
	pTimeBeginPeriod.Call(1)
	defer pTimeEndPeriod.Call(1)
	logLine("fenêtre ouverte")

	var msg struct {
		hwnd    uintptr
		message uint32
		wparam  uintptr
		lparam  uintptr
		time    uint32
		pt      struct{ x, y int32 }
	}
	started := time.Now()
	next := time.Now()
	quit := false
	for !quit {
		for {
			r, _, _ := pPeekMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0, 1)
			if r == 0 {
				break
			}
			if msg.message == 0x0012 {
				quit = true
				break
			}
			pTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
			pDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
		}
		if quit {
			break
		}
		now := time.Now()
		if now.Before(next) {
			wait := next.Sub(now).Milliseconds()
			if wait < 1 {
				wait = 1
			}
			pSleep.Call(uintptr(wait))
			continue
		}
		start := time.Now()
		onTick()
		gap := 4 * time.Millisecond
		if time.Since(start) > 8*time.Millisecond {
			gap = 6 * time.Millisecond
		}
		next = time.Now().Add(gap)
	}
	if time.Since(started) < 2*time.Second {
		alert("La fenêtre s'est fermée tout de suite. Relance Portee.exe. Si Windows affiche un écran bleu de protection, clique Informations complémentaires, puis Exécuter quand même.")
	}
}

func wndProc(hwnd, msg, wparam, lparam uintptr) uintptr {
	defer func() {
		if rec := recover(); rec != nil {
			alert(fmt.Sprintf("Erreur d'affichage: %v", rec))
		}
	}()
	switch msg {
	case wmCommand:
		return 0
	case 0x0202:
		y := int32(int16(lparam >> 16))
		if y >= 248 && y <= 336 {
			flipAuto()
		}
	case wmHotkey:
		flipAuto()
	case wmTimer:
		return 0
	case 0x0014:
		return 1
	case wmCtlColorStatic, wmCtlColorBtn:
		pSetBkMode.Call(wparam, transparent)
		pSetTextColor.Call(wparam, 0x00E7F1F6)
		return bgBrush
	case wmPaint:
		paint(hwnd)
	case wmClose:
		releaseHeld()
		pDestroyWindow.Call(hwnd)
	case wmDestroy:
		pPostQuitMessage.Call(0)
	default:
		r, _, _ := pDefWindowProcW.Call(hwnd, msg, wparam, lparam)
		return r
	}
	return 0
}

func paint(hwnd uintptr) {
	var ps struct {
		hdc       uintptr
		erase     int32
		rc        [4]int32
		restore   int32
		incUpdate int32
		reserved  [32]byte
	}
	hdc, _, _ := pBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
	var client [4]int32
	pGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&client)))
	cw, ch := client[2], client[3]
	if cw < 120 {
		cw = 420
	}
	if ch < 200 {
		ch = 600
	}
	pFillRect.Call(hdc, uintptr(unsafe.Pointer(&client)), bgBrush)
	pSetBkMode.Call(hdc, transparent)
	bar := [4]int32{0, 0, cw, 3}
	pFillRect.Call(hdc, uintptr(unsafe.Pointer(&bar)), onBrush)

	old, _, _ := pSelectObject.Call(hdc, fontTitle)
	pSetTextColor.Call(hdc, 0x00F7F4EF)
	draw(hdc, 24, 22, "Portée")
	pSelectObject.Call(hdc, fontSmall)
	pSetTextColor.Call(hdc, 0x00948C86)
	draw(hdc, 24, 52, "Aide pour les QTE FiveM")
	if fps > 0 {
		draw(hdc, int(cw)-92, 52, fmt.Sprintf("%d det.", fps))
	}

	fillRound(hdc, cardBrush, 16, 84, cw-16, 228, 20)
	fillRound(hdc, wellBrush, 32, 108, 124, 200, 92)
	letter := "–"
	if heldKey >= 'A' && heldKey <= 'Z' {
		letter = string(heldKey)
	}
	pSelectObject.Call(hdc, fontBig)
	if autoQTE && letter != "–" {
		pSetTextColor.Call(hdc, 0x0098E0B4)
	} else {
		pSetTextColor.Call(hdc, 0x00F7F4EF)
	}
	draw(hdc, 62, 132, letter)
	pSelectObject.Call(hdc, font)
	pSetTextColor.Call(hdc, 0x00F7F4EF)
	draw(hdc, 144, 128, status)
	pSelectObject.Call(hdc, fontSmall)
	pSetTextColor.Call(hdc, 0x00948C86)
	if autoQTE {
		draw(hdc, 144, 156, "L'aide enverra la touche")
	} else {
		draw(hdc, 144, 156, "Aucune touche n'est envoyée")
	}

	fillRound(hdc, cardBrush, 16, 244, cw-16, 336, 20)
	track := offBrush
	knobX := int32(40)
	if autoQTE {
		track = onBrush
		knobX = 70
	}
	fillRound(hdc, track, 36, 270, 104, 306, 36)
	fillRound(hdc, knobBrush, knobX, 274, knobX+28, 302, 28)
	pSelectObject.Call(hdc, font)
	pSetTextColor.Call(hdc, 0x00F7F4EF)
	draw(hdc, 124, 266, "Aide automatique")
	pSelectObject.Call(hdc, fontSmall)
	if autoQTE {
		pSetTextColor.Call(hdc, 0x0098E0B4)
		draw(hdc, 124, 292, "Activée  ·  F8 pour couper")
	} else {
		pSetTextColor.Call(hdc, 0x00948C86)
		draw(hdc, 124, 292, "Coupée  ·  clique ou appuie sur F8")
	}

	fillRound(hdc, cardBrush, 16, 352, cw-16, ch-16, 20)
	pSetTextColor.Call(hdc, 0x00948C86)
	draw(hdc, 32, 366, "Aperçu du centre")
	pSelectObject.Call(hdc, old)
	previewTop := int32(392)
	previewSide := cw - 64
	room := ch - 16 - previewTop - 16
	if room < previewSide {
		previewSide = room
	}
	if previewSide < 40 {
		previewSide = 40
	}
	previewLeft := (cw - previewSide) / 2
	fillRound(hdc, wellBrush, previewLeft-6, previewTop-6, previewLeft+previewSide+6, previewTop+previewSide+6, 16)
	drawFrame(hdc, previewLeft, previewTop, previewLeft+previewSide, previewTop+previewSide)
	pEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
}

func fillRound(hdc, brush uintptr, l, t, r, b, rad int32) {
	oldB, _, _ := pSelectObject.Call(hdc, brush)
	oldP, _, _ := pSelectObject.Call(hdc, nullPen)
	pRoundRect.Call(hdc, uintptr(l), uintptr(t), uintptr(r), uintptr(b), uintptr(rad), uintptr(rad))
	pSelectObject.Call(hdc, oldP)
	pSelectObject.Call(hdc, oldB)
}

func draw(hdc uintptr, x, y int, text string) {
	u := syscall.StringToUTF16(text)
	pTextOutW.Call(hdc, uintptr(x), uintptr(y), uintptr(unsafe.Pointer(&u[0])), uintptr(len(u)-1))
}

func flipAuto() {
	autoQTE = !autoQTE
	saveAuto(autoQTE)
	refreshStatus()
}

func refreshStatus() {
	if autoQTE {
		detail = "L'aide est active."
	} else {
		detail = "L'aide est coupée. Aucune touche n'est envoyée."
	}
	pInvalidateRect.Call(mainHwnd, 0, 0)
}

func drawFrame(hdc uintptr, l, t, r, b int32) {
	if r-l < 20 || b-t < 20 {
		return
	}
	if len(frame) < frameW*frameH*4 || frameW < 2 || frameH < 2 {
		old, _, _ := pSelectObject.Call(hdc, fontSmall)
		pSetTextColor.Call(hdc, 0x00A39890)
		draw(hdc, int(l), int(t)+12, "L'aperçu arrive avec FiveM.")
		pSelectObject.Call(hdc, old)
		return
	}
	side := r - l
	if b-t < side {
		side = b - t
	}
	var info bitmapHeader
	info.size = uint32(unsafe.Sizeof(info))
	info.width = int32(frameW)
	info.height = -int32(frameH)
	info.planes = 1
	info.bitCount = 32
	pStretchDIBits.Call(
		hdc,
		uintptr(l), uintptr(t), uintptr(side), uintptr(side),
		0, 0, uintptr(frameW), uintptr(frameH),
		uintptr(unsafe.Pointer(&frame[0])),
		uintptr(unsafe.Pointer(&info)),
		0, srcCopy,
	)
}

func onTick() {
	releaseIfDue()
	tickN++
	if fpsAt.IsZero() {
		fpsAt = time.Now()
	}
	fpsN++
	if time.Since(fpsAt) >= time.Second {
		fps = fpsN
		fpsN = 0
		fpsAt = time.Now()
	}
	if tickN%90 == 0 {
		pSetWindowPos.Call(mainHwnd, hwndTopmost, 0, 0, 0, 0, swpNomove|swpNosize|swpNoactivate)
	}
	game := findGame()
	if game == 0 {
		setStatus("En attente de FiveM")
		wasInZone = false
		qteDone = false
		heldKey = 0
		return
	}
	pix, w, h := captureCenter(game)
	if pix == nil {
		return
	}
	frame, frameW, frameH = pix, w, h
	if tickN%2 == 0 {
		pInvalidateRect.Call(mainHwnd, 0, 0)
	}
	if frameDark(pix) {
		setStatus("Image noire — FiveM en sans bordure")
		arm = Track{}
		heldKey = 0
		voteN = 0
		qteDone = false
		return
	}
	res := Analyze(pix, w, h)
	if !res.Seen {
		miss++
		if miss > 8 {
			arm = Track{}
			heldKey = 0
			voteN = 0
			qteDone = false
			pendingAt = time.Time{}
		}
		setStatus("En attente du cercle")
		return
	}
	miss = 0
	if qteDone {
		heldKey = 0
		voteN = 0
		setStatus("En attente du prochain")
		return
	}
	fresh := res.Key
	if res.Key != 0 {
		if voted := pushVote(res.Key); voted != 0 {
			heldKey = voted
			heldKeyAt = time.Now()
			res.Key = voted
		} else if heldKey != 0 && time.Since(heldKeyAt) < 400*time.Millisecond {
			res.Key = heldKey
		}
	} else if heldKey != 0 && time.Since(heldKeyAt) < 400*time.Millisecond {
		res.Key = heldKey
	}
	if res.Key == 0 && isQteKey(fresh) && !pendingAt.IsZero() {
		res.Key = fresh
	}
	line := "Lettre pas encore lue"
	if res.Key != 0 {
		line = "Prêt"
	}
	if res.HasNeedle {
		line = "Trait en approche"
	}
	press := arm.Press(res)
	if press && res.Key == 0 && isQteKey(fresh) {
		res.Key = fresh
	}
	if press && !isQteKey(res.Key) {
		pendingAt = time.Now()
		press = false
	}
	if !press && !pendingAt.IsZero() && time.Since(pendingAt) < 40*time.Millisecond && isQteKey(res.Key) {
		press = true
	}
	if press && !autoQTE {
		line = "Active l'aide pour envoyer"
	} else if press {
		line = "Touche envoyée"
	}
	setStatus(line)
	if autoQTE && press && isQteKey(res.Key) && time.Since(lastTap) > 90*time.Millisecond {
		lastTap = time.Now()
		pressKey(res.Key)
		qteDone = true
		heldKey = 0
		voteN = 0
		pendingAt = time.Time{}
		setStatus("En attente du prochain")
	}
}

func setStatus(next string) {
	if next == status {
		return
	}
	status = next
	pInvalidateRect.Call(mainHwnd, 0, 0)
}

func pushVote(key byte) byte {
	if voteN < len(voteRing) {
		voteRing[voteN] = key
		voteN++
	} else {
		copy(voteRing[:], voteRing[1:])
		voteRing[len(voteRing)-1] = key
	}
	var best byte
	bestN := 0
	for i := 0; i < voteN; i++ {
		n := 0
		for j := 0; j < voteN; j++ {
			if voteRing[j] == voteRing[i] {
				n++
			}
		}
		if n > bestN {
			bestN = n
			best = voteRing[i]
		}
	}
	if bestN >= 2 {
		return best
	}
	return 0
}

func findGame() uintptr {
	if cachedGame != 0 && time.Since(cachedWhen) < 250*time.Millisecond {
		vis, _, _ := pIsWindowVisible.Call(cachedGame)
		if vis != 0 {
			return cachedGame
		}
	}
	if enumCB == 0 {
		enumCB = syscall.NewCallback(enumGame)
	}
	gameHWND = 0
	gameArea = 0
	pEnumWindows.Call(enumCB, 0)
	cachedGame = gameHWND
	cachedWhen = time.Now()
	return cachedGame
}

func enumGame(hwnd, _ uintptr) uintptr {
	if hwnd == 0 || hwnd == mainHwnd {
		return 1
	}
	vis, _, _ := pIsWindowVisible.Call(hwnd)
	if vis == 0 || !isGameWindow(hwnd) {
		return 1
	}
	var r winRect
	ok, _, _ := pGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&r)))
	if ok == 0 {
		return 1
	}
	area := int(r.right-r.left) * int(r.bottom-r.top)
	if area > gameArea {
		gameArea = area
		gameHWND = hwnd
	}
	return 1
}

func isGameWindow(hwnd uintptr) bool {
	var pid uint32
	pGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	proc, _, _ := pOpenProcess.Call(processQuery, 0, uintptr(pid))
	name := ""
	if proc != 0 {
		buf := make([]uint16, 520)
		n := uint32(len(buf))
		pQueryFullProcessImageNameW.Call(proc, 0, uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&n)))
		pCloseHandle.Call(proc)
		name = strings.ToLower(syscall.UTF16ToString(buf))
	}
	if strings.Contains(name, "fivem") || strings.Contains(name, "gta5") || strings.Contains(name, "gtav") || strings.Contains(name, "citizenfx") {
		return true
	}
	titleBuf := make([]uint16, 256)
	pGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&titleBuf[0])), 256)
	title := strings.ToLower(syscall.UTF16ToString(titleBuf))
	return strings.Contains(title, "fivem") || strings.Contains(title, "grand theft auto") || strings.Contains(title, "cfx.re")
}

func frameDark(pix []byte) bool {
	if len(pix) < 64 {
		return true
	}
	sum, n := 0, 0
	step := 16 * 4
	for i := 0; i+2 < len(pix); i += step {
		sum += int(pix[i]) + int(pix[i+1]) + int(pix[i+2])
		n++
	}
	return n == 0 || sum/n < 20
}

type winRect struct{ left, top, right, bottom int32 }

func captureCenter(hwnd uintptr) ([]byte, int, int) {
	var r winRect
	ok, _, _ := pGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&r)))
	if ok == 0 {
		return nil, 0, 0
	}
	ww := int(r.right - r.left)
	hh := int(r.bottom - r.top)
	if ww < 100 || hh < 100 {
		return nil, 0, 0
	}
	side := ww
	if hh < side {
		side = hh
	}
	side = side * 30 / 100
	if side < 160 {
		side = 160
	}
	if side > 180 {
		side = 180
	}
	x := int(r.left) + ww/2 - side/2
	y := int(r.top) + hh/2 - side/2
	return grab(x, y, side, side)
}

type bitmapHeader struct {
	size      uint32
	width     int32
	height    int32
	planes    uint16
	bitCount  uint16
	compress  uint32
	sizeImage uint32
	xppm      int32
	yppm      int32
	clrUsed   uint32
	clrImp    uint32
}

var (
	grabMem uintptr
	grabBmp uintptr
	grabW   int
	grabH   int
	grabBuf []byte
)

func grab(x, y, w, h int) ([]byte, int, int) {
	if w < 2 || h < 2 || w > 320 || h > 320 {
		return nil, 0, 0
	}
	hdc, _, _ := pGetDC.Call(0)
	if hdc == 0 {
		return nil, 0, 0
	}
	defer pReleaseDC.Call(0, hdc)
	if grabMem == 0 {
		grabMem, _, _ = pCreateCompatibleDC.Call(hdc)
		if grabMem == 0 {
			return nil, 0, 0
		}
	}
	if grabBmp == 0 || grabW != w || grabH != h {
		if grabBmp != 0 {
			pDeleteObject.Call(grabBmp)
			grabBmp = 0
		}
		grabBmp, _, _ = pCreateCompatibleBitmap.Call(hdc, uintptr(w), uintptr(h))
		if grabBmp == 0 {
			return nil, 0, 0
		}
		grabW, grabH = w, h
		grabBuf = make([]byte, w*h*4)
	}
	old, _, _ := pSelectObject.Call(grabMem, grabBmp)
	ok, _, _ := pBitBlt.Call(grabMem, 0, 0, uintptr(w), uintptr(h), hdc, uintptr(x), uintptr(y), srcCopy)
	pSelectObject.Call(grabMem, old)
	if ok == 0 {
		pDeleteObject.Call(grabBmp)
		grabBmp = 0
		return nil, 0, 0
	}
	var info bitmapHeader
	info.size = uint32(unsafe.Sizeof(info))
	info.width = int32(w)
	info.height = -int32(h)
	info.planes = 1
	info.bitCount = 32
	pGetDIBits.Call(hdc, grabBmp, 0, uintptr(h), uintptr(unsafe.Pointer(&grabBuf[0])), uintptr(unsafe.Pointer(&info)), 0)
	return grabBuf, w, h
}

func focusGame(hwnd uintptr) {
	fg, _, _ := pGetForegroundWindow.Call()
	if fg == hwnd {
		return
	}
	our, _, _ := pGetCurrentThreadId.Call()
	fgThread, _, _ := pGetWindowThreadProcessId.Call(fg, 0)
	gameThread, _, _ := pGetWindowThreadProcessId.Call(hwnd, 0)
	if fgThread != 0 {
		pAttachThreadInput.Call(our, fgThread, 1)
	}
	if gameThread != 0 && gameThread != fgThread {
		pAttachThreadInput.Call(our, gameThread, 1)
	}
	pSetForegroundWindow.Call(hwnd)
	pBringWindowToTop.Call(hwnd)
	if gameThread != 0 && gameThread != fgThread {
		pAttachThreadInput.Call(our, gameThread, 0)
	}
	if fgThread != 0 {
		pAttachThreadInput.Call(our, fgThread, 0)
	}
}

func pressKey(key byte) {
	ch := rune(key + ('a' - 'A'))
	layout, _, _ := pGetKeyboardLayout.Call(0)
	vkPair, _, _ := pVkKeyScanExW.Call(uintptr(ch), layout)
	if int16(vkPair) == -1 {
		return
	}
	vk := uint16(vkPair & 0xff)
	scan, _, _ := pMapVirtualKeyW.Call(uintptr(vk), 0)
	if scan == 0 {
		scan = uintptr(vk)
	}
	send(vk, uint16(scan), keyeventfScan)
	pKeybdEvent.Call(uintptr(vk), scan, uintptr(keyeventfScan), 0)
	downKey = key
	releaseAt = time.Now().Add(35 * time.Millisecond)
}

func releaseIfDue() {
	if downKey != 0 && time.Now().After(releaseAt) {
		releaseHeld()
	}
}

func releaseHeld() {
	if downKey == 0 {
		return
	}
	ch := rune(downKey + ('a' - 'A'))
	layout, _, _ := pGetKeyboardLayout.Call(0)
	vkPair, _, _ := pVkKeyScanExW.Call(uintptr(ch), layout)
	vk := uint16(vkPair & 0xff)
	scan, _, _ := pMapVirtualKeyW.Call(uintptr(vk), 0)
	send(vk, uint16(scan), keyeventfScan|keyeventfKeyup)
	pKeybdEvent.Call(uintptr(vk), scan, uintptr(keyeventfScan|keyeventfKeyup), 0)
	downKey = 0
}

func send(vk, scan, flags uint16) {
	in := keyInput{Type: 1, Vk: vk, Scan: scan, Flags: uint32(flags)}
	pSendInput.Call(1, uintptr(unsafe.Pointer(&in)), unsafe.Sizeof(in))
}

func settingsFile() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = "."
	}
	dir = filepath.Join(dir, "Portee")
	_ = os.MkdirAll(dir, 0o700)
	return filepath.Join(dir, "reglages.txt")
}

func loadAuto() bool {
	b, err := os.ReadFile(settingsFile())
	if err != nil {
		return false
	}
	return strings.Contains(string(b), "auto=1")
}

func saveAuto(on bool) {
	val := "auto=0\n"
	if on {
		val = "auto=1\n"
	}
	_ = os.WriteFile(settingsFile(), []byte(val), 0o600)
}
