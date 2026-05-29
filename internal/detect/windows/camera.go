//go:build windows

package windows

import (
	"golang.org/x/sys/windows/registry"
)

const webcamConsentKey = `Software\Microsoft\Windows\CurrentVersion\CapabilityAccessManager\ConsentStore\webcam`

func cameraInUse() (bool, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, webcamConsentKey, registry.READ)
	if err != nil {
		return false, err
	}
	defer key.Close()

	for _, subkeyName := range []string{"NonPackaged", "NonPackagedSideloaded"} {
		if active, err := cameraActiveInSubkey(key, subkeyName); err != nil {
			return false, err
		} else if active {
			return true, nil
		}
	}

	return false, nil
}

func cameraActiveInSubkey(parent registry.Key, name string) (bool, error) {
	subkey, err := registry.OpenKey(parent, name, registry.READ)
	if err != nil {
		if err == registry.ErrNotExist {
			return false, nil
		}
		return false, err
	}
	defer subkey.Close()

	names, err := subkey.ReadSubKeyNames(0)
	if err != nil {
		return false, err
	}

	for _, appKey := range names {
		active, err := cameraActiveForApp(subkey, appKey)
		if err != nil {
			return false, err
		}
		if active {
			return true, nil
		}
	}

	return false, nil
}

func cameraActiveForApp(parent registry.Key, name string) (bool, error) {
	appKey, err := registry.OpenKey(parent, name, registry.READ)
	if err != nil {
		if err == registry.ErrNotExist {
			return false, nil
		}
		return false, err
	}
	defer appKey.Close()

	start, _, startErr := appKey.GetIntegerValue("LastUsedTimeStart")
	stop, _, stopErr := appKey.GetIntegerValue("LastUsedTimeStop")

	if startErr != nil && stopErr != nil {
		return false, nil
	}
	if startErr != nil {
		start = 0
	}
	if stopErr != nil {
		stop = 0
	}

	// Windows updates Stop when capture ends. While capture is active, Start
	// is newer than Stop (or Stop has not been written yet).
	return cameraActiveFromTimestamps(start, stop), nil
}

func cameraActiveFromTimestamps(start, stop uint64) bool {
	return start > stop
}
