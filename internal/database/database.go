package database

import "context"

type DB interface {
	Ping(context.Context) error
	Close() error
}
