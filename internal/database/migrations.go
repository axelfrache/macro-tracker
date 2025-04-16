package database

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
)

func (db *DB) ApplyMigrations(migrationsDir string) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("erreur lors de la création de la table migrations: %v", err)
	}

	files, err := ioutil.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("erreur lors de la lecture du dossier migrations: %v", err)
	}

	var migrations []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			migrations = append(migrations, file.Name())
		}
	}
	sort.Strings(migrations)

	for _, migration := range migrations {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM migrations WHERE name = $1", migration).Scan(&count)
		if err != nil {
			return fmt.Errorf("erreur lors de la vérification de la migration %s: %v", migration, err)
		}

		if count == 0 {
			content, err := ioutil.ReadFile(filepath.Join(migrationsDir, migration))
			if err != nil {
				return fmt.Errorf("erreur lors de la lecture du fichier %s: %v", migration, err)
			}

			_, err = db.Exec(string(content))
			if err != nil {
				return fmt.Errorf("erreur lors de l'exécution de la migration %s: %v", migration, err)
			}

			_, err = db.Exec("INSERT INTO migrations (name) VALUES ($1)", migration)
			if err != nil {
				return fmt.Errorf("erreur lors de l'enregistrement de la migration %s: %v", migration, err)
			}

			fmt.Printf("Migration appliquée: %s\n", migration)
		}
	}

	return nil
} 