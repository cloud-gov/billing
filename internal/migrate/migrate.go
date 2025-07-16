package migrate

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/tern/v2/migrate"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"

	"github.com/cloud-gov/billing/sql/migrations"
)

// Migrate migrates the database to the latest migrations for the application and the River queue. It is idempotent.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	err := migrateRiver(ctx, pool)
	if err != nil {
		return err
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}

	err = migrateTern(ctx, conn.Conn())
	return err
}

// migrateRiver uses the rivermigrate package to run "up" migrations for the River queue.
func migrateRiver(ctx context.Context, conn *pgxpool.Pool) error {
	migrator, err := rivermigrate.New(riverpgxv5.New(conn), nil)
	if err != nil {
		return err
	}
	// If already migrated to latest, this is a noop.
	_, err = migrator.Migrate(ctx, rivermigrate.DirectionUp, nil)
	if err != nil {
		return err
	}
	return nil
}

// migrateTern uses the tern package to execute "up" migrations for the billing service schema.
func migrateTern(ctx context.Context, conn *pgx.Conn) error {
	m, err := migrate.NewMigrator(context.Background(), conn, "schema_version")
	if err != nil {
		return err
	}
	err = m.LoadMigrations(migrations.FS)
	if err != nil {
		return err
	}

	// If already migrated to latest, this is a noop.
	err = m.Migrate(ctx)
	return err
}
