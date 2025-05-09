package database

import (
	"fmt"
	"os"
	"path/filepath"

	migrator "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/sirupsen/logrus"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
)

var (
	FCS *sqlx.DB
)

type SSLMode string

const (
	SSLModeEnable  SSLMode = "enable"
	SSLModeDisable SSLMode = "disable"
)

// ConnectAndMigrate function connects with a given database and returns error if there is any error
func ConnectAndMigrate(host, port, databaseName, user, password string, sslMode SSLMode) error {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, databaseName, sslMode)
	DB, err := sqlx.Open("postgres", connStr)

	if err != nil {
		return err
	}

	err = DB.Ping()
	if err != nil {
		return err
	}
	FCS = DB
	return migrateUp(DB)
}

func ShutdownDatabase() error {
	return FCS.Close()
}

func findMigrationsFolderRoot() string {
	workingDirectory, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	lastDir := workingDirectory
	myUniqueRelativePath := "database/migrations"
	for {
		currentPath := fmt.Sprintf("%s/%s", lastDir, myUniqueRelativePath)
		fi, statErr := os.Stat(currentPath)
		if statErr == nil {
			mode := fi.Mode()
			if mode.IsDir() {
				return currentPath
			}
		}
		newDir := filepath.Dir(lastDir)
		if newDir == "/" || newDir == lastDir {
			return ""
		}
		lastDir = newDir
	}
}

// migrateUp function migrate the database and handles the migration logic
func migrateUp(db *sqlx.DB) error {
	db.Driver()
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return err
	}
	path := findMigrationsFolderRoot()
	m, err := migrator.NewWithDatabaseInstance(
		"file://database/migrations",
		"postgres://postgres:Ved1234@localhost:5432/farmako_coupon_service?sslmode=disable&search_path=public",
		driver)
	fmt.Sprintf("file://%s", path)

	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && err != migrator.ErrNoChange {
		return err
	}
	return nil
}

// Tx provides the transaction wrapper
func Tx(fn func(tx *sqlx.Tx) error) error {
	tx, err := FCS.Beginx()
	if err != nil {
		return fmt.Errorf("failed to start a transaction: %+v", err)
	}
	defer func() {
		if err != nil {
			if rollBackErr := tx.Rollback(); rollBackErr != nil {
				logrus.Errorf("failed to rollback tx: %s", rollBackErr)
			}
			return
		}
		if commitErr := tx.Commit(); commitErr != nil {
			logrus.Errorf("failed to commit tx: %s", commitErr)
		}
	}()
	err = fn(tx)
	return err
}

// // SetupBindVars prepares the SQL statement for batch insert
// func SetupBindVars(stmt, bindVars string, length int) string {
// 	bindVars += ","
// 	stmt = fmt.Sprintf(stmt, strings.Repeat(bindVars, length))
// 	return replaceSQL(strings.TrimSuffix(stmt, ","), "?")
// }

// // replaceSQL replaces the instance occurrence of any string pattern with an increasing $n based sequence
// func replaceSQL(old, searchPattern string) string {
// 	tmpCount := strings.Count(old, searchPattern)
// 	for m := 1; m <= tmpCount; m++ {
// 		old = strings.Replace(old, searchPattern, "$"+strconv.Itoa(m), 1)
// 	}
// 	return old
// }

// func SetupColumnAndOrder(stmt, column, order string) (string, error) {
// 	if !strings.Contains(stmt, "$COLUMN") {
// 		return "", errors.New("query statement does not contain $COLUMN field")
// 	}
// 	if !strings.Contains(stmt, "$ORDER") {
// 		return "", errors.New("query statement does not contain $ORDER field")
// 	}
// 	sql := strings.Replace(strings.Replace(stmt, "$COLUMN", column, 1), "$ORDER", order, 1)
// 	return sql, nil
// }
