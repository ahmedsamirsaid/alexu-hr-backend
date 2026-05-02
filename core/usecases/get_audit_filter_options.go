package usecases

import (
	"context"

	"github.com/banumusa/backend/core/ports"
)

// FilterOption represents a filter option with bilingual labels.
type FilterOption struct {
	Value   string `json:"value"`   // Technical value (e.g., "approve")
	LabelEN string `json:"labelEn"` // English label
	LabelAR string `json:"labelAr"` // Arabic label
}

// GetAuditFilterOptionsInput defines the input for getting filter options.
type GetAuditFilterOptionsInput struct {
	Language string // "en" or "ar" (optional, for future use)
}

// GetAuditFilterOptionsOutput contains the available filter options.
type GetAuditFilterOptionsOutput struct {
	ActionTypes []FilterOption `json:"actionTypes"`
	EntityTypes []FilterOption `json:"entityTypes"`
}

// GetAuditFilterOptionsUseCase handles retrieving available filter options for audit logs.
type GetAuditFilterOptionsUseCase struct {
	auditRepo ports.AuditLogRepository
	db        ports.DB
}

// NewGetAuditFilterOptionsUseCase creates a new filter options use case.
func NewGetAuditFilterOptionsUseCase(
	auditRepo ports.AuditLogRepository,
	db ports.DB,
) *GetAuditFilterOptionsUseCase {
	return &GetAuditFilterOptionsUseCase{
		auditRepo: auditRepo,
		db:        db,
	}
}

// Execute retrieves available filter options from the audit log.
func (uc *GetAuditFilterOptionsUseCase) Execute(ctx context.Context, input GetAuditFilterOptionsInput) (*GetAuditFilterOptionsOutput, error) {
	// Get distinct action types
	actionTypes, err := uc.auditRepo.GetDistinctActionTypes(ctx, uc.db)
	if err != nil {
		return nil, err
	}

	// Get distinct entity types
	entityTypes, err := uc.auditRepo.GetDistinctEntityTypes(ctx, uc.db)
	if err != nil {
		return nil, err
	}

	// Map action types to filter options with labels
	actionOptions := make([]FilterOption, 0, len(actionTypes))
	for _, action := range actionTypes {
		labelEN, labelAR := getActionLabelsForFilter(action)
		actionOptions = append(actionOptions, FilterOption{
			Value:   action,
			LabelEN: labelEN,
			LabelAR: labelAR,
		})
	}

	// Map entity types to filter options with labels
	entityOptions := make([]FilterOption, 0, len(entityTypes))
	for _, entityType := range entityTypes {
		labelEN, labelAR := getEntityTypeLabelsForFilter(entityType)
		entityOptions = append(entityOptions, FilterOption{
			Value:   entityType,
			LabelEN: labelEN,
			LabelAR: labelAR,
		})
	}

	return &GetAuditFilterOptionsOutput{
		ActionTypes: actionOptions,
		EntityTypes: entityOptions,
	}, nil
}

// Action type translations (same as in list_enriched_audit_logs.go)
var actionTranslationsForFilter = map[string]struct {
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

// Entity type translations (same as in list_enriched_audit_logs.go)
var entityTypeTranslationsForFilter = map[string]struct {
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

// getActionLabelsForFilter returns the English and Arabic labels for an action.
func getActionLabelsForFilter(action string) (string, string) {
	if labels, ok := actionTranslationsForFilter[action]; ok {
		return labels.EN, labels.AR
	}
	// Return the action itself if no translation found
	return action, action
}

// getEntityTypeLabelsForFilter returns the English and Arabic labels for an entity type.
func getEntityTypeLabelsForFilter(entityType string) (string, string) {
	if labels, ok := entityTypeTranslationsForFilter[entityType]; ok {
		return labels.EN, labels.AR
	}
	// Return the entity type itself if no translation found
	return entityType, entityType
}
