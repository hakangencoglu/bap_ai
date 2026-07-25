# BAP AI - İş Akış Şeması (Workflow Schema)

Bu doküman, **BAP AI (Bilimsel Araştırma Projeleri Yönetim Sistemi)** kapsamındaki tüm kullanıcı rollerini, proje durum geçişlerini, onay mekanizmalarını ve onay sonrası süreçleri (Satın Alma, Değişiklik Talepleri) detaylı bir şekilde açıklar.

---

## 👥 Sistem Rolleri ve Sorumlulukları

Projedeki temel aktörler ve sistem üzerindeki yetki sınırları şu şekildedir:

1. **Akademisyen (Yürütücü / Araştırmacı):**
   - Yeni BAP projesi başvurusu oluşturur.
   - Proje ekibini (Araştırmacı, Danışman, Bursiyer) davet eder ve yönetir.
   - Taslak projeyi günceller, otomatik kaydetme ile veri kaybını önler ve başvuru yapar.
   - Onaylanmış projeler için **Satın Alma Talebi** ve **Süreç Değişiklik Talepleri** (ek süre, ek bütçe, dondurma vb.) iletir.
   - Proje sözleşmesini **E-İmza** ile imzalar.

2. **TTO Temsilcisi (Teknoloji Transfer Ofisi):**
   - Projelerin ön incelemesini yapar (`incelemede` aşaması).
   - Dekan onayından geçen projeleri Komisyon onayına sevk eder.
   - Komisyon onayından sonra hakem gerektiren projeler için **Hakem Ataması** yapar veya hakemsiz projeleri doğrudan sözleşme aşamasına taşır.
   - Hakem değerlendirmeleri bittiğinde süreci yönetir ve nihai sözleşme aşamasına sevk eder.
   - Satın alma ve süreç değişiklik taleplerini onaylar/reddeder.

3. **Fakülte Dekanı:**
   - Kendi fakültesindeki araştırmacıların başvurduğu projeleri inceler, onaylar, reddeder veya revizyon talep eder (`dekan_onayi_bekliyor`).

4. **BAP Komisyon Üyesi ve Komisyon Başkanı:**
   - Komisyona sevk edilen projeleri oylar (`komisyon_bekliyor`).
   - Komisyon Başkanı toplantılar düzenleyebilir ve nihai komisyon kararını doğrudan onaylama yetkisine sahiptir.
   - Komisyon üyelerinin tamamı onayladığında proje otomatik olarak bir sonraki aşamaya geçer. Herhangi bir üye red veya revizyon verirse süreç buna göre şekillenir.

5. **Hakem:**
   - Atandığı projeler için öncelikle **Gizlilik Taahhütnamesini** onaylayarak görevi kabul/red eder.
   - Kabul ettiği projeleri detaylı sorular eşliğinde puanlar ve görüş (Onay/Red/Revizyon) bildirir (`hakem_bekliyor`).

6. **Sistem Yöneticisi (Admin):**
   - Sistemdeki tüm roller üzerinde tam yetkiye sahiptir.
   - Hakem atayabilir, durumları manuel güncelleyebilir, sayfa/rol yetki matrisini yönetir ve BAP türlerini tanımlar.

---

## 📊 1. BAP Başvuru ve Onay Ana İş Akışı

Aşağıdaki şema, bir BAP projesinin araştırmacı tarafından taslak olarak oluşturulmasından yürürlüğe girmesine (ve tamamlanmasına) kadar olan ana iş akışını göstermektedir.

```mermaid
flowchart TD
    %% Stil Tanımları
    classDef start_end fill:#eceff1,stroke:#37474f,stroke-width:2px;
    classDef akademisyen fill:#e3f2fd,stroke:#1565c0,stroke-width:2px;
    classDef tto fill:#fff3e0,stroke:#ef6c00,stroke-width:2px;
    classDef dekan fill:#f3e5f5,stroke:#6a1b9a,stroke-width:2px;
    classDef komisyon fill:#e8f5e9,stroke:#2e7d32,stroke-width:2px;
    classDef hakem fill:#efebe9,stroke:#4e342e,stroke-width:2px;
    classDef durum fill:#f5f5f5,stroke:#9e9e9e,stroke-width:1px,stroke-dasharray: 5 5;

    %% Süreç Düğümleri
    Start([Başlangıç]) :::start_end
    
    subgraph A_Basvuru ["1. Başvuru & Taslak Aşaması (Araştırmacı)"]
        Create[Proje Başvurusu Oluştur] :::akademisyen
        DraftStatus{Durum: 'taslak'} :::durum
        StepForm[5 Adımlı Multi-Step Form Doldurma] :::akademisyen
        Autosave[Arka Planda Otomatik Kaydetme] :::akademisyen
        AddTeam[Ekip Arkadaşı Ekle & Davet Gönder] :::akademisyen
        PDFPreview[PDF Önizleme Motoru] :::akademisyen
        Submit[Projeyi Gönder / Kilitle] :::akademisyen
    end

    subgraph B_TTO_On["2. TTO Ön İnceleme"]
        TTO_On_Status{Durum: 'incelemede'} :::durum
        TTO_Review[TTO Temsilcisi İnceleme] :::tto
        TTO_Decision{TTO Kararı?} :::tto
    end

    subgraph C_Dekan["3. Dekan Onay Süreci"]
        Dekan_Status{Durum: 'dekan_onayi_bekliyor'} :::durum
        Dekan_Review[Fakülte Dekanı İnceleme] :::dekan
        Dekan_Decision{Dekan Kararı?} :::dekan
        Dekan_Approved_Status{Durum: 'dekan_onayladi'} :::durum
        TTO_Sevk_Komisyon[TTO Komisyona Sevk Eder] :::tto
    end

    subgraph D_Komisyon["4. BAP Komisyon Onay Süreci"]
        Komisyon_Status{Durum: 'komisyon_bekliyor'} :::durum
        Komisyon_Vote[Komisyon Üyeleri Oylama] :::komisyon
        Komisyon_Decision{Komisyon Kararı?} :::komisyon
        Komisyon_Approved_Status{Durum: 'komisyon_onayladi'} :::durum
        TTO_Sevk_Hakem[TTO Kararı / BAP Türü Kontrolü] :::tto
    end

    subgraph E_Hakem["5. Hakem Değerlendirme Aşaması"]
        Hakem_Atama_Status{Durum: 'hakem_atama_bekliyor'} :::durum
        TTO_Assign_Hakem[TTO/Admin Hakem Atar] :::tto
        Hakem_Sub[Hakem Alt Süreci - Detaylı Değerlendirme] :::hakem
        Hakem_Approved_Status{Durum: 'hakem_onayladi'} :::durum
        TTO_Final_Approve[TTO Sözleşmeye Sevk Eder] :::tto
    end

    subgraph F_Sozlesme["6. Sözleşme & Yürürlük"]
        Sozlesme_Status{Durum: 'sozlesme_imza'} :::durum
        Gen_Contract[Sözleşme PDF'i Oluşturulur] :::tto
        ESignature[E-İmza Süreci - Tüm Takım & TTO] :::akademisyen
        Active_Status{Durum: 'yururlukte'} :::durum
        Complete_Status{Durum: 'tamamlandi'} :::durum
    end

    %% Revizyon ve Red Durumları
    RevisionStatus{Durum: 'revizyon'} :::durum
    RejectedStatus{Durum: 'reddedildi'} :::durum
    End([Son]) :::start_end

    %% Bağlantılar ve Akış Yönleri
    Start --> Create
    Create --> DraftStatus
    DraftStatus --> StepForm
    StepForm --> Autosave
    StepForm --> AddTeam
    StepForm --> PDFPreview
    StepForm --> Submit
    Submit --> TTO_On_Status
    
    %% TTO Karar Akışı
    TTO_On_Status --> TTO_Review
    TTO_Review --> TTO_Decision
    TTO_Decision -- Reddet --> RejectedStatus
    TTO_Decision -- Revizyon --> RevisionStatus
    TTO_Decision -- Onayla --> Dekan_Status
    
    %% Dekan Karar Akışı
    Dekan_Status --> Dekan_Review
    Dekan_Review --> Dekan_Decision
    Dekan_Decision -- Reddet --> RejectedStatus
    Dekan_Decision -- Revizyon --> RevisionStatus
    Dekan_Decision -- Onayla --> Dekan_Approved_Status
    
    Dekan_Approved_Status --> TTO_Sevk_Komisyon
    TTO_Sevk_Komisyon --> Komisyon_Status
    
    %% Komisyon Karar Akışı
    Komisyon_Status --> Komisyon_Vote
    Komisyon_Vote --> Komisyon_Decision
    Komisyon_Decision -- Reddet --> RejectedStatus
    Komisyon_Decision -- Revizyon --> RevisionStatus
    Komisyon_Decision -- Onayla --> Komisyon_Approved_Status
    
    %% Komisyondan Sonraki Aşama
    Komisyon_Approved_Status --> TTO_Sevk_Hakem
    TTO_Sevk_Hakem -- "Hakem Gerekli Değil" --> Sozlesme_Status
    TTO_Sevk_Hakem -- "Hakem Gerekli" --> Hakem_Atama_Status
    
    %% Hakem Aşaması Akışı
    Hakem_Atama_Status --> TTO_Assign_Hakem
    TTO_Assign_Hakem --> Hakem_Sub
    
    %% Hakem Sonuç Değerlendirmesi
    Hakem_Sub -- "Düşük Puan / Red / Uyuşmazlık" --> Hakem_Atama_Status
    Hakem_Sub -- "Revizyon Talebi" --> RevisionStatus
    Hakem_Sub -- "Tüm Hakemler Onayladı" --> Hakem_Approved_Status
    
    Hakem_Approved_Status --> TTO_Final_Approve
    TTO_Final_Approve --> Sozlesme_Status
    
    %% Sözleşme ve Yürürlüğe Giriş
    Sozlesme_Status --> Gen_Contract
    Gen_Contract --> ESignature
    ESignature --> Active_Status
    Active_Status --> Complete_Status
    Complete_Status --> End
    RejectedStatus --> End
    
    %% Revizyondan Taslağa Dönüş Akışı
    RevisionStatus --> |"Araştırmacı Projeyi Düzenler (PUT)"| DraftStatus
```

---

## ⚖️ 2. Hakem Değerlendirme Detay Akışı (Alt Süreç)

BAP projelerinin bilimsel kalitesini ölçmek amacıyla işletilen hakem atama, kabul/red ve detaylı puanlama süreci aşağıdaki şemada detaylandırılmıştır:

```mermaid
flowchart TD
    classDef tto fill:#fff3e0,stroke:#ef6c00,stroke-width:2px;
    classDef hakem fill:#efebe9,stroke:#4e342e,stroke-width:2px;
    classDef durum fill:#f5f5f5,stroke:#9e9e9e,stroke-width:1px,stroke-dasharray: 5 5;

    Start_Hakem([Hakem Süreci Başlangıcı])
    Assign[TTO: Hakem Listesinden Seçim ve Davet] :::tto
    InviteSent{Hakem Daveti Gönderildi} :::durum
    
    InviteDecision{Hakem Kararı?} :::hakem
    
    %% Red Durumu
    RejectReason[Hakem Red Nedeni Girer] :::hakem
    RevertTTO[Atama Durumu: 'Reddedildi'\nProje TTO'ya İade Edilir] :::tto
    ReAssign[TTO: Yeni Hakem Ataması Yapar] :::tto
    
    %% Kabul Durumu
    AcceptTaahhut[Gizlilik Taahhütnamesi Onayı] :::hakem
    AwaitingReview{Durum: 'hakem_bekliyor'} :::durum
    
    %% Değerlendirme Formu
    FillForm[Dinamik Soru & Puanlama Formu doldurulur] :::hakem
    SubmitEvaluation[Değerlendirme Gönderilir] :::hakem
    
    EvalDecision{Hakem Nihai Kararı?} :::hakem
    
    %% Değerlendirme Sonuçları
    EvalRevizyon[Hakem Revizyon Talebi] :::hakem
    ReqRev[Proje: 'revizyon' durumuna geçer\nRevizyon Açıklaması Eklenir] :::durum
    
    EvalRed[Hakem Projeyi Reddetti] :::hakem
    ReqRed[Proje TTO Kararına İade Edilir\nDurum: 'hakem_atama_bekliyor'] :::durum
    
    EvalOnay[Hakem Projeyi Onayladı] :::hakem
    CheckAllFinished{Tüm Hakemlerin\nİşlemleri Bitti mi?} :::durum
    
    %% Nihai Karar Kontrolü
    ScoreCheck{Tüm Hakemlerin Puanı\n>= 70 mi?} :::durum
    Passed[Proje Durumu: 'hakem_onayladi'] :::durum
    Failed[Düşük Puan / Uyuşmazlık:\nDurum: 'hakem_atama_bekliyor'] :::durum

    %% Akış Bağlantıları
    Start_Hakem --> Assign
    Assign --> InviteSent
    InviteSent --> InviteDecision
    
    InviteDecision -- Red --> RejectReason
    RejectReason --> RevertTTO
    RevertTTO --> ReAssign
    ReAssign --> InviteSent
    
    InviteDecision -- Kabul --> AcceptTaahhut
    AcceptTaahhut --> AwaitingReview
    AwaitingReview --> FillForm
    FillForm --> SubmitEvaluation
    SubmitEvaluation --> EvalDecision
    
    EvalDecision -- "Revizyon" --> EvalRevizyon
    EvalRevizyon --> ReqRev
    
    EvalDecision -- "Reddedildi" --> EvalRed
    EvalRed --> ReqRed
    
    EvalDecision -- "Onaylandı" --> EvalOnay
    EvalOnay --> CheckAllFinished
    
    CheckAllFinished -- Hayır --> AwaitingReview
    CheckAllFinished -- Evet --> ScoreCheck
    
    ScoreCheck -- Evet --> Passed
    ScoreCheck -- "Hayır (En az biri < 70)" --> Failed
```

---

## 🛠️ 3. Yürürlükteki Proje Yönetimi (Onay Sonrası Alt Süreçler)

Proje onaylandıktan ve durumu `yururlukte` (Aktif) olduktan sonra, araştırmacı (proje yürütücüsü) proje süresince satın alma yapabilir veya idari talepler iletebilir.

### A. Satın Alma Talep İş Akışı
Proje bütçe kalemlerine (cihaz, sarf malzeme, hizmet vb.) uygun olarak yapılan satın alma süreçlerini tanımlar.

```mermaid
flowchart LR
    classDef ak fill:#e3f2fd,stroke:#1565c0,stroke-width:2px;
    classDef tto fill:#fff3e0,stroke:#ef6c00,stroke-width:2px;
    classDef state fill:#f5f5f5,stroke:#9e9e9e,stroke-dasharray:3 3;

    A[Yürütücü: Satın Alma Talebi Oluşturur] :::ak --> B{Bütçe Yeterli mi?} :::state
    B -- Hayır --> C[HATA: Yetersiz Bütçe] :::state
    B -- Evet --> D[TTO: Satın Alma İnceleme] :::tto
    D --> E{TTO Kararı?} :::tto
    E -- Reddet --> F[Talep Reddedildi\nBütçe Serbest Bırakılır] :::state
    E -- Onayla --> G[Talep Onaylandı\nSatın Alma İşlemi Başlatılır] :::state
```

### B. Değişiklik Talepleri İş Akışı
Projenin idari ve teknik koşullarında (ek süre, ek bütçe, ekip değişikliği vb.) meydana gelen revizyon taleplerini yönetir.

```mermaid
flowchart TD
    classDef ak fill:#e3f2fd,stroke:#1565c0,stroke-width:2px;
    classDef tto fill:#fff3e0,stroke:#ef6c00,stroke-width:2px;
    classDef state fill:#f5f5f5,stroke:#9e9e9e,stroke-dasharray:3 3;

    A[Yürütücü: Talep Gönderir\n/api/talep/:tip] :::ak --> B{Talep Tipi Nedir?} :::state
    
    B --> T1[Ek Süre\ntalep_ek_sure]
    B --> T2[Ek Bütçe\ntalep_ek_butce]
    B --> T3[Fasıl Aktarımı\ntalep_fasil_aktarimi]
    B --> T4[Ekip Değişikliği\ntalep_arastirmaci/bursiyer]
    B --> T5[İptal / Dondurma\ntalep_proje_iptali/dondurma]
    
    T1 & T2 & T3 & T4 & T5 --> C[TTO / Admin İncelemesi] :::tto
    C --> D{Onay Durumu?} :::tto
    
    D -- Reddet --> E[Talep Reddedildi\n(Gerekçe Bildirilir)] :::state
    D -- Onayla --> F[Proje Verileri Güncellenir\n(Süre, bütçe veya ekip güncellenir)] :::state
```

---

## 🔄 4. Proje Durum Değişiklikleri ve Açıklamaları

Veritabanında kayıtlı `proje_durum` tablosu değerleri ve bu durumların anlamları aşağıdaki gibidir:

| Durum Adı (`durum_adi`) | Ekran Karşılığı (`durum_etiketi`) | Açıklama / Süreçteki Yeri |
| :--- | :--- | :--- |
| `taslak` | Taslak | Akademisyen projeyi yazma aşamasındadır. Henüz onay sürecine gönderilmemiştir. |
| `incelemede` | İncelemede | Proje TTO Ön İnceleme sırasındadır. |
| `dekan_onayi_bekliyor` | Dekan Onayı Bekliyor | TTO ön incelemeyi onaylamış, proje Fakülte Dekanının onayını beklemektedir. |
| `dekan_onayladi` | Dekan Onayladı | Dekan projeyi onaylamış, proje komisyona sevk edilmek üzere TTO sırasına gelmiştir. |
| `komisyon_bekliyor` | Komisyon Onayı Bekliyor | Proje BAP Komisyon Üyelerinin oylamasına sunulmuştur. |
| `komisyon_onayladi` | Komisyon Onayladı | Komisyon oylaması olumlu tamamlanmıştır. TTO hakem gereksinimini denetler. |
| `hakem_atama_bekliyor` | Hakem Atama Bekleniyor | Projeye hakem ataması yapılması gerekmektedir (TTO/Admin tarafından). |
| `hakem_bekliyor` | Hakem İncelemesinde | Hakemler daveti kabul etmiş, değerlendirme raporunu doldurmaktadır. |
| `hakem_onayladi` | Hakem Onayladı | Tüm hakem değerlendirmeleri başarıyla (>=70 puan ve Onay kararıyla) tamamlanmıştır. |
| `sozlesme_imza` | Sözleşme / İmza Aşaması | Projenin onay aşaması bitmiş, sözleşme oluşturulmuş ve ıslak/e-imza sırasındadır. |
| `yururlukte` | Yürürlükte (Aktif) | Sözleşmeler imzalanmış, proje aktif olarak yürütülmektedir. Bütçe harcamalarına açıktır. |
| `tamamlandi` | Onaylandı (Tamamlandı) | Proje başarıyla nihayete ermiş ve kapatılmıştır. |
| `reddedildi` | Reddedildi | TTO, Dekan, Komisyon veya Sözleşme adımlarının herhangi birinde proje reddedilmiştir. |
| `revizyon` | Revizyon Gerekli | Onay mercilerinden veya hakemlerden biri düzeltme istemiştir. Proje taslak haline geri döner. |
| `tto_aktif` | TTO Onayı Bekliyor | Eski akışlardan kalan, TTO'nun projeyi aktifleştirmesini bekleyen ara aşama. |
