package helpers

import "strings"

// IsDatabaseLockedError verifica si un error es causado por una base de datos bloqueada
// Este es un error común en SQLite cuando hay múltiples escrituras concurrentes
func IsDatabaseLockedError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	return strings.Contains(errStr, "database is locked") ||
		strings.Contains(errStr, "failed to prefetch sessions: database is locked") ||
		strings.Contains(errStr, "database locked")
}

// IsTemporaryError verifica si un error es temporal y puede reintentarse
func IsTemporaryError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())

	// Errores de base de datos temporales
	if IsDatabaseLockedError(err) {
		return true
	}

	// Errores de red temporales
	temporaryPatterns := []string{
		"timeout",
		"connection refused",
		"connection reset",
		"temporary failure",
		"try again",
		"service unavailable",
		"too many requests",
	}

	for _, pattern := range temporaryPatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}
