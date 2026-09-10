package main

import (
	"database/sql"
	"fmt"
	"os"

	"bap_ai/configs"
	"bap_ai/backend/database"
	"bap_ai/backend/repository"
	"bap_ai/backend/service"
)

func main() {
	configs.LoadConfig()
	os.Setenv("DB_HOST", "localhost")
	configs.AppConfig.DBHost = "localhost"
	database.Connect()
	db := database.DB

	projeRepo := repository.NewProjeRepository(db)
	hakemRepo := repository.NewHakemRepository(db)
	revizyonRepo := repository.NewRevizyonRepository(db)
	projeSrv := service.NewProjeService(projeRepo, hakemRepo, revizyonRepo)

	// Direct repository call for tto
	projelerDirect, err := projeRepo.GetProjectsForWorkflow("tto", "yururlukte", 0)
	fmt.Printf("Direct Repo GetProjectsForWorkflow('tto', 'yururlukte'): count=%d, err=%v\n", len(projelerDirect), err)

	// Service call for tto
	projelerService, err := projeSrv.GetProjectsForWorkflow("tto", 0)
	fmt.Printf("Service GetProjectsForWorkflow('tto'): count=%d, err=%v\n", len(projelerService), err)

	// Service call for admin
	projelerAdmin, err := projeSrv.GetProjectsForWorkflow("admin", 0)
	fmt.Printf("Service GetProjectsForWorkflow('admin'): count=%d, err=%v\n", len(projelerAdmin), err)

	// Check all projects status in DB
	rows, err := db.Query("SELECT p.proje_id, p.proje_kodu, p.durum_id, pd.durum_adi, pd.durum_etiketi FROM proje p LEFT JOIN proje_durum pd ON p.durum_id = pd.durum_id")
	if err == nil {
		defer rows.Close()
		fmt.Println("--- DB Proje Durumları ---")
		for rows.Next() {
			var id int
			var kodu, durumAdi, durumEtiketi sql.NullString
			var durumID sql.NullInt64
			rows.Scan(&id, &kodu, &durumID, &durumAdi, &durumEtiketi)
			fmt.Printf("ID: %d, Kodu: %s, DurumID: %v, DurumAdi: %s, DurumEtiketi: %s\n", id, kodu.String, durumID.Int64, durumAdi.String, durumEtiketi.String)
		}
	}
}
