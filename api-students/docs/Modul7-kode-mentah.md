# Modul 7 - kode mentah untuk D.1

Branch ini adalah salinan `api-students` yang mengikuti potongan kode Bagian B Modul 7 sebelum sembilan kesalahan sengaja diperbaiki. Nama entity dan field disesuaikan dari user menjadi student/NIM/name/grade. Ini bahan percobaan D.1, bukan kode aplikasi siap pakai.

Lokasi yang sengaja mentah:

1. `helper/errors.go`: `Validation` masih berstatus 400.
2. `config/app.go`: akses `appErr.cause` dari package lain.
3. `config/app.go`: cabang log 4xx/5xx terbalik.
4. `middleware/middleware.go`: access log membaca status respons sebelum ErrorHandler.
5. `app/service/student_service.go`: `translateError` mengembalikan `nil` pada error yang tidak dikenal.
6. `helper/validator.go`: `strongpassword` memakai `!= ""`.
7. `app/model/student.go`: NIM pada PATCH bertipe `string`, sementara `student_rules.go` memperlakukannya sebagai pointer.
8. `app/repository/student_repository.go`: query cursor mengurut `ASC` walau cursor menggunakan `<`.
9. `helper/negotiate.go`: penulis CSV tidak memanggil `Flush`.

Jalankan `go build ./...` dan simpan keluaran compiler asli sebelum melakukan perbaikan. Perbaiki satu kesalahan, uji lagi, lalu lanjutkan. Setelah build berhasil, jalankan pengujian Bagian C pada database pengujian dan catat respons/log yang benar-benar muncul. Jangan memakai contoh keluaran di modul sebagai bukti percobaan.
