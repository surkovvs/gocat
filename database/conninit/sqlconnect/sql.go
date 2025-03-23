package sqlconnect

import (
	"database/sql"
	"fmt"

	"github.com/surkovvs/gocat/database/conninit"
	"github.com/surkovvs/gocat/database/conninit/dbloggers"

	_ "github.com/lib/pq"
	sqldblogger "github.com/simukti/sqldb-logger"
)

func InitSQL(cfg conninit.Config) (*sql.DB, error) {
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

	return db, nil
}
