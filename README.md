# Interior Backend API

Backend server untuk panel admin sistem manajemen proyek interior.
Dibangun menggunakan Arsitektur Modular Feature-Based (Golang + Gin + GORM + PostgreSQL + Redis).

## Fitur Utama

*   **RESTful API** dengan Gin Web Framework.
*   **RBAC Authentication** dengan JWT & Go-Redis (Token Blacklist).
*   **Database ORM** menggunakan GORM dengan PostgreSQL.
*   **Caching & Session Management** menggunakan Redis (Cache-Aside pattern).
*   **Rate Limiting Middleware** untuk proteksi endpoint login.
*   **Dynamic Data Seeder** untuk Role & Permissions saat startup.
*   **Modular Architecture**: terbagi menjadi Layer Handler, Service, Repository, DTO, dan Entity.

## Prasyarat

*   [Go](https://golang.org/dl/) 1.25.x
*   [PostgreSQL](https://www.postgresql.org/download/)
*   [Redis](https://redis.io/download)

## Instalasi & Menjalankan Service

1.  **Clone / Download Repository**

2.  **Konfigurasi Environment**
    Salin file `.env.example` menjadi `.env` lalu sesuaikan isinya:
    ```bash
    cp .env.example .env
    ```
    Sesuaikan `DB_HOST`, `DB_USER`, `DB_PASS`, `DB_NAME`, dan kredensial Redis.

3.  **Install Dependensi**
    ```bash
    go mod tidy
    ```

4.  **Menjalankan Aplikasi**
    Backend otomatis akan mengeksekusi *auto-migration* dan melakukan *seeding* permissions ketika berjalan.
    ```bash
    go run cmd/api/main.go
    ```

## Arsitektur Layer

Setiap modul / entitas dibagi ke dalam layer berikut:
*   `internal/entity/`: Representasi model GORM dengan tabel di database.
*   `internal/dto/`: Request/Response contract (data transfer object).
*   `internal/repository/`: Interaksi langsung dengan database (PostgreSQL).
*   `internal/service/`: Logika bisnis dan interaksi dengan Cache (Redis).
*   `internal/handler/`: Menerima request HTTP, memanggil Service, dan mengembalikan response JSON standar.
*   `internal/routes/`: Mendaftarkan endpoint handler dengan middleware yang sesuai.

## Response Standard

Seluruh endpoint mengembalikan JSON dengan struktur berikut:
```json
{
  "success": true,
  "message": "Pesan status",
  "data": { ... },
  "errors": null
}
```

## Teknologi & Package Utama

*   `github.com/gin-gonic/gin`: Web framework.
*   `gorm.io/gorm`: Object Relational Mapping (ORM).
*   `github.com/redis/go-redis/v9`: Redis client wrapper.
*   `github.com/golang-jwt/jwt/v5`: Implementasi JWT token.
*   `go.uber.org/zap`: Logging framework yang cepat.
*   `github.com/shopspring/decimal`: Presisi tinggi untuk data harga.
