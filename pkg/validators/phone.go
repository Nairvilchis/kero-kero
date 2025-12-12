package validators

import (
	"regexp"
	"strings"

	"kero-kero/pkg/errors"
)

// ValidatePhoneNumber valida y limpia un número de teléfono para WhatsApp
// Acepta números con o sin código de país, con o sin espacios/guiones
func ValidatePhoneNumber(phone string) (string, error) {
	if phone == "" {
		return "", errors.New(400, "El número de teléfono es requerido")
	}

	// Eliminar espacios, guiones, paréntesis y el símbolo +
	cleaned := strings.ReplaceAll(phone, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")
	cleaned = strings.ReplaceAll(cleaned, "+", "")

	// Validar que solo contenga dígitos
	if !regexp.MustCompile(`^\d+$`).MatchString(cleaned) {
		return "", errors.New(400, "El número de teléfono solo debe contener dígitos")
	}

	// Validar longitud (mínimo 10, máximo 15 dígitos según estándar E.164)
	if len(cleaned) < 10 {
		return "", errors.New(400, "El número de teléfono es demasiado corto (mínimo 10 dígitos)")
	}

	if len(cleaned) > 15 {
		return "", errors.New(400, "El número de teléfono es demasiado largo (máximo 15 dígitos)")
	}

	// Devolver el número limpio y validado
	return cleaned, nil
}

// MaskPhoneNumber enmascara un número de teléfono para logs
// Ejemplo: "5215512345678" -> "52****5678"
func MaskPhoneNumber(phone string) string {
	if len(phone) < 4 {
		return "***"
	}

	// Mostrar primeros 2 dígitos (código de país) y últimos 4
	if len(phone) > 6 {
		return phone[:2] + "****" + phone[len(phone)-4:]
	}

	// Si es muy corto, solo mostrar últimos 2
	return "**" + phone[len(phone)-2:]
}
