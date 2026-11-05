package main

import (
	"fmt"
	"log"
	"time"

	"wargapay-core/config"
	"wargapay-core/entity"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	_ = godotenv.Load()
	config.ConnectDB()

	log.Println("Memulai migrasi skema tabel...")
	err := config.DB.AutoMigrate(
		&entity.User{},
		&entity.WasteService{},
		&entity.Dues{},
		&entity.Transaction{},
		&entity.Expense{},
	)
	if err != nil {
		log.Fatalf("AutoMigrate gagal: %v", err)
	}
	log.Println("Migrasi tabel selesai.")

	// Hash password
	adminPass, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	wargaPass, _ := bcrypt.GenerateFromPassword([]byte("warga123"), bcrypt.DefaultCost)

	// 1. Seed Users
	users := []entity.User{
		{
			Name:     "Pengurus RT 05",
			Email:    "admin@wargapay.id",
			Password: string(adminPass),
			Role:     "admin",
		},
		{
			Name:     "Budi Santoso",
			Email:    "budi@wargapay.id",
			Password: string(wargaPass),
			Role:     "user",
		},
		{
			Name:     "Siti Rahmawati",
			Email:    "siti@wargapay.id",
			Password: string(wargaPass),
			Role:     "user",
		},
		{
			Name:     "Joko Prasetyo",
			Email:    "joko@wargapay.id",
			Password: string(wargaPass),
			Role:     "user",
		},
	}

	for i := range users {
		var existing entity.User
		if err := config.DB.Where("email = ?", users[i].Email).First(&existing).Error; err != nil {
			config.DB.Create(&users[i])
			log.Printf("User dibuat: %s (%s)", users[i].Name, users[i].Role)
		} else {
			users[i].ID = existing.ID
			log.Printf("User sudah ada: %s", users[i].Email)
		}
	}

	// 2. Seed Waste Services
	services := []entity.WasteService{
		{
			Name:        "Penjemputan Sampah Rumah Tangga Reguler",
			Description: "Pengambilan rutin sampah organik dan anorganik 3x seminggu langsung dari depan rumah.",
			Price:       35000,
			IsActive:    true,
		},
		{
			Name:        "Penjemputan Sampah Daur Ulang & Terpilah",
			Description: "Khusus sampah botol plastik, kardus, dan logam yang sudah dipilah rapi.",
			Price:       20000,
			IsActive:    true,
		},
		{
			Name:        "Penjemputan Puing & Perabot Bekas (Bulky Waste)",
			Description: "Layanan penjemputan barang besar seperti sofa, kasur, lemari, atau sisa renovasi.",
			Price:       75000,
			IsActive:    true,
		},
		{
			Name:        "Penjemputan Sampah Taman & Ranting",
			Description: "Pengangkutan dahan pohon, daun kering, dan limbah kebun cluster.",
			Price:       30000,
			IsActive:    true,
		},
	}

	for i := range services {
		var existing entity.WasteService
		if err := config.DB.Where("name = ?", services[i].Name).First(&existing).Error; err != nil {
			config.DB.Create(&services[i])
			log.Printf("Layanan sampah dibuat: %s", services[i].Name)
		} else {
			services[i].ID = existing.ID
		}
	}

	// 3. Seed Dues (Iuran Warga)
	now := time.Now()
	duesList := []entity.Dues{
		{
			UserID:      users[1].ID, // Budi
			Period:      "2026-10",
			WasteFee:    35000,
			SecurityFee: 30000,
			RTFund:      20000,
			Total:       85000,
			Status:      "paid",
			PaidAt:      &now,
		},
		{
			UserID:      users[1].ID, // Budi
			Period:      "2026-11",
			WasteFee:    35000,
			SecurityFee: 30000,
			RTFund:      20000,
			Total:       85000,
			Status:      "unpaid",
		},
		{
			UserID:      users[2].ID, // Siti
			Period:      "2026-11",
			WasteFee:    35000,
			SecurityFee: 30000,
			RTFund:      20000,
			Total:       85000,
			Status:      "paid",
			PaidAt:      &now,
		},
		{
			UserID:      users[3].ID, // Joko
			Period:      "2026-11",
			WasteFee:    35000,
			SecurityFee: 30000,
			RTFund:      20000,
			Total:       85000,
			Status:      "unpaid",
		},
	}

	for i := range duesList {
		var existing entity.Dues
		if err := config.DB.Where("user_id = ? AND period = ?", duesList[i].UserID, duesList[i].Period).First(&existing).Error; err != nil {
			config.DB.Create(&duesList[i])
			log.Printf("Tagihan iuran dibuat: User ID %d Periode %s (Status: %s)", duesList[i].UserID, duesList[i].Period, duesList[i].Status)
		} else {
			duesList[i].ID = existing.ID
		}
	}

	// 4. Seed Transactions
	pickupDate := now.AddDate(0, 0, 3)
	txList := []entity.Transaction{
		{
			UserID:          users[1].ID,
			OrderID:         fmt.Sprintf("WP-DUES-%d-1793850001", users[1].ID),
			Type:            "dues",
			DuesID:          &duesList[0].ID,
			Amount:          85000,
			Status:          "paid",
			PaymentType:     "bank_transfer",
			SnapToken:       "snap-budi-oct-paid",
			SnapRedirectURL: "https://app.sandbox.midtrans.com/snap/v2/vtweb/snap-budi-oct-paid",
			PaidAt:          &now,
		},
		{
			UserID:          users[2].ID,
			OrderID:         fmt.Sprintf("WP-DUES-%d-1793850002", users[2].ID),
			Type:            "dues",
			DuesID:          &duesList[2].ID,
			Amount:          85000,
			Status:          "paid",
			PaymentType:     "qris",
			SnapToken:       "snap-siti-nov-paid",
			SnapRedirectURL: "https://app.sandbox.midtrans.com/snap/v2/vtweb/snap-siti-nov-paid",
			PaidAt:          &now,
		},
		{
			UserID:          users[2].ID,
			OrderID:         fmt.Sprintf("WP-WAST-%d-1793850003", users[2].ID),
			Type:            "waste_service",
			WasteServiceID:  &services[2].ID, // Bulky waste
			Amount:          75000,
			Status:          "paid",
			PaymentType:     "gopay",
			SnapToken:       "snap-siti-waste-paid",
			SnapRedirectURL: "https://app.sandbox.midtrans.com/snap/v2/vtweb/snap-siti-waste-paid",
			Notes:           "Kasur busa dan meja belajar rusak di depan pagar",
			PickupAddress:   "Cluster Diamond Blok B No. 07",
			PickupDate:      &pickupDate,
			PaidAt:          &now,
		},
		{
			UserID:          users[3].ID,
			OrderID:         fmt.Sprintf("WP-WAST-%d-1793850004", users[3].ID),
			Type:            "waste_service",
			WasteServiceID:  &services[3].ID, // Ranting taman
			Amount:          30000,
			Status:          "pending",
			SnapToken:       "snap-joko-waste-pending",
			SnapRedirectURL: "https://app.sandbox.midtrans.com/snap/v2/vtweb/snap-joko-waste-pending",
			Notes:           "Dahan pohon mangga sudah diikat rapi",
			PickupAddress:   "Cluster Diamond Blok D No. 15",
			PickupDate:      &pickupDate,
		},
	}

	for i := range txList {
		var existing entity.Transaction
		if err := config.DB.Where("order_id = ?", txList[i].OrderID).First(&existing).Error; err != nil {
			config.DB.Create(&txList[i])
			log.Printf("Transaksi dibuat: %s (Status: %s, Total: Rp %d)", txList[i].OrderID, txList[i].Status, txList[i].Amount)
		}
	}

	// 5. Seed Pengeluaran Kas RT (Expenses)
	expenses := []entity.Expense{
		{
			Title:       "Honor & Operasional Petugas Keamanan Cluster",
			Category:    "Keamanan",
			Amount:      4500000,
			Vendor:      "Regu Satpam Pos Utama",
			Date:        "01 Nov 2026",
			Description: "Gaji bulanan 4 petugas satpam dan biaya ronda malam",
			ReceiptRef:  "KW-2026-001",
			ApprovedBy:  "Bpk. Hendra (Ketua RT)",
		},
		{
			Title:       "Retribusi Truk Pengangkutan Sampah DLH",
			Category:    "Kebersihan",
			Amount:      2800000,
			Vendor:      "Dinas Lingkungan Hidup / Vendor Mitra",
			Date:        "02 Nov 2026",
			Description: "Operasional truk pengangkut sampah 3x seminggu",
			ReceiptRef:  "KW-2026-002",
			ApprovedBy:  "Ibu Dewi (Bendahara)",
		},
		{
			Title:       "Peremajaan & Penggantian Lampu PJU Blok B",
			Category:    "Infrastruktur",
			Amount:      1250000,
			Vendor:      "Toko Listrik Terang Abadi",
			Date:        "03 Nov 2026",
			Description: "Penggantian 8 titik lampu LED PJU 50 Watt",
			ReceiptRef:  "KW-2026-003",
			ApprovedBy:  "Bpk. Agus (Seksi Pembangunan)",
		},
		{
			Title:       "Pengadaan Refill APAR & Kotak P3K Pos RT",
			Category:    "Operasional",
			Amount:      650000,
			Vendor:      "CV Citra Safety Mandiri",
			Date:        "04 Nov 2026",
			Description: "Inspeksi berkala dan refill 3 tabung pemadam pos satpam",
			ReceiptRef:  "KW-2026-004",
			ApprovedBy:  "Bpk. Hendra (Ketua RT)",
		},
	}

	for i := range expenses {
		var existing entity.Expense
		if err := config.DB.Where("receipt_ref = ?", expenses[i].ReceiptRef).First(&existing).Error; err != nil {
			config.DB.Create(&expenses[i])
			log.Printf("Pengeluaran kas dibuat: %s (Rp %d)", expenses[i].Title, expenses[i].Amount)
		}
	}

	log.Println("Seeding database WargaPay berhasil 100%!")
}

