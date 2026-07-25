package models

// ProjeDurum, proje durum_adi değerleri için sabitler.
// Türkçe Yorum: Bu sabitler veritabanındaki proje_durum.durum_adi sütunuyla bire bir eşleşmelidir.
// Kodun herhangi bir yerinde ham string ("incelemede" gibi) kullanmak yerine bu sabitler kullanılmalıdır.
const (
	// DurumTaslak: Proje henüz gönderilmemiş, düzenleme aşamasında.
	DurumTaslak = "taslak"

	// DurumIncelemede: TTO ön inceleme aşamasında.
	DurumIncelemede = "incelemede"

	// DurumDekanOnayiBekliyor: Dekan onayı bekleniyor.
	DurumDekanOnayiBekliyor = "dekan_onayi_bekliyor"

	// DurumDekanOnayladi: Dekan onayladı, TTO komisyona sevk edecek.
	DurumDekanOnayladi = "dekan_onayladi"

	// DurumKomisyonBekliyor: Komisyon oylaması bekleniyor.
	DurumKomisyonBekliyor = "komisyon_bekliyor"

	// DurumKomisyonOnayladi: Komisyon onayladı, hakem ataması/sözleşme aşaması.
	DurumKomisyonOnayladi = "komisyon_onayladi"

	// DurumHakemAtamaBekliyor: Hakem ataması TTO tarafından yapılacak.
	DurumHakemAtamaBekliyor = "hakem_atama_bekliyor"

	// DurumHakemBekliyor: Hakem değerlendirmesi bekleniyor.
	DurumHakemBekliyor = "hakem_bekliyor"

	// DurumHakemOnayladi: Hakem onayladı, TTO sözleşmeye sevk edecek.
	DurumHakemOnayladi = "hakem_onayladi"

	// DurumSozlesmeImza: Sözleşme imzalanmak üzere bekliyor.
	DurumSozlesmeImza = "sozlesme_imza"

	// DurumTTOAktif: (Eski uyumluluk) TTO aktif etmiş.
	DurumTTOAktif = "tto_aktif"

	// DurumYururlukte: Proje aktif, yürütülüyor.
	DurumYururlukte = "yururlukte"

	// DurumTamamlandi: Proje başarıyla tamamlandı.
	DurumTamamlandi = "tamamlandi"

	// DurumReddedildi: Proje herhangi bir aşamada reddedildi.
	DurumReddedildi = "reddedildi"

	// DurumRevizyon: Proje revizyon talep edildi, başvuru sahibi bekliyor.
	DurumRevizyon = "revizyon"
)

// ProjeAksiyon, workflow işlem aksiyon adları için sabitler.
// Türkçe Yorum: ProcessWorkflowAction fonksiyonuna iletilen aksiyon string değerleri.
const (
	AksiyonOnayla        = "onayla"
	AksiyonOnaylaHakemsiz = "onayla_hakemsiz"
	AksiyonReddet        = "reddet"
	AksiyonRevizyon      = "revizyon"
	AksiyonTamamla       = "tamamla"
)

// SistemRol, sistem_rol_tanimlama.rol_adi değerleri için sabitler.
// Türkçe Yorum: Rol kontrolleri her yerde farklı string kullanmak yerine bu sabitlerle yapılmalıdır.
const (
	RolAdmin           = "admin"
	RolTTO             = "tto"
	RolDekan           = "dekan"
	RolKomisyon        = "komisyon"
	RolKomisyonBaskani = "komisyon_baskani"
	RolHakem           = "hakem"
	RolAkademisyen     = "akademisyen"
	RolOgrenci         = "ogrenci"
)
