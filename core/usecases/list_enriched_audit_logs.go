package usecases

import (
	"context"
	"encoding/json"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

// EnrichedAuditLogDTO represents an audit log entry with resolved names and localized labels.
type EnrichedAuditLogDTO struct {
	UID               string                 `json:"uid"`
	ActorUID          string                 `json:"actorUid"`
	ActorName         *string                `json:"actorName"`         // Resolved actor name
	ActorType         string                 `json:"actorType"`         // "user" or "system"
	Action            string                 `json:"action"`            // Raw key
	ActionLabel       string                 `json:"actionLabel"`       // Translated based on request locale
	EntityType        string                 `json:"entityType"`        // Raw key
	EntityTypeLabel   string                 `json:"entityTypeLabel"`   // Translated based on request locale
	EntityUID         string                 `json:"entityUid"`
	EntityDescription *string                `json:"entityDescription"` // "Ahmed Al-Sayed - Annual Leave (Jan 15-20)"
	Meta              map[string]interface{} `json:"meta"`              // Translated field names
	OccurredAt        string                 `json:"occurredAt"`
	CreatedAt         string                 `json:"createdAt"`
	IsSystemAction    bool                   `json:"isSystemAction"` // True for auto-generated actions
}

// ListEnrichedAuditLogsInput defines the input for listing enriched audit logs.
type ListEnrichedAuditLogsInput struct {
	// Existing filters
	EntityType string
	ActorUID   string
	SearchText string

	// New filters
	ActorName   string     // Partial name match
	ActionTypes []string   // Filter by action types
	EntityTypes []string   // Filter by entity types
	StartDate   *time.Time // Filter by date range start
	EndDate     *time.Time // Filter by date range end

	// Pagination
	Page     int
	PageSize int
}

// ListEnrichedAuditLogsOutput contains the enriched audit log results.
type ListEnrichedAuditLogsOutput struct {
	Events      []EnrichedAuditLogDTO `json:"events"`
	Page        int                   `json:"page"`
	PageSize    int                   `json:"pageSize"`
	TotalCount  int                   `json:"totalCount"`
	HasNextPage bool                  `json:"hasNextPage"`
}

// ListEnrichedAuditLogsUseCase handles retrieving enriched audit logs with resolved names.
type ListEnrichedAuditLogsUseCase struct {
	auditRepo     ports.AuditLogRepository
	employeeRepo  ports.EmployeeRepository
	deptRepo      ports.DepartmentRepository
	leaveReqRepo  ports.LeaveRequestRepository
	leaveTypeRepo ports.LeaveTypeRepository
	attRecordRepo ports.AttendanceRecordRepository
	db            ports.DB
	i18n          ports.I18nService
}

// NewListEnrichedAuditLogsUseCase creates a new enriched audit logs use case.
func NewListEnrichedAuditLogsUseCase(
	auditRepo ports.AuditLogRepository,
	employeeRepo ports.EmployeeRepository,
	deptRepo ports.DepartmentRepository,
	leaveReqRepo ports.LeaveRequestRepository,
	leaveTypeRepo ports.LeaveTypeRepository,
	attRecordRepo ports.AttendanceRecordRepository,
	db ports.DB,
	i18n ports.I18nService,
) *ListEnrichedAuditLogsUseCase {
	return &ListEnrichedAuditLogsUseCase{
		auditRepo:     auditRepo,
		employeeRepo:  employeeRepo,
		deptRepo:      deptRepo,
		leaveReqRepo:  leaveReqRepo,
		leaveTypeRepo: leaveTypeRepo,
		attRecordRepo: attRecordRepo,
		db:            db,
		i18n:          i18n,
	}
}

// Execute retrieves enriched audit logs with resolved names and bilingual labels.
func (uc *ListEnrichedAuditLogsUseCase) Execute(ctx context.Context, input ListEnrichedAuditLogsInput) (*ListEnrichedAuditLogsOutput, error) {
	// 1. Build filters for repository - always hide system actions
	filters := ports.AuditLogFilters{
		EntityType:        input.EntityType,
		EntityTypes:       input.EntityTypes,
		ActorUID:          input.ActorUID,
		ActorName:         input.ActorName,
		ActionTypes:       input.ActionTypes,
		SearchText:        input.SearchText,
		StartDate:         input.StartDate,
		EndDate:           input.EndDate,
		HideSystemActions: true, // Always hide system actions
	}

	params := ports.ListParams{
		Page:     input.Page,
		PageSize: input.PageSize,
	}

	// 2. Query audit logs with filters
	logs, err := uc.auditRepo.ListAll(ctx, uc.db, filters, params)
	if err != nil {
		return nil, err
	}

	// 3. Get total count for pagination
	totalCount, err := uc.auditRepo.CountAll(ctx, uc.db, filters)
	if err != nil {
		return nil, err
	}

	// 4. Extract unique actor UIDs and entity UIDs by type
	actorUIDs := extractUniqueActorUIDs(logs)
	entityUIDsByType := extractEntityUIDsByType(logs)

	// 5. Batch fetch actors and entities
	actorMap, err := uc.fetchActors(ctx, actorUIDs)
	if err != nil {
		return nil, err
	}

	entityDescMap, err := uc.fetchEntityDescriptions(ctx, entityUIDsByType)
	if err != nil {
		return nil, err
	}

	// 6. Build enriched DTOs with resolved names
	events := make([]EnrichedAuditLogDTO, 0, len(logs))
	for _, log := range logs {
		enriched := uc.enrichAuditLog(ctx, log, actorMap, entityDescMap)
		events = append(events, enriched)
	}

	// 7. Calculate pagination metadata
	hasNextPage := len(logs) == input.PageSize

	return &ListEnrichedAuditLogsOutput{
		Events:      events,
		Page:        input.Page,
		PageSize:    input.PageSize,
		TotalCount:  totalCount,
		HasNextPage: hasNextPage,
	}, nil
}

// extractUniqueActorUIDs extracts unique actor UIDs from audit logs.
func extractUniqueActorUIDs(logs []*domain.AuditLog) []string {
	uidSet := make(map[string]bool)
	for _, log := range logs {
		uidSet[log.ActorUID] = true
	}

	uids := make([]string, 0, len(uidSet))
	for uid := range uidSet {
		uids = append(uids, uid)
	}
	return uids
}

// extractEntityUIDsByType extracts unique entity UIDs grouped by entity type.
func extractEntityUIDsByType(logs []*domain.AuditLog) map[string][]string {
	uidsByType := make(map[string]map[string]bool)

	for _, log := range logs {
		if uidsByType[log.EntityType] == nil {
			uidsByType[log.EntityType] = make(map[string]bool)
		}
		uidsByType[log.EntityType][log.EntityUID] = true
	}

	result := make(map[string][]string)
	for entityType, uidSet := range uidsByType {
		uids := make([]string, 0, len(uidSet))
		for uid := range uidSet {
			uids = append(uids, uid)
		}
		result[entityType] = uids
	}

	return result
}

// fetchActors batch fetches employees and builds actor name lookup map.
func (uc *ListEnrichedAuditLogsUseCase) fetchActors(ctx context.Context, actorUIDs []string) (map[string]*domain.Employee, error) {
	if len(actorUIDs) == 0 {
		return make(map[string]*domain.Employee), nil
	}

	employees, err := uc.employeeRepo.GetByUIDs(ctx, uc.db, actorUIDs)
	if err != nil {
		return nil, err
	}

	actorMap := make(map[string]*domain.Employee)
	for _, emp := range employees {
		actorMap[emp.UID] = emp
	}

	return actorMap, nil
}

// fetchEntityDescriptions batch fetches entities and builds description lookup map.
func (uc *ListEnrichedAuditLogsUseCase) fetchEntityDescriptions(ctx context.Context, entityUIDsByType map[string][]string) (map[string]string, error) {
	descMap := make(map[string]string)

	// Fetch employees for employee entities
	if employeeUIDs, ok := entityUIDsByType["employee"]; ok && len(employeeUIDs) > 0 {
		employees, err := uc.employeeRepo.GetByUIDs(ctx, uc.db, employeeUIDs)
		if err != nil {
			return nil, err
		}
		for _, emp := range employees {
			descMap[emp.UID] = emp.Name
		}
	}

	// Fetch departments for department entities
	if deptUIDs, ok := entityUIDsByType["department"]; ok && len(deptUIDs) > 0 {
		departments, err := uc.deptRepo.GetByUIDs(ctx, uc.db, deptUIDs)
		if err != nil {
			return nil, err
		}
		for _, dept := range departments {
			descMap[dept.UID] = formatDepartmentDescription(dept)
		}
	}

	// Fetch attendance records for attendance_record entities
	if attUIDs, ok := entityUIDsByType["attendance_record"]; ok && len(attUIDs) > 0 {
		records, err := uc.attRecordRepo.GetByUIDs(ctx, uc.db, attUIDs)
		if err != nil {
			return nil, err
		}

		// Fetch employee names for attendance records
		empUIDs := make([]string, 0, len(records))
		for _, rec := range records {
			if rec.EmployeeUID != "" {
				empUIDs = append(empUIDs, rec.EmployeeUID)
			}
		}
		empMap := make(map[string]string)
		if len(empUIDs) > 0 {
			employees, err := uc.employeeRepo.GetByUIDs(ctx, uc.db, empUIDs)
			if err != nil {
				return nil, err
			}
			for _, emp := range employees {
				empMap[emp.UID] = emp.Name
			}
		}

		for _, rec := range records {
			descMap[rec.UID] = formatAttendanceRecordDescription(rec, empMap)
		}
	}

	// Fetch leave requests for leave_request entities
	if leaveReqUIDs, ok := entityUIDsByType["leave_request"]; ok && len(leaveReqUIDs) > 0 {
		requests, err := uc.leaveReqRepo.GetByUIDs(ctx, uc.db, leaveReqUIDs)
		if err != nil {
			return nil, err
		}

		// Fetch employee names and leave type names
		empUIDs := make([]string, 0, len(requests))
		leaveTypeUIDs := make([]string, 0, len(requests))
		for _, req := range requests {
			empUIDs = append(empUIDs, req.EmployeeUID)
			leaveTypeUIDs = append(leaveTypeUIDs, req.LeaveTypeUID)
		}

		empMap := make(map[string]string)
		if len(empUIDs) > 0 {
			employees, err := uc.employeeRepo.GetByUIDs(ctx, uc.db, empUIDs)
			if err != nil {
				return nil, err
			}
			for _, emp := range employees {
				empMap[emp.UID] = emp.Name
			}
		}

		leaveTypeMap := make(map[string]string)
		for _, ltUID := range leaveTypeUIDs {
			leaveType, err := uc.leaveTypeRepo.GetByUID(ctx, uc.db, ltUID)
			if err == nil && leaveType != nil {
				leaveTypeMap[ltUID] = leaveType.NameEN
			}
		}

		for _, req := range requests {
			descMap[req.UID] = formatLeaveRequestDescription(req, empMap, leaveTypeMap)
		}
	}

	// Fetch leave types for leave_type entities
	if leaveTypeUIDs, ok := entityUIDsByType["leave_type"]; ok && len(leaveTypeUIDs) > 0 {
		for _, ltUID := range leaveTypeUIDs {
			leaveType, err := uc.leaveTypeRepo.GetByUID(ctx, uc.db, ltUID)
			if err == nil && leaveType != nil {
				descMap[ltUID] = leaveType.NameEN + " (" + leaveType.Code + ")"
			}
		}
	}

	return descMap, nil
}

// formatDepartmentDescription formats a department description.
func formatDepartmentDescription(dept *domain.Department) string {
	return dept.NameEN + " (" + dept.Code + ")"
}

// formatAttendanceRecordDescription formats an attendance record description.
func formatAttendanceRecordDescription(record *domain.AttendanceRecord, empMap map[string]string) string {
	empName := "Unknown"
	if record.EmployeeUID != "" {
		if name, ok := empMap[record.EmployeeUID]; ok {
			empName = name
		}
	}
	date := record.PunchedAt.Format("Jan 2, 2006")
	return empName + " - " + date
}

// formatLeaveRequestDescription formats a leave request description.
func formatLeaveRequestDescription(req *domain.LeaveRequest, empMap map[string]string, leaveTypeMap map[string]string) string {
	empName := "Unknown"
	if name, ok := empMap[req.EmployeeUID]; ok {
		empName = name
	}

	leaveTypeName := "Leave"
	if name, ok := leaveTypeMap[req.LeaveTypeUID]; ok {
		leaveTypeName = name
	}

	startDate := req.StartDate.Format("Jan 2")
	endDate := req.EndDate.Format("Jan 2, 2006")

	return empName + " - " + leaveTypeName + " (" + startDate + "-" + endDate + ")"
}

// Action type translations
var actionTranslations = map[string]struct {
	EN string
	AR string
}{
	"create":              {EN: "Created", AR: "تم الإنشاء"},
	"update":              {EN: "Updated", AR: "تم التحديث"},
	"delete":              {EN: "Deleted", AR: "تم الحذف"},
	"approve":             {EN: "Approved", AR: "تمت الموافقة"},
	"reject":              {EN: "Rejected", AR: "تم الرفض"},
	"submit":              {EN: "Submitted", AR: "تم التقديم"},
	"cancel":              {EN: "Cancelled", AR: "تم الإلغاء"},
	"assign_department":   {EN: "Assigned Department", AR: "تم تعيين القسم"},
	"remove_department":   {EN: "Removed Department", AR: "تم إزالة القسم"},
	"assign_role":         {EN: "Assigned Role", AR: "تم تعيين الدور"},
	"remove_role":         {EN: "Removed Role", AR: "تم إزالة الدور"},
	"update_permissions":  {EN: "Updated Permissions", AR: "تم تحديث الصلاحيات"},
	"adjust_balance":      {EN: "Adjusted Balance", AR: "تم تعديل الرصيد"},
	"deduct_balance":      {EN: "Deducted Balance", AR: "تم خصم الرصيد"},
	"assign_manager":      {EN: "Assigned Manager", AR: "تم تعيين المدير"},
	"remove_manager":      {EN: "Removed Manager", AR: "تم إزالة المدير"},
	"add_step":            {EN: "Added Step", AR: "تم إضافة خطوة"},
	"update_step":         {EN: "Updated Step", AR: "تم تحديث الخطوة"},
	"delete_step":         {EN: "Deleted Step", AR: "تم حذف الخطوة"},
	"activate":            {EN: "Activated", AR: "تم التفعيل"},
	"deactivate":          {EN: "Deactivated", AR: "تم التعطيل"},
	"auto_reject":         {EN: "Auto-Rejected", AR: "تم الرفض تلقائياً"},
	"annual_reset":        {EN: "Annual Reset", AR: "إعادة تعيين سنوية"},
	"system_adjustment":   {EN: "System Adjustment", AR: "تعديل النظام"},
	"cascade_delete":      {EN: "Cascade Delete", AR: "حذف متسلسل"},
	"create_transaction":  {EN: "Created Transaction", AR: "تم إنشاء معاملة"},
	"auto_create_transaction": {EN: "Auto-Created Transaction", AR: "تم إنشاء معاملة تلقائياً"},
}

// Entity type translations
var entityTypeTranslations = map[string]struct {
	EN string
	AR string
}{
	"employee":             {EN: "Employee", AR: "موظف"},
	"leave_request":        {EN: "Leave Request", AR: "طلب إجازة"},
	"leave_record":         {EN: "Leave Record", AR: "سجل إجازة"},
	"attendance_record":    {EN: "Attendance Record", AR: "سجل الحضور"},
	"approval_request":     {EN: "Approval Request", AR: "طلب موافقة"},
	"user":                 {EN: "User", AR: "مستخدم"},
	"department":           {EN: "Department", AR: "قسم"},
	"role":                 {EN: "Role", AR: "دور"},
	"approval_flow":        {EN: "Approval Flow", AR: "سير الموافقة"},
	"approval_flow_step":   {EN: "Approval Flow Step", AR: "خطوة سير الموافقة"},
	"leave_type":           {EN: "Leave Type", AR: "نوع الإجازة"},
	"shift":                {EN: "Shift", AR: "وردية"},
	"attendance_device":    {EN: "Attendance Device", AR: "جهاز الحضور"},
	"leave_balance":        {EN: "Leave Balance", AR: "رصيد الإجازة"},
	"holiday":              {EN: "Holiday", AR: "عطلة"},
}

// getActionLabels returns the English and Arabic labels for an action.
func getActionLabels(action string) (string, string) {
	if labels, ok := actionTranslations[action]; ok {
		return labels.EN, labels.AR
	}
	// Return the action itself if no translation found
	return action, action
}

// getEntityTypeLabels returns the English and Arabic labels for an entity type.
func getEntityTypeLabels(entityType string) (string, string) {
	if labels, ok := entityTypeTranslations[entityType]; ok {
		return labels.EN, labels.AR
	}
	// Return the entity type itself if no translation found
	return entityType, entityType
}

// enrichAuditLog builds an enriched DTO with resolved names and localized labels.
func (uc *ListEnrichedAuditLogsUseCase) enrichAuditLog(
	ctx context.Context,
	log *domain.AuditLog,
	actorMap map[string]*domain.Employee,
	entityDescMap map[string]string,
) EnrichedAuditLogDTO {
	enriched := EnrichedAuditLogDTO{
		UID:            log.UID,
		ActorUID:       log.ActorUID,
		Action:         log.Action,
		ActionLabel:    uc.i18n.T(ctx, "audit.action."+log.Action),
		EntityType:     log.EntityType,
		EntityTypeLabel: uc.i18n.T(ctx, "audit.entity."+log.EntityType),
		EntityUID:      log.EntityUID,
		OccurredAt:     log.OccurredAt.Format("2006-01-02T15:04:05Z07:00"),
		CreatedAt:      log.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		IsSystemAction: isSystemAction(log.Action),
	}

	// Resolve actor name
	if actor, found := actorMap[log.ActorUID]; found {
		enriched.ActorName = &actor.Name
		enriched.ActorType = "user"
	} else {
		enriched.ActorType = "system"
	}

	// Resolve entity description
	if desc, found := entityDescMap[log.EntityUID]; found {
		enriched.EntityDescription = &desc
	}

	// Translate metadata field names if metadata exists
	if log.Meta != nil && *log.Meta != "" {
		enriched.Meta = uc.translateMetadata(ctx, *log.Meta)
	}

	return enriched
}

// translateMetadata translates metadata field names recursively.
func (uc *ListEnrichedAuditLogsUseCase) translateMetadata(ctx context.Context, metaJSON string) map[string]interface{} {
	var meta map[string]interface{}
	if err := json.Unmarshal([]byte(metaJSON), &meta); err != nil {
		// If JSON parsing fails, return empty map
		return make(map[string]interface{})
	}

	return uc.translateMetadataRecursive(ctx, meta)
}

// translateMetadataRecursive recursively translates field names and well-known values in metadata.
func (uc *ListEnrichedAuditLogsUseCase) translateMetadataRecursive(ctx context.Context, meta map[string]interface{}) map[string]interface{} {
	translated := make(map[string]interface{})

	// action_key + action_params are consumed together to form a translated sentence.
	// They must not appear as individual keys in the output.
	skipKeys := map[string]bool{
		"details":       true,
		"action_key":    true,
		"action_params": true,
	}

	// If this record uses the new structured format, compose a translated action sentence.
	if actionKey, ok := meta["action_key"].(string); ok {
		var params map[string]interface{}
		if p, ok := meta["action_params"].(map[string]interface{}); ok {
			params = p
		}
		actionLabel := uc.i18n.T(ctx, "audit.metadata.action")
		translated[actionLabel] = uc.i18n.TWithParams(ctx, actionKey, params)
	} else if actionSentence, ok := meta["action"].(string); ok && actionSentence != "" {
		// Legacy: show the raw English sentence as-is (old DB records before migration).
		actionLabel := uc.i18n.T(ctx, "audit.metadata.action")
		translated[actionLabel] = actionSentence
		skipKeys["action"] = true
	}

	for key, value := range meta {
		if skipKeys[key] {
			continue
		}

		translatedKey := uc.i18n.T(ctx, "audit.metadata."+key)

		if nestedMap, ok := value.(map[string]interface{}); ok {
			translated[translatedKey] = uc.translateMetadataRecursive(ctx, nestedMap)
		} else {
			translated[translatedKey] = uc.translateMetadataValue(ctx, key, value)
		}
	}

	return translated
}

// translateMetadataValue translates a metadata value for well-known enum and boolean fields.
func (uc *ListEnrichedAuditLogsUseCase) translateMetadataValue(ctx context.Context, key string, value interface{}) interface{} {
	// Translate booleans
	if b, isBool := value.(bool); isBool {
		if b {
			return uc.i18n.T(ctx, "common.yes")
		}
		return uc.i18n.T(ctx, "common.no")
	}

	strVal, ok := value.(string)
	if !ok {
		return value
	}

	// Translate known enum values by key context
	switch key {
	case "old_status", "new_status", "status":
		candidate := uc.i18n.T(ctx, "audit.value.status."+strVal)
		if candidate != "audit.value.status."+strVal {
			return candidate
		}
	case "punch_type", "old_punch_type", "new_punch_type":
		candidate := uc.i18n.T(ctx, "audit.value.punch_type."+strVal)
		if candidate != "audit.value.punch_type."+strVal {
			return candidate
		}
	}

	// Format ISO date/datetime strings into readable form
	if isISODateString(strVal) {
		return formatAuditDateString(strVal)
	}

	return strVal
}

// isISODateString returns true if s looks like an ISO 8601 date or datetime.
func isISODateString(s string) bool {
	if len(s) < 10 {
		return false
	}
	return len(s) >= 4 && s[4] == '-' && len(s) >= 7 && s[7] == '-'
}

// formatAuditDateString formats an ISO date string to a readable display format.
func formatAuditDateString(s string) string {
	layouts := []string{
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05Z",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			if len(s) == 10 {
				return t.Format("Jan 2, 2006")
			}
			return t.Format("Jan 2, 2006 3:04 PM")
		}
	}
	return s
}

// isSystemAction determines if an action is system-generated.
func isSystemAction(action string) bool {
	systemActions := map[string]bool{
		"deduct_balance":          true,
		"annual_reset":            true,
		"auto_reject":             true,
		"system_adjustment":       true,
		"cascade_delete":          true,
		"auto_create_transaction": true,
	}
	return systemActions[action]
}
