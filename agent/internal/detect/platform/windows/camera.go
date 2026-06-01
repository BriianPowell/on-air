//go:build windows

package windows

import (
	"golang.org/x/sys/windows/registry"
)

const webcamConsentKey = `Software\Microsoft\Windows\CurrentVersion\CapabilityAccessManager\ConsentStore\webcam`

var webcamContainerSubkeys = []string{
	"NonPackaged",
	"NonPackagedSideloaded",
	"Packaged",
}

func cameraInUse() (bool, error) {
	for _, hive := range []registry.Key{registry.CURRENT_USER, registry.LOCAL_MACHINE} {
		active, err := cameraInUseFromHive(hive)
		if err != nil {
			return false, err
		}
		if active {
			return true, nil
		}
	}
	return false, nil
}

func cameraInUseFromHive(hive registry.Key) (bool, error) {
	key, err := registry.OpenKey(hive, webcamConsentKey, registry.READ)
	if err != nil {
		if err == registry.ErrNotExist {
			return false, nil
		}
		return false, err
	}
	defer key.Close()

	for _, subkeyName := range webcamContainerSubkeys {
		active, err := cameraActiveInSubkey(key, subkeyName)
		if err != nil {
			return false, err
		}
		if active {
			return true, nil
		}
	}

	// New Teams and other MSIX apps register directly under webcam (e.g.
	// MSTeams_8wekyb3d8bbwe), not under NonPackaged.
	names, err := key.ReadSubKeyNames(0)
	if err != nil {
		return false, err
	}

	for _, name := range names {
		if isWebcamContainerSubkey(name) {
			continue
		}
		active, err := cameraActiveForApp(key, name)
		if err != nil {
			return false, err
		}
		if active {
			return true, nil
		}
	}

	return false, nil
}

func isWebcamContainerSubkey(name string) bool {
	for _, container := range webcamContainerSubkeys {
		if name == container {
			return true
		}
	}
	return false
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
