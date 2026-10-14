//go:build windows
// +build windows

package edge

import (
	"syscall"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

// call invokes a vtable entry the way native code does.
func call(proc ComProc, args ...uintptr) uintptr {
	r, _, _ := syscall.SyscallN(uintptr(proc), args...)
	return r
}

func getString(t *testing.T, o *environmentOptions, proc ComProc) string {
	t.Helper()
	var p *uint16
	if hr := call(proc, uintptr(unsafe.Pointer(o)), uintptr(unsafe.Pointer(&p))); hr != sOK {
		t.Fatalf("getter: HRESULT %#x", hr)
	}
	if p == nil {
		t.Fatal("getter returned nil")
	}
	defer windows.CoTaskMemFree(unsafe.Pointer(p))
	return windows.UTF16PtrToString(p)
}

func TestEnvironmentOptionsGetters(t *testing.T) {
	o := newEnvironmentOptions("--proxy-server=http://127.0.0.1:1 --lang=ru")

	if got := getString(t, o, o.vtbl.GetAdditionalBrowserArguments); got != "--proxy-server=http://127.0.0.1:1 --lang=ru" {
		t.Errorf("AdditionalBrowserArguments = %q", got)
	}
	if got := getString(t, o, o.vtbl.GetLanguage); got != "" {
		t.Errorf("Language = %q", got)
	}
	if got := getString(t, o, o.vtbl.GetTargetCompatibleBrowserVersion); got != targetCompatibleBrowserVersion {
		t.Errorf("TargetCompatibleBrowserVersion = %q", got)
	}

	lang, _ := windows.UTF16PtrFromString("de")
	if hr := call(o.vtbl.PutLanguage, uintptr(unsafe.Pointer(o)), uintptr(unsafe.Pointer(lang))); hr != sOK {
		t.Fatalf("put_Language: HRESULT %#x", hr)
	}
	if got := getString(t, o, o.vtbl.GetLanguage); got != "de" {
		t.Errorf("Language after put = %q", got)
	}

	var sso int32 = 7
	if hr := call(o.vtbl.GetAllowSingleSignOnUsingOSPrimaryAccount, uintptr(unsafe.Pointer(o)), uintptr(unsafe.Pointer(&sso))); hr != sOK || sso != 0 {
		t.Errorf("AllowSingleSignOnUsingOSPrimaryAccount = %d, HRESULT %#x", sso, hr)
	}
}

func TestEnvironmentOptionsQueryInterface(t *testing.T) {
	o := newEnvironmentOptions("")
	this := uintptr(unsafe.Pointer(o))

	for _, iid := range []windows.GUID{iidIUnknown, iidICoreWebView2EnvironmentOptions} {
		iid := iid
		var out uintptr
		if hr := call(o.vtbl.QueryInterface, this, uintptr(unsafe.Pointer(&iid)), uintptr(unsafe.Pointer(&out))); hr != sOK || out != this {
			t.Errorf("QueryInterface(%v) = %#x, %#x", iid, hr, out)
		}
	}

	// ICoreWebView2EnvironmentOptions2 is not implemented.
	other := windows.GUID{Data1: 0xff85c98a, Data2: 0x1ba7, Data3: 0x4a6b, Data4: [8]byte{0x90, 0xc8, 0x2b, 0x75, 0x2c, 0x89, 0xe9, 0xe2}}
	out := uintptr(1)
	if hr := call(o.vtbl.QueryInterface, this, uintptr(unsafe.Pointer(&other)), uintptr(unsafe.Pointer(&out))); hr != eNoInterface || out != 0 {
		t.Errorf("QueryInterface(other) = %#x, %#x", hr, out)
	}
}
