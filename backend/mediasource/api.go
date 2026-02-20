// backend/mediasource/api.go
//
// This file defines the public API of the mediasource Manager module.
// For this module, the public API is limited to CLI utilities, as all
// runtime inter-module communication is handled via the EventBus.
//
// Contents:
// - Public CLI Utilities

package mediasource

import (
	"fmt"

	"github.com/pion/mediadevices"
)

// ============================================================================
// Public CLI Utilities
// ============================================================================

// ShowAllDevices enumerates and prints all available media devices to stdout.
// This is used by the -list-devices CLI flag and is not intended for use
// by other modules during runtime.
func ShowAllDevices() {
	fmt.Println("----------- Available Media Devices -----------")
	devices := mediadevices.EnumerateDevices()

	if len(devices) == 0 {
		fmt.Println("No media devices found.")
		fmt.Println("Note: You may need to run with appropriate permissions.")
		return
	}

	fmt.Println("INFO: Use the 'Friendly Name' or 'DeviceID' for your config.")
	fmt.Println()

	var videoDevices, audioDevices []mediadevices.MediaDeviceInfo
	for _, device := range devices {
		if device.Kind == mediadevices.VideoInput {
			videoDevices = append(videoDevices, device)
		} else if device.Kind == mediadevices.AudioInput {
			audioDevices = append(audioDevices, device)
		}
	}

	if len(videoDevices) > 0 {
		fmt.Println("VIDEO DEVICES:")
		for i, device := range videoDevices {
			fmt.Printf("  #%d:\n", i+1)
			fmt.Printf("    Friendly Name : %s\n", decodeDeviceLabel(device.Label))
			fmt.Printf("    DeviceID      : %s\n", device.DeviceID)
			fmt.Println()
		}
	}

	if len(audioDevices) > 0 {
		fmt.Println("AUDIO DEVICES:")
		for i, device := range audioDevices {
			fmt.Printf("  #%d:\n", i+1)
			fmt.Printf("    Friendly Name : %s\n", decodeDeviceLabel(device.Label))
			fmt.Printf("    DeviceID      : %s\n", device.DeviceID)
			fmt.Println()
		}
	}
	fmt.Println("----------------------------------------------")
}

