#!/usr/bin/env python3
"""Migrate all audit action sentences to action_key + action_params pattern."""
import re, os

BASE = os.path.dirname(os.path.abspath(__file__))

# Each entry: (filename, action_key, old_sprintf_pattern_regex, params_map_code, old_withmeta_line)
# We do two replacements per file:
#   1. Replace the actionSentence := fmt.Sprintf(...) block with actionParams map
#   2. Replace WithMeta("action", actionSentence) with action_key + action_params

MIGRATIONS = {
    "core/usecases/approve_request.go": [
        {
            "old_sentence": (
                '\tactionSentence := fmt.Sprintf(\n'
                '\t\t"%s (%s) approved %s (Employee\'s %s leave request for %s to %s (%d days)",\n'
                '\t\tactorName, actorRoleName, requester.Name, leaveTypeName,\n'
                '\t\tleaveStartDate, leaveEndDate, leaveDays,\n'
                '\t)'
            ),
            "new_params": (
                '\tactionParams := map[string]interface{}{\n'
                '\t\t"Actor":     actorName,\n'
                '\t\t"ActorRole": actorRoleName,\n'
                '\t\t"Requester": requester.Name,\n'
                '\t\t"LeaveType": leaveTypeName,\n'
                '\t\t"StartDate": leaveStartDate,\n'
                '\t\t"EndDate":   leaveEndDate,\n'
                '\t\t"Days":      leaveDays,\n'
                '\t}'
            ),
            "old_meta": '\t\tWithMeta("action", actionSentence).',
            "new_meta": '\t\tWithMeta("action_key", "audit.sentence.approve_leave").\n\t\tWithMeta("action_params", actionParams).',
        },
        {
            "old_sentence": (
                '\t\t\tWithMeta("action", fmt.Sprintf(\n'
                '\t\t\t\t"%s approved %s\'s %s leave — deducted %.1f days from balance (was %.1f used, now %.1f used)",\n'
                '\t\t\t\tactorName, requester.Name, leaveTypeName,\n'
                '\t\t\t\tfloat64(leaveRequest.Days), float64(oldUsedDays), float64(balance.UsedDays),\n'
                '\t\t\t)).'
            ),
            "new_params": (
                '\t\t\tWithMeta("action_key", "audit.sentence.deduct_leave_balance").\n'
                '\t\t\tWithMeta("action_params", map[string]interface{}{\n'
                '\t\t\t\t"Actor":     actorName,\n'
                '\t\t\t\t"Requester": requester.Name,\n'
                '\t\t\t\t"LeaveType": leaveTypeName,\n'
                '\t\t\t\t"Days":      leaveRequest.Days,\n'
                '\t\t\t\t"OldUsed":   oldUsedDays,\n'
                '\t\t\t\t"NewUsed":   balance.UsedDays,\n'
                '\t\t\t}).'
            ),
            "old_meta": None,
            "new_meta": None,
        },
    ],
}

SIMPLE_MIGRATIONS = [
    ("core/usecases/reject_request.go",
     r'actionSentence := fmt\.Sprintf\(\s*"[^"]+",\s*actorName, actorRoleName, requester\.Name, leaveTypeName,\s*leaveStartDate, leaveEndDate, leaveDays, reasonStr,\s*\)',
     'actionParams := map[string]interface{}{\n\t\t"Actor":     actorName,\n\t\t"ActorRole": actorRoleName,\n\t\t"Requester": requester.Name,\n\t\t"LeaveType": leaveTypeName,\n\t\t"StartDate": leaveStartDate,\n\t\t"EndDate":   leaveEndDate,\n\t\t"Days":      leaveDays,\n\t\t"Reason":    reasonStr,\n\t}',
     "audit.sentence.reject_leave"),
    ("core/usecases/create_manual_holiday.go",
     r'actionSentence := fmt\.Sprintf\(\s*"[^"]+",\s*actorName, nameEN, dateOnly\.Format\("[^"]+"\),\s*\)',
     'actionParams := map[string]interface{}{\n\t\t"Actor": actorName,\n\t\t"Name":  nameEN,\n\t\t"Date":  dateOnly.Format("Jan 2, 2006"),\n\t}',
     "audit.sentence.create_holiday"),
    ("core/usecases/update_holiday.go",
     r'actionSentence := fmt\.Sprintf\(\s*"[^"]+",\s*actorName, definition\.NameEN,\s*\)',
     'actionParams := map[string]interface{}{\n\t\t"Actor": actorName,\n\t\t"Name":  definition.NameEN,\n\t}',
     "audit.sentence.update_holiday"),
    ("core/usecases/delete_holiday.go",
     r'actionSentence := fmt\.Sprintf\(\s*"[^"]+",\s*actorName, definition\.NameEN, definition\.Date\.Format\("[^"]+"\),\s*\)',
     'actionParams := map[string]interface{}{\n\t\t"Actor": actorName,\n\t\t"Name":  definition.NameEN,\n\t\t"Date":  definition.Date.Format("Jan 2, 2006"),\n\t}',
     "audit.sentence.delete_holiday"),
    ("core/usecases/assign_department_manager.go",
     r'actionSentence := fmt\.Sprintf\(\s*"[^"]+",\s*actorName, managerName, department\.NameEN,\s*\)',
     'actionParams := map[string]interface{}{\n\t\t"Actor":      actorName,\n\t\t"Manager":    managerName,\n\t\t"Department": department.NameEN,\n\t}',
     "audit.sentence.assign_department_manager"),
    ("core/usecases/assign_employee_department.go",
     r'actionSentence := fmt\.Sprintf\(\s*"[^"]+",\s*actorName, employee\.Name, department\.NameEN,\s*\)',
     'actionParams := map[string]interface{}{\n\t\t"Actor":      actorName,\n\t\t"Employee":   employee.Name,\n\t\t"Department": department.NameEN,\n\t}',
     "audit.sentence.assign_employee_department"),
    ("core/usecases/remove_employee_department.go",
     r'actionSentence := fmt\.Sprintf\(\s*"[^"]+",\s*actorName, employee\.Name,\s*\)',
     'actionParams := map[string]interface{}{\n\t\t"Actor":    actorName,\n\t\t"Employee": employee.Name,\n\t}',
     "audit.sentence.remove_employee_department"),
    ("core/usecases/role_create.go",
     r'actionSentence := fmt\.Sprintf\("[^"]+", role\.Name, role\.ScopeType\)',
     'actionParams := map[string]interface{}{\n\t\t"Name":      role.Name,\n\t\t"ScopeType": role.ScopeType,\n\t}',
     "audit.sentence.create_role"),
    ("core/usecases/role_set_scope.go",
     r'actionSentence := fmt\.Sprintf\(\s*"[^"]+",\s*actorName, role\.Name, oldScopeType, normalizedScopeType,\s*\)',
     'actionParams := map[string]interface{}{\n\t\t"Actor":    actorName,\n\t\t"Role":     role.Name,\n\t\t"OldScope": oldScopeType,\n\t\t"NewScope": normalizedScopeType,\n\t}',
     "audit.sentence.set_role_scope"),
    ("core/usecases/role_set_permissions.go",
     r'actionSentence := fmt\.Sprintf\(\s*"[^"]+",\s*role\.Name, len\(addedPerms\), len\(removedPerms\),\s*\)',
     'actionParams := map[string]interface{}{\n\t\t"Role":    role.Name,\n\t\t"Added":   len(addedPerms),\n\t\t"Removed": len(removedPerms),\n\t}',
     "audit.sentence.set_role_permissions"),
    ("core/usecases/create_department.go",
     r'actionSentence := fmt\.Sprintf\("[^"]+", input\.NameEN, input\.Code\)',
     'actionParams := map[string]interface{}{\n\t\t"Name": input.NameEN,\n\t\t"Code": input.Code,\n\t}',
     "audit.sentence.create_department"),
    ("core/usecases/update_leave_type.go",
     r'actionSentence := fmt\.Sprintf\(\s*"[^"]+",\s*actorName, leaveType\.NameEN,\s*\)',
     'actionParams := map[string]interface{}{\n\t\t"Actor": actorName,\n\t\t"Name":  leaveType.NameEN,\n\t}',
     "audit.sentence.update_leave_type"),
    ("core/usecases/create_approval_flow.go",
     r'actionSentence := fmt\.Sprintf\(\s*"[^"]+",\s*actorName, input\.NameEN, input\.Code,\s*\)',
     'actionParams := map[string]interface{}{\n\t\t"Actor": actorName,\n\t\t"Name":  input.NameEN,\n\t\t"Code":  input.Code,\n\t}',
     "audit.sentence.create_approval_flow"),
    ("core/usecases/update_approval_flow.go",
     r'actionSentence := fmt\.Sprintf\("[^"]+", actorName, flow\.NameEN\)',
     'actionParams := map[string]interface{}{\n\t\t"Actor": actorName,\n\t\t"Name":  flow.NameEN,\n\t}',
     "audit.sentence.update_approval_flow"),
    ("core/usecases/update_approval_flow_step.go",
     r'actionSentence := fmt\.Sprintf\("[^"]+", actorName\)',
     'actionParams := map[string]interface{}{\n\t\t"Actor": actorName,\n\t}',
     "audit.sentence.update_approval_flow_step"),
    ("core/usecases/delete_approval_flow_step.go",
     r'actionSentence := fmt\.Sprintf\(\s*"[^"]+",\s*actorName, stepOrder,\s*\)',
     'actionParams := map[string]interface{}{\n\t\t"Actor":     actorName,\n\t\t"StepOrder": stepOrder,\n\t}',
     "audit.sentence.delete_approval_flow_step"),
    ("core/usecases/delete_attendance_device.go",
     r'actionSentence := fmt\.Sprintf\(\s*"[^"]+",\s*actorName, existing\.Name, existing\.IP, existing\.Port,\s*\)',
     'actionParams := map[string]interface{}{\n\t\t"Actor":  actorName,\n\t\t"Device": existing.Name,\n\t\t"IP":     existing.IP,\n\t\t"Port":   existing.Port,\n\t}',
     "audit.sentence.delete_attendance_device"),
    ("core/usecases/activate_attendance_device.go",
     r'actionSentence := fmt\.Sprintf\(\s*"[^"]+",\s*actorName, existing\.Name, existing\.IP, existing\.Port,\s*\)',
     'actionParams := map[string]interface{}{\n\t\t"Actor":  actorName,\n\t\t"Device": existing.Name,\n\t\t"IP":     existing.IP,\n\t\t"Port":   existing.Port,\n\t}',
     "audit.sentence.activate_attendance_device"),
    ("core/usecases/update_attendance_device.go",
     r'actionSentence := fmt\.Sprintf\("[^"]+", actorName, existing\.Name\)',
     'actionParams := map[string]interface{}{\n\t\t"Actor":  actorName,\n\t\t"Device": existing.Name,\n\t}',
     "audit.sentence.update_attendance_device"),
    ("core/usecases/register_attendance_device.go",
     r'actionSentence := fmt\.Sprintf\(\s*"[^"]+",\s*actorName, device\.Name, device\.IP, device\.Port, device\.SerialNumber,\s*\)',
     'actionParams := map[string]interface{}{\n\t\t"Actor":        actorName,\n\t\t"Device":       device.Name,\n\t\t"IP":           device.IP,\n\t\t"Port":         device.Port,\n\t\t"SerialNumber": device.SerialNumber,\n\t}',
     "audit.sentence.register_attendance_device"),
    ("core/usecases/update_department.go",
     r'actionSentence := fmt\.Sprintf\("[^"]+", department\.NameEN\)',
     'actionParams := map[string]interface{}{\n\t\t"Name": department.NameEN,\n\t}',
     "audit.sentence.update_department"),
    ("core/usecases/update_rejected_leave_request.go",
     r'actionSentence := fmt\.Sprintf\(\s*"[^"]+",\s*actorName,\s*\)',
     'actionParams := map[string]interface{}{\n\t\t"Actor": actorName,\n\t}',
     "audit.sentence.resubmit_leave"),
]

def replace_withmeta_action(content, action_key):
    """Replace WithMeta("action", actionSentence) with action_key + action_params."""
    # Match various indentation patterns
    patterns = [
        (r'WithMeta\("action", actionSentence\)\.', f'WithMeta("action_key", "{action_key}").\n\t\tWithMeta("action_params", actionParams).'),
        (r'WithMeta\("action", actionSentence\)\n', f'WithMeta("action_key", "{action_key}").\n\t\tWithMeta("action_params", actionParams)\n'),
        (r'WithMeta\("action", actionSentence\)', f'WithMeta("action_key", "{action_key}").\n\t\tWithMeta("action_params", actionParams)'),
    ]
    for old, new in patterns:
        new_content = re.sub(old, new, content, count=1)
        if new_content != content:
            return new_content
    return content

changed_files = []
errors = []

for relpath, pattern, new_params, action_key in SIMPLE_MIGRATIONS:
    filepath = os.path.join(BASE, relpath)
    if not os.path.exists(filepath):
        errors.append(f"NOT FOUND: {relpath}")
        continue
    with open(filepath, 'r') as f:
        content = f.read()
    
    # Replace actionSentence assignment
    new_content = re.sub(pattern, new_params, content, count=1, flags=re.DOTALL)
    if new_content == content:
        errors.append(f"PATTERN NOT MATCHED (sentence): {relpath}")
        continue
    
    # Replace WithMeta("action", actionSentence)
    new_content = replace_withmeta_action(new_content, action_key)
    
    with open(filepath, 'w') as f:
        f.write(new_content)
    changed_files.append(relpath)

print(f"Changed {len(changed_files)} files:")
for f in changed_files:
    print(f"  ✅ {f}")
if errors:
    print(f"\nErrors ({len(errors)}):")
    for e in errors:
        print(f"  ❌ {e}")
