package usecases

import (
	"context"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

// EnrichedAuditLogDTO represents an audit log entry with resolved names and bilingual labels.
type EnrichedAuditLogDTO struct {
	UID               string  `json:"uid"`
	ActorUID          string  `json:"actorUid"`
	ActorName         *string `json:"actorName"`         // Resolved actor name
	ActorType         string  `json:"actorType"`         // "user" or "system"
	Action            string  `json:"action"`
	ActionLabelEN     string  `json:"actionLabelEn"`     // "Approved"
	ActionLabelAR     string  `json:"actionLabelAr"`     // "تمت الموافقة"
	EntityType        string  `json:"entityType"`
	EntityTypeLabelEN string  `json:"entityTypeLabelEn"` // "Leave Request"
	EntityTypeLabelAR string  `json:"entityTypeLabelAr"` // "طلب إجازة"
	EntityUID         string  `json:"entityUid"`
	EntityDescription *string `json:"entityDescription"` // "Ahmed Al-Sayed - Annual Leave (Jan 15-20)"
	Meta              *string `json:"meta"`
	OccurredAt        string  `json:"occurredAt"`
	CreatedAt         string  `json:"createdAt"`
	IsSystemAction    bool    `json:"isSystemAction"` // True for auto-generated actions
}

// ListEnrichedAuditLogsInput defines the input for listing enriched audit logs.
type ListEnrichedAuditLogsInput struct {
	// Existing filters
	EntityType string
	ActorUID   string
	SearchText string

	// New filters
	ActorName         string     // Partial name match
	ActionTypes       []string   // Filter by action types
	EntityTypes       []string   // Filter by entity types
	StartDate         *time.Time // Filter by date range start
	EndDate           *time.Time // Filter by date range end
	HideSystemActions bool       // Default: true

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
) *ListEnrichedAuditLogsUseCase {
	return &ListEnrichedAuditLogsUseCase{
		auditRepo:     auditRepo,
		employeeRepo:  employeeRepo,
		deptRepo:      deptRepo,
		leaveReqRepo:  leaveReqRepo,
		leaveTypeRepo: leaveTypeRepo,
		attRecordRepo: attRecordRepo,
		db:            db,
	}
}

// Execute retrieves enriched audit logs with resolved names and bilingual labels.
func (uc *ListEnrichedAuditLogsUseCase) Execute(ctx context.Context, input ListEnrichedAuditLogsInput) (*ListEnrichedAuditLogsOutput, error) {
	// 1. Build filters for repository
	filters := ports.AuditLogFilters{
		EntityType:        input.EntityType,
		EntityTypes:       input.EntityTypes,
		ActorUID:          input.ActorUID,
		ActorName:         input.ActorName,
		ActionTypes:       input.ActionTypes,
		SearchText:        input.SearchText,
		StartDate:         input.StartDate,
		EndDate:           input.EndDate,
		HideSystemActions: input.HideSystemActions,
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
		enriched := uc.enrichAuditLog(log, actorMap, entityDescMap)
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

// enrichAuditLog builds an enriched DTO with resolved names and labels.
func (uc *ListEnrichedAuditLogsUseCase) enrichAuditLog(
	log *domain.AuditLog,
	actorMap map[string]*domain.Employee,
	entityDescMap map[string]string,
) EnrichedAuditLogDTO {
	enriched := EnrichedAuditLogDTO{
		UID:            log.UID,
		ActorUID:       log.ActorUID,
		Action:         log.Action,
		EntityType:     log.EntityType,
		EntityUID:      log.EntityUID,
		Meta:           log.Meta,
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

	// Add action labels with translations
	actionLabelEN, actionLabelAR := getActionLabels(log.Action)
	enriched.ActionLabelEN = actionLabelEN
	enriched.ActionLabelAR = actionLabelAR

	// Add entity type labels with translations
	entityTypeLabelEN, entityTypeLabelAR := getEntityTypeLabels(log.EntityType)
	enriched.EntityTypeLabelEN = entityTypeLabelEN
	enriched.EntityTypeLabelAR = entityTypeLabelAR

	return enriched
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
