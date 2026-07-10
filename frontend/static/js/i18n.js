// frontend/static/js/i18n.js

document.addEventListener('DOMContentLoaded', async () => {
    // 1. Dil Seçeneğini Yükle (Varsayılan: tr)
    let currentLang = localStorage.getItem('bap_lang') || 'tr';
    
    // Uygulama genelinde kullanılacak çeviri fonksiyonunu tanımla
    window.t = function(key, params = {}) {
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

    // 4. Dinamik Sidebar Menüsü Oluşturma
    const sidebarMenu = document.querySelector('.sidebar-menu');
    if (sidebarMenu) {
        const token = localStorage.getItem('jwt_token');
        if (token) {
            let allowedPages = [];
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

            try {
                const base64Url = token.split('.')[1];
                const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
                const jsonPayload = decodeURIComponent(atob(base64).split('').map(function(c) {
                    return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
                }).join(''));
                const claims = JSON.parse(jsonPayload);
                const role = claims.role;
                
                if (role) {
                    const roles = role.split(',').map(r => r.trim());
                    let menuHTML = '';
                    const path = window.location.pathname;
                    const search = window.location.search;

                    // 1. Admin Menüsü
                    if (roles.includes('admin')) {
                        menuHTML += `
                            <div class="menu-label">${window.t('nav.admin_menu')}</div>
                            <ul class="menu-list">
                                <li class="menu-item ${path.startsWith('/admin/dashboard') && !search.includes('tab') ? 'active' : ''}">
                                    <a href="/admin/dashboard">
                                        <i class="fas fa-shield-alt"></i>
                                        <span>${window.t('nav.admin_panel')}</span>
                                    </a>
                                </li>
                                <li class="menu-item ${path === '/admin/hakem-atama' ? 'active' : ''}">
                                    <a href="/admin/hakem-atama">
                                        <i class="fas fa-user-check"></i>
                                        <span>${window.t('nav.referee_assign')}</span>
                                    </a>
                                </li>
                                <li class="menu-item ${search.includes('tab=usersTab') ? 'active' : ''}">
                                    <a href="/admin/dashboard?tab=usersTab">
                                        <i class="fas fa-users-cog"></i>
                                        <span>${window.t('nav.user_management')}</span>
                                    </a>
                                </li>
                                <li class="menu-item ${search.includes('tab=bapTab') ? 'active' : ''}">
                                    <a href="/admin/dashboard?tab=bapTab">
                                        <i class="fas fa-folder-plus"></i>
                                        <span>${window.t('nav.bap_definition')}</span>
                                    </a>
                                </li>
                            </ul>
                            <div class="menu-label">${window.t('nav.reports')}</div>
                            <ul class="menu-list">
                                <li class="menu-item ${path === '/admin/projects/status' ? 'active' : ''}">
                                    <a href="/admin/projects/status">
                                        <i class="fas fa-chart-pie"></i>
                                        <span>${window.t('nav.status_reports')}</span>
                                    </a>
                                </li>
                            </ul>
                        `;
                    }

                    // 2. Dekan Menüsü
                    if (roles.includes('dekan')) {
                        menuHTML += `
                            <div class="menu-label">${window.t('nav.dekan_menu')}</div>
                            <ul class="menu-list">
                                <li class="menu-item ${path === '/dekan/dashboard' ? 'active' : ''}">
                                    <a href="/dekan/dashboard">
                                        <i class="fas fa-university"></i>
                                        <span>${window.t('nav.dekan_panel')}</span>
                                    </a>
                                </li>
                                ${allowedPages.includes('/eimza') ? `
                                <li class="menu-item ${path === '/eimza' ? 'active' : ''}">
                                    <a href="/eimza">
                                        <i class="fas fa-signature"></i>
                                        <span>${window.t('nav.eimza')}</span>
                                    </a>
                                </li>
                                ` : ''}
                            </ul>
                        `;
                    }

                    // 3. Komisyon Menüsü
                    if (roles.includes('komisyon')) {
                        menuHTML += `
                            <div class="menu-label">${window.t('nav.komisyon_menu')}</div>
                            <ul class="menu-list">
                                <li class="menu-item ${path === '/komisyon/dashboard' ? 'active' : ''}">
                                    <a href="/komisyon/dashboard">
                                        <i class="fas fa-gavel"></i>
                                        <span>${window.t('nav.komisyon_panel')}</span>
                                    </a>
                                </li>
                                ${allowedPages.includes('/eimza') ? `
                                <li class="menu-item ${path === '/eimza' ? 'active' : ''}">
                                    <a href="/eimza">
                                        <i class="fas fa-signature"></i>
                                        <span>${window.t('nav.eimza')}</span>
                                    </a>
                                </li>
                                ` : ''}
                            </ul>
                        `;
                    }

                    // 4. TTO Menüsü
                    if (roles.includes('tto')) {
                        const isDashboardActive = path === '/tto/dashboard' && !search.includes('section=satinalma');
                        const isSatinalmaActive = path === '/tto/satinalma' || search.includes('section=satinalma');
                        menuHTML += `
                            <div class="menu-label">${window.t('nav.tto_menu')}</div>
                            <ul class="menu-list">
                                <li class="menu-item ${isDashboardActive ? 'active' : ''}">
                                    <a href="/tto/dashboard">
                                        <i class="fas fa-rocket"></i>
                                        <span>${window.t('nav.tto_panel')}</span>
                                    </a>
                                </li>
                                ${allowedPages.includes('/tto/satinalma') ? `
                                <li class="menu-item ${isSatinalmaActive ? 'active' : ''}">
                                    <a href="/tto/satinalma">
                                        <i class="fas fa-shopping-cart"></i>
                                        <span>${window.t('nav.purchasing_management_tto')}</span>
                                    </a>
                                </li>
                                ` : ''}
                                ${allowedPages.includes('/eimza') ? `
                                <li class="menu-item ${path === '/eimza' ? 'active' : ''}">
                                    <a href="/eimza">
                                        <i class="fas fa-signature"></i>
                                        <span>${window.t('nav.eimza')}</span>
                                    </a>
                                </li>
                                ` : ''}
                            </ul>
                        `;
                    }

                    // 5. Hakem Menüsü
                    if (roles.includes('hakem')) {
                        menuHTML += `
                            <div class="menu-label">${window.t('nav.referee_menu')}</div>
                            <ul class="menu-list">
                                <li class="menu-item ${path === '/hakem/dashboard' || path.startsWith('/hakem/degerlendirme') ? 'active' : ''}">
                                    <a href="/hakem/dashboard">
                                        <i class="fas fa-gavel"></i>
                                        <span>${window.t('nav.referee_panel')}</span>
                                    </a>
                                </li>
                            </ul>
                        `;
                    }

                    // 6. Akademisyen / Öğrenci Menüsü (Varsayılan olarak bu iki rolden biri varsa veya hiçbiri özel değilse gösterelim)
                    const hasOtherRoles = roles.includes('admin') || roles.includes('dekan') || roles.includes('komisyon') || roles.includes('tto') || roles.includes('hakem');
                    const hasAcademicOrStudent = roles.includes('akademisyen') || roles.includes('ogrenci');

                    if (hasAcademicOrStudent || !hasOtherRoles) {
                        const showEimzaInAnaMenu = allowedPages.includes('/eimza') && !(roles.includes('dekan') || roles.includes('komisyon') || roles.includes('tto'));
                        menuHTML += `
                            <div class="menu-label">${window.t('nav.main_menu')}</div>
                            <ul class="menu-list">
                                <li class="menu-item ${path === '/anasayfa' ? 'active' : ''}">
                                    <a href="/anasayfa">
                                        <i class="fas fa-home"></i>
                                        <span>${window.t('nav.dashboard')}</span>
                                    </a>
                                </li>
                                ${showEimzaInAnaMenu ? `
                                <li class="menu-item ${path === '/eimza' ? 'active' : ''}">
                                    <a href="/eimza">
                                        <i class="fas fa-signature"></i>
                                        <span>${window.t('nav.eimza')}</span>
                                    </a>
                                </li>
                                ` : ''}
                            </ul>
                            
                            <div class="menu-label">${window.t('dash.quick_actions')}</div>
                            <ul class="menu-list">
                                <li class="menu-item ${path === '/basvuru' ? 'active' : ''}">
                                    <a href="/basvuru">
                                        <i class="fas fa-plus"></i>
                                        <span>${window.t('dash.quick_new_bap')}</span>
                                    </a>
                                </li>
                                <li class="menu-item">
                                    <a href="#" onclick="alert('BAP Başvuru Kılavuzu İndiriliyor...'); return false;">
                                        <i class="fas fa-file-pdf"></i>
                                        <span>${window.t('dash.quick_guide')}</span>
                                    </a>
                                </li>
                            </ul>

                            <div class="menu-label">${window.t('nav.management')}</div>
                            <ul class="menu-list">
                                <li class="menu-item">
                                    <a href="#">
                                        <i class="fas fa-chart-line"></i>
                                        <span>${window.t('nav.reports')}</span>
                                    </a>
                                </li>
                            </ul>
                        `;
                    }

                    sidebarMenu.innerHTML = menuHTML;
                }
            } catch (e) {
                console.error('Sidebar build error:', e);
            }
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
window.formatPhoneNumber = function(value) {
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
window.applyPhoneMaskToInput = function(input) {
    if (!input) return;
    
    input.setAttribute('maxlength', '15');
    input.setAttribute('placeholder', '(5XX) XXX XX XX');
    
    input.addEventListener('input', function(e) {
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
window.applySidebarPermissions = async function() {
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
window.initNotificationsSystem = async function() {
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
        } catch (err) {}
    }

    checkUnreadCountSilent();
    // Her 60 saniyede bir yeni bildirimleri sessizce kontrol et
    setInterval(checkUnreadCountSilent, 60000);
};

// Sayfa yüklendiğinde otomatik olarak çalıştır
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', window.initNotificationsSystem);
} else {
    window.initNotificationsSystem();
}

