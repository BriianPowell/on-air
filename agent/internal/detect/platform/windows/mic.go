//go:build windows

package windows

import (
	"github.com/go-ole/go-ole"
	"github.com/moutend/go-wca/pkg/wca"
)

func micInUse() (bool, error) {
	if err := initCOM(); err != nil {
		return false, err
	}

	var mmde *wca.IMMDeviceEnumerator
	if err := wca.CoCreateInstance(
		wca.CLSID_MMDeviceEnumerator,
		0,
		ole.CLSCTX_ALL,
		wca.IID_IMMDeviceEnumerator,
		&mmde,
	); err != nil {
		return false, err
	}
	defer mmde.Release()

	var devices *wca.IMMDeviceCollection
	if err := mmde.EnumAudioEndpoints(wca.ECapture, wca.DEVICE_STATE_ACTIVE, &devices); err != nil {
		return false, err
	}
	defer devices.Release()

	var count uint32
	if err := devices.GetCount(&count); err != nil {
		return false, err
	}

	for i := uint32(0); i < count; i++ {
		var device *wca.IMMDevice
		if err := devices.Item(i, &device); err != nil {
			continue
		}

		active, err := micActiveOnDevice(device)
		device.Release()
		if err != nil {
			continue
		}
		if active {
			return true, nil
		}
	}

	return false, nil
}

func micActiveOnDevice(device *wca.IMMDevice) (bool, error) {
	var sessionManager *wca.IAudioSessionManager2
	if err := device.Activate(
		wca.IID_IAudioSessionManager2,
		ole.CLSCTX_ALL,
		nil,
		&sessionManager,
	); err != nil {
		return false, err
	}
	defer sessionManager.Release()

	var sessionEnum *wca.IAudioSessionEnumerator
	if err := sessionManager.GetSessionEnumerator(&sessionEnum); err != nil {
		return false, err
	}
	defer sessionEnum.Release()

	var sessionCount int
	if err := sessionEnum.GetCount(&sessionCount); err != nil {
		return false, err
	}

	for i := 0; i < sessionCount; i++ {
		var sessionControl *wca.IAudioSessionControl
		if err := sessionEnum.GetSession(i, &sessionControl); err != nil {
			continue
		}

		var state uint32
		if err := sessionControl.GetState(&state); err != nil {
			sessionControl.Release()
			continue
		}
		sessionControl.Release()

		if state == wca.AudioSessionStateActive {
			return true, nil
		}
	}

	return false, nil
}
