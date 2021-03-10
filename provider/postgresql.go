package provider

import (
	"encoding/json"
	"errors"
	"net/url"

	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"
)

type PostgreSQL struct {
	*sqlx.DB
}

func NewPostgreSQLConnection(scheme string, host string, path string, user string, pass string, dbName string) (*PostgreSQL, error) {
	query := url.Values{}
	if dbName != "" {
		query.Add("dbname", dbName)
	}

	db, err := open(scheme, user, pass, host, path, DefaultMaxOpenConn, query)
	if err != nil {
		return nil, err
	}

	return &PostgreSQL{db}, nil
}

func getPostgreSQLError(b []byte) (code string) {
	er := struct {
		Code string `json:"error"`
	}{}

	json.Unmarshal(b, &er)

	return er.Code
}

func (postgre *PostgreSQL) PostgreSQLQueryRow(queryString string, args ...interface{}) ([]byte, error) {
	var respByte []byte

	if err := postgre.QueryRow(queryString, args...).Scan(&respByte); err != nil {
		return nil, err
	}

	if code := getPostgreSQLError(respByte); code != "" {
		return nil, errors.New(code)
	}

	return respByte, nil
}
