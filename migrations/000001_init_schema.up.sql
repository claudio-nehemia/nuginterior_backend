-- 1. DIVISI
CREATE TABLE IF NOT EXISTS divisis (
    id          BIGSERIAL PRIMARY KEY,
    nama_divisi VARCHAR(255) NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

-- 2. ROLES
CREATE TABLE IF NOT EXISTS roles (
    id         BIGSERIAL PRIMARY KEY,
    nama_role  VARCHAR(255) NOT NULL UNIQUE,
    divisi_id  BIGINT NOT NULL REFERENCES divisis(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 3. PERMISSIONS
CREATE TABLE IF NOT EXISTS permissions (
    id           BIGSERIAL PRIMARY KEY,
    name         VARCHAR(255) NOT NULL UNIQUE,
    display_name VARCHAR(255),
    "group"      VARCHAR(255),
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    updated_at   TIMESTAMPTZ DEFAULT NOW()
);

-- 4. ROLE_PERMISSION (pivot)
CREATE TABLE IF NOT EXISTS role_permission (
    id            BIGSERIAL PRIMARY KEY,
    role_id       BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id BIGINT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    updated_at    TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(role_id, permission_id)
);

-- 5. USERS
CREATE TABLE IF NOT EXISTS users (
    id                   BIGSERIAL PRIMARY KEY,
    name                 VARCHAR(255) NOT NULL,
    email                VARCHAR(255) NOT NULL UNIQUE,
    email_verified_at    TIMESTAMPTZ,
    password             VARCHAR(255) NOT NULL,
    role_id              BIGINT REFERENCES roles(id) ON DELETE SET NULL,
    fcm_token            TEXT,
    device_platform      VARCHAR(50),
    fcm_token_updated_at TIMESTAMPTZ,
    created_at           TIMESTAMPTZ DEFAULT NOW(),
    updated_at           TIMESTAMPTZ DEFAULT NOW()
);

-- 6. PRODUKS
CREATE TABLE IF NOT EXISTS produks (
    id          BIGSERIAL PRIMARY KEY,
    nama_produk VARCHAR(255) NOT NULL,
    harga       DECIMAL(18,2) DEFAULT 0,
    harga_jasa  DECIMAL(18,2) DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

-- 7. PRODUK_IMAGES
CREATE TABLE IF NOT EXISTS produk_images (
    id         BIGSERIAL PRIMARY KEY,
    produk_id  BIGINT NOT NULL REFERENCES produks(id) ON DELETE CASCADE,
    image      VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 8. ITEMS
CREATE TABLE IF NOT EXISTS items (
    id         BIGSERIAL PRIMARY KEY,
    nama_item  VARCHAR(255) NOT NULL,
    jenis_item VARCHAR(50) NOT NULL CHECK (jenis_item IN ('finishing_dalam','finishing_luar','aksesoris')),
    harga      DECIMAL(18,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 9. BAHAN_BAKUS
CREATE TABLE IF NOT EXISTS bahan_bakus (
    id               BIGSERIAL PRIMARY KEY,
    nama_bahan_baku  VARCHAR(255) NOT NULL,
    created_at       TIMESTAMPTZ DEFAULT NOW(),
    updated_at       TIMESTAMPTZ DEFAULT NOW()
);

-- 10. PRODUK_BAHAN_BAKUS (pivot)
CREATE TABLE IF NOT EXISTS produk_bahan_bakus (
    id            BIGSERIAL PRIMARY KEY,
    produk_id     BIGINT NOT NULL REFERENCES produks(id) ON DELETE CASCADE,
    bahan_baku_id BIGINT NOT NULL REFERENCES bahan_bakus(id) ON DELETE CASCADE,
    harga_dasar   DECIMAL(18,2) DEFAULT 0,
    harga_jasa    DECIMAL(18,2) DEFAULT 0,
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    updated_at    TIMESTAMPTZ DEFAULT NOW()
);

-- 11. JENIS_PENGUKURAN
CREATE TABLE IF NOT EXISTS jenis_pengukuran (
    id               BIGSERIAL PRIMARY KEY,
    nama_pengukuran  VARCHAR(255) NOT NULL,
    created_at       TIMESTAMPTZ DEFAULT NOW(),
    updated_at       TIMESTAMPTZ DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_jenis_pengukuran_deleted_at ON jenis_pengukuran(deleted_at);

-- 12. TERMINS
CREATE TABLE IF NOT EXISTS termins (
    id         BIGSERIAL PRIMARY KEY,
    kode_tipe  VARCHAR(255) NOT NULL,
    nama_tipe  VARCHAR(255) NOT NULL,
    deskripsi  TEXT,
    tahapan    JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
