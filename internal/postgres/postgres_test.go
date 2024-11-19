package postgres

import (
	"context"
	"reflect"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestNewPSQL(t *testing.T) {
	type args struct {
		user string
		pass string
		host string
		port string
		db   string
	}
	tests := []struct {
		name string
		args args
		want Settings
	}{
		{
			name: "1",
			args: args{},
			want: Settings{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewPSQL(tt.args.user, tt.args.pass, tt.args.host, tt.args.port, tt.args.db); reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPSQL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewPSQLStr(t *testing.T) {
	type args struct {
		connection string
	}
	tests := []struct {
		name string
		args args
		want Settings
	}{
		{
			name: "1",
			args: args{connection: "nil"},
			want: Settings{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewPSQLStr(tt.args.connection); reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPSQLStr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetDB(t *testing.T) {
	type args struct {
		ctx         context.Context
		DatabaseDSN string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "1",
			args: args{ctx: context.Background()},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetDB(tt.args.ctx, tt.args.DatabaseDSN)
		})
	}
}
