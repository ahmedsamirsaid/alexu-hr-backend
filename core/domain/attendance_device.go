package domain

import "time"

type AttendanceDeviceStatus string

const (
	AttendanceDeviceStatusOnline      AttendanceDeviceStatus = "online"
	AttendanceDeviceStatusOffline     AttendanceDeviceStatus = "offline"
	AttendanceDeviceStatusDeactivated AttendanceDeviceStatus = "deactivated"
)

type AttendanceDevice struct {
	ID           int64
	UID          string
	IP           string
	Port         int
	Name         string
	Location     string
	SerialNumber string
	Status       AttendanceDeviceStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewAttendanceDevice(ip string, port int, name string, location string, serialNumber string) *AttendanceDevice {
	return &AttendanceDevice{
		UID:          GenerateUID("adev"),
		IP:           ip,
		Port:         port,
		Name:         name,
		Location:     location,
		SerialNumber: serialNumber,
		Status:       AttendanceDeviceStatusOffline,
	}
}
