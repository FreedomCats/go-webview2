//go:build windows
// +build windows

package edge

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

// ICoreWebView2EnvironmentOptions is implemented on the Go side and passed to
// CreateCoreWebView2EnvironmentWithOptions. Only the base interface is
// implemented; the runtime asks for newer versions via QueryInterface and
// falls back to defaults when they are not provided.

var (
	iidIUnknown                        = windows.GUID{Data1: 0x00000000, Data2: 0x0000, Data3: 0x0000, Data4: [8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}
	iidICoreWebView2EnvironmentOptions = windows.GUID{Data1: 0x2fde08a8, Data2: 0x1e9a, Data3: 0x4766, Data4: [8]byte{0x8c, 0x05, 0x95, 0xa9, 0xce, 0xb9, 0xd1, 0xc5}}

	modole32           = windows.NewLazySystemDLL("ole32.dll")
	procCoTaskMemAlloc = modole32.NewProc("CoTaskMemAlloc")
)

const (
	sOK          = 0
	eNoInterface = 0x80004002
	eOutOfMemory = 0x8007000E
	ePointer     = 0x80004003
)

// targetCompatibleBrowserVersion is the minimum runtime version this binding
// works with; the loader picks an installed runtime at least this new.
const targetCompatibleBrowserVersion = "86.0.616.0"

type environmentOptionsVtbl struct {
	_IUnknownVtbl
	GetAdditionalBrowserArguments             ComProc
	PutAdditionalBrowserArguments             ComProc
	GetLanguage                               ComProc
	PutLanguage                               ComProc
	GetTargetCompatibleBrowserVersion         ComProc
	PutTargetCompatibleBrowserVersion         ComProc
	GetAllowSingleSignOnUsingOSPrimaryAccount ComProc
	PutAllowSingleSignOnUsingOSPrimaryAccount ComProc
}

type environmentOptions struct {
	vtbl *environmentOptionsVtbl

	additionalBrowserArguments     string
	language                       string
	targetCompatibleBrowserVersion string
	allowSSO                       int32
}

func newEnvironmentOptions(additionalBrowserArguments string) *environmentOptions {
	return &environmentOptions{
		vtbl:                           &environmentOptionsFn,
		additionalBrowserArguments:     additionalBrowserArguments,
		targetCompatibleBrowserVersion: targetCompatibleBrowserVersion,
	}
}

// coTaskMemString returns s as a NUL-terminated UTF-16 string allocated with
// CoTaskMemAlloc: the caller of a COM getter frees it with CoTaskMemFree.
func coTaskMemString(s string, out **uint16) uintptr {
	if out == nil {
		return ePointer
	}
	u, err := windows.UTF16FromString(s)
	if err != nil {
		u = []uint16{0}
	}
	size := uintptr(len(u)) * 2
	p, _, _ := procCoTaskMemAlloc.Call(size)
	if p == 0 {
		*out = nil
		return eOutOfMemory
	}
	// The memory is not managed by Go, so converting the address back to a
	// pointer is safe; reading it through &p keeps vet quiet about it.
	mem := *(*unsafe.Pointer)(unsafe.Pointer(&p))
	copy((*[1 << 29]uint16)(mem)[:len(u):len(u)], u)
	*out = (*uint16)(mem)
	return sOK
}

func _EnvironmentOptionsQueryInterface(this *environmentOptions, riid *windows.GUID, ppv *uintptr) uintptr {
	if ppv == nil {
		return ePointer
	}
	if riid != nil && (*riid == iidIUnknown || *riid == iidICoreWebView2EnvironmentOptions) {
		*ppv = uintptr(unsafe.Pointer(this))
		return sOK
	}
	*ppv = 0
	return eNoInterface
}

// The object is owned by Chromium and lives as long as it does, so the
// reference count is not tracked.
func _EnvironmentOptionsAddRef(this *environmentOptions) uintptr  { return 1 }
func _EnvironmentOptionsRelease(this *environmentOptions) uintptr { return 1 }

func _EnvironmentOptionsGetAdditionalBrowserArguments(this *environmentOptions, value **uint16) uintptr {
	return coTaskMemString(this.additionalBrowserArguments, value)
}

func _EnvironmentOptionsPutAdditionalBrowserArguments(this *environmentOptions, value *uint16) uintptr {
	this.additionalBrowserArguments = windows.UTF16PtrToString(value)
	return sOK
}

func _EnvironmentOptionsGetLanguage(this *environmentOptions, value **uint16) uintptr {
	return coTaskMemString(this.language, value)
}

func _EnvironmentOptionsPutLanguage(this *environmentOptions, value *uint16) uintptr {
	this.language = windows.UTF16PtrToString(value)
	return sOK
}

func _EnvironmentOptionsGetTargetCompatibleBrowserVersion(this *environmentOptions, value **uint16) uintptr {
	return coTaskMemString(this.targetCompatibleBrowserVersion, value)
}

func _EnvironmentOptionsPutTargetCompatibleBrowserVersion(this *environmentOptions, value *uint16) uintptr {
	this.targetCompatibleBrowserVersion = windows.UTF16PtrToString(value)
	return sOK
}

func _EnvironmentOptionsGetAllowSingleSignOnUsingOSPrimaryAccount(this *environmentOptions, value *int32) uintptr {
	if value == nil {
		return ePointer
	}
	*value = this.allowSSO
	return sOK
}

func _EnvironmentOptionsPutAllowSingleSignOnUsingOSPrimaryAccount(this *environmentOptions, value int32) uintptr {
	this.allowSSO = value
	return sOK
}

var environmentOptionsFn = environmentOptionsVtbl{
	_IUnknownVtbl{
		NewComProc(_EnvironmentOptionsQueryInterface),
		NewComProc(_EnvironmentOptionsAddRef),
		NewComProc(_EnvironmentOptionsRelease),
	},
	NewComProc(_EnvironmentOptionsGetAdditionalBrowserArguments),
	NewComProc(_EnvironmentOptionsPutAdditionalBrowserArguments),
	NewComProc(_EnvironmentOptionsGetLanguage),
	NewComProc(_EnvironmentOptionsPutLanguage),
	NewComProc(_EnvironmentOptionsGetTargetCompatibleBrowserVersion),
	NewComProc(_EnvironmentOptionsPutTargetCompatibleBrowserVersion),
	NewComProc(_EnvironmentOptionsGetAllowSingleSignOnUsingOSPrimaryAccount),
	NewComProc(_EnvironmentOptionsPutAllowSingleSignOnUsingOSPrimaryAccount),
}
