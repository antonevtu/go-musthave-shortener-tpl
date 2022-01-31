package app

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"log"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestExample(t *testing.T) {
	ctx := context.Background()

	// container and database
	container, db, err := CreateTestContainer(ctx, "pg")
	require.NoError(t, err)
	defer db.Close()
	defer container.Terminate(ctx)

	// migration
	//mig, err := NewPgMigrator(db)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//err = mig.Up()
	//if err != nil {
	//	t.Fatal(err)
	//}

	// test
	//r :=  &ItemRepository{db}
	//created, err := r.CreateItem(ctx, "desc")
	//if err != nil {
	//	t.Errorf("failed to create item: %s", err)
	//}
	//retrieved, err := r.GetItem(ctx, created.Id)
	//if err != nil {
	//	t.Errorf("failed to retrieve item: %s", err)
	//}
	//if created.Id != retrieved.Id {
	//	t.Errorf("created.Id (%s) != retrieved.Id (%s)", created.Id, retrieved.Id)
	//}
	//if created.Description != retrieved.Description {
	//	t.Errorf("created.Description != retrieved.Description (%s != %s)", created.Description, retrieved.Description)
	//}
}

func CreateTestContainer(ctx context.Context, dbname string) (testcontainers.Container, *pgxpool.Pool, error) {
	var env = map[string]string{
		"POSTGRES_PASSWORD": "password",
		"POSTGRES_USER":     "postgres",
		"POSTGRES_DB":       dbname,
	}
	var port = "5432/tcp"
	dbURL := func(port nat.Port) string {
		return fmt.Sprintf("postgres://postgres:password@localhost:%s/%s?sslmode=disable", port.Port(), dbname)
	}

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:latest",
			ExposedPorts: []string{port},
			Cmd:          []string{"postgres", "-c", "fsync=off"},
			Env:          env,
			WaitingFor:   wait.ForSQL(nat.Port(port), "postgres", dbURL).Timeout(time.Second * 50),
		},
		Started: true,
	}
	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return container, nil, fmt.Errorf("failed to start container: %s", err)
	}

	mappedPort, err := container.MappedPort(ctx, nat.Port(port))
	if err != nil {
		return container, nil, fmt.Errorf("failed to get container external port: %s", err)
	}

	log.Println("postgres container ready and running at port: ", mappedPort)

	url := fmt.Sprintf("postgres://postgres:password@localhost:%s/%s?sslmode=disable", mappedPort.Port(), dbname)

	db, err := pgxpool.Connect(ctx, url)
	//db, err := sql.Open("postgres", url)
	if err != nil {
		return container, db, fmt.Errorf("failed to establish database connection: %s", err)
	}

	return container, db, nil
}

func NewPgMigrator(db *sql.DB) (*migrate.Migrate, error) {
	_, path, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatalf("failed to get path")
	}

	sourceUrl := "file://" + filepath.Dir(path) + "/migrations"

	driver, err := postgres.WithInstance(db, &postgres.Config{})

	if err != nil {
		log.Fatalf("failed to create migrator driver: %s", err)
	}

	m, err := migrate.NewWithDatabaseInstance(sourceUrl, "postgres", driver)

	return m, err
}
