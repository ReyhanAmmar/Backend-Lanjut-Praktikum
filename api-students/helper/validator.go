package helper
 
import (
    "errors"
    "reflect"
    "strings"
    "unicode"
 
    "github.com/go-playground/validator/v10"
)

var validate = newValidator()

func newValidator() *validator.Validate {
	v := validator.New()

	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return field.Name
		}
		return name
	})

	_ = v.RegisterValidation("nim", func(fl validator.FieldLevel) bool {
		nim := fl.Field().String()
		if len(nim) < 6 || len(nim) > 20 {
			return false
		}
		for _, r := range nim {
			if !unicode.IsDigit(r) {
				return false
			}
		}
		return true
	})

	_ = v.RegisterValidation("studentname", func(fl validator.FieldLevel) bool {
		return strings.TrimSpace(fl.Field().String()) != ""
	})

	_ = v.RegisterValidation("strongpassword", func(fl validator.FieldLevel) bool {
		return PasswordStrength(fl.Field().String()) == ""
	})

	return v
}

func ValidateStruct(s any) map[string]string {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var invalid *validator.InvalidValidationError
	if errors.As(err, &invalid) {
		return map[string]string{"_": "objek yang divalidasi tidak sah"}
	}

	var fieldErrors validator.ValidationErrors
	if !errors.As(err, &fieldErrors) {
		return map[string]string{"_": "validasi gagal"}
	}

	result := make(map[string]string, len(fieldErrors))
	for _, fe := range fieldErrors {
		if _, exists := result[fe.Field()]; !exists {
			result[fe.Field()] = messageFor(fe)
		}
	}
	return result
}

func messageFor(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "wajib diisi"
	case "nim":
		return "NIM harus terdiri dari 6 sampai 20 digit"
	case "studentname":
		return "nama tidak boleh hanya berisi spasi"
	case "min":
		return "minimal " + fe.Param() + " karakter"
	case "max":
		return "maksimal " + fe.Param() + " karakter"
	case "gte":
		return "nilai minimal " + fe.Param()
	case "lte":
		return "nilai maksimal " + fe.Param()
	case "strongpassword":
		if value, ok := fe.Value().(string); ok {
			return PasswordStrength(value)
		}
		return "password tidak memenuhi syarat"
	default:
		return "tidak memenuhi aturan " + fe.Tag()
	}
}

func PasswordStrength(password string) string {
	if len(password) < 8 {
		return "minimal 8 karakter"
	}
	if len(password) > 72 {
		return "maksimal 72 byte"
	}

	var hasLetter, hasDigit bool
	for _, r := range password {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return "harus memuat huruf dan angka"
	}

	weak := map[string]bool{
		"password1":   true,
		"12345678":    true,
		"qwerty123":   true,
		"admin123":    true,
		"password123": true,
	}
	if weak[strings.ToLower(password)] {
		return "password terlalu umum"
	}
	return ""
}