# Komisyon Toplantısı ↔ Proje İlişkilendirme — Görev Listesi

## DB Katmanı
- [x] `schema.sql` → `komisyon_toplantisi` tablosuna `durum` kolonu ekle
- [x] `schema.sql` → `komisyon_toplanti_proje` köprü tablosunu ekle
- [x] `migration.go` → Mevcut DB için migration bloğu ekle

## Backend — Model
- [x] `models/komisyon_toplanti.go` oluştur

## Backend — Repository
- [x] `repository/komisyon_toplanti_repository.go` oluştur
  - [x] `AddProjeToToplanti`
  - [x] `RemoveProjeFromToplanti`
  - [x] `GetProjectsByToplanti`
  - [x] `GetToplantilerByProje`
  - [x] `SetProjeKarar`
  - [x] `GetToplantiBelgeDetay` (PDF için)

## Backend — Service
- [x] `service/komisyon_toplanti_service.go` oluştur

## Backend — Handler
- [x] `api/komisyon_toplanti_handler.go` oluştur
  - [x] Tüm CRUD + karar endpointleri

## Backend — PDF
- [x] `service/pdf_service.go` → `GenerateToplantTutanakPDF` ekle

## Backend — main.go
- [x] Yeni handler'ı kaydet ve route'ları ekle

## Frontend
- [x] `komisyon_baskani_dashboard.html` → Toplantı yönetimi (oluştur, proje ekle, karar, PDF)
- [x] `komisyon_dashboard.html` → Raportör için toplantı listesi + karar ver
