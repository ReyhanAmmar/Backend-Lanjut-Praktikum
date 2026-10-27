EXPLAIN ANALYZE
SELECT id, nim, name, grade, role, is_active, COALESCE(owner_id, 0), created_at
FROM students
ORDER BY created_at DESC, id DESC
LIMIT 4;