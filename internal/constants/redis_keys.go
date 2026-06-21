package constants

// Redis key patterns and TTL constants.
const (
	// Redis key prefixes
	KeyRolePermissions  = "role:permissions:"  // + {role_id}
	KeyJWTBlacklist     = "jwt:blacklist:"     // + {jti}
	KeyRateLimit        = "ratelimit:"         // + {ip}:{endpoint}
	KeyPermissionsAll   = "permissions:all"
	KeyBahanBakuAll     = "bahan_baku:all"
	KeyItemsJenis       = "items:jenis:"       // + {jenis}
	KeyDivisiAll        = "divisi:all"
	KeyJenisPengukuran  = "jenis_pengukuran:all"
	KeyTerminAll        = "termin:all"
	KeyUserSession      = "user:session:"      // + {user_id}
	KeyOrderAll         = "order:all"
	KeySurveyAll        = "survey:all"
	KeyMoodboardAll     = "moodboard:all"
	KeySettingAll       = "setting:all"
)
