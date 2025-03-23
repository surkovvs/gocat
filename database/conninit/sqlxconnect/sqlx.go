package sqlxconnect

import (
	"database/sql"
	"fmt"

	"github.com/surkovvs/gocat/database/conninit"
	"github.com/surkovvs/gocat/database/conninit/dbloggers"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	sqldblogger "github.com/simukti/sqldb-logger"
)

func InitSQLx(cfg conninit.Config) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
		cfg.Host, cfg.Port,
		cfg.DBName,
		cfg.User,
		cfg.Pass,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if cfg.Logger != nil {
		db = sqldblogger.OpenDriver(dsn, db.Driver(), dbloggers.NewSQLDBLogger(cfg.Logger))
	}

	dbx := sqlx.NewDb(db, "postgres")
	if err := dbx.Ping(); err != nil {
		return nil, err
	}

	return dbx, nil
}
