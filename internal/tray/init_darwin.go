//go:build darwin

package tray

import (
	"errors"
	"runtime"
	"sync"
	"unsafe"

	"github.com/go-webgpu/goffi/ffi"
	"github.com/go-webgpu/goffi/types"
)

var nsAppInit struct {
	once sync.Once
	err  error
}

// ensureNSApplicationInitialized creates the shared NSApplication before any
// AppKit/NSStatusBar calls. On some macOS configurations (observed on Intel
// macOS 15) calling [NSStatusBar systemStatusBar] before [NSApplication
// sharedApplication] aborts the process with:
//
//   Assertion failed: (CGAtomicGet(&is_initialized)), function CGSConnectionByID ...
func ensureNSApplicationInitialized() error {
	nsAppInit.once.Do(func() {
		nsAppInit.err = initNSApplication()
	})
	return nsAppInit.err
}

func initNSApplication() error {
	libobjc, err := ffi.LoadLibrary("/usr/lib/libobjc.A.dylib")
	if err != nil {
		return err
	}
	if _, err := ffi.LoadLibrary("/System/Library/Frameworks/Foundation.framework/Foundation"); err != nil {
		return err
	}
	if _, err := ffi.LoadLibrary("/System/Library/Frameworks/AppKit.framework/AppKit"); err != nil {
		return err
	}

	objcGetClass, err := ffi.GetSymbol(libobjc, "objc_getClass")
	if err != nil {
		return err
	}
	selRegisterName, err := ffi.GetSymbol(libobjc, "sel_registerName")
	if err != nil {
		return err
	}
	objcMsgSend, err := ffi.GetSymbol(libobjc, "objc_msgSend")
	if err != nil {
		return err
	}

	// CIF for functions that take a C string and return a pointer.
	cifCStringToPtr := &types.CallInterface{}
	if err := ffi.PrepareCallInterface(
		cifCStringToPtr,
		types.DefaultCall,
		types.PointerTypeDescriptor,
		[]*types.TypeDescriptor{types.PointerTypeDescriptor},
	); err != nil {
		return err
	}

	nsAppClassName := append([]byte("NSApplication"), 0)
	nsAppClassNamePtr := unsafe.Pointer(&nsAppClassName[0])
	var nsAppClass uintptr
	if err := ffi.CallFunction(
		cifCStringToPtr,
		objcGetClass,
		unsafe.Pointer(&nsAppClass),
		[]unsafe.Pointer{unsafe.Pointer(&nsAppClassNamePtr)},
	); err != nil {
		runtime.KeepAlive(nsAppClassName)
		return err
	}
	runtime.KeepAlive(nsAppClassName)
	if nsAppClass == 0 {
		return errors.New("unable to find NSApplication class")
	}

	sharedAppSelName := append([]byte("sharedApplication"), 0)
	sharedAppSelNamePtr := unsafe.Pointer(&sharedAppSelName[0])
	var sharedAppSel uintptr
	if err := ffi.CallFunction(
		cifCStringToPtr,
		selRegisterName,
		unsafe.Pointer(&sharedAppSel),
		[]unsafe.Pointer{unsafe.Pointer(&sharedAppSelNamePtr)},
	); err != nil {
		runtime.KeepAlive(sharedAppSelName)
		return err
	}
	runtime.KeepAlive(sharedAppSelName)
	if sharedAppSel == 0 {
		return errors.New("unable to register sharedApplication selector")
	}

	// CIF for objc_msgSend(self, _cmd) -> pointer
	cifMsgSend := &types.CallInterface{}
	if err := ffi.PrepareCallInterface(
		cifMsgSend,
		types.DefaultCall,
		types.PointerTypeDescriptor,
		[]*types.TypeDescriptor{
			types.PointerTypeDescriptor,
			types.PointerTypeDescriptor,
		},
	); err != nil {
		return err
	}

	var nsApp uintptr
	if err := ffi.CallFunction(
		cifMsgSend,
		objcMsgSend,
		unsafe.Pointer(&nsApp),
		[]unsafe.Pointer{
			unsafe.Pointer(&nsAppClass),
			unsafe.Pointer(&sharedAppSel),
		},
	); err != nil {
		return err
	}
	if nsApp == 0 {
		return errors.New("[NSApplication sharedApplication] returned nil")
	}

	return nil
}
