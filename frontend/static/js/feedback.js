// frontend/static/js/feedback.js
// Türkçe Yorum: Sağ taraftan kayarak açılan Geri Bildirim Modülü istemci kütüphanesi.

(function () {
    // Türkçe Yorum: Giriş, kayıt ve temel kök rotalarında modülü çalıştırmıyoruz.
    const pathname = window.location.pathname;
    if (pathname === '/' || pathname === '/login' || pathname === '/register') {
        return;
    }

    // Türkçe Yorum: Kullanıcının JWT token bilgisi yoksa çalıştırılmamalıdır.
    const token = localStorage.getItem('jwt_token');
    if (!token) return;

    // DOM yüklendiğinde yetki kontrolü yapılır
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', checkPermissionAndInitFeedback);
    } else {
        checkPermissionAndInitFeedback();
    }

    async function checkPermissionAndInitFeedback() {
        try {
            // Türkçe Yorum: Kullanıcının geri bildirim modülüne erişim yetkisi olup olmadığını sorguluyoruz.
            const checkRes = await fetch('/api/auth/check-page-access?path=' + encodeURIComponent('/api/feedback'), {
                headers: { 'Authorization': 'Bearer ' + token }
            });
            if (checkRes.ok) {
                const checkData = await checkRes.json();
                if (checkData.allowed) {
                    initFeedbackModule();
                }
            }
        } catch (e) {
            console.error('Geri bildirim yetki kontrolü hatası:', e);
        }
    }

    function initFeedbackModule() {
        if (document.getElementById('feedbackToggleBtn')) return;

        const body = document.body;

        // 1. Sağ Taraf Yüzen Aç/Kapat Butonu
        const toggleBtn = document.createElement('button');
        toggleBtn.id = 'feedbackToggleBtn';
        toggleBtn.className = 'feedback-toggle-btn';
        toggleBtn.title = 'Geri Bildirim Gönder';
        toggleBtn.setAttribute('aria-label', 'Geri Bildirim Gönder');
        toggleBtn.innerHTML = `
            <i class="fas fa-comment-alt"></i>
            <span class="feedback-toggle-label">Geri Bildirim</span>
        `;
        body.appendChild(toggleBtn);

        // 2. Karartma Arka Planı (Overlay)
        const overlay = document.createElement('div');
        overlay.id = 'feedbackOverlay';
        overlay.className = 'feedback-overlay';
        body.appendChild(overlay);

        // 3. Sağ Çekmece (Drawer) Paneli
        const drawer = document.createElement('div');
        drawer.id = 'feedbackDrawer';
        drawer.className = 'feedback-drawer';
        drawer.innerHTML = `
            <div class="feedback-header">
                <div class="feedback-header-title">
                    <i class="fas fa-envelope-open-text"></i>
                    <span>Geri Bildirim İletin</span>
                </div>
                <button class="feedback-close-btn" id="feedbackCloseBtn" title="Kapat">
                    <i class="fas fa-times"></i>
                </button>
            </div>

            <div class="feedback-body">
                <div class="feedback-info-box">
                    <i class="fas fa-info-circle"></i>
                    <span>İlettiğiniz geri bildirimler profil bilgileriniz ve sistem yetkileriniz ile birlikte e-posta olarak Sistem Yöneticilerine (Admin) iletilmektedir.</span>
                </div>

                <form id="feedbackForm">
                    <div class="feedback-form-group">
                        <label for="feedbackSubject">Konu / Kategori <span style="color:red;">*</span></label>
                        <select id="feedbackSubject" class="feedback-input" required>
                            <option value="Öneri / İstek">💡 Öneri / İstek</option>
                            <option value="Hata Bildirimi">🐛 Hata / Problem Bildirimi</option>
                            <option value="Sistem Kullanım Desteği">❓ Kullanım Desteği</option>
                            <option value="Diğer">📝 Diğer</option>
                        </select>
                    </div>

                    <div class="feedback-form-group">
                        <label for="feedbackMessage">Mesajınız <span style="color:red;">*</span></label>
                        <textarea id="feedbackMessage" class="feedback-textarea" rows="6" placeholder="Görüş, öneri veya karşılaştığınız hatayı detaylıca buraya yazabilirsiniz..." required></textarea>
                    </div>

                    <div class="feedback-form-group">
                        <label>Gönderilen Sayfa Bilgisi</label>
                        <div class="feedback-page-badge">
                            <i class="fas fa-link"></i>
                            <span id="feedbackCurrentUrl"></span>
                        </div>
                    </div>

                    <div id="feedbackAlert" class="feedback-alert" style="display:none;"></div>

                    <div class="feedback-footer">
                        <button type="button" class="feedback-btn-cancel" id="feedbackCancelBtn">İptal</button>
                        <button type="submit" class="feedback-btn-send" id="feedbackSendBtn">
                            <i class="fas fa-paper-plane"></i>
                            <span>Gönder</span>
                        </button>
                    </div>
                </form>
            </div>
        `;
        body.appendChild(drawer);

        // Current URL göster
        const currentUrlSpan = document.getElementById('feedbackCurrentUrl');
        if (currentUrlSpan) {
            currentUrlSpan.textContent = window.location.pathname + window.location.search;
        }

        // 4. Event Listener'ları Bağla
        const closeBtn = document.getElementById('feedbackCloseBtn');
        const cancelBtn = document.getElementById('feedbackCancelBtn');
        const form = document.getElementById('feedbackForm');
        const alertBox = document.getElementById('feedbackAlert');

        function openDrawer() {
            drawer.classList.add('open');
            overlay.classList.add('open');
            if (currentUrlSpan) {
                currentUrlSpan.textContent = window.location.pathname + window.location.search;
            }
            document.getElementById('feedbackMessage').focus();
        }

        function closeDrawer() {
            drawer.classList.remove('open');
            overlay.classList.remove('open');
            if (alertBox) {
                alertBox.style.display = 'none';
                alertBox.className = 'feedback-alert';
            }
        }

        toggleBtn.addEventListener('click', openDrawer);
        closeBtn.addEventListener('click', closeDrawer);
        cancelBtn.addEventListener('click', closeDrawer);
        overlay.addEventListener('click', closeDrawer);

        // ESC tuşu ile kapatma
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && drawer.classList.contains('open')) {
                closeDrawer();
            }
        });

        // Form Submit İletişimi
        form.addEventListener('submit', async (e) => {
            e.preventDefault();

            const subject = document.getElementById('feedbackSubject').value;
            const message = document.getElementById('feedbackMessage').value;
            const sendBtn = document.getElementById('feedbackSendBtn');

            if (!message.trim()) {
                showAlert('Lütfen bir mesaj yazınız.', 'error');
                return;
            }

            // UI Yükleniyor durumuna getir
            sendBtn.disabled = true;
            sendBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> <span>Gönderiliyor...</span>';

            try {
                const response = await fetch('/api/feedback', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'Authorization': 'Bearer ' + token
                    },
                    body: JSON.stringify({
                        konu: subject,
                        mesaj: message,
                        sayfa_url: window.location.href
                    })
                });

                const data = await response.json();

                if (response.ok && data.success) {
                    showAlert(data.message || 'Geri bildiriminiz yöneticilere başarıyla iletildi.', 'success');
                    form.reset();
                    setTimeout(() => {
                        closeDrawer();
                    }, 2000);
                } else {
                    showAlert(data.error || 'Geri bildirim gönderilirken bir hata oluştu.', 'error');
                }
            } catch (err) {
                console.error('Geri bildirim gönderme hatası:', err);
                showAlert('Sunucuya bağlanırken bir hata oluştu.', 'error');
            } finally {
                sendBtn.disabled = false;
                sendBtn.innerHTML = '<i class="fas fa-paper-plane"></i> <span>Gönder</span>';
            }
        });

        function showAlert(msg, type) {
            if (!alertBox) return;
            alertBox.textContent = msg;
            alertBox.className = 'feedback-alert ' + (type === 'success' ? 'alert-success' : 'alert-error');
            alertBox.style.display = 'block';
        }
    }
})();
