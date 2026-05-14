package db

import (
	"context"
	"testing"
	"time"

	"github.com/banumusa/backend/core/domain"
)

func TestEmployeeRepository_Create(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewEmployeeRepository()
	ctx := context.Background()

	employee := domain.NewEmployee(
		"Ahmed Hassan",
		"01012345678",
		"29901011234567",
		"UNI001",
		time.Date(2020, 6, 15, 0, 0, 0, 0, time.UTC),
	)

	err := repo.Create(ctx, tdb.SQLiteDB, employee)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if employee.ID == 0 {
		t.Error("Create() did not set ID")
	}
	if employee.CreatedAt.IsZero() {
		t.Error("Create() did not set CreatedAt")
	}
	if employee.UpdatedAt.IsZero() {
		t.Error("Create() did not set UpdatedAt")
	}
	if employee.Type != domain.EmployeeTypePermanent {
		t.Errorf("Create() Type = %v, want %v", employee.Type, domain.EmployeeTypePermanent)
	}
	if employee.SubType != domain.EmployeeSubTypeNormal {
		t.Errorf("Create() SubType = %v, want %v", employee.SubType, domain.EmployeeSubTypeNormal)
	}
}

func TestEmployeeRepository_Create_DuplicateMobile(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewEmployeeRepository()
	ctx := context.Background()

	employee1 := domain.NewEmployee("Ahmed", "01012345678", "gov1", "uni1", time.Now())
	employee2 := domain.NewEmployee("Mohamed", "01012345678", "gov2", "uni2", time.Now())

	if err := repo.Create(ctx, tdb.SQLiteDB, employee1); err != nil {
		t.Fatalf("Create() first employee error = %v", err)
	}

	err := repo.Create(ctx, tdb.SQLiteDB, employee2)
	if err == nil {
		t.Error("Create() expected error for duplicate mobile, got nil")
	}
}

func TestEmployeeRepository_GetByID(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewEmployeeRepository()
	ctx := context.Background()

	// Create employee
	employee := tdb.SeedEmployee("Ahmed Hassan")

	// Get by ID
	found, err := repo.GetByID(ctx, tdb.SQLiteDB, employee.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if found == nil {
		t.Fatal("GetByID() returned nil")
	}
	if found.Name != employee.Name {
		t.Errorf("GetByID() Name = %v, want %v", found.Name, employee.Name)
	}
	if found.UID != employee.UID {
		t.Errorf("GetByID() UID = %v, want %v", found.UID, employee.UID)
	}
	if found.Type != employee.Type {
		t.Errorf("GetByID() Type = %v, want %v", found.Type, employee.Type)
	}
	if found.SubType != employee.SubType {
		t.Errorf("GetByID() SubType = %v, want %v", found.SubType, employee.SubType)
	}
}

func TestEmployeeRepository_GetByID_NotFound(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewEmployeeRepository()
	ctx := context.Background()

	found, err := repo.GetByID(ctx, tdb.SQLiteDB, 99999)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if found != nil {
		t.Errorf("GetByID() expected nil for non-existent ID, got %+v", found)
	}
}

func TestEmployeeRepository_GetByUID(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewEmployeeRepository()
	ctx := context.Background()

	// Create employee
	employee := tdb.SeedEmployee("Mohamed Ali")

	// Get by UID
	found, err := repo.GetByUID(ctx, tdb.SQLiteDB, employee.UID)
	if err != nil {
		t.Fatalf("GetByUID() error = %v", err)
	}
	if found == nil {
		t.Fatal("GetByUID() returned nil")
	}
	if found.ID != employee.ID {
		t.Errorf("GetByUID() ID = %v, want %v", found.ID, employee.ID)
	}
	if found.Name != employee.Name {
		t.Errorf("GetByUID() Name = %v, want %v", found.Name, employee.Name)
	}
}

func TestEmployeeRepository_GetByUID_NotFound(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewEmployeeRepository()
	ctx := context.Background()

	found, err := repo.GetByUID(ctx, tdb.SQLiteDB, "emp_nonexistent")
	if err != nil {
		t.Fatalf("GetByUID() error = %v", err)
	}
	if found != nil {
		t.Errorf("GetByUID() expected nil for non-existent UID, got %+v", found)
	}
}

func TestEmployeeRepository_Update(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewEmployeeRepository()
	ctx := context.Background()

	// Create employee
	employee := tdb.SeedEmployee("Old Name")
	originalUpdatedAt := employee.UpdatedAt

	// Update
	time.Sleep(10 * time.Millisecond) // Ensure time difference
	employee.Name = "New Name"
	employee.Status = domain.EmployeeStatusInactive
	employee.Type = domain.EmployeeTypeTemporary
	employee.SubType = domain.EmployeeSubTypeComprehensiveBonus

	err := repo.Update(ctx, tdb.SQLiteDB, employee)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if employee.UpdatedAt.Equal(originalUpdatedAt) {
		t.Error("Update() did not update UpdatedAt")
	}

	// Verify update persisted
	found, err := repo.GetByID(ctx, tdb.SQLiteDB, employee.ID)
	if err != nil {
		t.Fatalf("GetByID() after update error = %v", err)
	}
	if found.Name != "New Name" {
		t.Errorf("Update() Name = %v, want %v", found.Name, "New Name")
	}
	if found.Status != domain.EmployeeStatusInactive {
		t.Errorf("Update() Status = %v, want %v", found.Status, domain.EmployeeStatusInactive)
	}
	if found.Type != domain.EmployeeTypeTemporary {
		t.Errorf("Update() Type = %v, want %v", found.Type, domain.EmployeeTypeTemporary)
	}
	if found.SubType != domain.EmployeeSubTypeComprehensiveBonus {
		t.Errorf("Update() SubType = %v, want %v", found.SubType, domain.EmployeeSubTypeComprehensiveBonus)
	}
}

func TestEmployeeRepository_Create_WithEmail(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewEmployeeRepository()
	ctx := context.Background()

	email := "ahmed@example.com"
	employee := domain.NewEmployee(
		"Ahmed",
		"01098765432",
		"gov123",
		"uni123",
		time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	)
	employee.Email = &email

	if err := repo.Create(ctx, tdb.SQLiteDB, employee); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	found, err := repo.GetByID(ctx, tdb.SQLiteDB, employee.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if found.Email == nil || *found.Email != email {
		t.Errorf("GetByID() Email = %v, want %v", found.Email, &email)
	}
}

func TestEmployeeRepository_Create_WithExtendedProfileFields(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewEmployeeRepository()
	ctx := context.Background()

	telephone := "0223456789"
	gender := "male"
	yearObtained := 2012
	appointmentDecisionNumber := "12345"
	appointmentType := "new"
	departmentName := "xxx"
	natureOfAppointment := "permanent"
	grade := "3.5"
	workEntity := "HR"
	decisionNumber := int64(42)
	personalPhotoURL := "minio://employees/photo.jpg"
	dateOfBirth := time.Date(1990, 5, 10, 0, 0, 0, 0, time.UTC)
	subscriptionDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	actualAppointmentReappointmentDate := time.Date(2015, 2, 2, 0, 0, 0, 0, time.UTC)
	appointmentDecisionDate := time.Date(2015, 2, 2, 0, 0, 0, 0, time.UTC)

	employee := domain.NewEmployee(
		"Ahmed",
		"01077777777",
		"gov-profile-1",
		"uni-profile-1",
		time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	)
	employee.TelephoneNumber = &telephone
	employee.PersonWithSpecialNeeds = true
	employee.DateOfBirth = &dateOfBirth
	employee.Gender = &gender
	employee.YearObtained = &yearObtained
	employee.ActualAppointmentReappointmentDate = &actualAppointmentReappointmentDate
	employee.AppointmentDecisionDate = &appointmentDecisionDate
	employee.AppointmentDecisionNumber = &appointmentDecisionNumber
	employee.AppointmentType = &appointmentType
	employee.DepartmentName = &departmentName
	employee.Grade = &grade
	employee.WorkEntity = &workEntity
	employee.SolidarityFund = true
	employee.SubscriptionDate = &subscriptionDate
	employee.DecisionNumber = &decisionNumber
	employee.NatureOfAppointment = &natureOfAppointment
	employee.PersonalPhotoURL = &personalPhotoURL
	employee.Reappointment = true
	employee.AppointmentSeniorityOrGradeWithdrawal = true

	if err := repo.Create(ctx, tdb.SQLiteDB, employee); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	found, err := repo.GetByID(ctx, tdb.SQLiteDB, employee.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if found.TelephoneNumber == nil || *found.TelephoneNumber != telephone {
		t.Fatalf("TelephoneNumber = %v, want %q", found.TelephoneNumber, telephone)
	}
	if !found.PersonWithSpecialNeeds {
		t.Fatal("PersonWithSpecialNeeds = false, want true")
	}
	if found.DateOfBirth == nil || !found.DateOfBirth.Equal(dateOfBirth) {
		t.Fatalf("DateOfBirth = %v, want %v", found.DateOfBirth, dateOfBirth)
	}
	if found.Gender == nil || *found.Gender != gender {
		t.Fatalf("Gender = %v, want %q", found.Gender, gender)
	}
	if found.YearObtained == nil || *found.YearObtained != yearObtained {
		t.Fatalf("YearObtained = %v, want %d", found.YearObtained, yearObtained)
	}
	if found.ActualAppointmentReappointmentDate == nil || !found.ActualAppointmentReappointmentDate.Equal(actualAppointmentReappointmentDate) {
		t.Fatalf("ActualAppointmentReappointmentDate = %v, want %v", found.ActualAppointmentReappointmentDate, actualAppointmentReappointmentDate)
	}
	if found.AppointmentDecisionDate == nil || !found.AppointmentDecisionDate.Equal(appointmentDecisionDate) {
		t.Fatalf("AppointmentDecisionDate = %v, want %v", found.AppointmentDecisionDate, appointmentDecisionDate)
	}
	if found.AppointmentDecisionNumber == nil || *found.AppointmentDecisionNumber != appointmentDecisionNumber {
		t.Fatalf("AppointmentDecisionNumber = %v, want %q", found.AppointmentDecisionNumber, appointmentDecisionNumber)
	}
	if found.AppointmentType == nil || *found.AppointmentType != appointmentType {
		t.Fatalf("AppointmentType = %v, want %q", found.AppointmentType, appointmentType)
	}
	if found.DepartmentName == nil || *found.DepartmentName != departmentName {
		t.Fatalf("DepartmentName = %v, want %q", found.DepartmentName, departmentName)
	}
	if found.Grade == nil || *found.Grade != grade {
		t.Fatalf("Grade = %v, want %v", found.Grade, grade)
	}
	if found.WorkEntity == nil || *found.WorkEntity != workEntity {
		t.Fatalf("WorkEntity = %v, want %q", found.WorkEntity, workEntity)
	}
	if !found.SolidarityFund {
		t.Fatal("SolidarityFund = false, want true")
	}
	if found.SubscriptionDate == nil || !found.SubscriptionDate.Equal(subscriptionDate) {
		t.Fatalf("SubscriptionDate = %v, want %v", found.SubscriptionDate, subscriptionDate)
	}
	if found.DecisionNumber == nil || *found.DecisionNumber != decisionNumber {
		t.Fatalf("DecisionNumber = %v, want %d", found.DecisionNumber, decisionNumber)
	}
	if found.NatureOfAppointment == nil || *found.NatureOfAppointment != natureOfAppointment {
		t.Fatalf("NatureOfAppointment = %v, want %q", found.NatureOfAppointment, natureOfAppointment)
	}
	if found.PersonalPhotoURL == nil || *found.PersonalPhotoURL != personalPhotoURL {
		t.Fatalf("PersonalPhotoURL = %v, want %q", found.PersonalPhotoURL, personalPhotoURL)
	}
	if !found.Reappointment {
		t.Fatal("Reappointment = false, want true")
	}
	if !found.AppointmentSeniorityOrGradeWithdrawal {
		t.Fatal("AppointmentSeniorityOrGradeWithdrawal = false, want true")
	}
}

func TestEmployeeRepository_Transaction(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewEmployeeRepository()
	ctx := context.Background()

	// Start transaction
	tx, err := tdb.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx() error = %v", err)
	}

	// Create employee in transaction
	employee := domain.NewEmployee("TxEmployee", "01055555555", "govtx", "unitx", time.Now())
	if err := repo.Create(ctx, tx, employee); err != nil {
		tx.Rollback()
		t.Fatalf("Create() in tx error = %v", err)
	}

	// Rollback transaction
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback() error = %v", err)
	}

	// Verify employee was not persisted
	found, err := repo.GetByID(ctx, tdb.SQLiteDB, employee.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if found != nil {
		t.Error("Employee should not exist after rollback")
	}
}

func TestEmployeeRepository_Transaction_Commit(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewEmployeeRepository()
	ctx := context.Background()

	// Start transaction
	tx, err := tdb.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx() error = %v", err)
	}

	// Create employee in transaction
	employee := domain.NewEmployee("TxEmployee", "01066666666", "govtx2", "unitx2", time.Now())
	if err := repo.Create(ctx, tx, employee); err != nil {
		tx.Rollback()
		t.Fatalf("Create() in tx error = %v", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit() error = %v", err)
	}

	// Verify employee was persisted
	found, err := repo.GetByID(ctx, tdb.SQLiteDB, employee.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if found == nil {
		t.Error("Employee should exist after commit")
	}
}
