# Modul 7 - api-students

Status: implementasi disiapkan. Hasil runtime dan EXPLAIN ANALYZE harus diisi dari PostgreSQL lokal; jangan menyalin hasil contoh modul sebagai bukti.

## Menjalankan

Dari direktori `api-students`, jalankan `go mod tidy`, `go build ./...`, dan `go test ./...`. Terapkan migration 001 sampai 006 secara berurutan pada database pengujian. Jalankan API mengikuti README, lalu login dan simpan access token sebagai `TOKEN`.

## Bukti curl (Git Bash)

```bash
BASE=http://localhost:3000/api/v1
curl -i "$BASE/students?limit=5" -H "Authorization: Bearer $TOKEN"
# Salin meta.next_cursor dari response sebagai CURSOR.
curl -i "$BASE/students?limit=5&cursor=$CURSOR" -H "Authorization: Bearer $TOKEN"
curl -i "$BASE/students?limit=3" -H "Authorization: Bearer $TOKEN" -H 'Accept: text/csv'
curl -i "$BASE/students" -H "Authorization: Bearer $TOKEN" -H 'Accept: application/xml'
curl -i "$BASE/students" -H "Authorization: Bearer $TOKEN" -H 'Accept: */*'
curl -i "$BASE/students?cursor=bukanbase64!!" -H "Authorization: Bearer $TOKEN"
curl -i "$BASE/students?is_active=mungkin" -H "Authorization: Bearer $TOKEN"
curl -i -X PATCH "$BASE/students/4" -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{"name":""}'
curl -i -X PATCH "$BASE/students/4" -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{"grade":88}'
curl -i -X PATCH "$BASE/students/4" -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{}'
```

Ganti ID 4 dengan student yang dapat diubah oleh token tersebut. Untuk bukti tanpa duplikasi, simpan halaman 1, tambahkan student baru dengan NIM unik, lalu ambil halaman 2 menggunakan cursor lama. Bandingkan seluruh ID.

## EXPLAIN ANALYZE

```sql
EXPLAIN (ANALYZE, BUFFERS)
SELECT id, nim, name, grade, is_active, created_at
FROM students ORDER BY created_at DESC, id DESC LIMIT 6;

EXPLAIN (ANALYZE, BUFFERS)
SELECT id, nim, name, grade, is_active, created_at
FROM students WHERE (created_at, id) < ('2026-09-13 14:12:40+00', 8)
ORDER BY created_at DESC, id DESC LIMIT 6;
```

Gunakan nilai `created_at` dan `id` hasil halaman 1 sendiri. Pada tabel yang kecil planner dapat memilih sequential scan walau index ada. Untuk memeriksa rencana index, perbanyak data uji lalu `ANALYZE students;` dan ulangi; laporkan rencana yang benar-benar muncul.

## Kode error endpoint students

| Status | Code | Pemicu |
| --- | --- | --- |
| 400 | BAD_REQUEST | ID, cursor, limit, filter, body rusak, atau PATCH kosong |
| 401 | UNAUTHORIZED | Token tidak ada, rusak, atau kedaluwarsa |
| 403 | FORBIDDEN | Hak akses kurang atau mencoba menghapus akun sendiri |
| 404 | NOT_FOUND | Student atau route tidak ditemukan |
| 406 | NOT_ACCEPTABLE | Accept tidak didukung untuk daftar |
| 409 | CONFLICT | NIM sudah dipakai |
| 415 | UNSUPPORTED_MEDIA_TYPE | Body bukan JSON |
| 422 | VALIDATION_ERROR | Aturan tag atau role tidak terpenuhi |
| 429 | TOO_MANY_REQUESTS | Batas login, bukan endpoint students |
| 500 | INTERNAL_ERROR | Kegagalan query atau operasi internal |

## Analisis D.5

Cursor cocok untuk daftar yang digulir bertahap. Bila antarmuka perlu nomor halaman dan jumlah total, saya akan menyediakan endpoint tabel administrasi terpisah berbasis offset dengan COUNT, sambil mempertahankan cursor untuk feed.

Cursor base64 dapat dibaca dan dipalsukan. Jangan simpan token akses, password, atau data rahasia di dalamnya. Client dapat mendekode dan mengambil rahasia tersebut; cursor juga tidak boleh dijadikan bukti otorisasi.

Aturan setidaknya satu field pada PATCH tidak dapat dinyatakan oleh tag per field karena aturan itu membandingkan keberadaan beberapa field sekaligus. Keberadaan role di database dan izin pemilik juga memerlukan pemeriksaan di service/repository.
