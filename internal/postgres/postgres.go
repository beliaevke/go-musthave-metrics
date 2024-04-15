package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Settings struct {
	User    string
	Pass    string
	Host    string
	Port    string
	DBName  string
	ConnStr string
}

func NewPSQL(user string, pass string, host string, port string, db string) Settings {
	return Settings{
		User:   user,
		Pass:   pass,
		Host:   host,
		Port:   port,
		DBName: db,
		ConnStr: fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
			host+":"+port, user, pass, db),
	}
}

func NewPSQLStr(connection string) Settings {
	return Settings{
		ConnStr: connection,
	}
}

func (s *Settings) Ping(ctx context.Context) error {
	dbpool, err := pgxpool.New(ctx, s.ConnStr)
	if err != nil {
		return err
	}
	defer dbpool.Close()
	err = dbpool.Ping(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *Settings) Update(t string, n string, v string) error {
	// depricated
	return nil
}

func (s *Settings) UpdateNew(t string, n string, delta *int64, value *float64) error {
	return nil
}

func (s *Settings) GetValue(t string, n string) (any, error) {
	var val any
	return val, nil
}
