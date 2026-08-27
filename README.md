# SPOS-E (Smart Point of Sale - Engine) F&B Edition

<img src="./spose-logo.webp" width="300px" alt="SPOS-E Logo">

**SPOS-E F&B Edition** adalah sistem Kasir POS (Point of Sale) modern serbaguna yang dirancang khusus untuk industri Food & Beverage (F&B) seperti Restoran, Kafe, Kedai Kopi, dan Usaha Kuliner. Aplikasi ini mengintegrasikan pemrosesan transaksi kasir instan, manajemen stok bahan baku otomatis (Bill of Materials / BOM), integrasi pembayaran QRIS dinamis via Midtrans, tampilan dapur real-time (KDS & WLD), serta pengolahan laporan finansial terpadu dalam satu aplikasi *Single Binary Standalone*.

---

## Fitur Utama Sistem

- **Point of Sale (POS) Kasir Touchscreen**:
  - **Mode Layanan**: *Dine In* (Makan di Tempat), *Take Away* (Bawa Pulang), dan *Delivery* (Pesan Antar).
  - **Denah Meja Interaktif (*Table Grid*)**: Status visual real-time (*Available*, *Occupied*, *Reserved*).
  - **Operasi Tagihan Fleksibel**: **Pisah Tagihan (*Split Bill*)** dan **Gabung Tagihan (*Merge Bill*)**.
  - **Kalkulasi Finansial Dinamis**: Pajak Resto (PB1 %), Service Charge (%), dan toggle *Harga Nett (Termasuk Pajak)*.

- **Stok Bahan Baku & Formula BOM (Bill of Materials)**:
  - Pemotongan stok bahan baku resep secara otomatis dan presisi saat transaksi dinyatakan Lunas (`PAID`).
  - Pencegahan otomatis penambahan menu jika stok salah satu bahan baku penyusunnya telah habis.
  - Peringatan stok minim (*Low Stock Alert*) dan log pergerakan stok inventaris.

- **Integrasi Pembayaran Midtrans Core API (Dynamic QRIS)**:
  - Penerbitan kode QRIS dinamis resmi format EMVCo via Gateway Midtrans Core API (`/v2/charge`).
  - **Dukungan Mode Sandbox & Production**: Checklist mode Live / Production langsung pada Setup Wizard & Pengaturan.
  - **Auto-Detect Status Credential**: Deteksi otomatis kredensial yang belum dikonfigurasi dengan kartu peringatan ramah pengguna (*"Midtrans Tidak Diatur"*).
  - Pemantauan status otomatis (*Auto-Status Polling*) 3 detik dan penerima notifikasi webhook asynchronous (*SHA-512 Signature Key*).
  - Dukungan penuh pelunasan Transaksi Tunai (*Cash Engine*) dengan kalkulasi uang kembalian instan.

- **Real-time Display System (KDS & WLD via Server-Sent Events)**:
  - **Kitchen Display System (`/kds`)**: Tampilan khusus tim dapur untuk memperbarui status pembuatan menu (*Queued* → *Cooking* → *Ready*).
  - **Waitingline Display (`/wld`)**: Tampilan khusus pramusaji untuk memantau makanan yang siap disajikan ke meja pelanggan.

- **Menu Pengaturan Sistem & Kredensial (`/admin/settings`)**:
  - Halaman terpusat khusus **Superadmin** untuk mengedit Identitas Outlet, Biaya Layanan, Kredensial Midtrans, serta Feature Flags modul aktif.
  - **Otorisasi Supervisor PIN 6-Digit**: Setiap pengubahan dan penyimpanan pengaturan wajib diverifikasi dengan PIN Supervisor.

- **Keamanan & Manajemen Akses Pengguna**:
  - Otentikasi berbasis **JWT Bearer Token** dan **Role-Based Access Control (RBAC)** (*Superadmin*, *Manager*, *Cashier*, *Kitchen*, *Waiter*).
  - Otorisasi pembatalan pesanan (**VOID Order**) yang dilindungi **6-Digit PIN Supervisor** berbasis enkripsi BCrypt.
  - Proteksi **SetupGuard Middleware** dengan *Pre-warmed Cache* yang mengunci rute inisialisasi sistem secara permanen pasca-setup demi mencegah pengambilalihan akun.

- **Laporan Finansial & Analisis Penjualan**:
  - Ringkasan Omset Harian (*Gross Revenue*), Total Modal (*COGS*), dan Laba Kotor (*Gross Profit*).
  - Analisis produk terlaris (*Top Selling Items*) dan perbandingan metode pembayaran.
  - Fitur **Ekspor Laporan ke Berkas CSV** (BOM UTF-8) siap olah.

- **GUI Double-Click Auto-Terminal Launcher**:
  - Saat file biner diklik (*double-click*) dari GUI File Manager tanpa CLI, aplikasi secara otomatis **memicu peluncuran jendela Terminal/CLI** baru yang menampilkan banner status server aktif, URL `http://localhost:8871`, status database, dan log transaksi real-time.

---

## Arsitektur & Teknologi

- **Backend**: Go (Golang) versi 1.21+
  - Web Framework: **Fiber v2** (`github.com/gofiber/fiber/v2`)
  - ORM Engine: **GORM** (`gorm.io/gorm`)
  - Autentikasi: JWT (`golang-jwt/jwt/v5`) & BCrypt
  - Real-time Engine: **Server-Sent Events (SSE)**
- **Frontend**: SvelteKit / Svelte 4
  - Styling: **Vanilla CSS / Modern Glassmorphism System**
  - State Management: Svelte Stores
  - Icons: **Lucide Svelte**
- **Dukungan Engine Database**:
  - **SQLite 3** (`gorm.io/driver/sqlite` / `glebarez/sqlite` - Database file lokal ringan tanpa dependensi external)
  - **PostgreSQL** (`gorm.io/driver/postgres` - Dukungan Supabase & Direct PostgreSQL IPv4/IPv6)

---

## Persyaratan Sistem (Prerequisites)

Untuk menjalankan atau melakukan build dari source code:

- **Go**: v1.21 atau lebih baru ([Unduh Go](https://golang.org/dl/))
- **Node.js**: v18.0.0 atau lebih baru & **npm** ([Unduh Node.js](https://nodejs.org/))

---

## Panduan Jalankan & Build Aplikasi

### Cara 1: Menjalankan Executable Binary Langsung (Rekomendasi Produksi)

File biner siap pakai untuk berbagai platform tersedia di folder `backend/bin/`:

- **Linux AMD64**: `./backend/bin/spos-e-fnb-linux-amd64`
- **Linux ARM64**: `./backend/bin/spos-e-fnb-linux-arm64`
- **Windows AMD64**: `backend\bin\spos-e-fnb-windows-amd64.exe`
- **macOS Apple Silicon**: `./backend/bin/spos-e-fnb-darwin-arm64`
- **macOS Intel**: `./backend/bin/spos-e-fnb-darwin-amd64`

> **Catatan**: Cukup klik 2x (*double-click*) file biner tersebut di File Manager Anda. Jendela Terminal/CLI akan otomatis terbuka menampilkan status server `http://localhost:8871`. Buka browser Anda dan akses alamat tersebut.

---

### Cara 2: Menjalankan Mode Development (Source Code)

#### 1. Menjalankan Backend (Go Server)
```bash
cd backend
go run cmd/api/main.go
```
*Server backend akan berjalan di `http://localhost:8871`.*

#### 2. Menjalankan Frontend (SvelteKit Dev Server)
```bash
cd frontend
npm install
npm run dev
```
*Frontend dev server akan berjalan di `http://localhost:5173`.*

---

### Cara 3: Build Single Binary dari Source Code

Untuk mengompilasi frontend SPA ke dalam file biner Go tunggal:

1. **Build Frontend SPA**:
   ```bash
   cd frontend
   npm run build
   ```
   *(Hasil build otomatis diekspor ke `backend/cmd/api/public`)*.

2. **Compile Executable Binary Go**:
   ```bash
   cd ../backend
   go build -ldflags="-s -w" -o spos-engine ./cmd/api/main.go
   ```

3. **Jalankan Binary**:
   ```bash
   ./spos-engine
   ```
   Buka browser di `http://localhost:8871`.

---

## Alur Penggunaan Aplikasi

### 1. Setup Wizard (Inisialisasi Pertama Kali)
Saat aplikasi dibuka pertama kali di browser, rute otomatis diarahkan ke `/setup`:
- **Langkah 1**: Pilih Database Engine (**SQLite 3** atau **PostgreSQL**).
- **Langkah 2**: Isi Nama Outlet, Tagline, Alamat, dan Nomor Telepon.
- **Langkah 3**: Pilih Tipe Bisnis (Resto, Cafe, Cloud Kitchen) & Mode Layanan.
- **Langkah 4**: Aktifkan Modul Fitur yang dibutuhkan (Meja, KDS, WLD, BOM Stock, Midtrans).
- **Langkah 5**: Atur Pajak PB1 (%), Service Charge (%), serta Kredensial Midtrans & Checklist Mode Production.
- **Langkah 6**: Buat Akun Superadmin & **6-Digit PIN Supervisor**.
- **Langkah 7**: Klik **Inisialisasi Sistem SPOS**. Sistem mengunci rute setup secara permanen.

### 2. Membuka Shift Kasir (*Opening Shift*)
- Login sebagai Superadmin atau Kasir.
- Pada halaman `/pos`, masukkan nominal **Kas Awal (*Opening Cash*)** di laci uang lalu klik **Buka Shift Register**.

### 3. Operasi Kasir POS (`/pos`)
- **Pilih Layanan & Meja**: Pilih *Dine In*, *Take Away*, atau *Delivery*. Jika *Dine In*, pilih Meja dari Denah Meja.
- **Proses Pembayaran**: Klik **Proses Bayar (F4)**:
  - **Tunai**: Masukkan nominal uang tunai, sistem menghitung uang kembalian secara presisi.
  - **QRIS Midtrans**: Kode QRIS Dinamis ditampilkan. Sistem mendeteksi status pembayaran secara otomatis dalam 3 detik.

### 4. Pelacakan Dapur & Waiter (KDS & WLD)
- **KDS (`/kds`)**: Dapur melihat item masuk, mengklik *Mulai Masak* → *Siap Disajikan*.
- **WLD (`/wld`)**: Pramusaji melihat pesanan matang lalu mengklik *Telah Disajikan*.

### 5. Pelunasan Pending & VOID Order (`/admin/orders`)
- **Pelunasan Transaksi Pending**: Lanjutkan transaksi pending langsung dari tabel riwayat order.
- **Batal Order (VOID)**: Klik *Batal*, masukkan **6-Digit PIN Supervisor** & Alasan VOID. Stok bahan baku BOM dan status meja akan dipulihkan otomatis.

### 6. Pengaturan Sistem (`/admin/settings`)
- Khusus akun ber-role **Superadmin**.
- Mengubah profil resto, persentase pajak/service fee, kredensial Midtrans, serta modul aktif kapan saja dengan otorisasi **PIN Supervisor**.

---

## Kebijakan Lisensi (License Agreement)

Software **SPOS-E (F&B Edition)** dilindungi oleh Undang-Undang Nomor 28 Tahun 2014 tentang Hak Cipta Republik Indonesia serta Perjanjian Lisensi Perangkat Lunak (*Software License Agreement*).

**Pemilik Hak Cipta**: Dillmux / The Lastofem / Berhembusnyaudara

### Skema Paket Lisensi:
1. **Paket Standard (Lisensi Binary / Installer Only / Hosted App)**:
   - Diberikan lisensi non-eksklusif untuk menginstal dan menjalankan file biner/installer pada perangkat sesuai kuota pembelian.
   - **Dilarang keras** melakukan pembongkaran kode (*reverse engineering*), dekompilasi, atau memodifikasi file biner.
2. **Paket Developer (Lisensi Source Code)**:
   - Diberikan akses ke kode sumber untuk penyesuaian kebutuhan internal yang disetujui.
   - **Dilarang keras** mempublikasikan atau mengunggah kode sumber ini ke platform publik (seperti GitHub, GitLab, dsb) tanpa izin tertulis dari Pemilik Hak Cipta.
