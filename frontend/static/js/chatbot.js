// frontend/static/js/chatbot.js

(function() {
    // Türkçe Yorum: Giriş sayfası, kayıt sayfası veya ana kök dizinde asistanın çalışmasını engelliyoruz
    const pathname = window.location.pathname;
    if (pathname === '/' || pathname === '/login' || pathname === '/register') {
        return;
    }

    // Türkçe Yorum: Sadece giriş yapmış kullanıcılar için chatbot'u başlatıyoruz.
    const token = localStorage.getItem('jwt_token');
    if (!token) return;

    // DOM öğelerinin çakışmaması için sayfa tamamen yüklendiğinde çalıştır
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', checkPermissionAndInit);
    } else {
        checkPermissionAndInit();
    }

    async function checkPermissionAndInit() {
        try {
            // Türkçe Yorum: Kullanıcının chatbot yetkisi olup olmadığını sorguluyoruz.
            const checkRes = await fetch('/api/auth/check-page-access?path=' + encodeURIComponent('/api/chat'), {
                headers: { 'Authorization': 'Bearer ' + token }
            });
            if (checkRes.ok) {
                const checkData = await checkRes.json();
                if (checkData.allowed) {
                    initChatbot();
                }
            }
        } catch (e) {
            console.error('Chatbot yetki kontrol hatası:', e);
        }
    }

    function initChatbot() {
        // Zaten eklenmişse tekrar ekleme
        if (document.getElementById('chatbotToggleBtn')) return;

        // 1. DOM Öğelerini Dinamik Olarak Oluştur ve Gövdeye Ekle
        const body = document.body;

        // Yüzen Aç/Kapat Butonu
        const toggleBtn = document.createElement('button');
        toggleBtn.id = 'chatbotToggleBtn';
        toggleBtn.className = 'chatbot-toggle-btn';
        toggleBtn.title = 'BAP AI Asistanı';
        toggleBtn.innerHTML = '<i class="fas fa-robot"></i><span class="chatbot-status-dot" id="chatbotStatusDot"></span>';
        body.appendChild(toggleBtn);

        // Sohbet Paneli Konteyneri
        const chatContainer = document.createElement('div');
        chatContainer.id = 'chatbotContainer';
        chatContainer.className = 'chatbot-container';
        
        chatContainer.innerHTML = `
            <div class="chatbot-header">
                <div class="chatbot-header-title">
                    <span>BAP AI Asistanı </span>
                    <i class="fas fa-robot" id="chatbotHeaderRobot" style="transition: color 0.3s; margin-left: 0.25rem;"></i>
                </div>
                <div class="chatbot-header-actions">
                    <button class="chatbot-header-btn" id="chatbotNewBtn" title="Yeni Sohbet"><i class="fas fa-plus"></i></button>
                    <button class="chatbot-header-btn" id="chatbotHistoryBtn" title="Geçmiş"><i class="fas fa-history"></i></button>
                    <button class="chatbot-header-btn" id="chatbotClearBtn" title="Sohbeti Temizle">
                        <i class="fas fa-trash-alt"></i>
                    </button>
                    <button class="chatbot-header-btn" id="chatbotCloseBtn" title="Kapat">
                        <i class="fas fa-times"></i>
                    </button>
                </div>
            </div>
            <div class="chatbot-messages" id="chatbotMessages"></div>
            <div class="chatbot-input-container">
                <textarea class="chatbot-input-field" id="chatbotInputField" placeholder="yapay zeka" rows="1"></textarea>
                <div class="chatbot-input-controls">
                    <div class="chatbot-input-actions">
                        <button class="chatbot-action-btn" title="Dosya Ekle"><i class="fas fa-plus"></i></button>
                        <div class="chatbot-model-selector" title="Model Bilgisi">
                            <i class="fas fa-robot"></i>
                            <span>BAP AI Asistanı</span>
                            <i class="fas fa-chevron-up" style="font-size: 0.6rem;"></i>
                        </div>
                        <button class="chatbot-action-btn" title="Sesle Yaz"><i class="fas fa-microphone"></i></button>
                    </div>
                    <button class="chatbot-send-btn" id="chatbotSendBtn" disabled>
                        <i class="fas fa-arrow-right"></i>
                    </button>
                </div>
            </div>
        `;
        body.appendChild(chatContainer);

        // 2. DOM Referanslarını Al
        const messagesContainer = document.getElementById('chatbotMessages');
        const inputField = document.getElementById('chatbotInputField');
        const sendBtn = document.getElementById('chatbotSendBtn');
        const closeBtn = document.getElementById('chatbotCloseBtn');
        const clearBtn = document.getElementById('chatbotClearBtn');
        const newBtn = document.getElementById('chatbotNewBtn');
        const historyBtn = document.getElementById('chatbotHistoryBtn');

        // 3. Olay Dinleyicileri (Event Listeners)
        newBtn.addEventListener('click', () => {
            if (confirm('Yeni bir sohbet başlatmak istiyor musunuz? Geçmiş temizlenecektir.')) {
                clearChatHistory();
            }
        });

        historyBtn.addEventListener('click', () => {
            alert('Sohbet geçmişiniz tarayıcınızda otomatik olarak saklanmaktadır.');
        });

        toggleBtn.addEventListener('click', () => {
            chatContainer.classList.toggle('active');
            if (chatContainer.classList.contains('active')) {
                document.body.classList.add('chatbot-open');
                inputField.focus();
                scrollToBottom();
            } else {
                document.body.classList.remove('chatbot-open');
            }
        });

        closeBtn.addEventListener('click', () => {
            chatContainer.classList.remove('active');
            document.body.classList.remove('chatbot-open');
        });

        clearBtn.addEventListener('click', () => {
            if (confirm('Sohbet geçmişini temizlemek istediğinizden emin misiniz?')) {
                clearChatHistory();
            }
        });

        inputField.addEventListener('input', () => {
            // Textarea yüksekliğini dinamik ayarla
            inputField.style.height = 'auto';
            inputField.style.height = (inputField.scrollHeight) + 'px';
            
            // Gönder butonunu aktif/pasif yap
            sendBtn.disabled = inputField.value.trim() === '';
        });

        inputField.addEventListener('keydown', (e) => {
            // Enter tuşuna basıldığında (Shift olmadan) mesajı gönder
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                sendMessage();
            }
        });

        sendBtn.addEventListener('click', sendMessage);

        // 4. Sohbet Geçmişini Yükle
        loadChatHistory();

        // 5. Yapay Zeka Sunucu Bağlantısını Kontrol Et (Yeşil/Kırmızı Durum Noktası)
        updateStatusIndicator();

        // --- Yardımcı Fonksiyonlar ---

        // Mesaj gönderme mantığı
        async function sendMessage() {
            const text = inputField.value.trim();
            if (!text) return;

            // Giriş alanını temizle ve sıfırla
            inputField.value = '';
            inputField.style.height = '40px';
            sendBtn.disabled = true;

            // Mevcut geçmişi oku (bu mesaj eklenmeden önce)
            const recentHistory = [];
            const rawHistory = localStorage.getItem('bap_chat_history');
            if (rawHistory) {
                try {
                    const savedHistory = JSON.parse(rawHistory);
                    // En fazla son 10 mesajı geçmiş olarak gönder
                    const startIdx = Math.max(0, savedHistory.length - 10);
                    for (let i = startIdx; i < savedHistory.length; i++) {
                        const h = savedHistory[i];
                        recentHistory.push({
                            role: h.sender === 'user' ? 'user' : 'assistant',
                            content: h.textContent || h.htmlContent.replace(/<[^>]*>/g, '').trim()
                        });
                    }
                } catch (e) {
                    console.error("Sohbet geçmişi işlenirken hata:", e);
                }
            }

            // Kullanıcı mesajını ekrana ekle ve kaydet
            appendMessage('user', text);
            saveChatHistory();

            // Yazıyor animasyonunu göster
            const typingIndicator = showTypingIndicator();

            try {
                // Backend API'sine istek at
                const response = await fetch('/api/chat', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'Authorization': 'Bearer ' + token
                    },
                    body: JSON.stringify({ 
                        message: text,
                        history: recentHistory
                    })
                });

                // Yazıyor göstergesini kaldır
                typingIndicator.remove();

                if (response.ok) {
                    const data = await response.json();
                    appendMessage('bot', data.response);
                } else {
                    const errData = await response.json();
                    appendMessage('bot', 'Üzgünüm, isteğinizi işlerken bir hata oluştu: ' + (errData.error || 'Bilinmeyen hata'));
                }
            } catch (err) {
                typingIndicator.remove();
                console.error('ChatBot API hatası:', err);
                appendMessage('bot', 'Bağlantı hatası! Yerel LLM sunucusu veya internet erişimi kontrol edilmelidir.');
            }

            saveChatHistory();
            scrollToBottom();
        }

        // Mesaj balonunu arayüze ekler
        function appendMessage(sender, text) {
            const time = new Date().toLocaleTimeString('tr-TR', { hour: '2-digit', minute: '2-digit' });
            const msgRow = document.createElement('div');
            msgRow.className = `chat-message-row ${sender}`;

            const bubble = document.createElement('div');
            bubble.className = 'chat-bubble';
            
            // Markdown formatını parse edip HTML olarak bas
            bubble.innerHTML = parseMarkdown(text) + `<span class="chat-bubble-time">${time}</span>`;
            
            msgRow.appendChild(bubble);
            messagesContainer.appendChild(msgRow);
            scrollToBottom();
        }

        // Yazıyor (Düşünüyor) animasyonu gösterir
        function showTypingIndicator() {
            const msgRow = document.createElement('div');
            msgRow.className = 'chat-message-row bot';

            const bubble = document.createElement('div');
            bubble.className = 'chat-bubble';
            bubble.innerHTML = `
                <div class="typing-indicator">
                    <div class="typing-dot"></div>
                    <div class="typing-dot"></div>
                    <div class="typing-dot"></div>
                </div>
            `;
            msgRow.appendChild(bubble);
            messagesContainer.appendChild(msgRow);
            scrollToBottom();
            return msgRow;
        }

        // Arayüzü en aşağı kaydırır
        function scrollToBottom() {
            messagesContainer.scrollTop = messagesContainer.scrollHeight;
        }

        // Sohbet geçmişini localStorage'a kaydeder
        function saveChatHistory() {
            const messages = [];
            const rows = messagesContainer.querySelectorAll('.chat-message-row');
            rows.forEach(row => {
                // Yazıyor animasyonunu kaydetme
                if (row.querySelector('.typing-indicator')) return;
                
                const isUser = row.classList.contains('user');
                const bubble = row.querySelector('.chat-bubble');
                const timeSpan = row.querySelector('.chat-bubble-time');
                
                // Zaman damgası ve dışındaki saf metni ayır
                let text = bubble.innerHTML;
                let textVal = bubble.innerText;
                if (timeSpan) {
                    text = text.replace(timeSpan.outerHTML, '');
                    textVal = textVal.replace(timeSpan.innerText, '');
                }

                messages.push({
                    sender: isUser ? 'user' : 'bot',
                    htmlContent: text,
                    textContent: textVal.trim()
                });
            });
            localStorage.setItem('bap_chat_history', JSON.stringify(messages));
        }

        // Sohbet geçmişini localStorage'dan yükler
        function loadChatHistory() {
            const raw = localStorage.getItem('bap_chat_history');
            if (raw) {
                try {
                    const messages = JSON.parse(raw);
                    messagesContainer.innerHTML = '';
                    messages.forEach(m => {
                        const msgRow = document.createElement('div');
                        msgRow.className = `chat-message-row ${m.sender}`;
                        const bubble = document.createElement('div');
                        bubble.className = 'chat-bubble';
                        
                        // Kayıtlı HTML içeriği bas
                        bubble.innerHTML = m.htmlContent + `<span class="chat-bubble-time">Geçmiş</span>`;
                        msgRow.appendChild(bubble);
                        messagesContainer.appendChild(msgRow);
                    });
                } catch (e) {
                    console.error('Geçmiş yükleme hatası:', e);
                }
            } else {
                // Hoş geldiniz mesajı
                appendMessage('bot', 'Merhaba! Ben **İZÜ BAP Yapay Zeka Asistanı** 🤖. \n\nİZÜ Bilimsel Araştırma Projeleri bütçe limitleri, kuralları, başvuru formu adımları ve satın alma işlemleri hakkında bilgi sahibiyim. Sorularınızı aşağıdaki alana yazarak bana iletebilirsiniz!');
            }
            scrollToBottom();
        }

        // Sohbet geçmişini sıfırlar
        function clearChatHistory() {
            localStorage.removeItem('bap_chat_history');
            messagesContainer.innerHTML = '';
            appendMessage('bot', 'Sohbet geçmişi başarıyla temizlendi. Nasıl yardımcı olabilirim?');
        }

        // Basit ve Premium Markdown Parser
        function parseMarkdown(text) {
            let html = text;

            // HTML etiket sızıntılarını önlemek için güvenli temizleme (escape)
            html = html.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");

            // Başlıklar
            html = html.replace(/^### (.*$)/gim, '<h4>$1</h4>');
            html = html.replace(/^## (.*$)/gim, '<h3>$1</h3>');
            html = html.replace(/^# (.*$)/gim, '<h2>$1</h2>');

            // Kalın yazılar
            html = html.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>');

            // Tabloları parse etme (Basic markdown table regex)
            const lines = html.split('\n');
            let inTable = false;
            let tableHTML = '';
            
            for (let i = 0; i < lines.length; i++) {
                const line = lines[i].trim();
                
                // Tablo satırı algılama
                if (line.startsWith('|') && line.endsWith('|')) {
                    if (!inTable) {
                        inTable = true;
                        tableHTML = '<table class="chatbot-table"><thead>';
                    }
                    
                    const cells = line.split('|').map(c => c.trim()).filter((c, idx, arr) => idx > 0 && idx < arr.length - 1);
                    
                    // Ayırıcı satır ise atla (e.g. | :--- | :--- |)
                    if (cells.every(c => c.startsWith(':') || c.startsWith('-') || c.endsWith('-'))) {
                        tableHTML = tableHTML.replace('<thead>', '<tbody>'); // Kapatıp tbody'ye geç
                        continue;
                    }
                    
                    tableHTML += '<tr>';
                    cells.forEach(cell => {
                        const tag = tableHTML.includes('<tbody>') ? 'td' : 'th';
                        tableHTML += `<${tag}>${cell}</${tag}>`;
                    });
                    tableHTML += '</tr>';
                    
                    if (tableHTML.includes('<tr>') && !tableHTML.includes('<tbody>') && !tableHTML.includes('</thead>')) {
                        tableHTML += '</thead>';
                    }
                    
                    lines[i] = ''; // Satırı boşalt, sonradan kaldıracağız
                } else {
                    if (inTable) {
                        inTable = false;
                        tableHTML += '</tbody></table>';
                        // Tabloyu bir önceki boşaltılan satıra enjekte et
                        let j = i - 1;
                        while (j >= 0 && lines[j] === '') {
                            j--;
                        }
                        lines[j + 1] = tableHTML;
                    }
                }
            }
            if (inTable) {
                tableHTML += '</tbody></table>';
                lines[lines.length - 1] = tableHTML;
            }
            
            html = lines.filter(l => l !== '').join('\n');

            // Liste elemanları (* veya - ile başlayanlar)
            html = html.replace(/^\s*[-*]\s+(.*$)/gim, '<li>$1</li>');
            // Ardışık <li> bloklarını <ul> içine al
            html = html.replace(/(<li>.*<\/li>)/gim, '<ul>$1</ul>');
            // Yan yana gelen </ul><ul> etiketlerini temizle
            html = html.replace(/<\/ul>\s*<ul>/g, '');

            // Satır sonları (\n)
            html = html.replace(/\n/g, '<br>');

            return html;
        }

        // Yapay zeka sağlayıcısının bağlantı durumunu sorgular ve durum göstergesini günceller
        async function updateStatusIndicator() {
            const statusDot = document.getElementById('chatbotStatusDot');
            if (!statusDot) return;
            try {
                const res = await fetch('/api/chat/status', {
                    headers: { 'Authorization': 'Bearer ' + token }
                });
                if (res.ok) {
                    const data = await res.json();
                    const headerRobot = document.getElementById('chatbotHeaderRobot');
                    if (data.connected) {
                        statusDot.classList.add('online');
                        statusDot.classList.remove('offline');
                        toggleBtn.classList.add('online');
                        toggleBtn.classList.remove('offline');
                        if (headerRobot) {
                            headerRobot.style.color = '#10b981';
                        }
                    } else {
                        statusDot.classList.add('offline');
                        statusDot.classList.remove('online');
                        toggleBtn.classList.add('offline');
                        toggleBtn.classList.remove('online');
                        if (headerRobot) {
                            headerRobot.style.color = '#ef4444';
                        }
                    }
                } else {
                    statusDot.classList.add('offline');
                    toggleBtn.classList.add('offline');
                    const headerRobot = document.getElementById('chatbotHeaderRobot');
                    if (headerRobot) headerRobot.style.color = '#ef4444';
                }
            } catch (e) {
                console.error('Bağlantı durum kontrol hatası:', e);
                statusDot.classList.add('offline');
                toggleBtn.classList.add('offline');
                const headerRobot = document.getElementById('chatbotHeaderRobot');
                if (headerRobot) headerRobot.style.color = '#ef4444';
            }
        }
    }
})();
