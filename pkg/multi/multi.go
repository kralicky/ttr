package multi

/*
#cgo LDFLAGS: -lxdo -lX11
#include <stdlib.h>
#include <xdo.h>
#include <X11/Xlib.h>
#include <X11/Xatom.h>
#include <X11/Xutil.h>
#include <X11/XKBlib.h>

Window window_at(Window* windows, int idx) {
	return windows[idx];
}

int event_type(XEvent* event) {
	return event->type;
}

XKeyEvent* key_event(XEvent* event) {
	return &event->xkey;
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

type XdoContext struct {
	xdt *C.xdo_t
}

type Window struct {
	xid *C.Window
}

type Manager struct {
	XdoContext

	windows   []Window
	onDispose []func()
}

func NewManager() (*Manager, error) {
	mgr := &Manager{}

	xdo := C.xdo_new(nil)
	xdo.debug = 1

	mgr.XdoContext = XdoContext{xdt: xdo}
	mgr.onDispose = append(mgr.onDispose, func() {
		C.xdo_free(mgr.xdt)
	})

	var windows *C.Window
	var nwindows C.uint
	if ret := C.xdo_search_windows(xdo, &C.struct_xdo_search{
		winname:    C.CString("^Toontown Rewritten$"),
		winclass:   C.CString("^Toontown Rewritten$"),
		searchmask: C.SEARCH_NAME | C.SEARCH_CLASS,
		max_depth:  2,
		require:    C.SEARCH_ALL,
	}, &windows, &nwindows); ret != 0 {
		return nil, fmt.Errorf("error searching for windows: %d", ret)
	}
	fmt.Println("found", nwindows, "windows")

	mgr.windows = make([]Window, 0, int(nwindows))
	for i := 0; i < int(nwindows); i++ {
		w := C.window_at(windows, C.int(i))
		mgr.windows = append(mgr.windows, Window{xid: &w})
	}

	return mgr, nil
}

func (m *Manager) sendKeyDown(key string) {
	keycstr := C.CString(key)
	for _, w := range m.windows {
		C.xdo_send_keysequence_window_down(m.xdt, *w.xid, keycstr, 0)
	}
	C.free(unsafe.Pointer(keycstr))
}

func (m *Manager) sendKeyUp(key string) {
	keycstr := C.CString(key)
	for _, w := range m.windows {
		C.xdo_send_keysequence_window_up(m.xdt, *w.xid, keycstr, 0)
	}
	C.free(unsafe.Pointer(keycstr))
}

func (m *Manager) RunInputWindow() {
	classHint := C.XClassHint{
		res_class: C.CString("Multitoon Input"),
		res_name:  C.CString("Multitoon Input"),
	}
	screen := C.XDefaultScreen(m.xdt.xdpy)
	root := C.XRootWindow(m.xdt.xdpy, screen)

	whitePixel := C.XWhitePixel(m.xdt.xdpy, screen)
	blackPixel := C.XBlackPixel(m.xdt.xdpy, screen)

	w := C.XCreateSimpleWindow(
		m.xdt.xdpy,
		root,
		0, 0,
		250, 200,
		1,
		blackPixel,
		whitePixel,
	)

	C.XSelectInput(m.xdt.xdpy, w, C.KeyPressMask|C.KeyReleaseMask)
	C.XStoreName(m.xdt.xdpy, w, C.CString("Multitoon Input"))
	C.XSetClassHint(m.xdt.xdpy, w, &classHint)

	C.XMapWindow(m.xdt.xdpy, w)
	C.XSync(m.xdt.xdpy, 0)

	C.XkbSetDetectableAutoRepeat(m.xdt.xdpy, C.True, nil)
	C.XFlush(m.xdt.xdpy)

	pressed := make([]bool, 1024)
	for {
		var event C.XEvent
		C.XNextEvent(m.xdt.xdpy, &event)

		var keyEvent *C.XKeyEvent
		var sendFunc func(string)
		switch C.event_type(&event) {
		case C.KeyPress:
			keyEvent = C.key_event(&event)
			if pressed[uint(keyEvent.keycode)] {
				continue
			}
			pressed[uint(keyEvent.keycode)] = true
			// fmt.Println("key press", keyEvent.keycode)
			sendFunc = m.sendKeyDown
		case C.KeyRelease:
			keyEvent = C.key_event(&event)
			pressed[uint(keyEvent.keycode)] = false
			// fmt.Println("key release", keyEvent.keycode)
			sendFunc = m.sendKeyUp
		default:
			continue
		}

		var modifier string
		if keyEvent.state&C.ShiftMask != 0 && (keyEvent.keycode != 50 && keyEvent.keycode != 62) {
			modifier = "shift+"
		}
		keysym := C.XkbKeycodeToKeysym(m.xdt.xdpy, C.KeyCode(keyEvent.keycode), 0, 0)
		keystr := C.XKeysymToString(keysym) // do not free the result

		fullKeyStr := modifier + C.GoString(keystr)
		// fmt.Printf("sending %q\n", fullKeyStr)
		sendFunc(fullKeyStr)
	}
}

func (m *Manager) Dispose() {
	for _, f := range m.onDispose {
		f()
	}
}
