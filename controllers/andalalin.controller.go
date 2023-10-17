package controllers

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Ikit777/Andalalin-Backend/initializers"
	"github.com/Ikit777/Andalalin-Backend/models"
	"github.com/Ikit777/Andalalin-Backend/repository"
	"github.com/Ikit777/Andalalin-Backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	_ "time/tzdata"

	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
)

type AndalalinController struct {
	DB *gorm.DB
}

type data struct {
	Jenis string
	Nilai int
	Hasil string
}

type komentar struct {
	Nama     string
	Komentar string
}

func interval(hasil float64) string {
	intervalNilai := ""
	if hasil < 26.00 {
		cekInterval := float64(hasil) / float64(25.00)
		intervalNilai = fmt.Sprintf("%.2f", cekInterval)
	} else if hasil >= 43.76 && hasil <= 62.50 {
		cekInterval := float64(hasil) / float64(25.00)
		intervalNilai = fmt.Sprintf("%.2f", cekInterval)
	} else if hasil >= 62.51 && hasil <= 81.25 {
		cekInterval := float64(hasil) / float64(25.00)
		intervalNilai = fmt.Sprintf("%.2f", cekInterval)
	} else if hasil >= 81.26 && hasil <= 100 {
		cekInterval := float64(hasil) / float64(25.00)
		intervalNilai = fmt.Sprintf("%.2f", cekInterval)
	}
	return intervalNilai
}

func mutu(hasil float64) string {
	mutuNilai := ""
	if hasil <= 43.75 {
		mutuNilai = "D"
	} else if hasil >= 43.76 && hasil <= 62.50 {
		mutuNilai = "C"
	} else if hasil >= 62.51 && hasil <= 81.25 {
		mutuNilai = "B"
	} else if hasil >= 81.26 && hasil <= 100 {
		mutuNilai = "A"
	}
	return mutuNilai
}

func kinerja(hasil float64) string {
	kinerjaNilai := ""
	if hasil <= 43.75 {
		kinerjaNilai = "Buruk"
	} else if hasil >= 43.76 && hasil <= 62.50 {
		kinerjaNilai = "Kurang baik"
	} else if hasil >= 62.51 && hasil <= 81.25 {
		kinerjaNilai = "Baik"
	} else if hasil >= 81.26 && hasil <= 100 {
		kinerjaNilai = "Sangat baik"
	}
	return kinerjaNilai
}

func getStartOfMonth(year int, month time.Month) time.Time {
	return time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
}

func getEndOfMonth(year int, month time.Month) time.Time {
	nextMonth := getStartOfMonth(year, month).AddDate(0, 1, 0)
	return nextMonth.Add(-time.Second)
}

func NewAndalalinController(DB *gorm.DB) AndalalinController {
	return AndalalinController{DB}
}

func (ac *AndalalinController) Pengajuan(ctx *gin.Context) {
	var payload *models.DataAndalalin
	currentUser := ctx.MustGet("currentUser").(models.User)

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.AndalalinPengajuanCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	if err := ctx.ShouldBind(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	loc, _ := time.LoadLocation("Asia/Singapore")
	nowTime := time.Now().In(loc)

	kode := "andalalin/" + utils.Generate(6)
	tanggal := nowTime.Format("02") + " " + utils.Bulan(nowTime.Month()) + " " + nowTime.Format("2006")

	t, err := template.ParseFiles("templates/tandaterimaTemplate.html")
	if err != nil {
		log.Fatal("Error reading the email template:", err)
		return
	}

	bukti := struct {
		Tanggal      string
		Waktu        string
		Kode         string
		Nama         string
		Instansi     string
		Nomor        string
		NomorSeluler string
	}{
		Tanggal:      tanggal,
		Waktu:        nowTime.Format("15:04:05"),
		Kode:         kode,
		Nama:         currentUser.Name,
		Instansi:     payload.Andalalin.NamaPerusahaan,
		Nomor:        payload.Andalalin.NomerPemohon,
		NomorSeluler: payload.Andalalin.NomerSelulerPemohon,
	}

	buffer := new(bytes.Buffer)
	if err = t.Execute(buffer, bukti); err != nil {
		log.Fatal("Eror saat membaca template:", err)
		return
	}

	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		log.Fatal("Eror generate pdf", err)
		return
	}

	// read the HTML page as a PDF page
	page := wkhtmltopdf.NewPageReader(bytes.NewReader(buffer.Bytes()))

	pdfg.AddPage(page)

	pdfg.Dpi.Set(300)
	pdfg.PageSize.Set(wkhtmltopdf.PageSizeA4)
	pdfg.Orientation.Set(wkhtmltopdf.OrientationPortrait)
	pdfg.MarginBottom.Set(20)
	pdfg.MarginLeft.Set(30)
	pdfg.MarginRight.Set(30)
	pdfg.MarginTop.Set(20)

	err = pdfg.Create()
	if err != nil {
		log.Fatal(err)
	}

	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	blobs := make(map[string][]byte)

	tambahan := []models.PersyaratanTambahanPermohonan{}

	for key, files := range form.File {
		for _, file := range files {
			// Save the uploaded file with key as prefix
			file, err := file.Open()

			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			defer file.Close()

			data, err := io.ReadAll(file)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			// Store the blob data in the map

			switch key {
			case "ktp":
				blobs[key] = data
			case "apb":
				blobs[key] = data
			case "sk":
				blobs[key] = data
			default:
				tambahan = append(tambahan, models.PersyaratanTambahanPermohonan{Persyaratan: key, Berkas: data})
			}
		}
	}

	permohonan := models.Andalalin{
		IdUser:                 currentUser.ID,
		JenisAndalalin:         "Dokumen analisa dampak lalu lintas",
		Kategori:               payload.Andalalin.KategoriJenisRencanaPembangunan,
		Jenis:                  payload.Andalalin.JenisRencanaPembangunan,
		Kode:                   kode,
		NikPemohon:             payload.Andalalin.NikPemohon,
		NamaPemohon:            currentUser.Name,
		EmailPemohon:           currentUser.Email,
		TempatLahirPemohon:     payload.Andalalin.TempatLahirPemohon,
		TanggalLahirPemohon:    payload.Andalalin.TanggalLahirPemohon,
		AlamatPemohon:          payload.Andalalin.AlamatPemohon,
		JenisKelaminPemohon:    payload.Andalalin.JenisKelaminPemohon,
		NomerPemohon:           payload.Andalalin.NomerPemohon,
		NomerSelulerPemohon:    payload.Andalalin.NomerSelulerPemohon,
		JabatanPemohon:         payload.Andalalin.JabatanPemohon,
		LokasiPengambilan:      payload.Andalalin.LokasiPengambilan,
		WaktuAndalalin:         nowTime.Format("15:04:05"),
		TanggalAndalalin:       tanggal,
		StatusAndalalin:        "Cek persyaratan",
		TandaTerimaPendaftaran: pdfg.Bytes(),

		NamaPerusahaan:       payload.Andalalin.NamaPerusahaan,
		AlamatPerusahaan:     payload.Andalalin.AlamatPerusahaan,
		NomerPerusahaan:      payload.Andalalin.NomerPerusahaan,
		EmailPerusahaan:      payload.Andalalin.EmailPerusahaan,
		ProvinsiPerusahaan:   payload.Andalalin.ProvinsiPerusahaan,
		KabupatenPerusahaan:  payload.Andalalin.KabupatenPerusahaan,
		KecamatanPerusahaan:  payload.Andalalin.KecamatanPerusahaan,
		KelurahaanPerusahaan: payload.Andalalin.KelurahaanPerusahaan,
		NamaPimpinan:         payload.Andalalin.NamaPimpinan,
		JabatanPimpinan:      payload.Andalalin.JabatanPimpinan,
		JenisKelaminPimpinan: payload.Andalalin.JenisKelaminPimpinan,
		JenisKegiatan:        payload.Andalalin.JenisKegiatan,
		Peruntukan:           payload.Andalalin.Peruntukan,
		LuasLahan:            payload.Andalalin.LuasLahan + "m²",
		AlamatPersil:         payload.Andalalin.AlamatPersil,
		KelurahanPersil:      payload.Andalalin.KelurahanPersil,
		NomerSKRK:            payload.Andalalin.NomerSKRK,
		TanggalSKRK:          payload.Andalalin.TanggalSKRK,

		KartuTandaPenduduk:  blobs["ktp"],
		AktaPendirianBadan:  blobs["apb"],
		SuratKuasa:          blobs["sk"],
		PersyaratanTambahan: tambahan,
	}

	result := ac.DB.Create(&permohonan)

	respone := &models.DaftarAndalalinResponse{
		IdAndalalin:      permohonan.IdAndalalin,
		Kode:             permohonan.Kode,
		TanggalAndalalin: permohonan.TanggalAndalalin,
		Nama:             permohonan.NamaPemohon,
		Alamat:           permohonan.AlamatPemohon,
		JenisAndalalin:   permohonan.JenisAndalalin,
		StatusAndalalin:  permohonan.StatusAndalalin,
	}

	if result.Error != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "eror saat mengirim data"})
		return
	} else {
		ac.ReleaseTicketLevel1(ctx, permohonan.IdAndalalin)
		ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": respone})
	}
}

func (ac *AndalalinController) PengajuanPerlalin(ctx *gin.Context) {
	var payload *models.DataPerlalin
	currentUser := ctx.MustGet("currentUser").(models.User)

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.AndalalinPengajuanCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	if err := ctx.ShouldBind(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	loc, _ := time.LoadLocation("Asia/Singapore")
	nowTime := time.Now().In(loc)

	kode := "perlalin/" + utils.Generate(6)
	tanggal := nowTime.Format("02") + " " + utils.Bulan(nowTime.Month()) + " " + nowTime.Format("2006")

	t, err := template.ParseFiles("templates/tandaterimaPerlalin.html")
	if err != nil {
		log.Fatal("Error reading the email template:", err)
		return
	}

	bukti := struct {
		Tanggal      string
		Waktu        string
		Kode         string
		Nama         string
		Nomor        string
		NomorSeluler string
	}{
		Tanggal:      tanggal,
		Waktu:        nowTime.Format("15:04:05"),
		Kode:         kode,
		Nama:         currentUser.Name,
		Nomor:        payload.Perlalin.NomerPemohon,
		NomorSeluler: payload.Perlalin.NomerSelulerPemohon,
	}

	buffer := new(bytes.Buffer)
	if err = t.Execute(buffer, bukti); err != nil {
		log.Fatal("Eror saat membaca template:", err)
		return
	}

	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		log.Fatal("Eror generate pdf", err)
		return
	}

	// read the HTML page as a PDF page
	page := wkhtmltopdf.NewPageReader(bytes.NewReader(buffer.Bytes()))

	pdfg.AddPage(page)

	pdfg.Dpi.Set(300)
	pdfg.PageSize.Set(wkhtmltopdf.PageSizeA4)
	pdfg.Orientation.Set(wkhtmltopdf.OrientationPortrait)
	pdfg.MarginBottom.Set(20)
	pdfg.MarginLeft.Set(30)
	pdfg.MarginRight.Set(30)
	pdfg.MarginTop.Set(20)

	err = pdfg.Create()
	if err != nil {
		log.Fatal(err)
	}

	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	blobs := make(map[string][]byte)

	tambahan := []models.PersyaratanTambahanPermohonan{}

	for key, files := range form.File {
		for _, file := range files {
			// Save the uploaded file with key as prefix
			file, err := file.Open()

			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			defer file.Close()

			data, err := io.ReadAll(file)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			// Store the blob data in the map

			switch key {
			case "ktp":
				blobs[key] = data
			case "sp":
				blobs[key] = data
			default:
				tambahan = append(tambahan, models.PersyaratanTambahanPermohonan{Persyaratan: key, Berkas: data})
			}
		}
	}

	permohonan := models.Perlalin{
		IdUser:                 currentUser.ID,
		JenisAndalalin:         "Perlengkapan lalu lintas",
		Kategori:               payload.Perlalin.Kategori,
		Jenis:                  payload.Perlalin.Jenis,
		Kode:                   kode,
		NikPemohon:             payload.Perlalin.NikPemohon,
		NamaPemohon:            currentUser.Name,
		EmailPemohon:           currentUser.Email,
		TempatLahirPemohon:     payload.Perlalin.TempatLahirPemohon,
		TanggalLahirPemohon:    payload.Perlalin.TanggalLahirPemohon,
		AlamatPemohon:          payload.Perlalin.AlamatPemohon,
		JenisKelaminPemohon:    payload.Perlalin.JenisKelaminPemohon,
		NomerPemohon:           payload.Perlalin.NomerPemohon,
		NomerSelulerPemohon:    payload.Perlalin.NomerSelulerPemohon,
		LokasiPengambilan:      payload.Perlalin.LokasiPengambilan,
		WaktuAndalalin:         nowTime.Format("15:04:05"),
		TanggalAndalalin:       tanggal,
		JenisKegiatan:          payload.Perlalin.JenisKegiatan,
		Peruntukan:             payload.Perlalin.Peruntukan,
		LuasLahan:              payload.Perlalin.LuasLahan + "m²",
		AlamatPersil:           payload.Perlalin.AlamatPersil,
		KelurahanPersil:        payload.Perlalin.KelurahanPersil,
		StatusAndalalin:        "Cek persyaratan",
		TandaTerimaPendaftaran: pdfg.Bytes(),

		KartuTandaPenduduk:  blobs["ktp"],
		SuratPermohonan:     blobs["sp"],
		PersyaratanTambahan: tambahan,
	}

	result := ac.DB.Create(&permohonan)

	respone := &models.DaftarAndalalinResponse{
		IdAndalalin:      permohonan.IdAndalalin,
		Kode:             permohonan.Kode,
		TanggalAndalalin: permohonan.TanggalAndalalin,
		Nama:             permohonan.NamaPemohon,
		Alamat:           permohonan.AlamatPemohon,
		JenisAndalalin:   permohonan.JenisAndalalin,
		StatusAndalalin:  permohonan.StatusAndalalin,
	}

	if result.Error != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "eror saat mengirim data"})
		return
	} else {
		ac.ReleaseTicketLevel1(ctx, permohonan.IdAndalalin)
		ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": respone})
	}
}

func (ac *AndalalinController) ReleaseTicketLevel1(ctx *gin.Context, id uuid.UUID) {
	tiket := models.TiketLevel1{
		IdAndalalin: id,
		Status:      "Buka",
	}

	result := ac.DB.Create(&tiket)

	if result.Error != nil && strings.Contains(result.Error.Error(), "duplicate key value violates unique") {
		ctx.JSON(http.StatusConflict, gin.H{"status": "fail", "message": "Tiket level 1 sudah tersedia"})
		return
	} else if result.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": "Telah terjadi sesuatu"})
		return
	}
}

func (ac *AndalalinController) CloseTiketLevel1(ctx *gin.Context, id uuid.UUID) {
	var tiket models.TiketLevel1

	result := ac.DB.Model(&tiket).Where("id_andalalin = ? AND status = ?", id, "Buka").Update("status", "Tutup")
	if result.Error != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Tiket level 1 tidak tersedia"})
		return
	}
}

func (ac *AndalalinController) ReleaseTicketLevel2(ctx *gin.Context, id uuid.UUID, petugas uuid.UUID) {
	var tiket1 models.TiketLevel1
	results := ac.DB.First(&tiket1, "id_andalalin = ?", id)

	if results.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": results.Error})
		return
	}

	tiket := models.TiketLevel2{
		IdTiketLevel1: tiket1.IdTiketLevel1,
		IdAndalalin:   id,
		IdPetugas:     petugas,
		Status:        "Buka",
	}

	result := ac.DB.Create(&tiket)

	if result.Error != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Telah terjadi sesuatu"})
		return
	}
}

func (ac *AndalalinController) CloseTiketLevel2(ctx *gin.Context, id uuid.UUID) {
	var tiket models.TiketLevel2

	result := ac.DB.Model(&tiket).Where("id_andalalin = ? AND status = ?", id, "Buka").Or("id_andalalin = ? AND status = ?", id, "Batal").Update("status", "Tutup")
	if result.Error != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Telah terjadi sesuatu"})
		return
	}
}

func (ac *AndalalinController) GetPermohonanByIdUser(ctx *gin.Context) {
	currentUser := ctx.MustGet("currentUser").(models.User)

	var andalalin []models.Andalalin
	var perlalin []models.Perlalin

	resultsAndalalin := ac.DB.Order("tanggal_andalalin").Find(&andalalin, "id_user = ?", currentUser.ID)
	resultsPerlalin := ac.DB.Order("tanggal_andalalin").Find(&perlalin, "id_user = ?", currentUser.ID)

	if resultsAndalalin.Error != nil && resultsPerlalin != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": "Tidak ditemukan"})
		return
	} else {
		var respone []models.DaftarAndalalinResponse
		for _, s := range andalalin {
			respone = append(respone, models.DaftarAndalalinResponse{
				IdAndalalin:      s.IdAndalalin,
				Kode:             s.Kode,
				TanggalAndalalin: s.TanggalAndalalin,
				Nama:             s.NamaPemohon,
				Alamat:           s.AlamatPemohon,
				JenisAndalalin:   s.JenisAndalalin,
				StatusAndalalin:  s.StatusAndalalin,
			})
		}
		for _, s := range perlalin {
			respone = append(respone, models.DaftarAndalalinResponse{
				IdAndalalin:      s.IdAndalalin,
				Kode:             s.Kode,
				TanggalAndalalin: s.TanggalAndalalin,
				Nama:             s.NamaPemohon,
				Alamat:           s.AlamatPemohon,
				JenisAndalalin:   s.JenisAndalalin,
				StatusAndalalin:  s.StatusAndalalin,
			})
		}
		ctx.JSON(http.StatusOK, gin.H{"status": "success", "results": len(respone), "data": respone})
	}
}

func (ac *AndalalinController) GetPermohonanByIdAndalalin(ctx *gin.Context) {
	id := ctx.Param("id_andalalin")

	currentUser := ctx.MustGet("currentUser").(models.User)

	var andalalin models.Andalalin
	var perlalin models.Perlalin

	resultsAndalalin := ac.DB.First(&andalalin, "id_andalalin = ?", id)
	resultsPerlalin := ac.DB.First(&perlalin, "id_andalalin = ?", id)

	if resultsAndalalin.Error != nil && resultsPerlalin.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": "Tidak ditemukan"})
		return
	}

	var ticket2 models.TiketLevel2
	resultTiket2 := ac.DB.Not("status = ?", "Tutup").Where("id_andalalin = ?", id).First(&ticket2)
	var status string
	if resultTiket2.Error != nil {
		status = "Kosong"
	} else {
		status = ticket2.Status
	}

	if andalalin.IdAndalalin != uuid.Nil {
		if currentUser.Role == "User" {
			dataUser := models.AndalalinResponseUser{
				IdAndalalin:             andalalin.IdAndalalin,
				JenisAndalalin:          andalalin.JenisAndalalin,
				JenisRencanaPembangunan: andalalin.Jenis,
				Kategori:                andalalin.Kategori,
				Kode:                    andalalin.Kode,
				NamaPemohon:             andalalin.NamaPemohon,
				LokasiPengambilan:       andalalin.LokasiPengambilan,
				TanggalAndalalin:        andalalin.TanggalAndalalin,
				StatusAndalalin:         andalalin.StatusAndalalin,
				TandaTerimaPendaftaran:  andalalin.TandaTerimaPendaftaran,
				NamaPerusahaan:          andalalin.NamaPerusahaan,
				JenisKegiatan:           andalalin.JenisKegiatan,
				Peruntukan:              andalalin.Peruntukan,
				LuasLahan:               andalalin.LuasLahan,
				PersyaratanTidakSesuai:  andalalin.PersyaratanTidakSesuai,
				FileSK:                  andalalin.FileSK,
			}

			ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": dataUser})
		} else {
			data := models.AndalalinResponse{
				IdAndalalin:                  andalalin.IdAndalalin,
				JenisAndalalin:               andalalin.JenisAndalalin,
				Kategori:                     andalalin.Kategori,
				Jenis:                        andalalin.Jenis,
				Kode:                         andalalin.Kode,
				NikPemohon:                   andalalin.NikPemohon,
				NamaPemohon:                  andalalin.NamaPemohon,
				EmailPemohon:                 andalalin.EmailPemohon,
				TempatLahirPemohon:           andalalin.TempatLahirPemohon,
				TanggalLahirPemohon:          andalalin.TanggalLahirPemohon,
				AlamatPemohon:                andalalin.AlamatPemohon,
				JenisKelaminPemohon:          andalalin.JenisKelaminPemohon,
				NomerPemohon:                 andalalin.NomerPemohon,
				NomerSelulerPemohon:          andalalin.NomerSelulerPemohon,
				JabatanPemohon:               andalalin.JabatanPemohon,
				LokasiPengambilan:            andalalin.LokasiPengambilan,
				WaktuAndalalin:               andalalin.WaktuAndalalin,
				TanggalAndalalin:             andalalin.TanggalAndalalin,
				StatusAndalalin:              andalalin.StatusAndalalin,
				TandaTerimaPendaftaran:       andalalin.TandaTerimaPendaftaran,
				NamaPerusahaan:               andalalin.NamaPerusahaan,
				AlamatPerusahaan:             andalalin.AlamatPerusahaan,
				NomerPerusahaan:              andalalin.NomerPerusahaan,
				EmailPerusahaan:              andalalin.EmailPerusahaan,
				ProvinsiPerusahaan:           andalalin.ProvinsiPerusahaan,
				KabupatenPerusahaan:          andalalin.KabupatenPerusahaan,
				KecamatanPerusahaan:          andalalin.KecamatanPerusahaan,
				KelurahaanPerusahaan:         andalalin.KelurahaanPerusahaan,
				NamaPimpinan:                 andalalin.NamaPimpinan,
				JabatanPimpinan:              andalalin.JabatanPimpinan,
				JenisKelaminPimpinan:         andalalin.JenisKelaminPimpinan,
				JenisKegiatan:                andalalin.JenisKegiatan,
				Peruntukan:                   andalalin.Peruntukan,
				LuasLahan:                    andalalin.LuasLahan,
				AlamatPersil:                 andalalin.AlamatPersil,
				KelurahanPersil:              andalalin.KelurahanPersil,
				NomerSKRK:                    andalalin.NomerSKRK,
				TanggalSKRK:                  andalalin.TanggalSKRK,
				KartuTandaPenduduk:           andalalin.KartuTandaPenduduk,
				AktaPendirianBadan:           andalalin.AktaPendirianBadan,
				SuratKuasa:                   andalalin.SuratKuasa,
				PersyaratanTidakSesuai:       andalalin.PersyaratanTidakSesuai,
				StatusTiketLevel2:            status,
				PersetujuanDokumen:           andalalin.PersetujuanDokumen,
				KeteranganPersetujuanDokumen: andalalin.KeteranganPersetujuanDokumen,
				NomerBAPDasar:                andalalin.NomerBAPDasar,
				NomerBAPPelaksanaan:          andalalin.NomerBAPPelaksanaan,
				TanggalBAP:                   andalalin.TanggalBAP,
				FileBAP:                      andalalin.FileBAP,
				FileSK:                       andalalin.FileSK,
				PersyaratanTambahan:          andalalin.PersyaratanTambahan,
			}
			ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": data})
		}
	}

	if perlalin.IdAndalalin != uuid.Nil {
		if currentUser.Role == "User" {
			dataUser := models.AndalalinResponseUser{
				IdAndalalin:             perlalin.IdAndalalin,
				JenisAndalalin:          perlalin.JenisAndalalin,
				Kategori:                perlalin.Kategori,
				JenisRencanaPembangunan: perlalin.Jenis,
				Kode:                    perlalin.Kode,
				NamaPemohon:             perlalin.NamaPemohon,
				LokasiPengambilan:       perlalin.LokasiPengambilan,
				TanggalAndalalin:        perlalin.TanggalAndalalin,
				StatusAndalalin:         perlalin.StatusAndalalin,
				TandaTerimaPendaftaran:  perlalin.TandaTerimaPendaftaran,
				JenisKegiatan:           perlalin.JenisKegiatan,
				Peruntukan:              perlalin.Peruntukan,
				LuasLahan:               perlalin.LuasLahan,
				PersyaratanTidakSesuai:  perlalin.PersyaratanTidakSesuai,
			}

			ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": dataUser})
		} else {
			data := models.PerlalinResponse{
				IdAndalalin:            perlalin.IdAndalalin,
				JenisAndalalin:         perlalin.JenisAndalalin,
				Kategori:               perlalin.Kategori,
				Jenis:                  perlalin.Jenis,
				Kode:                   perlalin.Kode,
				NikPemohon:             perlalin.NikPemohon,
				NamaPemohon:            perlalin.NamaPemohon,
				EmailPemohon:           perlalin.EmailPemohon,
				TempatLahirPemohon:     perlalin.TempatLahirPemohon,
				TanggalLahirPemohon:    perlalin.TanggalLahirPemohon,
				AlamatPemohon:          perlalin.AlamatPemohon,
				JenisKelaminPemohon:    perlalin.JenisKelaminPemohon,
				NomerPemohon:           perlalin.NomerPemohon,
				NomerSelulerPemohon:    perlalin.NomerSelulerPemohon,
				LokasiPengambilan:      perlalin.LokasiPengambilan,
				WaktuAndalalin:         perlalin.WaktuAndalalin,
				TanggalAndalalin:       perlalin.TanggalAndalalin,
				StatusAndalalin:        perlalin.StatusAndalalin,
				TandaTerimaPendaftaran: perlalin.TandaTerimaPendaftaran,
				JenisKegiatan:          perlalin.JenisKegiatan,
				Peruntukan:             perlalin.Peruntukan,
				LuasLahan:              perlalin.LuasLahan,
				AlamatPersil:           perlalin.AlamatPersil,
				KelurahanPersil:        perlalin.KelurahanPersil,
				KartuTandaPenduduk:     perlalin.KartuTandaPenduduk,
				SuratPermohonan:        perlalin.SuratPermohonan,
				PersyaratanTidakSesuai: perlalin.PersyaratanTidakSesuai,
				IdPetugas:              perlalin.IdPetugas,
				NamaPetugas:            perlalin.NamaPetugas,
				EmailPetugas:           perlalin.EmailPetugas,
				StatusTiketLevel2:      status,
				LaporanSurvei:          perlalin.LaporanSurvei,
				PersyaratanTambahan:    perlalin.PersyaratanTambahan,
				Tindakan:               perlalin.Tindakan,
				PertimbanganTindakan:   perlalin.PertimbanganTindakan,
			}
			ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": data})
		}
	}
}

func (ac *AndalalinController) GetPermohonanByStatus(ctx *gin.Context) {
	status := ctx.Param("status_andalalin")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.AndalalinGetCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var andalalin []models.Andalalin
	var perlalin []models.Perlalin

	resultsAndalalin := ac.DB.Order("tanggal_andalalin").Find(&andalalin, "status_andalalin = ?", status)
	resultsPerlalin := ac.DB.Order("tanggal_andalalin").Find(&perlalin, "status_andalalin = ?", status)

	if resultsAndalalin.Error != nil && resultsPerlalin != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": "Tidak ditemukan"})
		return
	} else {
		var respone []models.DaftarAndalalinResponse
		for _, s := range andalalin {
			respone = append(respone, models.DaftarAndalalinResponse{
				IdAndalalin:      s.IdAndalalin,
				Kode:             s.Kode,
				TanggalAndalalin: s.TanggalAndalalin,
				Nama:             s.NamaPemohon,
				Alamat:           s.AlamatPemohon,
				JenisAndalalin:   s.JenisAndalalin,
				StatusAndalalin:  s.StatusAndalalin,
			})
		}
		for _, s := range perlalin {
			respone = append(respone, models.DaftarAndalalinResponse{
				IdAndalalin:      s.IdAndalalin,
				Kode:             s.Kode,
				TanggalAndalalin: s.TanggalAndalalin,
				Nama:             s.NamaPemohon,
				Alamat:           s.AlamatPemohon,
				JenisAndalalin:   s.JenisAndalalin,
				StatusAndalalin:  s.StatusAndalalin,
			})
		}
		ctx.JSON(http.StatusOK, gin.H{"status": "success", "results": len(respone), "data": respone})
	}
}

func (ac *AndalalinController) GetPermohonan(ctx *gin.Context) {
	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.AndalalinGetCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var andalalin []models.Andalalin
	var perlalin []models.Perlalin

	resultsAndalalin := ac.DB.Order("tanggal_andalalin").Find(&andalalin)
	resultsPerlalin := ac.DB.Order("tanggal_andalalin").Find(&perlalin)

	if resultsAndalalin.Error != nil && resultsPerlalin != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": "Tidak ditemukan"})
		return
	} else {
		var respone []models.DaftarAndalalinResponse
		for _, s := range andalalin {
			respone = append(respone, models.DaftarAndalalinResponse{
				IdAndalalin:      s.IdAndalalin,
				Kode:             s.Kode,
				TanggalAndalalin: s.TanggalAndalalin,
				Nama:             s.NamaPemohon,
				Alamat:           s.AlamatPemohon,
				JenisAndalalin:   s.JenisAndalalin,
				StatusAndalalin:  s.StatusAndalalin,
			})
		}
		for _, s := range perlalin {
			respone = append(respone, models.DaftarAndalalinResponse{
				IdAndalalin:      s.IdAndalalin,
				Kode:             s.Kode,
				TanggalAndalalin: s.TanggalAndalalin,
				Nama:             s.NamaPemohon,
				Alamat:           s.AlamatPemohon,
				JenisAndalalin:   s.JenisAndalalin,
				StatusAndalalin:  s.StatusAndalalin,
			})
		}
		ctx.JSON(http.StatusOK, gin.H{"status": "success", "results": len(respone), "data": respone})
	}
}

func (ac *AndalalinController) GetAndalalinTicketLevel1(ctx *gin.Context) {
	status := ctx.Param("status")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.AndalalinTicket1Credential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var ticket []models.TiketLevel1

	results := ac.DB.Find(&ticket, "status = ?", status)

	if results.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": results.Error})
		return
	} else {
		var respone []models.DaftarAndalalinResponse
		for _, s := range ticket {
			var andalalin models.Andalalin
			var perlalin models.Perlalin
			ac.DB.First(&andalalin, "id_andalalin = ?", s.IdAndalalin)
			ac.DB.First(&perlalin, "id_andalalin = ?", s.IdAndalalin)

			resultsAndalalin := ac.DB.First(&andalalin, "id_andalalin = ?", s.IdAndalalin)
			resultsPerlalin := ac.DB.First(&perlalin, "id_andalalin = ?", s.IdAndalalin)

			if resultsAndalalin.Error != nil && resultsPerlalin.Error != nil {
				ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": "Tidak ditemukan"})
				return
			}

			if andalalin.IdAndalalin != uuid.Nil {
				respone = append(respone, models.DaftarAndalalinResponse{
					IdAndalalin:      andalalin.IdAndalalin,
					Kode:             andalalin.Kode,
					TanggalAndalalin: andalalin.TanggalAndalalin,
					Nama:             andalalin.NamaPemohon,
					Alamat:           andalalin.AlamatPemohon,
					JenisAndalalin:   andalalin.JenisAndalalin,
					StatusAndalalin:  andalalin.StatusAndalalin,
				})
			}

			if perlalin.IdAndalalin != uuid.Nil {
				respone = append(respone, models.DaftarAndalalinResponse{
					IdAndalalin:      perlalin.IdAndalalin,
					Kode:             perlalin.Kode,
					TanggalAndalalin: perlalin.TanggalAndalalin,
					Nama:             perlalin.NamaPemohon,
					Alamat:           perlalin.AlamatPemohon,
					JenisAndalalin:   perlalin.JenisAndalalin,
					StatusAndalalin:  perlalin.StatusAndalalin,
				})
			}

		}
		ctx.JSON(http.StatusOK, gin.H{"status": "success", "results": len(respone), "data": respone})
	}
}

func (ac *AndalalinController) UpdatePersyaratan(ctx *gin.Context) {
	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.AndalalinUpdateCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	id := ctx.Param("id_andalalin")
	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var andalalin *models.Andalalin
	var perlalin *models.Perlalin

	resultsAndalalin := ac.DB.Find(&andalalin, "id_andalalin = ?", id)
	resultsPerlalin := ac.DB.Find(&perlalin, "id_andalalin = ?", id)

	if resultsAndalalin.Error != nil && resultsPerlalin != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": "Tidak ditemukan"})
		return
	}

	if andalalin.IdAndalalin != uuid.Nil {
		for key, files := range form.File {
			for _, file := range files {
				// Save the uploaded file with key as prefix
				file, err := file.Open()

				if err != nil {
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				defer file.Close()

				data, err := io.ReadAll(file)
				if err != nil {
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				switch key {
				case "Kartu tanda penduduk":
					andalalin.KartuTandaPenduduk = data
				case "Akta pendirian badan":
					andalalin.AktaPendirianBadan = data
				case "Surat kuasa":
					andalalin.SuratKuasa = data
				default:
					for i := range andalalin.PersyaratanTambahan {
						if andalalin.PersyaratanTambahan[i].Persyaratan == key {
							andalalin.PersyaratanTambahan[i].Berkas = data
							break
						}
					}
				}
			}
		}

		andalalin.PersyaratanTidakSesuai = nil
		andalalin.StatusAndalalin = "Cek persyaratan"

		ac.DB.Save(&andalalin)
	}

	if perlalin.IdAndalalin != uuid.Nil {
		for key, files := range form.File {
			for _, file := range files {
				// Save the uploaded file with key as prefix
				file, err := file.Open()

				if err != nil {
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				defer file.Close()

				data, err := io.ReadAll(file)
				if err != nil {
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				switch key {
				case "Kartu tanda penduduk":
					perlalin.KartuTandaPenduduk = data
				case "Surat permohonan":
					perlalin.SuratPermohonan = data
				default:
					for i := range andalalin.PersyaratanTambahan {
						if andalalin.PersyaratanTambahan[i].Persyaratan == key {
							andalalin.PersyaratanTambahan[i].Berkas = data
							break
						}
					}
				}
			}
		}

		perlalin.PersyaratanTidakSesuai = nil
		perlalin.StatusAndalalin = "Cek persyaratan"

		ac.DB.Save(&perlalin)
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "msg": "persyaratan berhasil diupdate"})
}

func (ac *AndalalinController) PersyaratanTerpenuhi(ctx *gin.Context) {
	id := ctx.Param("id_andalalin")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.AndalalinStatusCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var andalalin models.Andalalin
	var perlalin models.Perlalin

	resultsAndalalin := ac.DB.First(&andalalin, "id_andalalin = ?", id)
	resultsPerlalin := ac.DB.First(&perlalin, "id_andalalin = ?", id)

	if resultsAndalalin.Error != nil && resultsPerlalin.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": "Tidak ditemukan"})
		return
	}

	if andalalin.IdAndalalin != uuid.Nil {
		andalalin.StatusAndalalin = "Berita acara pemeriksaan"
		ac.DB.Save(&andalalin)
		ac.ReleaseTicketLevel2(ctx, andalalin.IdAndalalin, andalalin.IdAndalalin)
	}

	if perlalin.IdAndalalin != uuid.Nil {
		perlalin.StatusAndalalin = "Persyaratan terpenuhi"
		ac.DB.Save(&perlalin)
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (ac *AndalalinController) PersyaratanTidakSesuai(ctx *gin.Context) {
	id := ctx.Param("id_andalalin")
	var payload *models.PersayaratanTidakSesuaiInput

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.AndalalinStatusCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	var andalalin models.Andalalin
	var perlalin models.Perlalin

	resultsAndalalin := ac.DB.First(&andalalin, "id_andalalin = ?", id)
	resultsPerlalin := ac.DB.First(&perlalin, "id_andalalin = ?", id)

	if resultsAndalalin.Error != nil && resultsPerlalin.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": "Tidak ditemukan"})
		return
	}

	if andalalin.IdAndalalin != uuid.Nil {
		andalalin.StatusAndalalin = "Persyaratan tidak terpenuhi"
		andalalin.PersyaratanTidakSesuai = payload.Persyaratan

		ac.DB.Save(&andalalin)

		justString := strings.Join(payload.Persyaratan, "\n")

		data := utils.PersyaratanTidakSesuai{
			Nomer:       andalalin.Kode,
			Nama:        andalalin.NamaPemohon,
			Alamat:      andalalin.AlamatPemohon,
			Tlp:         andalalin.NomerPemohon,
			Waktu:       andalalin.WaktuAndalalin,
			Izin:        andalalin.JenisAndalalin,
			Status:      andalalin.StatusAndalalin,
			Persyaratan: justString,
			Subject:     "Persyaratan tidak terpenuhi",
		}

		utils.SendEmailPersyaratan(andalalin.EmailPemohon, &data)

		var user models.User
		resultUser := ac.DB.First(&user, "id = ?", andalalin.IdUser)
		if resultUser.Error != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "User tidak ditemukan"})
			return
		}

		simpanNotif := models.Notifikasi{
			IdUser: user.ID,
			Title:  "Persyaratan tidak terpenuhi",
			Body:   "Permohonan anda dengan kode " + andalalin.Kode + " terdapat persyaratan yang tidak sesuai, harap cek email untuk lebih jelas",
		}

		ac.DB.Create(&simpanNotif)

		if user.PushToken != "" {
			notif := utils.Notification{
				IdUser: user.ID,
				Title:  "Persyaratan tidak terpenuhi",
				Body:   "Permohonan anda dengan kode " + andalalin.Kode + " terdapat persyaratan yang tidak sesuai, harap cek email untuk lebih jelas",
				Token:  user.PushToken,
			}

			utils.SendPushNotifications(&notif)
		}

	}

	if perlalin.IdAndalalin != uuid.Nil {
		perlalin.StatusAndalalin = "Persyaratan tidak terpenuhi"
		perlalin.PersyaratanTidakSesuai = payload.Persyaratan

		ac.DB.Save(&perlalin)

		justString := strings.Join(payload.Persyaratan, "\n")

		data := utils.PersyaratanTidakSesuai{
			Nomer:       perlalin.Kode,
			Nama:        perlalin.NamaPemohon,
			Alamat:      perlalin.AlamatPemohon,
			Tlp:         perlalin.NomerPemohon,
			Waktu:       perlalin.WaktuAndalalin,
			Izin:        perlalin.JenisAndalalin,
			Status:      perlalin.StatusAndalalin,
			Persyaratan: justString,
			Subject:     "Persyaratan tidak terpenuhi",
		}

		utils.SendEmailPersyaratan(perlalin.EmailPemohon, &data)

		var user models.User
		resultUser := ac.DB.First(&user, "id = ?", perlalin.IdUser)
		if resultUser.Error != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "User tidak ditemukan"})
			return
		}

		simpanNotif := models.Notifikasi{
			IdUser: user.ID,
			Title:  "Persyaratan tidak terpenuhi",
			Body:   "Permohonan anda dengan kode " + perlalin.Kode + " terdapat persyaratan yang tidak sesuai, harap cek email untuk lebih jelas",
		}

		ac.DB.Create(&simpanNotif)

		if user.PushToken != "" {
			notif := utils.Notification{
				IdUser: user.ID,
				Title:  "Persyaratan tidak terpenuhi",
				Body:   "Permohonan anda dengan kode " + perlalin.Kode + " terdapat persyaratan yang tidak sesuai, harap cek email untuk lebih jelas",
				Token:  user.PushToken,
			}

			utils.SendPushNotifications(&notif)
		}

	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (ac *AndalalinController) UpdateStatusPermohonan(ctx *gin.Context) {
	status := ctx.Param("status")
	id := ctx.Param("id_andalalin")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.AndalalinStatusCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var andalalin models.Andalalin
	var perlalin models.Perlalin

	resultsAndalalin := ac.DB.First(&andalalin, "id_andalalin = ?", id)
	resultsPerlalin := ac.DB.First(&perlalin, "id_andalalin = ?", id)

	if resultsAndalalin.Error != nil && resultsPerlalin.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": "Tidak ditemukan"})
		return
	}

	if andalalin.IdAndalalin != uuid.Nil {
		andalalin.StatusAndalalin = status
		ac.DB.Save(&andalalin)

	}

	if perlalin.IdAndalalin != uuid.Nil {
		perlalin.StatusAndalalin = status
		ac.DB.Save(&perlalin)
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (ac *AndalalinController) TambahPetugas(ctx *gin.Context) {
	var payload *models.TambahPetugas
	id := ctx.Param("id_andalalin")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.AndalalinAddOfficerCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	var perlalin models.Perlalin

	resultsPerlalin := ac.DB.First(&perlalin, "id_andalalin = ?", id)

	if resultsPerlalin.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": "Tidak ditemukan"})
		return
	}

	if perlalin.IdAndalalin != uuid.Nil {
		perlalin.IdPetugas = payload.IdPetugas
		perlalin.NamaPetugas = payload.NamaPetugas
		perlalin.EmailPetugas = payload.EmailPetugas
		perlalin.StatusAndalalin = "Survei lapangan"

		ac.DB.Save(&perlalin)

		ac.ReleaseTicketLevel2(ctx, perlalin.IdAndalalin, payload.IdPetugas)
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Tambah petugas berhasil"})
}

func (ac *AndalalinController) GantiPetugas(ctx *gin.Context) {
	var payload *models.TambahPetugas
	id := ctx.Param("id_andalalin")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.AndalalinAddOfficerCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	var perlalin models.Perlalin

	resultsPerlalin := ac.DB.First(&perlalin, "id_andalalin = ?", id)

	if resultsPerlalin.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": "Tidak ditemukan"})
		return
	}

	if perlalin.IdAndalalin != uuid.Nil {
		perlalin.IdPetugas = payload.IdPetugas
		perlalin.NamaPetugas = payload.NamaPetugas
		perlalin.EmailPetugas = payload.EmailPetugas
		if perlalin.StatusAndalalin == "Survei lapangan" {
			ac.CloseTiketLevel2(ctx, perlalin.IdAndalalin)

			ac.ReleaseTicketLevel2(ctx, perlalin.IdAndalalin, payload.IdPetugas)
		}

		ac.DB.Save(&perlalin)

	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Ubah petugas berhasil"})
}

func (ac *AndalalinController) GetAndalalinTicketLevel2(ctx *gin.Context) {
	status := ctx.Param("status")
	currentUser := ctx.MustGet("currentUser").(models.User)

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.AndalalinTicket2Credential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var ticket []models.TiketLevel2

	results := ac.DB.Find(&ticket, "status = ? AND id_petugas = ?", status, currentUser.ID)

	if results.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": results.Error})
		return
	} else {
		var respone []models.DaftarAndalalinResponse
		for _, s := range ticket {
			var andalalin models.Andalalin
			var perlalin models.Perlalin
			var usulan models.UsulanPengelolaan

			resultsAndalalin := ac.DB.First(&andalalin, "id_andalalin = ? AND id_petugas = ?", s.IdAndalalin, currentUser.ID)
			resultsPerlalin := ac.DB.First(&perlalin, "id_andalalin = ? AND id_petugas = ?", s.IdAndalalin, currentUser.ID)

			if resultsAndalalin.Error != nil && resultsPerlalin.Error != nil {
				ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": "Tidak ditemukan"})
				return
			}

			ac.DB.First(&usulan, "id_andalalin = ?", s.IdAndalalin)

			if usulan.IdUsulan == uuid.Nil {
				if andalalin.IdAndalalin != uuid.Nil {
					respone = append(respone, models.DaftarAndalalinResponse{
						IdAndalalin:      andalalin.IdAndalalin,
						Kode:             andalalin.Kode,
						TanggalAndalalin: andalalin.TanggalAndalalin,
						Nama:             andalalin.NamaPemohon,
						Alamat:           andalalin.AlamatPemohon,
						JenisAndalalin:   andalalin.JenisAndalalin,
						StatusAndalalin:  andalalin.StatusAndalalin,
					})
				}

				if perlalin.IdAndalalin != uuid.Nil {
					respone = append(respone, models.DaftarAndalalinResponse{
						IdAndalalin:      perlalin.IdAndalalin,
						Kode:             perlalin.Kode,
						TanggalAndalalin: perlalin.TanggalAndalalin,
						Nama:             perlalin.NamaPemohon,
						Alamat:           perlalin.AlamatPemohon,
						JenisAndalalin:   perlalin.JenisAndalalin,
						StatusAndalalin:  perlalin.StatusAndalalin,
					})
				}
			}
		}
		ctx.JSON(http.StatusOK, gin.H{"status": "success", "results": len(respone), "data": respone})
	}
}

func (ac *AndalalinController) IsiSurvey(ctx *gin.Context) {
	var payload *models.DataSurvey
	currentUser := ctx.MustGet("currentUser").(models.User)
	id := ctx.Param("id_andalalin")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.AndalalinOfficerCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	if err := ctx.ShouldBind(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	loc, _ := time.LoadLocation("Asia/Singapore")
	nowTime := time.Now().In(loc)

	tanggal := nowTime.Format("02") + " " + utils.Bulan(nowTime.Month()) + " " + nowTime.Format("2006")

	var ticket1 models.TiketLevel1
	var ticket2 models.TiketLevel2

	resultTiket1 := ac.DB.Find(&ticket1, "id_andalalin = ?", id)
	resultTiket2 := ac.DB.Find(&ticket2, "id_andalalin = ?", id)
	if resultTiket1.Error != nil && resultTiket2.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": "Tiket tidak ditemukan"})
		return
	}

	var perlalin models.Perlalin
	resultsPerlalin := ac.DB.First(&perlalin, "id_andalalin = ?", id)

	if resultsPerlalin.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": "Tidak ditemukan"})
		return
	}

	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	blobs := make(map[string][]byte)

	for key, files := range form.File {
		for _, file := range files {
			// Save the uploaded file with key as prefix
			file, err := file.Open()

			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			defer file.Close()

			data, err := io.ReadAll(file)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			// Store the blob data in the map
			blobs[key] = data
		}
	}

	if perlalin.IdAndalalin != uuid.Nil {
		survey := models.Survei{
			IdAndalalin:   perlalin.IdAndalalin,
			IdTiketLevel1: ticket1.IdTiketLevel1,
			IdTiketLevel2: ticket2.IdTiketLevel2,
			IdPetugas:     currentUser.ID,
			Petugas:       currentUser.Name,
			EmailPetugas:  currentUser.Email,
			Lokasi:        payload.Data.Lokasi,
			Keterangan:    payload.Data.Keterangan,
			Foto1:         blobs["foto1"],
			Foto2:         blobs["foto2"],
			Foto3:         blobs["foto3"],
			Latitude:      payload.Data.Latitude,
			Longitude:     payload.Data.Longitude,
			TanggalSurvei: tanggal,
			WaktuSurvei:   nowTime.Format("15:04:05"),
		}

		result := ac.DB.Create(&survey)

		if result.Error != nil && strings.Contains(result.Error.Error(), "duplicate key value violates unique") {
			ctx.JSON(http.StatusConflict, gin.H{"status": "fail", "message": "Data survey sudah tersedia"})
			return
		} else if result.Error != nil {
			ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": "Telah terjadi sesuatu"})
			return
		}

		perlalin.StatusAndalalin = "Laporan survei"

		ac.DB.Save(&perlalin)

		ac.CloseTiketLevel2(ctx, perlalin.IdAndalalin)
	}

	ctx.JSON(http.StatusCreated, gin.H{"status": "success"})
}

func (ac *AndalalinController) GetAllSurvey(ctx *gin.Context) {
	currentUser := ctx.MustGet("currentUser").(models.User)
	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.AndalalinSurveyCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var survey []models.Survei

	results := ac.DB.Find(&survey, "id_petugas = ?", currentUser.ID)

	if results.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": results.Error})
		return
	} else {
		var respone []models.DaftarAndalalinResponse
		for _, s := range survey {
			var perlalin models.Perlalin

			ac.DB.First(&perlalin, "id_andalalin = ?", s.IdAndalalin)

			if perlalin.IdAndalalin != uuid.Nil {
				respone = append(respone, models.DaftarAndalalinResponse{
					IdAndalalin:      perlalin.IdAndalalin,
					Kode:             perlalin.Kode,
					TanggalAndalalin: perlalin.TanggalAndalalin,
					Nama:             perlalin.NamaPemohon,
					Alamat:           perlalin.AlamatPemohon,
					JenisAndalalin:   perlalin.JenisAndalalin,
					StatusAndalalin:  perlalin.StatusAndalalin,
				})
			}

		}
		ctx.JSON(http.StatusOK, gin.H{"status": "success", "results": len(respone), "data": respone})
	}
}

func (ac *AndalalinController) GetSurvey(ctx *gin.Context) {
	id := ctx.Param("id_andalalin")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.AndalalinSurveyCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var survey *models.Survei

	result := ac.DB.First(&survey, "id_andalalin = ?", id)
	if result.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": result.Error})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"status": "success", "data": survey})
}

func (ac *AndalalinController) LaporanBAP(ctx *gin.Context) {
	var payload *models.BAPData
	id := ctx.Param("id_andalalin")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.AndalalinBAPCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	if err := ctx.ShouldBind(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	file, err := ctx.FormFile("bap")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uploadedFile, err := file.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer uploadedFile.Close()

	data, err := io.ReadAll(uploadedFile)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var andalalin models.Andalalin

	resultAndalalin := ac.DB.First(&andalalin, "id_andalalin = ?", id)
	if resultAndalalin.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultAndalalin.Error})
		return
	}

	andalalin.NomerBAPDasar = payload.Data.NomerBAPDasar
	andalalin.NomerBAPPelaksanaan = payload.Data.NomerBAPPelaksanaan
	andalalin.TanggalBAP = payload.Data.TanggalBAP
	andalalin.FileBAP = data
	andalalin.StatusAndalalin = "Persetujuan dokumen"

	result := ac.DB.Save(&andalalin)

	if result.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": result.Error})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"status": "success"})
}

func (ac *AndalalinController) PersetujuanDokumen(ctx *gin.Context) {
	var payload *models.Persetujuan
	id := ctx.Param("id_andalalin")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.AndalalinPersetujuanCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	var andalalin models.Andalalin

	result := ac.DB.First(&andalalin, "id_andalalin = ?", id)
	if result.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": result.Error})
		return
	}

	andalalin.PersetujuanDokumen = payload.Persetujuan
	andalalin.KeteranganPersetujuanDokumen = payload.Keterangan
	if payload.Persetujuan == "Dokumen disetujui" {
		andalalin.StatusAndalalin = "Pembuatan surat keputusan"
		ac.CloseTiketLevel2(ctx, andalalin.IdAndalalin)
	} else {
		andalalin.StatusAndalalin = "Berita acara pemeriksaan"
	}

	ac.DB.Save(&andalalin)

	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (ac *AndalalinController) LaporanSK(ctx *gin.Context) {
	id := ctx.Param("id_andalalin")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.AndalalinSKCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var andalalin models.Andalalin

	result := ac.DB.First(&andalalin, "id_andalalin = ?", id)
	if result.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": result.Error})
		return
	}

	file, err := ctx.FormFile("sk")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uploadedFile, err := file.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer uploadedFile.Close()

	data, err := io.ReadAll(uploadedFile)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	andalalin.FileSK = data

	resultSK := ac.DB.Save(&andalalin)

	if resultSK.Error != nil && strings.Contains(resultSK.Error.Error(), "duplicate key value violates unique") {
		ctx.JSON(http.StatusConflict, gin.H{"status": "fail", "message": "Data SK sudah tersedia"})
		return
	} else if resultSK.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": "Telah terjadi sesuatu"})
		return
	}

	ac.CloseTiketLevel1(ctx, andalalin.IdAndalalin)

	ac.PermohonanSelesai(ctx, andalalin.IdAndalalin)

	ctx.JSON(http.StatusCreated, gin.H{"status": "success"})
}

func (ac *AndalalinController) PermohonanSelesai(ctx *gin.Context, id uuid.UUID) {
	var andalalin models.Andalalin

	result := ac.DB.First(&andalalin, "id_andalalin = ?", id)
	if result.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": result.Error})
		return
	}

	andalalin.StatusAndalalin = "Permohonan selesai"

	ac.DB.Save(&andalalin)

	data := utils.PermohonanSelesai{
		Nomer:   andalalin.Kode,
		Nama:    andalalin.NamaPemohon,
		Alamat:  andalalin.AlamatPemohon,
		Tlp:     andalalin.NomerPemohon,
		Waktu:   andalalin.WaktuAndalalin,
		Izin:    andalalin.JenisAndalalin,
		Status:  andalalin.StatusAndalalin,
		Subject: "Permohonan telah selesai",
	}

	utils.SendEmailPermohonanSelesai(andalalin.EmailPemohon, &data)

	var user models.User
	resultUser := ac.DB.First(&user, "id = ?", andalalin.IdUser)
	if resultUser.Error != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "User tidak ditemukan"})
		return
	}

	simpanNotif := models.Notifikasi{
		IdUser: user.ID,
		Title:  "Permohonan selesai",
		Body:   "Permohonan anda dengan kode " + andalalin.Kode + " telah selesai, harap cek email untuk lebih jelas",
	}

	ac.DB.Create(&simpanNotif)

	if user.PushToken != "" {
		notif := utils.Notification{
			IdUser: user.ID,
			Title:  "Permohonan selesai",
			Body:   "Permohonan anda dengan kode " + andalalin.Kode + " telah selesai, harap cek email untuk lebih jelas",
			Token:  user.PushToken,
		}

		utils.SendPushNotifications(&notif)
	}
}

func (ac *AndalalinController) UsulanTindakanPengelolaan(ctx *gin.Context) {
	var payload *models.InputUsulanPengelolaan
	id := ctx.Param("id_andalalin")
	currentUser := ctx.MustGet("currentUser").(models.User)

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.AndalalinKelolaTiket]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	var ticket1 models.TiketLevel1
	var ticket2 models.TiketLevel2

	resultTiket1 := ac.DB.Find(&ticket1, "id_andalalin = ?", id)
	resultTiket2 := ac.DB.Find(&ticket2, "id_andalalin = ?", id)
	if resultTiket1.Error != nil && resultTiket2.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": "Tiket tidak ditemukan"})
		return
	}

	var perlalin models.Perlalin

	resultsPerlalin := ac.DB.First(&perlalin, "id_andalalin = ?", id)

	if resultsPerlalin.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": "Tidak ditemukan"})
		return
	}

	if perlalin.IdAndalalin != uuid.Nil {
		usul := models.UsulanPengelolaan{
			IdAndalalin:                perlalin.IdAndalalin,
			IdTiketLevel1:              ticket1.IdTiketLevel1,
			IdTiketLevel2:              ticket2.IdTiketLevel2,
			IdPengusulTindakan:         currentUser.ID,
			NamaPengusulTindakan:       currentUser.Name,
			PertimbanganUsulanTindakan: payload.PertimbanganUsulanTindakan,
			KeteranganUsulanTindakan:   payload.KeteranganUsulanTindakan,
		}

		result := ac.DB.Create(&usul)

		if result.Error != nil && strings.Contains(result.Error.Error(), "duplicate key value violates unique") {
			ctx.JSON(http.StatusConflict, gin.H{"status": "fail", "message": "Usulan sudah ada"})
			return
		} else if result.Error != nil {
			ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": "Telah terjadi sesuatu"})
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (ac *AndalalinController) GetUsulan(ctx *gin.Context) {
	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.AndalalinKelolaTiket]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var usulan []models.UsulanPengelolaan

	results := ac.DB.Find(&usulan)

	if results.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": results.Error})
		return
	} else {
		var respone []models.DaftarAndalalinResponse
		for _, s := range usulan {
			var ticket2 models.TiketLevel2
			resultTiket2 := ac.DB.Not("status = ?", "Tunda").Where("id_andalalin = ? AND status = ?", s.IdAndalalin, "Buka").First(&ticket2)
			if resultTiket2.Error == nil {
				var perlalin models.Perlalin

				resultsPerlalin := ac.DB.First(&perlalin, "id_andalalin = ?", ticket2.IdAndalalin)

				if resultsPerlalin.Error != nil {
					ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": "Tidak ditemukan"})
					return
				}

				if perlalin.IdAndalalin != uuid.Nil {
					respone = append(respone, models.DaftarAndalalinResponse{
						IdAndalalin:      perlalin.IdAndalalin,
						Kode:             perlalin.Kode,
						TanggalAndalalin: perlalin.TanggalAndalalin,
						Nama:             perlalin.NamaPemohon,
						Alamat:           perlalin.AlamatPemohon,
						JenisAndalalin:   perlalin.JenisAndalalin,
						StatusAndalalin:  perlalin.StatusAndalalin,
					})
				}

			}
		}
		ctx.JSON(http.StatusOK, gin.H{"status": "success", "results": len(respone), "data": respone})
	}
}

func (ac *AndalalinController) GetDetailUsulan(ctx *gin.Context) {
	id := ctx.Param("id_andalalin")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.AndalalinKelolaTiket]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var usulan models.UsulanPengelolaan

	resultUsulan := ac.DB.First(&usulan, "id_andalalin = ?", id)
	if resultUsulan.Error != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Telah terjadi sesuatu"})
		return
	}

	data := struct {
		NamaPengusulTindakan       string  `json:"nama,omitempty"`
		PertimbanganUsulanTindakan string  `json:"pertimbangan,omitempty"`
		KeteranganUsulanTindakan   *string `json:"keterangan,omitempty"`
	}{
		NamaPengusulTindakan:       usulan.NamaPengusulTindakan,
		PertimbanganUsulanTindakan: usulan.PertimbanganUsulanTindakan,
		KeteranganUsulanTindakan:   usulan.KeteranganUsulanTindakan,
	}

	ctx.JSON(http.StatusCreated, gin.H{"status": "success", "data": data})
}

func (ac *AndalalinController) TindakanPengelolaan(ctx *gin.Context) {
	id := ctx.Param("id_andalalin")
	jenis := ctx.Param("jenis_pelaksanaan")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.AndalalinKelolaTiket]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var tiket models.TiketLevel2

	result := ac.DB.Model(&tiket).Where("id_andalalin = ? AND status = ?", id, "Buka").Or("id_andalalin = ? AND status = ?", id, "Tunda").Update("status", jenis)
	if result.Error != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Telah terjadi sesuatu"})
		return
	}

	var usulan models.UsulanPengelolaan

	ac.DB.First(&usulan, "id_andalalin = ?", id)
	var perlalin models.Perlalin

	resultsPerlalin := ac.DB.First(&perlalin, "id_andalalin = ?", id)

	if resultsPerlalin.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": "Tidak ditemukan"})
		return
	}

	var userPengusul models.User
	resulPengusul := ac.DB.First(&userPengusul, "id = ?", usulan.IdPengusulTindakan)
	if resulPengusul.Error != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "User tidak ditemukan"})
		return
	}

	if perlalin.IdAndalalin != uuid.Nil {
		var userPetugas models.User
		resultPetugas := ac.DB.First(&userPetugas, "id = ?", perlalin.IdPetugas)
		if resultPetugas.Error != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "User tidak ditemukan"})
			return
		}

		switch jenis {
		case "Tunda":
			simpanNotifPengusul := models.Notifikasi{
				IdUser: userPengusul.ID,
				Title:  "Pelaksanaan survei ditunda",
				Body:   "Usulan tindakan anda pada permohonan dengan kode " + perlalin.Kode + " telah diputuskan bahwa pelaksanaan survei ditunda",
			}

			ac.DB.Create(&simpanNotifPengusul)

			simpanNotifPetugas := models.Notifikasi{
				IdUser: userPetugas.ID,
				Title:  "Pelaksanaan survei ditunda",
				Body:   "Pelakasnaan survei pada permohonan dengan kode " + perlalin.Kode + " dibatalkan",
			}

			ac.DB.Create(&simpanNotifPetugas)

			if userPengusul.PushToken != "" {
				notifPengusul := utils.Notification{
					IdUser: userPengusul.ID,
					Title:  "Pelaksanaan survei ditunda",
					Body:   "Usulan tindakan anda pada permohonan dengan kode " + perlalin.Kode + " telah diputuskan bahwa pelaksanaan survei ditunda",
					Token:  userPengusul.PushToken,
				}

				utils.SendPushNotifications(&notifPengusul)
			}

			if userPetugas.PushToken != "" {
				notifPetugas := utils.Notification{
					IdUser: userPetugas.ID,
					Title:  "Pelaksanaan survei ditunda",
					Body:   "Pelakasnaan survei pada permohonan dengan kode " + perlalin.Kode + " ditunda",
					Token:  userPetugas.PushToken,
				}

				utils.SendPushNotifications(&notifPetugas)
			}
		case "Batal":
			simpanNotifPengusul := models.Notifikasi{
				IdUser: userPengusul.ID,
				Title:  "Pelaksanaan survei dibatalkan",
				Body:   "Usulan tindakan anda pada permohonan dengan kode " + perlalin.Kode + " telah diputuskan bahwa pelaksanaan survei dibatalkan",
			}

			ac.DB.Create(&simpanNotifPengusul)

			simpanNotifPetugas := models.Notifikasi{
				IdUser: userPetugas.ID,
				Title:  "Pelaksanaan survei dibatalkan",
				Body:   "Pelakasnaan survei pada permohonan dengan kode " + perlalin.Kode + " dibatalkan",
			}

			ac.DB.Create(&simpanNotifPetugas)

			if userPengusul.PushToken != "" {
				notifPengusul := utils.Notification{
					IdUser: userPengusul.ID,
					Title:  "Pelaksanaan survei dibatalkan",
					Body:   "Usulan tindakan anda pada permohonan dengan kode " + perlalin.Kode + " telah diputuskan bahwa pelaksanaan survei dibatalkan",
					Token:  userPengusul.PushToken,
				}

				utils.SendPushNotifications(&notifPengusul)
			}

			if userPetugas.PushToken != "" {
				notifPetugas := utils.Notification{
					IdUser: userPetugas.ID,
					Title:  "Pelaksanaan survei dibatalkan",
					Body:   "Pelakasnaan survei pada permohonan dengan kode " + perlalin.Kode + " dibatalkan",
					Token:  userPetugas.PushToken,
				}

				utils.SendPushNotifications(&notifPetugas)
			}
		case "Buka":
			simpanNotifPetugas := models.Notifikasi{
				IdUser: userPetugas.ID,
				Title:  "Pelaksanaan survei dilanjutkan",
				Body:   "Pelaksanaan survei pada permohonan dengan kode " + perlalin.Kode + " telah dilanjutkan kembali",
			}

			ac.DB.Create(&simpanNotifPetugas)

			if userPetugas.PushToken != "" {
				notifPetugas := utils.Notification{
					IdUser: userPetugas.ID,
					Title:  "Pelaksanaan survei dilanjutkan",
					Body:   "Pelaksanaan survei pada permohonan dengan kode " + perlalin.Kode + " telah dilanjutkan kembali",
					Token:  userPetugas.PushToken,
				}

				utils.SendPushNotifications(&notifPetugas)
			}
		}
	}

	if jenis == "Batal" || jenis == "Buka" {
		ac.DB.Delete(&models.UsulanPengelolaan{}, "id_andalalin = ?", id)
	}

	ctx.JSON(http.StatusCreated, gin.H{"status": "success"})
}

func (ac *AndalalinController) HapusUsulan(ctx *gin.Context) {
	id := ctx.Param("id_andalalin")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.AndalalinKelolaTiket]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var usulan models.UsulanPengelolaan

	resultUsulan := ac.DB.First(&usulan, "id_andalalin = ?", id)
	if resultUsulan.Error != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "User tidak ditemukan"})
		return
	}

	var perlalin models.Perlalin

	resultsPerlalin := ac.DB.First(&perlalin, "id_andalalin = ?", id)

	if resultsPerlalin.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": "Tidak ditemukan"})
		return
	}

	var userPengusul models.User
	resulPengusul := ac.DB.First(&userPengusul, "id = ?", usulan.IdPengusulTindakan)
	if resulPengusul.Error != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "User tidak ditemukan"})
		return
	}

	if perlalin.IdAndalalin != uuid.Nil {
		simpanNotifPengusul := models.Notifikasi{
			IdUser: userPengusul.ID,
			Title:  "Usulan tindakan dihapus",
			Body:   "Usulan tindakan anda pada permohonan dengan kode " + perlalin.Kode + " telah dihapus",
		}

		ac.DB.Create(&simpanNotifPengusul)

		if userPengusul.PushToken != "" {
			notifPengusul := utils.Notification{
				IdUser: userPengusul.ID,
				Title:  "Usulan tindakan dihapus",
				Body:   "Usulan tindakan anda pada permohonan dengan kode " + perlalin.Kode + " telah dihapus",
				Token:  userPengusul.PushToken,
			}

			utils.SendPushNotifications(&notifPengusul)
		}
	}

	results := ac.DB.Delete(&models.UsulanPengelolaan{}, "id_andalalin = ?", id)

	if results.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": results.Error})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"status": "success"})
}

func (ac *AndalalinController) GetAllAndalalinByTiketLevel2(ctx *gin.Context) {
	status := ctx.Param("status")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.AndalalinTicket2Credential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var ticket []models.TiketLevel2

	results := ac.DB.Find(&ticket, "status = ?", status)

	if results.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": results.Error})
		return
	} else {
		var respone []models.DaftarAndalalinResponse
		for _, s := range ticket {
			var perlalin models.Perlalin

			resultsPerlalin := ac.DB.First(&perlalin, "id_andalalin = ?", s.IdAndalalin)

			if resultsPerlalin.Error != nil {
				ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": "Tidak ditemukan"})
				return
			}

			if perlalin.IdAndalalin != uuid.Nil {
				respone = append(respone, models.DaftarAndalalinResponse{
					IdAndalalin:      perlalin.IdAndalalin,
					Kode:             perlalin.Kode,
					TanggalAndalalin: perlalin.TanggalAndalalin,
					Nama:             perlalin.NamaPemohon,
					Alamat:           perlalin.AlamatPemohon,
					JenisAndalalin:   perlalin.JenisAndalalin,
					StatusAndalalin:  perlalin.StatusAndalalin,
				})
			}

		}
		ctx.JSON(http.StatusOK, gin.H{"status": "success", "results": len(respone), "data": respone})
	}
}

func (ac *AndalalinController) LaporanSurvei(ctx *gin.Context) {
	id := ctx.Param("id_andalalin")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.AndalalinSurveyCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var perlalin models.Perlalin

	result := ac.DB.First(&perlalin, "id_andalalin = ?", id)
	if result.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": result.Error})
		return
	}

	file, err := ctx.FormFile("ls")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uploadedFile, err := file.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer uploadedFile.Close()

	data, err := io.ReadAll(uploadedFile)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	perlalin.LaporanSurvei = data
	perlalin.StatusAndalalin = "Menunggu hasil keputusan"

	resultLaporan := ac.DB.Save(&perlalin)

	if resultLaporan.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": "Telah terjadi sesuatu"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"status": "success"})
}

func (ac *AndalalinController) IsiSurveyMandiri(ctx *gin.Context) {
	var payload *models.DataSurvey
	currentUser := ctx.MustGet("currentUser").(models.User)

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.AndalalinOfficerCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	if err := ctx.ShouldBind(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	loc, _ := time.LoadLocation("Asia/Singapore")
	nowTime := time.Now().In(loc)

	tanggal := nowTime.Format("02") + " " + utils.Bulan(nowTime.Month()) + " " + nowTime.Format("2006")

	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	blobs := make(map[string][]byte)

	for key, files := range form.File {
		for _, file := range files {
			// Save the uploaded file with key as prefix
			file, err := file.Open()

			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			defer file.Close()

			data, err := io.ReadAll(file)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			// Store the blob data in the map
			blobs[key] = data
		}
	}

	survey := models.SurveiMandiri{
		IdPetugas:     currentUser.ID,
		Petugas:       currentUser.Name,
		EmailPetugas:  currentUser.Email,
		Lokasi:        payload.Data.Lokasi,
		Keterangan:    payload.Data.Keterangan,
		Foto1:         blobs["foto1"],
		Foto2:         blobs["foto2"],
		Foto3:         blobs["foto3"],
		Latitude:      payload.Data.Latitude,
		Longitude:     payload.Data.Longitude,
		StatusSurvei:  "Perlu tindakan",
		TanggalSurvei: tanggal,
		WaktuSurvei:   nowTime.Format("15:04:05"),
	}

	result := ac.DB.Create(&survey)

	if result.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": "Telah terjadi sesuatu"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"status": "success", "data": survey})
}

func (ac *AndalalinController) GetAllSurveiMandiri(ctx *gin.Context) {
	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.AndalalinSurveyCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var survey []models.SurveiMandiri

	results := ac.DB.Find(&survey)

	if results.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": results.Error})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "success", "results": len(survey), "data": survey})
}

func (ac *AndalalinController) GetAllSurveiMandiriByPetugas(ctx *gin.Context) {
	currentUser := ctx.MustGet("currentUser").(models.User)
	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.AndalalinSurveyCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var survey []models.SurveiMandiri

	results := ac.DB.Find(&survey, "id_petugas = ?", currentUser.ID)

	if results.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": results.Error})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "success", "results": len(survey), "data": survey})
}

func (ac *AndalalinController) GetSurveiMandiri(ctx *gin.Context) {
	id := ctx.Param("id_survei")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.AndalalinSurveyCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var survey *models.SurveiMandiri

	result := ac.DB.First(&survey, "id_survey = ?", id)
	if result.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": result.Error})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"status": "success", "data": survey})
}

func (ac *AndalalinController) TerimaSurvei(ctx *gin.Context) {
	id := ctx.Param("id_survei")
	keterangan := ctx.Param("keterangan")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.AndalalinSurveyCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var survey *models.SurveiMandiri

	result := ac.DB.First(&survey, "id_survey = ?", id)
	if result.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": result.Error})
		return
	}
	survey.StatusSurvei = "Survei diterima"
	survey.KeteranganTindakan = keterangan

	ac.DB.Save(&survey)

	ctx.JSON(http.StatusCreated, gin.H{"status": "success", "data": survey})
}

func (ac *AndalalinController) KeputusanHasil(ctx *gin.Context) {
	id := ctx.Param("id_andalalin")
	var payload *models.KeputusanHasil

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.AndalalinKeputusanHasil]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var perlalin models.Perlalin

	result := ac.DB.First(&perlalin, "id_andalalin = ?", id)
	if result.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": result.Error})
		return
	}

	if payload.Keputusan == "Pemasangan ditunda" {
		perlalin.Tindakan = payload.Keputusan
		perlalin.PertimbanganTindakan = payload.Pertimbangan
		perlalin.StatusAndalalin = "Tunda pemasangan"
	} else if payload.Keputusan == "Pemasangan disegerakan" {
		perlalin.Tindakan = payload.Keputusan
		perlalin.StatusAndalalin = "Pemasangan sedang dilakukan"
	}

	resultKeputusan := ac.DB.Save(&perlalin)

	if resultKeputusan.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": "Telah terjadi sesuatu"})
		return
	}

	var mutex sync.Mutex

	updateChannelTunda := make(chan struct{})
	updateChannelDisegerakan := make(chan struct{})

	if payload.Keputusan == "Pemasangan ditunda" {
		go func() {
			duration := 3 * 24 * time.Hour
			timer := time.NewTimer(duration)

			select {
			case <-timer.C:
				mutex.Lock()
				defer mutex.Unlock()

				var data models.Perlalin

				result := ac.DB.First(&data, "id_andalalin = ?", id)
				if result.Error != nil {
					ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": result.Error})
					return
				}

				if data.StatusAndalalin == "Tunda pemasangan" {
					ac.CloseTiketLevel1(ctx, data.IdAndalalin)
					ac.BatalkanPermohonan(ctx, id)
					data.Tindakan = "Permohonan dibatalkan"
					data.PertimbanganTindakan = "Permohonan dibatalkan"
					data.StatusAndalalin = "Permohonan dibatalkan"
					ac.DB.Save(&data)
					updateChannelTunda <- struct{}{}
				}
			case <-updateChannelTunda:
				// The update was canceled, do nothing
			}
		}()
	} else if payload.Keputusan == "Pemasangan disegerakan" {
		if perlalin.StatusAndalalin == "Tunda pemasangan" {
			close(updateChannelTunda)
		}

		go func() {
			duration := 3 * 24 * time.Hour
			timer := time.NewTimer(duration)

			select {
			case <-timer.C:
				mutex.Lock()
				defer mutex.Unlock()

				var data models.Perlalin

				result := ac.DB.First(&data, "id_andalalin = ?", id)
				if result.Error != nil {
					ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": result.Error})
					return
				}

				if data.StatusAndalalin == "Pemasangan sedang dilakukan" {
					data.Tindakan = "Pemasangan ditunda"
					data.PertimbanganTindakan = "Pemasangan ditunda"
					data.StatusAndalalin = "Tunda pemasangan"
					ac.DB.Save(&data)

					updateChannelTunda = make(chan struct{})

					go func() {
						duration := 3 * 24 * time.Hour
						timer := time.NewTimer(duration)

						select {
						case <-timer.C:
							mutex.Lock()
							defer mutex.Unlock()

							var data models.Perlalin

							result := ac.DB.First(&data, "id_andalalin = ?", id)
							if result.Error != nil {
								ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": result.Error})
								return
							}

							if data.StatusAndalalin == "Tunda pemasangan" {
								ac.CloseTiketLevel1(ctx, data.IdAndalalin)
								ac.BatalkanPermohonan(ctx, id)
								updateChannelTunda <- struct{}{}
								updateChannelDisegerakan <- struct{}{}
							}

						case <-updateChannelTunda:
							// The update was canceled, do nothing
						}
					}()
				}
			case <-updateChannelDisegerakan:
				// The update was canceled, do nothing
			}
		}()
	} else if payload.Keputusan == "Batalkan permohonan" {
		if perlalin.StatusAndalalin == "Tunda pemasangan" {
			close(updateChannelTunda)
		}

		ac.CloseTiketLevel1(ctx, perlalin.IdAndalalin)
		ac.BatalkanPermohonan(ctx, id)
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (ac *AndalalinController) BatalkanPermohonan(ctx *gin.Context, id string) {
	var permohonan models.Perlalin

	result := ac.DB.First(&permohonan, "id_andalalin = ?", id)
	if result.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": result.Error})
		return
	}

	permohonan.Tindakan = "Permohonan dibatalkan"
	permohonan.PertimbanganTindakan = "Permohonan dibatalkan"
	permohonan.StatusAndalalin = "Permohonan dibatalkan"
	ac.DB.Save(&permohonan)

	data := utils.PermohonanSelesai{
		Nomer:   permohonan.Kode,
		Nama:    permohonan.NamaPemohon,
		Alamat:  permohonan.AlamatPemohon,
		Tlp:     permohonan.NomerPemohon,
		Waktu:   permohonan.WaktuAndalalin,
		Izin:    permohonan.JenisAndalalin,
		Status:  permohonan.StatusAndalalin,
		Subject: "Permohonan dibatalkan",
	}

	utils.SendEmailPermohonanDibatalkan(permohonan.EmailPemohon, &data)

	var user models.User
	resultUser := ac.DB.First(&user, "id = ?", permohonan.IdUser)
	if resultUser.Error != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "User tidak ditemukan"})
		return
	}

	simpanNotif := models.Notifikasi{
		IdUser: user.ID,
		Title:  "Permohonan dibatalkan",
		Body:   "Permohonan anda dengan kode " + permohonan.Kode + " telah dibatalkan, harap cek email untuk lebih jelas",
	}

	ac.DB.Create(&simpanNotif)

	if user.PushToken != "" {
		notif := utils.Notification{
			IdUser: user.ID,
			Title:  "Permohonan dibatalkan",
			Body:   "Permohonan anda dengan kode " + permohonan.Kode + " telah dibatalkan, harap cek email untuk lebih jelas",
			Token:  user.PushToken,
		}

		utils.SendPushNotifications(&notif)
	}
}

func (ac *AndalalinController) SurveiKepuasan(ctx *gin.Context) {
	var payload *models.SurveiKepuasanInput
	id := ctx.Param("id_andalalin")
	currentUser := ctx.MustGet("currentUser").(models.User)

	if err := ctx.ShouldBind(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	loc, _ := time.LoadLocation("Asia/Singapore")
	nowTime := time.Now().In(loc)

	tanggal := nowTime.Format("02") + " " + utils.Bulan(nowTime.Month()) + " " + nowTime.Format("2006")

	var andalalin models.Perlalin
	var perlalin models.Andalalin

	resultsAndalalin := ac.DB.First(&andalalin, "id_andalalin = ?", id)
	resultsPerlalin := ac.DB.First(&perlalin, "id_andalalin = ?", id)

	if resultsAndalalin.Error != nil && resultsPerlalin.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": "Tidak ditemukan"})
		return
	}

	if andalalin.IdAndalalin != uuid.Nil {
		kepuasan := models.SurveiKepuasan{
			IdAndalalin:        andalalin.IdAndalalin,
			IdUser:             currentUser.ID,
			Nama:               currentUser.Name,
			Email:              currentUser.Email,
			KritikSaran:        payload.KritikSaran,
			TanggalPelaksanaan: tanggal,
			DataSurvei:         payload.DataSurvei,
		}

		result := ac.DB.Create(&kepuasan)

		if result.Error != nil {
			ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": "Telah terjadi sesuatu"})
			return
		}
	}

	if perlalin.IdAndalalin != uuid.Nil {
		kepuasan := models.SurveiKepuasan{
			IdAndalalin:        perlalin.IdAndalalin,
			IdUser:             currentUser.ID,
			Nama:               currentUser.Name,
			Email:              currentUser.Email,
			KritikSaran:        payload.KritikSaran,
			TanggalPelaksanaan: tanggal,
			DataSurvei:         payload.DataSurvei,
		}

		result := ac.DB.Create(&kepuasan)

		if result.Error != nil {
			ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": "Telah terjadi sesuatu"})
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (ac *AndalalinController) CekSurveiKepuasan(ctx *gin.Context) {
	id := ctx.Param("id_andalalin")

	var survei models.SurveiKepuasan

	result := ac.DB.First(&survei, "id_andalalin", id)

	if result.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": "Telah terjadi sesuatu"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (ac *AndalalinController) HasilSurveiKepuasan(ctx *gin.Context) {
	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.AndalalinSurveiKepuasan]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	loc, _ := time.LoadLocation("Asia/Singapore")
	nowTime := time.Now().In(loc)

	startOfMonth := getStartOfMonth(nowTime.Year(), nowTime.Month())

	endOfMonth := getEndOfMonth(nowTime.Year(), nowTime.Month())

	periode := startOfMonth.Format("02") + " - " + endOfMonth.Format("02") + " " + utils.Bulan(nowTime.Month()) + " " + nowTime.Format("2006")

	var survei []models.SurveiKepuasan

	result := ac.DB.Where("tanggal_pelaksanaan LIKE ?", fmt.Sprintf("%%%s%%", utils.Bulan(nowTime.Month()))).Find(&survei)

	if result.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": "Telah terjadi sesuatu"})
		return
	}

	nilai := []data{}

	nilai = append(nilai, data{Jenis: "Persyaratan pelayanan", Nilai: 0, Hasil: "0"})
	nilai = append(nilai, data{Jenis: "Prosedur pelayanan", Nilai: 0, Hasil: "0"})
	nilai = append(nilai, data{Jenis: "Waktu pelayanan", Nilai: 0, Hasil: "0"})
	nilai = append(nilai, data{Jenis: "Biaya / tarif pelayanan", Nilai: 0, Hasil: "0"})
	nilai = append(nilai, data{Jenis: "Produk pelayanan", Nilai: 0, Hasil: "0"})
	nilai = append(nilai, data{Jenis: "Kompetensi pelaksana", Nilai: 0, Hasil: "0"})
	nilai = append(nilai, data{Jenis: "Perilaku / sikap petugas", Nilai: 0, Hasil: "0"})
	nilai = append(nilai, data{Jenis: "Maklumat pelayanan", Nilai: 0, Hasil: "0"})
	nilai = append(nilai, data{Jenis: "Ketersediaan sarana pengaduan", Nilai: 0, Hasil: "0"})

	komen := []komentar{}

	for _, data := range survei {
		komen = append(komen, komentar{Nama: data.Nama, Komentar: *data.KritikSaran})
		for _, isi := range data.DataSurvei {
			for i, item := range nilai {
				if item.Jenis == isi.Jenis {
					switch isi.Nilai {
					case "Sangat baik":
						nilai[i].Nilai = nilai[i].Nilai + 4
					case "Baik":
						nilai[i].Nilai = nilai[i].Nilai + 3
					case "Kurang baik":
						nilai[i].Nilai = nilai[i].Nilai + 2
					case "Buruk":
						nilai[i].Nilai = nilai[i].Nilai + 1
					}
					break
				}
			}
		}
	}

	total := 0

	for i, item := range nilai {
		hasil := float64(item.Nilai) * float64(100) / float64(len(survei)) / float64(4)
		nilai[i].Hasil = fmt.Sprintf("%.2f", hasil)
		total = total + item.Nilai
	}

	indeksHasil := float64(total) * float64(100) / float64(9) / float64(4) / float64(len(survei))
	indeks := fmt.Sprintf("%.2f", indeksHasil)

	hasil := struct {
		Periode        string     `json:"periode,omitempty"`
		Responden      string     `json:"responden,omitempty"`
		IndeksKepuasan string     `json:"indeks_kepuasan,omitempty"`
		NilaiInterval  string     `json:"nilai_interval,omitempty"`
		Mutu           string     `json:"mutu,omitempty"`
		Kinerja        string     `json:"kinerja,omitempty"`
		DataHasil      []data     `json:"hasil,omitempty"`
		Komentar       []komentar `json:"komentar,omitempty"`
	}{
		Periode:        periode,
		Responden:      strconv.Itoa(len(survei)),
		IndeksKepuasan: indeks,
		NilaiInterval:  interval(indeksHasil),
		Mutu:           mutu(indeksHasil),
		Kinerja:        kinerja(indeksHasil),
		DataHasil:      nilai,
		Komentar:       komen,
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": hasil})
}

func (ac *AndalalinController) GetPermohonanPemasanganLalin(ctx *gin.Context) {
	currentUser := ctx.MustGet("currentUser").(models.User)

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.AndalalinGetCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var perlalin []models.Perlalin

	ac.DB.Order("tanggal_andalalin").Find(&perlalin)

	var respone []models.DaftarAndalalinResponse
	for _, s := range perlalin {
		if s.StatusAndalalin == "Pemasangan sedang dilakukan" && s.IdPetugas == currentUser.ID {
			respone = append(respone, models.DaftarAndalalinResponse{
				IdAndalalin:      s.IdAndalalin,
				Kode:             s.Kode,
				TanggalAndalalin: s.TanggalAndalalin,
				Nama:             s.NamaPemohon,
				Alamat:           s.AlamatPemohon,
				JenisAndalalin:   s.JenisAndalalin,
				StatusAndalalin:  s.StatusAndalalin,
			})
		}
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "success", "results": len(respone), "data": respone})
}

func (ac *AndalalinController) PemasanganPerlengkapanLaluLintas(ctx *gin.Context) {
	var payload *models.DataSurvey
	currentUser := ctx.MustGet("currentUser").(models.User)
	id := ctx.Param("id_andalalin")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.AndalalinOfficerCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	if err := ctx.ShouldBind(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	loc, _ := time.LoadLocation("Asia/Singapore")
	nowTime := time.Now().In(loc)

	tanggal := nowTime.Format("02") + " " + utils.Bulan(nowTime.Month()) + " " + nowTime.Format("2006")

	var ticket1 models.TiketLevel1

	resultTiket1 := ac.DB.Find(&ticket1, "id_andalalin = ?", id)
	if resultTiket1.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": "Tiket tidak ditemukan"})
		return
	}

	var perlalin models.Perlalin
	resultsPerlalin := ac.DB.First(&perlalin, "id_andalalin = ?", id)

	if resultsPerlalin.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": "Tidak ditemukan"})
		return
	}

	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	blobs := make(map[string][]byte)

	for key, files := range form.File {
		for _, file := range files {
			// Save the uploaded file with key as prefix
			file, err := file.Open()

			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			defer file.Close()

			data, err := io.ReadAll(file)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			// Store the blob data in the map
			blobs[key] = data
		}
	}

	survey := models.Pemasangan{
		IdAndalalin:       perlalin.IdAndalalin,
		IdTiketLevel1:     ticket1.IdTiketLevel1,
		IdPetugas:         currentUser.ID,
		Petugas:           currentUser.Name,
		EmailPetugas:      currentUser.Email,
		Lokasi:            payload.Data.Lokasi,
		Keterangan:        payload.Data.Keterangan,
		Foto1:             blobs["foto1"],
		Foto2:             blobs["foto2"],
		Foto3:             blobs["foto3"],
		Latitude:          payload.Data.Latitude,
		Longitude:         payload.Data.Longitude,
		WaktuPemasangan:   tanggal,
		TanggalPemasangan: nowTime.Format("15:04:05"),
	}

	result := ac.DB.Create(&survey)

	if result.Error != nil && strings.Contains(result.Error.Error(), "duplicate key value violates unique") {
		ctx.JSON(http.StatusConflict, gin.H{"status": "fail", "message": "Data survey sudah tersedia"})
		return
	} else if result.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": "Telah terjadi sesuatu"})
		return
	}

	perlalin.StatusAndalalin = "Pemasangan selesai"

	ac.DB.Save(&perlalin)

	ac.PemasanganSelesai(ctx, perlalin)
	ac.CloseTiketLevel1(ctx, perlalin.IdAndalalin)

	ctx.JSON(http.StatusCreated, gin.H{"status": "success"})
}

func (ac *AndalalinController) PemasanganSelesai(ctx *gin.Context, permohonan models.Perlalin) {
	data := utils.Pemasangan{
		Nomer:   permohonan.Kode,
		Nama:    permohonan.NamaPemohon,
		Alamat:  permohonan.AlamatPemohon,
		Tlp:     permohonan.NomerPemohon,
		Waktu:   permohonan.WaktuAndalalin,
		Izin:    permohonan.JenisAndalalin,
		Status:  permohonan.StatusAndalalin,
		Subject: "Pemasangan selesai",
	}

	utils.SendEmailPemasangan(permohonan.EmailPemohon, &data)

	var user models.User
	resultUser := ac.DB.First(&user, "id = ?", permohonan.IdUser)
	if resultUser.Error != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "User tidak ditemukan"})
		return
	}

	simpanNotif := models.Notifikasi{
		IdUser: user.ID,
		Title:  "Pemasangan selesai",
		Body:   "Permohonan anda dengan kode " + permohonan.Kode + " telah selesai pemasangan perlengkapan lalu lintas, harap cek email untuk lebih jelas",
	}

	ac.DB.Create(&simpanNotif)

	if user.PushToken != "" {
		notif := utils.Notification{
			IdUser: user.ID,
			Title:  "Pemasangan selesai",
			Body:   "Permohonan anda dengan kode " + permohonan.Kode + " telah selesai pemasangan perlengkapan lalu lintas, harap cek email untuk lebih jelas",
			Token:  user.PushToken,
		}

		utils.SendPushNotifications(&notif)
	}
}

func (ac *AndalalinController) GetAllPemasangan(ctx *gin.Context) {
	currentUser := ctx.MustGet("currentUser").(models.User)
	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.AndalalinOfficerCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var pemasangan []models.Pemasangan

	results := ac.DB.Find(&pemasangan, "id_petugas = ?", currentUser.ID)

	if results.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": results.Error})
		return
	} else {
		var respone []models.DaftarAndalalinResponse
		for _, s := range pemasangan {
			var perlalin models.Perlalin

			ac.DB.First(&perlalin, "id_andalalin = ?", s.IdAndalalin)

			if perlalin.IdAndalalin != uuid.Nil {
				respone = append(respone, models.DaftarAndalalinResponse{
					IdAndalalin:      perlalin.IdAndalalin,
					Kode:             perlalin.Kode,
					TanggalAndalalin: perlalin.TanggalAndalalin,
					Nama:             perlalin.NamaPemohon,
					Alamat:           perlalin.AlamatPemohon,
					JenisAndalalin:   perlalin.JenisAndalalin,
					StatusAndalalin:  perlalin.StatusAndalalin,
				})
			}

		}
		ctx.JSON(http.StatusOK, gin.H{"status": "success", "results": len(respone), "data": respone})
	}
}

func (ac *AndalalinController) GetPemasangan(ctx *gin.Context) {
	id := ctx.Param("id_andalalin")

	var pemasangan *models.Pemasangan

	result := ac.DB.First(&pemasangan, "id_andalalin = ?", id)
	if result.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": result.Error})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"status": "success", "data": pemasangan})
}
