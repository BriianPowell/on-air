//go:build windows

package windows

import (
	"sync"

	"github.com/go-ole/go-ole"
)

var (
	comInitOnce sync.Once
	comInitErr  error
)

func initCOM() error {
	comInitOnce.Do(func() {
		// MTA: background poll loop has no message pump (STA would need one).
		comInitErr = ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED)
		if comInitErr == nil {
			return
		}
		oleErr, ok := comInitErr.(*ole.OleError)
		if !ok {
			return
		}
		switch oleErr.Code() {
		case 1: // S_FALSE — COM already initialized on this thread
			comInitErr = nil
		case 0x80010106: // RPC_E_CHANGED_MODE
			comInitErr = nil
		}
	})
	return comInitErr
}
