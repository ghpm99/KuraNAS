package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// Conexão com o banco de dados
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Teste 1: Query simples sem JOIN
	fmt.Println("=== Teste 1: Query simples ===")
	simpleQuery := `
		SELECT 
			hf.id,
			hf."name",
			hf."path",
			hf.parent_path,
			hf.format,
			hf."size",
			hf.updated_at,
			hf.created_at,
			hf.last_interaction,
			hf.last_backup,
			hf."type",
			hf.checksum,
			hf.deleted_at,
			hf.starred
		FROM home_file hf
		WHERE hf.format IN ($1)
		ORDER BY hf.TYPE, hf.NAME, hf.id DESC
		LIMIT $2 OFFSET $3
	`

	imageFormats := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".svg", ".webp"}
	rows, err := db.Query(simpleQuery, imageFormats, 10, 0)
	if err != nil {
		log.Printf("Erro na query simples: %v", err)
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
		var id int
		var name, path, parentPath, format, type_, checksum string
		var size int
		var updatedAt, createdAt, lastInteraction, lastBackup sql.NullTime
		var deletedAt sql.NullTime
		var starred bool

		err := rows.Scan(
			&id, &name, &path, &parentPath, &format, &size,
			&updatedAt, &createdAt, &lastInteraction, &lastBackup,
			&type_, &checksum, &deletedAt, &starred,
		)
		if err != nil {
			log.Printf("Erro no scan: %v", err)
			continue
		}

		fmt.Printf("Arquivo %d: %s (%s)\n", id, name, format)
	}
	fmt.Printf("Total de arquivos encontrados (query simples): %d\n\n", count)

	// Teste 2: Query com contagem total
	fmt.Println("=== Teste 2: Contagem total ===")
	var totalCount int
	countQuery := "SELECT COUNT(*) FROM home_file WHERE format IN ($1)"
	err = db.QueryRow(countQuery, imageFormats).Scan(&totalCount)
	if err != nil {
		log.Printf("Erro na contagem: %v", err)
	} else {
		fmt.Printf("Total de imagens no banco: %d\n\n", totalCount)
	}

	// Teste 3: Verificar formatos existentes
	fmt.Println("=== Teste 3: Formatos existentes ==="
	formatQuery := `
		SELECT format, COUNT(*) 
		FROM home_file 
		WHERE format IS NOT NULL 
		GROUP BY format 
		ORDER BY COUNT(*) DESC
	`
	rows, err = db.Query(formatQuery)
	if err != nil {
		log.Printf("Erro na query de formatos: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var format string
		var count int
		if err := rows.Scan(&format, &count); err != nil {
			continue
		}
		fmt.Printf("Formato %s: %d arquivos\n", format, count)
	}
}
