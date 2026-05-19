package database

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func RunMigrations(db *pgxpool.Pool) error {
	ctx := context.Background()

	steps := []struct {
		name string
		fn   func(context.Context, *pgxpool.Pool) error
	}{
		{"create schema",  createSchema},
		{"create enums",   createEnums},
		{"create tables",  createTables},
		{"add foreign keys", addForeignKeys},
		{"seed data",      seedData},
	}

	for _, step := range steps {
		log.Printf("migration: running — %s", step.name)
		if err := step.fn(ctx, db); err != nil {
			return fmt.Errorf("migration failed at [%s]: %w", step.name, err)
		}
		log.Printf("migration: done    — %s", step.name)
	}

	log.Println("all migrations completed successfully")
	return nil
}

// ── Step 1: Schema ────────────────────────────────────────────────────────────

func createSchema(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `CREATE SCHEMA IF NOT EXISTS master`)
	if err != nil {
		return fmt.Errorf("createSchema: %w", err)
	}
	return nil
}

// ── Step 2: Enums ─────────────────────────────────────────────────────────────

func createEnums(ctx context.Context, db *pgxpool.Pool) error {
	query := `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_type t
				JOIN pg_namespace n ON n.oid = t.typnamespace
				WHERE t.typname = 'status_enum'
				AND n.nspname = 'master'
			) THEN
				CREATE TYPE master.status_enum AS ENUM ('A', 'I', 'D');
			END IF;
		END $$;
	`
	if _, err := db.Exec(ctx, query); err != nil {
		return fmt.Errorf("createEnums: %w", err)
	}
	return nil
}

// ── Step 3: Tables ────────────────────────────────────────────────────────────

func createTables(ctx context.Context, db *pgxpool.Pool) error {
	tables := []struct {
		name  string
		query string
	}{
		{
			name: "roles",
			query: `
				CREATE TABLE IF NOT EXISTS master.roles (
					role_no      SMALLSERIAL                PRIMARY KEY,
					role_id      UUID                       NOT NULL DEFAULT gen_random_uuid(),
					role_name    VARCHAR(65)                NOT NULL,
					role_details JSONB                      NOT NULL DEFAULT '[]',
					status       master.status_enum         NOT NULL DEFAULT 'A',
					created_at   TIMESTAMPTZ                NOT NULL DEFAULT NOW(),
					created_by   INTEGER                             DEFAULT NULL,
					updated_at   TIMESTAMPTZ                NOT NULL DEFAULT NOW(),
					updated_by   INTEGER                             DEFAULT NULL
				);
				CREATE INDEX IF NOT EXISTS idx_roles_role_id ON master.roles(role_id);
				CREATE INDEX IF NOT EXISTS idx_roles_status  ON master.roles(status);
			`,
		},
		{
			name: "sections",
			query: `
				CREATE TABLE IF NOT EXISTS master.sections (
					section_no    SMALLSERIAL                PRIMARY KEY,
					section_id    UUID                       NOT NULL DEFAULT gen_random_uuid(),
					section_name  VARCHAR(65)                NOT NULL,
					section_path  VARCHAR(65)                NOT NULL,
					section_icon  VARCHAR(65)                NOT NULL,
					section_order SMALLINT                   NOT NULL DEFAULT 0,
					status        master.status_enum         NOT NULL DEFAULT 'A',
					created_at    TIMESTAMPTZ                NOT NULL DEFAULT NOW(),
					created_by    INTEGER                             DEFAULT NULL,
					updated_at    TIMESTAMPTZ                NOT NULL DEFAULT NOW(),
					updated_by    INTEGER                             DEFAULT NULL
				);
				CREATE INDEX IF NOT EXISTS idx_sections_section_id    ON master.sections(section_id);
				CREATE INDEX IF NOT EXISTS idx_sections_section_order  ON master.sections(section_order);
				CREATE INDEX IF NOT EXISTS idx_sections_status         ON master.sections(status);
			`,
		},
		{
			name: "pages",
			query: `
				CREATE TABLE IF NOT EXISTS master.pages (
					page_no    SMALLSERIAL                PRIMARY KEY,
					page_id    UUID                       NOT NULL DEFAULT gen_random_uuid(),
					section_no SMALLINT                   NOT NULL,
					page_name  VARCHAR(65)                NOT NULL,
					page_path  VARCHAR(255)               NOT NULL,
					page_order SMALLINT                   NOT NULL DEFAULT 0,
					status     master.status_enum         NOT NULL DEFAULT 'A',
					created_at TIMESTAMPTZ                NOT NULL DEFAULT NOW(),
					created_by INTEGER                             DEFAULT NULL,
					updated_at TIMESTAMPTZ                NOT NULL DEFAULT NOW(),
					updated_by INTEGER                             DEFAULT NULL
				);
				CREATE INDEX IF NOT EXISTS idx_pages_page_id    ON master.pages(page_id);
				CREATE INDEX IF NOT EXISTS idx_pages_section_no ON master.pages(section_no);
				CREATE INDEX IF NOT EXISTS idx_pages_page_order ON master.pages(page_order);
				CREATE INDEX IF NOT EXISTS idx_pages_status     ON master.pages(status);
			`,
		},
		{
			name: "users",
			query: `
				CREATE TABLE IF NOT EXISTS master.users (
					profile_no    SERIAL                     PRIMARY KEY,
					profile_id    UUID                       NOT NULL DEFAULT gen_random_uuid(),
					role_no       SMALLINT                            DEFAULT NULL,
					user_fullname VARCHAR(65)                NOT NULL,
					email_id      VARCHAR(65)  UNIQUE        NOT NULL,
					gender        VARCHAR(10)                         DEFAULT NULL,
					dob           DATE                                DEFAULT NULL,
					mobile_no     VARCHAR(15)                         DEFAULT NULL,
					address       JSONB                      NOT NULL DEFAULT '{}',
					x_api_key     VARCHAR(55)                         DEFAULT NULL,
					secret_key    VARCHAR(55)                         DEFAULT NULL,
					status        master.status_enum         NOT NULL DEFAULT 'A',
					created_at    TIMESTAMPTZ                NOT NULL DEFAULT NOW(),
					created_by    INTEGER                             DEFAULT NULL,
					updated_at    TIMESTAMPTZ                NOT NULL DEFAULT NOW(),
					updated_by    INTEGER                             DEFAULT NULL
				);
				CREATE INDEX IF NOT EXISTS idx_users_profile_id ON master.users(profile_id);
				CREATE INDEX IF NOT EXISTS idx_users_role_no    ON master.users(role_no);
				CREATE INDEX IF NOT EXISTS idx_users_email_id   ON master.users(email_id);
				CREATE INDEX IF NOT EXISTS idx_users_status     ON master.users(status);
			`,
		},
		{
			name: "credentials",
			query: `
				CREATE TABLE IF NOT EXISTS master.credentials (
					credential_no SERIAL                     PRIMARY KEY,
					credential_id UUID                       NOT NULL DEFAULT gen_random_uuid(),
					profile_no    INTEGER                    NOT NULL,
					username      VARCHAR(65)  UNIQUE        NOT NULL,
					password      VARCHAR(255)               NOT NULL,
					status        master.status_enum         NOT NULL DEFAULT 'I',
					created_at    TIMESTAMPTZ                NOT NULL DEFAULT NOW(),
					created_by    INTEGER                             DEFAULT NULL,
					updated_at    TIMESTAMPTZ                NOT NULL DEFAULT NOW(),
					updated_by    INTEGER                             DEFAULT NULL
				);
				CREATE INDEX IF NOT EXISTS idx_credentials_credential_id ON master.credentials(credential_id);
				CREATE INDEX IF NOT EXISTS idx_credentials_profile_no    ON master.credentials(profile_no);
				CREATE INDEX IF NOT EXISTS idx_credentials_username      ON master.credentials(username);
				CREATE INDEX IF NOT EXISTS idx_credentials_status        ON master.credentials(status);
			`,
		},
	}

	for _, t := range tables {
		if _, err := db.Exec(ctx, t.query); err != nil {
			return fmt.Errorf("createTables: [%s]: %w", t.name, err)
		}
	}

	return nil
}

// ── Step 4: Foreign Keys ──────────────────────────────────────────────────────

func addForeignKeys(ctx context.Context, db *pgxpool.Pool) error {
	constraints := []struct {
		name  string
		query string
	}{
		{
			name: "fk_users_role_no",
			query: `
				DO $$ BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM information_schema.table_constraints
						WHERE constraint_name = 'fk_users_role_no'
					) THEN
						ALTER TABLE master.users
						ADD CONSTRAINT fk_users_role_no
						FOREIGN KEY (role_no) REFERENCES master.roles(role_no)
						ON UPDATE SET NULL ON DELETE SET NULL;
					END IF;
				END $$;
			`,
		},
		{
			name: "fk_credentials_profile_no",
			query: `
				DO $$ BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM information_schema.table_constraints
						WHERE constraint_name = 'fk_credentials_profile_no'
					) THEN
						ALTER TABLE master.credentials
						ADD CONSTRAINT fk_credentials_profile_no
						FOREIGN KEY (profile_no) REFERENCES master.users(profile_no)
						ON DELETE CASCADE;
					END IF;
				END $$;
			`,
		},
		{
			name: "fk_pages_section_no",
			query: `
				DO $$ BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM information_schema.table_constraints
						WHERE constraint_name = 'fk_pages_section_no'
					) THEN
						ALTER TABLE master.pages
						ADD CONSTRAINT fk_pages_section_no
						FOREIGN KEY (section_no) REFERENCES master.sections(section_no)
						ON DELETE CASCADE;
					END IF;
				END $$;
			`,
		},
	}

	for _, c := range constraints {
		if _, err := db.Exec(ctx, c.query); err != nil {
			return fmt.Errorf("addForeignKeys: [%s]: %w", c.name, err)
		}
	}

	return nil
}

// ── Step 5: Seed Data (fully transactional with rollback) ─────────────────────

func seedData(ctx context.Context, db *pgxpool.Pool) error {
	// Skip if already seeded
	var count int
	if err := db.QueryRow(ctx, `SELECT COUNT(*) FROM master.users`).Scan(&count); err != nil {
		return fmt.Errorf("seedData: count check: %w", err)
	}
	if count > 0 {
		log.Println("seed data already exists, skipping")
		return nil
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("seedData: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
			log.Printf("seedData: rolled back: %v", err)
		}
	}()

	// ── Roles ─────────────────────────────────────────────────────────────────
	var roleNo int
	err = tx.QueryRow(ctx, `
		INSERT INTO master.roles (role_id, role_name, role_details, status)
		VALUES ($1, $2, $3, $4)
		RETURNING role_no
	`, uuid.New(), "Super Admin", "[]", "A").Scan(&roleNo)
	if err != nil {
		return fmt.Errorf("seedData: insert role: %w", err)
	}

	// ── Sections ──────────────────────────────────────────────────────────────
	type sectionRow struct {
		id    uuid.UUID
		no    int
	}

	sectionInputs := []struct {
		name  string
		path  string
		icon  string
		order int
	}{
		{"Webmasters", "webmasters", "fa-laptop", 1},
		{"Employees",  "systems",    "fa-users",  2},
	}

	sectionNos := make([]int, len(sectionInputs))
	for i, s := range sectionInputs {
		err = tx.QueryRow(ctx, `
			INSERT INTO master.sections (section_id, section_name, section_path, section_icon, section_order, status)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING section_no
		`, uuid.New(), s.name, s.path, s.icon, s.order, "A").Scan(&sectionNos[i])
		if err != nil {
			return fmt.Errorf("seedData: insert section [%s]: %w", s.name, err)
		}
	}

	// ── Pages ─────────────────────────────────────────────────────────────────
	pageInputs := []struct {
		sectionNo int
		name      string
		path      string
		order     int
	}{
		{sectionNos[0], "Sections",  "section",   1},
		{sectionNos[0], "Pages",     "pages",     2},
		{sectionNos[1], "Roles",     "roles",     1},
		{sectionNos[1], "Employees", "employees", 2},
	}

	for _, p := range pageInputs {
		_, err = tx.Exec(ctx, `
			INSERT INTO master.pages (page_id, section_no, page_name, page_path, page_order, status)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, uuid.New(), p.sectionNo, p.name, p.path, p.order, "A")
		if err != nil {
			return fmt.Errorf("seedData: insert page [%s]: %w", p.name, err)
		}
	}

	// ── Admin User ────────────────────────────────────────────────────────────
	var profileNo int
	err = tx.QueryRow(ctx, `
		INSERT INTO master.users (profile_id, role_no, user_fullname, email_id, mobile_no, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING profile_no
	`, uuid.New(), roleNo, "Chand Kumar Magar", "chand.magar@gmail.com", "9804590230", "A").Scan(&profileNo)
	if err != nil {
		return fmt.Errorf("seedData: insert admin user: %w", err)
	}

	// ── Admin Credentials ─────────────────────────────────────────────────────
	hashed, err := bcrypt.GenerateFromPassword([]byte("Admin@123"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("seedData: hash password: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO master.credentials (credential_id, profile_no, username, password, status)
		VALUES ($1, $2, $3, $4, $5)
	`, uuid.New(), profileNo, "chand.magar", string(hashed), "A")
	if err != nil {
		return fmt.Errorf("seedData: insert credentials: %w", err)
	}

	// ── Commit ────────────────────────────────────────────────────────────────
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("seedData: commit: %w", err)
	}

	log.Println("seed data inserted successfully")
	return nil
}