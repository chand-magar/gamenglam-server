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
		{"create schema",    createSchema},
		{"create enums",     createEnums},
		{"create tables",    createTables},
		{"add foreign keys", addForeignKeys},
		{"seed data",        seedData},
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
	queries := []string{
		`CREATE SCHEMA IF NOT EXISTS auth`,
		`CREATE SCHEMA IF NOT EXISTS product`,
	}
	for _, q := range queries {
		if _, err := db.Exec(ctx, q); err != nil {
			return fmt.Errorf("createSchema: %w", err)
		}
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
				AND n.nspname = 'auth'
			) THEN
				CREATE TYPE auth.status_enum AS ENUM ('A', 'I', 'D');
			END IF;

			IF NOT EXISTS (
				SELECT 1 FROM pg_type t
				JOIN pg_namespace n ON n.oid = t.typnamespace
				WHERE t.typname = 'product_status_enum'
				AND n.nspname = 'product'
			) THEN
				CREATE TYPE product.product_status_enum AS ENUM ('draft', 'active', 'archived', 'D');
			END IF;

			IF NOT EXISTS (
				SELECT 1 FROM pg_type t
				JOIN pg_namespace n ON n.oid = t.typnamespace
				WHERE t.typname = 'attribute_type_enum'
				AND n.nspname = 'product'
			) THEN
				CREATE TYPE product.attribute_type_enum AS ENUM (
					'select', 'multi_select', 'boolean', 'text', 'number', 'color'
				);
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
		// ── product schema ────────────────────────────────────────────────────
		{
			name: "product.categories",
			query: `
				CREATE TABLE IF NOT EXISTS product.categories (
					category_no   SERIAL                       PRIMARY KEY,
					category_id   UUID                         NOT NULL DEFAULT gen_random_uuid(),
					parent_no     INTEGER                               DEFAULT NULL,
					category_name VARCHAR(255)                 NOT NULL,
					slug          VARCHAR(255)                 NOT NULL,
					details       TEXT                                  DEFAULT NULL,
					status        auth.status_enum           NOT NULL DEFAULT 'A',
					created_at    TIMESTAMPTZ                  NOT NULL DEFAULT NOW(),
					created_by    INTEGER                               DEFAULT NULL,
					updated_at    TIMESTAMPTZ                  NOT NULL DEFAULT NOW(),
					updated_by    INTEGER                               DEFAULT NULL
				);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_categories_category_id ON product.categories(category_id);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_categories_slug        ON product.categories(slug);
				CREATE        INDEX IF NOT EXISTS idx_categories_parent_no   ON product.categories(parent_no);
				CREATE        INDEX IF NOT EXISTS idx_categories_status      ON product.categories(status);
			`,
		},
		{
			name: "product.attributes",
			query: `
				CREATE TABLE IF NOT EXISTS product.attributes (
					attribute_no   SERIAL                           PRIMARY KEY,
					attribute_id   UUID                             NOT NULL DEFAULT gen_random_uuid(),
					category_no    INTEGER                          NOT NULL,
					attribute_name VARCHAR(100)                     NOT NULL,
					code           VARCHAR(50)                      NOT NULL,
					type           product.attribute_type_enum      NOT NULL,
					is_required    BOOLEAN                          NOT NULL DEFAULT FALSE,
					is_variant     BOOLEAN                          NOT NULL DEFAULT FALSE,
					sort_order     INTEGER                          NOT NULL DEFAULT 0,
					status         auth.status_enum               NOT NULL DEFAULT 'A',
					created_at     TIMESTAMPTZ                      NOT NULL DEFAULT NOW(),
					created_by     INTEGER                                   DEFAULT NULL,
					updated_at     TIMESTAMPTZ                      NOT NULL DEFAULT NOW(),
					updated_by     INTEGER                                   DEFAULT NULL
				);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_attributes_attribute_id ON product.attributes(attribute_id);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_attributes_code         ON product.attributes(code);
				CREATE        INDEX IF NOT EXISTS idx_attributes_category_no  ON product.attributes(category_no);
				CREATE        INDEX IF NOT EXISTS idx_attributes_is_variant   ON product.attributes(is_variant);
				CREATE        INDEX IF NOT EXISTS idx_attributes_status       ON product.attributes(status);
			`,
		},
		{
			name: "product.attribute_values",
			query: `
				CREATE TABLE IF NOT EXISTS product.attribute_values (
					attribute_value_no SERIAL                   PRIMARY KEY,
					attribute_value_id UUID                     NOT NULL DEFAULT gen_random_uuid(),
					attribute_no       INTEGER                  NOT NULL,
					value              VARCHAR(255)             NOT NULL,
					display_label      VARCHAR(255)             NOT NULL,
					color_hex          CHAR(7)                           DEFAULT NULL,
					sort_order         INTEGER                  NOT NULL DEFAULT 0,
					status             auth.status_enum       NOT NULL DEFAULT 'A',
					created_at         TIMESTAMPTZ              NOT NULL DEFAULT NOW(),
					created_by         INTEGER                           DEFAULT NULL,
					updated_at         TIMESTAMPTZ              NOT NULL DEFAULT NOW(),
					updated_by         INTEGER                           DEFAULT NULL
				);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_attribute_values_attribute_value_id ON product.attribute_values(attribute_value_id);
				CREATE        INDEX IF NOT EXISTS idx_attribute_values_attribute_no       ON product.attribute_values(attribute_no);
				CREATE        INDEX IF NOT EXISTS idx_attribute_values_status             ON product.attribute_values(status);
			`,
		},
		{
			name: "product.products",
			query: `
				CREATE TABLE IF NOT EXISTS product.products (
					product_no   SERIAL                            PRIMARY KEY,
					product_id   UUID                              NOT NULL DEFAULT gen_random_uuid(),
					category_no  INTEGER                           NOT NULL,
					sku          VARCHAR(100)                      NOT NULL,
					product_name VARCHAR(255)                      NOT NULL,
					slug         VARCHAR(255)                      NOT NULL,
					description  TEXT                                       DEFAULT NULL,
					base_price   NUMERIC(12,2)                     NOT NULL DEFAULT 0.00,
					status       product.product_status_enum       NOT NULL DEFAULT 'draft',
					is_featured  BOOLEAN                           NOT NULL DEFAULT FALSE,
					stock_qty    INTEGER                           NOT NULL DEFAULT 0,
					created_at   TIMESTAMPTZ                       NOT NULL DEFAULT NOW(),
					created_by   INTEGER                                    DEFAULT NULL,
					updated_at   TIMESTAMPTZ                       NOT NULL DEFAULT NOW(),
					updated_by   INTEGER                                    DEFAULT NULL
				);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_products_product_id ON product.products(product_id);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_products_sku        ON product.products(sku);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_products_slug       ON product.products(slug);
				CREATE        INDEX IF NOT EXISTS idx_products_category_no ON product.products(category_no);
				CREATE        INDEX IF NOT EXISTS idx_products_status      ON product.products(status);
				CREATE        INDEX IF NOT EXISTS idx_products_is_featured ON product.products(is_featured);
				CREATE        INDEX IF NOT EXISTS idx_products_created_at  ON product.products(created_at);
			`,
		},
		{
			name: "product.product_variants",
			query: `
				CREATE TABLE IF NOT EXISTS product.product_variants (
					variant_no     SERIAL                           PRIMARY KEY,
					variant_id     UUID                             NOT NULL DEFAULT gen_random_uuid(),
					product_no     INTEGER                          NOT NULL,
					sku            VARCHAR(100)                     NOT NULL,
					price_override NUMERIC(12,2)                             DEFAULT NULL,
					stock_qty      INTEGER                          NOT NULL DEFAULT 0,
					weight_kg      NUMERIC(8,3)                              DEFAULT NULL,
					image_url      VARCHAR(512)                              DEFAULT NULL,
					status         product.product_status_enum      NOT NULL DEFAULT 'draft',
					created_at     TIMESTAMPTZ                      NOT NULL DEFAULT NOW(),
					created_by     INTEGER                                   DEFAULT NULL,
					updated_at     TIMESTAMPTZ                      NOT NULL DEFAULT NOW(),
					updated_by     INTEGER                                   DEFAULT NULL
				);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_product_variants_variant_id ON product.product_variants(variant_id);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_product_variants_sku        ON product.product_variants(sku);
				CREATE        INDEX IF NOT EXISTS idx_product_variants_product_no ON product.product_variants(product_no);
				CREATE        INDEX IF NOT EXISTS idx_product_variants_status     ON product.product_variants(status);
			`,
		},
		{
			name: "product.variant_attribute_values",
			query: `
				CREATE TABLE IF NOT EXISTS product.variant_attribute_values (
					variant_attribute_value_no SERIAL    PRIMARY KEY,
					variant_attribute_value_id UUID      NOT NULL DEFAULT gen_random_uuid(),
					variant_no                 INTEGER   NOT NULL,
					attribute_value_no         INTEGER   NOT NULL,
					created_at                 TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					created_by                 INTEGER            DEFAULT NULL
				);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_vav_variant_attribute_value_id
					ON product.variant_attribute_values(variant_attribute_value_id);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_vav_variant_attribute_value_unique
					ON product.variant_attribute_values(variant_no, attribute_value_no);
				CREATE INDEX IF NOT EXISTS idx_vav_variant_no
					ON product.variant_attribute_values(variant_no);
				CREATE INDEX IF NOT EXISTS idx_vav_attribute_value_no
					ON product.variant_attribute_values(attribute_value_no);
			`,
		},

		// ── auth schema ─────────────────────────────────────────────────────
		{
			name: "roles",
			query: `
				CREATE TABLE IF NOT EXISTS auth.roles (
					role_no      SMALLSERIAL                PRIMARY KEY,
					role_id      UUID                       NOT NULL DEFAULT gen_random_uuid(),
					role_name    VARCHAR(65)                NOT NULL,
					role_details JSONB                      NOT NULL DEFAULT '[]',
					status       auth.status_enum         NOT NULL DEFAULT 'A',
					created_at   TIMESTAMPTZ                NOT NULL DEFAULT NOW(),
					created_by   INTEGER                             DEFAULT NULL,
					updated_at   TIMESTAMPTZ                NOT NULL DEFAULT NOW(),
					updated_by   INTEGER                             DEFAULT NULL
				);
				CREATE INDEX IF NOT EXISTS idx_roles_role_id ON auth.roles(role_id);
				CREATE INDEX IF NOT EXISTS idx_roles_status  ON auth.roles(status);
			`,
		},
		{
			name: "sections",
			query: `
				CREATE TABLE IF NOT EXISTS auth.sections (
					section_no    SMALLSERIAL                PRIMARY KEY,
					section_id    UUID                       NOT NULL DEFAULT gen_random_uuid(),
					section_name  VARCHAR(65)                NOT NULL,
					section_path  VARCHAR(65)                NOT NULL,
					section_icon  VARCHAR(65)                NOT NULL,
					section_order SMALLINT                   NOT NULL DEFAULT 0,
					status        auth.status_enum         NOT NULL DEFAULT 'A',
					created_at    TIMESTAMPTZ                NOT NULL DEFAULT NOW(),
					created_by    INTEGER                             DEFAULT NULL,
					updated_at    TIMESTAMPTZ                NOT NULL DEFAULT NOW(),
					updated_by    INTEGER                             DEFAULT NULL
				);
				CREATE INDEX IF NOT EXISTS idx_sections_section_id    ON auth.sections(section_id);
				CREATE INDEX IF NOT EXISTS idx_sections_section_order  ON auth.sections(section_order);
				CREATE INDEX IF NOT EXISTS idx_sections_status         ON auth.sections(status);
			`,
		},
		{
			name: "pages",
			query: `
				CREATE TABLE IF NOT EXISTS auth.pages (
					page_no    SMALLSERIAL                PRIMARY KEY,
					page_id    UUID                       NOT NULL DEFAULT gen_random_uuid(),
					section_no SMALLINT                   NOT NULL,
					page_name  VARCHAR(65)                NOT NULL,
					page_path  VARCHAR(255)               NOT NULL,
					page_order SMALLINT                   NOT NULL DEFAULT 0,
					status     auth.status_enum         NOT NULL DEFAULT 'A',
					created_at TIMESTAMPTZ                NOT NULL DEFAULT NOW(),
					created_by INTEGER                             DEFAULT NULL,
					updated_at TIMESTAMPTZ                NOT NULL DEFAULT NOW(),
					updated_by INTEGER                             DEFAULT NULL
				);
				CREATE INDEX IF NOT EXISTS idx_pages_page_id    ON auth.pages(page_id);
				CREATE INDEX IF NOT EXISTS idx_pages_section_no ON auth.pages(section_no);
				CREATE INDEX IF NOT EXISTS idx_pages_page_order ON auth.pages(page_order);
				CREATE INDEX IF NOT EXISTS idx_pages_status     ON auth.pages(status);
			`,
		},
		{
			name: "users",
			query: `
				CREATE TABLE IF NOT EXISTS auth.users (
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
					status        auth.status_enum         NOT NULL DEFAULT 'A',
					created_at    TIMESTAMPTZ                NOT NULL DEFAULT NOW(),
					created_by    INTEGER                             DEFAULT NULL,
					updated_at    TIMESTAMPTZ                NOT NULL DEFAULT NOW(),
					updated_by    INTEGER                             DEFAULT NULL
				);
				CREATE INDEX IF NOT EXISTS idx_users_profile_id ON auth.users(profile_id);
				CREATE INDEX IF NOT EXISTS idx_users_role_no    ON auth.users(role_no);
				CREATE INDEX IF NOT EXISTS idx_users_email_id   ON auth.users(email_id);
				CREATE INDEX IF NOT EXISTS idx_users_status     ON auth.users(status);
			`,
		},
		{
			name: "credentials",
			query: `
				CREATE TABLE IF NOT EXISTS auth.credentials (
					credential_no SERIAL                     PRIMARY KEY,
					credential_id UUID                       NOT NULL DEFAULT gen_random_uuid(),
					profile_no    INTEGER                    NOT NULL,
					username      VARCHAR(65)  UNIQUE        NOT NULL,
					password      VARCHAR(255)               NOT NULL,
					status        auth.status_enum         NOT NULL DEFAULT 'I',
					created_at    TIMESTAMPTZ                NOT NULL DEFAULT NOW(),
					created_by    INTEGER                             DEFAULT NULL,
					updated_at    TIMESTAMPTZ                NOT NULL DEFAULT NOW(),
					updated_by    INTEGER                             DEFAULT NULL
				);
				CREATE INDEX IF NOT EXISTS idx_credentials_credential_id ON auth.credentials(credential_id);
				CREATE INDEX IF NOT EXISTS idx_credentials_profile_no    ON auth.credentials(profile_no);
				CREATE INDEX IF NOT EXISTS idx_credentials_username      ON auth.credentials(username);
				CREATE INDEX IF NOT EXISTS idx_credentials_status        ON auth.credentials(status);
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
		// ── product schema ────────────────────────────────────────────────────
		{
			name: "fk_categories_parent_no",
			query: `
				DO $$ BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM information_schema.table_constraints
						WHERE constraint_name = 'fk_categories_parent_no'
					) THEN
						ALTER TABLE product.categories
						ADD CONSTRAINT fk_categories_parent_no
						FOREIGN KEY (parent_no) REFERENCES product.categories(category_no)
						ON UPDATE CASCADE ON DELETE SET NULL;
					END IF;
				END $$;
			`,
		},
		{
			name: "fk_attributes_category_no",
			query: `
				DO $$ BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM information_schema.table_constraints
						WHERE constraint_name = 'fk_attributes_category_no'
					) THEN
						ALTER TABLE product.attributes
						ADD CONSTRAINT fk_attributes_category_no
						FOREIGN KEY (category_no) REFERENCES product.categories(category_no)
						ON DELETE CASCADE;
					END IF;
				END $$;
			`,
		},
		{
			name: "fk_attribute_values_attribute_no",
			query: `
				DO $$ BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM information_schema.table_constraints
						WHERE constraint_name = 'fk_attribute_values_attribute_no'
					) THEN
						ALTER TABLE product.attribute_values
						ADD CONSTRAINT fk_attribute_values_attribute_no
						FOREIGN KEY (attribute_no) REFERENCES product.attributes(attribute_no)
						ON DELETE CASCADE;
					END IF;
				END $$;
			`,
		},
		{
			name: "fk_products_category_no",
			query: `
				DO $$ BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM information_schema.table_constraints
						WHERE constraint_name = 'fk_products_category_no'
					) THEN
						ALTER TABLE product.products
						ADD CONSTRAINT fk_products_category_no
						FOREIGN KEY (category_no) REFERENCES product.categories(category_no)
						ON DELETE RESTRICT;
					END IF;
				END $$;
			`,
		},
		{
			name: "fk_product_variants_product_no",
			query: `
				DO $$ BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM information_schema.table_constraints
						WHERE constraint_name = 'fk_product_variants_product_no'
					) THEN
						ALTER TABLE product.product_variants
						ADD CONSTRAINT fk_product_variants_product_no
						FOREIGN KEY (product_no) REFERENCES product.products(product_no)
						ON DELETE CASCADE;
					END IF;
				END $$;
			`,
		},
		{
			name: "fk_vav_variant_no",
			query: `
				DO $$ BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM information_schema.table_constraints
						WHERE constraint_name = 'fk_vav_variant_no'
					) THEN
						ALTER TABLE product.variant_attribute_values
						ADD CONSTRAINT fk_vav_variant_no
						FOREIGN KEY (variant_no) REFERENCES product.product_variants(variant_no)
						ON DELETE CASCADE;
					END IF;
				END $$;
			`,
		},
		{
			name: "fk_vav_attribute_value_no",
			query: `
				DO $$ BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM information_schema.table_constraints
						WHERE constraint_name = 'fk_vav_attribute_value_no'
					) THEN
						ALTER TABLE product.variant_attribute_values
						ADD CONSTRAINT fk_vav_attribute_value_no
						FOREIGN KEY (attribute_value_no) REFERENCES product.attribute_values(attribute_value_no)
						ON DELETE CASCADE;
					END IF;
				END $$;
			`,
		},

		// ── auth schema ─────────────────────────────────────────────────────
		{
			name: "fk_users_role_no",
			query: `
				DO $$ BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM information_schema.table_constraints
						WHERE constraint_name = 'fk_users_role_no'
					) THEN
						ALTER TABLE auth.users
						ADD CONSTRAINT fk_users_role_no
						FOREIGN KEY (role_no) REFERENCES auth.roles(role_no)
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
						ALTER TABLE auth.credentials
						ADD CONSTRAINT fk_credentials_profile_no
						FOREIGN KEY (profile_no) REFERENCES auth.users(profile_no)
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
						ALTER TABLE auth.pages
						ADD CONSTRAINT fk_pages_section_no
						FOREIGN KEY (section_no) REFERENCES auth.sections(section_no)
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
	if err := db.QueryRow(ctx, `SELECT COUNT(*) FROM auth.users`).Scan(&count); err != nil {
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
		INSERT INTO auth.roles (role_id, role_name, role_details, status)
		VALUES ($1, $2, $3, $4)
		RETURNING role_no
	`, uuid.New(), "Super Admin", "[]", "A").Scan(&roleNo)
	if err != nil {
		return fmt.Errorf("seedData: insert role: %w", err)
	}

	// ── Sections ──────────────────────────────────────────────────────────────
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
			INSERT INTO auth.sections (section_id, section_name, section_path, section_icon, section_order, status)
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
			INSERT INTO auth.pages (page_id, section_no, page_name, page_path, page_order, status)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, uuid.New(), p.sectionNo, p.name, p.path, p.order, "A")
		if err != nil {
			return fmt.Errorf("seedData: insert page [%s]: %w", p.name, err)
		}
	}

	// ── Admin User ────────────────────────────────────────────────────────────
	var profileNo int
	err = tx.QueryRow(ctx, `
		INSERT INTO auth.users (profile_id, role_no, user_fullname, email_id, mobile_no, status)
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
		INSERT INTO auth.credentials (credential_id, profile_no, username, password, status)
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