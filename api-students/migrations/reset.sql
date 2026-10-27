DELETE FROM enrollments
WHERE student_id IN (SELECT id FROM students WHERE nim LIKE 'mhs%');

UPDATE courses c
SET enrolled_count = (SELECT COUNT(*) FROM enrollments e WHERE e.course_id = c.id);
