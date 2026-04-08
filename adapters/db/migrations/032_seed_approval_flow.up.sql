-- Seed default leave approval flow
INSERT INTO approval_flows (uid, code, name_en, name_ar, description, is_active)
VALUES ('apf_leave_default', 'leave', 'Leave Request Approval', 'اعتماد طلب الإجازة', 'Default approval flow for leave requests requiring approval', 1);

-- Seed Department Manager role if not exists
INSERT OR IGNORE INTO roles (uid, name, description, is_system)
VALUES ('role_department_manager', 'Department Manager', 'Department Manager with approval authority for their department', 0);

-- Seed default approval flow step (Step 1: department_manager)
INSERT INTO approval_flow_steps (uid, approval_flow_uid, step_order, role_uid)
VALUES ('afs_leave_step1', 'apf_leave_default', 1, 'role_department_manager');
