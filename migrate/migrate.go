package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/Ikit777/Andalalin-Backend/initializers"
	"github.com/Ikit777/Andalalin-Backend/models"
	"github.com/Ikit777/Andalalin-Backend/utils"

	_ "time/tzdata"
)

func init() {
	config, err := initializers.LoadConfig()
	if err != nil {
		log.Fatal("Could not load environment variables", err)
	}

	initializers.ConnectDB(&config)
}

func removeExtension(fileName string) string {
	return path.Base(fileName[:len(fileName)-len(path.Ext(fileName))])
}

func main() {
	initializers.DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")

	initializers.DB.Migrator().DropTable(&models.User{})
	initializers.DB.Migrator().DropTable(&models.Andalalin{})
	initializers.DB.Migrator().DropTable(&models.Perlalin{})
	initializers.DB.Migrator().DropTable(&models.Survei{})
	initializers.DB.Migrator().DropTable(&models.SurveiMandiri{})
	initializers.DB.Migrator().DropTable(&models.TiketLevel1{})
	initializers.DB.Migrator().DropTable(&models.TiketLevel2{})
	initializers.DB.Migrator().DropTable(&models.Notifikasi{})
	initializers.DB.Migrator().DropTable(&models.DataMaster{})
	initializers.DB.Migrator().DropTable(&models.UsulanPengelolaan{})
	initializers.DB.Migrator().DropTable(&models.SurveiKepuasan{})
	initializers.DB.Migrator().DropTable(&models.Pemasangan{})

	initializers.DB.AutoMigrate(&models.User{})
	initializers.DB.AutoMigrate(&models.Andalalin{})
	initializers.DB.AutoMigrate(&models.Perlalin{})
	initializers.DB.AutoMigrate(&models.Survei{})
	initializers.DB.AutoMigrate(&models.SurveiMandiri{})
	initializers.DB.AutoMigrate(&models.TiketLevel1{})
	initializers.DB.AutoMigrate(&models.TiketLevel2{})
	initializers.DB.AutoMigrate(&models.Notifikasi{})
	initializers.DB.AutoMigrate(&models.DataMaster{})
	initializers.DB.AutoMigrate(&models.UsulanPengelolaan{})
	initializers.DB.AutoMigrate(&models.SurveiKepuasan{})
	initializers.DB.AutoMigrate(&models.Pemasangan{})

	loc, _ := time.LoadLocation("Asia/Singapore")
	now := time.Now().In(loc).Format("02-01-2006")
	hashedPassword, err := utils.HashPassword("superadmin")
	if err != nil {
		return
	}

	filePath := "assets/default.png"
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal("Error reading the file:", err)
	}

	initializers.DB.Create(&models.User{
		Name:      "Super admin",
		Email:     strings.ToLower("superadmin@gmail.com"),
		Password:  hashedPassword,
		Role:      "Super Admin",
		Photo:     fileData,
		Verified:  true,
		CreatedAt: now,
		UpdatedAt: now,
	})

	lokasi := []string{"Banjarmasin"}

	jenis_kegiatan := []string{"Pusat kegiatan", "Pemukiman", "Infrastruktur", "Lainnya"}

	pusat_kegiatan := []string{"Pusat perbelanjaan atau retail", "Perkantoran", "Industri dan pergudangan", "Sekolah atau universitas",
		"Lembaga kursus", "Rumah sakit", "Klinik bersama", "Bank", "Stasiun pengisin bahan bakar", "Hotel", "Gedung pertemuan",
		"Restoran", "Fasilitan olah raga", "Bengkel kendaraan bermotor", "Pencucian mobil"}
	infrastruktur := []string{"Akses ke dan dari jalan tol", "Pelabuhan", "Bandar udara", "Terminal", "Stasiun kereta api", "Pool kendaraan", "Fasilitas parkir umum", "Flyover", "Underpass", "Terowongan"}
	pemukiman := []string{"Perumahan sederhana", "Perumahan menengan-atas", "Rumah susun sederhana", "Apartemen", "Asrama", "Ruko"}

	rencana := []models.Rencana{}
	rencana = append(rencana, models.Rencana{Kategori: "Pusat kegiatan", JenisRencana: pusat_kegiatan})
	rencana = append(rencana, models.Rencana{Kategori: "Pemukiman", JenisRencana: pemukiman})
	rencana = append(rencana, models.Rencana{Kategori: "Infrastruktur", JenisRencana: infrastruktur})

	ketegori_perlengkapan := []string{"Rambu peringatan", "Rambu larangan", "Rambu perintah", "Rambu petunjunk", "Lainnya"}

	persyaratan := models.PersyaratanTambahan{
		PersyaratanTambahanAndalalin: []models.PersyaratanTambahanInput{},
		PersyaratanTambahanPerlalin:  []models.PersyaratanTambahanInput{},
	}

	perlengkapanPeringatan := []models.PerlengkapanItem{}

	folderPeringatan := "assets/Perlalin/Peringatan"

	folder1, err := os.Open(folderPeringatan)
	if err != nil {
		fmt.Println("Error opening folder:", err)
		return
	}
	defer folder1.Close()

	filePeringatan, err := folder1.Readdir(0)
	if err != nil {
		fmt.Println("Error reading folder contents:", err)
		return
	}

	for _, fileInfo := range filePeringatan {
		if fileInfo.Mode().IsRegular() {
			filePath := filepath.Join(folderPeringatan, fileInfo.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
				fmt.Printf("Error reading file %s: %v\n", fileInfo.Name(), err)
				continue
			}
			perlengkapanPeringatan = append(perlengkapanPeringatan, models.PerlengkapanItem{JenisPerlengkapan: removeExtension(fileInfo.Name()), GambarPerlengkapan: content})
		}
	}

	perlengkapanLarangan := []models.PerlengkapanItem{}

	folderLarangan := "assets/Perlalin/Larangan"

	folder2, err := os.Open(folderLarangan)
	if err != nil {
		fmt.Println("Error opening folder:", err)
		return
	}
	defer folder2.Close()

	fileLarangan, err := folder2.Readdir(0)
	if err != nil {
		fmt.Println("Error reading folder contents:", err)
		return
	}

	for _, fileInfo := range fileLarangan {
		if fileInfo.Mode().IsRegular() {
			filePath := filepath.Join(folderLarangan, fileInfo.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
				fmt.Printf("Error reading file %s: %v\n", fileInfo.Name(), err)
				continue
			}
			perlengkapanLarangan = append(perlengkapanLarangan, models.PerlengkapanItem{JenisPerlengkapan: removeExtension(fileInfo.Name()), GambarPerlengkapan: content})
		}
	}

	perlengkapanPerintah := []models.PerlengkapanItem{}

	folderPerintah := "assets/Perlalin/Perintah"

	folder3, err := os.Open(folderPerintah)
	if err != nil {
		fmt.Println("Error opening folder:", err)
		return
	}
	defer folder3.Close()

	filePerintah, err := folder3.Readdir(0)
	if err != nil {
		fmt.Println("Error reading folder contents:", err)
		return
	}

	for _, fileInfo := range filePerintah {
		if fileInfo.Mode().IsRegular() {
			filePath := filepath.Join(folderPerintah, fileInfo.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
				fmt.Printf("Error reading file %s: %v\n", fileInfo.Name(), err)
				continue
			}
			perlengkapanPerintah = append(perlengkapanPerintah, models.PerlengkapanItem{JenisPerlengkapan: removeExtension(fileInfo.Name()), GambarPerlengkapan: content})
		}
	}

	perlengkapanPetunjuk := []models.PerlengkapanItem{}

	folderPetunjuk := "assets/Perlalin/Petunjuk"

	folder4, err := os.Open(folderPetunjuk)
	if err != nil {
		fmt.Println("Error opening folder:", err)
		return
	}
	defer folder4.Close()

	filePetunjuk, err := folder4.Readdir(0)
	if err != nil {
		fmt.Println("Error reading folder contents:", err)
		return
	}

	for _, fileInfo := range filePetunjuk {
		if fileInfo.Mode().IsRegular() {
			filePath := filepath.Join(folderPetunjuk, fileInfo.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
				fmt.Printf("Error reading file %s: %v\n", fileInfo.Name(), err)
				continue
			}
			perlengkapanPetunjuk = append(perlengkapanPetunjuk, models.PerlengkapanItem{JenisPerlengkapan: removeExtension(fileInfo.Name()), GambarPerlengkapan: content})
		}
	}

	perlengkapan := []models.JenisPerlengkapan{}
	perlengkapan = append(perlengkapan, models.JenisPerlengkapan{Kategori: "Rambu peringatan", Perlengkapan: perlengkapanPeringatan})
	perlengkapan = append(perlengkapan, models.JenisPerlengkapan{Kategori: "Rambu larangan", Perlengkapan: perlengkapanLarangan})
	perlengkapan = append(perlengkapan, models.JenisPerlengkapan{Kategori: "Rambu perintah", Perlengkapan: perlengkapanPerintah})
	perlengkapan = append(perlengkapan, models.JenisPerlengkapan{Kategori: "Rambu petunjuk", Perlengkapan: perlengkapanPetunjuk})

	initializers.DB.Create(&models.DataMaster{
		LokasiPengambilan:       lokasi,
		JenisRencanaPembangunan: jenis_kegiatan,
		RencanaPembangunan:      rencana,
		PersyaratanTambahan:     persyaratan,
		KategoriPerlengkapan:    ketegori_perlengkapan,
		PerlengkapanLaluLintas:  perlengkapan,
	})

	fmt.Println("Migration complete")
}
