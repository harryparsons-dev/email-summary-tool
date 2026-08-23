package migrations

import "testing"

func TestSuccessfulUpMigration(t *testing.T) {
	version, name, ok := successfulUpMigration(
		"%v (%v)\n",
		"20260823182333/u create_projects",
		"12.5ms",
	)

	if !ok {
		t.Fatal("expected successful up migration message to be recognized")
	}
	if version != "20260823182333" {
		t.Fatalf("expected version 20260823182333, got %q", version)
	}
	if name != "create_projects" {
		t.Fatalf("expected migration name create_projects, got %q", name)
	}
}

func TestSuccessfulUpMigrationRejectsOtherMessages(t *testing.T) {
	tests := []struct {
		name   string
		format string
		args   []any
	}{
		{
			name:   "down migration",
			format: "%v (%v)\n",
			args:   []any{"20260823182333/d create_projects", "12.5ms"},
		},
		{
			name:   "verbose start message",
			format: "Read and execute %v\n",
			args:   []any{"20260823182333/u create_projects"},
		},
		{
			name:   "error",
			format: "error: %v",
			args:   []any{"migration failed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, _, ok := successfulUpMigration(tt.format, tt.args...); ok {
				t.Fatal("expected message not to be recognized as a successful up migration")
			}
		})
	}
}
