// /frontend/static/js/talep_detay_helper.js
// Türkçe Bilgilendirme: BAP resmi talep detay modallarında tip bazlı alanları okunaklı gösterir.

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

    // formatTalepDetayValue alan değerini ekranda gösterilecek metne çevirir
    function formatTalepDetayValue(key, value) {
        if (value === null || value === undefined || value === '') return '-';
        if (key === 'tutar_tl') {
            return Number(value).toLocaleString('tr-TR', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) + ' TL';
        }
        if (key === 'islem_turu') {
            return ISLEM_TURU_ETIKETLERI[String(value)] || String(value);
        }
        if (typeof value === 'object') {
            return JSON.stringify(value);
        }
        return String(value);
    }

    // buildFasilAktarimiDetayHtml fasıl aktarımına özel özet kartı üretir
    function buildFasilAktarimiDetayHtml(detay) {
        const kaynak = detay.kaynak_kalem || '-';
        const hedef = detay.hedef_kalem || '-';
        const tutar = formatTalepDetayValue('tutar_tl', detay.tutar_tl);

        return `
            <div style="margin-bottom:1rem; background:#fff; border:1px solid var(--border-color,#e2e8f0); border-radius:8px; padding:1rem;">
                <label style="font-weight:600; color:var(--primary); display:block; margin-bottom:0.75rem;">
                    <i class="fas fa-exchange-alt"></i> Fasıl Aktarım Detayları
                </label>
                <div style="display:grid; grid-template-columns:1fr auto 1fr; gap:0.75rem; align-items:center; margin-bottom:0.75rem;">
                    <div style="background:rgba(239,68,68,0.06); border:1px solid rgba(239,68,68,0.2); border-radius:8px; padding:0.75rem;">
                        <div style="font-size:0.75rem; color:var(--text-muted); text-transform:uppercase; font-weight:600; margin-bottom:0.25rem;">Kaynak Fasıl</div>
                        <div style="font-weight:700; color:#b91c1c;">${kaynak}</div>
                    </div>
                    <div style="text-align:center; color:var(--primary); font-size:1.25rem; font-weight:700;"><i class="fas fa-arrow-right"></i></div>
                    <div style="background:rgba(34,197,94,0.06); border:1px solid rgba(34,197,94,0.2); border-radius:8px; padding:0.75rem;">
                        <div style="font-size:0.75rem; color:var(--text-muted); text-transform:uppercase; font-weight:600; margin-bottom:0.25rem;">Hedef Fasıl</div>
                        <div style="font-weight:700; color:#15803d;">${hedef}</div>
                    </div>
                </div>
                <div style="background:rgba(38,74,150,0.05); border-left:4px solid var(--primary,#264a96); padding:0.75rem 1rem; border-radius:6px;">
                    <div style="font-size:0.78rem; color:var(--text-muted); text-transform:uppercase; font-weight:600;">Aktarılacak Tutar</div>
                    <div style="font-size:1.15rem; font-weight:800; color:var(--primary,#264a96); margin-top:0.15rem;">${tutar}</div>
                </div>
            </div>
        `;
    }

    // buildTalepDetaySectionHtml talep tipine göre detay bölümünü HTML olarak döner
    window.buildTalepDetaySectionHtml = function(talep) {
        if (!talep || !talep.detay || Object.keys(talep.detay).length === 0) {
            return '';
        }

        if (talep.talep_tipi === 'fasil_aktarimi') {
            return buildFasilAktarimiDetayHtml(talep.detay);
        }

        const rows = Object.entries(talep.detay).map(([key, value]) => `
            <div style="display:flex; justify-content:space-between; gap:1rem; padding:0.45rem 0; border-bottom:1px dashed var(--border-color,#e2e8f0);">
                <strong style="color:var(--text-secondary);">${ALAN_ETIKETLERI[key] || key.replace(/_/g, ' ')}:</strong>
                <span style="text-align:right;">${formatTalepDetayValue(key, value)}</span>
            </div>
        `).join('');

        return `
            <div style="margin-bottom:1rem; background:#fff; border:1px solid var(--border-color,#e2e8f0); border-radius:8px; padding:1rem;">
                <label style="font-weight:600; color:var(--primary); display:block; margin-bottom:0.6rem;">
                    <i class="fas fa-info-circle"></i> Talep Detayları
                </label>
                ${rows}
            </div>
        `;
    };
})();
