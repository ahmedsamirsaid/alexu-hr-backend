package usecases

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type ListAttendanceDevicesInput struct {
	Page        int
	PageSize    int
	Status      *domain.AttendanceDeviceStatus
	Search      string
	SearchField string
}

type AttendanceDeviceListItem struct {
	UID               string `json:"uid"`
	IP                string `json:"ip"`
	Port              int    `json:"port"`
	Name              string `json:"name"`
	Location          string `json:"location"`
	SerialNumber      string `json:"serialNumber"`
	Status            string `json:"status"`
	LastStatusKnownAt string `json:"lastStatusKnownAt"`
	MatchedBy         string `json:"matchedBy,omitempty"`
}

type ListAttendanceDevicesOutput struct {
	Devices    []AttendanceDeviceListItem `json:"devices"`
	Total      int                        `json:"total"`
	Page       int                        `json:"page"`
	PageSize   int                        `json:"pageSize"`
	TotalPages int                        `json:"totalPages"`
}

type ListAttendanceDevicesUseCase struct {
	db   ports.DB
	repo ports.AttendanceDeviceRepository
}

func NewListAttendanceDevicesUseCase(db ports.DB, repo ports.AttendanceDeviceRepository) *ListAttendanceDevicesUseCase {
	return &ListAttendanceDevicesUseCase{db: db, repo: repo}
}

func (uc *ListAttendanceDevicesUseCase) Execute(ctx context.Context, input ListAttendanceDevicesInput) (*ListAttendanceDevicesOutput, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 {
		input.PageSize = 20
	}
	if input.PageSize > 100 {
		input.PageSize = 100
	}

	offset := (input.Page - 1) * input.PageSize
	searchField := normalizeAttendanceDeviceSearchField(input.SearchField)
	search := strings.TrimSpace(input.Search)
	filter := ports.AttendanceDeviceListFilter{
		Status:      input.Status,
		Search:      search,
		SearchField: searchField,
	}

	devices, err := uc.repo.List(ctx, uc.db, filter, input.PageSize, offset)
	if err != nil {
		return nil, err
	}

	total, err := uc.repo.CountFiltered(ctx, uc.db, filter)
	if err != nil {
		return nil, err
	}

	items := make([]AttendanceDeviceListItem, len(devices))
	for i, dev := range devices {
		matchedBy := detectAttendanceDeviceMatchedBy(searchField, search, dev)
		items[i] = AttendanceDeviceListItem{
			UID:               dev.UID,
			IP:                dev.IP,
			Port:              dev.Port,
			Name:              dev.Name,
			Location:          dev.Location,
			SerialNumber:      dev.SerialNumber,
			Status:            string(dev.Status),
			LastStatusKnownAt: dev.UpdatedAt.Format(time.RFC3339),
			MatchedBy:         matchedBy,
		}
	}

	totalPages := (total + input.PageSize - 1) / input.PageSize

	return &ListAttendanceDevicesOutput{
		Devices:    items,
		Total:      total,
		Page:       input.Page,
		PageSize:   input.PageSize,
		TotalPages: totalPages,
	}, nil
}

func normalizeAttendanceDeviceSearchField(searchField string) string {
	normalized := strings.ToLower(strings.TrimSpace(searchField))
	switch normalized {
	case "name", "location", "ip", "serial", "port":
		return normalized
	default:
		return "all"
	}
}

func detectAttendanceDeviceMatchedBy(searchField string, search string, device *domain.AttendanceDevice) string {
	needle := strings.ToLower(strings.TrimSpace(search))
	if needle == "" {
		return ""
	}

	matchers := []struct {
		key   string
		value string
	}{
		{key: "name", value: device.Name},
		{key: "location", value: device.Location},
		{key: "ip", value: device.IP},
		{key: "serial", value: device.SerialNumber},
		{key: "port", value: fmt.Sprint(device.Port)},
	}

	if searchField != "all" {
		for _, matcher := range matchers {
			if matcher.key == searchField && strings.Contains(strings.ToLower(matcher.value), needle) {
				return matcher.key
			}
		}
		return ""
	}

	tokenized, _ := parseAttendanceSearchTerms(search)
	if len(tokenized) > 0 {
		for _, field := range []string{"name", "location", "ip", "serial", "port"} {
			for _, token := range tokenized[field] {
				for _, matcher := range matchers {
					if matcher.key == field && strings.Contains(strings.ToLower(matcher.value), strings.ToLower(token)) {
						return field
					}
				}
			}
		}
	}

	for _, matcher := range matchers {
		if strings.Contains(strings.ToLower(matcher.value), needle) {
			return matcher.key
		}
	}

	return ""
}

func parseAttendanceSearchTerms(search string) (map[string][]string, []string) {
	tokenized := map[string][]string{}
	freeTerms := make([]string, 0)

	for _, part := range strings.Fields(search) {
		tokens := strings.SplitN(part, ":", 2)
		if len(tokens) == 2 {
			field := strings.ToLower(strings.TrimSpace(tokens[0]))
			value := strings.TrimSpace(tokens[1])
			if value != "" {
				switch field {
				case "name", "location", "ip", "serial", "port":
					tokenized[field] = append(tokenized[field], value)
					continue
				}
			}
		}

		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			freeTerms = append(freeTerms, trimmed)
		}
	}

	return tokenized, freeTerms
}
