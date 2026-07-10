package handler

import (
	"fmt"
	"net/mail"
	"strings"
)

type fieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func validateRequired(val, field string) *fieldError {
	if strings.TrimSpace(val) == "" {
		return &fieldError{Field: field, Message: field + " wajib diisi"}
	}
	return nil
}

func validateEmail(val string) *fieldError {
	if _, err := mail.ParseAddress(val); err != nil {
		return &fieldError{Field: "email", Message: "email tidak valid"}
	}
	return nil
}

func validateMinLen(val string, min int, field string) *fieldError {
	if len(val) < min {
		return &fieldError{Field: field, Message: fmt.Sprintf("%s minimal %d karakter", field, min)}
	}
	return nil
}

func validatePort(port int) *fieldError {
	if port < 1 || port > 65535 {
		return &fieldError{Field: "api_port", Message: "port harus 1-65535"}
	}
	return nil
}

func validateCount(count int) *fieldError {
	if count < 1 || count > 500 {
		return &fieldError{Field: "count", Message: "count harus 1-500"}
	}
	return nil
}
