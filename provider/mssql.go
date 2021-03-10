package provider

import (
	"database/sql"
	"github.com/DmitriBeattie/custom-framework/utils/slice"
	"errors"
	mssql "github.com/denisenkom/go-mssqldb"
	"net/url"
	"reflect"

	_ "github.com/denisenkom/go-mssqldb"

	"github.com/jmoiron/sqlx"
)

type MSSQL struct {
	*sqlx.DB
}

func NewMSSQLConnection(scheme string, host string, path string, failoverPartnerHost string, user string, pass string, databaseName string) (*MSSQL, error) {
	query := url.Values{}

	if failoverPartnerHost != "" {
		query.Add("failoverpartner", failoverPartnerHost)
	}

	if databaseName != "" {
		query.Add("database", databaseName)
	}

	db, err := open(scheme, user, pass, host, path, DefaultMaxOpenConn, query)
	if err != nil {
		return nil, err
	}

	return &MSSQL{db}, nil
}

type List struct {
	ID int `db:"id"`
}

type LongList struct {
	ID int64 `db:"int64"`
}


func (w *MSSQL) GetDataByIDs(procName string, paramName string, ids interface{}, scanTo interface{}) error {
	if reflect.TypeOf(scanTo).Kind() != reflect.Ptr {
		return errors.New("Not a pointer")
	}

	tvp := mssql.TVP{}

	switch idsWithType := ids.(type) {
	case []int:
		slice.DistinctFromINT32Slice(&idsWithType)

		list := make([]List, 0, len(idsWithType))

		for i := range idsWithType {
			list[i].ID = idsWithType[i]
		}

		tvp.Value = list
		tvp.TypeName = "dbo.List"
	case []int64:
		slice.DistinctFromINT64Slice(&idsWithType)

		longList := make([]LongList, len(idsWithType))

		for i := range idsWithType {
			longList[i].ID = idsWithType[i]
		}

		tvp.Value = longList
		tvp.TypeName = "dbo.LongList"
	default:
		return errors.New("Not supported list")
	}

	rowsx, err := w.Queryx("EXEC " + procName + " @" + paramName, sql.Named(paramName, tvp))
	if rowsx != nil {
		defer rowsx.Close()
	}
	if err != nil {
		return err
	}

	return sqlx.StructScan(rowsx, scanTo)
}
