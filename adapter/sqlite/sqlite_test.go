package sqlite

import (
	"os"
	"path/filepath"
	"testing"
)

// testDB creates a temporary SQLite database with migrations applied.
// It returns the DB and a cleanup function.
func testDB(t *testing.T) *DB {
	t.Helper()
	dir := t.TempDir()
	db, err := Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestOpen_InvalidPath(t *testing.T) {
	_, err := Open(filepath.Join(os.DevNull, "impossible", "test.db"))
	if err == nil {
		t.Fatal("expected error for invalid path")
	}
}

func TestMigrate_Idempotent(t *testing.T) {
	db := testDB(t)
	// Running migrate again should be a no-op.
	if err := Migrate(db); err != nil {
		t.Fatalf("second migrate: %v", err)
	}
}
