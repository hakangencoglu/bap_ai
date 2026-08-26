// Satın alma talebi detay modalı — TTO ve araştırmacı ekranlarında ortak kullanılır.
(function() {
    'use strict';

    function escapeHtml(str) {
        if (!str) return '';
        return String(str)
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#039;');
    }

    function formatMoney(val) {
        const n = Number(val);
        if (isNaN(n)) return '₺0,00';
        return '₺' + n.toLocaleString('tr-TR', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
    }

    function getAuthToken() {
        return localStorage.getItem('jwt_token') || '';
    }

    // findButceHavuzu bütçe raporunda kalemi veya kategori havuzunu bulur
    function findButceHavuzu(kalemler, item) {
        if (!Array.isArray(kalemler) || !item) return null;
        let h = kalemler.find(k => Number(k.kalem_id) === Number(item.kalem_id));
        if (h) return h;
        if (item.butce_kategori_adi) {
            h = kalemler.find(k => k.kategori_adi === item.butce_kategori_adi);
        }
        return h || null;
    }

    function findButceHavuzuByKalemId(kalemler, kalemId, kategoriAdi) {
        if (!Array.isArray(kalemler) || !kalemId) return null;
        let h = kalemler.find(k => Number(k.kalem_id) === Number(kalemId));
        if (h) return h;
        if (kategoriAdi) {
            h = kalemler.find(k => k.kategori_adi === kategoriAdi);
        }
        return h || null;
    }

    // computeBudgetImpact mutabakat öncesi/sonrası kalan bütçeyi hesaplar
    function computeBudgetImpact(odeme, origHavuz, farkHavuz) {
        const taahhut = Number(odeme.taahhut_tutari) || 0;
        const fiili = Number(odeme.fiili_tutar) || 0;
        const fark = fiili - taahhut;
        const origKalemId = Number(odeme.kalem_id);
        const farkKalemId = odeme.fark_kalem_id ? Number(odeme.fark_kalem_id) : origKalemId;
        const sameKalem = farkKalemId === origKalemId;

        const afterOrig = origHavuz != null ? Number(origHavuz.kullanilabilir) : null;
        const afterFark = farkHavuz != null ? Number(farkHavuz.kullanilabilir) : afterOrig;

        let beforeOrig = null;
        let beforeFark = null;

        if (Math.abs(fark) < 0.01) {
            beforeOrig = afterOrig;
            beforeFark = afterFark;
        } else if (sameKalem) {
            beforeOrig = afterOrig != null ? afterOrig - fark : null;
            beforeFark = beforeOrig;
        } else if (fark > 0) {
            beforeOrig = afterOrig != null ? afterOrig + taahhut : null;
            beforeFark = afterFark != null ? afterFark + fark : null;
        } else {
            const transfer = taahhut - fiili;
            beforeOrig = afterOrig != null ? afterOrig + taahhut : null;
            beforeFark = afterFark != null ? afterFark - transfer : null;
        }

        return { taahhut, fiili, fark, sameKalem, farkKalemId, beforeOrig, afterOrig, beforeFark, afterFark };
    }

    function farkYonuLabel(yon, fark) {
        if (yon === 'fazla' || fark > 0.009) return 'Fazla';
        if (yon === 'eksik' || fark < -0.009) return 'Eksik';
        return 'Eşit';
    }

    function renderBudgetRow(label, before, after) {
        if (before == null && after == null) return '';
        const degisim = (after != null && before != null) ? after - before : 0;
        let degisimHtml = '—';
        if (Math.abs(degisim) > 0.009) {
            const cls = degisim > 0 ? 'color:#15803d;' : 'color:#b91c1c;';
            const sign = degisim > 0 ? '+' : '';
            degisimHtml = `<span style="${cls} font-weight:700;">${sign}${formatMoney(degisim)}</span>`;
        } else {
            degisimHtml = '<span style="color:var(--text-muted);">Değişim yok</span>';
        }
        return `
            <tr style="border-bottom:1px solid var(--border-color,#e2e8f0);">
                <td style="padding:0.55rem 0.75rem; font-weight:600;">${escapeHtml(label)}</td>
                <td style="padding:0.55rem 0.75rem; text-align:right;">${before != null ? formatMoney(before) : '—'}</td>
                <td style="padding:0.55rem 0.75rem; text-align:right; font-weight:700;">${after != null ? formatMoney(after) : '—'}</td>
                <td style="padding:0.55rem 0.75rem; text-align:right;">${degisimHtml}</td>
            </tr>
        `;
    }

    // renderSatinalmaDetayHtml talep grubu detay HTML'ini üretir
    function renderSatinalmaDetayHtml(talepler, odemeler, butceKalemler, talepNo) {
        if (!talepler || talepler.length === 0) {
            return '<div style="text-align:center;padding:2rem;color:var(--error);">Talep detayları bulunamadı.</div>';
        }

        const t0 = talepler[0];
        const durumMap = {
            'Kapatildi': { label: 'Kapatıldı', cls: 'background:#eff6ff;color:#1d4ed8;border:1px solid #bfdbfe;' },
            'IptalEdildi': { label: 'İptal Edildi', cls: 'background:#f3f4f6;color:#4b5563;border:1px solid #e5e7eb;' },
            'Onaylandı': { label: 'Onaylandı', cls: 'background:#f0fdf4;color:#15803d;border:1px solid #dcfce7;' },
            'Beklemede': { label: 'Beklemede', cls: 'background:#fff7ed;color:#c2410c;border:1px solid #ffedd5;' },
            'Reddedildi': { label: 'Reddedildi', cls: 'background:#fef2f2;color:#b91c1c;border:1px solid #fee2e2;' }
        };
        const durumInfo = durumMap[t0.durum] || { label: t0.durum, cls: 'background:#f3f4f6;color:#374151;' };
        const olusturmaTarihi = t0.olusturma_tarihi ? new Date(t0.olusturma_tarihi).toLocaleString('tr-TR') : '—';

        const itemRows = talepler.map((item, idx) => {
            const taahhut = Number(item.revize_toplam_fiyat != null ? item.revize_toplam_fiyat : item.toplam_fiyat) || 0;
            const odeme = odemeler.find(o => Number(o.talep_id) === Number(item.talep_id));
            const fiili = odeme ? Number(odeme.fiili_tutar) : null;
            const fark = fiili != null ? fiili - taahhut : null;
            const farkLabel = fark != null ? farkYonuLabel(odeme ? odeme.fark_yonu : '', fark) : '—';
            let farkStyle = '';
            if (fark != null && fark > 0.009) farkStyle = 'color:#c2410c;font-weight:700;';
            else if (fark != null && fark < -0.009) farkStyle = 'color:#15803d;font-weight:700;';

            const revizeBirim = item.revize_birim_fiyat != null
                ? formatMoney(item.revize_birim_fiyat)
                : '—';

            return `
                <tr style="border-bottom:1px solid #e2e8f0;">
                    <td style="padding:0.6rem 0.75rem; font-weight:600;">${idx + 1}. ${escapeHtml(item.malzeme_adi || '—')}</td>
                    <td style="padding:0.6rem 0.75rem; text-align:center;">${item.miktar || 1}</td>
                    <td style="padding:0.6rem 0.75rem; text-align:right;">${formatMoney(item.birim_fiyat)}${item.revize_birim_fiyat != null ? `<br><small style="color:#2563eb;">Revize: ${revizeBirim}</small>` : ''}</td>
                    <td style="padding:0.6rem 0.75rem; text-align:right; font-weight:700;">${formatMoney(taahhut)}</td>
                    <td style="padding:0.6rem 0.75rem; text-align:right; font-weight:700;">${fiili != null ? formatMoney(fiili) : '<span style="color:var(--text-muted);">Mutabakat bekliyor</span>'}</td>
                    <td style="padding:0.6rem 0.75rem; text-align:right; ${farkStyle}">${fark != null ? formatMoney(fark) : '—'}</td>
                    <td style="padding:0.6rem 0.75rem; text-align:center;">${farkLabel}</td>
                    <td style="padding:0.6rem 0.75rem;">${escapeHtml(item.butce_kategori_adi || item.kalem_adi || '—')}</td>
                </tr>
            `;
        }).join('');

        const budgetSections = [];
        talepler.forEach(item => {
            const odeme = odemeler.find(o => Number(o.talep_id) === Number(item.talep_id));
            if (!odeme) return;

            const origHavuz = findButceHavuzu(butceKalemler, item);
            const farkKalemId = odeme.fark_kalem_id || odeme.kalem_id;
            const farkHavuz = findButceHavuzuByKalemId(
                butceKalemler,
                farkKalemId,
                odeme.fark_kalem_kategori_adi || odeme.butce_kategori_adi
            );
            const impact = computeBudgetImpact(odeme, origHavuz, farkHavuz);

            let rows = renderBudgetRow(
                `Satın alma kalemi (${escapeHtml(item.butce_kategori_adi || odeme.butce_kategori_adi || 'Kalem')})`,
                impact.beforeOrig,
                impact.afterOrig
            );

            if (!impact.sameKalem && Math.abs(impact.fark) > 0.009) {
                const farkLabel = odeme.fark_kalem_kategori_adi || 'Fark kalemi';
                rows += renderBudgetRow(
                    `Fark kalemi (${escapeHtml(farkLabel)})`,
                    impact.beforeFark,
                    impact.afterFark
                );
            }

            if (rows) {
                budgetSections.push(`
                    <div style="margin-bottom:1rem;">
                        <div style="font-size:0.85rem; font-weight:700; color:var(--primary); margin-bottom:0.35rem;">
                            <i class="fas fa-box"></i> ${escapeHtml(item.malzeme_adi || 'Kalem')} — Bütçe etkisi
                        </div>
                        <div style="overflow-x:auto;">
                            <table style="width:100%; border-collapse:collapse; font-size:0.85rem;">
                                <thead>
                                    <tr style="background:var(--bg-surface-light,#f1f5f9);">
                                        <th style="padding:0.45rem 0.75rem; text-align:left;">Bütçe kalemi</th>
                                        <th style="padding:0.45rem 0.75rem; text-align:right;">Mutabakat öncesi kalan</th>
                                        <th style="padding:0.45rem 0.75rem; text-align:right;">Mutabakat sonrası kalan</th>
                                        <th style="padding:0.45rem 0.75rem; text-align:right;">Değişim</th>
                                    </tr>
                                </thead>
                                <tbody>${rows}</tbody>
                            </table>
                        </div>
                        ${odeme.fatura_no ? `<div style="font-size:0.8rem; color:var(--text-muted); margin-top:0.35rem;">Fatura: <strong>${escapeHtml(odeme.fatura_no)}</strong>${odeme.fatura_tarihi ? ' — ' + new Date(odeme.fatura_tarihi).toLocaleDateString('tr-TR') : ''}</div>` : ''}
                        ${odeme.tto_gerekce ? `<div style="font-size:0.8rem; color:var(--text-secondary); margin-top:0.25rem;">Mutabakat gerekçesi: ${escapeHtml(odeme.tto_gerekce)}</div>` : ''}
                    </div>
                `);
            }
        });

        const gerekcelerStr = talepler.map(t => t.gerekce).filter(Boolean).join(' | ') || 'Gerekçe belirtilmemiş.';
        const redNedenleri = talepler.map(t => t.red_nedeni).filter(Boolean).join(' | ');
        const revizyonGerekceleri = talepler.map(t => t.revizyon_gerekcesi).filter(Boolean).join(' | ');

        const toplamTaahhut = talepler.reduce((s, i) => s + Number(i.revize_toplam_fiyat != null ? i.revize_toplam_fiyat : i.toplam_fiyat), 0);
        const talepOdemeler = odemeler.filter(o => talepler.some(t => Number(t.talep_id) === Number(o.talep_id)));
        const toplamFiili = talepOdemeler.reduce((s, o) => s + Number(o.fiili_tutar || 0), 0);

        return `
            <div style="background:var(--bg-surface-light,#f8fafc); padding:1rem; border-radius:8px; border:1px solid var(--border-color,#e2e8f0); margin-bottom:1.25rem;">
                <div style="display:flex; justify-content:space-between; align-items:center; flex-wrap:wrap; gap:0.5rem; margin-bottom:0.75rem;">
                    <div style="display:flex; align-items:center; gap:0.6rem;">
                        <span style="font-size:1.1rem; font-weight:700; color:var(--primary,#264a96);"><code>${escapeHtml(talepNo)}</code></span>
                        <span style="font-size:0.85rem; font-weight:600; padding:0.35rem 0.75rem; border-radius:20px; ${durumInfo.cls}">${durumInfo.label}</span>
                    </div>
                    <div style="font-size:0.85rem; color:var(--text-muted);"><i class="fas fa-calendar-alt"></i> ${olusturmaTarihi}</div>
                </div>
                <div style="display:grid; grid-template-columns:repeat(auto-fit,minmax(180px,1fr)); gap:0.75rem; font-size:0.9rem;">
                    <div><strong>Proje:</strong> <code>${escapeHtml(t0.proje_kodu || '—')}</code></div>
                    <div><strong>Talep eden:</strong> ${escapeHtml(t0.uye_ad_soyad || '—')}</div>
                    <div><strong>Bütçe kalemi:</strong> ${escapeHtml(t0.butce_kategori_adi || '—')}</div>
                </div>
                ${t0.proje_baslik ? `<div style="margin-top:0.5rem;font-size:0.88rem;color:var(--text-secondary);"><strong>Proje başlığı:</strong> ${escapeHtml(t0.proje_baslik)}</div>` : ''}
            </div>

            <div style="display:grid; grid-template-columns:repeat(auto-fit,minmax(150px,1fr)); gap:0.75rem; margin-bottom:1.25rem;">
                <div style="background:rgba(38,74,150,0.04); border-left:4px solid var(--primary); padding:0.75rem 1rem; border-radius:6px;">
                    <div style="font-size:0.75rem; color:var(--text-muted); text-transform:uppercase; font-weight:600;">Toplam taahhüt</div>
                    <div style="font-size:1.15rem; font-weight:800; color:var(--primary); margin-top:0.2rem;">${formatMoney(toplamTaahhut)}</div>
                </div>
                <div style="background:rgba(34,197,94,0.04); border-left:4px solid #22c55e; padding:0.75rem 1rem; border-radius:6px;">
                    <div style="font-size:0.75rem; color:var(--text-muted); text-transform:uppercase; font-weight:600;">Toplam fiili</div>
                    <div style="font-size:1.15rem; font-weight:800; color:#15803d; margin-top:0.2rem;">${talepOdemeler.length ? formatMoney(toplamFiili) : '—'}</div>
                </div>
                <div style="background:rgba(234,88,12,0.04); border-left:4px solid #ea580c; padding:0.75rem 1rem; border-radius:6px;">
                    <div style="font-size:0.75rem; color:var(--text-muted); text-transform:uppercase; font-weight:600;">Toplam fark</div>
                    <div style="font-size:1.15rem; font-weight:800; margin-top:0.2rem;">${talepOdemeler.length ? formatMoney(toplamFiili - toplamTaahhut) : '—'}</div>
                </div>
            </div>

            <h4 style="font-size:0.95rem; margin-bottom:0.5rem;"><i class="fas fa-boxes"></i> Ürün / hizmet kalemleri</h4>
            <div style="overflow-x:auto; margin-bottom:1.25rem;">
                <table style="width:100%; border-collapse:collapse; font-size:0.85rem;">
                    <thead>
                        <tr style="background:var(--bg-surface-light,#f1f5f9);">
                            <th style="padding:0.5rem 0.75rem; text-align:left;">Malzeme / hizmet</th>
                            <th style="padding:0.5rem 0.75rem; text-align:center;">Miktar</th>
                            <th style="padding:0.5rem 0.75rem; text-align:right;">Birim fiyat</th>
                            <th style="padding:0.5rem 0.75rem; text-align:right;">Taahhüt</th>
                            <th style="padding:0.5rem 0.75rem; text-align:right;">Fiili alım</th>
                            <th style="padding:0.5rem 0.75rem; text-align:right;">Fark</th>
                            <th style="padding:0.5rem 0.75rem; text-align:center;">Yön</th>
                            <th style="padding:0.5rem 0.75rem; text-align:left;">Bütçe kalemi</th>
                        </tr>
                    </thead>
                    <tbody>${itemRows}</tbody>
                </table>
            </div>

            ${budgetSections.length ? `
                <h4 style="font-size:0.95rem; margin-bottom:0.5rem;"><i class="fas fa-wallet"></i> Bütçe kalemi kalan tutarları</h4>
                <div style="background:var(--bg-surface-light,#f8fafc); border:1px solid var(--border-color); border-radius:8px; padding:1rem; margin-bottom:1.25rem;">
                    ${budgetSections.join('')}
                    <p style="font-size:0.78rem; color:var(--text-muted); margin:0;"><i class="fas fa-info-circle"></i> Mutabakat sonrası kalan tutarlar güncel bütçe raporundan alınır; öncesi değerler mutabakat hareketine göre geri hesaplanır.</p>
                </div>
            ` : (t0.durum === 'Onaylandı' ? '<p style="font-size:0.85rem;color:var(--text-muted);margin-bottom:1rem;"><i class="fas fa-hourglass-half"></i> Mutabakat tamamlanmadığı için fiili tutar ve bütçe etkisi henüz oluşmamıştır.</p>' : '')}

            <div style="display:flex; flex-direction:column; gap:0.75rem;">
                <div style="background:var(--bg-surface-light,#f8fafc); padding:0.75rem 1rem; border-radius:6px; border:1px solid #e2e8f0; font-size:0.88rem;">
                    <strong><i class="fas fa-comment-alt"></i> Talep gerekçesi:</strong>
                    <p style="margin:0.25rem 0 0 0; color:var(--text-secondary);">${escapeHtml(gerekcelerStr)}</p>
                </div>
                ${revizyonGerekceleri ? `<div style="background:#eff6ff; padding:0.75rem 1rem; border-radius:6px; border:1px solid #bfdbfe; font-size:0.88rem; color:#1e40af;"><strong><i class="fas fa-edit"></i> TTO revizyon gerekçesi:</strong><p style="margin:0.25rem 0 0 0;">${escapeHtml(revizyonGerekceleri)}</p></div>` : ''}
                ${redNedenleri ? `<div style="background:#fef2f2; padding:0.75rem 1rem; border-radius:6px; border:1px solid #fecaca; font-size:0.88rem; color:#991b1b;"><strong><i class="fas fa-exclamation-circle"></i> Red nedeni:</strong><p style="margin:0.25rem 0 0 0;">${escapeHtml(redNedenleri)}</p></div>` : ''}
            </div>
        `;
    }

    // openSaDetayModal seçilen talep grubunun detay modalını açar
    window.openSaDetayModal = async function(projeId, talepNo) {
        const modal = document.getElementById('saDetayModal');
        const content = document.getElementById('saDetayContent');
        const title = document.getElementById('saDetayTitle');
        if (!modal || !content) return;

        if (title) title.textContent = 'Satın Alma Talebi Detayı — ' + talepNo;
        content.innerHTML = '<div style="text-align:center;padding:3rem;"><i class="fas fa-spinner fa-spin fa-2x"></i><br><br>Detaylar yükleniyor...</div>';
        modal.style.display = 'flex';

        const token = getAuthToken();
        const headers = { 'Authorization': 'Bearer ' + token };
        let talepler = [];
        let odemeler = [];
        let butceKalemler = [];

        try {
            const [resTalep, resOdeme, resButce] = await Promise.all([
                fetch('/api/satinalma/proje/' + projeId, { headers }),
                fetch('/api/satinalma/proje/' + projeId + '/odemeler', { headers }),
                fetch('/api/satinalma/proje/' + projeId + '/butce-raporu', { headers })
            ]);
            if (resTalep.ok) {
                const data = await resTalep.json();
                talepler = (data.talepler || []).filter(t => t.talep_no === talepNo);
            }
            if (resOdeme.ok) {
                const data = await resOdeme.json();
                odemeler = data.odemeler || [];
            }
            if (resButce.ok) {
                const data = await resButce.json();
                butceKalemler = data.kalemler || [];
            }
        } catch (e) {
            console.error(e);
            content.innerHTML = '<div style="text-align:center;padding:2rem;color:var(--error);">Bağlantı hatası oluştu.</div>';
            return;
        }

        content.innerHTML = renderSatinalmaDetayHtml(talepler, odemeler, butceKalemler, talepNo);
    };

    window.closeSaDetayModal = function() {
        const modal = document.getElementById('saDetayModal');
        if (modal) modal.style.display = 'none';
    };
})();
