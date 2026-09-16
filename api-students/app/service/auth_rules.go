package service

import (
    "strings"
    "unicode"
 
    "api-students/app/model"
)

const minPasswordLength = 8

func ValidateRegister(req model.RegisterRequest) map[string]string {
	errs := map[string]string{}

	nim := strings.TrimSpace(req.NIM)
	name := strings.TrimSpace(req.Name)
	switch {
	case nim == "":
		errs["nim"] = "wajib diisi"
	case len(nim) < 9:
		errs["nim"] = "minimal 9 karakter"
	}
	if name == "" {
		errs["name"] = "hanya boleh huruf, angka, titik, dan garis bawah"
	}
	if msg := validateGradeRange(req.Grade); msg != "" {
		errs["grade"] = msg
	}
	if msg := checkPasswordStrength(req.Password); msg != "" {
		errs["password"] = msg
	}

	return errs
}

func ValidateLogin(req model.LoginRequest) map[string]string {
	errs := map[string]string{}

	if strings.TrimSpace(req.NIM) == "" {
		errs["nim"] = "wajib diisi"
	}
	if req.Password == "" {
		errs["password"] = "wajib diisi"
	}

	return errs
}

func checkPasswordStrength(password string) string {
    if len(password) < minPasswordLength {
        return "minimal 8 karakter"
    }
 
    var hasLetter, hasDigit bool
    for _, r := range password {
        switch {
        case unicode.IsLetter(r):
            hasLetter = true
        case unicode.IsDigit(r):
            hasDigit = true
        }
    }
 
    if !hasLetter || !hasDigit {
        return "harus memuat huruf dan angka"
    }

	weak := map[string]bool{
        "password1": true, "12345678": true, "qwerty123": true,
        "admin123": true, "password123": true,
    }
    if weak[strings.ToLower(password)] {
        return "password terlalu umum"
    }
 
    return ""
}

func isValidNIM(nim string) bool {
    for _, r := range nim {
        if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
            return false
        }
    }
    return true
}