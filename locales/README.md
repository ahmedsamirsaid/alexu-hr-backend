# Translation Files

This directory contains JSON translation files for the Banu Musa ERP backend i18n system.

## Directory Structure

```
locales/
├── en/          # English translations
│   ├── audit.json
│   ├── errors.json
│   ├── notifications.json
│   └── validation.json
└── ar/          # Arabic translations
    ├── audit.json
    ├── errors.json
    ├── notifications.json
    └── validation.json
```

## File Format

All translation files must:
- Use **UTF-8 encoding** (required for Arabic character support)
- Follow valid JSON syntax
- Use nested objects for hierarchical keys
- Contain only string values

## Translation Key Naming Conventions

| Category | Pattern | Example |
|---|---|---|
| Audit Actions | `audit.action.{action}` | `audit.action.approve` |
| Audit Entities | `audit.entity.{entity}` | `audit.entity.leave_request` |
| Audit Metadata | `audit.metadata.{field}` | `audit.metadata.old_value` |
| Errors | `error.{category}.{error}` | `error.leave.insufficient_balance` |
| Validation | `validation.{rule}` | `validation.required_field` |
| Notifications | `notification.{type}.{part}` | `notification.leave_approved.title` |

## Example Translation File

```json
{
  "audit": {
    "action": {
      "create": "Created",
      "update": "Updated",
      "delete": "Deleted"
    },
    "entity": {
      "employee": "Employee",
      "leave_request": "Leave Request"
    }
  }
}
```

## Adding New Translations

1. Add the translation key to both `en/` and `ar/` files
2. Ensure UTF-8 encoding is preserved
3. Follow the naming conventions above
4. Restart the application to load new translations

## UTF-8 Encoding

All JSON files in this directory **must** use UTF-8 encoding to properly support:
- Arabic characters (العربية)
- Special characters and diacritics
- Emoji and symbols

When editing these files, ensure your text editor is configured to save as UTF-8.
