package usecases

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/banumusa/backend/core/audit"
	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type UpdateEmployeeProfileInput struct {
	EmployeeUID      string
	ActorUserID      int64
	ActorEmployeeUID string
	Changes          map[string]any
}

type UpdateEmployeeProfileOutput struct {
	Employee *domain.Employee
}

type UpdateEmployeeProfileUseCase struct {
	db           ports.DB
	employeeRepo ports.EmployeeRepository
	userRepo     ports.UserRepository
	roleRepo     ports.RoleRepository
	auditor      audit.Auditor
}

func NewUpdateEmployeeProfileUseCase(
	db ports.DB,
	employeeRepo ports.EmployeeRepository,
	userRepo ports.UserRepository,
	roleRepo ports.RoleRepository,
	auditor audit.Auditor,
) *UpdateEmployeeProfileUseCase {
	return &UpdateEmployeeProfileUseCase{
		db:           db,
		employeeRepo: employeeRepo,
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		auditor:      auditor,
	}
}

func (uc *UpdateEmployeeProfileUseCase) Execute(ctx context.Context, input UpdateEmployeeProfileInput) (*UpdateEmployeeProfileOutput, error) {
	if len(input.Changes) == 0 {
		return nil, ErrNoProfileChangesRequested
	}

	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	authorized, err := uc.roleRepo.IsUserAuthorizedApprover(ctx, tx, input.ActorUserID, informationCenterRoleUID, "")
	if err != nil {
		return nil, err
	}
	if !authorized {
		return nil, ErrNotAuthorizedApprover
	}

	employee, err := uc.employeeRepo.GetByUID(ctx, tx, input.EmployeeUID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, ErrEmployeeNotFound
	}

	oldValues := make(map[string]any)
	newValues := make(map[string]any)

	for field, value := range input.Changes {
		changed, err := applyEmployeeProfileChangeField(employee, field, value, oldValues, newValues)
		if err != nil {
			return nil, err
		}
		if !changed {
			delete(oldValues, field)
			delete(newValues, field)
		}
	}

	if len(newValues) == 0 {
		return nil, ErrNoProfileChangesRequested
	}

	if err := validateUpdatedEmployeeUniqueness(ctx, tx, uc.employeeRepo, uc.userRepo, employee, oldValues); err != nil {
		return nil, err
	}
	employee.ApplyClassificationDefaults()
	if !domain.IsValidEmployeeSubTypeForType(employee.Type, employee.SubType) {
		return nil, fmt.Errorf("%w: subtype %q for type %q", ErrInvalidEmployeeClassification, employee.SubType, employee.Type)
	}

	if err := uc.employeeRepo.Update(ctx, tx, employee); err != nil {
		return nil, err
	}

	if _, ok := newValues["mobile"]; ok {
		user, err := uc.userRepo.GetByEmployeeUID(ctx, tx, employee.UID)
		if err != nil {
			return nil, err
		}
		if user != nil {
			user.Phone = employee.Mobile
			if err := uc.userRepo.Update(ctx, tx, user); err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	defer uc.auditor.Actor(input.ActorEmployeeUID).
		Did(audit.ActionUpdate).
		On(audit.EntityEmployee, employee.UID).
		WithMeta("changed_fields", keysOfMap(newValues)).
		WithMeta("old_values", oldValues).
		WithMeta("new_values", newValues).
		Save(ctx)

	return &UpdateEmployeeProfileOutput{Employee: employee}, nil
}

func applyEmployeeProfileChangeField(employee *domain.Employee, field string, value any, oldValues, newValues map[string]any) (bool, error) {
	switch field {
	case "name":
		v, ok := value.(string)
		if !ok {
			return false, fmt.Errorf("invalid %s", field)
		}
		v = strings.TrimSpace(v)
		if v == "" {
			return false, fmt.Errorf("%s is required", field)
		}
		if employee.Name == v {
			return false, nil
		}
		oldValues[field] = employee.Name
		employee.Name = v
		newValues[field] = employee.Name
		return true, nil
	case "mobile":
		v, ok := value.(string)
		if !ok {
			return false, fmt.Errorf("invalid %s", field)
		}
		v = strings.TrimSpace(v)
		if v == "" {
			return false, fmt.Errorf("%s is required", field)
		}
		if employee.Mobile == v {
			return false, nil
		}
		oldValues[field] = employee.Mobile
		employee.Mobile = v
		newValues[field] = employee.Mobile
		return true, nil
	case "governmentId":
		v, ok := value.(string)
		if !ok {
			return false, fmt.Errorf("invalid %s", field)
		}
		v = strings.TrimSpace(v)
		if v == "" {
			return false, fmt.Errorf("%s is required", field)
		}
		if employee.GovernmentID == v {
			return false, nil
		}
		oldValues[field] = employee.GovernmentID
		employee.GovernmentID = v
		newValues[field] = employee.GovernmentID
		return true, nil
	case "universityId":
		v, ok := value.(string)
		if !ok {
			return false, fmt.Errorf("invalid %s", field)
		}
		v = strings.TrimSpace(v)
		if v == "" {
			return false, fmt.Errorf("%s is required", field)
		}
		if employee.UniversityID == v {
			return false, nil
		}
		oldValues[field] = employee.UniversityID
		employee.UniversityID = v
		newValues[field] = employee.UniversityID
		return true, nil
	case "email":
		return setStringPtrField(field, &employee.Email, value, oldValues, newValues)
	case "telephoneNumber":
		return setStringPtrField(field, &employee.TelephoneNumber, value, oldValues, newValues)
	case "gender":
		return setStringPtrField(field, &employee.Gender, value, oldValues, newValues)
	case "religion":
		return setStringPtrField(field, &employee.Religion, value, oldValues, newValues)
	case "maritalStatus":
		return setStringPtrField(field, &employee.MaritalStatus, value, oldValues, newValues)
	case "address":
		return setStringPtrField(field, &employee.Address, value, oldValues, newValues)
	case "placeOfBirth":
		return setStringPtrField(field, &employee.PlaceOfBirth, value, oldValues, newValues)
	case "placeOfResidence":
		return setStringPtrField(field, &employee.PlaceOfResidence, value, oldValues, newValues)
	case "policeStation":
		return setStringPtrField(field, &employee.PoliceStation, value, oldValues, newValues)
	case "academicLevel":
		return setStringPtrField(field, &employee.AcademicLevel, value, oldValues, newValues)
	case "educationalQualification":
		return setStringPtrField(field, &employee.EducationalQualification, value, oldValues, newValues)
	case "universityName":
		return setStringPtrField(field, &employee.UniversityName, value, oldValues, newValues)
	case "faculty":
		return setStringPtrField(field, &employee.Faculty, value, oldValues, newValues)
	case "specialization":
		return setStringPtrField(field, &employee.Specialization, value, oldValues, newValues)
	case "appointmentDecisionNumber":
		return setStringPtrField(field, &employee.AppointmentDecisionNumber, value, oldValues, newValues)
	case "appointmentType":
		return setStringPtrField(field, &employee.AppointmentType, value, oldValues, newValues)
	case "departmentName":
		return setStringPtrField(field, &employee.DepartmentName, value, oldValues, newValues)
	case "grade":
		return setStringPtrField(field, &employee.Grade, value, oldValues, newValues)
	case "workEntity":
		return setStringPtrField(field, &employee.WorkEntity, value, oldValues, newValues)
	case "employeeFileNumber":
		return setStringPtrField(field, &employee.EmployeeFileNumber, value, oldValues, newValues)
	case "insuranceNumber":
		return setStringPtrField(field, &employee.InsuranceNumber, value, oldValues, newValues)
	case "employmentStatus":
		return setStringPtrField(field, &employee.EmploymentStatus, value, oldValues, newValues)
	case "militaryStatus":
		return setStringPtrField(field, &employee.MilitaryStatus, value, oldValues, newValues)
	case "medicalCadre":
		return setStringPtrField(field, &employee.MedicalCadre, value, oldValues, newValues)
	case "personalPhotoUrl":
		return setStringPtrField(field, &employee.PersonalPhotoURL, value, oldValues, newValues)
	case "nationalCardImageUrl":
		return setStringPtrField(field, &employee.NationalCardImageURL, value, oldValues, newValues)
	case "qualificationCertificateImageUrl":
		return setStringPtrField(field, &employee.QualificationCertificateImageURL, value, oldValues, newValues)
	case "cvUrl":
		return setStringPtrField(field, &employee.CVURL, value, oldValues, newValues)
	case "memberNumber":
		return setStringPtrField(field, &employee.MemberNumber, value, oldValues, newValues)
	case "insuranceCode":
		return setStringPtrField(field, &employee.InsuranceCode, value, oldValues, newValues)
	case "jobGroup":
		return setStringPtrField(field, &employee.JobGroup, value, oldValues, newValues)
	case "qualitativeGroup":
		return setStringPtrField(field, &employee.QualitativeGroup, value, oldValues, newValues)
	case "jobTitleAtLevel":
		return setStringPtrField(field, &employee.JobTitleAtLevel, value, oldValues, newValues)
	case "jobTitleBeforePlacement":
		return setStringPtrField(field, &employee.JobTitleBeforePlacement, value, oldValues, newValues)
	case "financialGrade":
		return setStringPtrField(field, &employee.FinancialGrade, value, oldValues, newValues)
	case "previousFinancialGrade":
		return setStringPtrField(field, &employee.PreviousFinancialGrade, value, oldValues, newValues)
	case "jobLevel":
		return setStringPtrField(field, &employee.JobLevel, value, oldValues, newValues)
	case "natureOfAppointment":
		return setStringPtrField(field, &employee.NatureOfAppointment, value, oldValues, newValues)
	case "notes":
		return setStringPtrField(field, &employee.Notes, value, oldValues, newValues)
	case "decisionFileUrl":
		return setStringPtrField(field, &employee.DecisionFileURL, value, oldValues, newValues)
	case "departmentUid":
		return setStringPtrField(field, &employee.DepartmentUID, value, oldValues, newValues)
	case "shiftUid":
		return setStringPtrField(field, &employee.ShiftUID, value, oldValues, newValues)
	case "hireDate":
		return setTimeField(field, &employee.HireDate, value, oldValues, newValues)
	case "dateOfBirth":
		return setTimePtrField(field, &employee.DateOfBirth, value, oldValues, newValues)
	case "idCardValidUntil":
		return setTimePtrField(field, &employee.IDCardValidUntil, value, oldValues, newValues)
	case "subscriptionDate":
		return setTimePtrField(field, &employee.SubscriptionDate, value, oldValues, newValues)
	case "actualAppointmentReappointmentDate":
		return setTimePtrField(field, &employee.ActualAppointmentReappointmentDate, value, oldValues, newValues)
	case "appointmentDecisionDate":
		return setTimePtrField(field, &employee.AppointmentDecisionDate, value, oldValues, newValues)
	case "decisionDate":
		return setTimePtrField(field, &employee.DecisionDate, value, oldValues, newValues)
	case "gradeGrantDate":
		return setTimePtrField(field, &employee.GradeGrantDate, value, oldValues, newValues)
	case "status":
		v, ok := value.(string)
		if !ok {
			return false, fmt.Errorf("invalid %s", field)
		}
		next := domain.EmployeeStatus(strings.TrimSpace(strings.ToLower(v)))
		if next == "" {
			return false, fmt.Errorf("%s is required", field)
		}
		if employee.Status == next {
			return false, nil
		}
		oldValues[field] = string(employee.Status)
		employee.Status = next
		newValues[field] = string(employee.Status)
		return true, nil
	case "type":
		v, ok := value.(string)
		if !ok {
			return false, fmt.Errorf("invalid %s", field)
		}
		next := domain.NormalizeEmployeeType(v)
		if next == "" {
			return false, ErrInvalidEmployeeClassification
		}
		if employee.Type == next {
			return false, nil
		}
		oldValues[field] = string(employee.Type)
		employee.Type = next
		newValues[field] = string(employee.Type)
		return true, nil
	case "subType":
		v, ok := value.(string)
		if !ok {
			return false, fmt.Errorf("invalid %s", field)
		}
		next := domain.NormalizeEmployeeSubType(v)
		if next == "" {
			return false, ErrInvalidEmployeeClassification
		}
		if employee.SubType == next {
			return false, nil
		}
		oldValues[field] = string(employee.SubType)
		employee.SubType = next
		newValues[field] = string(employee.SubType)
		return true, nil
	case "personWithSpecialNeeds":
		return setBoolField(field, &employee.PersonWithSpecialNeeds, value, oldValues, newValues)
	case "solidarityFund":
		return setBoolField(field, &employee.SolidarityFund, value, oldValues, newValues)
	case "reappointment":
		return setBoolField(field, &employee.Reappointment, value, oldValues, newValues)
	case "appointmentSeniorityOrGradeWithdrawal":
		return setBoolField(field, &employee.AppointmentSeniorityOrGradeWithdrawal, value, oldValues, newValues)
	case "yearObtained":
		return setIntPtrField(field, &employee.YearObtained, value, oldValues, newValues)
	case "decisionNumber":
		return setInt64PtrField(field, &employee.DecisionNumber, value, oldValues, newValues)
	default:
		return false, fmt.Errorf("unsupported field %q", field)
	}
}

func validateUpdatedEmployeeUniqueness(ctx context.Context, q ports.Querier, employeeRepo ports.EmployeeRepository, userRepo ports.UserRepository, employee *domain.Employee, oldValues map[string]any) error {
	if _, ok := oldValues["mobile"]; ok {
		existing, err := employeeRepo.ExistingMobiles(ctx, q, []string{employee.Mobile})
		if err != nil {
			return err
		}
		if len(existing) > 0 && existing[0] == employee.Mobile {
			current, err := employeeRepo.GetByUID(ctx, q, employee.UID)
			if err != nil {
				return err
			}
			if current == nil || current.Mobile != employee.Mobile {
				return ErrEmployeeMobileAlreadyExists
			}
		}
		existingUser, err := userRepo.GetByPhone(ctx, q, employee.Mobile)
		if err != nil {
			return err
		}
		if existingUser != nil && (existingUser.EmployeeUID == nil || *existingUser.EmployeeUID != employee.UID) {
			return ErrPhoneAlreadyExists
		}
	}
	if _, ok := oldValues["governmentId"]; ok {
		existing, err := employeeRepo.ExistingGovernmentIDs(ctx, q, []string{employee.GovernmentID})
		if err != nil {
			return err
		}
		if len(existing) > 0 {
			current, err := employeeRepo.GetByUID(ctx, q, employee.UID)
			if err != nil {
				return err
			}
			if current == nil || current.GovernmentID != employee.GovernmentID {
				return ErrGovernmentIDAlreadyExists
			}
		}
	}
	if _, ok := oldValues["universityId"]; ok {
		existing, err := employeeRepo.ExistingUniversityIDs(ctx, q, []string{employee.UniversityID})
		if err != nil {
			return err
		}
		if len(existing) > 0 {
			current, err := employeeRepo.GetByUID(ctx, q, employee.UID)
			if err != nil {
				return err
			}
			if current == nil || current.UniversityID != employee.UniversityID {
				return ErrUniversityIDAlreadyExists
			}
		}
	}
	return nil
}

func setStringPtrField(field string, target **string, value any, oldValues, newValues map[string]any) (bool, error) {
	var next *string
	switch v := value.(type) {
	case string:
		next = trimStringPtr(&v)
	case *string:
		next = trimStringPtr(v)
	case nil:
		next = nil
	default:
		return false, fmt.Errorf("invalid %s", field)
	}
	if equalStringPointers(*target, next) {
		return false, nil
	}
	oldValues[field] = stringPtrValue(*target)
	*target = next
	newValues[field] = stringPtrValue(*target)
	return true, nil
}

func setTimePtrField(field string, target **time.Time, value any, oldValues, newValues map[string]any) (bool, error) {
	var next *time.Time
	switch v := value.(type) {
	case time.Time:
		vv := v
		next = &vv
	case *time.Time:
		next = cloneTimePtr(v)
	case nil:
		next = nil
	default:
		return false, fmt.Errorf("invalid %s", field)
	}
	if equalDatePointers(*target, next) {
		return false, nil
	}
	oldValues[field] = formatOptionalDate(*target)
	*target = next
	newValues[field] = formatOptionalDate(*target)
	return true, nil
}

func setTimeField(field string, target *time.Time, value any, oldValues, newValues map[string]any) (bool, error) {
	var next time.Time
	switch v := value.(type) {
	case time.Time:
		next = v
	default:
		return false, fmt.Errorf("invalid %s", field)
	}
	if target.Equal(next) {
		return false, nil
	}
	oldValues[field] = target.Format("2006-01-02")
	*target = next
	newValues[field] = target.Format("2006-01-02")
	return true, nil
}

func setBoolField(field string, target *bool, value any, oldValues, newValues map[string]any) (bool, error) {
	v, ok := value.(bool)
	if !ok {
		return false, fmt.Errorf("invalid %s", field)
	}
	if *target == v {
		return false, nil
	}
	oldValues[field] = *target
	*target = v
	newValues[field] = *target
	return true, nil
}

func setIntPtrField(field string, target **int, value any, oldValues, newValues map[string]any) (bool, error) {
	var next *int
	switch v := value.(type) {
	case int:
		vv := v
		next = &vv
	case *int:
		next = v
	case nil:
		next = nil
	default:
		return false, fmt.Errorf("invalid %s", field)
	}
	if equalIntPointers(*target, next) {
		return false, nil
	}
	oldValues[field] = intPtrValue(*target)
	*target = next
	newValues[field] = intPtrValue(*target)
	return true, nil
}

func setInt64PtrField(field string, target **int64, value any, oldValues, newValues map[string]any) (bool, error) {
	var next *int64
	switch v := value.(type) {
	case int64:
		vv := v
		next = &vv
	case *int64:
		next = v
	case nil:
		next = nil
	default:
		return false, fmt.Errorf("invalid %s", field)
	}
	if equalInt64Pointers(*target, next) {
		return false, nil
	}
	oldValues[field] = int64PtrValue(*target)
	*target = next
	newValues[field] = int64PtrValue(*target)
	return true, nil
}

func cloneStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	v := *value
	return &v
}

func cloneTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	v := *value
	return &v
}

func formatOptionalDate(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format("2006-01-02")
}

func equalStringPointers(left, right *string) bool {
	return stringPtrValue(left) == stringPtrValue(right)
}

func equalDatePointers(left, right *time.Time) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return left.Equal(*right)
}

func equalIntPointers(left, right *int) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func equalInt64Pointers(left, right *int64) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func intPtrValue(value *int) any {
	if value == nil {
		return nil
	}
	return *value
}

func int64PtrValue(value *int64) any {
	if value == nil {
		return nil
	}
	return *value
}

func keysOfMap(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return keys
}
