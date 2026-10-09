package helper

import (
    "encoding/csv"
    "strconv"
    "strings"
    "github.com/gofiber/fiber/v2"
    "api-students/app/model"
)

const (FormatJSON = fiber.MIMEApplicationJSON; FormatCSV = "text/csv")

func Negotiate(c *fiber.Ctx, offered ...string) (string,error) {
    accept := strings.TrimSpace(c.Get(fiber.HeaderAccept))
    if accept == "" || accept == "*/*" { return offered[0],nil }
    chosen := c.Accepts(offered...)
    if chosen == "" { return "",NotAcceptable("format yang diminta tidak tersedia, pilih salah satu dari: "+strings.Join(offered,", ")) }
    return chosen,nil
}

func WriteStudentsCSV(c *fiber.Ctx, students []model.Student) error {
    var b strings.Builder
    w := csv.NewWriter(&b)
    if err := w.Write([]string{"id","nim","name","grade","is_active","created_at"}); err != nil { return Internal(err) }
    for _, s := range students {
        if err := w.Write([]string{strconv.Itoa(s.ID),s.NIM,s.Name,strconv.FormatFloat(s.Grade,'f',2,64),strconv.FormatBool(s.IsActive),s.CreatedAt.UTC().Format("2006-01-02T15:04:05Z")}); err != nil { return Internal(err) }
    }
    w.Flush()
    if err := w.Error(); err != nil { return Internal(err) }
    c.Set(fiber.HeaderContentType,FormatCSV+"; charset=utf-8")
    c.Set(fiber.HeaderContentDisposition,`attachment; filename="students.csv"`)
    return c.SendString(b.String())
}
