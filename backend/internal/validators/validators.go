package validators

import (
	"strings"
)

// ContainsSQLInjection проверяет строку на наличие SQL injection паттернов
func ContainsSQLInjection(input string) bool {
	sqlPatterns := []string{
		"'", "\"", ";", "--", "/*", "*/", "union", "select", "insert",
		"update", "delete", "drop", "create", "alter", "exec", "execute",
		"script", "javascript", "vbscript", "onload", "onerror",
	}

	lowerInput := strings.ToLower(input)
	for _, pattern := range sqlPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}
	return false
}

// ContainsXSS проверяет строку на наличие XSS паттернов
func ContainsXSS(input string) bool {
	xssPatterns := []string{
		"<script", "</script>", "javascript:", "vbscript:", "onload=",
		"onerror=", "onclick=", "onmouseover=", "<iframe", "<object",
		"embed", "<link", "<meta", "eval(", "expression(",
	}

	lowerInput := strings.ToLower(input)
	for _, pattern := range xssPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}
	return false
}
