SELECT outbox_id, student_id, channel, subject, sent_at 
FROM notifications 
WHERE outbox_id = 8
ORDER BY student_id ASC;