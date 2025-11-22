package utils

import (
	"regexp"
	"strings"
)

// IsNumeric checks if a string contains only digits
func IsNumeric(s string) bool {
	matched, _ := regexp.MatchString(`^\d+$`, s)
	return matched
}

// ContainsAsterisk checks if a string contains asterisks (redacted check)
func ContainsAsterisk(s string) bool {
	return strings.Contains(s, "*")
}

// NormalizeCustomerType normalizes customer type to "Individual" or "SME"
func NormalizeCustomerType(customerType string) string {
	normalized := strings.ToLower(strings.TrimSpace(customerType))
	switch normalized {
	case "individual":
		return "Individual"
	case "sme":
		return "SME"
	default:
		return customerType // Return as-is if not recognized
	}
}

// ExtractNumericCode extracts numeric code from string (e.g., "1 - Youth Banking" -> "1")
func ExtractNumericCode(code string) string {
	code = strings.TrimSpace(code)
	if idx := strings.Index(code, "-"); idx != -1 {
		return strings.TrimSpace(code[:idx])
	}
	return code
}



