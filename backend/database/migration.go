package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// RunMigrations fonksiyonu, migrations/ klasöründeki SQL dosyalarını sırasıyla çalıştırır.
func RunMigrations(db *sql.DB, migrationsDir string) error {
	// Migration dosyalarını oku
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("migration dizini okunamadı (%s): %v", migrationsDir, err)
	}

	// Sadece .sql dosyalarını filtrele ve sırala
	var sqlFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			sqlFiles = append(sqlFiles, entry.Name())
		}
	}
	sort.Strings(sqlFiles)

	if len(sqlFiles) == 0 {
		log.Println("Migration: Çalıştırılacak SQL dosyası bulunamadı.")
		return nil
	}

	// Her SQL dosyasını sırayla çalıştır
	for _, fileName := range sqlFiles {
		filePath := filepath.Join(migrationsDir, fileName)

		// SQL dosyasını oku
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("migration dosyası okunamadı (%s): %v", fileName, err)
		}

		// SQL'i çalıştır
		_, err = db.Exec(string(content))
		if err != nil {
			return fmt.Errorf("migration çalıştırılamadı (%s): %v", fileName, err)
		}

		log.Printf("Migration başarıyla çalıştırıldı: %s\n", fileName)
	}

	log.Printf("Toplam %d migration başarıyla tamamlandı.\n", len(sqlFiles))
	return nil
}
