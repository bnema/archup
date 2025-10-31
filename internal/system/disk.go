package system

import (
	"encoding/json"
	"fmt"
	"strings"
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

// Partition represents a disk partition
type Partition struct {
	Name       string `json:"name"`
	Size       string `json:"size"`
	Type       string `json:"type"`
	Mountpoint string `json:"mountpoint"`
	UUID       string `json:"uuid"`
	Label      string `json:"label"`
}

// lsblkOutput represents the JSON output from lsblk
type lsblkOutput struct {
	BlockDevices []struct {
		Name       string `json:"name"`
		Size       string `json:"size"`
		Type       string `json:"type"`
		Model      string `json:"model"`
		Serial     string `json:"serial"`
		Vendor     string `json:"vendor"`
		Mountpoint string `json:"mountpoint"`
		UUID       string `json:"uuid"`
		Label      string `json:"label"`
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

// ListPartitions returns a list of partitions for a given disk
func ListPartitions(diskPath string) ([]Partition, error) {
	diskName := strings.TrimPrefix(diskPath, "/dev/")

	result := RunSimple("lsblk", "-J", "-o", "NAME,SIZE,TYPE,MOUNTPOINT,UUID,LABEL", diskName)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list partitions: %w", result.Error)
	}

	var output lsblkOutput
	if err := json.Unmarshal([]byte(result.Output), &output); err != nil {
		return nil, fmt.Errorf("failed to parse lsblk output: %w", err)
	}

	var partitions []Partition
	for _, dev := range output.BlockDevices {
		if dev.Type == "part" {
			partitions = append(partitions, Partition{
				Name:       dev.Name,
				Size:       dev.Size,
				Type:       dev.Type,
				Mountpoint: dev.Mountpoint,
				UUID:       dev.UUID,
				Label:      dev.Label,
			})
		}
	}

	return partitions, nil
}

// IsDiskMounted checks if any partition on a disk is mounted
func IsDiskMounted(diskPath string) (bool, error) {
	partitions, err := ListPartitions(diskPath)
	if err != nil {
		return false, err
	}

	for _, part := range partitions {
		if part.Mountpoint != "" {
			return true, nil
		}
	}

	return false, nil
}

// WipeDisk wipes the partition table of a disk
func WipeDisk(diskPath string) error {
	result := RunSimple("wipefs", "-a", diskPath)
	if result.Error != nil {
		return fmt.Errorf("failed to wipe disk: %w", result.Error)
	}
	return nil
}

// CreateGPTPartitionTable creates a new GPT partition table
func CreateGPTPartitionTable(diskPath string) error {
	result := RunSimple("parted", "-s", diskPath, "mklabel", "gpt")
	if result.Error != nil {
		return fmt.Errorf("failed to create GPT partition table: %w", result.Error)
	}
	return nil
}
