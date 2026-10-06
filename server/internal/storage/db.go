package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/valkey-io/valkey-go"
)

type Postgres struct {
	Pool *pgxpool.Pool
}

func NewPostgres(dsn string) (*Postgres, error) {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("pgxpool: %w", err)
	}
	
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	
	return &Postgres{Pool: pool}, nil
}

func (p *Postgres) Close() {
	p.Pool.Close()
}

type Valkey struct {
	Client valkey.Client
}

func NewValkey(addr string) (*Valkey, error) {
	client, err := valkey.NewClient(valkey.ClientOption{InitAddress: []string{addr}})
	if err != nil {
		return nil, fmt.Errorf("valkey: %w", err)
	}
	return &Valkey{Client: client}, nil
}

func (v *Valkey) Close() {
	v.Client.Close()
}