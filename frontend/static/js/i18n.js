// frontend/static/js/i18n.js

document.addEventListener('DOMContentLoaded', async () => {
    // 1. Dil Seçeneğini Yükle (Varsayılan: tr)
    let currentLang = localStorage.getItem('bap_lang') || 'tr';

    // Uygulama genelinde kullanılacak çeviri fonksiyonunu tanımla
    window.t = function (key, params = {}) {
        const langData = window.translations[currentLang];
        if (!langData) return key;

        let text = langData[key] || key;

        // Parametreleri değiştir (ör. {name} => Ali)
        for (const [k, v] of Object.entries(params)) {
            text = text.replace(new RegExp(`{${k}}`, 'g'), v);
        }

        return text;
    };

    // 2. Sayfadaki data-i18n etiketlerini çevir
    function applyTranslations() {
        // Set document text direction based on active language (ar is RTL, others LTR)
        document.documentElement.dir = currentLang === 'ar' ? 'rtl' : 'ltr';

        const elements = document.querySelectorAll('[data-i18n]');
        elements.forEach(el => {
            const key = el.getAttribute('data-i18n');
            const translation = window.t(key);

            // Çeviri bulunamazsa (key kendisi dönüyorsa) atla
            if (translation === key) return;

            const tag = el.tagName;

            // Input ve Textarea için placeholder/value güncelle
            if (tag === 'INPUT' || tag === 'TEXTAREA') {
                if (el.hasAttribute('placeholder')) {
                    el.placeholder = translation;
                } else if (el.type === 'button' || el.type === 'submit') {
                    el.value = translation;
                }
            }
            // Option ve Button elementleri için doğrudan textContent güncelle
            else if (tag === 'OPTION' || tag === 'BUTTON') {
                el.textContent = translation;
            }
            // Diğer elementler (span, strong, h3, td, th, vb.)
            else {
                // Eğer elementin içi sadece metin ise (çocuk element yoksa)
                if (el.children.length === 0) {
                    el.textContent = translation;
                } else {
                    // Child node'lar arasında text node bul ve güncelle
                    let textNodeUpdated = false;
                    for (let i = 0; i < el.childNodes.length; i++) {
                        if (el.childNodes[i].nodeType === 3 && el.childNodes[i].nodeValue.trim().length > 0) {
                            el.childNodes[i].nodeValue = ' ' + translation;
                            textNodeUpdated = true;
                            break;
                        }
                    }
                    // Text node bulunamadıysa sonuna ekle
                    if (!textNodeUpdated) {
                        el.appendChild(document.createTextNode(' ' + translation));
                    }
                }
            }
        });

        document.documentElement.lang = currentLang;
    }

    // İlk yüklemede çevirileri uygula
    applyTranslations();

    // 3. Otomatik Dil Seçici (Language Switcher) Ekleme
    // Sayfada "#lang-switcher" ID'li bir kap varsa veya topbar/right-side varsa oraya ekleyeceğiz.
    const topBarRight = document.querySelector('.topbar-right') || document.querySelector('.auth-header');

    if (topBarRight && !document.getElementById('lang-selector-container')) {
        const langContainer = document.createElement('div');
        langContainer.id = 'lang-selector-container';
        langContainer.style.display = 'inline-flex';
        langContainer.style.alignItems = 'center';
        langContainer.style.marginRight = '15px';

        const langToggleBtn = document.createElement('button');
        langToggleBtn.id = 'langToggleBtn';
        langToggleBtn.className = 'theme-toggle';

        const tooltipMap = {
            'tr': 'English / العربية',
            'en': 'Türkçe / العربية',
            'ar': 'Türkçe / English'
        };
        langToggleBtn.title = tooltipMap[currentLang] || 'Dili Değiştir / Change Language';
        langToggleBtn.style.width = '40px';
        langToggleBtn.style.height = '40px';
        langToggleBtn.style.display = 'flex';
        langToggleBtn.style.alignItems = 'center';
        langToggleBtn.style.justifyContent = 'center';
        langToggleBtn.style.fontWeight = 'bold';
        langToggleBtn.style.fontSize = '14px';

        // Tema rengine göre ters renk:
        langToggleBtn.style.backgroundColor = 'var(--text-primary)';
        langToggleBtn.style.color = 'var(--bg-surface)';
        langToggleBtn.style.border = 'none';

        // Gösterilecek metin (Mevcut ekrandaki dili göster)
        langToggleBtn.innerHTML = currentLang.toUpperCase();

        // Değişim olayını dinle (tr -> en -> ar -> tr)
        langToggleBtn.addEventListener('click', (e) => {
            e.preventDefault();
            e.stopPropagation();
            if (currentLang === 'tr') {
                currentLang = 'en';
            } else if (currentLang === 'en') {
                currentLang = 'ar';
            } else {
                currentLang = 'tr';
            }
            localStorage.setItem('bap_lang', currentLang);
            applyTranslations();

            // Eğer sayfa özelinde ekstra Javascript güncellemesi gerekiyorsa (ör dashboard js'i gibi),
            // CustomEvent fırlatabiliriz.
            document.dispatchEvent(new CustomEvent('languageChanged', { detail: { lang: currentLang } }));

            // Kolaylık olsun diye sayfayı yeniliyoruz, böylece template render sırasında ekli js verileri de güncellensin.
            window.location.reload();
        });

        langContainer.appendChild(langToggleBtn);

        if (document.querySelector('.topbar-right')) {
            // Theme toggle'ın yanına ekle
            const themeToggle = document.getElementById('themeToggle');
            if (themeToggle) {
                document.querySelector('.topbar-right').insertBefore(langContainer, themeToggle);
            } else {
                document.querySelector('.topbar-right').prepend(langContainer);
            }
        } else if (document.querySelector('.auth-header')) {
            langContainer.style.position = 'absolute';
            langContainer.style.top = '15px';
            langContainer.style.right = '15px';
            document.querySelector('.auth-card').style.position = 'relative';
            document.querySelector('.auth-card').appendChild(langContainer);
        }
    }

    // 4. Dinamik Sidebar Menüsü Oluşturma (Garantili Fallback Sistemi)
    const sidebarMenu = document.querySelector('.sidebar-menu');
    if (sidebarMenu) {
        const token = localStorage.getItem('jwt_token');
        let allowedPages = [];
        let userRole = '';

        if (token) {
            try {
                const payload = JSON.parse(atob(token.split('.')[1]));
                userRole = (payload.rol || payload.role || '').toLowerCase();
            } catch(e){}

            try {
                const res = await fetch('/api/auth/my-allowed-pages', {
                    headers: { 'Authorization': 'Bearer ' + token }
                });
                if (res.ok) {
                    const data = await res.json();
                    allowedPages = data.allowed_pages || [];
                }
            } catch (e) {
                console.error('Failed to load allowed pages:', e);
            }
        }

        const path = window.location.pathname;

        // Fallback: allowedPages boşsa sayfa yolu veya kullanıcı rolüne göre varsayılan yetki listesi oluştur
        if (!allowedPages.length) {
            if (userRole === 'admin' || path.startsWith('/admin')) {
                allowedPages = ['/admin/dashboard', '/admin/hakem-atama', '/admin/proje-basvurulari', '/admin/zamanlanmis-gorevler', '/admin/projects/status', '/anasayfa', '/eimza'];
            } else if (userRole === 'dekan' || path.startsWith('/dekan')) {
                allowedPages = ['/dekan/dashboard', '/anasayfa'];
            } else if (userRole === 'komisyon' || userRole === 'komisyon_baskani' || path.startsWith('/komisyon')) {
                allowedPages = ['/komisyon/dashboard', '/komisyon/baskan/dashboard', '/anasayfa'];
            } else if (userRole === 'tto' || path.startsWith('/tto')) {
                allowedPages = ['/tto/dashboard', '/tto/satinalma', '/tto/satinalma/mutabakat', '/tto/talepler', '/admin/proje-basvurulari', '/anasayfa'];
            } else if (userRole === 'hakem' || path.startsWith('/hakem')) {
                allowedPages = ['/hakem/dashboard', '/anasayfa'];
            } else {
                allowedPages = ['/anasayfa', '/basvuru', '/satinalma', '/profil'];
            }
        }

        try {
            const search = window.location.search;
            const isPageAdmin = path.startsWith('/admin') || userRole === 'admin';

            // Türkçe Yorum: Sidebar menüsünün TAMAMI tek bir tanım listesinden (registry) üretilir.
            // Hiçbir menü öğesi sabit HTML olarak yazılmaz. Her öğe, 'yetki' alanındaki sayfa yolu
            // kullanıcının yetki listesinde (my-allowed-pages) bulunuyorsa çizilir.
            // 'temel: true' olan öğeler yetki aranmadan gösterilir (ör. Anasayfa).
            // Hiç görünür öğesi olmayan bölümün başlığı da çizilmez.
            const menuBolumleri = [
                {
                    etiket: window.t('nav.admin_menu'),
                    ogeler: [
                        { href: '/admin/dashboard', icon: 'fa-shield-alt', etiket: window.t('nav.admin_panel'), aktifMi: () => path === '/admin/dashboard' && !search.includes('tab') },
                        { href: '/admin/hakem-atama', icon: 'fa-user-check', etiket: window.t('nav.referee_assign') },
                        { href: '/admin/dashboard?tab=usersTab', yetki: '/admin/dashboard', icon: 'fa-users-cog', etiket: window.t('nav.user_management'), aktifMi: () => search.includes('tab=usersTab') },
                        { href: '/admin/dashboard?tab=bapTab', yetki: '/admin/dashboard', icon: 'fa-folder-plus', etiket: window.t('nav.bap_definition'), aktifMi: () => search.includes('tab=bapTab') },
                        { href: '/admin/proje-basvurulari', icon: 'fa-file-signature', renk: '#7c3aed', etiket: window.t('nav.project_applications'), aktifMi: () => path === '/admin/proje-basvurulari' || search.includes('section=talepler') },
                        { href: '/admin/zamanlanmis-gorevler', icon: 'fa-clock', renk: '#f59e0b', etiket: window.t('nav.scheduled_tasks') }
                    ]
                },
                {
                    etiket: window.t('nav.reports'),
                    ogeler: [
                        { href: '/admin/projects/status', icon: 'fa-chart-pie', etiket: window.t('nav.status_reports') }
                    ]
                },
                {
                    etiket: window.t('nav.dekan_menu'),
                    ogeler: [
                        { href: '/dekan/dashboard', icon: 'fa-university', etiket: window.t('nav.dekan_panel') }
                    ]
                },
                {
                    etiket: window.t('nav.komisyon_menu'),
                    ogeler: [
                        { href: '/komisyon/dashboard', icon: 'fa-gavel', etiket: window.t('nav.komisyon_panel') },
                        { href: '/komisyon/baskan/dashboard', icon: 'fa-tasks', etiket: window.t('nav.komisyon_yonetim') }
                    ]
                },
                {
                    etiket: window.t('nav.tto_menu'),
                    ogeler: [
                        { href: '/tto/dashboard', icon: 'fa-rocket', etiket: window.t('nav.tto_panel'), aktifMi: () => path === '/tto/dashboard' && !search.includes('section=satinalma') },
                        { href: '/tto/satinalma', icon: 'fa-shopping-cart', etiket: window.t('nav.purchasing_management_tto'), aktifMi: () => path === '/tto/satinalma' || search.includes('section=satinalma') },
                        { href: '/tto/talepler', icon: 'fa-file-signature', renk: '#7c3aed', etiket: window.t('nav.project_requests_tto'), aktifMi: () => path === '/tto/talepler' || search.includes('section=talepler') }
                    ]
                },
                {
                    etiket: window.t('nav.referee_menu'),
                    ogeler: [
                        { href: '/hakem/dashboard', icon: 'fa-gavel', etiket: window.t('nav.referee_panel'), aktifMi: () => path === '/hakem/dashboard' || path.startsWith('/hakem/degerlendirme') }
                    ]
                },
                {
                    etiket: isPageAdmin ? window.t('dash.quick_actions') : window.t('nav.main_menu'),
                    ogeler: [
                        { href: '/anasayfa', icon: 'fa-home', etiket: window.t('nav.dashboard'), temel: true },
                        { href: '/basvuru', icon: 'fa-plus-circle', etiket: window.t('nav.new_application'), adminSayfasindaGizle: true },
                        { href: '/satinalma', icon: 'fa-shopping-cart', etiket: window.t('nav.purchase_requests'), adminSayfasindaGizle: true },
                        { href: '/eimza', icon: 'fa-signature', etiket: window.t('nav.eimza') }
                    ]
                }
            ];

            // ogeGorunurMu bir menü öğesinin kullanıcının yetkilerine göre çizilip çizilmeyeceğini belirler
            const ogeGorunurMu = (oge) => {
                if (oge.adminSayfasindaGizle && isPageAdmin) return false;
                if (oge.temel) return true;
                const yetkiYolu = oge.yetki || oge.href.split('?')[0];
                return allowedPages.includes(yetkiYolu);
            };

            // ogeHTML tek bir menü öğesinin HTML çıktısını üretir
            const ogeHTML = (oge) => {
                const aktif = oge.aktifMi ? oge.aktifMi() : path === oge.href.split('?')[0];
                const renkStil = oge.renk ? ` style="color:${oge.renk};"` : '';
                return `
                        <li class="menu-item ${aktif ? 'active' : ''}">
                            <a href="${oge.href}">
                                <i class="fas ${oge.icon}"${renkStil}></i>
                                <span>${oge.etiket}</span>
                            </a>
                        </li>`;
            };

            // Türkçe Yorum: Bölümler sırayla gezilir; görünür öğesi olan bölümler menüye eklenir.
            let menuHTML = '';
            menuBolumleri.forEach(bolum => {
                const gorunurOgeler = bolum.ogeler.filter(ogeGorunurMu);
                if (!gorunurOgeler.length) return;
                menuHTML += `
                    <div class="menu-label">${bolum.etiket}</div>
                    <ul class="menu-list">${gorunurOgeler.map(ogeHTML).join('')}
                    </ul>
                `;
            });

            sidebarMenu.innerHTML = menuHTML;
        } catch (e) {
            console.error('Sidebar build error:', e);
        }
    }

    // 4.5. Dairesel Tema Seçici Yönetimi (Açık, Koyu, Açık Mezuniyet, Koyu Mezuniyet)
    const oldToggle = document.getElementById('themeToggle');
    if (oldToggle) {
        // Türkçe Yorum: Butonu klonlayarak üzerindeki diğer tüm olay dinleyicilerini (event listener) sıfırlıyoruz.
        const newToggle = oldToggle.cloneNode(true);
        oldToggle.parentNode.replaceChild(newToggle, oldToggle);

        const themes = ['light', 'dark', 'light-mezuniyet', 'dark-mezuniyet'];

        newToggle.addEventListener('click', (e) => {
            e.preventDefault();
            const currentTheme = localStorage.getItem('bap_theme') || 'light';
            let nextIndex = (themes.indexOf(currentTheme) + 1) % themes.length;
            if (nextIndex === -1) nextIndex = 0;
            const nextTheme = themes[nextIndex];

            // Türkçe Yorum: Temayı uyguluyoruz ve localStorage'a kaydediyoruz
            document.documentElement.setAttribute('data-theme', nextTheme);
            localStorage.setItem('bap_theme', nextTheme);

            // Türkçe Yorum: Buton başlığını (tooltip) güncelliyoruz
            let themeTitle = "Açık İZÜ Teması";
            if (nextTheme === 'dark') themeTitle = "Koyu İZÜ Teması";
            else if (nextTheme === 'light-mezuniyet') themeTitle = "Açık Mezuniyet Teması";
            else if (nextTheme === 'dark-mezuniyet') themeTitle = "Koyu Mezuniyet Teması";
            newToggle.title = themeTitle;
        });
    }

    // 5. Chatbot scriptini otomatik yükleme (Giriş sayfası, kayıt sayfası veya ana kök dizinde asistan yüklenmez)
    const pathname = window.location.pathname;
    if (pathname !== '/' && pathname !== '/login' && pathname !== '/register') {
        const chatbotScript = document.createElement('script');
        chatbotScript.src = '/static/js/chatbot.js';
        chatbotScript.defer = true;
        document.head.appendChild(chatbotScript);
    }
});

// window.formatPhoneNumber receives a string of digits and returns '(5XX) XXX XX XX' format
window.formatPhoneNumber = function (value) {
    if (!value) return '';
    let cleaned = value.replace(/\D/g, '');
    if (cleaned.startsWith('0')) {
        cleaned = cleaned.substring(1);
    }
    cleaned = cleaned.substring(0, 10);
    if (cleaned.length === 0) return '';

    let result = '';
    if (cleaned.length > 0) {
        result += '(' + cleaned.substring(0, Math.min(cleaned.length, 3));
    }
    if (cleaned.length >= 3) {
        result += ') ';
    }
    if (cleaned.length > 3) {
        result += cleaned.substring(3, Math.min(cleaned.length, 6));
    }
    if (cleaned.length > 6) {
        result += ' ' + cleaned.substring(6, Math.min(cleaned.length, 8));
    }
    if (cleaned.length > 8) {
        result += ' ' + cleaned.substring(8, Math.min(cleaned.length, 10));
    }
    return result;
};

// window.applyPhoneMaskToInput attaches key listeners to format and mask inputs in (5XX) XXX XX XX writing format
window.applyPhoneMaskToInput = function (input) {
    if (!input) return;

    input.setAttribute('maxlength', '15');
    input.setAttribute('placeholder', '(5XX) XXX XX XX');

    input.addEventListener('input', function (e) {
        let cursorPosition = e.target.selectionStart;
        let originalValue = e.target.value;
        let digitsOnly = originalValue.replace(/\D/g, '');
        if (digitsOnly.startsWith('0')) {
            digitsOnly = digitsOnly.substring(1);
        }
        digitsOnly = digitsOnly.substring(0, 10);

        let formatted = '';
        if (digitsOnly.length > 0) {
            formatted += '(' + digitsOnly.substring(0, Math.min(digitsOnly.length, 3));
        }
        if (digitsOnly.length >= 3) {
            formatted += ') ';
        }
        if (digitsOnly.length > 3) {
            formatted += digitsOnly.substring(3, Math.min(digitsOnly.length, 6));
        }
        if (digitsOnly.length > 6) {
            formatted += ' ' + digitsOnly.substring(6, Math.min(digitsOnly.length, 8));
        }
        if (digitsOnly.length > 8) {
            formatted += ' ' + digitsOnly.substring(8, Math.min(digitsOnly.length, 10));
        }

        e.target.value = formatted;

        let digitsBeforeCursor = originalValue.substring(0, cursorPosition).replace(/\D/g, '').length;
        if (originalValue.startsWith('0') && cursorPosition > 0) {
            digitsBeforeCursor = Math.max(0, digitsBeforeCursor - 1);
        }

        let newCursor = 0;
        let digitCount = 0;
        for (let i = 0; i < formatted.length; i++) {
            if (/\d/.test(formatted[i])) {
                digitCount++;
            }
            newCursor = i + 1;
            if (digitCount === digitsBeforeCursor) {
                while (newCursor < formatted.length && /\D/.test(formatted[newCursor])) {
                    newCursor++;
                }
                break;
            }
        }
        e.target.setSelectionRange(newCursor, newCursor);
    });
};

// applySidebarPermissions sol menüdeki linkleri kullanıcının yetkilerine göre dinamik olarak gizler/gösterir.
window.applySidebarPermissions = async function () {
    const token = localStorage.getItem('jwt_token');
    if (!token) return;

    try {
        const res = await fetch('/api/auth/my-allowed-pages', {
            headers: { 'Authorization': 'Bearer ' + token }
        });
        if (!res.ok) return;

        const data = await res.json();
        const allowedPages = data.allowed_pages || [];

        // Sol menüdeki tüm a elementlerini seç
        const menuLinks = document.querySelectorAll('.sidebar-menu .menu-list a');
        menuLinks.forEach(link => {
            let path = link.getAttribute('href');
            if (!path || path === '#' || path === 'javascript:void(0)') {
                const onClickAttr = link.getAttribute('onclick') || '';
                if (onClickAttr.includes('openUsersTab')) {
                    path = '/admin/dashboard';
                } else if (onClickAttr.includes('openBapTab')) {
                    path = '/admin/dashboard';
                } else if (onClickAttr.includes('openYetkiTab')) {
                    path = '/admin/dashboard';
                } else {
                    return;
                }
            }

            // Path'in query parametrelerini temizleyelim
            const cleanPath = path.split('?')[0];

            // Anasayfa ve profil sayfaları her zaman açık olmalı (temel erişim)
            if (cleanPath === '/anasayfa' || cleanPath === '/profil' || cleanPath === '/profil-tamamla') {
                return;
            }

            // Proje Başvuruları özel yetki kontrolü (href'te section=talepler veya /admin/proje-basvurulari ise)
            if (path.includes('section=talepler') || cleanPath === '/admin/proje-basvurulari') {
                const isTaleplerAllowed = allowedPages.includes('/admin/proje-basvurulari') || allowedPages.includes('/admin/dashboard');
                const menuItem = link.closest('.menu-item') || link.closest('li');
                if (menuItem) {
                    menuItem.style.display = isTaleplerAllowed ? '' : 'none';
                }
                return;
            }

            // Eğer izin verilen sayfalar listesinde bu yol yoksa, menü öğesini gizle
            const isAllowed = allowedPages.some(allowedUrl => {
                return cleanPath === allowedUrl ||
                    (cleanPath.startsWith('/admin') && allowedUrl === '/admin/dashboard');
            });

            const menuItem = link.closest('.menu-item') || link.closest('li');
            if (menuItem) {
                if (isAllowed) {
                    menuItem.style.display = '';
                } else {
                    menuItem.style.display = 'none';
                }
            }
        });
    } catch (err) {
        console.error('Sol menü yetki kontrolü hatası:', err);
    }
};

// Sayfa yüklendiğinde otomatik olarak çalıştır
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', window.applySidebarPermissions);
} else {
    window.applySidebarPermissions();
}

// initNotificationsSystem, tüm sayfalarda bildirimleri dinamik olarak yükler ve arayüze ekler.
// Türkçe Yorum: Her sayfadaki topbar-right elementinin yanına bildirim zili ekler, tıklanınca açılan menü ve modal detayını yönetir.
window.initNotificationsSystem = async function () {
    const token = localStorage.getItem('jwt_token');
    const topBarRight = document.querySelector('.topbar-right');
    if (!token || !topBarRight) return;

    // 1. Gerekli CSS Stillerini Ekle
    const styleEl = document.createElement('style');
    styleEl.textContent = `
        .notification-bell-container { position: relative; margin-right: 15px; display: inline-flex; align-items: center; }
        .notification-bell-btn { background: transparent; border: none; font-size: 1.25rem; color: var(--text-secondary); cursor: pointer; position: relative; padding: 0.5rem; transition: color 0.2s; display: flex; align-items: center; justify-content: center; }
        .notification-bell-btn:hover { color: var(--primary); }
        .notification-badge { position: absolute; top: 0; right: 0; background: var(--error, #f44336); color: white; border-radius: 50%; min-width: 16px; height: 16px; font-size: 0.65rem; display: flex; align-items: center; justify-content: center; font-weight: bold; border: 2px solid var(--bg-surface); padding: 1px; }
        .notification-dropdown { display: none; position: absolute; right: 0; top: 120%; width: 340px; background: var(--bg-surface); border: 1px solid var(--border-color); border-radius: var(--radius-md, 8px); box-shadow: var(--shadow-lg, 0 10px 15px -3px rgba(0,0,0,0.1)); z-index: 10001; max-height: 420px; overflow-y: auto; flex-direction: column; }
        .notification-dropdown.active { display: flex; }
        .notification-dropdown-header { display: flex; justify-content: space-between; align-items: center; padding: 0.75rem 1rem; border-bottom: 1px solid var(--border-color); font-weight: bold; font-size: 0.88rem; }
        .notification-dropdown-header button { background: none; border: none; color: var(--primary); font-size: 0.75rem; cursor: pointer; font-weight: 600; padding: 0; }
        .notification-list { font-size: 0.85rem; display: flex; flex-direction: column; overflow-y: auto; }
        .notification-item { padding: 0.85rem 1rem; border-bottom: 1px solid var(--border-color); cursor: pointer; transition: background 0.2s; display: flex; flex-direction: column; gap: 0.25rem; text-align: left; }
        .notification-item:hover { background: rgba(0,0,0,0.02); }
        .notification-item.unread { background: rgba(30, 60, 114, 0.04); border-left: 3px solid var(--primary); }
        .notification-item-title { font-weight: 600; color: var(--text-primary); }
        .notification-item-time { font-size: 0.72rem; color: var(--text-muted); }
    `;
    document.head.appendChild(styleEl);

    // 2. Bildirim Zili HTML Enjekte Et
    const bellContainer = document.createElement('div');
    bellContainer.className = 'notification-bell-container';
    bellContainer.innerHTML = `
        <button class="notification-bell-btn" id="notificationToggleBtn" title="Bildirimler">
            <i class="fas fa-bell"></i>
            <span class="notification-badge" id="notificationBadge" style="display: none;">0</span>
        </button>
        <div class="notification-dropdown" id="notificationDropdown">
            <div class="notification-dropdown-header">
                <span>Bildirimler</span>
                <button id="markAllReadBtn">Tümünü Okundu Yap</button>
            </div>
            <div class="notification-list" id="notificationList">
                <div style="text-align: center; padding: 1.5rem; color: var(--text-muted);">Bildirimler yükleniyor...</div>
            </div>
        </div>
    `;

    // Topbar-right içindeki profil dropdown'ın hemen soluna ekleyelim
    const profileToggle = document.getElementById('profileToggle');
    if (profileToggle) {
        topBarRight.insertBefore(bellContainer, profileToggle);
    } else {
        topBarRight.appendChild(bellContainer);
    }

    // 3. Modal HTML Enjekte Et
    if (!document.getElementById('notificationModal')) {
        const modalEl = document.createElement('div');
        modalEl.id = 'notificationModal';
        modalEl.className = 'custom-modal-overlay';
        modalEl.innerHTML = `
            <div class="custom-modal" style="max-width: 600px;">
                <div class="custom-modal-header">
                    <h2 id="notificationModalTitle">Bildirim Detayı</h2>
                    <button class="custom-modal-close" id="closeNotificationModalBtn">&times;</button>
                </div>
                <div class="custom-modal-body" id="notificationModalBody" style="padding: 1.5rem; max-height: 70vh; overflow-y: auto;">
                    <!-- Bildirim İçeriği (HTML) -->
                </div>
                <div class="custom-modal-footer">
                    <button class="btn btn-outline" id="closeNotificationModalBtn2">Kapat</button>
                </div>
            </div>
        `;
        document.body.appendChild(modalEl);
    }

    // Elemanları al
    const toggleBtn = document.getElementById('notificationToggleBtn');
    const dropdown = document.getElementById('notificationDropdown');
    const badge = document.getElementById('notificationBadge');
    const listContainer = document.getElementById('notificationList');
    const markAllBtn = document.getElementById('markAllReadBtn');
    const modal = document.getElementById('notificationModal');
    const modalTitle = document.getElementById('notificationModalTitle');
    const modalBody = document.getElementById('notificationModalBody');

    // Kapatma butonları
    const closeBtns = [
        document.getElementById('closeNotificationModalBtn'),
        document.getElementById('closeNotificationModalBtn2')
    ];
    closeBtns.forEach(btn => btn?.addEventListener('click', () => {
        modal.style.display = 'none';
    }));

    // Dropdown açma/kapama
    toggleBtn.addEventListener('click', (e) => {
        e.preventDefault();
        e.stopPropagation();
        dropdown.classList.toggle('active');
        if (dropdown.classList.contains('active')) {
            loadNotifications();
        }
    });

    // Boş yere tıklayınca dropdown'ı kapat
    document.addEventListener('click', () => {
        dropdown.classList.remove('active');
    });
    dropdown.addEventListener('click', (e) => {
        e.stopPropagation();
    });

    // Bildirimleri yükleyen ve listeleyen fonksiyon
    async function loadNotifications() {
        try {
            const res = await fetch('/api/bildirimler', {
                headers: { 'Authorization': 'Bearer ' + token }
            });
            if (!res.ok) throw new Error();
            const data = await res.json();
            const notifications = data.bildirimler || [];

            // Okunmamış sayısını güncelle
            const unreadCount = notifications.filter(n => !n.okundu).length;
            if (unreadCount > 0) {
                badge.textContent = unreadCount;
                badge.style.display = 'flex';
            } else {
                badge.style.display = 'none';
            }

            listContainer.innerHTML = '';
            if (notifications.length === 0) {
                listContainer.innerHTML = '<div style="text-align: center; padding: 2rem; color: var(--text-muted);">Henüz bildiriminiz bulunmuyor.</div>';
                return;
            }

            notifications.forEach(n => {
                const item = document.createElement('div');
                item.className = `notification-item ${n.okundu ? '' : 'unread'}`;

                const timeStr = new Date(n.olusturma_tarihi).toLocaleString('tr-TR', {
                    day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit'
                });

                item.innerHTML = `
                    <span class="notification-item-title">${n.baslik}</span>
                    <span class="notification-item-time"><i class="far fa-clock"></i> ${timeStr}</span>
                `;

                item.addEventListener('click', async () => {
                    dropdown.classList.remove('active');

                    // Modalı aç ve içeriği yükle
                    modalTitle.textContent = n.baslik;
                    modalBody.innerHTML = n.icerik;
                    modal.style.display = 'flex';

                    // Okunmadıysa okundu yap
                    if (!n.okundu) {
                        try {
                            const readRes = await fetch(`/api/bildirimler/${n.bildirim_id}/oku`, {
                                method: 'POST',
                                headers: { 'Authorization': 'Bearer ' + token }
                            });
                            if (readRes.ok) {
                                n.okundu = true;
                                item.classList.remove('unread');
                                updateBadgeCount(notifications);
                            }
                        } catch (err) {
                            console.error(err);
                        }
                    }
                });

                listContainer.appendChild(item);
            });
        } catch (err) {
            listContainer.innerHTML = '<div style="text-align: center; padding: 2rem; color: var(--error);">Bildirimler alınamadı.</div>';
        }
    }

    // Badge sayısını anlık güncelleyen yardımcı fonksiyon
    function updateBadgeCount(notifications) {
        const unreadCount = notifications.filter(n => !n.okundu).length;
        if (unreadCount > 0) {
            badge.textContent = unreadCount;
            badge.style.display = 'flex';
        } else {
            badge.style.display = 'none';
        }
    }

    // Tümünü okundu yap butonu
    const markAllAllReadBtn = async () => {
        try {
            const res = await fetch('/api/bildirimler/oku-hepsi', {
                method: 'POST',
                headers: { 'Authorization': 'Bearer ' + token }
            });
            if (res.ok) {
                loadNotifications();
            }
        } catch (err) {
            console.error(err);
        }
    };
    markAllBtn.addEventListener('click', markAllAllReadBtn);

    // İlk sayfa açılışında sessizce okunmamış sayısını çek
    async function checkUnreadCountSilent() {
        try {
            const res = await fetch('/api/bildirimler', {
                headers: { 'Authorization': 'Bearer ' + token }
            });
            if (res.ok) {
                const data = await res.json();
                const list = data.bildirimler || [];
                const unread = list.filter(n => !n.okundu).length;
                if (unread > 0) {
                    badge.textContent = unread;
                    badge.style.display = 'flex';
                } else {
                    badge.style.display = 'none';
                }
            }
        } catch (err) { }
    }

    checkUnreadCountSilent();
    // Her 60 saniyede bir yeni bildirimleri sessizce kontrol et
    setInterval(checkUnreadCountSilent, 60000);
};

// Central Profile Dropdown & Logout Handler
// Tüm sayfalarda profil menüsünün ve çıkış yap butonunun sorunsuz çalışmasını garanti eder.
window.initProfileDropdown = function () {
    const profileToggle = document.getElementById('profileToggle');
    const profileDropdown = document.getElementById('profileDropdown');
    const logoutBtn = document.getElementById('logoutBtn');

    if (profileToggle && profileDropdown && !profileToggle.dataset.dropdownInit) {
        profileToggle.dataset.dropdownInit = 'true';
        profileToggle.addEventListener('click', (e) => {
            e.stopPropagation();
            profileDropdown.classList.toggle('active');
        });
        document.addEventListener('click', (e) => {
            if (!profileToggle.contains(e.target) && !profileDropdown.contains(e.target)) {
                profileDropdown.classList.remove('active');
            }
        });
    }

    if (logoutBtn && !logoutBtn.dataset.logoutInit) {
        logoutBtn.dataset.logoutInit = 'true';
        logoutBtn.addEventListener('click', (e) => {
            e.preventDefault();
            localStorage.removeItem('jwt_token');
            window.location.href = '/';
        });
    }
};

// Evrensel Tablo Başlığı Sıralama Sistemi (Universal Table Header Sorting)
// Tüm sayfa, modül ve sekmelerdeki tablolarda başlığa tıklayarak (A-Z, Z-A, sayısal, tarih) sıralama yapılmasını sağlar.
window.initUniversalTableSorting = function () {
    function applySortingToTable(table) {
        if (table.dataset.sortableInit === 'true') return;

        const thead = table.querySelector('thead');
        if (!thead) return;

        const headers = thead.querySelectorAll('th');
        if (!headers.length) return;

        table.dataset.sortableInit = 'true';

        headers.forEach((th, colIndex) => {
            const text = th.textContent.trim().toLowerCase();
            // "işlem", "işlemler", "aksiyon", "detay", "seç", "sil" gibi buton / aksiyon sütunlarını sıralama dışı tutalım
            if (text.includes('işlem') || text.includes('aksiyon') || text.includes('yönet') || text.includes('seç') || text.includes('durumunu değiştir')) {
                return;
            }

            th.classList.add('sortable-header');
            if (!th.hasAttribute('title')) {
                th.title = 'Sıralamak için tıklayın';
            }

            // Icon ekleyelim (eğer yoksa)
            let icon = th.querySelector('.sort-icon, i.fa-sort, i.fa-sort-up, i.fa-sort-down, i.fa-sort-alpha-down, i.fa-sort-numeric-down');
            if (!icon) {
                icon = document.createElement('i');
                icon.className = 'fas fa-sort sort-icon';
                th.appendChild(icon);
            }

            th.addEventListener('click', (e) => {
                // Eğer özel JS sıralama fonksiyonu tanımlıysa (örn: sortUsers), çakışmayı önleyelim
                if (th.hasAttribute('onclick') && th.getAttribute('onclick').includes('sort')) {
                    return;
                }

                const tbody = table.querySelector('tbody');
                if (!tbody) return;

                const rows = Array.from(tbody.querySelectorAll('tr'));
                if (rows.length <= 1) return;

                // Yükleniyor veya boş mesaj satırlarını ayıkla
                const validRows = rows.filter(row => {
                    const rowText = row.textContent.trim();
                    return !rowText.includes('Yükleniyor...') && 
                           !rowText.includes('Kayıt bulunamadı') && 
                           !rowText.includes('veri bulunamadı') &&
                           !rowText.includes('Henüz') &&
                           row.children.length > 1;
                });
                if (validRows.length <= 1) return;

                const currentDir = th.dataset.sortDir === 'asc' ? 'desc' : 'asc';

                // Diğer tüm header'ları sıfırla
                headers.forEach(h => {
                    if (h !== th) {
                        h.dataset.sortDir = '';
                        h.classList.remove('sorted-asc', 'sorted-desc');
                        const hIcon = h.querySelector('.sort-icon');
                        if (hIcon) {
                            hIcon.className = 'fas fa-sort sort-icon';
                        }
                    }
                });

                th.dataset.sortDir = currentDir;
                th.classList.remove('sorted-asc', 'sorted-desc');
                th.classList.add(currentDir === 'asc' ? 'sorted-asc' : 'sorted-desc');

                if (icon) {
                    icon.className = currentDir === 'asc' 
                        ? 'fas fa-sort-up sort-icon' 
                        : 'fas fa-sort-down sort-icon';
                }
                // Satırları sırala
                validRows.sort((rowA, rowB) => {
                    const cellA = rowA.children[colIndex] ? rowA.children[colIndex].textContent.trim() : '';
                    const cellB = rowB.children[colIndex] ? rowB.children[colIndex].textContent.trim() : '';

                    return compareCells(cellA, cellB, currentDir);
                });

                // Sıralanmış satırları DOM'a tekrar ekle
                validRows.forEach(row => tbody.appendChild(row));
            });
        });
    }

    function compareCells(valA, valB, direction) {
        const mult = direction === 'asc' ? 1 : -1;

        // 1. TARİH KONTROLÜ (Öncelikle tarih formatı var mı bakılır - GG.AA.YYYY, YYYY-AA-GG vb.)
        const dateA = parseTurkishDate(valA);
        const dateB = parseTurkishDate(valB);
        if (dateA !== null && dateB !== null) {
            return (dateA - dateB) * mult;
        }

        // 2. SAYISAL TEMİZLEME VE KONTROL (#1, ₺50.000, $100, %18 vb.)
        // Sadece tarih olmayan ve sayı simgesi içeren hücreler için sayısal kontrol yap
        const isNumCandidateA = /^[#№₺$%\s]*[+-]?\d+([.,]\d+)?[#№₺$%\s]*$/.test(valA.trim());
        const isNumCandidateB = /^[#№₺$%\s]*[+-]?\d+([.,]\d+)?[#№₺$%\s]*$/.test(valB.trim());

        if (isNumCandidateA && isNumCandidateB) {
            const cleanA = valA.replace(/^[#№]\s*/, '').replace(/[₺$\s%]/g, '').replace(/\.(?=\d{3})/g, '').replace(',', '.');
            const cleanB = valB.replace(/^[#№]\s*/, '').replace(/[₺$\s%]/g, '').replace(/\.(?=\d{3})/g, '').replace(',', '.');

            const numA = parseFloat(cleanA);
            const numB = parseFloat(cleanB);

            if (!isNaN(numA) && !isNaN(numB)) {
                return (numA - numB) * mult;
            }
        }

        // 3. TÜRKÇE ALFABETİK KARŞILAŞTIRMA (A-Z / Z-A)
        return valA.localeCompare(valB, 'tr', { sensitivity: 'base', numeric: true }) * mult;
    }

    function parseTurkishDate(str) {
        if (!str || typeof str !== 'string') return null;
        const cleanStr = str.trim();
        if (!cleanStr) return null;

        // 1. GG.AA.YYYY HH:mm:ss veya GG.AA.YYYY HH:mm veya GG.AA.YYYY (ör: 04.07.2026, 14.07.2026, 11.08.2026)
        const matchDot = cleanStr.match(/(\d{1,2})\.(\d{1,2})\.(\d{4})(?:\s+(\d{1,2}):(\d{1,2})(?::(\d{1,2}))?)?/);
        if (matchDot) {
            const day = parseInt(matchDot[1], 10);
            const month = parseInt(matchDot[2], 10) - 1; // 0-indexed ay (Ocak = 0)
            const year = parseInt(matchDot[3], 10);
            const hour = matchDot[4] ? parseInt(matchDot[4], 10) : 0;
            const min = matchDot[5] ? parseInt(matchDot[5], 10) : 0;
            const sec = matchDot[6] ? parseInt(matchDot[6], 10) : 0;
            const d = new Date(year, month, day, hour, min, sec);
            if (!isNaN(d.getTime())) return d.getTime();
        }

        // 2. YYYY-AA-GG HH:mm:ss veya YYYY-AA-GG (ISO formatı)
        const matchDash = cleanStr.match(/(\d{4})-(\d{1,2})-(\d{1,2})(?:[T\s](\d{1,2}):(\d{1,2})(?::(\d{1,2}))?)?/);
        if (matchDash) {
            const year = parseInt(matchDash[1], 10);
            const month = parseInt(matchDash[2], 10) - 1;
            const day = parseInt(matchDash[3], 10);
            const hour = matchDash[4] ? parseInt(matchDash[4], 10) : 0;
            const min = matchDash[5] ? parseInt(matchDash[5], 10) : 0;
            const sec = matchDash[6] ? parseInt(matchDash[6], 10) : 0;
            const d = new Date(year, month, day, hour, min, sec);
            if (!isNaN(d.getTime())) return d.getTime();
        }

        // 3. GG/AA/YYYY
        const matchSlash = cleanStr.match(/(\d{1,2})\/(\d{1,2})\/(\d{4})/);
        if (matchSlash) {
            const day = parseInt(matchSlash[1], 10);
            const month = parseInt(matchSlash[2], 10) - 1;
            const year = parseInt(matchSlash[3], 10);
            const d = new Date(year, month, day);
            if (!isNaN(d.getTime())) return d.getTime();
        }

        // 4. Türkçe Ay İsmi ("14 Temmuz 2026", "11 Ağustos 2026")
        const trMonths = {
            'ocak': 0, 'şubat': 1, 'mart': 2, 'nisan': 3, 'mayıs': 4, 'haziran': 5,
            'temmuz': 6, 'ağustos': 7, 'eylül': 8, 'ekim': 9, 'kasım': 10, 'aralık': 11
        };
        const matchText = cleanStr.toLowerCase().match(/(\d{1,2})\s+([a-zğüşıöç]+)\s+(\d{4})/);
        if (matchText) {
            const day = parseInt(matchText[1], 10);
            const monthName = matchText[2];
            const year = parseInt(matchText[3], 10);
            if (trMonths[monthName] !== undefined) {
                const d = new Date(year, trMonths[monthName], day);
                if (!isNaN(d.getTime())) return d.getTime();
            }
        }

        return null;
    }

    function scanAndApply() {
        document.querySelectorAll('table').forEach(applySortingToTable);
    }

    scanAndApply();

    if (!window._tableSortObserver) {
        window._tableSortObserver = new MutationObserver(() => {
            scanAndApply();
        });
        window._tableSortObserver.observe(document.body, { childList: true, subtree: true });
    }
};

// trNormalizeTableText: Türkçe karakterleri arama uyumluluğu için normalize eder
function trNormalizeTableText(str) {
    if (!str) return '';
    return str
        .toLocaleLowerCase('tr')
        .replace(/i̇/g, 'i')
        .replace(/ı/g, 'i')
        .replace(/ğ/g, 'g')
        .replace(/ü/g, 'u')
        .replace(/ş/g, 's')
        .replace(/ö/g, 'o')
        .replace(/ç/g, 'c');
}

// applyTableColumnFilters: Tablonun sütun arama kutularındaki girdilere göre tbody satırlarını canlı filtreler
window.applyTableColumnFilters = function (table) {
    if (!table) return;
    const filterRow = table.querySelector('.table-filter-row');
    if (!filterRow) return;

    const inputs = Array.from(filterRow.querySelectorAll('input.table-col-filter'));
    const activeFilters = inputs
        .filter(inp => inp.value.trim() !== '')
        .map(inp => {
            const rawVal = inp.value.trim();
            return {
                index: parseInt(inp.dataset.colIndex, 10),
                rawVal: rawVal.toLocaleLowerCase('tr'),
                normVal: trNormalizeTableText(rawVal)
            };
        });

    const tbody = table.querySelector('tbody');
    if (!tbody) return;

    const rows = Array.from(tbody.querySelectorAll('tr'));

    rows.forEach(row => {
        // Yükleniyor veya bildirim mesaj satırlarını atla
        const text = row.textContent.trim();
        if (
            text.includes('Yükleniyor...') ||
            text.includes('Gösterilecek') ||
            text.includes('Kayıt bulunamadı') ||
            text.includes('veri bulunamadı') ||
            text.includes('Henüz') ||
            row.classList.contains('no-filter-match-row')
        ) {
            return;
        }

        if (activeFilters.length === 0) {
            row.style.display = '';
            return;
        }

        let isMatch = true;
        for (const filter of activeFilters) {
            const cell = row.children[filter.index];
            if (!cell) {
                isMatch = false;
                break;
            }

            const cellText = cell.textContent.trim();
            const cellLower = cellText.toLocaleLowerCase('tr');
            const cellNorm = trNormalizeTableText(cellText);

            if (!cellLower.includes(filter.rawVal) && !cellNorm.includes(filter.normVal)) {
                isMatch = false;
                break;
            }
        }

        row.style.display = isMatch ? '' : 'none';
    });
};

// clearTableColumnFilters: Seçilen tablodaki tüm sütun arama kutularını temizler
window.clearTableColumnFilters = function (btnOrTable) {
    const table = btnOrTable.closest ? btnOrTable.closest('table') : btnOrTable;
    if (!table) return;
    const filterRow = table.querySelector('.table-filter-row');
    if (!filterRow) return;

    filterRow.querySelectorAll('input.table-col-filter').forEach(inp => {
        inp.value = '';
    });
    window.applyTableColumnFilters(table);
};

// initUniversalTableFiltering: Tüm tablolarda başlıkların altında arama kutuları oluşturur
window.initUniversalTableFiltering = function () {
    function applyFilteringToTable(table) {
        const thead = table.querySelector('thead');
        if (!thead) return;

        const firstHeaderRow = thead.querySelector('tr:not(.table-filter-row)');
        if (!firstHeaderRow) return;

        const headers = firstHeaderRow.querySelectorAll('th');
        if (!headers.length) return;

        let filterRow = thead.querySelector('.table-filter-row');
        if (!filterRow) {
            filterRow = document.createElement('tr');
            filterRow.className = 'table-filter-row';

            let hasSearchableCols = false;

            headers.forEach((th, colIndex) => {
                const filterTh = document.createElement('th');
                const rawText = th.textContent.trim();
                const cleanText = rawText.replace(/[\u2195\u25B2\u25BC\u2191\u2193]/g, '').trim();
                const lowerText = cleanText.toLowerCase();

                const isActionCol =
                    lowerText.includes('işlem') ||
                    lowerText.includes('aksiyon') ||
                    lowerText.includes('yönet') ||
                    lowerText.includes('seç') ||
                    lowerText.includes('durumunu değiştir') ||
                    colIndex === headers.length - 1;

                if (isActionCol) {
                    filterTh.style.textAlign = 'center';
                    filterTh.innerHTML = `
                        <button type="button" class="btn-clear-table-filters" title="Aramaları Temizle" onclick="window.clearTableColumnFilters(this)">
                            <i class="fas fa-times"></i> Temizle
                        </button>
                    `;
                } else {
                    hasSearchableCols = true;
                    const input = document.createElement('input');
                    input.type = 'text';
                    input.className = 'table-col-filter';
                    input.placeholder = `🔍 ${cleanText || 'Ara'}...`;
                    input.dataset.colIndex = colIndex;

                    input.addEventListener('click', (e) => e.stopPropagation());
                    input.addEventListener('keydown', (e) => {
                        e.stopPropagation();
                        if (e.key === 'Escape') {
                            input.value = '';
                            window.applyTableColumnFilters(table);
                        }
                    });
                    input.addEventListener('input', (e) => {
                        e.stopPropagation();
                        window.applyTableColumnFilters(table);
                    });

                    filterTh.appendChild(input);
                }
                filterRow.appendChild(filterTh);
            });

            if (hasSearchableCols) {
                thead.appendChild(filterRow);
            }
        } else {
            filterRow.querySelectorAll('input.table-col-filter').forEach(input => {
                if (!input.dataset.boundInit) {
                    input.dataset.boundInit = 'true';
                    input.addEventListener('click', (e) => e.stopPropagation());
                    input.addEventListener('keydown', (e) => {
                        e.stopPropagation();
                        if (e.key === 'Escape') {
                            input.value = '';
                            window.applyTableColumnFilters(table);
                        }
                    });
                    input.addEventListener('input', (e) => {
                        e.stopPropagation();
                        window.applyTableColumnFilters(table);
                    });
                }
            });
        }

        window.applyTableColumnFilters(table);
    }

    function scanAndApplyFiltering() {
        document.querySelectorAll('table').forEach(applyFilteringToTable);
    }

    scanAndApplyFiltering();

    if (!window._tableFilterObserver) {
        window._tableFilterObserver = new MutationObserver(() => {
            scanAndApplyFiltering();
        });
        window._tableFilterObserver.observe(document.body, { childList: true, subtree: true });
    }
};

// Sayfa yüklendiğinde otomatik olarak çalıştır
function loadGlobalModuleScripts() {
    if (!document.getElementById('feedbackModuleScript')) {
        const s = document.createElement('script');
        s.id = 'feedbackModuleScript';
        s.src = '/static/js/feedback.js';
        document.head.appendChild(s);
    }
    if (!document.getElementById('chatbotModuleScript')) {
        const s = document.createElement('script');
        s.id = 'chatbotModuleScript';
        s.src = '/static/js/chatbot.js';
        document.head.appendChild(s);
    }
}

if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
        window.initNotificationsSystem();
        window.initProfileDropdown();
        window.initUniversalTableSorting();
        window.initUniversalTableFiltering();
        loadGlobalModuleScripts();
    });
} else {
    window.initNotificationsSystem();
    window.initProfileDropdown();
    window.initUniversalTableSorting();
    window.initUniversalTableFiltering();
    loadGlobalModuleScripts();
}


