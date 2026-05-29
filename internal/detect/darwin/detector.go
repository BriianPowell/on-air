//go:build darwin

package darwin

/*
#cgo LDFLAGS: -framework CoreAudio -framework CoreFoundation -framework CoreMediaIO

#include <CoreAudio/CoreAudio.h>
#include <CoreMediaIO/CMIOHardware.h>
#include <CoreFoundation/CoreFoundation.h>
#include <stdlib.h>

static int isIgnoredCameraUID(CFStringRef uid) {
	if (uid == NULL) {
		return 1;
	}
	if (CFStringCompare(uid, CFSTR("obs-virtual-cam-device"), kCFCompareCaseInsensitive) == kCFCompareEqualTo) {
		return 1;
	}
	return 0;
}

static int cameraInUse(void) {
	CMIOObjectPropertyAddress propertyAddress = {
		kCMIOHardwarePropertyDevices,
		kCMIOObjectPropertyScopeGlobal,
		kCMIOObjectPropertyElementMain
	};

	UInt32 dataSize = 0;
	OSStatus status = CMIOObjectGetPropertyDataSize(
		kCMIOObjectSystemObject,
		&propertyAddress,
		0,
		NULL,
		&dataSize
	);
	if (status != noErr || dataSize == 0) {
		return 0;
	}

	CMIOObjectID *devices = (CMIOObjectID *)malloc(dataSize);
	if (devices == NULL) {
		return 0;
	}

	UInt32 dataUsed = 0;
	status = CMIOObjectGetPropertyData(
		kCMIOObjectSystemObject,
		&propertyAddress,
		0,
		NULL,
		dataSize,
		&dataUsed,
		devices
	);
	if (status != noErr) {
		free(devices);
		return 0;
	}

	int deviceCount = (int)(dataSize / sizeof(CMIOObjectID));
	int inUse = 0;

	for (int i = 0; i < deviceCount; i++) {
		propertyAddress.mSelector = kCMIODevicePropertyDeviceUID;
		propertyAddress.mScope = kCMIOObjectPropertyScopeWildcard;
		propertyAddress.mElement = kCMIOObjectPropertyElementWildcard;

		dataSize = 0;
		status = CMIOObjectGetPropertyDataSize(devices[i], &propertyAddress, 0, NULL, &dataSize);
		if (status != noErr || dataSize == 0) {
			continue;
		}

		CFStringRef uid = NULL;
		dataUsed = 0;
		status = CMIOObjectGetPropertyData(devices[i], &propertyAddress, 0, NULL, dataSize, &dataUsed, &uid);
		if (status != noErr || uid == NULL || isIgnoredCameraUID(uid)) {
			if (uid != NULL) {
				CFRelease(uid);
			}
			continue;
		}
		CFRelease(uid);

		propertyAddress.mSelector = kCMIODevicePropertyDeviceIsRunningSomewhere;
		propertyAddress.mScope = kCMIOObjectPropertyScopeWildcard;
		propertyAddress.mElement = kCMIOObjectPropertyElementWildcard;

		UInt32 isRunning = 0;
		dataSize = sizeof(isRunning);
		dataUsed = 0;
		status = CMIOObjectGetPropertyData(devices[i], &propertyAddress, 0, NULL, dataSize, &dataUsed, &isRunning);
		if (status == noErr && isRunning) {
			inUse = 1;
			break;
		}
	}

	free(devices);
	return inUse;
}

static int micInUseFromProcesses(void) {
	AudioObjectPropertyAddress address = {
		kAudioHardwarePropertyProcessObjectList,
		kAudioObjectPropertyScopeGlobal,
		kAudioObjectPropertyElementMain
	};

	UInt32 dataSize = 0;
	OSStatus status = AudioObjectGetPropertyDataSize(
		kAudioObjectSystemObject,
		&address,
		0,
		NULL,
		&dataSize
	);
	if (status != noErr) {
		return -1;
	}
	if (dataSize == 0) {
		return 0;
	}

	AudioObjectID *processes = (AudioObjectID *)malloc(dataSize);
	if (processes == NULL) {
		return -1;
	}

	status = AudioObjectGetPropertyData(
		kAudioObjectSystemObject,
		&address,
		0,
		NULL,
		&dataSize,
		processes
	);
	if (status != noErr) {
		free(processes);
		return -1;
	}

	int processCount = (int)(dataSize / sizeof(AudioObjectID));
	int inUse = 0;

	for (int i = 0; i < processCount; i++) {
		AudioObjectPropertyAddress runningInputAddress = {
			kAudioProcessPropertyIsRunningInput,
			kAudioObjectPropertyScopeGlobal,
			kAudioObjectPropertyElementMain
		};

		UInt32 isRunning = 0;
		UInt32 runningSize = sizeof(isRunning);
		status = AudioObjectGetPropertyData(
			processes[i],
			&runningInputAddress,
			0,
			NULL,
			&runningSize,
			&isRunning
		);
		if (status == noErr && isRunning) {
			inUse = 1;
			break;
		}
	}

	free(processes);
	return inUse;
}

static int micInUseFromDevices(void) {
	AudioObjectPropertyAddress address = {
		kAudioHardwarePropertyDevices,
		kAudioObjectPropertyScopeGlobal,
		kAudioObjectPropertyElementMain
	};

	UInt32 dataSize = 0;
	OSStatus status = AudioObjectGetPropertyDataSize(
		kAudioObjectSystemObject,
		&address,
		0,
		NULL,
		&dataSize
	);
	if (status != noErr || dataSize == 0) {
		return 0;
	}

	AudioDeviceID *devices = (AudioDeviceID *)malloc(dataSize);
	if (devices == NULL) {
		return 0;
	}

	status = AudioObjectGetPropertyData(
		kAudioObjectSystemObject,
		&address,
		0,
		NULL,
		&dataSize,
		devices
	);
	if (status != noErr) {
		free(devices);
		return 0;
	}

	int deviceCount = (int)(dataSize / sizeof(AudioDeviceID));
	int inUse = 0;

	for (int i = 0; i < deviceCount; i++) {
		address.mSelector = kAudioDevicePropertyStreamConfiguration;
		address.mScope = kAudioDevicePropertyScopeInput;
		address.mElement = kAudioObjectPropertyElementMain;

		dataSize = 0;
		status = AudioObjectGetPropertyDataSize(devices[i], &address, 0, NULL, &dataSize);
		if (status != noErr || dataSize == 0) {
			continue;
		}

		AudioBufferList *bufferList = (AudioBufferList *)malloc(dataSize);
		if (bufferList == NULL) {
			continue;
		}

		status = AudioObjectGetPropertyData(devices[i], &address, 0, NULL, &dataSize, bufferList);
		if (status != noErr) {
			free(bufferList);
			continue;
		}

		UInt32 inputChannels = 0;
		for (UInt32 j = 0; j < bufferList->mNumberBuffers; j++) {
			inputChannels += bufferList->mBuffers[j].mNumberChannels;
		}
		free(bufferList);

		if (inputChannels == 0) {
			continue;
		}

		address.mSelector = kAudioDevicePropertyDeviceIsRunning;
		address.mScope = kAudioDevicePropertyScopeInput;

		UInt32 isRunning = 0;
		dataSize = sizeof(isRunning);
		status = AudioObjectGetPropertyData(devices[i], &address, 0, NULL, &dataSize, &isRunning);
		if (status == noErr && isRunning) {
			inUse = 1;
			break;
		}

		address.mSelector = kAudioDevicePropertyDeviceIsRunningSomewhere;
		address.mScope = kAudioObjectPropertyScopeGlobal;

		isRunning = 0;
		dataSize = sizeof(isRunning);
		status = AudioObjectGetPropertyData(devices[i], &address, 0, NULL, &dataSize, &isRunning);
		if (status == noErr && isRunning) {
			inUse = 1;
			break;
		}
	}

	free(devices);
	return inUse;
}

static int micInUse(void) {
	int fromProcesses = micInUseFromProcesses();
	if (fromProcesses >= 0) {
		return fromProcesses;
	}
	return micInUseFromDevices();
}
*/
import "C"

import (
	"github.com/brianpowell/on-air/internal/detect"
)

type Detector struct{}

func New() *Detector {
	return &Detector{}
}

func (d *Detector) Poll() (detect.Status, error) {
	return detect.Status{
		MicActive:    C.micInUse() != 0,
		CameraActive: C.cameraInUse() != 0,
	}, nil
}
