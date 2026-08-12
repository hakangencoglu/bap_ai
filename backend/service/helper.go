package service

import (
	"errors"
	"regexp"
)

// cleanAndValidatePhone telefon numarasını temizler (sadece rakam bırakır), başında 0 varsa kaldırır,
// 10 haneli olup olmadığını ve 5XX ile başlayan geçerli bir cep telefonu numarası olduğunu kontrol eder.
func cleanAndValidatePhone(phone string) (string, error) {
	if phone == "" {
		return "", nil
	}

	// Sadece rakamları ayıkla
	reg := regexp.MustCompile(`[^0-9]`)
	cleaned := reg.ReplaceAllString(phone, "")

	// Başında 0 varsa kaldır
	if len(cleaned) > 0 && cleaned[0] == '0' {
		cleaned = cleaned[1:]
	}

	// 10 hane kontrolü
	if len(cleaned) != 10 {
		return "", errors.New("lütfen telefon numaranızı başında 0 olmadan tam olarak 10 haneli olarak giriniz (Örn: 5321234567)")
	}

	// Alan kodu ve cep telefonu kontrolü (Türkiye cep telefonları 5 ile başlamalıdır)
	if cleaned[0] != '5' {
		return "", errors.New("lütfen geçerli bir cep telefonu numarası giriniz (numaranız 5XX ile başlamalıdır, sabit hatlar kabul edilmemektedir)")
	}

	return cleaned, nil
}
