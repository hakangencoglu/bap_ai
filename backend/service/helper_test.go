package service

import (
	"testing"
)

// TestCleanAndValidatePhone telefon temizleme ve doğrulama fonksiyonunun doğruluğunu test eder.
func TestCleanAndValidatePhone(t *testing.T) {
	tests := []struct {
		name    string
		phone   string
		want    string
		wantErr bool
	}{
		{
			name:    "Bos telefon numarasi",
			phone:   "",
			want:    "",
			wantErr: false,
		},
		{
			name:    "Gecerli 10 haneli telefon numarasi",
			phone:   "5051234567",
			want:    "5051234567",
			wantErr: false,
		},
		{
			name:    "Basinda 0 olan gecerli telefon numarasi",
			phone:   "05051234567",
			want:    "5051234567",
			wantErr: false,
		},
		{
			name:    "Formatli telefon numarasi",
			phone:   "(505) 123 45 67",
			want:    "5051234567",
			wantErr: false,
		},
		{
			name:    "Eksik haneli telefon numarasi",
			phone:   "505123456",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Fazla haneli telefon numarasi",
			phone:   "50512345678",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Karakter iceren gecersiz telefon numarasi",
			phone:   "505-ABC-4567",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cleanAndValidatePhone(tt.phone)
			if (err != nil) != tt.wantErr {
				t.Errorf("cleanAndValidatePhone() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("cleanAndValidatePhone() got = %v, want %v", got, tt.want)
			}
		})
	}
}
