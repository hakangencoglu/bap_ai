# AGENTS.md


## Proje hakkında

Bu doküman içerisinde geliştirilecek proje kapsamında kullanılacak yapay zeka ajanlarının isimleri, nitelikleri ve kullanmaları gereken dosya yapılarını tanıtılacaktır. 
Bu projenin amacı ise üniversitelerde kullanılan bilimsel araştırma projesi (BAP) için gerekli sistemin geliştirilmesidir. 

Bu projede ana konular aşağıdaki gibi olacaktır.
- **accounts**: Kullanıcı doğrulaması (giriş yap, kayıt ol, çıkış yap, rol ataması ...)
- **web sayfaları**: Karşılama sayfası ve ilgili BAP türüne göre BAP başvurusu yapılması
- **önizleme ve otomatik kaydetme**: belirli sayfalarda yapılan değişiklikler otomatik kayıt edilecek ve en sonunda pdf üzerinde önizleme olarak kullanıcıya PDF formatında gösterilecek.

## Agent list
- isim: [Frontend Dev]
- amaç: Yapılacak web projesindeki arayüzleri tasarlanacak.
- teknolojiler: arayüz tasarımları için basit ve dinamik html/css yapısıyla çalışılacak.

- isim: [Backend Dev]
- amaç: Yapılacak web projesindeki backend üzerinde çalışacak.
- teknolojiler: backendde golang güncel en stabil versiyonu kullanılacak.

- isim: [DB admin]
- amaç: Yapılacak projedeki veritabanından sorumlu admin olacak.
- teknolojiler: PostgreSQL kullanılacak ve detaylı, modüler bir veritabanı tasarlayacak.


## Environment
- **golang**: >= 1.23
- **database**: >= 16.0

## Önemli 1: Bu projedeki genel yapı modüler olmak zorundadır. Statik bir yapı kabul edilmez. 
## Önemli 2: Bu projede her sectionda özellikle backendde yapılan geliştirmelerde TÜRKÇE yorum satırları ile kısa kısa bilgilendirmeler yapılmadır.


## Kod yapısı

- Bu projede yazılan fonksiyonlarda camelCase kullanılmalıdır. 
- Bu projede yazılan fonksiyonların üzerilerinde alttaki fonksiyon ne yaptığına dair kısa bilgilendirme olmalıdır.
- Bu projede geliştirilirken yazılan kod parçalarında her zaman aşağıdaki gibi uygulama yapmalısın.

'''go
// CheckEquality iki tam sayıyı karşılaştırır
func CheckEquality(a int, b int) {
    if a == b {
        fmt.Printf("%d ve %d birbirine eşittir.\n", a, b)
    } else {
        fmt.Printf("%d ve %d birbirine eşit değildir.\n", a, b)
    }
}
'''


bap_ai/
├── README.md                    # 🌍 Proje Tanıtımı (Genel Bakış)
├── AGENTS.md                    # 🛠️ Geliştirme Standartları (Ajan Kılavuzu)
├── cmd/                         # Uygulamanın giriş noktası (Entry Point)
│   └── server/
│       └── main.go              # Sunucuyu başlatan ana dosya
├── backend/                     # Backend Dev & DB Admin Çalışma Alanı
│   ├── api/                     # HTTP Katmanı (Handlers & Middlewares)
│   ├── service/                 # İş Mantığı (Business Logic)
│   ├── repository/              # SQL Sorguları (Go Tarafı)
│   ├── models/                  # DB Struct Yapıları
│   └── database/                # Bağlantı Yönetimi
├── frontend/                    # Frontend Dev Çalışma Alanı
│   ├── templates/               # Dinamik HTML Şablonları
│   └── static/                  # CSS, JS ve Medya Dosyaları
├── migrations/                  # DB Admin: SQL Şema Dosyaları
├── configs/                     # Yapılandırma (.env, vb.)
├── go.mod                       # Modül Bağımlılıkları
└── go.sum