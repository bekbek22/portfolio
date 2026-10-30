package testdb_test

import (
	"testing"

	"github.com/bekbek101/portfolio/server/internal/database/testdb"
)

const probeSlug = "testdb-isolation-probe"

func TestTxIsRolledBack(t *testing.T) {
	t.Run("insert inside test transaction", func(t *testing.T) {
		tx := testdb.Tx(t)
		if _, err := tx.Exec(t.Context(),
			"INSERT INTO tags (slug, name, category) VALUES ($1, 'Probe', 'tool')", probeSlug); err != nil {
			t.Fatalf("insert: %v", err)
		}
		var n int
		if err := tx.QueryRow(t.Context(),
			"SELECT count(*) FROM tags WHERE slug = $1", probeSlug).Scan(&n); err != nil {
			t.Fatalf("count: %v", err)
		}
		if n != 1 {
			t.Fatalf("inserted row not visible inside its own transaction: count = %d", n)
		}
	}) // the subtest's cleanup rolls back here

	var n int
	if err := testdb.Tx(t).QueryRow(t.Context(),
		"SELECT count(*) FROM tags WHERE slug = $1", probeSlug).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 0 {
		t.Fatalf("row leaked after rollback: count = %d", n)
	}
}

func TestParallelTestsDoNotSeeEachOther(t *testing.T) {
	for _, name := range []string{"a", "b", "c"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			tx := testdb.Tx(t)

			if _, err := tx.Exec(t.Context(),
				"INSERT INTO tags (slug, name, category) VALUES ($1, $1, 'tool')", "testdb-parallel-"+name); err != nil {
				t.Fatalf("insert: %v", err)
			}
			var n int
			if err := tx.QueryRow(t.Context(),
				"SELECT count(*) FROM tags WHERE slug LIKE 'testdb-parallel-%'").Scan(&n); err != nil {
				t.Fatalf("count: %v", err)
			}
			if n != 1 {
				t.Errorf("saw %d parallel rows, want only this test's own row", n)
			}
		})
	}
}
