package migrations

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
)

type migration struct {
	Name string
	// Requires lists the names of migrations that must run before this one.
	// It is the explicit dependency edge used to order the queue — a migration
	// that creates a table referenced by a foreign key declares the migration
	// that creates the referenced table here. Empty for migrations that only
	// depend on their natural numeric sequence.
	Requires []string
	Migrate  func(*sql.Tx) error
}

//go:embed queries/create_migrations_table.sql
var createMigrationDatabaseQuery string

//go:embed queries/insert_migration.sql
var insertMigrationQuery string

//go:embed queries/migration_exists.sql
var migrationExistsQuery string

var migrationList = []migration{}

var (
	initMigrationListFn       = initMigrationList
	createMigrationDatabaseFn = createMigrationDatabase
	runMigrationFn            = runMigration
	orderMigrationsFn         = orderMigrations
)

func Init(db *sql.DB) {
	if db == nil {
		log.Println("Database connection is nil")
		panic("Database connection is nil")
	}
	initMigrationListFn()
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		panic("Failed to begin transaction: " + err.Error())
	}
	defer tx.Rollback()

	err = createMigrationDatabaseFn(tx)

	if err != nil {
		panic("Failed to create migrations table: " + err.Error())
	}

	ordered, err := orderMigrationsFn(migrationList)
	if err != nil {
		panic("Failed to order migrations: " + err.Error())
	}

	for _, m := range ordered {
		if err := runMigrationFn(tx, m.Name, m.Migrate); err != nil {
			panic("Failed to run migration " + m.Name + ": " + err.Error())
		}
	}

	if err := tx.Commit(); err != nil {
		panic("Failed to commit transaction: " + err.Error())
	}
	log.Println("All migrations applied successfully")

}

func initMigrationList() {
	logMigrationList()
	diaryMigrationList()
	fileMigrationList()
	musicMigrationList()
	videoMigrationList()
	workerMigrationList()
	configurationMigrationList()
	notificationsMigrationList()
	systemEventMigrationList()
	capturesMigrationList()
	librariesMigrationList()
	aiProvidersMigrationList()
	watchFoldersMigrationList()
	assistantMigrationList()
	accessControlMigrationList()
	trashMigrationList()
	storageRootsMigrationList()
	emailMigrationList()
}

func createMigrationDatabase(tx *sql.Tx) error {
	_, err := tx.Exec(createMigrationDatabaseQuery)
	return err
}

func recordMigration(tx *sql.Tx, name string) error {
	_, err := tx.Exec(insertMigrationQuery, name)
	return err
}

func migrationExists(tx *sql.Tx, name string) (bool, error) {
	rows, err := tx.Query(migrationExistsQuery, name)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	if rows.Next() {
		var count int
		if err := rows.Scan(&count); err != nil {
			return false, err
		}
		return count > 0, nil
	}
	return false, nil
}

func runMigration(tx *sql.Tx, name string, migrationFunc func(*sql.Tx) error) error {
	exists, err := migrationExists(tx, name)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	if err := migrationFunc(tx); err != nil {
		return err
	}

	return recordMigration(tx, name)
}

func addMigration(name string, migrationFunc func(*sql.Tx) error) {
	migrationList = append(migrationList,
		migration{
			Name:    name,
			Migrate: migrationFunc,
		})

}

// addMigrationRequiring registers a migration that must run after the named
// prerequisites, regardless of the order it is appended to the list. Use it
// whenever a migration depends on a table/column created by another (typically
// a foreign key), so the queue ordering enforces the dependency explicitly.
func addMigrationRequiring(name string, requires []string, migrationFunc func(*sql.Tx) error) {
	migrationList = append(migrationList,
		migration{
			Name:     name,
			Requires: requires,
			Migrate:  migrationFunc,
		})
}

// migrationSequence extracts the leading numeric prefix of a migration name
// ("0040_create_email_message_table" -> 40), used as the natural execution
// order. A name without a numeric prefix sorts last but deterministically (by
// name), so ordering is always stable.
func migrationSequence(name string) int {
	digits := name
	if idx := strings.IndexByte(name, '_'); idx >= 0 {
		digits = name[:idx]
	}
	seq, err := strconv.Atoi(digits)
	if err != nil {
		return math.MaxInt32
	}
	return seq
}

// orderMigrations returns the migrations in execution order: ascending numeric
// sequence, with every declared prerequisite guaranteed to run before its
// dependents. It is a topological sort seeded by the sequence number (Kahn's
// algorithm, always draining the lowest-sequence ready migration first), so the
// default order is the natural numeric one and explicit Requires only add hard
// constraints. It fails on an unknown prerequisite or a dependency cycle rather
// than silently running migrations in a broken order.
func orderMigrations(list []migration) ([]migration, error) {
	byName := make(map[string]migration, len(list))
	for _, m := range list {
		byName[m.Name] = m
	}

	indegree := make(map[string]int, len(list))
	dependents := make(map[string][]string, len(list))
	for _, m := range list {
		if _, seen := indegree[m.Name]; !seen {
			indegree[m.Name] = 0
		}
		for _, req := range m.Requires {
			if _, ok := byName[req]; !ok {
				return nil, fmt.Errorf("migration %q requires unknown migration %q", m.Name, req)
			}
			indegree[m.Name]++
			dependents[req] = append(dependents[req], m.Name)
		}
	}

	less := func(a, b string) bool {
		sa, sb := migrationSequence(a), migrationSequence(b)
		if sa != sb {
			return sa < sb
		}
		return a < b
	}

	ready := make([]string, 0, len(indegree))
	for name, degree := range indegree {
		if degree == 0 {
			ready = append(ready, name)
		}
	}

	ordered := make([]migration, 0, len(indegree))
	for len(ready) > 0 {
		sort.Slice(ready, func(i, j int) bool { return less(ready[i], ready[j]) })
		next := ready[0]
		ready = ready[1:]
		ordered = append(ordered, byName[next])
		for _, dep := range dependents[next] {
			indegree[dep]--
			if indegree[dep] == 0 {
				ready = append(ready, dep)
			}
		}
	}

	if len(ordered) != len(indegree) {
		return nil, fmt.Errorf("migration dependency cycle detected")
	}
	return ordered, nil
}
