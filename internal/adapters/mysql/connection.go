package mysql

import (
	"TaskControlService/internal/config"
	"fmt"
	// "time"

	"github.com/jmoiron/sqlx"
)

func NewDB(cfg config.Config) (*sqlx.DB, error) {

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true",
		cfg.MySQL.User,
		cfg.MySQL.Password,
		cfg.MySQL.Host,
		cfg.MySQL.Port,
		cfg.MySQL.DB,
	)

	db, err := sqlx.Connect("mysql", dsn)
	// for i := 0; i < 10; i++ {
	// 	err := db.Ping()
	// if err == nil {
	// 	break
	// }

	// 	time.Sleep(2 * time.Second)
	// }
	
	if err != nil {
		return nil, err
	}


	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

	return db, nil
}