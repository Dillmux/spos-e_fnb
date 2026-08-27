# SPOS-E (Smart Point of Sale - Engine) F&B Edition

<img src="./spose-logo.webp" width="300px">

**SPOS-E F&B Edition** adalah sistem Kasir POS (Point of Sale) modern serbaguna yang dirancang khusus untuk industri Food & Beverage (F&B) seperti Restoran, Kafe, Kedai Kopi, dan Usaha Kuliner. Aplikasi ini mengintegrasikan pemrosesan transaksi kasir instan, manajemen stok bahan baku otomatis (Bill of Materials / BOM), integrasi pembayaran QRIS dinamis via Midtrans, serta tampilan dapur real-time dalam satu platform terpadu.

---

## Fitur Utama Sistem

- **Point of Sale (POS) Kasir Touchscreen**:
  - Dukungan mode layanan: _Dine In_ (Makan di Tempat), _Take Away_ (Bawa Pulang), dan _Delivery_ (Pesan Antar).
  - Manajemen Denah Meja Interaktif (_Table Grid_) dengan status _Available_, _Occupied_, dan _Reserved_.
  - Operasi tagihan fleksibel: **Pisah Tagihan (_Split Bill_)** dan **Gabung Tagihan (_Merge Bill_)**.
  - Kalkulasi finansial dinamis: Pajak Resto & Service Charge sesuai profil outlet terdaftar.

- **Stok Bahan Baku & Formula BOM (Bill of Materials)**:
  - Pemotongan stok bahan baku secara otomatis dan presisi saat transaksi dinyatakan Lunas (`PAID`).
  - Pencegahan otomatis penambahan menu jika stok salah satu bahan baku penyusunnya telah habis.
  - Peringatan stok minim (_Low Stock Alert_) dan bahan habis (_Depleted Inventory Log_).

- **Integrasi Pembayaran Midtrans Core API (Dynamic QRIS)**:
  - Penerbitan kode QRIS dinamis resmi format EMVCo via Gateway Midtrans Core API (`/v2/charge`).
  - Pemantauan status otomatis (_Auto-Status Polling_) 3 detik dan penerima notifikasi webhook asynchronous (`SHA-512 Signature Key`).
  - Dukungan penuh pelunasan Transaksi Tunai (_Cash Engine_) dengan kalkulasi uang kembalian instan.

- **Real-time Display System (KDS & WLD)**:
  - **Kitchen Display System (`/kds`)**: Tampilan khusus tim dapur untuk memperbarui status pembuatan menu (_Queued_ → _Cooking_ → _Ready_).
  - **Waitingline Display (`/wld`)**: Tampilan khusus pramusaji untuk memantau makanan yang siap disajikan ke meja pelanggan.

- **Keamanan & Manajemen Akses Pengguna**:
  - Otentikasi berbasis **JWT Bearer Token** dan **Role-Based Access Control (RBAC)** (_Superadmin_, _Manager_, _Cashier_, _Kitchen_, _Waiter_).
  - Otorisasi pembatalan pesanan (**VOID Order**) yang dilindungi **6-Digit PIN Supervisor** berbasis enkripsi BCrypt.
  - Proteksi **SetupGuard Middleware** yang mengunci rute inisialisasi sistem secara permanen pasca-setup demi mencegah pengambilalihan akun.

- **Laporan Finansial & Analisis Penjualan**:
  - Ringkasan Omset Harian (_Gross Revenue_), Total Modal (_COGS_), dan Laba Kotor (_Gross Profit_).
  - Analisis produk terlaris (_Top Selling Items_) dan perbandingan metode pembayaran.
  - Fitur **Ekspor Laporan ke Berkas CSV** (BOM UTF-8) siap olah.

---

## Arsitektur & Teknologi

- **Backend**: Go (Golang) versi 1.21+
  - Web Framework: **Fiber v2**
  - ORM Engine: **GORM**
  - Autentikasi: JWT (`golang-jwt/jwt/v5`) & BCrypt
  - Driver Database: SQLite3 & PostgreSQL (`gorm.io/driver/postgres`)
- **Frontend**: SvelteKit / Svelte 4
  - Styling: **Vanilla CSS / Tailwind CSS**
  - State Management: Svelte Stores
  - Icons: **Lucide Svelte**
- **Database Engine Support**:
  - **SQLite** (Database file lokal ringan untuk pengujian/single-node)
  - **PostgreSQL / Supabase** (Transaction Pooler Port 6543 IPv4 & IPv6 Direct)

---

## Persyaratan Sistem (Prerequisites)

Sebelum menginstal dan menjalankan aplikasi, pastikan komputer Anda telah terpasang:

- **Go**: v1.21 atau lebih baru ([Unduh Go](https://golang.org/dl/))
- **Node.js**: v18.0.0 atau lebih baru & **npm** ([Unduh Node.js](https://nodejs.org/))
- **GCC / C Compiler**: Diperlukan jika menggunakan driver SQLite lokal (biasanya sudah ada pada Linux/macOS, atau tdm-gcc di Windows).

---

## Panduan Instalasi & Jalankan Aplikasi

### Langkah 1: Clone atau Buka Direktori Proyek

Buka terminal dan masuk ke direktori proyek `SPOS-E`:

```bash
cd SPOS-E
```

---

### Langkah 2: Menjalankan Backend (Go Server)

1. Masuk ke direktori `backend`:
   ```bash
   cd backend
   ```
2. Unduh seluruh dependensi Go:
   ```bash
   go mod tidy
   ```
3. Jalankan server backend:
   ```bash
   go run cmd/api/main.go
   ```
   _Server backend akan berjalan secara default di `http://localhost:8080`._

---

### Langkah 3: Menjalankan Frontend (SvelteKit Development)

1. Buka terminal baru dan masuk ke direktori `frontend`:
   ```bash
   cd frontend
   ```
2. Pasang dependensi Node.js:
   ```bash
   npm install
   ```
3. Jalankan server frontend development:
   ```bash
   npm run dev
   ```
   _Aplikasi frontend akan dapat diakses melalui browser di `http://localhost:5173` (atau `http://localhost:8080` jika diakses via bundle statis backend)._

---

### Build Produksi Single-Binary (Deployment Produksi)

Untuk mendistribusikan aplikasi sebagai satu berkas biner mandiri yang sudah memuat aset frontend di dalam backend Go:

1. Compile frontend ke bentuk bundle statis:

   ```bash
   cd frontend
   npm run build
   ```

   _(Hasil build otomatis diekspor ke direktori `../backend/public`)_.

2. Compile backend Go menjadi executable binary:

   ```bash
   cd ../backend
   go build -o spos-engine cmd/api/main.go
   ```

3. Jalankan executable:
   ```bash
   ./spos-engine
   ```
   Buka browser Anda dan akses `http://localhost:8080`.

---

## Cara Penggunaan & Alur Kerja Aplikasi

### 1. Inisialisasi Toko (Setup Wizard Pertama Kali)

Saat pertama kali aplikasi dibuka di browser, **Setup Guard** akan secara otomatis mengarahkan Anda ke halaman **Setup Wizard** (`/setup`):

- **Langkah 1**: Isi Nama Toko/Outlet, Alamat, dan Nomor Telepon.
- **Langkah 2**: Buat Akun Superadmin (Username, Email, Password) & Buat **6-Digit PIN Supervisor**.
- **Langkah 3**: Pilih Database Engine (Gunakan **SQLite** untuk pengujian lokal, atau masukkan DSN **Supabase PostgreSQL** untuk produksi cloud).
- **Langkah 4**: Aktifkan modul fitur restoran yang dibutuhkan (Meja, KDS, WLD, BOM Stock, Midtrans).
- **Langkah 5**: Atur Service Charge (%), Pajak Resto (%), serta masukkan Midtrans Server Key & Client Key jika menggunakan QRIS Midtrans.
- Klik **Inisialisasi Sistem SPOS**. Sistem akan otomatis membuat tabel, akun admin, dan mengunci rute setup demi keamanan.

---

### 2. Membuka Shift Kasir (_Opening Shift_)

1. Login menggunakan akun Superadmin / Kasir yang telah dibuat.
2. Saat pertama kali masuk ke menu Kasir POS (`/pos`), sistem akan meminta Anda membuka Shift Register.
3. Masukkan nominal **Kas Awal (_Opening Cash_)** di laci uang (misal: `Rp 500.000`) dan klik **Buka Shift Register**.

---

### 3. Operasi Transaksi Kasir POS (`/pos`)

- **Menambah Menu**: Klik item menu pada katalog atau gunakan bilah pencarian (Tekan `F2` untuk pencarian cepat).
- **Pilih Layanan & Meja**: Pilih mode layanan (_Dine In_, _Take Away_, atau _Delivery_). Jika _Dine In_, pilih Meja yang tersedia pada tab _Denah Meja_.
- **Proses Pembayaran**: Klik **Proses Bayar (F4)**:
  - **Tunai (Cash)**: Masukkan nominal uang yang diterima pelanggan. Sistem akan langsung menampilkan jumlah kembalian.
  - **QRIS Midtrans**: Sistem akan menampilkan Kode QRIS Dinamis. Minta pelanggan melakukan scan via aplikasi e-wallet / m-banking. Setelah pembayaran diselesaikan, sistem otomatis mendeteksi status settlement dan melunasi pesanan.

---

### 4. Pelacakan Pesanan Dapur & Waiter (KDS & WLD)

- **Kitchen Display System (`/kds`)**: Tim dapur dapat melihat pesanan baru yang masuk, mengklik **Mulai Masak** saat hidangan diproses, dan mengklik **Siap Disajikan** setelah masakan matang.
- **Waitingline Display (`/wld`)**: Pramusaji dapat melihat daftar hidangan yang berstatus _Ready_ dan mengklik **Telah Disajikan** setelah makanan diantarkan ke meja pelanggan.

---

### 5. Pelunasan Transaksi Pending & Pembatalan (VOID)

- **Melanjutkan Pembayaran Pending**: Jika ada pesanan QRIS/Kasir yang tertunda, buka halaman **Riwayat Transaksi** (`/admin/orders`). Klik tombol **`PENDING (BAYAR)`** pada kolom status untuk melanjutkan pembayaran tanpa membuat order baru.
- **Batalkan Order (VOID)**: Klik tombol **Batal** pada transaksi (maksimal 30 menit dari pembuatan). Masukkan 6-Digit PIN Supervisor dan alasan pembatalan. Sistem akan mengosongkan meja dan memulihkan stok bahan baku inventaris secara otomatis.

---

### 6. Penutupan Shift Kasir & Laporan Penjualan

- **Closing Shift**: Pada akhir operasional, klik tombol **Tutup Shift** pada header POS. Masukkan jumlah fisik uang tunai di laci untuk diaudit dengan kalkulasi sistem.
- **Ekspor Laporan Penjualan (`/admin/reports`)**: Buka menu Laporan, pilih rentang tanggal yang diinginkan, lalu klik tombol **`Ekspor CSV`** untuk mengunduh laporan omset, COGS, dan laba kotor dalam format file CSV.

---

## Kebijakan Hak Cipta & Lisensi (License Agreement)

Software **SPOS-E (F&B Edition)** dilindungi oleh Undang-Undang Nomor 28 Tahun 2014 tentang Hak Cipta Republik Indonesia serta Perjanjian Lisensi Perangkat Lunak (_Software License Agreement_).

**Pemilik Hak Cipta**: Dillmux / The Lastofem / Berhembusnyaudara

### Skema Paket Lisensi:

1. **Paket Standard (Lisensi Binary / Installer Only / Hosted App)**:
   - Diberikan lisensi non-eksklusif untuk menginstal dan menjalankan file biner/installer pada perangkat sesuai kuota pembelian.
   - **Dilarang keras** melakukan pembongkaran kode (_reverse engineering_), dekompilasi, atau mencoba memodifikasi file biner Software.
   - Pengguna tidak berhak atas file kode sumber (_source code_) dalam bentuk apa pun.

2. **Paket Developer (Lisensi Source Code)**:
   - Diberikan akses penuh ke file kode sumber (_source code_) milik Software untuk penyesuaian kebutuhan internal atau proyek spesifik yang disetujui.
   - **Dilarang keras** mempublikasikan atau mengunggah kode sumber ini ke platform publik (seperti GitHub, GitLab, atau situs web publik lainnya) baik sebagian maupun seluruhnya tanpa persetujuan tertulis dari Pemilik Hak Cipta.

### Batasan Umum & Sanksi Hukum:

- **Dilarang keras** menjual kembali, menyewakan, membagikan, atau mendistribusikan ulang Software (baik berupa file installer maupun kode sumber) kepada pihak ketiga tanpa izin tertulis dari Pemilik Hak Cipta.
- Setiap pelanggaran terhadap ketentuan lisensi ini (terutama pembajakan atau penyebaran kode sumber secara ilegal) akan mengakibatkan pencabutan lisensi secara otomatis tanpa pengembalian uang (_refund_), dan Pemilik Hak Cipta berhak menempuh jalur hukum perdata maupun pidana sesuai undang-undang yang berlaku.
