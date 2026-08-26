// /frontend/static/js/pdf_preview.js
// Türkçe Bilgilendirme: Tüm sistem genelinde ortak PDF görüntüleme (önizleme) modal yapısını dinamik olarak oluşturan ve yöneten modül.

(function() {
    // PDF spin animasyonu için CSS kuralını dinamik ekle
    const addPdfStyles = () => {
        if (document.getElementById('pdf-preview-styles')) return;
        const style = document.createElement('style');
        style.id = 'pdf-preview-styles';
        style.innerHTML = `
            @keyframes pdfSpin {
                0% { transform: rotate(0deg); }
                100% { transform: rotate(360deg); }
            }
            .pdf-spin {
                width: 48px;
                height: 48px;
                border: 4px solid #eee;
                border-top-color: var(--primary, #264A96);
                border-radius: 50%;
                animation: pdfSpin 0.8s linear infinite;
            }
        `;
        document.head.appendChild(style);
    };

    // Modal HTML yapısını sayfa yüklenince (veya ilk çağrıda) dinamik olarak body'ye ekle
    const initPdfModal = () => {
        if (document.getElementById('pdfModal')) return;
        addPdfStyles();

        const modalDiv = document.createElement('div');
        modalDiv.id = 'pdfModal';
        modalDiv.className = 'custom-modal-overlay';
        modalDiv.style.cssText = 'display: none; justify-content: center; align-items: center; z-index: 2000; position: fixed; top: 0; left: 0; width: 100%; height: 100%; background: rgba(0, 0, 0, 0.5); backdrop-filter: blur(4px);';
        
        modalDiv.innerHTML = `
            <div class="custom-modal" style="width: 96%; height: 94%; max-width: 1400px; display: flex; flex-direction: column; overflow: hidden; padding: 0; background: var(--bg-surface, #fff); border-radius: var(--radius-lg, 8px); box-shadow: var(--shadow-xl); border: 1px solid var(--border-color, #e2e8f0);">
                <!-- Modal Header -->
                <div class="custom-modal-header" style="padding: 1rem 1.5rem; border-bottom: 1px solid var(--border-color, #e2e8f0); display: flex; justify-content: space-between; align-items: center; width: 100%; box-sizing: border-box;">
                    <h2 id="pdfModalTitle" style="margin: 0; color: var(--primary, #264A96); font-size: 1.25rem; display: flex; align-items: center; gap: 0.5rem; font-weight: 700;">
                        <i class="fas fa-file-pdf"></i> Doküman Detayları (PDF)
                    </h2>
                    <button class="custom-modal-close" onclick="closePdfModal()" style="background: none; border: none; font-size: 1.75rem; cursor: pointer; color: var(--text-muted, #64748b);">&times;</button>
                </div>
                <!-- PDF Content -->
                <div id="pdfIframeContainer" style="flex: 1; padding: 0; overflow: hidden; position: relative; width: 100%; min-height: 400px; background: #f8fafc;">
                    <div id="pdfLoadingSpinner" style="position: absolute; top: 50%; left: 50%; transform: translate(-50%, -50%); display: flex; flex-direction: column; align-items: center; gap: 1rem; z-index: 5;">
                        <div class="pdf-spin"></div>
                        <span style="color: var(--text-secondary, #475569); font-size: 0.9rem; font-weight: 500;">PDF yükleniyor...</span>
                    </div>
                    <iframe id="pdfPreviewIframe" src="" style="width: 100%; height: 100%; border: none; display: block;"></iframe>
                </div>
                <!-- Modal Footer -->
                <div class="custom-modal-footer" style="padding: 1rem 1.5rem; border-top: 1px solid var(--border-color, #e2e8f0); display: flex; justify-content: flex-end; width: 100%; box-sizing: border-box; background: var(--bg-surface, #fff);">
                    <button type="button" class="btn btn-outline" onclick="closePdfModal()" style="padding: 0.5rem 1.5rem; cursor: pointer; border-radius: var(--radius-md, 6px); font-weight: 600;">Kapat</button>
                </div>
            </div>
        `;
        
        // Modalın arka planına tıklandığında kapanmasını sağla
        modalDiv.addEventListener('click', (e) => {
            if (e.target === modalDiv) {
                closePdfModal();
            }
        });

        document.body.appendChild(modalDiv);
    };

    // Global fonksiyon: PDF belgesini indirip blob url ile modal içinde açar
    window.viewPDF = async function(url, title = 'Doküman Detayları (PDF)', options = null) {
        initPdfModal();
        const token = localStorage.getItem('jwt_token');
        const modal = document.getElementById('pdfModal');
        const iframe = document.getElementById('pdfPreviewIframe');
        const spinner = document.getElementById('pdfLoadingSpinner');
        const titleEl = document.getElementById('pdfModalTitle');

        if (!modal || !iframe || !spinner) return;

        if (titleEl) {
            titleEl.innerHTML = `<i class="fas fa-file-pdf"></i> ${title}`;
        }

        modal.style.display = 'flex';
        spinner.style.display = 'flex';
        iframe.style.display = 'none';

        try {
            const headers = { 'Authorization': 'Bearer ' + token };
            const fetchOptions = { headers };
            
            if (options) {
                if (options.method) fetchOptions.method = options.method;
                if (options.body) {
                    fetchOptions.body = options.body;
                    headers['Content-Type'] = 'application/json';
                }
            }

            const response = await fetch(url, fetchOptions);

            if (!response.ok) {
                throw new Error('PDF yüklenemedi');
            }

            const pdfBlob = await response.blob();
            const pdfUrl = URL.createObjectURL(pdfBlob);

            iframe.onload = () => {
                spinner.style.display = 'none';
                iframe.style.display = 'block';
            };
            iframe.src = pdfUrl + '#view=FitH&navpanes=0';
        } catch (err) {
            console.error('PDF Önizleme Hatası:', err);
            alert('PDF yüklenirken bir hata oluştu veya bu belgeye erişim yetkiniz bulunmuyor.');
            closePdfModal();
        }
    };

    // Global fonksiyon: PDF modalını kapatıp blob url'i bellekten temizler
    window.closePdfModal = function() {
        const modal = document.getElementById('pdfModal');
        const iframe = document.getElementById('pdfPreviewIframe');
        if (modal) modal.style.display = 'none';
        if (iframe && iframe.src) {
            URL.revokeObjectURL(iframe.src.split('#')[0]);
            iframe.src = '';
        }
    };
})();
