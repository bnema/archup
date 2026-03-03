package system

import (
	"encoding/json"
	"fmt"
)

// Disk represents a storage device
type Disk struct {
	Name   string `json:"name"`
	Size   string `json:"size"`
	Type   string `json:"type"`
	Model  string `json:"model"`
	Serial string `json:"serial"`
	Vendor string `json:"vendor"`
	Path   string `json:"path"`
}

// lsblkOutput represents the JSON output from lsblk
type lsblkOutput struct {
	BlockDevices []struct {
		Name   string `json:"name"`
		Size   string `json:"size"`
		Type   string `json:"type"`
		Model  string `json:"model"`
		Serial string `json:"serial"`
		Vendor string `json:"vendor"`
	} `json:"blockdevices"`
}

// ListDisks returns a list of available disks (excluding loop devices)
func ListDisks() ([]Disk, error) {
	result := RunSimple("lsblk", "-J", "-o", "NAME,SIZE,TYPE,MODEL,SERIAL,VENDOR", "-d", "-e", "7")
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list disks: %w", result.Error)
	}

	var output lsblkOutput
	if err := json.Unmarshal([]byte(result.Output), &output); err != nil {
		return nil, fmt.Errorf("failed to parse lsblk output: %w", err)
	}

	var disks []Disk
	for _, dev := range output.BlockDevices {
		if dev.Type == "disk" {
			disks = append(disks, Disk{
				Name:   dev.Name,
				Size:   dev.Size,
				Type:   dev.Type,
				Model:  dev.Model,
				Serial: dev.Serial,
				Vendor: dev.Vendor,
				Path:   "/dev/" + dev.Name,
			})
		}
	}

	return disks, nil
}
