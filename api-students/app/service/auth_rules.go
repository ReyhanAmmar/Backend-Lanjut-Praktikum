package service

import (
    "strings"
    "api-students/app/model"
)

func ValidateLogin(req model.LoginRequest) map[string]string {
    errs := map[string]string{}
    if strings.TrimSpace(req.NIM) == "" { errs["nim"] = "wajib diisi" }
    if req.Password == "" { errs["password"] = "wajib diisi" }
    return errs
}
