# Dokumentasi API Mahasiswa (Student API)

Repositori ini berisi implementasi RESTful API untuk pengelolaan data mahasiswa menggunakan bahasa pemrograman Go dan framework **[Fiber](https://gofiber.io/)**.

## Tabel Kontrak API

Berikut adalah tabel kontrak endpoint API lengkap dengan metode, endpoint, parameter, contoh body, status code, dan contoh responsnya:

| Metode | Endpoint | Parameter | Contoh Body Permintaan | Status yang Mungkin Dikembalikan | Contoh Respons |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **GET** | `/health` | *Tidak ada* | *Tidak ada* | `200 OK` | `{"success": true, "message": "service mahasiswa berjalan normal", "data": {"timestamp": "2026-08-27T00:00:00Z"}}` |
| **GET** | `/api/v1/students` | **Query Params (Opsional):**<br>• `page` (int, default: `1`)<br>• `limit` (int, default: `10`)<br>• `search` (string)<br>• `sort` (`id`, `nim`, `name`, `grade`, `created_at`)<br>• `order` (`asc`, `desc`)<br>• `is_active` (`true`, `false`) | *Tidak ada* | `200 OK` | `{"success": true, "message": "daftar mahasiswa berhasil diambil", "data": [{"id": 1, "nim": "434241061", "name": "Muhammad Reyhan Ammar", "grade": 90, "is_active": true, "created_at": "2026-08-27T00:04:32Z"}], "meta": {"page": 1, "limit": 10, "total": 1, "total_pages": 1}}` |
| **GET** | `/api/v1/students/:id` | **Path Param:**<br>• `id` (integer positif, wajib) | *Tidak ada* | • `200 OK`<br>• `400 Bad Request`<br>• `404 Not Found` | `{"success": true, "message": "mahasiswa ditemukan", "data": {"id": 1, "nim": "434241061", "name": "Muhammad Reyhan Ammar", "grade": 90, "is_active": true, "created_at": "2026-08-27T00:04:32Z"}}` |
| **POST** | `/api/v1/students` | **Header (Wajib):**<br>• `Content-Type: application/json` | `{"nim": "434241061", "name": "Muhammad Reyhan Ammar", "grade": 90.0}` | • `201 Created`<br>• `400 Bad Request`<br>• `409 Conflict`<br>• `415 Unsupported Media Type`<br>• `422 Unprocessable Entity` | `{"success": true, "message": "mahasiswa berhasil didaftarkan", "data": {"id": 1, "nim": "434241061", "name": "Muhammad Reyhan Ammar", "grade": 90, "is_active": true, "created_at": "2026-08-27T00:04:32Z"}}` |
| **PUT** | `/api/v1/students/:id` | **Path Param:**<br>• `id` (integer positif, wajib)<br>**Header (Wajib):**<br>• `Content-Type: application/json` | `{"nim": "434241061", "name": "Muhammad Reyhan Ammar", "grade": 95.0, "is_active": true}` | • `200 OK`<br>• `400 Bad Request`<br>• `404 Not Found`<br>• `415 Unsupported Media Type`<br>• `422 Unprocessable Entity` | `{"success": true, "message": "data mahasiswa berhasil diganti seluruhnya", "data": {"id": 1, "nim": "434241061", "name": "Muhammad Reyhan Ammar", "grade": 95, "is_active": true, "created_at": "2026-08-27T00:04:32Z"}}` |
| **PATCH** | `/api/v1/students/:id` | **Path Param:**<br>• `id` (integer positif, wajib)<br>**Header (Wajib):**<br>• `Content-Type: application/json` | `{"grade": 98.0, "is_active": false}` | • `200 OK`<br>• `400 Bad Request`<br>• `404 Not Found`<br>• `415 Unsupported Media Type`<br>• `422 Unprocessable Entity` | `{"success": true, "message": "data mahasiswa berhasil diperbarui sebagian", "data": {"id": 1, "nim": "434241061", "name": "Muhammad Reyhan Ammar", "grade": 98, "is_active": false, "created_at": "2026-08-27T00:04:32Z"}}` |
| **DELETE** | `/api/v1/students/:id` | **Path Param:**<br>• `id` (integer positif, wajib) | *Tidak ada* | • `204 No Content`<br>• `400 Bad Request`<br>• `404 Not Found` | *(Body kosong / No Content)* |


## Keterangan Status Code HTTP

| Status Code | Nama Status | Keterangan |
| :--- | :--- | :--- |
| **`200`** | **OK** | Permintaan berhasil diproses dan data berhasil dikembalikan/diperbarui. |
| **`201`** | **Created** | Data mahasiswa baru berhasil disimpan ke dalam server. |
| **`204`** | **No Content** | Data berhasil dihapus dari server tanpa mengembalikan konten body. |
| **`400`** | **Bad Request** | Parameter ID bukan angka bulat positif atau format JSON rusak/tidak valid. |
| **`404`** | **Not Found** | Endpoint tidak ditemukan atau ID mahasiswa tidak ada di database/memori. |
| **`409`** | **Conflict** | Terjadi konflik duplikasi data unik (NIM sudah digunakan oleh mahasiswa lain). |
| **`415`** | **Unsupported Media Type** | Header `Content-Type` pada request ber-body bukan `application/json`. |
| **`422`** | **Unprocessable Entity** | Format JSON valid namun melanggar aturan validasi (field wajib kosong / nilai di luar rentang `0.00 - 100.00`). |
