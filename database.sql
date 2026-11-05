
-- 1. Skema Tabel Users
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'user' NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 2. Skema Tabel Layanan Sampah (Waste Services)
CREATE TABLE IF NOT EXISTS waste_services (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    price BIGINT NOT NULL,
    is_active BOOLEAN DEFAULT TRUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 3. Skema Tabel Tagihan Iuran Warga (Dues)
CREATE TABLE IF NOT EXISTS dues (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    period VARCHAR(7) NOT NULL,
    waste_fee BIGINT NOT NULL,
    security_fee BIGINT NOT NULL,
    rt_fund BIGINT NOT NULL,
    total BIGINT NOT NULL,
    status VARCHAR(20) DEFAULT 'unpaid' NOT NULL,
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT idx_user_period UNIQUE (user_id, period)
);

-- 4. Skema Tabel Transaksi & Pembayaran (Transactions)
CREATE TABLE IF NOT EXISTS transactions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    order_id VARCHAR(100) UNIQUE NOT NULL,
    type VARCHAR(50) NOT NULL, -- 'dues' atau 'waste_service'
    dues_id BIGINT REFERENCES dues(id) ON UPDATE CASCADE ON DELETE SET NULL,
    waste_service_id BIGINT REFERENCES waste_services(id) ON UPDATE CASCADE ON DELETE SET NULL,
    amount BIGINT NOT NULL,
    status VARCHAR(30) DEFAULT 'pending' NOT NULL, -- 'pending', 'paid', 'failed'
    payment_type VARCHAR(50),
    snap_token VARCHAR(255),
    snap_redirect_url TEXT,
    notes TEXT,
    pickup_address TEXT,
    pickup_date TIMESTAMPTZ,
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Indexing untuk optimasi query pencarian
CREATE INDEX IF NOT EXISTS idx_dues_user_id ON dues(user_id);
CREATE INDEX IF NOT EXISTS idx_dues_status ON dues(status);
CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status);
CREATE INDEX IF NOT EXISTS idx_transactions_order_id ON transactions(order_id);

-- ========================================================
-- SEED DATA AWAL (Akun Default, Layanan, Iuran, Transaksi)
-- Password Default:
-- Admin: admin123
-- Warga: warga123
-- ========================================================

-- Insert Users (Password bcrypt: admin123 & warga123)
INSERT INTO users (id, name, email, password, role) VALUES
(1, 'Pengurus RT 05', 'admin@wargapay.id', '$2a$10$wK1y1sP8PZpGZkJqI4qgxeI7mD3mGsmNf0a5vjI7oK3I7eE9Z7aFe', 'admin'),
(2, 'Budi Santoso', 'budi@wargapay.id', '$2a$10$aE6xR4U1A4t3F.WfS5wB6.s1D2H3J4K5L6M7N8O9P0Q1R2S3T4U5V', 'user'),
(3, 'Siti Rahmawati', 'siti@wargapay.id', '$2a$10$aE6xR4U1A4t3F.WfS5wB6.s1D2H3J4K5L6M7N8O9P0Q1R2S3T4U5V', 'user'),
(4, 'Joko Prasetyo', 'joko@wargapay.id', '$2a$10$aE6xR4U1A4t3F.WfS5wB6.s1D2H3J4K5L6M7N8O9P0Q1R2S3T4U5V', 'user')
ON CONFLICT (email) DO NOTHING;

-- Insert Layanan Sampah
INSERT INTO waste_services (id, name, description, price, is_active) VALUES
(1, 'Penjemputan Sampah Rumah Tangga Reguler', 'Pengambilan rutin sampah organik dan anorganik 3x seminggu langsung dari depan rumah.', 35000, true),
(2, 'Penjemputan Sampah Daur Ulang & Terpilah', 'Khusus sampah botol plastik, kardus, dan logam yang sudah dipilah rapi.', 20000, true),
(3, 'Penjemputan Puing & Perabot Bekas (Bulky Waste)', 'Layanan penjemputan barang besar seperti sofa, kasur, lemari, atau sisa renovasi.', 75000, true),
(4, 'Penjemputan Sampah Taman & Ranting', 'Pengangkutan dahan pohon, daun kering, dan limbah kebun cluster.', 30000, true)
ON CONFLICT (id) DO NOTHING;

-- Insert Tagihan Iuran
INSERT INTO dues (id, user_id, period, waste_fee, security_fee, rt_fund, total, status, paid_at) VALUES
(1, 2, '2026-10', 35000, 30000, 20000, 85000, 'paid', NOW() - INTERVAL '15 days'),
(2, 2, '2026-11', 35000, 30000, 20000, 85000, 'unpaid', NULL),
(3, 3, '2026-11', 35000, 30000, 20000, 85000, 'paid', NOW() - INTERVAL '2 days'),
(4, 4, '2026-11', 35000, 30000, 20000, 85000, 'unpaid', NULL)
ON CONFLICT (user_id, period) DO NOTHING;

-- Reset sequence id
SELECT setval('users_id_seq', (SELECT MAX(id) FROM users));
SELECT setval('waste_services_id_seq', (SELECT MAX(id) FROM waste_services));
SELECT setval('dues_id_seq', (SELECT MAX(id) FROM dues));

-- 5. Skema Tabel Pengeluaran Kas RT (Expenses)
CREATE TABLE IF NOT EXISTS expenses (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(150) NOT NULL,
    category VARCHAR(50) NOT NULL,
    amount BIGINT NOT NULL,
    vendor VARCHAR(100) NOT NULL,
    date VARCHAR(30) NOT NULL,
    description TEXT,
    receipt_ref VARCHAR(50),
    approved_by VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_expenses_category ON expenses(category);

-- Insert Pengeluaran Kas RT Awal
INSERT INTO expenses (id, title, category, amount, vendor, date, description, receipt_ref, approved_by) VALUES
(1, 'Honor & Operasional Petugas Keamanan Cluster', 'Keamanan', 4500000, 'Regu Satpam Pos Utama', '01 Nov 2026', 'Gaji bulanan 4 petugas satpam dan biaya ronda malam', 'KW-2026-001', 'Bpk. Hendra (Ketua RT)'),
(2, 'Retribusi Truk Pengangkutan Sampah DLH', 'Kebersihan', 2800000, 'Dinas Lingkungan Hidup / Vendor Mitra', '02 Nov 2026', 'Operasional truk pengangkut sampah 3x seminggu', 'KW-2026-002', 'Ibu Dewi (Bendahara)'),
(3, 'Peremajaan & Penggantian Lampu PJU Blok B', 'Infrastruktur', 1250000, 'Toko Listrik Terang Abadi', '03 Nov 2026', 'Penggantian 8 titik lampu LED PJU 50 Watt', 'KW-2026-003', 'Bpk. Agus (Seksi Pembangunan)'),
(4, 'Pengadaan Refill APAR & Kotak P3K Pos RT', 'Operasional', 650000, 'CV Citra Safety Mandiri', '04 Nov 2026', 'Inspeksi berkala dan refill 3 tabung pemadam pos satpam', 'KW-2026-004', 'Bpk. Hendra (Ketua RT)')
ON CONFLICT (id) DO NOTHING;

SELECT setval('expenses_id_seq', (SELECT MAX(id) FROM expenses));


