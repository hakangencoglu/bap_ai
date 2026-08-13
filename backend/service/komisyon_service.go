package service

import (
	"fmt"
	"time"

	"bap_ai/backend/models"
	"bap_ai/backend/repository"
)

// KomisyonService komisyon toplantı süreçlerini yönetir.
// Türkçe Yorum: Toplantı numarasının otomatik oluşturulması ve kayıt öncesi bütçe/yetki doğrulama kontrolleri burada yapılır.
type KomisyonService struct {
	KomisyonRepo *repository.KomisyonRepository
}

// NewKomisyonService yeni bir KomisyonService oluşturur.
func NewKomisyonService(komisyonRepo *repository.KomisyonRepository) *KomisyonService {
	return &KomisyonService{KomisyonRepo: komisyonRepo}
}

// GetCommissionMembers komisyon üye listesini döner.
// Türkçe Yorum: Komisyon paneline üye yoklama listesini basabilmek için ilgili repository metodunu çağırır.
func (s *KomisyonService) GetCommissionMembers() ([]*models.UyeWithDetay, error) {
	return s.KomisyonRepo.GetCommissionMembers()
}

// GetNextMeetingNumber seçilen tarihe göre bir sonraki toplantı numarasını belirler.
// Türkçe Yorum: Gönderilen tarih verisinin yılını ayrıştırır ve veritabanı sayacına göre sıradaki toplantı numarasını 'YIL/SIRA' (Örn: 2026/004) formatında üretir.
func (s *KomisyonService) GetNextMeetingNumber(meetingDate time.Time) (string, error) {
	year := meetingDate.Year()
	count, err := s.KomisyonRepo.GetMeetingsCountByYear(year)
	if err != nil {
		return "", err
	}
	// Sıra numarasını 3 haneli dolduracak şekilde formatla (örn: 2026/001)
	nextNo := fmt.Sprintf("%d/%03d", year, count+1)
	return nextNo, nil
}

// CreateMeeting yeni bir toplantı oluşturur.
// Türkçe Yorum: Gerekli alanların doğruluğunu denetler, otomatik numara üretir ve toplantıyı kaydeder.
func (s *KomisyonService) CreateMeeting(meeting *models.KomisyonToplantisi) error {
	if meeting.Tarih.IsZero() {
		return fmt.Errorf("toplantı tarihi belirtilmelidir")
	}
	if meeting.Gundem == "" {
		return fmt.Errorf("toplantı gündemi boş olamaz")
	}
	if meeting.Karar == "" {
		return fmt.Errorf("toplantı kararı boş olamaz")
	}
	if len(meeting.Katilimcilar) == 0 {
		return fmt.Errorf("toplantıya en az bir katılımcı eklenmelidir")
	}

	// Otomatik toplantı numarası oluştur
	nextNo, err := s.GetNextMeetingNumber(meeting.Tarih)
	if err != nil {
		return fmt.Errorf("toplantı numarası oluşturulamadı: %w", err)
	}
	meeting.ToplantiNo = nextNo

	return s.KomisyonRepo.CreateMeeting(meeting)
}

// GetMeetingByID belirtilen ID'ye sahip toplantıyı döner.
// Türkçe Yorum: Toplantı detaylarını ve katılımcı listesini getirir.
func (s *KomisyonService) GetMeetingByID(id int) (*models.KomisyonToplantisi, error) {
	return s.KomisyonRepo.GetMeetingByID(id)
}

// ListMeetings tüm toplantıları listeler.
// Türkçe Yorum: Rapor geçmişini komisyon başkanına sunmak için tüm toplantı kayıtlarını çeker.
func (s *KomisyonService) ListMeetings() ([]*models.KomisyonToplantisi, error) {
	return s.KomisyonRepo.ListMeetings()
}

// DeleteMeeting toplantıyı kalıcı siler (yalnızca super-delete).
func (s *KomisyonService) DeleteMeeting(toplantiID int) error {
	return s.KomisyonRepo.DeleteMeeting(toplantiID)
}
