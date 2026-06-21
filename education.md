# Panduan Arsitektur & Layer Backend (Education)

Dokumen ini ditujukan sebagai bahan edukasi untuk memahami arsitektur, fungsi dari masing-masing direktori, dan bagaimana aliran data terjadi di dalam aplikasi backend *Nuginterior*.

Aplikasi ini menggunakan pendekatan **Modular Feature-Based** dan **Layered Architecture** dengan bahasa pemrograman Golang.

---

## 1. Aliran Data (Data Flow)

Setiap request dari user/client akan melewati urutan layer berikut:

`Client Request` ➔ `Routes` ➔ `Middleware` ➔ `Handler` ➔ `Service` ➔ `Repository` ➔ `Database/Cache`

1. **Routes**: Menerima request HTTP dari klien dan mengarahkannya ke Handler yang tepat.
2. **Middleware**: Berfungsi sebagai penjaga pintu gerbang (contoh: mengecek token JWT valid atau tidak, rate limiting, validasi akses/permissions).
3. **Handler**: Bertanggung jawab menerima input (JSON body, query, params), memvalidasinya menggunakan *DTO*, dan mengirim balasan (response JSON) kembali ke klien.
4. **Service**: Inti dari sistem (Business Logic). Disini semua perhitungan, validasi aturan bisnis, dan akses *cache* (Redis) terjadi.
5. **Repository**: Layer khusus yang bertugas "berbicara" secara langsung ke database (PostgreSQL) menggunakan GORM. Layer ini hanya memikirkan urusan CRUD.
6. **Database (Entity)**: Struktur model data kita dalam bentuk database yang merepresentasikan struktur tabel asli di PostgreSQL.

---

## 2. Fungsi Setiap Direktori (Folder)

Struktur folder dibuat rapi dan modular agar lebih mudah dirawat (maintainable).

### 📁 `cmd/`
Berisi file utama (entry point) aplikasi.
*   **`api/main.go`**: Titik awal ketika aplikasi dijalankan. File ini memanggil fungsi-fungsi *bootstrap* untuk menginisialisasi seluruh layer.

### 📁 `internal/`
Berisi kode private dari aplikasi yang tidak boleh/tidak akan di-import oleh aplikasi Golang lain.

*   **`bootstrap/`**: Bertugas me-wiring atau "merangkai" semua komponen (Database, Config, Service, Repository, Handler) menjadi satu kesatuan (Dependency Injection). Disini juga *seeder* dijalankan.
*   **`config/`**: Membaca variabel *environment* dari file `.env` ke dalam object Struct (`Config`) sehingga lebih aman digunakan di seluruh aplikasi.
*   **`constants/`**: Menyimpan variabel statis, enumerasi, list permisions, atau key statis (misal *Key Redis*).
*   **`database/`**: Menyimpan logika untuk terhubung ke PostgreSQL, auto-migrate GORM, dan auto-create DB (jika belum ada).
*   **`dto/ (Data Transfer Object)`**: Berisi *struct* (bentuk data) untuk menangkap input JSON dari *Request* (contoh: `CreateUserRequest`) dan *struct* untuk format balik *Response* (contoh: `UserResponse`).
*   **`entity/`**: Berisi *struct* model GORM yang mewakili tabel asli (fisik) pada database PostgreSQL.
*   **`handler/`**: Berisi sekumpulan fungsi untuk menangani request dari `routes`, memvalidasi dari `dto`, memanggil `service`, dan mengirim hasil fungsi balik ke client.
*   **`helper/`**: Fungsi utilitas ringan. Contoh: Format Respon JSON yang seragam (`OK`, `BadRequest`), Enkripsi Password (Bcrypt), pembaca param ID di URL.
*   **`middleware/`**: Logika pencegatan rute/URL sebelum sampai ke `handler`. Meliputi pengecekan validitas token JWT, pengecekan *Permissions (RBAC)*, CORS, dan *Rate Limiter*.
*   **`repository/`**: Layer untuk melakukan "query" (CREATE, READ, UPDATE, DELETE) ke database GORM. *Tidak* mengandung validasi bisnis.
*   **`routes/`**: Tempat mendaftarkan semua alamat URL/Endpoint (contoh: `POST /api/auth/login`).
*   **`service/`**: Tempat menaruh seluruh *Business Logic*. Menyediakan fungsi-fungsi khusus untuk handler, yang di dalamnya dapat menyimpan ke Cache Redis, menghitung diskon harga, dsb.

### 📁 `pkg/`
Berisi kode publik/utilitas khusus yang bisa di-*reuse* (digunakan ulang) yang tidak terkait langsung dengan logika *interior_backend*.
*   **`cache/`**: Wrapper (pembungkus) *go-redis* untuk mempermudah akses Get/Set JSON data dari Redis.
*   **`logger/`**: Konfigurasi logger tingkat lanjut dengan menggunakan *Uber Zap* untuk menghasilkan log terstruktur.

### 📁 `migrations/`
Berisi file `.sql` biasa untuk keperluan sejarah perubahan *database scheme*. Sangat berguna apabila ingin menjalankan secara manual maupun untuk acuan migrasi ke depan.

---

## 3. Cara Kerja Auto-Create Database

Aplikasi juga memiliki fungsi *Auto-Create Database* saat dinyalakan (`main.go`). Prosesnya sbb:
1. Mengambil kredensial dari `.env`.
2. Koneksi ke *database default* `postgres`.
3. Menjalankan fungsi pengecekan tabel: `SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = 'interior_db')`
4. Bila *false* (belum ada), maka menjalankan perintah `CREATE DATABASE interior_db`.
5. Lalu, menutup koneksi *default* dan kembali terkoneksi langsung ke database `interior_db`.
6. Semua fungsi GORM `AutoMigrate` (membuat tabel) berjalan secara otomatis di database yang baru dibuat.
