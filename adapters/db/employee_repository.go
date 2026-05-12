package db

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type EmployeeRepository struct{}

const employeeSelectColumns = `
	id, uid, name, mobile, government_id, university_id, email,
	hire_date, status, type, sub_type,
	telephone_number, person_with_special_needs, date_of_birth, gender, religion, marital_status,
	address, place_of_birth, place_of_residence, police_station, id_card_valid_until,
	academic_level, educational_qualification, university_name, faculty, specialization,
	year_obtained,
	actual_appointment_reappointment_date, appointment_decision_date, appointment_decision_number, appointment_type, department_name,
	grade, work_entity, employee_file_number, insurance_number, employment_status,
	solidarity_fund, subscription_date, military_status, medical_cadre,
	personal_photo_url, national_card_image_url, qualification_certificate_image_url, cv_url,
	member_number, insurance_code, job_group, qualitative_group, job_title_at_level,
	job_title_before_placement, financial_grade, previous_financial_grade, job_level,
	decision_date, decision_number, grade_grant_date, nature_of_appointment, notes, reappointment, decision_file_url,
	appointment_seniority_or_grade_withdrawal,
	department_uid, shift_uid, manager_uid, created_at, updated_at`

func NewEmployeeRepository() *EmployeeRepository {
	return &EmployeeRepository{}
}

func (r *EmployeeRepository) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.Employee, error) {
	query := `SELECT ` + employeeSelectColumns + ` FROM employees WHERE id = ?`

	return r.scanEmployee(q.QueryRowContext(ctx, query, id))
}

func (r *EmployeeRepository) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.Employee, error) {
	query := `SELECT ` + employeeSelectColumns + ` FROM employees WHERE uid = ?`

	return r.scanEmployee(q.QueryRowContext(ctx, query, uid))
}

func (r *EmployeeRepository) GetByUIDs(ctx context.Context, q ports.Querier, uids []string) ([]*domain.Employee, error) {
	if len(uids) == 0 {
		return []*domain.Employee{}, nil
	}

	// Build IN clause with placeholders
	query := `SELECT ` + employeeSelectColumns + ` FROM employees WHERE uid IN (`

	args := make([]any, len(uids))
	for i, uid := range uids {
		if i > 0 {
			query += ", "
		}
		query += "?"
		args[i] = uid
	}
	query += ")"

	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		slog.Error("employee_repository.GetByUIDs.query", "error", err, "count", len(uids))
		return nil, err
	}
	defer rows.Close()

	var employees []*domain.Employee
	for rows.Next() {
		emp, err := r.scanEmployeeRow(rows)
		if err != nil {
			slog.Error("employee_repository.GetByUIDs.scan", "error", err)
			return nil, err
		}
		employees = append(employees, emp)
	}

	if err := rows.Err(); err != nil {
		slog.Error("employee_repository.GetByUIDs.rows_err", "error", err)
		return nil, err
	}

	return employees, nil
}
func (r *EmployeeRepository) Create(ctx context.Context, q ports.Querier, employee *domain.Employee) error {
	query := `
		INSERT INTO employees (
			uid, name, mobile, government_id, university_id, email,
			hire_date, status, type, sub_type,
			telephone_number, person_with_special_needs, date_of_birth, gender, religion, marital_status,
			address, place_of_birth, place_of_residence, police_station, id_card_valid_until,
			academic_level, educational_qualification, university_name, faculty, specialization,
			year_obtained,
			actual_appointment_reappointment_date, appointment_decision_date, appointment_decision_number, appointment_type, department_name,
			grade, work_entity, employee_file_number, insurance_number, employment_status,
			solidarity_fund, subscription_date, military_status, medical_cadre,
			personal_photo_url, national_card_image_url, qualification_certificate_image_url, cv_url,
			member_number, insurance_code, job_group, qualitative_group, job_title_at_level,
			job_title_before_placement, financial_grade, previous_financial_grade, job_level,
			decision_date, decision_number, grade_grant_date, nature_of_appointment, notes, reappointment, decision_file_url,
			appointment_seniority_or_grade_withdrawal,
			department_uid, shift_uid, manager_uid, created_at, updated_at
		)
		VALUES (
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, 
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?, ?, ?, ?,
			?,
			?, ?, ?, ?
		)`

	now := time.Now()
	employee.ApplyClassificationDefaults()
	employee.CreatedAt = now
	employee.UpdatedAt = now

	result, err := q.ExecContext(ctx, query,
		employee.UID, employee.Name, employee.Mobile,
		employee.GovernmentID, employee.UniversityID, employee.Email,
		employee.HireDate, employee.Status, employee.Type, employee.SubType,
		employee.TelephoneNumber, employee.PersonWithSpecialNeeds, nullableDate(employee.DateOfBirth), employee.Gender, employee.Religion, employee.MaritalStatus,
		employee.Address, employee.PlaceOfBirth, employee.PlaceOfResidence, employee.PoliceStation, nullableDate(employee.IDCardValidUntil),
		employee.AcademicLevel, employee.EducationalQualification, employee.UniversityName, employee.Faculty, employee.Specialization,
		employee.YearObtained,
		nullableDate(employee.ActualAppointmentReappointmentDate), nullableDate(employee.AppointmentDecisionDate), employee.AppointmentDecisionNumber, employee.AppointmentType, employee.DepartmentName,
		employee.Grade, employee.WorkEntity, employee.EmployeeFileNumber, employee.InsuranceNumber, employee.EmploymentStatus,
		employee.SolidarityFund, nullableDate(employee.SubscriptionDate), employee.MilitaryStatus, employee.MedicalCadre,
		employee.PersonalPhotoURL, employee.NationalCardImageURL, employee.QualificationCertificateImageURL, employee.CVURL,
		employee.MemberNumber, employee.InsuranceCode, employee.JobGroup, employee.QualitativeGroup, employee.JobTitleAtLevel,
		employee.JobTitleBeforePlacement, employee.FinancialGrade, employee.PreviousFinancialGrade, employee.JobLevel,
		nullableDate(employee.DecisionDate), employee.DecisionNumber, nullableDate(employee.GradeGrantDate), employee.NatureOfAppointment, employee.Notes, employee.Reappointment, employee.DecisionFileURL,
		employee.AppointmentSeniorityOrGradeWithdrawal,
		employee.DepartmentUID, employee.ShiftUID, employee.ManagerUID, employee.CreatedAt, employee.UpdatedAt)
	if err != nil {
		slog.Error("employee_repository.Create.exec_query", "error", err, "uid", employee.UID)
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		slog.Error("employee_repository.Create.last_insert_id", "error", err, "uid", employee.UID)
		return err
	}
	employee.ID = id

	return nil
}

func (r *EmployeeRepository) Update(ctx context.Context, q ports.Querier, employee *domain.Employee) error {
	query := `
		UPDATE employees
		SET name = ?, mobile = ?, government_id = ?, university_id = ?, email = ?,
		    hire_date = ?, status = ?, type = ?, sub_type = ?,
		    telephone_number = ?, person_with_special_needs = ?, date_of_birth = ?, gender = ?, religion = ?, marital_status = ?,
		    address = ?, place_of_birth = ?, place_of_residence = ?, police_station = ?, id_card_valid_until = ?,
		    academic_level = ?, educational_qualification = ?, university_name = ?, faculty = ?, specialization = ?, year_obtained = ?,
		    actual_appointment_reappointment_date = ?, appointment_decision_date = ?, appointment_decision_number = ?, appointment_type = ?, department_name = ?,
		    grade = ?, work_entity = ?, employee_file_number = ?, insurance_number = ?, employment_status = ?,
		    solidarity_fund = ?, subscription_date = ?, military_status = ?, medical_cadre = ?,
		    personal_photo_url = ?, national_card_image_url = ?, qualification_certificate_image_url = ?, cv_url = ?,
		    member_number = ?, insurance_code = ?, job_group = ?, qualitative_group = ?, job_title_at_level = ?,
		    job_title_before_placement = ?, financial_grade = ?, previous_financial_grade = ?, job_level = ?,
		    decision_date = ?, decision_number = ?, grade_grant_date = ?, nature_of_appointment = ?, notes = ?, reappointment = ?, decision_file_url = ?,
		    appointment_seniority_or_grade_withdrawal = ?, department_uid = ?, shift_uid = ?, manager_uid = ?, updated_at = ?
		WHERE id = ?`

	employee.ApplyClassificationDefaults()
	employee.UpdatedAt = time.Now()

	_, err := q.ExecContext(ctx, query,
		employee.Name, employee.Mobile, employee.GovernmentID, employee.UniversityID,
		employee.Email, employee.HireDate, employee.Status, employee.Type, employee.SubType,
		employee.TelephoneNumber, employee.PersonWithSpecialNeeds, nullableDate(employee.DateOfBirth), employee.Gender, employee.Religion, employee.MaritalStatus,
		employee.Address, employee.PlaceOfBirth, employee.PlaceOfResidence, employee.PoliceStation, nullableDate(employee.IDCardValidUntil),
		employee.AcademicLevel, employee.EducationalQualification, employee.UniversityName, employee.Faculty, employee.Specialization,
		employee.YearObtained,
		nullableDate(employee.ActualAppointmentReappointmentDate), nullableDate(employee.AppointmentDecisionDate), employee.AppointmentDecisionNumber, employee.AppointmentType, employee.DepartmentName,
		employee.Grade, employee.WorkEntity, employee.EmployeeFileNumber, employee.InsuranceNumber, employee.EmploymentStatus,
		employee.SolidarityFund, nullableDate(employee.SubscriptionDate), employee.MilitaryStatus, employee.MedicalCadre,
		employee.PersonalPhotoURL, employee.NationalCardImageURL, employee.QualificationCertificateImageURL, employee.CVURL,
		employee.MemberNumber, employee.InsuranceCode, employee.JobGroup, employee.QualitativeGroup, employee.JobTitleAtLevel,
		employee.JobTitleBeforePlacement, employee.FinancialGrade, employee.PreviousFinancialGrade, employee.JobLevel,
		nullableDate(employee.DecisionDate), employee.DecisionNumber, nullableDate(employee.GradeGrantDate), employee.NatureOfAppointment, employee.Notes, employee.Reappointment, employee.DecisionFileURL,
		employee.AppointmentSeniorityOrGradeWithdrawal, employee.DepartmentUID, employee.ShiftUID, employee.ManagerUID, employee.UpdatedAt, employee.ID)
	if err != nil {
		slog.Error("employee_repository.Update.exec_query", "error", err, "uid", employee.UID)
	}

	return err
}

func (r *EmployeeRepository) List(ctx context.Context, q ports.Querier, filter *ports.EmployeeListFilter) ([]*domain.Employee, error) {
	query := `SELECT ` + employeeSelectColumns + ` FROM employees WHERE 1=1`

	var args []any

	if filter != nil {
		if filter.Status != nil {
			query += ` AND status = ?`
			args = append(args, *filter.Status)
		}
		if filter.HireDateFrom != nil {
			query += ` AND hire_date >= ?`
			args = append(args, *filter.HireDateFrom)
		}
		if filter.HireDateTo != nil {
			query += ` AND hire_date <= ?`
			args = append(args, *filter.HireDateTo)
		}
	}

	query += ` ORDER BY name ASC`

	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		slog.Error("employee_repository.List.query", "error", err)
		return nil, err
	}
	defer rows.Close()

	var employees []*domain.Employee
	for rows.Next() {
		e, err := r.scanEmployeeRow(rows)
		if err != nil {
			slog.Error("employee_repository.List.scan_row", "error", err)
			return nil, err
		}
		employees = append(employees, e)
	}

	if err := rows.Err(); err != nil {
		slog.Error("employee_repository.List.rows_iteration", "error", err)
		return nil, err
	}

	return employees, nil
}

func (r *EmployeeRepository) ExistingGovernmentIDs(ctx context.Context, q ports.Querier, governmentIDs []string) ([]string, error) {
	if len(governmentIDs) == 0 {
		return nil, nil
	}
	return r.existingValues(ctx, q, "government_id", governmentIDs)
}

func (r *EmployeeRepository) ExistingMobiles(ctx context.Context, q ports.Querier, mobiles []string) ([]string, error) {
	if len(mobiles) == 0 {
		return nil, nil
	}
	return r.existingValues(ctx, q, "mobile", mobiles)
}

func (r *EmployeeRepository) ExistingUniversityIDs(ctx context.Context, q ports.Querier, universityIDs []string) ([]string, error) {
	if len(universityIDs) == 0 {
		return nil, nil
	}
	return r.existingValues(ctx, q, "university_id", universityIDs)
}

func (r *EmployeeRepository) existingValues(ctx context.Context, q ports.Querier, column string, values []string) ([]string, error) {
	placeholders := make([]byte, 0, len(values)*2)
	args := make([]any, len(values))
	for i, v := range values {
		if i > 0 {
			placeholders = append(placeholders, ',')
		}
		placeholders = append(placeholders, '?')
		args[i] = v
	}

	query := `SELECT ` + column + ` FROM employees WHERE ` + column + ` IN (` + string(placeholders) + `)`

	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		slog.Error("employee_repository.existingValues.query", "error", err, "column", column)
		return nil, err
	}
	defer rows.Close()

	var existing []string
	for rows.Next() {
		var val string
		if err := rows.Scan(&val); err != nil {
			slog.Error("employee_repository.existingValues.scan_row", "error", err, "column", column)
			return nil, err
		}
		existing = append(existing, val)
	}

	if err := rows.Err(); err != nil {
		slog.Error("employee_repository.existingValues.rows_iteration", "error", err, "column", column)
		return nil, err
	}

	return existing, nil
}

func (r *EmployeeRepository) scanEmployee(row *sql.Row) (*domain.Employee, error) {
	var e domain.Employee
	var hireDate, createdAt, updatedAt domain.Time
	var dateOfBirth, idCardValidUntil, subscriptionDate, actualAppointmentReappointmentDate, appointmentDecisionDate sql.NullString
	var decisionDate, gradeGrantDate sql.NullString
	var telephoneNumber, gender, religion, maritalStatus, address, placeOfBirth, placeOfResidence sql.NullString
	var policeStation, academicLevel, educationalQualification, universityName, faculty, specialization sql.NullString
	var yearObtained sql.NullInt64
	var appointmentDecisionNumber, appointmentType, departmentName sql.NullString
	var workEntity, employeeFileNumber, insuranceNumber, employmentStatus, militaryStatus, medicalCadre sql.NullString
	var personalPhotoURL, nationalCardImageURL, qualificationCertificateImageURL, cvURL sql.NullString
	var memberNumber, insuranceCode, jobGroup, qualitativeGroup, jobTitleAtLevel, jobTitleBeforePlacement sql.NullString
	var financialGrade, previousFinancialGrade, jobLevel, natureOfAppointment, notes, decisionFileURL sql.NullString
	var grade sql.NullString
	var decisionNumber sql.NullInt64
	err := row.Scan(
		&e.ID, &e.UID, &e.Name, &e.Mobile, &e.GovernmentID, &e.UniversityID,
		&e.Email, &hireDate, &e.Status, &e.Type, &e.SubType,
		&telephoneNumber, &e.PersonWithSpecialNeeds, &dateOfBirth, &gender, &religion, &maritalStatus,
		&address, &placeOfBirth, &placeOfResidence, &policeStation, &idCardValidUntil,
		&academicLevel, &educationalQualification, &universityName, &faculty, &specialization, &yearObtained,
		&actualAppointmentReappointmentDate, &appointmentDecisionDate, &appointmentDecisionNumber, &appointmentType, &departmentName,
		&grade, &workEntity, &employeeFileNumber, &insuranceNumber, &employmentStatus,
		&e.SolidarityFund, &subscriptionDate, &militaryStatus, &medicalCadre,
		&personalPhotoURL, &nationalCardImageURL, &qualificationCertificateImageURL, &cvURL,
		&memberNumber, &insuranceCode, &jobGroup, &qualitativeGroup, &jobTitleAtLevel,
		&jobTitleBeforePlacement, &financialGrade, &previousFinancialGrade, &jobLevel,
		&decisionDate, &decisionNumber, &gradeGrantDate, &natureOfAppointment, &notes, &e.Reappointment, &decisionFileURL,
		&e.AppointmentSeniorityOrGradeWithdrawal,
		&e.DepartmentUID, &e.ShiftUID, &e.ManagerUID, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		slog.Error("employee_repository.scanEmployee.scan_row", "error", err)
		return nil, err
	}
	e.TelephoneNumber = nullableStringPtr(telephoneNumber)
	e.DateOfBirth = nullableDatePtr(dateOfBirth)
	e.Gender = nullableStringPtr(gender)
	e.Religion = nullableStringPtr(religion)
	e.MaritalStatus = nullableStringPtr(maritalStatus)
	e.Address = nullableStringPtr(address)
	e.PlaceOfBirth = nullableStringPtr(placeOfBirth)
	e.PlaceOfResidence = nullableStringPtr(placeOfResidence)
	e.PoliceStation = nullableStringPtr(policeStation)
	e.IDCardValidUntil = nullableDatePtr(idCardValidUntil)
	e.AcademicLevel = nullableStringPtr(academicLevel)
	e.EducationalQualification = nullableStringPtr(educationalQualification)
	e.UniversityName = nullableStringPtr(universityName)
	e.Faculty = nullableStringPtr(faculty)
	e.Specialization = nullableStringPtr(specialization)
	e.YearObtained = nullableIntPtr(yearObtained)
	e.ActualAppointmentReappointmentDate = nullableDatePtr(actualAppointmentReappointmentDate)
	e.AppointmentDecisionDate = nullableDatePtr(appointmentDecisionDate)
	e.AppointmentDecisionNumber = nullableStringPtr(appointmentDecisionNumber)
	e.AppointmentType = nullableStringPtr(appointmentType)
	e.DepartmentName = nullableStringPtr(departmentName)
	e.Grade = nullableStringPtr(grade)
	e.WorkEntity = nullableStringPtr(workEntity)
	e.EmployeeFileNumber = nullableStringPtr(employeeFileNumber)
	e.InsuranceNumber = nullableStringPtr(insuranceNumber)
	e.EmploymentStatus = nullableStringPtr(employmentStatus)
	e.SubscriptionDate = nullableDatePtr(subscriptionDate)
	e.MilitaryStatus = nullableStringPtr(militaryStatus)
	e.MedicalCadre = nullableStringPtr(medicalCadre)
	e.PersonalPhotoURL = nullableStringPtr(personalPhotoURL)
	e.NationalCardImageURL = nullableStringPtr(nationalCardImageURL)
	e.QualificationCertificateImageURL = nullableStringPtr(qualificationCertificateImageURL)
	e.CVURL = nullableStringPtr(cvURL)
	e.MemberNumber = nullableStringPtr(memberNumber)
	e.InsuranceCode = nullableStringPtr(insuranceCode)
	e.JobGroup = nullableStringPtr(jobGroup)
	e.QualitativeGroup = nullableStringPtr(qualitativeGroup)
	e.JobTitleAtLevel = nullableStringPtr(jobTitleAtLevel)
	e.JobTitleBeforePlacement = nullableStringPtr(jobTitleBeforePlacement)
	e.FinancialGrade = nullableStringPtr(financialGrade)
	e.PreviousFinancialGrade = nullableStringPtr(previousFinancialGrade)
	e.JobLevel = nullableStringPtr(jobLevel)
	e.DecisionDate = nullableDatePtr(decisionDate)
	e.DecisionNumber = nullableInt64Ptr(decisionNumber)
	e.GradeGrantDate = nullableDatePtr(gradeGrantDate)
	e.NatureOfAppointment = nullableStringPtr(natureOfAppointment)
	e.Notes = nullableStringPtr(notes)
	e.DecisionFileURL = nullableStringPtr(decisionFileURL)
	e.HireDate = hireDate.Time
	e.CreatedAt = createdAt.Time
	e.UpdatedAt = updatedAt.Time
	return &e, nil
}

func (r *EmployeeRepository) scanEmployeeRow(rows *sql.Rows) (*domain.Employee, error) {
	var e domain.Employee
	var hireDate, createdAt, updatedAt domain.Time
	var dateOfBirth, idCardValidUntil, subscriptionDate, actualAppointmentReappointmentDate, appointmentDecisionDate sql.NullString
	var decisionDate, gradeGrantDate sql.NullString
	var telephoneNumber, gender, religion, maritalStatus, address, placeOfBirth, placeOfResidence sql.NullString
	var policeStation, academicLevel, educationalQualification, universityName, faculty, specialization sql.NullString
	var yearObtained sql.NullInt64
	var appointmentDecisionNumber, appointmentType, departmentName sql.NullString
	var workEntity, employeeFileNumber, insuranceNumber, employmentStatus, militaryStatus, medicalCadre sql.NullString
	var personalPhotoURL, nationalCardImageURL, qualificationCertificateImageURL, cvURL sql.NullString
	var memberNumber, insuranceCode, jobGroup, qualitativeGroup, jobTitleAtLevel, jobTitleBeforePlacement sql.NullString
	var financialGrade, previousFinancialGrade, jobLevel, natureOfAppointment, notes, decisionFileURL sql.NullString
	var grade sql.NullString
	var decisionNumber sql.NullInt64
	err := rows.Scan(
		&e.ID, &e.UID, &e.Name, &e.Mobile, &e.GovernmentID, &e.UniversityID,
		&e.Email, &hireDate, &e.Status, &e.Type, &e.SubType,
		&telephoneNumber, &e.PersonWithSpecialNeeds, &dateOfBirth, &gender, &religion, &maritalStatus,
		&address, &placeOfBirth, &placeOfResidence, &policeStation, &idCardValidUntil,
		&academicLevel, &educationalQualification, &universityName, &faculty, &specialization, &yearObtained,
		&actualAppointmentReappointmentDate, &appointmentDecisionDate, &appointmentDecisionNumber, &appointmentType, &departmentName,
		&grade, &workEntity, &employeeFileNumber, &insuranceNumber, &employmentStatus,
		&e.SolidarityFund, &subscriptionDate, &militaryStatus, &medicalCadre,
		&personalPhotoURL, &nationalCardImageURL, &qualificationCertificateImageURL, &cvURL,
		&memberNumber, &insuranceCode, &jobGroup, &qualitativeGroup, &jobTitleAtLevel,
		&jobTitleBeforePlacement, &financialGrade, &previousFinancialGrade, &jobLevel,
		&decisionDate, &decisionNumber, &gradeGrantDate, &natureOfAppointment, &notes, &e.Reappointment, &decisionFileURL,
		&e.AppointmentSeniorityOrGradeWithdrawal,
		&e.DepartmentUID, &e.ShiftUID, &e.ManagerUID, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	e.TelephoneNumber = nullableStringPtr(telephoneNumber)
	e.DateOfBirth = nullableDatePtr(dateOfBirth)
	e.Gender = nullableStringPtr(gender)
	e.Religion = nullableStringPtr(religion)
	e.MaritalStatus = nullableStringPtr(maritalStatus)
	e.Address = nullableStringPtr(address)
	e.PlaceOfBirth = nullableStringPtr(placeOfBirth)
	e.PlaceOfResidence = nullableStringPtr(placeOfResidence)
	e.PoliceStation = nullableStringPtr(policeStation)
	e.IDCardValidUntil = nullableDatePtr(idCardValidUntil)
	e.AcademicLevel = nullableStringPtr(academicLevel)
	e.EducationalQualification = nullableStringPtr(educationalQualification)
	e.UniversityName = nullableStringPtr(universityName)
	e.Faculty = nullableStringPtr(faculty)
	e.Specialization = nullableStringPtr(specialization)
	e.YearObtained = nullableIntPtr(yearObtained)
	e.ActualAppointmentReappointmentDate = nullableDatePtr(actualAppointmentReappointmentDate)
	e.AppointmentDecisionDate = nullableDatePtr(appointmentDecisionDate)
	e.AppointmentDecisionNumber = nullableStringPtr(appointmentDecisionNumber)
	e.AppointmentType = nullableStringPtr(appointmentType)
	e.DepartmentName = nullableStringPtr(departmentName)
	e.Grade = nullableStringPtr(grade)
	e.WorkEntity = nullableStringPtr(workEntity)
	e.EmployeeFileNumber = nullableStringPtr(employeeFileNumber)
	e.InsuranceNumber = nullableStringPtr(insuranceNumber)
	e.EmploymentStatus = nullableStringPtr(employmentStatus)
	e.SubscriptionDate = nullableDatePtr(subscriptionDate)
	e.MilitaryStatus = nullableStringPtr(militaryStatus)
	e.MedicalCadre = nullableStringPtr(medicalCadre)
	e.PersonalPhotoURL = nullableStringPtr(personalPhotoURL)
	e.NationalCardImageURL = nullableStringPtr(nationalCardImageURL)
	e.QualificationCertificateImageURL = nullableStringPtr(qualificationCertificateImageURL)
	e.CVURL = nullableStringPtr(cvURL)
	e.MemberNumber = nullableStringPtr(memberNumber)
	e.InsuranceCode = nullableStringPtr(insuranceCode)
	e.JobGroup = nullableStringPtr(jobGroup)
	e.QualitativeGroup = nullableStringPtr(qualitativeGroup)
	e.JobTitleAtLevel = nullableStringPtr(jobTitleAtLevel)
	e.JobTitleBeforePlacement = nullableStringPtr(jobTitleBeforePlacement)
	e.FinancialGrade = nullableStringPtr(financialGrade)
	e.PreviousFinancialGrade = nullableStringPtr(previousFinancialGrade)
	e.JobLevel = nullableStringPtr(jobLevel)
	e.DecisionDate = nullableDatePtr(decisionDate)
	e.DecisionNumber = nullableInt64Ptr(decisionNumber)
	e.GradeGrantDate = nullableDatePtr(gradeGrantDate)
	e.NatureOfAppointment = nullableStringPtr(natureOfAppointment)
	e.Notes = nullableStringPtr(notes)
	e.DecisionFileURL = nullableStringPtr(decisionFileURL)
	e.HireDate = hireDate.Time
	e.CreatedAt = createdAt.Time
	e.UpdatedAt = updatedAt.Time
	return &e, nil
}

func nullableStringPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func nullableInt64Ptr(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	return &value.Int64
}

func nullableIntPtr(value sql.NullInt64) *int {
	if !value.Valid {
		return nil
	}
	converted := int(value.Int64)
	return &converted
}

func nullableDate(value *time.Time) any {
	if value == nil {
		return nil
	}
	return domain.Date{Time: *value}
}

func nullableDatePtr(value sql.NullString) *time.Time {
	if !value.Valid || value.String == "" {
		return nil
	}
	var parsed domain.Date
	if err := parsed.Scan(value.String); err != nil {
		slog.Error("employee_repository.nullableDatePtr.scan", "error", err, "value", value.String)
		return nil
	}
	t := parsed.Time
	return &t
}

func (r *EmployeeRepository) Count(ctx context.Context, q ports.Querier) (int, error) {
	query := `SELECT COUNT(*) FROM employees`

	var count int
	err := q.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		slog.Error("employee_repository.Count.scan", "error", err)
		return 0, err
	}
	return count, nil
}
