package provider

import (
	"net/url"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const (
	DefaultMaxOpenConn = 45
)

func open(scheme string, userName string, pass string, host string, path string, maxOpenConnections int, query url.Values) (*sqlx.DB, error) {
	u := &url.URL{
		Scheme:   scheme,
		User:     url.UserPassword(userName, pass),
		Host:     host,
		Path:     path,
		RawQuery: query.Encode(),
	}

	db, err := sqlx.Open(scheme, u.String())
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	if maxOpenConnections > 0 {
		db.SetMaxOpenConns(maxOpenConnections)
	} else {
		db.SetMaxOpenConns(DefaultMaxOpenConn)
	}

	return db, nil
}
