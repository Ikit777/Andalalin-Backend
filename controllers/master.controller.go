package controllers

import (
	"archive/zip"
	"encoding/base64"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/Ikit777/Andalalin-Backend/initializers"
	"github.com/Ikit777/Andalalin-Backend/models"
	"github.com/Ikit777/Andalalin-Backend/repository"
	"github.com/Ikit777/Andalalin-Backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DataMasterControler struct {
	DB *gorm.DB
}

type file struct {
	Name string
	File []byte
}

func NewDataMasterControler(DB *gorm.DB) DataMasterControler {
	return DataMasterControler{DB}
}

func (dm *DataMasterControler) GetDataMaster(ctx *gin.Context) {
	var data models.DataMaster

	results := dm.DB.First(&data)

	if results.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": results.Error})
		return
	}

	respone := struct {
		IdDataMaster           uuid.UUID                  `json:"id_data_master,omitempty"`
		Lokasi                 []string                   `json:"lokasi_pengambilan,omitempty"`
		JenisRencana           []string                   `json:"jenis_rencana,omitempty"`
		RencanaPembangunan     []models.Rencana           `json:"rencana_pembangunan,omitempty"`
		KategoriPerlengkapan   []string                   `json:"kategori_perlengkapan,omitempty"`
		PerlengkapanLaluLintas []models.JenisPerlengkapan `json:"perlengkapan,omitempty"`
		PersyaratanTambahan    models.PersyaratanTambahan `json:"persyaratan_tambahan,omitempty"`
	}{
		IdDataMaster:           data.IdDataMaster,
		Lokasi:                 data.LokasiPengambilan,
		JenisRencana:           data.JenisRencanaPembangunan,
		RencanaPembangunan:     data.RencanaPembangunan,
		KategoriPerlengkapan:   data.KategoriPerlengkapan,
		PerlengkapanLaluLintas: data.PerlengkapanLaluLintas,
		PersyaratanTambahan:    data.PersyaratanTambahan,
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": respone})
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func compressFiles(zipFileName string, fileData []file) error {
	zipFile, err := os.Create(zipFileName)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, data := range fileData {
		tmpFile, _ := os.CreateTemp("", "persyaratan.pdf")
		defer os.Remove(tmpFile.Name())

		_, _ = tmpFile.Write(data.File)
		tmpFile.Close()

		file, err := os.Open(tmpFile.Name())
		if err != nil {
			return err
		}
		defer file.Close()

		fileInfo, err := file.Stat()
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(fileInfo)
		if err != nil {
			return err
		}

		header.Name = data.Name

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		_, err = io.Copy(writer, file)
		if err != nil {
			return err
		}
	}

	return nil
}

func (dm *DataMasterControler) TambahLokasi(ctx *gin.Context) {
	lokasi := ctx.Param("lokasi")
	id := ctx.Param("id")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.ProductAddCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var master models.DataMaster

	resultsData := dm.DB.Where("id_data_master", id).First(&master)

	if resultsData.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsData.Error})
		return
	}

	exist := contains(master.LokasiPengambilan, lokasi)

	if exist {
		ctx.JSON(http.StatusConflict, gin.H{"status": "fail", "message": "Data sudah ada"})
		return
	}

	master.LokasiPengambilan = append(master.LokasiPengambilan, lokasi)

	resultsSave := dm.DB.Save(&master)
	if resultsSave.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsSave.Error})
		return
	}

	respone := struct {
		IdDataMaster           uuid.UUID                  `json:"id_data_master,omitempty"`
		Lokasi                 []string                   `json:"lokasi_pengambilan,omitempty"`
		JenisRencana           []string                   `json:"jenis_rencana,omitempty"`
		RencanaPembangunan     []models.Rencana           `json:"rencana_pembangunan,omitempty"`
		KategoriPerlengkapan   []string                   `json:"kategori_perlengkapan,omitempty"`
		PerlengkapanLaluLintas []models.JenisPerlengkapan `json:"perlengkapan,omitempty"`
		PersyaratanTambahan    models.PersyaratanTambahan `json:"persyaratan_tambahan,omitempty"`
	}{
		IdDataMaster:           master.IdDataMaster,
		Lokasi:                 master.LokasiPengambilan,
		JenisRencana:           master.JenisRencanaPembangunan,
		RencanaPembangunan:     master.RencanaPembangunan,
		KategoriPerlengkapan:   master.KategoriPerlengkapan,
		PerlengkapanLaluLintas: master.PerlengkapanLaluLintas,
		PersyaratanTambahan:    master.PersyaratanTambahan,
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": respone})
}

func (dm *DataMasterControler) HapusLokasi(ctx *gin.Context) {
	lokasi := ctx.Param("lokasi")
	id := ctx.Param("id")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.ProductDeleteCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var master models.DataMaster

	resultsData := dm.DB.Where("id_data_master", id).First(&master)

	if resultsData.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsData.Error})
		return
	}

	for i, item := range master.LokasiPengambilan {
		if item == lokasi {
			master.LokasiPengambilan = append(master.LokasiPengambilan[:i], master.LokasiPengambilan[i+1:]...)
			break
		}
	}

	resultsSave := dm.DB.Save(&master)
	if resultsSave.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsSave.Error})
		return
	}

	respone := struct {
		IdDataMaster           uuid.UUID                  `json:"id_data_master,omitempty"`
		Lokasi                 []string                   `json:"lokasi_pengambilan,omitempty"`
		JenisRencana           []string                   `json:"jenis_rencana,omitempty"`
		RencanaPembangunan     []models.Rencana           `json:"rencana_pembangunan,omitempty"`
		KategoriPerlengkapan   []string                   `json:"kategori_perlengkapan,omitempty"`
		PerlengkapanLaluLintas []models.JenisPerlengkapan `json:"perlengkapan,omitempty"`
		PersyaratanTambahan    models.PersyaratanTambahan `json:"persyaratan_tambahan,omitempty"`
	}{
		IdDataMaster:           master.IdDataMaster,
		Lokasi:                 master.LokasiPengambilan,
		JenisRencana:           master.JenisRencanaPembangunan,
		RencanaPembangunan:     master.RencanaPembangunan,
		KategoriPerlengkapan:   master.KategoriPerlengkapan,
		PerlengkapanLaluLintas: master.PerlengkapanLaluLintas,
		PersyaratanTambahan:    master.PersyaratanTambahan,
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": respone})
}

func (dm *DataMasterControler) EditLokasi(ctx *gin.Context) {
	lokasi := ctx.Param("lokasi")
	newLokasi := ctx.Param("new_lokasi")
	id := ctx.Param("id")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.ProductUpdateCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var master models.DataMaster

	resultsData := dm.DB.Where("id_data_master", id).First(&master)

	if resultsData.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsData.Error})
		return
	}

	itemIndex := -1

	for i, item := range master.LokasiPengambilan {
		if item == lokasi {
			itemIndex = i
			break
		}
	}

	if itemIndex != -1 {
		master.LokasiPengambilan[itemIndex] = newLokasi
	}

	resultsSave := dm.DB.Save(&master)
	if resultsSave.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsSave.Error})
		return
	}

	respone := struct {
		IdDataMaster           uuid.UUID                  `json:"id_data_master,omitempty"`
		Lokasi                 []string                   `json:"lokasi_pengambilan,omitempty"`
		JenisRencana           []string                   `json:"jenis_rencana,omitempty"`
		RencanaPembangunan     []models.Rencana           `json:"rencana_pembangunan,omitempty"`
		KategoriPerlengkapan   []string                   `json:"kategori_perlengkapan,omitempty"`
		PerlengkapanLaluLintas []models.JenisPerlengkapan `json:"perlengkapan,omitempty"`
		PersyaratanTambahan    models.PersyaratanTambahan `json:"persyaratan_tambahan,omitempty"`
	}{
		IdDataMaster:           master.IdDataMaster,
		Lokasi:                 master.LokasiPengambilan,
		JenisRencana:           master.JenisRencanaPembangunan,
		RencanaPembangunan:     master.RencanaPembangunan,
		KategoriPerlengkapan:   master.KategoriPerlengkapan,
		PerlengkapanLaluLintas: master.PerlengkapanLaluLintas,
		PersyaratanTambahan:    master.PersyaratanTambahan,
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": respone})
}

func (dm *DataMasterControler) TambahKategori(ctx *gin.Context) {
	kategori := ctx.Param("kategori")
	id := ctx.Param("id")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.ProductAddCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var master models.DataMaster

	resultsData := dm.DB.Where("id_data_master", id).First(&master)

	if resultsData.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsData.Error})
		return
	}

	exist := contains(master.JenisRencanaPembangunan, kategori)

	if exist {
		ctx.JSON(http.StatusConflict, gin.H{"status": "fail", "message": "Data sudah ada"})
		return
	}

	master.JenisRencanaPembangunan = append(master.JenisRencanaPembangunan, kategori)

	resultsSave := dm.DB.Save(&master)
	if resultsSave.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsSave.Error})
		return
	}

	respone := struct {
		IdDataMaster           uuid.UUID                  `json:"id_data_master,omitempty"`
		Lokasi                 []string                   `json:"lokasi_pengambilan,omitempty"`
		JenisRencana           []string                   `json:"jenis_rencana,omitempty"`
		RencanaPembangunan     []models.Rencana           `json:"rencana_pembangunan,omitempty"`
		KategoriPerlengkapan   []string                   `json:"kategori_perlengkapan,omitempty"`
		PerlengkapanLaluLintas []models.JenisPerlengkapan `json:"perlengkapan,omitempty"`
		PersyaratanTambahan    models.PersyaratanTambahan `json:"persyaratan_tambahan,omitempty"`
	}{
		IdDataMaster:           master.IdDataMaster,
		Lokasi:                 master.LokasiPengambilan,
		JenisRencana:           master.JenisRencanaPembangunan,
		RencanaPembangunan:     master.RencanaPembangunan,
		KategoriPerlengkapan:   master.KategoriPerlengkapan,
		PerlengkapanLaluLintas: master.PerlengkapanLaluLintas,
		PersyaratanTambahan:    master.PersyaratanTambahan,
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": respone})
}

func (dm *DataMasterControler) HapusKategori(ctx *gin.Context) {
	kategori := ctx.Param("kategori")
	id := ctx.Param("id")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.ProductDeleteCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var master models.DataMaster

	resultsData := dm.DB.Where("id_data_master", id).First(&master)

	if resultsData.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsData.Error})
		return
	}

	for i, item := range master.JenisRencanaPembangunan {
		if item == kategori {
			master.JenisRencanaPembangunan = append(master.JenisRencanaPembangunan[:i], master.JenisRencanaPembangunan[i+1:]...)
			break
		}
	}

	for i, item := range master.RencanaPembangunan {
		if item.Kategori == kategori {
			master.RencanaPembangunan = append(master.RencanaPembangunan[:i], master.RencanaPembangunan[i+1:]...)
			break
		}
	}

	resultsSave := dm.DB.Save(&master)
	if resultsSave.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsSave.Error})
		return
	}

	respone := struct {
		IdDataMaster           uuid.UUID                  `json:"id_data_master,omitempty"`
		Lokasi                 []string                   `json:"lokasi_pengambilan,omitempty"`
		JenisRencana           []string                   `json:"jenis_rencana,omitempty"`
		RencanaPembangunan     []models.Rencana           `json:"rencana_pembangunan,omitempty"`
		KategoriPerlengkapan   []string                   `json:"kategori_perlengkapan,omitempty"`
		PerlengkapanLaluLintas []models.JenisPerlengkapan `json:"perlengkapan,omitempty"`
		PersyaratanTambahan    models.PersyaratanTambahan `json:"persyaratan_tambahan,omitempty"`
	}{
		IdDataMaster:           master.IdDataMaster,
		Lokasi:                 master.LokasiPengambilan,
		JenisRencana:           master.JenisRencanaPembangunan,
		RencanaPembangunan:     master.RencanaPembangunan,
		KategoriPerlengkapan:   master.KategoriPerlengkapan,
		PerlengkapanLaluLintas: master.PerlengkapanLaluLintas,
		PersyaratanTambahan:    master.PersyaratanTambahan,
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": respone})
}

func (dm *DataMasterControler) EditKategori(ctx *gin.Context) {
	kategori := ctx.Param("kategori")
	newKategori := ctx.Param("new_kategori")
	id := ctx.Param("id")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.ProductUpdateCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var master models.DataMaster

	resultsData := dm.DB.Where("id_data_master", id).First(&master)

	if resultsData.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsData.Error})
		return
	}

	itemIndex := -1
	itemIndexRencana := -1

	for i, item := range master.JenisRencanaPembangunan {
		if item == kategori {
			itemIndex = i
			break
		}
	}

	if itemIndex != -1 {
		master.JenisRencanaPembangunan[itemIndex] = newKategori
	}

	for i, item := range master.RencanaPembangunan {
		if item.Kategori == kategori {
			itemIndexRencana = i
			break
		}
	}

	if itemIndexRencana != -1 {
		master.RencanaPembangunan[itemIndexRencana].Kategori = newKategori
	}

	resultsSave := dm.DB.Save(&master)
	if resultsSave.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsSave.Error})
		return
	}

	respone := struct {
		IdDataMaster           uuid.UUID                  `json:"id_data_master,omitempty"`
		Lokasi                 []string                   `json:"lokasi_pengambilan,omitempty"`
		JenisRencana           []string                   `json:"jenis_rencana,omitempty"`
		RencanaPembangunan     []models.Rencana           `json:"rencana_pembangunan,omitempty"`
		KategoriPerlengkapan   []string                   `json:"kategori_perlengkapan,omitempty"`
		PerlengkapanLaluLintas []models.JenisPerlengkapan `json:"perlengkapan,omitempty"`
		PersyaratanTambahan    models.PersyaratanTambahan `json:"persyaratan_tambahan,omitempty"`
	}{
		IdDataMaster:           master.IdDataMaster,
		Lokasi:                 master.LokasiPengambilan,
		JenisRencana:           master.JenisRencanaPembangunan,
		RencanaPembangunan:     master.RencanaPembangunan,
		KategoriPerlengkapan:   master.KategoriPerlengkapan,
		PerlengkapanLaluLintas: master.PerlengkapanLaluLintas,
		PersyaratanTambahan:    master.PersyaratanTambahan,
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": respone})
}

func (dm *DataMasterControler) TambahJenisRencanaPembangunan(ctx *gin.Context) {
	kategori := ctx.Param("kategori")
	rencana := ctx.Param("rencana")
	id := ctx.Param("id")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.ProductAddCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var master models.DataMaster

	resultsData := dm.DB.Where("id_data_master", id).First(&master)

	if resultsData.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsData.Error})
		return
	}

	kategoriExists := false
	jenisExists := false
	itemIndex := 0

	for i := range master.RencanaPembangunan {
		if master.RencanaPembangunan[i].Kategori == kategori {
			kategoriExists = true
			itemIndex = i
			for _, item := range master.RencanaPembangunan[i].JenisRencana {
				if item == rencana {
					jenisExists = true
					ctx.JSON(http.StatusConflict, gin.H{"status": "fail", "message": "Data sudah ada"})
					return
				}
			}
		}
	}

	if !kategoriExists {
		jenis := models.Rencana{
			Kategori:     kategori,
			JenisRencana: []string{rencana},
		}
		master.RencanaPembangunan = append(master.RencanaPembangunan, jenis)
	}

	if !jenisExists && kategoriExists {
		master.RencanaPembangunan[itemIndex].JenisRencana = append(master.RencanaPembangunan[itemIndex].JenisRencana, rencana)
	}

	resultsSave := dm.DB.Save(&master)
	if resultsSave.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsSave.Error})
		return
	}

	respone := struct {
		IdDataMaster           uuid.UUID                  `json:"id_data_master,omitempty"`
		Lokasi                 []string                   `json:"lokasi_pengambilan,omitempty"`
		JenisRencana           []string                   `json:"jenis_rencana,omitempty"`
		RencanaPembangunan     []models.Rencana           `json:"rencana_pembangunan,omitempty"`
		KategoriPerlengkapan   []string                   `json:"kategori_perlengkapan,omitempty"`
		PerlengkapanLaluLintas []models.JenisPerlengkapan `json:"perlengkapan,omitempty"`
		PersyaratanTambahan    models.PersyaratanTambahan `json:"persyaratan_tambahan,omitempty"`
	}{
		IdDataMaster:           master.IdDataMaster,
		Lokasi:                 master.LokasiPengambilan,
		JenisRencana:           master.JenisRencanaPembangunan,
		RencanaPembangunan:     master.RencanaPembangunan,
		KategoriPerlengkapan:   master.KategoriPerlengkapan,
		PerlengkapanLaluLintas: master.PerlengkapanLaluLintas,
		PersyaratanTambahan:    master.PersyaratanTambahan,
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": respone})
}

func (dm *DataMasterControler) HapusJenisRencanaPembangunan(ctx *gin.Context) {
	kategori := ctx.Param("kategori")
	rencana := ctx.Param("rencana")
	id := ctx.Param("id")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.ProductDeleteCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var master models.DataMaster

	resultsData := dm.DB.Where("id_data_master", id).First(&master)

	if resultsData.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsData.Error})
		return
	}

	for i := range master.RencanaPembangunan {
		if master.RencanaPembangunan[i].Kategori == kategori {
			for j, item := range master.RencanaPembangunan[i].JenisRencana {
				if item == rencana {
					master.RencanaPembangunan[i].JenisRencana = append(master.RencanaPembangunan[i].JenisRencana[:j], master.RencanaPembangunan[i].JenisRencana[j+1:]...)
				}
			}
		}
	}

	resultsSave := dm.DB.Save(&master)
	if resultsSave.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsSave.Error})
		return
	}

	respone := struct {
		IdDataMaster           uuid.UUID                  `json:"id_data_master,omitempty"`
		Lokasi                 []string                   `json:"lokasi_pengambilan,omitempty"`
		JenisRencana           []string                   `json:"jenis_rencana,omitempty"`
		RencanaPembangunan     []models.Rencana           `json:"rencana_pembangunan,omitempty"`
		KategoriPerlengkapan   []string                   `json:"kategori_perlengkapan,omitempty"`
		PerlengkapanLaluLintas []models.JenisPerlengkapan `json:"perlengkapan,omitempty"`
		PersyaratanTambahan    models.PersyaratanTambahan `json:"persyaratan_tambahan,omitempty"`
	}{
		IdDataMaster:           master.IdDataMaster,
		Lokasi:                 master.LokasiPengambilan,
		JenisRencana:           master.JenisRencanaPembangunan,
		RencanaPembangunan:     master.RencanaPembangunan,
		KategoriPerlengkapan:   master.KategoriPerlengkapan,
		PerlengkapanLaluLintas: master.PerlengkapanLaluLintas,
		PersyaratanTambahan:    master.PersyaratanTambahan,
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": respone})
}

func (dm *DataMasterControler) EditJenisRencanaPembangunan(ctx *gin.Context) {
	kategori := ctx.Param("kategori")
	rencana := ctx.Param("rencana")
	newRencana := ctx.Param("rencana_new")
	id := ctx.Param("id")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.ProductUpdateCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var master models.DataMaster

	resultsData := dm.DB.Where("id_data_master", id).First(&master)

	if resultsData.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsData.Error})
		return
	}

	itemIndexKategori := -1
	itemIndexRencana := -1

	for i := range master.RencanaPembangunan {
		if master.RencanaPembangunan[i].Kategori == kategori {
			itemIndexKategori = i
			for j, item := range master.RencanaPembangunan[i].JenisRencana {
				if item == rencana {
					itemIndexRencana = j
					break
				}
			}
		}
	}

	if itemIndexKategori != -1 && itemIndexRencana != -1 {
		master.RencanaPembangunan[itemIndexKategori].JenisRencana[itemIndexRencana] = newRencana
	}

	resultsSave := dm.DB.Save(&master)
	if resultsSave.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsSave.Error})
		return
	}

	respone := struct {
		IdDataMaster           uuid.UUID                  `json:"id_data_master,omitempty"`
		Lokasi                 []string                   `json:"lokasi_pengambilan,omitempty"`
		JenisRencana           []string                   `json:"jenis_rencana,omitempty"`
		RencanaPembangunan     []models.Rencana           `json:"rencana_pembangunan,omitempty"`
		KategoriPerlengkapan   []string                   `json:"kategori_perlengkapan,omitempty"`
		PerlengkapanLaluLintas []models.JenisPerlengkapan `json:"perlengkapan,omitempty"`
		PersyaratanTambahan    models.PersyaratanTambahan `json:"persyaratan_tambahan,omitempty"`
	}{
		IdDataMaster:           master.IdDataMaster,
		Lokasi:                 master.LokasiPengambilan,
		JenisRencana:           master.JenisRencanaPembangunan,
		RencanaPembangunan:     master.RencanaPembangunan,
		KategoriPerlengkapan:   master.KategoriPerlengkapan,
		PerlengkapanLaluLintas: master.PerlengkapanLaluLintas,
		PersyaratanTambahan:    master.PersyaratanTambahan,
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": respone})
}

func (dm *DataMasterControler) TambahKategoriPerlengkapan(ctx *gin.Context) {
	kategori := ctx.Param("kategori")
	id := ctx.Param("id")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.ProductAddCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var master models.DataMaster

	resultsData := dm.DB.Where("id_data_master", id).First(&master)

	if resultsData.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsData.Error})
		return
	}

	exist := contains(master.KategoriPerlengkapan, kategori)

	if exist {
		ctx.JSON(http.StatusConflict, gin.H{"status": "fail", "message": "Data sudah ada"})
		return
	}

	master.KategoriPerlengkapan = append(master.KategoriPerlengkapan, kategori)

	resultsSave := dm.DB.Save(&master)
	if resultsSave.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsSave.Error})
		return
	}

	respone := struct {
		IdDataMaster           uuid.UUID                  `json:"id_data_master,omitempty"`
		Lokasi                 []string                   `json:"lokasi_pengambilan,omitempty"`
		JenisRencana           []string                   `json:"jenis_rencana,omitempty"`
		RencanaPembangunan     []models.Rencana           `json:"rencana_pembangunan,omitempty"`
		KategoriPerlengkapan   []string                   `json:"kategori_perlengkapan,omitempty"`
		PerlengkapanLaluLintas []models.JenisPerlengkapan `json:"perlengkapan,omitempty"`
		PersyaratanTambahan    models.PersyaratanTambahan `json:"persyaratan_tambahan,omitempty"`
	}{
		IdDataMaster:           master.IdDataMaster,
		Lokasi:                 master.LokasiPengambilan,
		JenisRencana:           master.JenisRencanaPembangunan,
		RencanaPembangunan:     master.RencanaPembangunan,
		KategoriPerlengkapan:   master.KategoriPerlengkapan,
		PerlengkapanLaluLintas: master.PerlengkapanLaluLintas,
		PersyaratanTambahan:    master.PersyaratanTambahan,
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": respone})
}

func (dm *DataMasterControler) HapusKategoriPerlengkapan(ctx *gin.Context) {
	kategori := ctx.Param("kategori")
	id := ctx.Param("id")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.ProductDeleteCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var master models.DataMaster

	resultsData := dm.DB.Where("id_data_master", id).First(&master)

	if resultsData.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsData.Error})
		return
	}

	for i, item := range master.KategoriPerlengkapan {
		if item == kategori {
			master.KategoriPerlengkapan = append(master.KategoriPerlengkapan[:i], master.KategoriPerlengkapan[i+1:]...)
			break
		}
	}

	for i, item := range master.PerlengkapanLaluLintas {
		if item.Kategori == kategori {
			master.PerlengkapanLaluLintas = append(master.PerlengkapanLaluLintas[:i], master.PerlengkapanLaluLintas[i+1:]...)
			break
		}
	}

	resultsSave := dm.DB.Save(&master)
	if resultsSave.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsSave.Error})
		return
	}

	respone := struct {
		IdDataMaster           uuid.UUID                  `json:"id_data_master,omitempty"`
		Lokasi                 []string                   `json:"lokasi_pengambilan,omitempty"`
		JenisRencana           []string                   `json:"jenis_rencana,omitempty"`
		RencanaPembangunan     []models.Rencana           `json:"rencana_pembangunan,omitempty"`
		KategoriPerlengkapan   []string                   `json:"kategori_perlengkapan,omitempty"`
		PerlengkapanLaluLintas []models.JenisPerlengkapan `json:"perlengkapan,omitempty"`
		PersyaratanTambahan    models.PersyaratanTambahan `json:"persyaratan_tambahan,omitempty"`
	}{
		IdDataMaster:           master.IdDataMaster,
		Lokasi:                 master.LokasiPengambilan,
		JenisRencana:           master.JenisRencanaPembangunan,
		RencanaPembangunan:     master.RencanaPembangunan,
		KategoriPerlengkapan:   master.KategoriPerlengkapan,
		PerlengkapanLaluLintas: master.PerlengkapanLaluLintas,
		PersyaratanTambahan:    master.PersyaratanTambahan,
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": respone})
}

func (dm *DataMasterControler) EditKategoriPerlengkapan(ctx *gin.Context) {
	kategori := ctx.Param("kategori")
	newKategori := ctx.Param("new_kategori")
	id := ctx.Param("id")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.ProductUpdateCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var master models.DataMaster

	resultsData := dm.DB.Where("id_data_master", id).First(&master)

	if resultsData.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsData.Error})
		return
	}

	itemIndex := -1
	itemIndexKategori := -1

	for i, item := range master.KategoriPerlengkapan {
		if item == kategori {
			itemIndex = i
			break
		}
	}

	if itemIndex != -1 {
		master.KategoriPerlengkapan[itemIndex] = newKategori
	}

	for i, item := range master.PerlengkapanLaluLintas {
		if item.Kategori == kategori {
			itemIndexKategori = i
			break
		}
	}

	if itemIndexKategori != -1 {
		master.PerlengkapanLaluLintas[itemIndexKategori].Kategori = newKategori
	}

	resultsSave := dm.DB.Save(&master)
	if resultsSave.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsSave.Error})
		return
	}

	respone := struct {
		IdDataMaster           uuid.UUID                  `json:"id_data_master,omitempty"`
		Lokasi                 []string                   `json:"lokasi_pengambilan,omitempty"`
		JenisRencana           []string                   `json:"jenis_rencana,omitempty"`
		RencanaPembangunan     []models.Rencana           `json:"rencana_pembangunan,omitempty"`
		KategoriPerlengkapan   []string                   `json:"kategori_perlengkapan,omitempty"`
		PerlengkapanLaluLintas []models.JenisPerlengkapan `json:"perlengkapan,omitempty"`
		PersyaratanTambahan    models.PersyaratanTambahan `json:"persyaratan_tambahan,omitempty"`
	}{
		IdDataMaster:           master.IdDataMaster,
		Lokasi:                 master.LokasiPengambilan,
		JenisRencana:           master.JenisRencanaPembangunan,
		RencanaPembangunan:     master.RencanaPembangunan,
		KategoriPerlengkapan:   master.KategoriPerlengkapan,
		PerlengkapanLaluLintas: master.PerlengkapanLaluLintas,
		PersyaratanTambahan:    master.PersyaratanTambahan,
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": respone})
}

func (dm *DataMasterControler) TambahPerlengkapan(ctx *gin.Context) {
	kategori := ctx.Param("kategori")
	perlengkapan := ctx.Param("perlengkapan")
	id := ctx.Param("id")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.ProductAddCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var master models.DataMaster

	resultsData := dm.DB.Where("id_data_master", id).First(&master)

	if resultsData.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsData.Error})
		return
	}

	kategoriExists := false
	perlengkapanExist := false
	itemIndex := 0

	for i := range master.PerlengkapanLaluLintas {
		if master.PerlengkapanLaluLintas[i].Kategori == kategori {
			kategoriExists = true
			itemIndex = i
			for _, item := range master.PerlengkapanLaluLintas[i].Perlengkapan {
				if item.JenisPerlengkapan == perlengkapan {
					perlengkapanExist = true
					ctx.JSON(http.StatusConflict, gin.H{"status": "fail", "message": "Data sudah ada"})
					return
				}
			}
		}
	}

	file, err := ctx.FormFile("perlengkapan")
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

	if !kategoriExists {
		perlengkapan := models.PerlengkapanItem{
			JenisPerlengkapan:  perlengkapan,
			GambarPerlengkapan: data,
		}
		jenis := models.JenisPerlengkapan{
			Kategori:     kategori,
			Perlengkapan: []models.PerlengkapanItem{perlengkapan},
		}
		master.PerlengkapanLaluLintas = append(master.PerlengkapanLaluLintas, jenis)
	}

	if !perlengkapanExist && kategoriExists {
		perlengkapan := models.PerlengkapanItem{
			JenisPerlengkapan:  perlengkapan,
			GambarPerlengkapan: data,
		}
		master.PerlengkapanLaluLintas[itemIndex].Perlengkapan = append(master.PerlengkapanLaluLintas[itemIndex].Perlengkapan, perlengkapan)
	}

	resultsSave := dm.DB.Save(&master)
	if resultsSave.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsSave.Error})
		return
	}

	respone := struct {
		IdDataMaster           uuid.UUID                  `json:"id_data_master,omitempty"`
		Lokasi                 []string                   `json:"lokasi_pengambilan,omitempty"`
		JenisRencana           []string                   `json:"jenis_rencana,omitempty"`
		RencanaPembangunan     []models.Rencana           `json:"rencana_pembangunan,omitempty"`
		KategoriPerlengkapan   []string                   `json:"kategori_perlengkapan,omitempty"`
		PerlengkapanLaluLintas []models.JenisPerlengkapan `json:"perlengkapan,omitempty"`
		PersyaratanTambahan    models.PersyaratanTambahan `json:"persyaratan_tambahan,omitempty"`
	}{
		IdDataMaster:           master.IdDataMaster,
		Lokasi:                 master.LokasiPengambilan,
		JenisRencana:           master.JenisRencanaPembangunan,
		RencanaPembangunan:     master.RencanaPembangunan,
		KategoriPerlengkapan:   master.KategoriPerlengkapan,
		PerlengkapanLaluLintas: master.PerlengkapanLaluLintas,
		PersyaratanTambahan:    master.PersyaratanTambahan,
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": respone})
}

func (dm *DataMasterControler) HapuspPerlengkapan(ctx *gin.Context) {
	kategori := ctx.Param("kategori")
	perlengkapan := ctx.Param("perlengkapan")
	id := ctx.Param("id")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.ProductDeleteCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var master models.DataMaster

	resultsData := dm.DB.Where("id_data_master", id).First(&master)

	if resultsData.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsData.Error})
		return
	}

	for i := range master.PerlengkapanLaluLintas {
		if master.PerlengkapanLaluLintas[i].Kategori == kategori {
			for j, item := range master.PerlengkapanLaluLintas[i].Perlengkapan {
				if item.JenisPerlengkapan == perlengkapan {
					master.PerlengkapanLaluLintas[i].Perlengkapan = append(master.PerlengkapanLaluLintas[i].Perlengkapan[:j], master.PerlengkapanLaluLintas[i].Perlengkapan[j+1:]...)
				}
			}
		}
	}

	resultsSave := dm.DB.Save(&master)
	if resultsSave.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsSave.Error})
		return
	}

	respone := struct {
		IdDataMaster           uuid.UUID                  `json:"id_data_master,omitempty"`
		Lokasi                 []string                   `json:"lokasi_pengambilan,omitempty"`
		JenisRencana           []string                   `json:"jenis_rencana,omitempty"`
		RencanaPembangunan     []models.Rencana           `json:"rencana_pembangunan,omitempty"`
		KategoriPerlengkapan   []string                   `json:"kategori_perlengkapan,omitempty"`
		PerlengkapanLaluLintas []models.JenisPerlengkapan `json:"perlengkapan,omitempty"`
		PersyaratanTambahan    models.PersyaratanTambahan `json:"persyaratan_tambahan,omitempty"`
	}{
		IdDataMaster:           master.IdDataMaster,
		Lokasi:                 master.LokasiPengambilan,
		JenisRencana:           master.JenisRencanaPembangunan,
		RencanaPembangunan:     master.RencanaPembangunan,
		KategoriPerlengkapan:   master.KategoriPerlengkapan,
		PerlengkapanLaluLintas: master.PerlengkapanLaluLintas,
		PersyaratanTambahan:    master.PersyaratanTambahan,
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": respone})
}

func (dm *DataMasterControler) EditPerlengkapan(ctx *gin.Context) {
	kategori := ctx.Param("kategori")
	perlengkapan := ctx.Param("perlengkapan")
	newPerlengkapan := ctx.Param("perlengkapan_new")
	id := ctx.Param("id")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.ProductUpdateCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var master models.DataMaster

	resultsData := dm.DB.Where("id_data_master", id).First(&master)

	if resultsData.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsData.Error})
		return
	}

	itemIndexKategori := -1
	itemIndexPerlengkapan := -1

	for i := range master.PerlengkapanLaluLintas {
		if master.PerlengkapanLaluLintas[i].Kategori == kategori {
			itemIndexKategori = i
			for j, item := range master.PerlengkapanLaluLintas[i].Perlengkapan {
				if item.JenisPerlengkapan == perlengkapan {
					itemIndexPerlengkapan = j
					break
				}
			}
		}
	}

	file, _ := ctx.FormFile("perlengkapan")

	if itemIndexKategori != -1 && itemIndexPerlengkapan != -1 {
		master.PerlengkapanLaluLintas[itemIndexKategori].Perlengkapan[itemIndexPerlengkapan].JenisPerlengkapan = newPerlengkapan
		if file != nil {
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

			master.PerlengkapanLaluLintas[itemIndexKategori].Perlengkapan[itemIndexPerlengkapan].GambarPerlengkapan = data
		}

	}

	resultsSave := dm.DB.Save(&master)
	if resultsSave.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsSave.Error})
		return
	}

	respone := struct {
		IdDataMaster           uuid.UUID                  `json:"id_data_master,omitempty"`
		Lokasi                 []string                   `json:"lokasi_pengambilan,omitempty"`
		JenisRencana           []string                   `json:"jenis_rencana,omitempty"`
		RencanaPembangunan     []models.Rencana           `json:"rencana_pembangunan,omitempty"`
		KategoriPerlengkapan   []string                   `json:"kategori_perlengkapan,omitempty"`
		PerlengkapanLaluLintas []models.JenisPerlengkapan `json:"perlengkapan,omitempty"`
		PersyaratanTambahan    models.PersyaratanTambahan `json:"persyaratan_tambahan,omitempty"`
	}{
		IdDataMaster:           master.IdDataMaster,
		Lokasi:                 master.LokasiPengambilan,
		JenisRencana:           master.JenisRencanaPembangunan,
		RencanaPembangunan:     master.RencanaPembangunan,
		KategoriPerlengkapan:   master.KategoriPerlengkapan,
		PerlengkapanLaluLintas: master.PerlengkapanLaluLintas,
		PersyaratanTambahan:    master.PersyaratanTambahan,
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": respone})
}

func (dm *DataMasterControler) TambahPersyaratanAndalalin(ctx *gin.Context) {
	var payload *models.PersyaratanTambahanInput
	id := ctx.Param("id")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.ProductAddCredential]

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

	var master models.DataMaster

	resultsData := dm.DB.Where("id_data_master", id).First(&master)

	if resultsData.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsData.Error})
		return
	}

	persyaratanExist := false

	for i := range master.PersyaratanTambahan.PersyaratanTambahanAndalalin {
		if master.PersyaratanTambahan.PersyaratanTambahanAndalalin[i].Persyaratan == payload.Persyaratan {
			persyaratanExist = true
			ctx.JSON(http.StatusConflict, gin.H{"status": "fail", "message": "Data sudah ada"})
			return
		}
	}

	if !persyaratanExist {
		persyaratan := models.PersyaratanTambahanInput{
			Persyaratan:           payload.Persyaratan,
			KeteranganPersyaratan: payload.KeteranganPersyaratan,
		}
		master.PersyaratanTambahan.PersyaratanTambahanAndalalin = append(master.PersyaratanTambahan.PersyaratanTambahanAndalalin, persyaratan)
	}

	resultsSave := dm.DB.Save(&master)
	if resultsSave.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsSave.Error})
		return
	}

	respone := struct {
		IdDataMaster           uuid.UUID                  `json:"id_data_master,omitempty"`
		Lokasi                 []string                   `json:"lokasi_pengambilan,omitempty"`
		JenisRencana           []string                   `json:"jenis_rencana,omitempty"`
		RencanaPembangunan     []models.Rencana           `json:"rencana_pembangunan,omitempty"`
		KategoriPerlengkapan   []string                   `json:"kategori_perlengkapan,omitempty"`
		PerlengkapanLaluLintas []models.JenisPerlengkapan `json:"perlengkapan,omitempty"`
		PersyaratanTambahan    models.PersyaratanTambahan `json:"persyaratan_tambahan,omitempty"`
	}{
		IdDataMaster:           master.IdDataMaster,
		Lokasi:                 master.LokasiPengambilan,
		JenisRencana:           master.JenisRencanaPembangunan,
		RencanaPembangunan:     master.RencanaPembangunan,
		KategoriPerlengkapan:   master.KategoriPerlengkapan,
		PerlengkapanLaluLintas: master.PerlengkapanLaluLintas,
		PersyaratanTambahan:    master.PersyaratanTambahan,
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": respone})
}

func (dm *DataMasterControler) HapusPersyaratanAndalalin(ctx *gin.Context) {
	persyaratan := ctx.Param("persyaratan")
	id := ctx.Param("id")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.ProductDeleteCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var master models.DataMaster

	resultsData := dm.DB.Where("id_data_master", id).First(&master)

	if resultsData.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsData.Error})
		return
	}

	for i := range master.PersyaratanTambahan.PersyaratanTambahanAndalalin {
		if master.PersyaratanTambahan.PersyaratanTambahanAndalalin[i].Persyaratan == persyaratan {
			master.PersyaratanTambahan.PersyaratanTambahanAndalalin = append(master.PersyaratanTambahan.PersyaratanTambahanAndalalin[:i], master.PersyaratanTambahan.PersyaratanTambahanAndalalin[i+1:]...)
			break
		}
	}

	var andalalin []models.Andalalin

	results := dm.DB.Find(&andalalin)

	if results.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": results.Error})
		return
	} else {
		dataFile := []file{}
		for i, permohonan := range andalalin {
			for j, tambahan := range permohonan.PersyaratanTambahan {
				if tambahan.Persyaratan == persyaratan {
					oldSubstr := "/"
					newSubstr := "-"

					result := strings.Replace(permohonan.Kode, oldSubstr, newSubstr, -1)
					fileName := result + ".pdf"

					dataFile = append(dataFile, file{Name: fileName, File: tambahan.Berkas})
					andalalin[i].PersyaratanTambahan = append(andalalin[i].PersyaratanTambahan[:j], andalalin[i].PersyaratanTambahan[j+1:]...)
					break
				}
			}
		}

		zipFile := persyaratan + ".zip"
		error = compressFiles(zipFile, dataFile)
		if error != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": error})
			return
		}

		zipData, errorZip := os.ReadFile(zipFile)
		if errorZip != nil {
			ctx.JSON(http.StatusNoContent, gin.H{"status": "error", "message": errorZip})
			return
		}

		base64ZipData := base64.StdEncoding.EncodeToString(zipData)

		dm.DB.Save(&andalalin)
		resultsSave := dm.DB.Save(&master)
		if resultsSave.Error != nil {
			ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsSave.Error})
			return
		}

		respone := struct {
			IdDataMaster           uuid.UUID                  `json:"id_data_master,omitempty"`
			Lokasi                 []string                   `json:"lokasi_pengambilan,omitempty"`
			JenisRencana           []string                   `json:"jenis_rencana,omitempty"`
			RencanaPembangunan     []models.Rencana           `json:"rencana_pembangunan,omitempty"`
			KategoriPerlengkapan   []string                   `json:"kategori_perlengkapan,omitempty"`
			PerlengkapanLaluLintas []models.JenisPerlengkapan `json:"perlengkapan,omitempty"`
			PersyaratanTambahan    models.PersyaratanTambahan `json:"persyaratan_tambahan,omitempty"`
		}{
			IdDataMaster:           master.IdDataMaster,
			Lokasi:                 master.LokasiPengambilan,
			JenisRencana:           master.JenisRencanaPembangunan,
			RencanaPembangunan:     master.RencanaPembangunan,
			KategoriPerlengkapan:   master.KategoriPerlengkapan,
			PerlengkapanLaluLintas: master.PerlengkapanLaluLintas,
			PersyaratanTambahan:    master.PersyaratanTambahan,
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": respone, "file": base64ZipData})
	}
}

func (dm *DataMasterControler) EditPersyaratanAndalalin(ctx *gin.Context) {
	var payload *models.PersyaratanTambahanInput
	id := ctx.Param("id")
	syarat := ctx.Param("persyaratan")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.ProductUpdateCredential]

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

	var master models.DataMaster

	resultsData := dm.DB.Where("id_data_master", id).First(&master)

	if resultsData.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsData.Error})
		return
	}

	itemIndex := -1

	for i := range master.PersyaratanTambahan.PersyaratanTambahanAndalalin {
		if master.PersyaratanTambahan.PersyaratanTambahanAndalalin[i].Persyaratan == syarat {
			itemIndex = i
			break
		}
	}

	if itemIndex != -1 {
		if master.PersyaratanTambahan.PersyaratanTambahanAndalalin[itemIndex].Persyaratan != payload.Persyaratan {
			master.PersyaratanTambahan.PersyaratanTambahanAndalalin[itemIndex].Persyaratan = payload.Persyaratan
		}

		if master.PersyaratanTambahan.PersyaratanTambahanAndalalin[itemIndex].KeteranganPersyaratan != payload.KeteranganPersyaratan {
			master.PersyaratanTambahan.PersyaratanTambahanAndalalin[itemIndex].KeteranganPersyaratan = payload.KeteranganPersyaratan
		}

	}

	resultsSave := dm.DB.Save(&master)
	if resultsSave.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsSave.Error})
		return
	}

	respone := struct {
		IdDataMaster           uuid.UUID                  `json:"id_data_master,omitempty"`
		Lokasi                 []string                   `json:"lokasi_pengambilan,omitempty"`
		JenisRencana           []string                   `json:"jenis_rencana,omitempty"`
		RencanaPembangunan     []models.Rencana           `json:"rencana_pembangunan,omitempty"`
		KategoriPerlengkapan   []string                   `json:"kategori_perlengkapan,omitempty"`
		PerlengkapanLaluLintas []models.JenisPerlengkapan `json:"perlengkapan,omitempty"`
		PersyaratanTambahan    models.PersyaratanTambahan `json:"persyaratan_tambahan,omitempty"`
	}{
		IdDataMaster:           master.IdDataMaster,
		Lokasi:                 master.LokasiPengambilan,
		JenisRencana:           master.JenisRencanaPembangunan,
		RencanaPembangunan:     master.RencanaPembangunan,
		KategoriPerlengkapan:   master.KategoriPerlengkapan,
		PerlengkapanLaluLintas: master.PerlengkapanLaluLintas,
		PersyaratanTambahan:    master.PersyaratanTambahan,
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": respone})
}

func (dm *DataMasterControler) TambahPersyaratanPerlalin(ctx *gin.Context) {
	var payload *models.PersyaratanTambahanInput
	id := ctx.Param("id")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.ProductAddCredential]

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

	var master models.DataMaster

	resultsData := dm.DB.Where("id_data_master", id).First(&master)

	if resultsData.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsData.Error})
		return
	}

	persyaratanExist := false

	for i := range master.PersyaratanTambahan.PersyaratanTambahanPerlalin {
		if master.PersyaratanTambahan.PersyaratanTambahanPerlalin[i].Persyaratan == payload.Persyaratan {
			persyaratanExist = true
			ctx.JSON(http.StatusConflict, gin.H{"status": "fail", "message": "Data sudah ada"})
			return
		}
	}

	if !persyaratanExist {
		persyaratan := models.PersyaratanTambahanInput{
			Persyaratan:           payload.Persyaratan,
			KeteranganPersyaratan: payload.KeteranganPersyaratan,
		}
		master.PersyaratanTambahan.PersyaratanTambahanPerlalin = append(master.PersyaratanTambahan.PersyaratanTambahanPerlalin, persyaratan)
	}

	resultsSave := dm.DB.Save(&master)
	if resultsSave.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsSave.Error})
		return
	}

	respone := struct {
		IdDataMaster           uuid.UUID                  `json:"id_data_master,omitempty"`
		Lokasi                 []string                   `json:"lokasi_pengambilan,omitempty"`
		JenisRencana           []string                   `json:"jenis_rencana,omitempty"`
		RencanaPembangunan     []models.Rencana           `json:"rencana_pembangunan,omitempty"`
		KategoriPerlengkapan   []string                   `json:"kategori_perlengkapan,omitempty"`
		PerlengkapanLaluLintas []models.JenisPerlengkapan `json:"perlengkapan,omitempty"`
		PersyaratanTambahan    models.PersyaratanTambahan `json:"persyaratan_tambahan,omitempty"`
	}{
		IdDataMaster:           master.IdDataMaster,
		Lokasi:                 master.LokasiPengambilan,
		JenisRencana:           master.JenisRencanaPembangunan,
		RencanaPembangunan:     master.RencanaPembangunan,
		KategoriPerlengkapan:   master.KategoriPerlengkapan,
		PerlengkapanLaluLintas: master.PerlengkapanLaluLintas,
		PersyaratanTambahan:    master.PersyaratanTambahan,
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": respone})
}

func (dm *DataMasterControler) HapusPersyaratanPerlalin(ctx *gin.Context) {
	persyaratan := ctx.Param("persyaratan")
	id := ctx.Param("id")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.ProductDeleteCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var master models.DataMaster

	resultsData := dm.DB.Where("id_data_master", id).First(&master)

	if resultsData.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsData.Error})
		return
	}

	for i := range master.PersyaratanTambahan.PersyaratanTambahanPerlalin {
		if master.PersyaratanTambahan.PersyaratanTambahanPerlalin[i].Persyaratan == persyaratan {
			master.PersyaratanTambahan.PersyaratanTambahanPerlalin = append(master.PersyaratanTambahan.PersyaratanTambahanPerlalin[:i], master.PersyaratanTambahan.PersyaratanTambahanPerlalin[i+1:]...)
			break
		}
	}

	var perlalin []models.Perlalin

	results := dm.DB.Find(&perlalin)

	if results.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": results.Error})
		return
	} else {
		dataFile := []file{}
		for i, permohonan := range perlalin {
			for j, tambahan := range permohonan.PersyaratanTambahan {
				if tambahan.Persyaratan == persyaratan {
					oldSubstr := "/"
					newSubstr := "-"

					result := strings.Replace(permohonan.Kode, oldSubstr, newSubstr, -1)
					fileName := result + ".pdf"

					dataFile = append(dataFile, file{Name: fileName, File: tambahan.Berkas})
					perlalin[i].PersyaratanTambahan = append(perlalin[i].PersyaratanTambahan[:j], perlalin[i].PersyaratanTambahan[j+1:]...)
					break
				}
			}
		}

		zipFile := persyaratan + ".zip"
		error = compressFiles(zipFile, dataFile)
		if error != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": error})
			return
		}

		zipData, errorZip := os.ReadFile(zipFile)
		if errorZip != nil {
			ctx.JSON(http.StatusNoContent, gin.H{"status": "error", "message": errorZip})
			return
		}

		base64ZipData := base64.StdEncoding.EncodeToString(zipData)

		dm.DB.Save(&perlalin)
		resultsSave := dm.DB.Save(&master)
		if resultsSave.Error != nil {
			ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsSave.Error})
			return
		}

		respone := struct {
			IdDataMaster           uuid.UUID                  `json:"id_data_master,omitempty"`
			Lokasi                 []string                   `json:"lokasi_pengambilan,omitempty"`
			JenisRencana           []string                   `json:"jenis_rencana,omitempty"`
			RencanaPembangunan     []models.Rencana           `json:"rencana_pembangunan,omitempty"`
			KategoriPerlengkapan   []string                   `json:"kategori_perlengkapan,omitempty"`
			PerlengkapanLaluLintas []models.JenisPerlengkapan `json:"perlengkapan,omitempty"`
			PersyaratanTambahan    models.PersyaratanTambahan `json:"persyaratan_tambahan,omitempty"`
		}{
			IdDataMaster:           master.IdDataMaster,
			Lokasi:                 master.LokasiPengambilan,
			JenisRencana:           master.JenisRencanaPembangunan,
			RencanaPembangunan:     master.RencanaPembangunan,
			KategoriPerlengkapan:   master.KategoriPerlengkapan,
			PerlengkapanLaluLintas: master.PerlengkapanLaluLintas,
			PersyaratanTambahan:    master.PersyaratanTambahan,
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": respone, "file": base64ZipData})
	}
}

func (dm *DataMasterControler) EditPersyaratanPerlalin(ctx *gin.Context) {
	var payload *models.PersyaratanTambahanInput
	id := ctx.Param("id")
	syarat := ctx.Param("persyaratan")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.ProductUpdateCredential]

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

	var master models.DataMaster

	resultsData := dm.DB.Where("id_data_master", id).First(&master)

	if resultsData.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsData.Error})
		return
	}

	itemIndex := -1

	for i := range master.PersyaratanTambahan.PersyaratanTambahanPerlalin {
		if master.PersyaratanTambahan.PersyaratanTambahanPerlalin[i].Persyaratan == syarat {
			itemIndex = i
			break
		}
	}

	if itemIndex != -1 {
		if master.PersyaratanTambahan.PersyaratanTambahanPerlalin[itemIndex].Persyaratan != payload.Persyaratan {
			master.PersyaratanTambahan.PersyaratanTambahanPerlalin[itemIndex].Persyaratan = payload.Persyaratan
		}

		if master.PersyaratanTambahan.PersyaratanTambahanPerlalin[itemIndex].KeteranganPersyaratan != payload.KeteranganPersyaratan {
			master.PersyaratanTambahan.PersyaratanTambahanPerlalin[itemIndex].KeteranganPersyaratan = payload.KeteranganPersyaratan
		}
	}

	resultsSave := dm.DB.Save(&master)
	if resultsSave.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": resultsSave.Error})
		return
	}

	respone := struct {
		IdDataMaster           uuid.UUID                  `json:"id_data_master,omitempty"`
		Lokasi                 []string                   `json:"lokasi_pengambilan,omitempty"`
		JenisRencana           []string                   `json:"jenis_rencana,omitempty"`
		RencanaPembangunan     []models.Rencana           `json:"rencana_pembangunan,omitempty"`
		KategoriPerlengkapan   []string                   `json:"kategori_perlengkapan,omitempty"`
		PerlengkapanLaluLintas []models.JenisPerlengkapan `json:"perlengkapan,omitempty"`
		PersyaratanTambahan    models.PersyaratanTambahan `json:"persyaratan_tambahan,omitempty"`
	}{
		IdDataMaster:           master.IdDataMaster,
		Lokasi:                 master.LokasiPengambilan,
		JenisRencana:           master.JenisRencanaPembangunan,
		RencanaPembangunan:     master.RencanaPembangunan,
		KategoriPerlengkapan:   master.KategoriPerlengkapan,
		PerlengkapanLaluLintas: master.PerlengkapanLaluLintas,
		PersyaratanTambahan:    master.PersyaratanTambahan,
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": respone})
}
