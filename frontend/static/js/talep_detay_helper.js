// /frontend/static/js/talep_detay_helper.js
// Türkçe Bilgilendirme: BAP resmi talep detay modallarında tip bazlı detay tablolarını gösterir.

(function() {
    const ALAN_ETIKETLERI = {
        kaynak_kalem: 'Kaynak Bütçe Kalemi',
        hedef_kalem: 'Hedef Bütçe Kalemi',
        tutar_tl: 'Aktarılacak Tutar',
        butce_kalemi: 'Bütçe Kalemi',
        ek_sure_ay: 'Ek Süre (Ay)',
        islem_turu: 'İşlem Türü',
        arastirmaci_adi: 'Araştırmacı Adı Soyadı',
        bursiyer_kimlik: 'Bursiyer TC Kimlik No',
        bursiyer_adi: 'Bursiyer Adı Soyadı',
        degisiklik_tanimi: 'Değişiklik Tanımı',
        dondurma_sure_ay: 'Dondurma Süresi (Ay)',
        guncelleme_tanimi: 'Güncelleme Tanımı'
    };

    const ISLEM_TURU_ETIKETLERI = {
        ekleme: 'Ekleme',
        cikarma: 'Çıkarma',
        eklenmesi: 'Eklenmesi',
        cikarilmasi: 'Çıkarılması',
        degistirilmesi: 'Değiştirilmesi'
    };

    // formatTL tutarı Türkçe para birimi cinsinden biçimlendirir (Örn: 120.000,00 TL veya 120.000 TL)
    function formatTL(val) {
        if (val === null || val === undefined || isNaN(val)) return '0,00 TL';
        const num = Number(val);
        return num.toLocaleString('tr-TR', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) + ' TL';
    }

    // buildFasilAktarimiDetayHtml: Fasıl aktarımına özel 5 sütunlu detay tablosunu üretir
    function buildFasilAktarimiDetayHtml(detay, durum) {
        const kaynak = detay.kaynak_kalem || 'Belirtilmedi';
        const hedef = detay.hedef_kalem || 'Belirtilmedi';
        const tutar = Number(detay.tutar_tl || 0);

        let kaynakOncesi = Number(detay.kaynak_oncesi_butce || 0);
        let hedefOncesi = Number(detay.hedef_oncesi_butce || 0);
        let kaynakSonrasi = 0;
        let hedefSonrasi = 0;

        if (durum === 'onaylandi') {
            // Onaylanmış ise DB bütçesi zaten güncellenmiş haldedir
            kaynakSonrasi = kaynakOncesi;
            kaynakOncesi = kaynakOncesi + tutar;

            hedefSonrasi = hedefOncesi;
            hedefOncesi = Math.max(0, hedefOncesi - tutar);
        } else {
            // Beklemede veya reddedildi durumunda aktarım öncesi verisi mevcut durumdur
            kaynakSonrasi = Math.max(0, kaynakOncesi - tutar);
            hedefSonrasi = hedefOncesi + tutar;
        }

        return `
            <div style="margin-bottom:1.25rem; background:#fff; border:1px solid var(--border-color,#e2e8f0); border-radius:8px; padding:1rem; box-shadow:0 1px 3px rgba(0,0,0,0.04);">
                <div style="font-weight:700; font-size:1.05rem; color:#1e293b; margin-bottom:0.85rem; border-bottom:1px solid #f1f5f9; padding-bottom:0.5rem;">
                    <i class="fas fa-exchange-alt" style="color:#7c3aed; margin-right:0.5rem;"></i><strong>İşlem Türü:</strong> Fasıl Aktarımı
                </div>
                <div class="table-responsive" style="overflow-x:auto;">
                    <table style="width:100%; border-collapse:collapse; font-size:0.88rem; text-align:left;">
                        <thead>
                            <tr style="background:#f8fafc; border-bottom:2px solid #e2e8f0; color:#475569;">
                                <th style="padding:0.65rem 0.75rem; font-weight:700;">İşlem</th>
                                <th style="padding:0.65rem 0.75rem; font-weight:700;">Bütçe Kalemi</th>
                                <th style="padding:0.65rem 0.75rem; font-weight:700; text-align:right;">Aktarım Öncesi Bütçe</th>
                                <th style="padding:0.65rem 0.75rem; font-weight:700; text-align:right;">Aktarım Tutarı</th>
                                <th style="padding:0.65rem 0.75rem; font-weight:700; text-align:right;">Aktarım Sonrası Bütçe</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr style="border-bottom:1px solid #f1f5f9;">
                                <td style="padding:0.65rem 0.75rem; font-weight:600; color:#b91c1c;">Aktarım Yapılan Kalem</td>
                                <td style="padding:0.65rem 0.75rem; font-weight:600; color:#334155;">${kaynak}</td>
                                <td style="padding:0.65rem 0.75rem; text-align:right; font-weight:600; color:#475569;">${formatTL(kaynakOncesi)}</td>
                                <td style="padding:0.65rem 0.75rem; text-align:right; font-weight:700; color:#ef4444;">-${formatTL(tutar)}</td>
                                <td style="padding:0.65rem 0.75rem; text-align:right; font-weight:700; color:#0f172a;">${formatTL(kaynakSonrasi)}</td>
                            </tr>
                            <tr style="border-bottom:1px solid #f1f5f9;">
                                <td style="padding:0.65rem 0.75rem; font-weight:600; color:#15803d;">Aktarım Yapılacak Kalem</td>
                                <td style="padding:0.65rem 0.75rem; font-weight:600; color:#334155;">${hedef}</td>
                                <td style="padding:0.65rem 0.75rem; text-align:right; font-weight:600; color:#475569;">${formatTL(hedefOncesi)}</td>
                                <td style="padding:0.65rem 0.75rem; text-align:right; font-weight:700; color:#22c55e;">+${formatTL(tutar)}</td>
                                <td style="padding:0.65rem 0.75rem; text-align:right; font-weight:700; color:#0f172a;">${formatTL(hedefSonrasi)}</td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>
        `;
    }

    // buildEkButceDetayHtml: Ek bütçeye özel 5 sütunlu detay tablosunu üretir
    function buildEkButceDetayHtml(detay, durum) {
        const kalem = detay.butce_kalemi || 'Belirtilmedi';
        const tutar = Number(detay.tutar_tl || 0);

        let mevcutButce = Number(detay.mevcut_butce || 0);
        let sonrakiButce = 0;

        if (durum === 'onaylandi') {
            sonrakiButce = mevcutButce;
            mevcutButce = Math.max(0, mevcutButce - tutar);
        } else {
            sonrakiButce = mevcutButce + tutar;
        }

        return `
            <div style="margin-bottom:1.25rem; background:#fff; border:1px solid var(--border-color,#e2e8f0); border-radius:8px; padding:1rem; box-shadow:0 1px 3px rgba(0,0,0,0.04);">
                <div style="font-weight:700; font-size:1.05rem; color:#1e293b; margin-bottom:0.85rem; border-bottom:1px solid #f1f5f9; padding-bottom:0.5rem;">
                    <i class="fas fa-coins" style="color:#0284c7; margin-right:0.5rem;"></i><strong>İşlem Türü:</strong> Ek Bütçe Talebi
                </div>
                <div class="table-responsive" style="overflow-x:auto;">
                    <table style="width:100%; border-collapse:collapse; font-size:0.88rem; text-align:left;">
                        <thead>
                            <tr style="background:#f8fafc; border-bottom:2px solid #e2e8f0; color:#475569;">
                                <th style="padding:0.65rem 0.75rem; font-weight:700;">İşlem</th>
                                <th style="padding:0.65rem 0.75rem; font-weight:700;">Bütçe Kalemi</th>
                                <th style="padding:0.65rem 0.75rem; font-weight:700; text-align:right;">Mevcut Bütçe</th>
                                <th style="padding:0.65rem 0.75rem; font-weight:700; text-align:right;">Ek Bütçe Tutarı</th>
                                <th style="padding:0.65rem 0.75rem; font-weight:700; text-align:right;">Talepten Sonraki Bütçe</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr style="border-bottom:1px solid #f1f5f9;">
                                <td style="padding:0.65rem 0.75rem; font-weight:600; color:#0284c7;">Ek Bütçe Talebi</td>
                                <td style="padding:0.65rem 0.75rem; font-weight:600; color:#334155;">${kalem}</td>
                                <td style="padding:0.65rem 0.75rem; text-align:right; font-weight:600; color:#475569;">${formatTL(mevcutButce)}</td>
                                <td style="padding:0.65rem 0.75rem; text-align:right; font-weight:700; color:#22c55e;">+${formatTL(tutar)}</td>
                                <td style="padding:0.65rem 0.75rem; text-align:right; font-weight:700; color:#0f172a;">${formatTL(sonrakiButce)}</td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>
        `;
    }

    // buildAvansDetayHtml: Avans talebine özel detay tablosunu üretir
    function buildAvansDetayHtml(detay, durum) {
        const kalem = detay.butce_kalemi || 'Belirtilmedi';
        const tutar = Number(detay.tutar_tl || 0);
        let mevcutButce = Number(detay.mevcut_butce || 0);
        let kalanButce = Math.max(0, mevcutButce - tutar);

        return `
            <div style="margin-bottom:1.25rem; background:#fff; border:1px solid var(--border-color,#e2e8f0); border-radius:8px; padding:1rem; box-shadow:0 1px 3px rgba(0,0,0,0.04);">
                <div style="font-weight:700; font-size:1.05rem; color:#1e293b; margin-bottom:0.85rem; border-bottom:1px solid #f1f5f9; padding-bottom:0.5rem;">
                    <i class="fas fa-hand-holding-usd" style="color:#d97706; margin-right:0.5rem;"></i><strong>İşlem Türü:</strong> Avans Talebi
                </div>
                <div class="table-responsive" style="overflow-x:auto;">
                    <table style="width:100%; border-collapse:collapse; font-size:0.88rem; text-align:left;">
                        <thead>
                            <tr style="background:#f8fafc; border-bottom:2px solid #e2e8f0; color:#475569;">
                                <th style="padding:0.65rem 0.75rem; font-weight:700;">İşlem</th>
                                <th style="padding:0.65rem 0.75rem; font-weight:700;">Bütçe Kalemi</th>
                                <th style="padding:0.65rem 0.75rem; font-weight:700; text-align:right;">Mevcut Kalem Bütçesi</th>
                                <th style="padding:0.65rem 0.75rem; font-weight:700; text-align:right;">Talep Edilen Avans</th>
                                <th style="padding:0.65rem 0.75rem; font-weight:700; text-align:right;">Kalan Bütçe</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr style="border-bottom:1px solid #f1f5f9;">
                                <td style="padding:0.65rem 0.75rem; font-weight:600; color:#d97706;">Avans Talebi</td>
                                <td style="padding:0.65rem 0.75rem; font-weight:600; color:#334155;">${kalem}</td>
                                <td style="padding:0.65rem 0.75rem; text-align:right; font-weight:600; color:#475569;">${formatTL(mevcutButce)}</td>
                                <td style="padding:0.65rem 0.75rem; text-align:right; font-weight:700; color:#d97706;">${formatTL(tutar)}</td>
                                <td style="padding:0.65rem 0.75rem; text-align:right; font-weight:700; color:#0f172a;">${formatTL(kalanButce)}</td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>
        `;
    }

    // buildEkSureDetayHtml: Ek süre talebine özel detay tablosu
    function buildEkSureDetayHtml(detay) {
        const ay = detay.ek_sure_ay || 0;
        return `
            <div style="margin-bottom:1.25rem; background:#fff; border:1px solid var(--border-color,#e2e8f0); border-radius:8px; padding:1rem; box-shadow:0 1px 3px rgba(0,0,0,0.04);">
                <div style="font-weight:700; font-size:1.05rem; color:#1e293b; margin-bottom:0.85rem; border-bottom:1px solid #f1f5f9; padding-bottom:0.5rem;">
                    <i class="fas fa-clock" style="color:#2563eb; margin-right:0.5rem;"></i><strong>İşlem Türü:</strong> Ek Süre Talebi
                </div>
                <div class="table-responsive" style="overflow-x:auto;">
                    <table style="width:100%; border-collapse:collapse; font-size:0.88rem; text-align:left;">
                        <thead>
                            <tr style="background:#f8fafc; border-bottom:2px solid #e2e8f0; color:#475569;">
                                <th style="padding:0.65rem 0.75rem; font-weight:700;">İşlem</th>
                                <th style="padding:0.65rem 0.75rem; font-weight:700; text-align:center;">Talep Edilen Ek Süre</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr style="border-bottom:1px solid #f1f5f9;">
                                <td style="padding:0.65rem 0.75rem; font-weight:600; color:#2563eb;">Proje Süre Uzatımı</td>
                                <td style="padding:0.65rem 0.75rem; text-align:center; font-weight:700; color:#0f172a;">${ay} Ay</td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>
        `;
    }

    // buildArastirmaciDetayHtml: Araştırmacı değişikliğine özel detay tablosu
    function buildArastirmaciDetayHtml(detay) {
        const islem = ISLEM_TURU_ETIKETLERI[detay.islem_turu] || detay.islem_turu || '-';
        const kisi = detay.arastirmaci_adi || '-';
        return `
            <div style="margin-bottom:1.25rem; background:#fff; border:1px solid var(--border-color,#e2e8f0); border-radius:8px; padding:1rem; box-shadow:0 1px 3px rgba(0,0,0,0.04);">
                <div style="font-weight:700; font-size:1.05rem; color:#1e293b; margin-bottom:0.85rem; border-bottom:1px solid #f1f5f9; padding-bottom:0.5rem;">
                    <i class="fas fa-user-plus" style="color:#0891b2; margin-right:0.5rem;"></i><strong>İşlem Türü:</strong> Araştırmacı Değişikliği
                </div>
                <div class="table-responsive" style="overflow-x:auto;">
                    <table style="width:100%; border-collapse:collapse; font-size:0.88rem; text-align:left;">
                        <thead>
                            <tr style="background:#f8fafc; border-bottom:2px solid #e2e8f0; color:#475569;">
                                <th style="padding:0.65rem 0.75rem; font-weight:700;">İşlem</th>
                                <th style="padding:0.65rem 0.75rem; font-weight:700;">İşlem Türü</th>
                                <th style="padding:0.65rem 0.75rem; font-weight:700;">Araştırmacı Adı Soyadı</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr style="border-bottom:1px solid #f1f5f9;">
                                <td style="padding:0.65rem 0.75rem; font-weight:600; color:#0891b2;">Ekip Güncellemesi</td>
                                <td style="padding:0.65rem 0.75rem; font-weight:600; color:#334155;">${islem}</td>
                                <td style="padding:0.65rem 0.75rem; font-weight:700; color:#0f172a;">${kisi}</td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>
        `;
    }

    // buildBursiyerDetayHtml: Bursiyer işlemine özel detay tablosu
    function buildBursiyerDetayHtml(detay) {
        const islem = ISLEM_TURU_ETIKETLERI[detay.islem_turu] || detay.islem_turu || '-';
        const kisi = detay.bursiyer_adi || '-';
        const kimlik = detay.bursiyer_kimlik || '-';
        return `
            <div style="margin-bottom:1.25rem; background:#fff; border:1px solid var(--border-color,#e2e8f0); border-radius:8px; padding:1rem; box-shadow:0 1px 3px rgba(0,0,0,0.04);">
                <div style="font-weight:700; font-size:1.05rem; color:#1e293b; margin-bottom:0.85rem; border-bottom:1px solid #f1f5f9; padding-bottom:0.5rem;">
                    <i class="fas fa-user-graduate" style="color:#4f46e5; margin-right:0.5rem;"></i><strong>İşlem Türü:</strong> Bursiyer İşlemi
                </div>
                <div class="table-responsive" style="overflow-x:auto;">
                    <table style="width:100%; border-collapse:collapse; font-size:0.88rem; text-align:left;">
                        <thead>
                            <tr style="background:#f8fafc; border-bottom:2px solid #e2e8f0; color:#475569;">
                                <th style="padding:0.65rem 0.75rem; font-weight:700;">İşlem</th>
                                <th style="padding:0.65rem 0.75rem; font-weight:700;">İşlem Türü</th>
                                <th style="padding:0.65rem 0.75rem; font-weight:700;">Bursiyer Adı Soyadı</th>
                                <th style="padding:0.65rem 0.75rem; font-weight:700;">TC Kimlik No</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr style="border-bottom:1px solid #f1f5f9;">
                                <td style="padding:0.65rem 0.75rem; font-weight:600; color:#4f46e5;">Bursiyer İşlemi</td>
                                <td style="padding:0.65rem 0.75rem; font-weight:600; color:#334155;">${islem}</td>
                                <td style="padding:0.65rem 0.75rem; font-weight:700; color:#0f172a;">${kisi}</td>
                                <td style="padding:0.65rem 0.75rem; font-weight:600; color:#64748b;">${kimlik}</td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>
        `;
    }

    // buildJenerikDetayHtml: Diğer talep türleri için genel detay tablosu
    function buildJenerikDetayHtml(talep) {
        const detay = talep.detay || {};
        const tipEtiket = talep.talep_tipi_etiketi || talep.talep_tipi || 'Talep';
        const rows = Object.entries(detay).map(([key, value]) => `
            <tr style="border-bottom:1px solid #f1f5f9;">
                <td style="padding:0.65rem 0.75rem; font-weight:600; color:#475569;">${ALAN_ETIKETLERI[key] || key.replace(/_/g, ' ')}</td>
                <td style="padding:0.65rem 0.75rem; font-weight:700; color:#0f172a;">${typeof value === 'object' ? JSON.stringify(value) : value}</td>
            </tr>
        `).join('');

        return `
            <div style="margin-bottom:1.25rem; background:#fff; border:1px solid var(--border-color,#e2e8f0); border-radius:8px; padding:1rem; box-shadow:0 1px 3px rgba(0,0,0,0.04);">
                <div style="font-weight:700; font-size:1.05rem; color:#1e293b; margin-bottom:0.85rem; border-bottom:1px solid #f1f5f9; padding-bottom:0.5rem;">
                    <i class="fas fa-info-circle" style="color:#0284c7; margin-right:0.5rem;"></i><strong>İşlem Türü:</strong> ${tipEtiket}
                </div>
                <div class="table-responsive" style="overflow-x:auto;">
                    <table style="width:100%; border-collapse:collapse; font-size:0.88rem; text-align:left;">
                        <thead>
                            <tr style="background:#f8fafc; border-bottom:2px solid #e2e8f0; color:#475569;">
                                <th style="padding:0.65rem 0.75rem; font-weight:700;">Detay Bilgisi</th>
                                <th style="padding:0.65rem 0.75rem; font-weight:700;">Açıklama / Değer</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${rows || '<tr><td colspan="2" style="padding:0.65rem 0.75rem; color:#94a3b8;">Ek detay bulunmamaktadır.</td></tr>'}
                        </tbody>
                    </table>
                </div>
            </div>
        `;
    }

    // buildTalepDetaySectionHtml: Talep tipine göre detay tablosunu döner
    window.buildTalepDetaySectionHtml = function(talep) {
        if (!talep) return '';

        const tip = talep.talep_tipi;
        const detay = talep.detay || {};
        const durum = talep.durum || 'beklemede';

        if (tip === 'fasil_aktarimi') {
            return buildFasilAktarimiDetayHtml(detay, durum);
        }
        if (tip === 'ek_butce') {
            return buildEkButceDetayHtml(detay, durum);
        }
        if (tip === 'avans') {
            return buildAvansDetayHtml(detay, durum);
        }
        if (tip === 'ek_sure') {
            return buildEkSureDetayHtml(detay);
        }
        if (tip === 'arastirmaci') {
            return buildArastirmaciDetayHtml(detay);
        }
        if (tip === 'bursiyer') {
            return buildBursiyerDetayHtml(detay);
        }

        return buildJenerikDetayHtml(talep);
    };
})();
