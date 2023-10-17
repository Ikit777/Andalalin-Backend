package models

import (
	"github.com/google/uuid"
)

type Andalalin struct {
	//Data Pemohon
	IdAndalalin            uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	IdUser                 uuid.UUID `gorm:"type:varchar(255);not null"`
	JenisAndalalin         string    `gorm:"type:varchar(255);not null"`
	Kategori               string    `gorm:"type:varchar(255);not null"`
	Jenis                  string    `gorm:"type:varchar(255);not null"`
	Kode                   string    `gorm:"type:varchar(255);not null"`
	NikPemohon             string    `gorm:"type:varchar(255);not null"`
	NamaPemohon            string    `gorm:"type:varchar(255);not null"`
	EmailPemohon           string    `gorm:"type:varchar(255);not null"`
	TempatLahirPemohon     string    `gorm:"type:varchar(255);not null"`
	TanggalLahirPemohon    string    `gorm:"type:varchar(255);not null"`
	AlamatPemohon          string    `gorm:"type:varchar(255);not null"`
	JenisKelaminPemohon    string    `sql:"type:enum('Laki-laki', 'Perempuan');not null"`
	NomerPemohon           string    `gorm:"type:varchar(255);not null"`
	NomerSelulerPemohon    string    `gorm:"type:varchar(255);not null"`
	JabatanPemohon         string    `gorm:"type:varchar(255);not null"`
	LokasiPengambilan      string    `gorm:"type:varchar(255);not null"`
	WaktuAndalalin         string    `gorm:"not null"`
	TanggalAndalalin       string    `gorm:"not null"`
	StatusAndalalin        string    `sql:"type:enum('Cek persyaratan', 'Persyaratan tidak terpenuhi', 'Berita acara pemeriksaan', 'Persetujuan dokumen', 'Pembuatan surat keputusan', 'Permohonan selesai')"`
	TandaTerimaPendaftaran []byte

	//Data Perusahaan
	NamaPerusahaan       string `gorm:"type:varchar(255);not null"`
	AlamatPerusahaan     string `gorm:"type:varchar(255);not null"`
	NomerPerusahaan      string `gorm:"type:varchar(255);not null"`
	EmailPerusahaan      string `gorm:"type:varchar(255);not null"`
	ProvinsiPerusahaan   string `gorm:"type:varchar(255);not null"`
	KabupatenPerusahaan  string `gorm:"type:varchar(255);not null"`
	KecamatanPerusahaan  string `gorm:"type:varchar(255);not null"`
	KelurahaanPerusahaan string `gorm:"type:varchar(255);not null"`
	NamaPimpinan         string `gorm:"type:varchar(255);not null"`
	JabatanPimpinan      string `gorm:"type:varchar(255);not null"`
	JenisKelaminPimpinan string `sql:"type:enum('Laki-laki', 'Perempuan');not null"`
	JenisKegiatan        string `gorm:"type:varchar(255);not null"`
	Peruntukan           string `gorm:"type:varchar(255);not null"`
	LuasLahan            string `gorm:"type:varchar(255);not null"`
	AlamatPersil         string `gorm:"type:varchar(255);not null"`
	KelurahanPersil      string `gorm:"type:varchar(255);not null"`
	NomerSKRK            string `gorm:"type:varchar(255);not null"`
	TanggalSKRK          string `gorm:"type:varchar(255);not null"`

	//Data Persyaratan
	KartuTandaPenduduk []byte
	AktaPendirianBadan []byte
	SuratKuasa         []byte

	PersyaratanTambahan []PersyaratanTambahanPermohonan `gorm:"serializer:json"`

	//Persyaratan tidak terpenuhi
	PersyaratanTidakSesuai []string `gorm:"serializer:json"`

	//Data Persetujuan Dokumen
	PersetujuanDokumen           string `gorm:"type:varchar(255);"`
	KeteranganPersetujuanDokumen *string

	//Data BAP
	NomerBAPDasar       string `gorm:"type:varchar(255);"`
	NomerBAPPelaksanaan string `gorm:"type:varchar(255);"`
	TanggalBAP          string `gorm:"type:varchar(255);"`
	FileBAP             []byte

	//Data SK
	FileSK []byte
}

type Perlalin struct {
	IdAndalalin            uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	IdUser                 uuid.UUID `gorm:"type:varchar(255);not null"`
	IdPetugas              uuid.UUID `gorm:"type:varchar(255);"`
	JenisAndalalin         string    `gorm:"type:varchar(255);not null"`
	Kategori               string    `gorm:"type:varchar(255);not null"`
	Jenis                  string    `gorm:"type:varchar(255);not null"`
	Kode                   string    `gorm:"type:varchar(255);not null"`
	NikPemohon             string    `gorm:"type:varchar(255);not null"`
	NamaPemohon            string    `gorm:"type:varchar(255);not null"`
	EmailPemohon           string    `gorm:"type:varchar(255);not null"`
	TempatLahirPemohon     string    `gorm:"type:varchar(255);not null"`
	TanggalLahirPemohon    string    `gorm:"type:varchar(255);not null"`
	AlamatPemohon          string    `gorm:"type:varchar(255);not null"`
	JenisKelaminPemohon    string    `sql:"type:enum('Laki-laki', 'Perempuan');not null"`
	NomerPemohon           string    `gorm:"type:varchar(255);not null"`
	NomerSelulerPemohon    string    `gorm:"type:varchar(255);not null"`
	JenisKegiatan          string    `gorm:"type:varchar(255);not null"`
	Peruntukan             string    `gorm:"type:varchar(255);not null"`
	LuasLahan              string    `gorm:"type:varchar(255);not null"`
	AlamatPersil           string    `gorm:"type:varchar(255);not null"`
	KelurahanPersil        string    `gorm:"type:varchar(255);not null"`
	LokasiPengambilan      string    `gorm:"type:varchar(255);not null"`
	WaktuAndalalin         string    `gorm:"not null"`
	TanggalAndalalin       string    `gorm:"not null"`
	StatusAndalalin        string    `sql:"type:enum('Cek persyaratan', 'Persyaratan tidak terpenuhi', 'Persyaratan terpenuhi', 'Survei lapangan', 'Laporan survei', 'Menunggu hasil keputusan', 'Tunda pemasangan', 'Pemasangan sedang dilakukan', 'Permohonan dibatalkan', 'Permohonan selesai')"`
	NamaPetugas            string    `gorm:"type:varchar(255);"`
	EmailPetugas           string    `gorm:"type:varchar(255);"`
	TandaTerimaPendaftaran []byte

	KartuTandaPenduduk []byte
	SuratPermohonan    []byte

	PersyaratanTambahan []PersyaratanTambahanPermohonan `gorm:"serializer:json"`

	//Persyaratan tidak terpenuhi
	PersyaratanTidakSesuai []string `gorm:"serializer:json"`

	LaporanSurvei []byte

	Tindakan             string
	PertimbanganTindakan string
}

type PersyaratanTambahanPermohonan struct {
	Persyaratan string
	Berkas      []byte
}

type InputAndalalin struct {
	KategoriJenisRencanaPembangunan string `json:"kategori" binding:"required"`
	JenisRencanaPembangunan         string `json:"jenis_rencana_pembangunan" binding:"required"`
	NikPemohon                      string `json:"nik_pemohon" binding:"required"`
	TempatLahirPemohon              string `json:"tempat_lahir_pemohon" binding:"required"`
	TanggalLahirPemohon             string `json:"tanggal_lahir_pemohon" binding:"required"`
	AlamatPemohon                   string `json:"alamat_pemohon" binding:"required"`
	JenisKelaminPemohon             string `json:"jenis_kelamin_pemohon" binding:"required"`
	NomerPemohon                    string `json:"nomer_pemohon" binding:"required"`
	NomerSelulerPemohon             string `json:"nomer_seluler_pemohon" binding:"required"`
	JabatanPemohon                  string `json:"jabatan_pemohon" binding:"required"`
	LokasiPengambilan               string `json:"lokasi_pengambilan" binding:"required"`
	NamaPerusahaan                  string `json:"nama_perusahaan" binding:"required"`
	AlamatPerusahaan                string `json:"alamat_perusahaan" binding:"required"`
	NomerPerusahaan                 string `json:"nomer_perusahaan" binding:"required"`
	EmailPerusahaan                 string `json:"email_perusahaan" binding:"required"`
	ProvinsiPerusahaan              string `json:"provinsi_perusahaan" binding:"required"`
	KabupatenPerusahaan             string `json:"kabupaten_perusahaan" binding:"required"`
	KecamatanPerusahaan             string `json:"kecamatan_perusahaan" binding:"required"`
	KelurahaanPerusahaan            string `json:"kelurahan_perusahaan" binding:"required"`
	NamaPimpinan                    string `json:"nama_pimpinan" binding:"required"`
	JabatanPimpinan                 string `json:"jabatan_pimpinan" binding:"required"`
	JenisKelaminPimpinan            string `json:"jenis_kelamin_pimpinan" binding:"required"`
	JenisKegiatan                   string `json:"jenis_kegiatan" binding:"required"`
	Peruntukan                      string `json:"peruntukan" binding:"required"`
	LuasLahan                       string `json:"luas_lahan" binding:"required"`
	AlamatPersil                    string `json:"alamat_persil" binding:"required"`
	KelurahanPersil                 string `json:"kelurahan_persil" binding:"required"`
	NomerSKRK                       string `json:"nomer_skrk" binding:"required"`
	TanggalSKRK                     string `json:"tanggal_skrk" binding:"required"`
}

type InputPerlalin struct {
	Kategori            string `json:"kategori" binding:"required"`
	Jenis               string `json:"jenis_perlengkapan" binding:"required"`
	NikPemohon          string `json:"nik_pemohon" binding:"required"`
	TempatLahirPemohon  string `json:"tempat_lahir_pemohon" binding:"required"`
	TanggalLahirPemohon string `json:"tanggal_lahir_pemohon" binding:"required"`
	AlamatPemohon       string `json:"alamat_pemohon" binding:"required"`
	JenisKelaminPemohon string `json:"jenis_kelamin_pemohon" binding:"required"`
	NomerPemohon        string `json:"nomer_pemohon" binding:"required"`
	NomerSelulerPemohon string `json:"nomer_seluler_pemohon" binding:"required"`
	LokasiPengambilan   string `json:"lokasi_pengambilan" binding:"required"`
	JenisKegiatan       string `json:"jenis_kegiatan" binding:"required"`
	Peruntukan          string `json:"peruntukan" binding:"required"`
	LuasLahan           string `json:"luas_lahan" binding:"required"`
	AlamatPersil        string `json:"alamat_persil" binding:"required"`
	KelurahanPersil     string `json:"kelurahan_persil" binding:"required"`
}

type DataAndalalin struct {
	Andalalin InputAndalalin `form:"data"`
}

type DataPerlalin struct {
	Perlalin InputPerlalin `form:"data"`
}

type DaftarAndalalinResponse struct {
	IdAndalalin      uuid.UUID `json:"id_andalalin,omitempty"`
	Kode             string    `json:"kode_andalalin,omitempty"`
	TanggalAndalalin string    `json:"tanggal_andalalin,omitempty"`
	Nama             string    `json:"nama_pemohon,omitempty"`
	Alamat           string    `json:"alamat_pemohon,omitempty"`
	JenisAndalalin   string    `json:"jenis_andalalin,omitempty"`
	StatusAndalalin  string    `json:"status_andalalin,omitempty"`
}

type AndalalinResponse struct {
	//Data Pemohon
	IdAndalalin            uuid.UUID `json:"id_andalalin,omitempty"`
	JenisAndalalin         string    `json:"jenis_andalalin,omitempty"`
	Kategori               string    `json:"kategori,omitempty"`
	Jenis                  string    `json:"jenis_rencana_pembangunan,omitempty"`
	Kode                   string    `json:"kode_andalalin,omitempty"`
	NikPemohon             string    `json:"nik_pemohon,omitempty"`
	NamaPemohon            string    `json:"nama_pemohon,omitempty"`
	EmailPemohon           string    `json:"email_pemohon,omitempty"`
	TempatLahirPemohon     string    `json:"tempat_lahir_pemohon,omitempty"`
	TanggalLahirPemohon    string    `json:"tanggal_lahir_pemohon,omitempty"`
	AlamatPemohon          string    `json:"alamat_pemohon,omitempty"`
	JenisKelaminPemohon    string    `json:"jenis_kelamin_pemohon,omitempty"`
	NomerPemohon           string    `json:"nomer_pemohon,omitempty"`
	NomerSelulerPemohon    string    `json:"nomer_seluler_pemohon,omitempty"`
	JabatanPemohon         string    `json:"jabatan_pemohon,omitempty"`
	LokasiPengambilan      string    `json:"lokasi_pengambilan,omitempty"`
	WaktuAndalalin         string    `json:"waktu_andalalin,omitempty"`
	TanggalAndalalin       string    `json:"tanggal_andalalin,omitempty"`
	StatusAndalalin        string    `json:"status_andalalin,omitempty"`
	TandaTerimaPendaftaran []byte    `json:"tanda_terima_pendaftaran,omitempty"`

	//Data Perusahaan
	NamaPerusahaan       string `json:"nama_perusahaan,omitempty"`
	AlamatPerusahaan     string `json:"alamat_perusahaan,omitempty"`
	NomerPerusahaan      string `json:"nomer_perusahaan,omitempty"`
	EmailPerusahaan      string `json:"email_perusahaan,omitempty"`
	ProvinsiPerusahaan   string `json:"perovinsi_perusahaan,omitempty"`
	KabupatenPerusahaan  string `json:"kabupaten_perusahaan,omitempty"`
	KecamatanPerusahaan  string `json:"kecamatan_perusahaan,omitempty"`
	KelurahaanPerusahaan string `json:"kelurahan_perusahaan,omitempty"`
	NamaPimpinan         string `json:"nama_pimpinan,omitempty"`
	JabatanPimpinan      string `json:"jabatan_pimpinan,omitempty"`
	JenisKelaminPimpinan string `json:"jenis_kelamin,omitempty"`
	JenisKegiatan        string `json:"jenis_kegiatan,omitempty"`
	Peruntukan           string `json:"peruntukan,omitempty"`
	LuasLahan            string `json:"luas_lahan,omitempty"`
	AlamatPersil         string `json:"alamat_persil,omitempty"`
	KelurahanPersil      string `json:"kelurahan_persil,omitempty"`
	NomerSKRK            string `json:"nomer_skrk,omitempty"`
	TanggalSKRK          string `json:"tanggal_skrk,omitempty"`

	//Data Persyaratan
	KartuTandaPenduduk []byte `json:"ktp,omitempty"`
	AktaPendirianBadan []byte `json:"akta_pendirian_badan,omitempty"`
	SuratKuasa         []byte `json:"surat_kuasa,omitempty"`

	//Persyaratan tidak terpenuhi
	PersyaratanTidakSesuai []string `json:"persyaratan_tidak_sesuai,omitempty"`

	//Data Petugas
	IdPetugas         uuid.UUID `json:"id_petugas,omitempty"`
	NamaPetugas       string    `json:"nama_petugas,omitempty"`
	EmailPetugas      string    `json:"email_petugas,omitempty"`
	StatusTiketLevel2 string    `json:"status_tiket,omitempty"`

	//Data Persetujuan
	PersetujuanDokumen           string  `json:"persetujuan,omitempty"`
	KeteranganPersetujuanDokumen *string `json:"keterangan_persetujuan,omitempty"`

	//Data BAP
	NomerBAPDasar       string `json:"nomer_bap_dasar,omitempty"`
	NomerBAPPelaksanaan string `json:"nomer_bap_pelaksanaan,omitempty"`
	TanggalBAP          string `json:"tanggal_bap,omitempty"`
	FileBAP             []byte `json:"file_bap,omitempty"`

	//Data SK
	FileSK []byte `json:"file_sk,omitempty"`

	PersyaratanTambahan []PersyaratanTambahanPermohonan `json:"persyaratan_tambahan,omitempty"`
}

type PerlalinResponse struct {
	IdAndalalin            uuid.UUID `json:"id_andalalin,omitempty"`
	JenisAndalalin         string    `json:"jenis_andalalin,omitempty"`
	Kategori               string    `json:"kategori,omitempty"`
	Jenis                  string    `json:"jenis_rencana_pembangunan,omitempty"`
	Kode                   string    `json:"kode_andalalin,omitempty"`
	NikPemohon             string    `json:"nik_pemohon,omitempty"`
	NamaPemohon            string    `json:"nama_pemohon,omitempty"`
	EmailPemohon           string    `json:"email_pemohon,omitempty"`
	TempatLahirPemohon     string    `json:"tempat_lahir_pemohon,omitempty"`
	TanggalLahirPemohon    string    `json:"tanggal_lahir_pemohon,omitempty"`
	AlamatPemohon          string    `json:"alamat_pemohon,omitempty"`
	JenisKelaminPemohon    string    `json:"jenis_kelamin_pemohon,omitempty"`
	NomerPemohon           string    `json:"nomer_pemohon,omitempty"`
	NomerSelulerPemohon    string    `json:"nomer_seluler_pemohon,omitempty"`
	LokasiPengambilan      string    `json:"lokasi_pengambilan,omitempty"`
	WaktuAndalalin         string    `json:"waktu_andalalin,omitempty"`
	TanggalAndalalin       string    `json:"tanggal_andalalin,omitempty"`
	StatusAndalalin        string    `json:"status_andalalin,omitempty"`
	TandaTerimaPendaftaran []byte    `json:"tanda_terima_pendaftaran,omitempty"`
	JenisKegiatan          string    `json:"jenis_kegiatan,omitempty"`
	Peruntukan             string    `json:"peruntukan,omitempty"`
	LuasLahan              string    `json:"luas_lahan,omitempty"`
	AlamatPersil           string    `json:"alamat_persil,omitempty"`
	KelurahanPersil        string    `json:"kelurahan_persil,omitempty"`

	//Data Persyaratan
	KartuTandaPenduduk []byte `json:"ktp,omitempty"`
	SuratPermohonan    []byte `json:"surat_kuasa,omitempty"`

	//Persyaratan tidak terpenuhi
	PersyaratanTidakSesuai []string `json:"persyaratan_tidak_sesuai,omitempty"`

	//Data Petugas
	IdPetugas         uuid.UUID `json:"id_petugas,omitempty"`
	NamaPetugas       string    `json:"nama_petugas,omitempty"`
	EmailPetugas      string    `json:"email_petugas,omitempty"`
	StatusTiketLevel2 string    `json:"status_tiket,omitempty"`

	LaporanSurvei []byte `json:"laporan_survei,omitempty"`

	PersyaratanTambahan []PersyaratanTambahanPermohonan `json:"persyaratan_tambahan,omitempty"`

	Tindakan             string `json:"keputusan_hasil,omitempty"`
	PertimbanganTindakan string `json:"pertimbangan,omitempty"`
}

type AndalalinResponseUser struct {
	//Data Pemohon
	IdAndalalin             uuid.UUID `json:"id_andalalin,omitempty"`
	JenisAndalalin          string    `json:"jenis_andalalin,omitempty"`
	Kode                    string    `json:"kode_andalalin,omitempty"`
	NamaPemohon             string    `json:"nama_pemohon,omitempty"`
	LokasiPengambilan       string    `json:"lokasi_pengambilan,omitempty"`
	TanggalAndalalin        string    `json:"tanggal_andalalin,omitempty"`
	StatusAndalalin         string    `json:"status_andalalin,omitempty"`
	TandaTerimaPendaftaran  []byte    `json:"tanda_terima_pendaftaran,omitempty"`
	JenisRencanaPembangunan string    `json:"jenis_rencana_pembangunan,omitempty"`
	Kategori                string    `json:"kategori,omitempty"`

	//Data Perusahaan
	NamaPerusahaan string `json:"nama_perusahaan,omitempty"`
	JenisKegiatan  string `json:"jenis_kegiatan,omitempty"`
	Peruntukan     string `json:"peruntukan,omitempty"`
	LuasLahan      string `json:"luas_lahan,omitempty"`

	//Persyaratan tidak terpenuhi
	PersyaratanTidakSesuai []string `json:"persyaratan_tidak_sesuai,omitempty"`

	//Data SK
	FileSK []byte `json:"file_sk,omitempty"`
}

type PersayaratanTidakSesuaiInput struct {
	Persyaratan []string `json:"persyaratan" binding:"required"`
}

type Persetujuan struct {
	Persetujuan string  `json:"persetujuan" binding:"required"`
	Keterangan  *string `json:"keterangan" binding:"required"`
}

type KeputusanHasil struct {
	Keputusan    string `json:"keputusan" binding:"required"`
	Pertimbangan string `json:"pertimbangan" binding:"required"`
}

type InputBAP struct {
	NomerBAPDasar       string `json:"nomer_dasar" binding:"required"`
	NomerBAPPelaksanaan string `json:"nomer_pelaksanaan" binding:"required"`
	TanggalBAP          string `json:"tanggal" binding:"required"`
}

type BAPData struct {
	Data InputBAP `form:"data"`
}

type TambahPetugas struct {
	IdPetugas    uuid.UUID `json:"id_petugas" binding:"required"`
	NamaPetugas  string    `json:"nama_petugas" binding:"required"`
	EmailPetugas string    `json:"email_petugas" binding:"required"`
}

type Survei struct {
	IdSurvey      uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	IdAndalalin   uuid.UUID `gorm:"type:varchar(255);uniqueIndex;not null"`
	IdTiketLevel1 uuid.UUID `gorm:"type:varchar(255);not null"`
	IdTiketLevel2 uuid.UUID `gorm:"type:varchar(255);not null"`
	IdPetugas     uuid.UUID `gorm:"type:varchar(255);not null"`
	Petugas       string    `gorm:"type:varchar(255);not null"`
	EmailPetugas  string    `gorm:"type:varchar(255);not null"`
	Keterangan    *string
	Foto1         []byte
	Foto2         []byte
	Foto3         []byte
	Lokasi        string
	Latitude      float64
	Longitude     float64
	WaktuSurvei   string `gorm:"not null"`
	TanggalSurvei string `gorm:"not null"`
}

type SurveiMandiri struct {
	IdSurvey           uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	IdPetugas          uuid.UUID `gorm:"type:varchar(255);not null"`
	Petugas            string    `gorm:"type:varchar(255);not null"`
	EmailPetugas       string    `gorm:"type:varchar(255);not null"`
	Keterangan         *string
	Foto1              []byte
	Foto2              []byte
	Foto3              []byte
	Lokasi             string
	Latitude           float64
	Longitude          float64
	WaktuSurvei        string `gorm:"not null"`
	TanggalSurvei      string `gorm:"not null"`
	StatusSurvei       string
	KeteranganTindakan string
}

type InputSurvey struct {
	Lokasi     string  `json:"lokasi" binding:"required"`
	Keterangan *string `json:"keterangan" binding:"required"`
	Latitude   float64 `protobuf:"fixed64,1,opt,name=latitude,proto3" json:"latitude" binding:"required"`
	Longitude  float64 `protobuf:"fixed64,2,opt,name=longitude,proto3" json:"longtitude" binding:"required"`
}

type DataSurvey struct {
	Data InputSurvey `form:"data"`
}

type TiketLevel1 struct {
	IdTiketLevel1 uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	IdAndalalin   uuid.UUID `gorm:"type:varchar(255);uniqueIndex;not null"`
	Status        string    `sql:"type:enum('Buka', 'Tutup', 'Tunda', 'Batal');not null"`
}

type TiketLevel2 struct {
	IdTiketLevel2 uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	IdTiketLevel1 uuid.UUID `gorm:"type:varchar(255);not null"`
	IdAndalalin   uuid.UUID `gorm:"type:varchar(255);not null"`
	IdPetugas     uuid.UUID `gorm:"type:varchar(255);not null"`
	Status        string    `sql:"type:enum('Buka', 'Tutup', 'Tunda', 'Batal');not null"`
}

type UsulanPengelolaan struct {
	IdUsulan                   uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	IdAndalalin                uuid.UUID `gorm:"type:varchar(255);uniqueIndex;not null"`
	IdTiketLevel1              uuid.UUID `gorm:"type:varchar(255);not null"`
	IdTiketLevel2              uuid.UUID `gorm:"type:varchar(255);not null"`
	IdPengusulTindakan         uuid.UUID `gorm:"type:varchar(255);not null"`
	NamaPengusulTindakan       string    `gorm:"type:varchar(255);not null"`
	PertimbanganUsulanTindakan string    `gorm:"type:varchar(255);not null"`
	KeteranganUsulanTindakan   *string   `gorm:"type:varchar(255);not null"`
}

type InputUsulanPengelolaan struct {
	PertimbanganUsulanTindakan string  `json:"pertimbangan" binding:"required"`
	KeteranganUsulanTindakan   *string `json:"keterangan" binding:"required"`
}

type SurveiKepuasan struct {
	IdSurvey           uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	IdAndalalin        uuid.UUID `gorm:"type:varchar(255);not null"`
	IdUser             uuid.UUID `gorm:"type:varchar(255);not null"`
	Nama               string    `gorm:"type:varchar(255);not null"`
	Email              string    `gorm:"type:varchar(255);not null"`
	KritikSaran        *string
	TanggalPelaksanaan string
	DataSurvei         []Kepuasan `gorm:"serializer:json"`
}

type SurveiKepuasanInput struct {
	KritikSaran *string    `json:"saran" binding:"required"`
	DataSurvei  []Kepuasan `json:"data" binding:"required"`
}

type Kepuasan struct {
	Jenis string
	Nilai string
}

type Pemasangan struct {
	IdPemasangan      uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	IdAndalalin       uuid.UUID `gorm:"type:varchar(255);uniqueIndex;not null"`
	IdTiketLevel1     uuid.UUID `gorm:"type:varchar(255);not null"`
	IdPetugas         uuid.UUID `gorm:"type:varchar(255);not null"`
	Petugas           string    `gorm:"type:varchar(255);not null"`
	EmailPetugas      string    `gorm:"type:varchar(255);not null"`
	Keterangan        *string
	Foto1             []byte
	Foto2             []byte
	Foto3             []byte
	Lokasi            string
	Latitude          float64
	Longitude         float64
	WaktuPemasangan   string `gorm:"not null"`
	TanggalPemasangan string `gorm:"not null"`
}
