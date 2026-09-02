# api-students

REST API data mahasiswa dengan Fiber v2 dan PostgreSQL, disusun memakai pola repository. Proyek ini adalah lanjutan dari tugas pertemuan 1 (struct Student) dan pertemuan 2 (REST API di memori), sekarang datanya tersimpan permanen di PostgreSQL.

## Menyiapkan basis data dari nol

1. Pastikan PostgreSQL sudah terpasang dan menyala.
2. Buat basis data kosong.

   ```
   psql -U postgres -c "CREATE DATABASE students;"
   ```

3. Jalankan berkas migrasi untuk membuat tabel `students`.

   ```
   psql -U postgres -d students -f migrations/001_create_students.sql
   ```

4. Periksa hasilnya.

   ```
   psql -U postgres -d students -c "\d students"
   ```

## Skema tabel students

| Kolom      | Tipe          | Batasan                                   |
|------------|---------------|--------------------------------------------|
| id         | SERIAL        | PRIMARY KEY                                 |
| nim        | VARCHAR(20)   | NOT NULL, unik lewat UNIQUE INDEX           |
| name       | VARCHAR(150)  | NOT NULL                                    |
| grade      | NUMERIC(5,2)  | NOT NULL, DEFAULT 0, CHECK 0–100            |
| is_active  | BOOLEAN       | NOT NULL, DEFAULT TRUE                      |
| created_at | TIMESTAMPTZ   | NOT NULL, DEFAULT NOW()                     |

Indeks yang dibuat:

- `students_nim_key`, UNIQUE INDEX pada kolom `nim`. Menjaga keunikan NIM di level basis data, bukan hanya di kode Go, supaya tidak ada celah balapan ketika dua permintaan datang bersamaan.
- `students_grade_idx`, INDEX biasa pada kolom `grade`. Mempercepat query dengan filter `min_grade` dan `max_grade`, serta query dengan pengurutan berdasarkan grade.

## Variabel environment

Salin `.env.example` menjadi `.env`, lalu isi sesuai lingkungan Anda. Berkas `.env` tidak pernah ikut ter-commit.

| Variabel      | Kegunaan                              | Nilai bawaan         |
|---------------|----------------------------------------|-----------------------|
| APP_PORT      | Port server Fiber                     |                   
|
| DB_HOST       | Host PostgreSQL                       | 
|
| DB_PORT       | Port PostgreSQL                       |                 
|
| DB_USER       | Username PostgreSQL                   |               
|
| DB_PASSWORD   | Password PostgreSQL                   | 
|
| DB_NAME       | Nama database                         |
|
| DB_SSLMODE    | Mode SSL koneksi                      | 
|
| DB_MAX_CONNS  | Jumlah maksimum koneksi dalam pool    | 
|

## Menjalankan proyek

```
go mod tidy
go run .
```

Server berjalan di `http://localhost:3000`. Karena proyek ini terdiri atas beberapa berkas dan paket, gunakan `go run .`, bukan `go run main.go`.

## Kontrak API

Amplop respons yang dipakai konsisten di seluruh endpoint.

Berhasil, satu data:
```json
{ "success": true, "message": "...", "data": { } }
```

Berhasil, daftar data:
```json
{ "success": true, "message": "...", "data": [ ], "meta": { "page": 1, "limit": 10, "total": 0, "total_pages": 0 } }
```

Gagal biasa:
```json
{ "success": false, "message": "student tidak ditemukan" }
```

Gagal validasi:
```json
{ "success": false, "message": "validasi gagal", "errors": { "nim": "wajib diisi" } }
```

| Metode | Endpoint                | Parameter / Body                                                                 | Contoh body permintaan                                                        | Status mungkin      | Contoh respons singkat                                   |
|--------|--------------------------|-----------------------------------------------------------------------------------|--------------------------------------------------------------------------------|----------------------|-------------------------------------------------------------|
| GET    | /api/v1/health           | –                                                                                   | –                                                                                | 200, 503             | `{ "success": true, "message": "server dan database berjalan" }` |
| GET    | /api/v1/students         | query: page, limit, search, sort, order, is_active, min_grade, max_grade          | –                                                                                | 200                   | `{ "success": true, "data": [ ], "meta": { } }`              |
| GET    | /api/v1/students/:id     | path: id                                                                            | –                                                                                | 200, 400, 404         | `{ "success": true, "data": { "id": 1, "nim": "123" } }`     |
| POST   | /api/v1/students         | body: nim, name, grade                                                             | `{"nim":"123456","name":"Sari","grade":85}`                                     | 201, 400, 415, 422, 409 | header `Location: /api/v1/students/1`                    |
| PUT    | /api/v1/students/:id     | path: id · body: nim, name, grade, is_active (semua wajib)                        | `{"nim":"123456","name":"Sari Dewi","grade":90,"is_active":true}`               | 200, 400, 404, 415, 422, 409 | `{ "success": true, "message": "student berhasil diganti seluruhnya" }` |
| PATCH  | /api/v1/students/:id     | path: id · body: sebagian dari nim, name, grade, is_active                        | `{"is_active":false}`                                                           | 200, 400, 404, 415, 422, 409 | `{ "success": true, "message": "student berhasil diperbarui sebagian" }` |
| DELETE | /api/v1/students/:id     | path: id                                                                            | –                                                                                | 204, 400, 404         | tanpa body                                                    |

## Query string pada GET /api/v1/students

| Parameter  | Kegunaan                              | Nilai bawaan aman |
|------------|-----------------------------------------|---------------------|
| page       | Halaman keberapa                       | 1                   |
| limit      | Baris per halaman, batas atas 50       | 10                  |
| search     | Cari pada kolom name, tidak case-sensitive (ILIKE) | kosong |
| sort       | Kolom pengurutan, daftar putih: id, nim, name, grade, created_at | id |
| order      | asc atau desc                          | asc                 |
| is_active  | Saring berdasarkan status aktif        | tidak menyaring     |
| min_grade  | Saring nilai minimum                   | tidak menyaring     |
| max_grade  | Saring nilai maksimum                  | tidak menyaring     |