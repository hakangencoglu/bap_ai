package service

import (
	"fmt"
	"strings"
	"time"

	"bap_ai/backend/models"
	"bap_ai/backend/repository"
)

// pageAccessChecker mutabakat yetkisi için sayfa erişim kontrolü arayüzüdür.
type pageAccessChecker interface {
	CheckPageAccess(roles []string, path string) (bool, error)
}

// SatinalmaService yapısı, satın alma işlemlerine ait iş mantığını yönetir.
// Türkçe Yorum: Satın alma talepleri oluşturulurken bütçe kalemi limit kontrolü ve proje durum doğrulaması yapan servis katmanıdır.
type SatinalmaService struct {
	SatinalmaRepo    *repository.SatinalmaRepository
	ProjeRepo        *repository.ProjeRepository
	PageAccess       pageAccessChecker
	OnPurchaseAction func(talepID int, eventType string, islemYapanID int)
}

// MutabakatSayfaYolu admin'in TTO'ya verebileceği mutabakat yetki URL'sidir.
const MutabakatSayfaYolu = "/tto/satinalma/mutabakat"

// NewSatinalmaService yeni bir SatinalmaService nesnesi oluşturur.
// Türkçe Yorum: SatinalmaService için dependency injection kurucusu.
func NewSatinalmaService(satinalmaRepo *repository.SatinalmaRepository, projeRepo *repository.ProjeRepository) *SatinalmaService {
	return &SatinalmaService{
		SatinalmaRepo: satinalmaRepo,
		ProjeRepo:     projeRepo,
	}
}

// CreatePurchaseRequests yeni bir satın alma talebi grubu (toplu talep) oluşturur.
// Türkçe Yorum: Proje yetki kontrolü yapar, proje durumunun aktif (yururlukte) olduğunu ve talep edilen toplam tutarın ilgili bütçe kalemindeki rezerve (onaylı + bekleyen) bakiyeyi aşmadığını denetler.
func (s *SatinalmaService) CreatePurchaseRequests(reqs []*models.SatinalmaTalebi, requestorRole string) error {
	if len(reqs) == 0 {
		return fmt.Errorf("en az bir satın alma kalemi gönderilmelidir")
	}

	firstReq := reqs[0]

	// Yetkilendirme Kontrolü: Admin dışındaki tüm kullanıcıların projenin ekibinde olması zorunludur.
	isAdmin := false
	for _, r := range strings.Split(requestorRole, ",") {
		if strings.TrimSpace(r) == "admin" {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		isUye, err := s.ProjeRepo.IsProjeUyesi(firstReq.ProjeID, firstReq.UyeID)
		if err != nil {
			return fmt.Errorf("proje yetki kontrolü yapılamadı: %w", err)
		}
		if !isUye {
			return fmt.Errorf("bu proje için satın alma talebi oluşturma yetkiniz bulunmamaktadır")
		}
	}

	// 1. Projeyi sorgula ve durumunu kontrol et
	proje, err := s.ProjeRepo.GetProjeByID(firstReq.ProjeID)
	if err != nil {
		return fmt.Errorf("proje bilgisi alınamadı: %w", err)
	}

	// Türkçe Yorum: TTO onayından geçerek aktifleşmiş projelerin durumu 'yururlukte' olmalıdır.
	if proje.DurumAdi != "yururlukte" {
		return fmt.Errorf("satın alma talebi sadece TTO tarafından onaylanmış ve sözleşmesi imzalanmış (aktif) projeler için yapılabilir")
	}

	// 2. Rezerve bütçe (taahhüt + fiili + bekleyen) limit kontrolünü yap
	kalanButce, err := s.SatinalmaRepo.GetReservedBudget(firstReq.ProjeID, firstReq.KalemID)
	if err != nil {
		return fmt.Errorf("rezerve bütçe bilgisi sorgulanamadı: %w", err)
	}

	// Toplam talep tutarını hesapla
	var toplamTalepTutar float64
	for _, req := range reqs {
		toplamTalepTutar += float64(req.Miktar) * req.BirimFiyat
	}

	if toplamTalepTutar > kalanButce {
		return fmt.Errorf("talep edilen toplam tutar (%.2f ₺), bu bütçe kaleminin onaylanmış ve bekleyen taleplerden kalan limitini (%.2f ₺) aşmaktadır", toplamTalepTutar, kalanButce)
	}

	// 3. Talepleri veritabanına ekle
	err = s.SatinalmaRepo.CreatePurchaseRequests(reqs)
	if err == nil && s.OnPurchaseAction != nil {
		for _, req := range reqs {
			go s.OnPurchaseAction(req.TalepID, "create", req.UyeID)
		}
	}
	return err
}

// CreatePurchaseRequest yeni bir satın alma talebi oluşturur.
// Türkçe Yorum: Geriye dönük uyumluluk için tekli satın alma talebi ekleme isteklerini toplu ekleme metoduna yönlendirir.
func (s *SatinalmaService) CreatePurchaseRequest(req *models.SatinalmaTalebi, requestorRole string) error {
	return s.CreatePurchaseRequests([]*models.SatinalmaTalebi{req}, requestorRole)
}

// GetPurchaseRequestsByProject bir projeye ait tüm talepleri listeler.
// Türkçe Yorum: Belirli bir proje altındaki tüm satın alma işlemlerini yetkilendirme kontrolü yaparak listeler.
func (s *SatinalmaService) GetPurchaseRequestsByProject(projeID int, requestorID int, requestorRole string) ([]models.SatinalmaTalebi, error) {
	// Yetkilendirme Kontrolü: Admin ve TTO rolleri dışındaki kullanıcıların proje üyesi olması zorunludur.
	hasAccess := false
	for _, r := range strings.Split(requestorRole, ",") {
		rClean := strings.TrimSpace(r)
		if rClean == "admin" || rClean == "tto" {
			hasAccess = true
			break
		}
	}

	if !hasAccess {
		isUye, err := s.ProjeRepo.IsProjeUyesi(projeID, requestorID)
		if err != nil {
			return nil, fmt.Errorf("proje yetki kontrolü yapılamadı: %w", err)
		}
		if !isUye {
			return nil, fmt.Errorf("bu projenin satın alma taleplerini görüntüleme yetkiniz bulunmamaktadır")
		}
	}

	return s.SatinalmaRepo.GetPurchaseRequestsByProject(projeID)
}

// GetAllPurchaseRequests tüm sistemdeki satın alma taleplerini listeler.
// Türkçe Yorum: TTO yetkilileri için tüm talepleri listeler.
func (s *SatinalmaService) GetAllPurchaseRequests() ([]models.SatinalmaTalebi, error) {
	return s.SatinalmaRepo.GetAllPurchaseRequests()
}

// UpdatePurchaseStatus satın alma talebini onaylar veya reddeder.
// Türkçe Yorum: TTO yetkilisinin verdiği karara göre talebi 'Onaylandı' veya 'Reddedildi' durumuna getirir. Onay durumunda bütçe limitini tekrar doğrular.
func (s *SatinalmaService) UpdatePurchaseStatus(talepID int, status string, redNedeni string, islemYapanID int) error {
	// 1. Talebi bul
	talep, err := s.SatinalmaRepo.GetPurchaseRequestByID(talepID)
	if err != nil {
		return fmt.Errorf("satın alma talebi bulunamadı: %w", err)
	}

	if talep.Durum != models.SatinalmaDurumBeklemede {
		return fmt.Errorf("sadece 'Beklemede' durumundaki satın alma talepleri güncellenebilir")
	}

	// 2. Eğer onaylanıyorsa kalan bütçeyi son bir kez daha kontrol et
	if status == models.SatinalmaDurumOnaylandi {
		effective := talep.EffectiveAmount()
		kalanButce, err := s.SatinalmaRepo.GetRemainingBudget(talep.ProjeID, talep.KalemID)
		if err != nil {
			return fmt.Errorf("bütçe kalemi kalan limiti doğrulanamadı: %w", err)
		}

		if effective > kalanButce {
			return fmt.Errorf("bu satın alma talebi onaylandığında bütçe kalem limiti (Kalan: %.2f ₺) aşılacaktır", kalanButce)
		}
	}

	// 3. Durumu güncelle
	err = s.SatinalmaRepo.UpdatePurchaseStatus(talepID, status, redNedeni)
	if err != nil {
		return err
	}

	// 4. Onayda grup kalemleri için rezervasyon ledger kaydı yaz
	if status == models.SatinalmaDurumOnaylandi {
		group, gErr := s.collectApprovalGroup(talep)
		if gErr == nil {
			tx, txErr := s.SatinalmaRepo.BeginTx()
			if txErr == nil {
				for _, item := range group {
					_ = s.SatinalmaRepo.RecordApprovalReservationTx(tx, &item, islemYapanID)
				}
				_ = tx.Commit()
			}
		}
	}

	if s.OnPurchaseAction != nil {
		go s.OnPurchaseAction(talepID, "update", islemYapanID)
	}
	return nil
}

// collectApprovalGroup aynı talep_no grubundaki kalemleri toplar.
func (s *SatinalmaService) collectApprovalGroup(talep *models.SatinalmaTalebi) ([]models.SatinalmaTalebi, error) {
	all, err := s.SatinalmaRepo.GetPurchaseRequestsByProject(talep.ProjeID)
	if err != nil {
		return nil, err
	}
	var group []models.SatinalmaTalebi
	for _, t := range all {
		if talep.TalepNo != "" && talep.TalepNo != "-" {
			if t.TalepNo == talep.TalepNo && t.Durum == models.SatinalmaDurumOnaylandi {
				group = append(group, t)
			}
		} else if t.TalepID == talep.TalepID {
			group = append(group, t)
		}
	}
	if len(group) == 0 {
		group = append(group, *talep)
	}
	return group, nil
}

// RevisePurchaseRequest satın alma talebinin fiyatını günceller ve gerekçe kaydeder.
// Türkçe Yorum: İstek yapan kullanıcının yetki durumunu kontrol eder (sadece admin veya tto).
// Ardından güncellenecek yeni fiyatın projenin kalan bütçe limitlerini aşmadığını doğrulayarak repository katmanına yansıtır.
func (s *SatinalmaService) RevisePurchaseRequest(talepID int, yeniBirimFiyat float64, gerekce string, yetkiliID int, requestorRole string) error {
	// 1. Yetki Kontrolü: Yalnızca TTO ve Admin rolleri bütçe revizyonu yapabilir.
	isAuthorized := false
	for _, r := range strings.Split(requestorRole, ",") {
		rClean := strings.TrimSpace(r)
		if rClean == "admin" || rClean == "tto" {
			isAuthorized = true
			break
		}
	}
	if !isAuthorized {
		return fmt.Errorf("bütçe kalemi fiyatını düzenleme yetkiniz bulunmamaktadır")
	}

	// 2. Talebi veritabanından çek
	talep, err := s.SatinalmaRepo.GetPurchaseRequestByID(talepID)
	if err != nil {
		return fmt.Errorf("satın alma talebi bulunamadı: %w", err)
	}

	if talep.Durum != models.SatinalmaDurumBeklemede {
		return fmt.Errorf("sadece 'Beklemede' durumundaki satın alma talepleri revize edilebilir")
	}

	if yeniBirimFiyat <= 0 {
		return fmt.Errorf("yeni birim fiyat sıfırdan büyük olmalıdır")
	}

	// 3. Projenin kalan bütçesini kontrol et
	mevcutTutar := talep.EffectiveAmount()
	yeniTutar := yeniBirimFiyat * float64(talep.Miktar)
	farkTutar := yeniTutar - mevcutTutar

	// Eğer yeni tutar eskisinden büyükse bütçe aşım kontrolü yapmalıyız
	if farkTutar > 0 {
		// Bu bütçe kalemindeki rezerve edilmemiş kalan limiti al
		kalanButceLimit, err := s.SatinalmaRepo.GetReservedBudget(talep.ProjeID, talep.KalemID)
		if err != nil {
			return fmt.Errorf("bütçe limiti doğrulanamadı: %w", err)
		}

		// kalanButceLimit zaten bu talebin MEVCUT tutarını da düşmüş durumda.
		// Bu nedenle, ek getireceğimiz farkTutar, kalanButceLimit'ten büyük olamaz.
		if farkTutar > kalanButceLimit {
			return fmt.Errorf("girdiğiniz yeni fiyat ile oluşacak fark tutar (%.2f ₺), bu bütçe kaleminin kalan limitini (%.2f ₺) aşmaktadır", farkTutar, kalanButceLimit)
		}
	}

	// 4. Güncellemeyi kaydet
	err = s.SatinalmaRepo.RevisePurchaseRequest(talepID, yeniBirimFiyat, gerekce, yetkiliID)
	if err == nil && s.OnPurchaseAction != nil {
		go s.OnPurchaseAction(talepID, "update", yetkiliID)
	}
	return err
}

// GetProjectBudgetReport projenin bütçe kalemi bazlı harcama raporunu döner.
func (s *SatinalmaService) GetProjectBudgetReport(projeID int) (*models.ProjeButceHarcamaRaporu, error) {
	if projeID <= 0 {
		return nil, fmt.Errorf("geçersiz proje ID'si")
	}
	return s.SatinalmaRepo.GetProjectBudgetReport(projeID)
}

// EnsureMutabakatYetkisi admin tarafından TTO'ya verilen mutabakat sayfa yetkisini doğrular.
// Türkçe Yorum: Admin her zaman yetkilidir; diğer roller sayfa_rol_yetki üzerinden kontrol edilir.
func (s *SatinalmaService) EnsureMutabakatYetkisi(requestorRole string) error {
	roles := []string{}
	for _, r := range strings.Split(requestorRole, ",") {
		r = strings.TrimSpace(r)
		if r != "" {
			roles = append(roles, r)
		}
	}
	for _, r := range roles {
		if r == models.RolAdmin {
			return nil
		}
	}
	if s.PageAccess == nil {
		// Fallback: TTO rolü varsa izin ver (PageAccess bağlanmamış ortamlarda)
		for _, r := range roles {
			if r == models.RolTTO {
				return nil
			}
		}
		return fmt.Errorf("satın alma mutabakat yetkiniz bulunmamaktadır")
	}
	ok, err := s.PageAccess.CheckPageAccess(roles, MutabakatSayfaYolu)
	if err != nil {
		return fmt.Errorf("mutabakat yetkisi doğrulanamadı: %w", err)
	}
	if !ok {
		return fmt.Errorf("satın alma mutabakat yetkiniz bulunmamaktadır; admin bu yetkiyi TTO rolüne vermelidir")
	}
	return nil
}

// ProcessMutabakat TTO'nun fiili tutar mutabakatını işler.
// Türkçe Yorum: Onaylı talebi kapatır (fiili tutar) veya iptal eder; fazla farkta kalan bütçe kontrolü yapar.
func (s *SatinalmaService) ProcessMutabakat(istek *models.MutabakatIstek, ttoUyeID int, requestorRole string) (*models.SatinalmaOdeme, error) {
	if err := s.EnsureMutabakatYetkisi(requestorRole); err != nil {
		return nil, err
	}

	karar := strings.ToLower(strings.TrimSpace(istek.Karar))
	if karar != models.MutabakatKararOnayla && karar != models.MutabakatKararIptal {
		return nil, fmt.Errorf("geçersiz karar; 'onayla' veya 'iptal' olmalıdır")
	}

	talep, err := s.SatinalmaRepo.GetPurchaseRequestByID(istek.TalepID)
	if err != nil {
		return nil, fmt.Errorf("satın alma talebi bulunamadı: %w", err)
	}
	if talep.Durum != models.SatinalmaDurumOnaylandi {
		return nil, fmt.Errorf("mutabakat yalnızca 'Onaylandı' durumundaki talepler için yapılabilir")
	}

	exists, err := s.SatinalmaRepo.HasApprovedOdeme(talep.TalepID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("bu talep için zaten onaylı bir mutabakat kaydı bulunmaktadır")
	}

	taahhut := talep.EffectiveAmount()
	fiili := istek.FiiliTutar
	yeniDurum := models.SatinalmaDurumKapatildi

	if karar == models.MutabakatKararIptal {
		fiili = 0
		yeniDurum = models.SatinalmaDurumIptalEdildi
		if strings.TrimSpace(istek.Gerekce) == "" {
			return nil, fmt.Errorf("iptal mutabakatı için gerekçe zorunludur")
		}
	} else {
		if fiili < 0 {
			return nil, fmt.Errorf("fiili tutar negatif olamaz")
		}
		// Fazla fark: kalan bütçe yeterli mi?
		if fiili > taahhut {
			kalan, kErr := s.SatinalmaRepo.GetReservedBudget(talep.ProjeID, talep.KalemID)
			if kErr != nil {
				return nil, fmt.Errorf("bütçe kontrolü yapılamadı: %w", kErr)
			}
			// Bu talep hâlâ açık taahhütte; GetReservedBudget taahhüdü düşmüş.
			// Fazla fark = fiili - taahhut; kullanılabilir kalan bu farkı karşılamalı.
			ekstra := fiili - taahhut
			if ekstra > kalan {
				return nil, fmt.Errorf(
					"fiili tutar taahhüdü %.2f ₺ aşıyor; kalan kullanılabilir bütçe %.2f ₺. Önce ek bütçe veya fasıl aktarımı talep edilmelidir",
					ekstra, kalan,
				)
			}
		}
		if fiili != taahhut && strings.TrimSpace(istek.Gerekce) == "" {
			return nil, fmt.Errorf("taahhüt ile fiili tutar farklıysa gerekçe zorunludur")
		}
	}

	fark := fiili - taahhut
	farkYonu := models.FarkYonuEsit
	if fark > 0 {
		farkYonu = models.FarkYonuFazla
	} else if fark < 0 {
		farkYonu = models.FarkYonuEksik
	}

	odeme := &models.SatinalmaOdeme{
		TalepID:       talep.TalepID,
		TalepNo:       talep.TalepNo,
		ProjeID:       talep.ProjeID,
		KalemID:       talep.KalemID,
		TaahhutTutari: taahhut,
		FiiliTutar:    fiili,
		FarkTutari:    fark,
		FarkYonu:      farkYonu,
		ParaBirimi:    "TRY",
		Durum:         models.OdemeDurumOnaylandi,
		TtoUyeID:      &ttoUyeID,
	}
	if g := strings.TrimSpace(istek.Gerekce); g != "" {
		odeme.TtoGerekce = &g
	}
	if f := strings.TrimSpace(istek.FaturaNo); f != "" {
		odeme.FaturaNo = &f
	}
	if d := parseOptionalDate(istek.FaturaTarihi); d != nil {
		odeme.FaturaTarihi = d
	}
	if d := parseOptionalDate(istek.OdemeTarihi); d != nil {
		odeme.OdemeTarihi = d
	}

	tx, err := s.SatinalmaRepo.BeginTx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if err = s.SatinalmaRepo.CreateMutabakatTx(tx, odeme, yeniDurum, ttoUyeID); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	if s.OnPurchaseAction != nil {
		go s.OnPurchaseAction(talep.TalepID, "mutabakat", ttoUyeID)
	}
	return odeme, nil
}

// parseOptionalDate YYYY-MM-DD stringini time.Time'a çevirir.
func parseOptionalDate(s string) *time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil
	}
	return &t
}

// ListPendingMutabakat mutabakat bekleyen onaylı talepleri listeler.
func (s *SatinalmaService) ListPendingMutabakat(requestorRole string) ([]models.SatinalmaTalebi, error) {
	if err := s.EnsureMutabakatYetkisi(requestorRole); err != nil {
		return nil, err
	}
	return s.SatinalmaRepo.ListPendingMutabakat()
}

// ListOdemelerByProje proje ödeme/mutabakat kayıtlarını listeler.
func (s *SatinalmaService) ListOdemelerByProje(projeID int) ([]models.SatinalmaOdeme, error) {
	if projeID <= 0 {
		return nil, fmt.Errorf("geçersiz proje ID'si")
	}
	return s.SatinalmaRepo.ListOdemelerByProje(projeID)
}
