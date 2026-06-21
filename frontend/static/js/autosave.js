/**
 * autosave.js
 * BAP Başvuru Formu - Otomatik Kaydetme Modülü
 * 
 * Bu modül, formda yapılan değişiklikleri belirli aralıklarla
 * localStorage'a kaydeder. Sayfa yenilendiğinde verileri geri yükler.
 */

(function () {
    'use strict';

    // Sabitler
    const AUTOSAVE_KEY = 'bap_application_draft';
    const AUTOSAVE_INTERVAL = 15000; // 15 saniyede bir otomatik kayıt
    const FORM_ID = 'applicationForm';

    // DOM elemanları
    const form = document.getElementById(FORM_ID);
    const statusDot = document.querySelector('#autosaveStatus .dot');
    const statusText = document.getElementById('autosaveText');

    // Form yoksa modülü devre dışı bırak
    if (!form) return;

    // showSavingState - Kaydediliyor durumunu gösterir
    function showSavingState() {
        if (statusDot) statusDot.classList.add('saving');
        if (statusText) statusText.textContent = 'Kaydediliyor...';
    }

    // showSavedState - Kaydedildi durumunu gösterir
    function showSavedState() {
        if (statusDot) statusDot.classList.remove('saving');

        const now = new Date();
        const timeStr = now.toLocaleTimeString('tr-TR', {
            hour: '2-digit',
            minute: '2-digit'
        });
        if (statusText) statusText.textContent = 'Kaydedildi - ' + timeStr;
    }

    // collectFormData - Formdaki tüm input değerlerini toplar
    function collectFormData() {
        const data = {};
        const inputs = form.querySelectorAll('input, select, textarea');

        inputs.forEach(function (input) {
            if (!input.name) return;

            if (input.type === 'checkbox') {
                data[input.name] = input.checked;
            } else {
                data[input.name] = input.value;
            }
        });

        return data;
    }

    // saveToLocalStorage - Verileri localStorage'a kaydeder
    function saveToLocalStorage() {
        showSavingState();

        try {
            const data = collectFormData();
            data._savedAt = new Date().toISOString();
            data._projeId = window.editProjeId || null; // Projenin ID'sini de taslağa kaydediyoruz
            localStorage.setItem(AUTOSAVE_KEY, JSON.stringify(data));
        } catch (e) {
            console.warn('Otomatik kayıt sırasında hata oluştu:', e);
        }

        // Kısa bir gecikme ile "kaydedildi" göster (UX)
        setTimeout(showSavedState, 500);
    }

    // loadFromLocalStorage - Kaydedilmiş verileri geri yükler
    function loadFromLocalStorage() {
        try {
            const raw = localStorage.getItem(AUTOSAVE_KEY);
            if (!raw) return;

            const data = JSON.parse(raw);
            if (!data) return;

            // Proje ID kontrolü: Taslağın şu anki sayfa (düzenleme veya yeni proje) ile uyumlu olup olmadığını kontrol et
            const currentEditId = window.editProjeId || null;
            const draftEditId = data._projeId || null;
            if (currentEditId !== draftEditId) {
                console.log('Taslak proje ID eşleşmediği için geri yüklenmedi.');
                return;
            }

            const inputs = form.querySelectorAll('input, select, textarea');
            inputs.forEach(function (input) {
                if (!input.name || !(input.name in data)) return;

                if (input.type === 'checkbox') {
                    input.checked = data[input.name];
                } else {
                    input.value = data[input.name];
                    if (input.id === 'bapType') {
                        // Dropdown seçenekleri henüz yüklenmediği için değeri globalde saklıyoruz
                        window.draftBapTuruId = data[input.name];
                    }
                }
            });

            // Bütçe toplamını güncelle (eğer fonksiyon mevcutsa)
            if (typeof calculateBudget === 'function') {
                calculateBudget();
            }

            console.log('Taslak yüklendi:', data._savedAt);
        } catch (e) {
            console.warn('Otomatik yükleme sırasında hata oluştu:', e);
        }
    }

    // clearDraft - Taslağı temizler
    function clearDraft() {
        localStorage.removeItem(AUTOSAVE_KEY);
    }

    // Zamanlayıcı ile düzenli kayıt
    setInterval(saveToLocalStorage, AUTOSAVE_INTERVAL);

    // Sayfa yüklendiğinde taslağı geri yükle
    loadFromLocalStorage();

    // Form gönderildiğinde taslağı sil
    form.addEventListener('submit', function () {
        clearDraft();
    });

    // Input değişikliklerinde kısa bir gecikmeyle kaydet (debounce)
    let debounceTimer = null;
    form.addEventListener('input', function () {
        clearTimeout(debounceTimer);
        debounceTimer = setTimeout(saveToLocalStorage, 2000);
    });

})();
