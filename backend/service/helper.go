package service

import (
	"errors"
	"regexp"
)

// cleanAndValidatePhone telefon numarasını temizler (sadece rakam bırakır) ve 10 haneli olup olmadığını kontrol eder.
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
		return "", errors.New("telefon numarası başında 0 olmadan tam olarak 10 haneli olmalıdır")
	}
	
	return cleaned, nil
}
