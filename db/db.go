package db

import (
	"database/sql"
	"errors"
	"flag"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var pollDB = flag.String("poll_db", "mysql://root@unix("+SockFile+")/polls?", "Polls DB URI")
var db_test_base = flag.String("db_test_base", "mysql://root@unix("+SockFile+")/", "Database file")

var SockFile = getSocketFile()

func getSocketFile() string {
	var sock = "/tmp/mysql.sock"
	if _, err := os.Stat(sock); os.IsNotExist(err) {
		// try common linux/Ubuntu socket file location
		return "/var/run/mysqld/mysqld.sock"
	}
	return sock
}

func Connect() (*sql.DB, error) {
	dbc, err := ConnectWithURI(*pollDB)
	if err != nil {
		return nil, err
	}
	return dbc, nil
}

func ConnectWithURI(uri string) (*sql.DB, error) {
	dbc, err := connect(uri)
	if err != nil {
		return nil, err
	}

	return dbc, nil
}

func connect(connectStr string) (*sql.DB, error) {
	const prefix = "mysql://"
	if !strings.HasPrefix(connectStr, prefix) {
		return nil, errors.New("db: URI is missing mysql:// prefix")
	}
	connectStr = connectStr[len(prefix):]

	if connectStr[len(connectStr)-1] != '?' {
		connectStr += "&"
	}
	connectStr += "parseTime=true&collation=utf8mb4_general_ci"

	dbc, err := sql.Open("mysql", connectStr)
	if err != nil {
		return nil, err
	}

	dbc.SetMaxOpenConns(100)
	dbc.SetMaxIdleConns(50)
	dbc.SetConnMaxLifetime(time.Minute)

	return dbc, nil
}

func ConnectForTesting(t *testing.T) *sql.DB {
	return connectAndResetForTesting(
		t, "/src/lightning-poll/db/schema.sql")
}

func connectAndResetForTesting(
	t *testing.T, schemaPath string) *sql.DB {

	uri := os.Getenv("DB_TEST_BASE")
	if uri == "" {
		uri = *db_test_base
	}

	uri += "test?"

	dbc, err := connect(uri)
	if err != nil {
		t.Fatalf("connect error: %v", err)
		return nil
	}

	// Multiple connections are problematic for unit tests since they
	// introduce concurrency issues.
	dbc.SetMaxOpenConns(1)

	if _, err := dbc.Exec("set time_zone='+00:00';"); err != nil {
		t.Errorf("Error setting time_zone: %v", err)
	}
	_, err = dbc.Exec("set sql_mode=if(@@version<'5.7', 'STRICT_TRANS_TABLES,NO_ENGINE_SUBSTITUTION', @@sql_mode);")
	if err != nil {
		t.Errorf("Error setting strict mode: %v", err)
	}

	schema, err := ioutil.ReadFile(os.Getenv("GOPATH") + schemaPath)
	if err != nil {
		t.Errorf("Error reading schema: %s", err.Error())
		return nil
	}
	for _, q := range strings.Split(string(schema), ";") {
		q = strings.TrimSpace(q)
		if q == "" {
			continue
		}

		q = strings.Replace(
			q, "create table", "create temporary table", 1)

		// Temporary tables don't support fulltext indexes.
		q = strings.Replace(
			q, "fulltext", "index", -1)

		_, err = dbc.Exec(q)
		if err != nil {
			t.Fatalf("Error executing %s: %s", q, err.Error())
			return nil
		}
	}

	return dbc
}
