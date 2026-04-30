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
            
            // Eğer element input, textarea ise placeholder'ı veya value'yu güncelle
            if (el.tagName === 'INPUT' || el.tagName === 'TEXTAREA') {
                if (el.hasAttribute('placeholder')) {
                    el.placeholder = translation;
                } else if (el.type === 'button' || el.type === 'submit') {
                    el.value = translation;
                }
            } else {
                // Sadece metin içeriğini değiştir (içindeki HTML/icon bozulmasın diye innerText veya textNode takibi yapılabilir 
                // ancak bu basitlik adına child nodes içinde text node'u güncelleyeceğiz:
                let textNodeUpdated = false;
                for (let i = 0; i < el.childNodes.length; i++) {
                    if (el.childNodes[i].nodeType === 3 && el.childNodes[i].nodeValue.trim().length > 0) {
                        el.childNodes[i].nodeValue = translation;
                        textNodeUpdated = true;
                        break;
                    }
                }
                
                // Eğer text node bulunamadıysa (içi boşsa veya sadece element varsa), başa metin olarak ekle veya innerHTML kullanmadan yap
                if (!textNodeUpdated) {
                    if (el.children.length === 0) {
                        el.textContent = translation;
                    } else {
                        // Eğer içinde i etiketi var ama text yoksa, arkasına text ekle
                        el.appendChild(document.createTextNode(" " + translation));
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
});
