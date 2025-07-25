package postgres

const (
	fieldID               = "id"
	fieldUserID           = "user_id"
	fieldClientDiagramID  = "client_diagram_id"
	fieldCode             = "code"
	fieldObjectStorageKey = "object_storage_key"
	fieldName             = "name"
	fieldTablesCount      = "tables_count"

	fieldLogin        = "login"
	fieldPasswordHash = "password_hash"
	fieldType         = "type"
	fieldConfirmedAt  = "confirmed_at"
	fieldExpiresAt    = "expires_at"

	fieldCreatedAt = "created_at"
	fieldUpdatedAt = "updated_at"
	fieldDeletedAt = "deleted_at"
)

const (
	returning = "RETURNING "
	separator = ","
)

const (
	asc  = "asc"
	desc = "desc"
)
