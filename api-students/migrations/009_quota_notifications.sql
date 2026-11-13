-- Menghapus constraint lama
ALTER TABLE notifications
DROP CONSTRAINT IF EXISTS notifications_outbox_unique;

-- Satu event dapat menghasilkan banyak notifikasi,
-- tetapi satu mahasiswa hanya menerima satu notifikasi
-- untuk event yang sama.
ALTER TABLE notifications
ADD CONSTRAINT notifications_outbox_student_unique
UNIQUE (outbox_id, student_id);