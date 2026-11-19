// Package database is the local re-implementation of go-common's database
// package (constitution, Template Exception 4). Names and signatures match
// go-common v1.0.0; only the subset this service uses is implemented.
//
// Unlike go-common, Connect registers no audit plugin: this service has no
// users, its audit columns are *_by_name columns filled by the usecases, and
// no model has a CreatedBy/UpdatedBy field for such a plugin to manage.
package database

// DBConfig holds the PostgreSQL connection and pool settings.
type DBConfig struct {
	Host            string
	User            string
	Password        string
	DBName          string
	Port            int
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime int // minutes
}
