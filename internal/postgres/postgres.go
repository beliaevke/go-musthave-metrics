package postgres

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"musthave-metrics/internal/logger"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

type Settings struct {
	User    string
	Pass    string
	Host    string
	Port    string
	DBName  string
	ConnStr string
}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
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

func (s *Settings) Updates(ctx context.Context, db *pgxpool.Pool, metrics []Metrics) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint
	for _, m := range metrics {
		err := s.UpdateNew(ctx, db, m.MType, m.ID, m.Delta, m.Value)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (s *Settings) UpdateNew(ctx context.Context, db *pgxpool.Pool, t string, n string, d *int64, v *float64) error {
	if t == "gauge" {
		value := strconv.FormatFloat(*v, 'g', -1, 64)
		result := db.QueryRow(ctx, `
			SELECT gauges.mname
			FROM
				public.gauges
			WHERE
				gauges.mname=$1
		`, n)
		switch err := result.Scan(&n); err {
		case pgx.ErrNoRows:
			_, err = db.Exec(ctx, `
				INSERT INTO public.gauges
				(mname, mvalue)
				VALUES
				($1, $2);
			`, n, value)
			if err != nil {
				logger.Warnf("INSERT INTO Gauges: " + err.Error())
				return err
			}
		case nil:
			_, err = db.Exec(ctx, `
				UPDATE public.gauges
				SET mvalue=$2
				WHERE mname=$1;
			`, n, value)
			if err != nil {
				logger.Warnf("UPDATE Gauges: " + err.Error())
				return err
			}
		case err:
			logger.Warnf("QueryRow Gauges: " + err.Error())
			return err
		}
	} else if t == "counter" {
		delta := strconv.FormatInt(*d, 10)
		result := db.QueryRow(ctx, `
			SELECT counters.mname
			FROM
				public.counters
			WHERE
				counters.mname=$1
		`, n)
		switch err := result.Scan(&n); err {
		case pgx.ErrNoRows:
			_, err = db.Exec(ctx, `
				INSERT INTO public.counters
				(mname, mvalue)
				VALUES
				($1, $2);
			`, n, delta)
			if err != nil {
				logger.Warnf("INSERT INTO Counters: " + err.Error())
				return err
			}
		case nil:
			_, err = db.Exec(ctx, `
				UPDATE public.counters
				SET mvalue=mvalue+$2
				WHERE mname=$1;
			`, n, delta)
			if err != nil {
				logger.Warnf("UPDATE Counters: " + err.Error())
				return err
			}
		case err:
			logger.Warnf("QueryRow Counters: " + err.Error())
			return err
		}
	}
	return nil
}

func (s *Settings) GetValue(ctx context.Context, DatabaseDSN string, t string, n string) (string, error) {
	var val string
	db, err := pgxpool.New(ctx, DatabaseDSN)
	if err != nil {
		logger.Warnf("sql.Open(): " + err.Error())
	}
	defer db.Close()
	if t == "gauge" {
		result := db.QueryRow(ctx, `
			SELECT gauges.mvalue
			FROM
				public.gauges
			WHERE
				gauges.mname=$1
		`, n)
		switch err := result.Scan(&val); err {
		case pgx.ErrNoRows:
			return "0", nil
		case nil:
			return val, nil
		case err:
			logger.Warnf("QueryRow Gauges: " + err.Error())
			return "0", err
		}
	} else if t == "counter" {
		result := db.QueryRow(ctx, `
			SELECT counters.mvalue
			FROM
				public.counters
			WHERE
			counters.mname=$1
		`, n)
		switch err := result.Scan(&val); err {
		case pgx.ErrNoRows:
			return "0", nil
		case nil:
			return val, nil
		case err:
			logger.Warnf("QueryRow Counters: " + err.Error())
			return "0", err
		}
	}
	return val, nil
}

func SetDB(ctx context.Context, DatabaseDSN string) {
	db, err := sql.Open("pgx", DatabaseDSN)
	if err != nil {
		logger.Warnf("sql.Open(): " + err.Error())
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Warnf("goose: failed to close DB: " + err.Error())
		}
	}()
	goose.SetBaseFS(embedMigrations)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := goose.UpContext(ctx, db, "migrations"); err != nil {
		logger.Warnf("goose: run failed  " + err.Error())
	}
}
