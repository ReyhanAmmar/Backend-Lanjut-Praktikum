package helper

import (
    "encoding/base64"
    "errors"
    "strconv"
    "strings"
    "time"
    "github.com/gofiber/fiber/v2"
    "api-students/app/model"
)

var ErrInvalidCursor = errors.New("cursor tidak valid")

func EncodeCursor(createdAt time.Time, id int) string {
    raw := strconv.FormatInt(createdAt.UTC().UnixNano(), 10) + "|" + strconv.Itoa(id)
    return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func DecodeCursor(encoded string) (model.Cursor, error) {
    decoded, err := base64.RawURLEncoding.DecodeString(encoded)
    if err != nil || encoded == "" { return model.Cursor{}, ErrInvalidCursor }
    parts := strings.Split(string(decoded), "|")
    if len(parts) != 2 { return model.Cursor{}, ErrInvalidCursor }
    nanos, err := strconv.ParseInt(parts[0], 10, 64)
    if err != nil || nanos <= 0 { return model.Cursor{}, ErrInvalidCursor }
    id, err := strconv.Atoi(parts[1])
    if err != nil || id < 1 { return model.Cursor{}, ErrInvalidCursor }
    return model.Cursor{CreatedAt: time.Unix(0,nanos).UTC(),ID:id},nil
}

func ParseCursorQuery(c *fiber.Ctx) (model.CursorQuery,error) {
    q := model.CursorQuery{Limit: 10, Search: strings.TrimSpace(c.Query("search"))}
    if raw := c.Query("limit"); raw != "" {
        n, err := strconv.Atoi(raw)
        if err != nil || n < 1 { return q, BadRequest("limit harus angka positif") }
        q.Limit = n
    }
    if q.Limit > 50 { q.Limit = 50 }
    if raw := c.Query("is_active"); raw != "" {
        value, err := strconv.ParseBool(raw)
        if err != nil { return q, BadRequest("is_active harus boolean") }
        q.IsActive = &value
    }
    if raw := c.Query("cursor"); raw != "" {
        value, err := DecodeCursor(raw)
        if err != nil { return q, BadRequest("cursor tidak valid") }
        q.After = &value
    }
    return q,nil
}
