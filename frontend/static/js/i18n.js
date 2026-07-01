// frontend/static/js/i18n.js

document.addEventListener('DOMContentLoaded', () => {
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
        langToggleBtn.title = currentLang === 'tr' ? 'İngilizceye Geç' : 'Switch to Turkish';
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
        langToggleBtn.innerHTML = currentLang === 'tr' ? 'TR' : 'EN';
        
        // Değişim olayını dinle
        langToggleBtn.addEventListener('click', (e) => {
            e.preventDefault();
            e.stopPropagation();
            currentLang = currentLang === 'tr' ? 'en' : 'tr';
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
                                <li class="menu-item ${path === '/eimza' ? 'active' : ''}">
                                    <a href="/eimza">
                                        <i class="fas fa-signature"></i>
                                        <span>${window.t('nav.eimza')}</span>
                                    </a>
                                </li>
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
                                <li class="menu-item ${path === '/eimza' ? 'active' : ''}">
                                    <a href="/eimza">
                                        <i class="fas fa-signature"></i>
                                        <span>${window.t('nav.eimza')}</span>
                                    </a>
                                </li>
                            </ul>
                        `;
                    }

                    // 4. TTO Menüsü
                    if (roles.includes('tto')) {
                        menuHTML += `
                            <div class="menu-label">${window.t('nav.tto_menu')}</div>
                            <ul class="menu-list">
                                <li class="menu-item ${path === '/tto/dashboard' ? 'active' : ''}">
                                    <a href="/tto/dashboard">
                                        <i class="fas fa-rocket"></i>
                                        <span>${window.t('nav.tto_panel')}</span>
                                    </a>
                                </li>
                                <li class="menu-item ${path === '/eimza' ? 'active' : ''}">
                                    <a href="/eimza">
                                        <i class="fas fa-signature"></i>
                                        <span>${window.t('nav.eimza')}</span>
                                    </a>
                                </li>
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
                        const showEimzaInAnaMenu = !(roles.includes('dekan') || roles.includes('komisyon') || roles.includes('tto'));
                        menuHTML += `
                            <div class="menu-label">${window.t('nav.main_menu')}</div>
                            <ul class="menu-list">
                                <li class="menu-item ${path === '/anasayfa' ? 'active' : ''}">
                                    <a href="/anasayfa">
                                        <i class="fas fa-home"></i>
                                        <span>${window.t('nav.dashboard')}</span>
                                    </a>
                                </li>
                                <li class="menu-item ${path === '/basvuru' ? 'active' : ''}">
                                    <a href="/basvuru">
                                        <i class="fas fa-file-signature"></i>
                                        <span>${window.t('nav.new_application')}</span>
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

