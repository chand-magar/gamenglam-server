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
		{"create schema", createSchema},
		{"create enums", createEnums},
		{"create tables", createTables},
		{"add foreign keys", addForeignKeys},
		{"seed data", seedData},
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
		`CREATE SCHEMA IF NOT EXISTS catalog`,
		`CREATE SCHEMA IF NOT EXISTS customer`,
		`CREATE SCHEMA IF NOT EXISTS orders`,
		`CREATE SCHEMA IF NOT EXISTS payment`,
		`CREATE SCHEMA IF NOT EXISTS delivery`,
		`CREATE SCHEMA IF NOT EXISTS notification`,
		`CREATE SCHEMA IF NOT EXISTS analytics`,
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
			-- auth
			IF NOT EXISTS (
				SELECT 1 FROM pg_type t
				JOIN pg_namespace n ON n.oid = t.typnamespace
				WHERE t.typname = 'status_enum' AND n.nspname = 'auth'
			) THEN
				CREATE TYPE auth.status_enum AS ENUM ('A', 'I', 'D');
			END IF;

			-- catalog
			IF NOT EXISTS (
				SELECT 1 FROM pg_type t
				JOIN pg_namespace n ON n.oid = t.typnamespace
				WHERE t.typname = 'product_status_enum' AND n.nspname = 'catalog'
			) THEN
				CREATE TYPE catalog.product_status_enum AS ENUM ('draft', 'active', 'archived', 'D');
			END IF;

			IF NOT EXISTS (
				SELECT 1 FROM pg_type t
				JOIN pg_namespace n ON n.oid = t.typnamespace
				WHERE t.typname = 'attribute_type_enum' AND n.nspname = 'catalog'
			) THEN
				CREATE TYPE catalog.attribute_type_enum AS ENUM (
					'select', 'multi_select', 'boolean', 'text', 'number', 'color'
				);
			END IF;

			-- customer
			IF NOT EXISTS (
				SELECT 1 FROM pg_type t
				JOIN pg_namespace n ON n.oid = t.typnamespace
				WHERE t.typname = 'gender_enum' AND n.nspname = 'customer'
			) THEN
				CREATE TYPE customer.gender_enum AS ENUM (
					'male', 'female', 'other', 'prefer_not_to_say'
				);
			END IF;

			-- orders
			IF NOT EXISTS (
				SELECT 1 FROM pg_type t
				JOIN pg_namespace n ON n.oid = t.typnamespace
				WHERE t.typname = 'order_status_enum' AND n.nspname = 'orders'
			) THEN
				CREATE TYPE orders.order_status_enum AS ENUM (
					'pending', 'confirmed', 'processing',
					'shipped', 'delivered', 'cancelled', 'refunded'
				);
			END IF;

			-- payment
			IF NOT EXISTS (
				SELECT 1 FROM pg_type t
				JOIN pg_namespace n ON n.oid = t.typnamespace
				WHERE t.typname = 'payment_status_enum' AND n.nspname = 'payment'
			) THEN
				CREATE TYPE payment.payment_status_enum AS ENUM (
					'pending', 'success', 'failed', 'cancelled'
				);
			END IF;

			IF NOT EXISTS (
				SELECT 1 FROM pg_type t
				JOIN pg_namespace n ON n.oid = t.typnamespace
				WHERE t.typname = 'refund_status_enum' AND n.nspname = 'payment'
			) THEN
				CREATE TYPE payment.refund_status_enum AS ENUM (
					'requested', 'processing', 'completed', 'rejected'
				);
			END IF;

			-- delivery
			IF NOT EXISTS (
				SELECT 1 FROM pg_type t
				JOIN pg_namespace n ON n.oid = t.typnamespace
				WHERE t.typname = 'shipment_status_enum' AND n.nspname = 'delivery'
			) THEN
				CREATE TYPE delivery.shipment_status_enum AS ENUM (
					'pending', 'picked_up', 'in_transit',
					'out_for_delivery', 'delivered', 'failed', 'returned'
				);
			END IF;

			-- notification
			IF NOT EXISTS (
				SELECT 1 FROM pg_type t
				JOIN pg_namespace n ON n.oid = t.typnamespace
				WHERE t.typname = 'notification_channel_enum' AND n.nspname = 'notification'
			) THEN
				CREATE TYPE notification.notification_channel_enum AS ENUM (
					'email', 'sms', 'push'
				);
			END IF;

			IF NOT EXISTS (
				SELECT 1 FROM pg_type t
				JOIN pg_namespace n ON n.oid = t.typnamespace
				WHERE t.typname = 'notification_status_enum' AND n.nspname = 'notification'
			) THEN
				CREATE TYPE notification.notification_status_enum AS ENUM (
					'pending', 'sent', 'failed', 'skipped'
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

		// ── auth schema ───────────────────────────────────────────────────────
		{
			name: "auth.roles",
			query: `
				CREATE TABLE IF NOT EXISTS auth.roles (
					role_no      SMALLSERIAL              PRIMARY KEY,
					role_id      UUID                     NOT NULL DEFAULT gen_random_uuid(),
					role_name    VARCHAR(65)              NOT NULL,
					role_details JSONB                    NOT NULL DEFAULT '[]',
					status       auth.status_enum         NOT NULL DEFAULT 'A',
					created_at   TIMESTAMPTZ              NOT NULL DEFAULT NOW(),
					created_by   INTEGER                           DEFAULT NULL,
					updated_at   TIMESTAMPTZ              NOT NULL DEFAULT NOW(),
					updated_by   INTEGER                           DEFAULT NULL
				);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_role_id ON auth.roles(role_id);
				CREATE        INDEX IF NOT EXISTS idx_roles_status  ON auth.roles(status);
			`,
		},
		{
			name: "auth.sections",
			query: `
				CREATE TABLE IF NOT EXISTS auth.sections (
					section_no    SMALLSERIAL              PRIMARY KEY,
					section_id    UUID                     NOT NULL DEFAULT gen_random_uuid(),
					section_name  VARCHAR(65)              NOT NULL,
					section_path  VARCHAR(65)              NOT NULL,
					section_icon  VARCHAR(65)              NOT NULL,
					section_order SMALLINT                 NOT NULL DEFAULT 0,
					status        auth.status_enum         NOT NULL DEFAULT 'A',
					created_at    TIMESTAMPTZ              NOT NULL DEFAULT NOW(),
					created_by    INTEGER                           DEFAULT NULL,
					updated_at    TIMESTAMPTZ              NOT NULL DEFAULT NOW(),
					updated_by    INTEGER                           DEFAULT NULL
				);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_sections_section_id    ON auth.sections(section_id);
				CREATE        INDEX IF NOT EXISTS idx_sections_section_order ON auth.sections(section_order);
				CREATE        INDEX IF NOT EXISTS idx_sections_status        ON auth.sections(status);
			`,
		},
		{
			name: "auth.pages",
			query: `
				CREATE TABLE IF NOT EXISTS auth.pages (
					page_no    SMALLSERIAL              PRIMARY KEY,
					page_id    UUID                     NOT NULL DEFAULT gen_random_uuid(),
					section_no SMALLINT                 NOT NULL,
					page_name  VARCHAR(65)              NOT NULL,
					page_path  VARCHAR(255)             NOT NULL,
					page_order SMALLINT                 NOT NULL DEFAULT 0,
					status     auth.status_enum         NOT NULL DEFAULT 'A',
					created_at TIMESTAMPTZ              NOT NULL DEFAULT NOW(),
					created_by INTEGER                           DEFAULT NULL,
					updated_at TIMESTAMPTZ              NOT NULL DEFAULT NOW(),
					updated_by INTEGER                           DEFAULT NULL
				);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_pages_page_id    ON auth.pages(page_id);
				CREATE        INDEX IF NOT EXISTS idx_pages_section_no ON auth.pages(section_no);
				CREATE        INDEX IF NOT EXISTS idx_pages_page_order ON auth.pages(page_order);
				CREATE        INDEX IF NOT EXISTS idx_pages_status     ON auth.pages(status);
			`,
		},
		{
			name: "auth.users",
			query: `
				CREATE TABLE IF NOT EXISTS auth.users (
					profile_no    SERIAL                   PRIMARY KEY,
					profile_id    UUID                     NOT NULL DEFAULT gen_random_uuid(),
					role_no       SMALLINT                          DEFAULT NULL,
					user_fullname VARCHAR(65)              NOT NULL,
					email_id      VARCHAR(65)  UNIQUE      NOT NULL,
					gender        VARCHAR(10)                       DEFAULT NULL,
					dob           DATE                              DEFAULT NULL,
					mobile_no     VARCHAR(15)                       DEFAULT NULL,
					address       JSONB                    NOT NULL DEFAULT '{}',
					x_api_key     VARCHAR(55)                       DEFAULT NULL,
					secret_key    VARCHAR(55)                       DEFAULT NULL,
					status        auth.status_enum         NOT NULL DEFAULT 'A',
					created_at    TIMESTAMPTZ              NOT NULL DEFAULT NOW(),
					created_by    INTEGER                           DEFAULT NULL,
					updated_at    TIMESTAMPTZ              NOT NULL DEFAULT NOW(),
					updated_by    INTEGER                           DEFAULT NULL
				);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_users_profile_id ON auth.users(profile_id);
				CREATE        INDEX IF NOT EXISTS idx_users_role_no    ON auth.users(role_no);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_id   ON auth.users(email_id);
				CREATE        INDEX IF NOT EXISTS idx_users_status     ON auth.users(status);
			`,
		},
		{
			name: "auth.credentials",
			query: `
				CREATE TABLE IF NOT EXISTS auth.credentials (
					credential_no SERIAL                   PRIMARY KEY,
					credential_id UUID                     NOT NULL DEFAULT gen_random_uuid(),
					profile_no    INTEGER                  NOT NULL,
					username      VARCHAR(65)  UNIQUE      NOT NULL,
					password      VARCHAR(255)             NOT NULL,
					status        auth.status_enum         NOT NULL DEFAULT 'I',
					created_at    TIMESTAMPTZ              NOT NULL DEFAULT NOW(),
					created_by    INTEGER                           DEFAULT NULL,
					updated_at    TIMESTAMPTZ              NOT NULL DEFAULT NOW(),
					updated_by    INTEGER                           DEFAULT NULL
				);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_credentials_credential_id ON auth.credentials(credential_id);
				CREATE        INDEX IF NOT EXISTS idx_credentials_profile_no    ON auth.credentials(profile_no);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_credentials_username      ON auth.credentials(username);
				CREATE        INDEX IF NOT EXISTS idx_credentials_status        ON auth.credentials(status);
			`,
		},

		// ── catalog schema ────────────────────────────────────────────────────
		{
			name: "catalog.categories",
			query: `
				CREATE TABLE IF NOT EXISTS catalog.categories (
					category_no   SERIAL                   PRIMARY KEY,
					category_id   UUID                     NOT NULL DEFAULT gen_random_uuid(),
					parent_no     INTEGER                           DEFAULT NULL,
					category_name VARCHAR(255)             NOT NULL,
					slug          VARCHAR(255)             NOT NULL,
					details       TEXT                              DEFAULT NULL,
					status        auth.status_enum         NOT NULL DEFAULT 'A',
					created_at    TIMESTAMPTZ              NOT NULL DEFAULT NOW(),
					created_by    INTEGER                           DEFAULT NULL,
					updated_at    TIMESTAMPTZ              NOT NULL DEFAULT NOW(),
					updated_by    INTEGER                           DEFAULT NULL
				);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_categories_category_id ON catalog.categories(category_id);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_categories_slug        ON catalog.categories(slug);
				CREATE        INDEX IF NOT EXISTS idx_categories_parent_no   ON catalog.categories(parent_no);
				CREATE        INDEX IF NOT EXISTS idx_categories_status      ON catalog.categories(status);
			`,
		},
		{
			name: "catalog.attributes",
			query: `
				CREATE TABLE IF NOT EXISTS catalog.attributes (
					attribute_no   SERIAL                        PRIMARY KEY,
					attribute_id   UUID                          NOT NULL DEFAULT gen_random_uuid(),
					category_no    INTEGER                       NOT NULL,
					attribute_name VARCHAR(100)                  NOT NULL,
					code           VARCHAR(50)                   NOT NULL,
					type           catalog.attribute_type_enum   NOT NULL,
					is_required    BOOLEAN                       NOT NULL DEFAULT FALSE,
					is_variant     BOOLEAN                       NOT NULL DEFAULT FALSE,
					show_in_filter BOOLEAN 						 NOT NULL DEFAULT FALSE,
					sort_order     INTEGER                       NOT NULL DEFAULT 0,
					status         auth.status_enum              NOT NULL DEFAULT 'A',
					created_at     TIMESTAMPTZ                   NOT NULL DEFAULT NOW(),
					created_by     INTEGER                                DEFAULT NULL,
					updated_at     TIMESTAMPTZ                   NOT NULL DEFAULT NOW(),
					updated_by     INTEGER                                DEFAULT NULL
				);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_attributes_attribute_id ON catalog.attributes(attribute_id);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_attributes_code         ON catalog.attributes(code);
				CREATE        INDEX IF NOT EXISTS idx_attributes_category_no  ON catalog.attributes(category_no);
				CREATE        INDEX IF NOT EXISTS idx_attributes_is_variant   ON catalog.attributes(is_variant);
				CREATE        INDEX IF NOT EXISTS idx_attributes_status       ON catalog.attributes(status);
			`,
		},
		{
			name: "catalog.attribute_values",
			query: `
				CREATE TABLE IF NOT EXISTS catalog.attribute_values (
					attribute_value_no SERIAL                PRIMARY KEY,
					attribute_value_id UUID                  NOT NULL DEFAULT gen_random_uuid(),
					attribute_no       INTEGER               NOT NULL,
					value              VARCHAR(255)          NOT NULL,
					display_label      VARCHAR(255)          NOT NULL,
					color_hex          CHAR(7)                        DEFAULT NULL,
					sort_order         INTEGER               NOT NULL DEFAULT 0,
					status             auth.status_enum      NOT NULL DEFAULT 'A',
					created_at         TIMESTAMPTZ           NOT NULL DEFAULT NOW(),
					created_by         INTEGER                        DEFAULT NULL,
					updated_at         TIMESTAMPTZ           NOT NULL DEFAULT NOW(),
					updated_by         INTEGER                        DEFAULT NULL
				);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_attribute_values_id     ON catalog.attribute_values(attribute_value_id);
				CREATE        INDEX IF NOT EXISTS idx_attribute_values_attr    ON catalog.attribute_values(attribute_no);
				CREATE        INDEX IF NOT EXISTS idx_attribute_values_status  ON catalog.attribute_values(status);
			`,
		},
		{
			name: "catalog.products",
			query: `
				CREATE TABLE IF NOT EXISTS catalog.products (
					product_no     SERIAL                          PRIMARY KEY,
					product_id     UUID                            NOT NULL DEFAULT gen_random_uuid(),
					category_no    INTEGER                         NOT NULL,
					sku            VARCHAR(100)                    NOT NULL,
					product_name   VARCHAR(255)                    NOT NULL,
					slug           VARCHAR(255)                    NOT NULL,
					description    TEXT                                     DEFAULT NULL,
					image_url      VARCHAR(512)                             DEFAULT NULL,
					features       JSONB NOT NULL                            DEFAULT '[]',
					specifications JSONB NOT NULL                            DEFAULT '[]',
					status         catalog.product_status_enum     NOT NULL DEFAULT 'draft',
					is_featured    BOOLEAN                         NOT NULL DEFAULT FALSE,
					created_at     TIMESTAMPTZ                     NOT NULL DEFAULT NOW(),
					created_by     INTEGER                                  DEFAULT NULL,
					updated_at     TIMESTAMPTZ                     NOT NULL DEFAULT NOW(),
					updated_by     INTEGER                                  DEFAULT NULL
				);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_products_product_id  ON catalog.products(product_id);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_products_sku         ON catalog.products(sku);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_products_slug        ON catalog.products(slug);
				CREATE        INDEX IF NOT EXISTS idx_products_category_no ON catalog.products(category_no);
				CREATE        INDEX IF NOT EXISTS idx_products_status      ON catalog.products(status);
				CREATE        INDEX IF NOT EXISTS idx_products_is_featured ON catalog.products(is_featured);
				CREATE        INDEX IF NOT EXISTS idx_products_created_at  ON catalog.products(created_at);
			`,
		},
		{
			name: "catalog.product_variants",
			query: `
				CREATE TABLE IF NOT EXISTS catalog.product_variants (
					variant_no     SERIAL                        PRIMARY KEY,
					variant_id     UUID                          NOT NULL DEFAULT gen_random_uuid(),
					product_no     INTEGER                       NOT NULL,
					sku            VARCHAR(100)                  NOT NULL,
					cost_price     NUMERIC(10,2) 						  DEFAULT NULL,
					mrp            NUMERIC(10,2)                          DEFAULT NULL,
					selling_price  NUMERIC(10,2) 						  DEFAULT NULL,
					stock_qty      INTEGER                       NOT NULL DEFAULT 0,
					weight_kg      NUMERIC(8,3)                           DEFAULT NULL,
					image_url      VARCHAR(512)                           DEFAULT NULL,
					status         catalog.product_status_enum   NOT NULL DEFAULT 'draft',
					created_at     TIMESTAMPTZ                   NOT NULL DEFAULT NOW(),
					created_by     INTEGER                                DEFAULT NULL,
					updated_at     TIMESTAMPTZ                   NOT NULL DEFAULT NOW(),
					updated_by     INTEGER                                DEFAULT NULL
				);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_product_variants_id         ON catalog.product_variants(variant_id);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_product_variants_sku        ON catalog.product_variants(sku);
				CREATE        INDEX IF NOT EXISTS idx_product_variants_product_no ON catalog.product_variants(product_no);
				CREATE        INDEX IF NOT EXISTS idx_product_variants_status     ON catalog.product_variants(status);
			`,
		},
		{
			name: "catalog.variant_attribute_values",
			query: `
				CREATE TABLE IF NOT EXISTS catalog.variant_attribute_values (
					variant_attribute_value_no SERIAL      PRIMARY KEY,
					variant_attribute_value_id UUID        NOT NULL DEFAULT gen_random_uuid(),
					variant_no                 INTEGER     NOT NULL,
					attribute_value_no         INTEGER     NOT NULL,
					created_at                 TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					created_by                 INTEGER              DEFAULT NULL
				);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_vav_id
					ON catalog.variant_attribute_values(variant_attribute_value_id);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_vav_unique
					ON catalog.variant_attribute_values(variant_no, attribute_value_no);
				CREATE        INDEX IF NOT EXISTS idx_vav_variant_no
					ON catalog.variant_attribute_values(variant_no);
				CREATE        INDEX IF NOT EXISTS idx_vav_attribute_value_no
					ON catalog.variant_attribute_values(attribute_value_no);
			`,
		},
		{
			name: "catalog.product_variant_images",
			query: `
				CREATE TABLE IF NOT EXISTS catalog.product_variant_images (
					image_no   SERIAL        PRIMARY KEY,
					image_id   UUID          NOT NULL DEFAULT gen_random_uuid(),
					variant_no INTEGER       NOT NULL,
					image_url  TEXT          NOT NULL,
					sort_order INTEGER       NOT NULL DEFAULT 0,
					is_primary BOOLEAN       NOT NULL DEFAULT FALSE,
					created_at TIMESTAMPTZ   NOT NULL DEFAULT NOW()
				);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_variant_images_image_id
					ON catalog.product_variant_images(image_id);
				CREATE        INDEX IF NOT EXISTS idx_variant_images_variant_no
					ON catalog.product_variant_images(variant_no);
				CREATE        INDEX IF NOT EXISTS idx_variant_images_is_primary
					ON catalog.product_variant_images(is_primary);
			`,
		},

		// ── customer schema ───────────────────────────────────────────────────
		{
			name: "customer.customers",
			query: `
				CREATE TABLE IF NOT EXISTS customer.customers (
					customer_no SERIAL                      PRIMARY KEY,
					customer_id UUID                        NOT NULL DEFAULT gen_random_uuid(),
					first_name  VARCHAR(65)                 NOT NULL,
					last_name   VARCHAR(65)                 NOT NULL,
					email_id    VARCHAR(100) UNIQUE         NOT NULL,
					mobile_no   VARCHAR(15)                          DEFAULT NULL,
					gender      customer.gender_enum                  DEFAULT NULL,
					dob         DATE                                  DEFAULT NULL,
					status      auth.status_enum            NOT NULL DEFAULT 'A',
					created_at  TIMESTAMPTZ                 NOT NULL DEFAULT NOW(),
					updated_at  TIMESTAMPTZ                 NOT NULL DEFAULT NOW()
				);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_customers_customer_id ON customer.customers(customer_id);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_customers_email_id    ON customer.customers(email_id);
				CREATE        INDEX IF NOT EXISTS idx_customers_status      ON customer.customers(status);
			`,
		},
		{
			name: "customer.credentials",
			query: `
				CREATE TABLE IF NOT EXISTS customer.credentials (
					credential_no SERIAL                    PRIMARY KEY,
					credential_id UUID                      NOT NULL DEFAULT gen_random_uuid(),
					customer_no   INTEGER                   NOT NULL,
					username      VARCHAR(65)  UNIQUE       NOT NULL,
					password      VARCHAR(255)              NOT NULL,
					last_login_at TIMESTAMPTZ                        DEFAULT NULL,
					status        auth.status_enum          NOT NULL DEFAULT 'A',
					created_at    TIMESTAMPTZ               NOT NULL DEFAULT NOW(),
					updated_at    TIMESTAMPTZ               NOT NULL DEFAULT NOW()
				);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_customer_creds_credential_id ON customer.credentials(credential_id);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_customer_creds_username      ON customer.credentials(username);
				CREATE        INDEX IF NOT EXISTS idx_customer_creds_customer_no   ON customer.credentials(customer_no);
				CREATE        INDEX IF NOT EXISTS idx_customer_creds_status        ON customer.credentials(status);
			`,
		},
		{
			name: "customer.addresses",
			query: `
				CREATE TABLE IF NOT EXISTS customer.addresses (
					address_no    SERIAL                   PRIMARY KEY,
					address_id    UUID                     NOT NULL DEFAULT gen_random_uuid(),
					customer_no   INTEGER                  NOT NULL,
					label         VARCHAR(65)              NOT NULL DEFAULT 'Home',
					address_line1 VARCHAR(255)             NOT NULL,
					address_line2 VARCHAR(255)                      DEFAULT NULL,
					city          VARCHAR(100)             NOT NULL,
					state         VARCHAR(100)             NOT NULL,
					country       VARCHAR(100)             NOT NULL DEFAULT 'Nepal',
					postal_code   VARCHAR(20)                       DEFAULT NULL,
					is_default    BOOLEAN                  NOT NULL DEFAULT FALSE,
					status        auth.status_enum         NOT NULL DEFAULT 'A',
					created_at    TIMESTAMPTZ              NOT NULL DEFAULT NOW(),
					updated_at    TIMESTAMPTZ              NOT NULL DEFAULT NOW()
				);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_addresses_address_id  ON customer.addresses(address_id);
				CREATE        INDEX IF NOT EXISTS idx_addresses_customer_no  ON customer.addresses(customer_no);
				CREATE        INDEX IF NOT EXISTS idx_addresses_is_default   ON customer.addresses(is_default);
				CREATE        INDEX IF NOT EXISTS idx_addresses_status       ON customer.addresses(status);
			`,
		},
		{
			name: "customer.wishlists",
			query: `
				CREATE TABLE IF NOT EXISTS customer.wishlists (
					wishlist_no SERIAL      PRIMARY KEY,
					wishlist_id UUID        NOT NULL DEFAULT gen_random_uuid(),
					customer_no INTEGER     NOT NULL,
					product_no  INTEGER     NOT NULL,
					created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
				);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_wishlists_wishlist_id      ON customer.wishlists(wishlist_id);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_wishlists_customer_product ON customer.wishlists(customer_no, product_no);
				CREATE        INDEX IF NOT EXISTS idx_wishlists_customer_no      ON customer.wishlists(customer_no);
			`,
		},

		// ── orders schema ─────────────────────────────────────────────────────
		{
			name: "orders.carts",
			query: `
				CREATE TABLE IF NOT EXISTS orders.carts (
					cart_no     SERIAL      PRIMARY KEY,
					cart_id     UUID        NOT NULL DEFAULT gen_random_uuid(),
					customer_no INTEGER     NOT NULL,
					created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
				);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_carts_cart_id     ON orders.carts(cart_id);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_carts_customer_no ON orders.carts(customer_no);
			`,
		},
		{
			name: "orders.cart_items",
			query: `
				CREATE TABLE IF NOT EXISTS orders.cart_items (
					cart_item_no SERIAL        PRIMARY KEY,
					cart_item_id UUID          NOT NULL DEFAULT gen_random_uuid(),
					cart_no      INTEGER       NOT NULL,
					variant_no   INTEGER       NOT NULL,
					quantity     INTEGER       NOT NULL DEFAULT 1,
					unit_price   NUMERIC(12,2) NOT NULL,
					created_at   TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
					updated_at   TIMESTAMPTZ   NOT NULL DEFAULT NOW()
				);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_cart_items_cart_item_id  ON orders.cart_items(cart_item_id);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_cart_items_cart_variant   ON orders.cart_items(cart_no, variant_no);
				CREATE        INDEX IF NOT EXISTS idx_cart_items_cart_no       ON orders.cart_items(cart_no);
			`,
		},
		{
			name: "orders.orders",
			query: `
				CREATE TABLE IF NOT EXISTS orders.orders (
					order_no     SERIAL                      PRIMARY KEY,
					order_id     UUID                        NOT NULL DEFAULT gen_random_uuid(),
					customer_no  INTEGER                     NOT NULL,
					address_no   INTEGER                     NOT NULL,
					order_ref    VARCHAR(50)                 NOT NULL,
					subtotal     NUMERIC(12,2)               NOT NULL DEFAULT 0.00,
					discount     NUMERIC(12,2)               NOT NULL DEFAULT 0.00,
					tax          NUMERIC(12,2)               NOT NULL DEFAULT 0.00,
					delivery_fee NUMERIC(12,2)               NOT NULL DEFAULT 0.00,
					total        NUMERIC(12,2)               NOT NULL DEFAULT 0.00,
					order_status orders.order_status_enum    NOT NULL DEFAULT 'pending',
					notes        TEXT                                 DEFAULT NULL,
					created_at   TIMESTAMPTZ                 NOT NULL DEFAULT NOW(),
					updated_at   TIMESTAMPTZ                 NOT NULL DEFAULT NOW()
				);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_orders_order_id    ON orders.orders(order_id);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_orders_order_ref   ON orders.orders(order_ref);
				CREATE        INDEX IF NOT EXISTS idx_orders_customer_no ON orders.orders(customer_no);
				CREATE        INDEX IF NOT EXISTS idx_orders_status      ON orders.orders(order_status);
				CREATE        INDEX IF NOT EXISTS idx_orders_created_at  ON orders.orders(created_at);
			`,
		},
		{
			name: "orders.order_items",
			query: `
				CREATE TABLE IF NOT EXISTS orders.order_items (
					order_item_no SERIAL        PRIMARY KEY,
					order_item_id UUID          NOT NULL DEFAULT gen_random_uuid(),
					order_no      INTEGER       NOT NULL,
					variant_no    INTEGER       NOT NULL,
					product_name  VARCHAR(255)  NOT NULL,
					sku           VARCHAR(100)  NOT NULL,
					attributes    JSONB         NOT NULL DEFAULT '{}',
					quantity      INTEGER       NOT NULL,
					unit_price    NUMERIC(12,2) NOT NULL,
					total_price   NUMERIC(12,2) NOT NULL,
					created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW()
				);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_order_items_order_item_id ON orders.order_items(order_item_id);
				CREATE        INDEX IF NOT EXISTS idx_order_items_order_no      ON orders.order_items(order_no);
				CREATE        INDEX IF NOT EXISTS idx_order_items_variant_no    ON orders.order_items(variant_no);
			`,
		},
		{
			name: "orders.order_status_history",
			query: `
				CREATE TABLE IF NOT EXISTS orders.order_status_history (
					history_no  SERIAL                     PRIMARY KEY,
					history_id  UUID                       NOT NULL DEFAULT gen_random_uuid(),
					order_no    INTEGER                    NOT NULL,
					from_status orders.order_status_enum            DEFAULT NULL,
					to_status   orders.order_status_enum   NOT NULL,
					changed_by  INTEGER                             DEFAULT NULL,
					note        TEXT                                DEFAULT NULL,
					changed_at  TIMESTAMPTZ                NOT NULL DEFAULT NOW()
				);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_order_status_history_id       ON orders.order_status_history(history_id);
				CREATE        INDEX IF NOT EXISTS idx_order_status_history_order_no ON orders.order_status_history(order_no);
				CREATE        INDEX IF NOT EXISTS idx_order_status_history_changed  ON orders.order_status_history(changed_at);
			`,
		},

		// ── payment schema ────────────────────────────────────────────────────
		{
			name: "payment.payments",
			query: `
				CREATE TABLE IF NOT EXISTS payment.payments (
					payment_no     SERIAL                       PRIMARY KEY,
					payment_id     UUID                         NOT NULL DEFAULT gen_random_uuid(),
					order_no       INTEGER                      NOT NULL,
					customer_no    INTEGER                      NOT NULL,
					gateway        VARCHAR(50)                  NOT NULL,
					gateway_ref    VARCHAR(255)                          DEFAULT NULL,
					amount         NUMERIC(12,2)                NOT NULL,
					currency       VARCHAR(3)                   NOT NULL DEFAULT 'NPR',
					payment_method VARCHAR(50)                  NOT NULL,
					payment_status payment.payment_status_enum  NOT NULL DEFAULT 'pending',
					paid_at        TIMESTAMPTZ                           DEFAULT NULL,
					created_at     TIMESTAMPTZ                  NOT NULL DEFAULT NOW(),
					updated_at     TIMESTAMPTZ                  NOT NULL DEFAULT NOW()
				);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_payments_payment_id  ON payment.payments(payment_id);
				CREATE        INDEX IF NOT EXISTS idx_payments_order_no    ON payment.payments(order_no);
				CREATE        INDEX IF NOT EXISTS idx_payments_customer_no ON payment.payments(customer_no);
				CREATE        INDEX IF NOT EXISTS idx_payments_gateway_ref ON payment.payments(gateway_ref);
				CREATE        INDEX IF NOT EXISTS idx_payments_status      ON payment.payments(payment_status);
			`,
		},
		{
			name: "payment.refunds",
			query: `
				CREATE TABLE IF NOT EXISTS payment.refunds (
					refund_no     SERIAL                      PRIMARY KEY,
					refund_id     UUID                        NOT NULL DEFAULT gen_random_uuid(),
					payment_no    INTEGER                     NOT NULL,
					order_no      INTEGER                     NOT NULL,
					amount        NUMERIC(12,2)               NOT NULL,
					reason        TEXT                                 DEFAULT NULL,
					refund_status payment.refund_status_enum  NOT NULL DEFAULT 'requested',
					gateway_ref   VARCHAR(255)                         DEFAULT NULL,
					refunded_at   TIMESTAMPTZ                          DEFAULT NULL,
					created_at    TIMESTAMPTZ                 NOT NULL DEFAULT NOW(),
					updated_at    TIMESTAMPTZ                 NOT NULL DEFAULT NOW()
				);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_refunds_refund_id  ON payment.refunds(refund_id);
				CREATE        INDEX IF NOT EXISTS idx_refunds_payment_no ON payment.refunds(payment_no);
				CREATE        INDEX IF NOT EXISTS idx_refunds_order_no   ON payment.refunds(order_no);
				CREATE        INDEX IF NOT EXISTS idx_refunds_status     ON payment.refunds(refund_status);
			`,
		},

		// ── delivery schema ───────────────────────────────────────────────────
		{
			name: "delivery.shipments",
			query: `
				CREATE TABLE IF NOT EXISTS delivery.shipments (
					shipment_no     SERIAL                         PRIMARY KEY,
					shipment_id     UUID                           NOT NULL DEFAULT gen_random_uuid(),
					order_no        INTEGER                        NOT NULL,
					carrier         VARCHAR(100)                            DEFAULT NULL,
					tracking_no     VARCHAR(255)                            DEFAULT NULL,
					label_url       VARCHAR(512)                            DEFAULT NULL,
					estimated_date  DATE                                    DEFAULT NULL,
					shipment_status delivery.shipment_status_enum  NOT NULL DEFAULT 'pending',
					shipped_at      TIMESTAMPTZ                             DEFAULT NULL,
					delivered_at    TIMESTAMPTZ                             DEFAULT NULL,
					created_at      TIMESTAMPTZ                    NOT NULL DEFAULT NOW(),
					updated_at      TIMESTAMPTZ                    NOT NULL DEFAULT NOW()
				);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_shipments_shipment_id ON delivery.shipments(shipment_id);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_shipments_tracking_no ON delivery.shipments(tracking_no);
				CREATE        INDEX IF NOT EXISTS idx_shipments_order_no    ON delivery.shipments(order_no);
				CREATE        INDEX IF NOT EXISTS idx_shipments_status      ON delivery.shipments(shipment_status);
			`,
		},
		{
			name: "delivery.tracking_events",
			query: `
				CREATE TABLE IF NOT EXISTS delivery.tracking_events (
					event_no     SERIAL       PRIMARY KEY,
					event_id     UUID         NOT NULL DEFAULT gen_random_uuid(),
					shipment_no  INTEGER      NOT NULL,
					location     VARCHAR(255)          DEFAULT NULL,
					description  VARCHAR(512) NOT NULL,
					event_status VARCHAR(100) NOT NULL,
					event_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
				);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_tracking_events_event_id    ON delivery.tracking_events(event_id);
				CREATE        INDEX IF NOT EXISTS idx_tracking_events_shipment_no ON delivery.tracking_events(shipment_no);
				CREATE        INDEX IF NOT EXISTS idx_tracking_events_event_at    ON delivery.tracking_events(event_at);
			`,
		},

		// ── notification schema ───────────────────────────────────────────────
		{
			name: "notification.templates",
			query: `
				CREATE TABLE IF NOT EXISTS notification.templates (
					template_no  SERIAL                                 PRIMARY KEY,
					template_id  UUID                                   NOT NULL DEFAULT gen_random_uuid(),
					template_key VARCHAR(100)                           NOT NULL,
					channel      notification.notification_channel_enum NOT NULL,
					subject      VARCHAR(255)                                    DEFAULT NULL,
					body         TEXT                                   NOT NULL,
					status       auth.status_enum                       NOT NULL DEFAULT 'A',
					created_at   TIMESTAMPTZ                            NOT NULL DEFAULT NOW(),
					created_by   INTEGER                                         DEFAULT NULL,
					updated_at   TIMESTAMPTZ                            NOT NULL DEFAULT NOW(),
					updated_by   INTEGER                                         DEFAULT NULL
				);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_templates_template_id ON notification.templates(template_id);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_templates_key_channel ON notification.templates(template_key, channel);
				CREATE        INDEX IF NOT EXISTS idx_templates_status      ON notification.templates(status);
			`,
		},
		{
			name: "notification.notifications",
			query: `
				CREATE TABLE IF NOT EXISTS notification.notifications (
					notification_no SERIAL                                  PRIMARY KEY,
					notification_id UUID                                    NOT NULL DEFAULT gen_random_uuid(),
					customer_no     INTEGER                                          DEFAULT NULL,
					template_no     INTEGER                                          DEFAULT NULL,
					channel         notification.notification_channel_enum  NOT NULL,
					recipient       VARCHAR(255)                            NOT NULL,
					payload         JSONB                                   NOT NULL DEFAULT '{}',
					sent_status     notification.notification_status_enum   NOT NULL DEFAULT 'pending',
					sent_at         TIMESTAMPTZ                                      DEFAULT NULL,
					error_message   TEXT                                             DEFAULT NULL,
					created_at      TIMESTAMPTZ                             NOT NULL DEFAULT NOW()
				);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_notifications_id          ON notification.notifications(notification_id);
				CREATE        INDEX IF NOT EXISTS idx_notifications_customer_no ON notification.notifications(customer_no);
				CREATE        INDEX IF NOT EXISTS idx_notifications_status      ON notification.notifications(sent_status);
				CREATE        INDEX IF NOT EXISTS idx_notifications_created_at  ON notification.notifications(created_at);
			`,
		},

		// ── analytics schema ──────────────────────────────────────────────────
		{
			name: "analytics.daily_sales",
			query: `
				CREATE TABLE IF NOT EXISTS analytics.daily_sales (
					id            SERIAL        PRIMARY KEY,
					date          DATE          NOT NULL,
					total_orders  INTEGER       NOT NULL DEFAULT 0,
					total_revenue NUMERIC(14,2) NOT NULL DEFAULT 0.00,
					total_refunds NUMERIC(14,2) NOT NULL DEFAULT 0.00,
					net_revenue   NUMERIC(14,2) NOT NULL DEFAULT 0.00,
					created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW()
				);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_daily_sales_date ON analytics.daily_sales(date);
			`,
		},
		{
			name: "analytics.product_performance",
			query: `
				CREATE TABLE IF NOT EXISTS analytics.product_performance (
					id          SERIAL        PRIMARY KEY,
					product_no  INTEGER       NOT NULL,
					date        DATE          NOT NULL,
					views       INTEGER       NOT NULL DEFAULT 0,
					add_to_cart INTEGER       NOT NULL DEFAULT 0,
					purchases   INTEGER       NOT NULL DEFAULT 0,
					revenue     NUMERIC(14,2) NOT NULL DEFAULT 0.00,
					created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW()
				);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_product_performance_product_date
					ON analytics.product_performance(product_no, date);
				CREATE        INDEX IF NOT EXISTS idx_product_performance_date
					ON analytics.product_performance(date);
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
		// ── auth schema ───────────────────────────────────────────────────────
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

		// ── catalog schema ────────────────────────────────────────────────────
		{
			name: "fk_categories_parent_no",
			query: `
				DO $$ BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM information_schema.table_constraints
						WHERE constraint_name = 'fk_categories_parent_no'
					) THEN
						ALTER TABLE catalog.categories
						ADD CONSTRAINT fk_categories_parent_no
						FOREIGN KEY (parent_no) REFERENCES catalog.categories(category_no)
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
						ALTER TABLE catalog.attributes
						ADD CONSTRAINT fk_attributes_category_no
						FOREIGN KEY (category_no) REFERENCES catalog.categories(category_no)
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
						ALTER TABLE catalog.attribute_values
						ADD CONSTRAINT fk_attribute_values_attribute_no
						FOREIGN KEY (attribute_no) REFERENCES catalog.attributes(attribute_no)
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
						ALTER TABLE catalog.products
						ADD CONSTRAINT fk_products_category_no
						FOREIGN KEY (category_no) REFERENCES catalog.categories(category_no)
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
						ALTER TABLE catalog.product_variants
						ADD CONSTRAINT fk_product_variants_product_no
						FOREIGN KEY (product_no) REFERENCES catalog.products(product_no)
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
						ALTER TABLE catalog.variant_attribute_values
						ADD CONSTRAINT fk_vav_variant_no
						FOREIGN KEY (variant_no) REFERENCES catalog.product_variants(variant_no)
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
						ALTER TABLE catalog.variant_attribute_values
						ADD CONSTRAINT fk_vav_attribute_value_no
						FOREIGN KEY (attribute_value_no) REFERENCES catalog.attribute_values(attribute_value_no)
						ON DELETE CASCADE;
					END IF;
				END $$;
			`,
		},
		{
			name: "fk_variant_images_variant_no",
			query: `
				DO $$ BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM information_schema.table_constraints
						WHERE constraint_name = 'fk_variant_images_variant_no'
					) THEN
						ALTER TABLE catalog.product_variant_images
						ADD CONSTRAINT fk_variant_images_variant_no
						FOREIGN KEY (variant_no) REFERENCES catalog.product_variants(variant_no)
						ON DELETE CASCADE;
					END IF;
				END $$;
			`,
		},

		// ── customer schema ───────────────────────────────────────────────────
		{
			name: "fk_customer_credentials_customer_no",
			query: `
				DO $$ BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM information_schema.table_constraints
						WHERE constraint_name = 'fk_customer_credentials_customer_no'
					) THEN
						ALTER TABLE customer.credentials
						ADD CONSTRAINT fk_customer_credentials_customer_no
						FOREIGN KEY (customer_no) REFERENCES customer.customers(customer_no)
						ON DELETE CASCADE;
					END IF;
				END $$;
			`,
		},
		{
			name: "fk_addresses_customer_no",
			query: `
				DO $$ BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM information_schema.table_constraints
						WHERE constraint_name = 'fk_addresses_customer_no'
					) THEN
						ALTER TABLE customer.addresses
						ADD CONSTRAINT fk_addresses_customer_no
						FOREIGN KEY (customer_no) REFERENCES customer.customers(customer_no)
						ON DELETE CASCADE;
					END IF;
				END $$;
			`,
		},
		{
			name: "fk_wishlists_customer_no",
			query: `
				DO $$ BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM information_schema.table_constraints
						WHERE constraint_name = 'fk_wishlists_customer_no'
					) THEN
						ALTER TABLE customer.wishlists
						ADD CONSTRAINT fk_wishlists_customer_no
						FOREIGN KEY (customer_no) REFERENCES customer.customers(customer_no)
						ON DELETE CASCADE;
					END IF;
				END $$;
			`,
		},
		{
			name: "fk_wishlists_product_no",
			query: `
				DO $$ BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM information_schema.table_constraints
						WHERE constraint_name = 'fk_wishlists_product_no'
					) THEN
						ALTER TABLE customer.wishlists
						ADD CONSTRAINT fk_wishlists_product_no
						FOREIGN KEY (product_no) REFERENCES catalog.products(product_no)
						ON DELETE CASCADE;
					END IF;
				END $$;
			`,
		},

		// ── orders schema ─────────────────────────────────────────────────────
		{
			name: "fk_carts_customer_no",
			query: `
				DO $$ BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM information_schema.table_constraints
						WHERE constraint_name = 'fk_carts_customer_no'
					) THEN
						ALTER TABLE orders.carts
						ADD CONSTRAINT fk_carts_customer_no
						FOREIGN KEY (customer_no) REFERENCES customer.customers(customer_no)
						ON DELETE CASCADE;
					END IF;
				END $$;
			`,
		},
		{
			name: "fk_cart_items_cart_no",
			query: `
				DO $$ BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM information_schema.table_constraints
						WHERE constraint_name = 'fk_cart_items_cart_no'
					) THEN
						ALTER TABLE orders.cart_items
						ADD CONSTRAINT fk_cart_items_cart_no
						FOREIGN KEY (cart_no) REFERENCES orders.carts(cart_no)
						ON DELETE CASCADE;
					END IF;
				END $$;
			`,
		},
		{
			name: "fk_orders_customer_no",
			query: `
				DO $$ BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM information_schema.table_constraints
						WHERE constraint_name = 'fk_orders_customer_no'
					) THEN
						ALTER TABLE orders.orders
						ADD CONSTRAINT fk_orders_customer_no
						FOREIGN KEY (customer_no) REFERENCES customer.customers(customer_no)
						ON DELETE RESTRICT;
					END IF;
				END $$;
			`,
		},
		{
			name: "fk_orders_address_no",
			query: `
				DO $$ BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM information_schema.table_constraints
						WHERE constraint_name = 'fk_orders_address_no'
					) THEN
						ALTER TABLE orders.orders
						ADD CONSTRAINT fk_orders_address_no
						FOREIGN KEY (address_no) REFERENCES customer.addresses(address_no)
						ON DELETE RESTRICT;
					END IF;
				END $$;
			`,
		},
		{
			name: "fk_order_items_order_no",
			query: `
				DO $$ BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM information_schema.table_constraints
						WHERE constraint_name = 'fk_order_items_order_no'
					) THEN
						ALTER TABLE orders.order_items
						ADD CONSTRAINT fk_order_items_order_no
						FOREIGN KEY (order_no) REFERENCES orders.orders(order_no)
						ON DELETE CASCADE;
					END IF;
				END $$;
			`,
		},
		{
			name: "fk_order_status_history_order_no",
			query: `
				DO $$ BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM information_schema.table_constraints
						WHERE constraint_name = 'fk_order_status_history_order_no'
					) THEN
						ALTER TABLE orders.order_status_history
						ADD CONSTRAINT fk_order_status_history_order_no
						FOREIGN KEY (order_no) REFERENCES orders.orders(order_no)
						ON DELETE CASCADE;
					END IF;
				END $$;
			`,
		},

		// ── payment schema ────────────────────────────────────────────────────
		{
			name: "fk_payments_order_no",
			query: `
				DO $$ BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM information_schema.table_constraints
						WHERE constraint_name = 'fk_payments_order_no'
					) THEN
						ALTER TABLE payment.payments
						ADD CONSTRAINT fk_payments_order_no
						FOREIGN KEY (order_no) REFERENCES orders.orders(order_no)
						ON DELETE RESTRICT;
					END IF;
				END $$;
			`,
		},
		{
			name: "fk_payments_customer_no",
			query: `
				DO $$ BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM information_schema.table_constraints
						WHERE constraint_name = 'fk_payments_customer_no'
					) THEN
						ALTER TABLE payment.payments
						ADD CONSTRAINT fk_payments_customer_no
						FOREIGN KEY (customer_no) REFERENCES customer.customers(customer_no)
						ON DELETE RESTRICT;
					END IF;
				END $$;
			`,
		},
		{
			name: "fk_refunds_payment_no",
			query: `
				DO $$ BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM information_schema.table_constraints
						WHERE constraint_name = 'fk_refunds_payment_no'
					) THEN
						ALTER TABLE payment.refunds
						ADD CONSTRAINT fk_refunds_payment_no
						FOREIGN KEY (payment_no) REFERENCES payment.payments(payment_no)
						ON DELETE RESTRICT;
					END IF;
				END $$;
			`,
		},
		{
			name: "fk_refunds_order_no",
			query: `
				DO $$ BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM information_schema.table_constraints
						WHERE constraint_name = 'fk_refunds_order_no'
					) THEN
						ALTER TABLE payment.refunds
						ADD CONSTRAINT fk_refunds_order_no
						FOREIGN KEY (order_no) REFERENCES orders.orders(order_no)
						ON DELETE RESTRICT;
					END IF;
				END $$;
			`,
		},

		// ── delivery schema ───────────────────────────────────────────────────
		{
			name: "fk_shipments_order_no",
			query: `
				DO $$ BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM information_schema.table_constraints
						WHERE constraint_name = 'fk_shipments_order_no'
					) THEN
						ALTER TABLE delivery.shipments
						ADD CONSTRAINT fk_shipments_order_no
						FOREIGN KEY (order_no) REFERENCES orders.orders(order_no)
						ON DELETE RESTRICT;
					END IF;
				END $$;
			`,
		},
		{
			name: "fk_tracking_events_shipment_no",
			query: `
				DO $$ BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM information_schema.table_constraints
						WHERE constraint_name = 'fk_tracking_events_shipment_no'
					) THEN
						ALTER TABLE delivery.tracking_events
						ADD CONSTRAINT fk_tracking_events_shipment_no
						FOREIGN KEY (shipment_no) REFERENCES delivery.shipments(shipment_no)
						ON DELETE CASCADE;
					END IF;
				END $$;
			`,
		},

		// ── notification schema ───────────────────────────────────────────────
		{
			name: "fk_notifications_template_no",
			query: `
				DO $$ BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM information_schema.table_constraints
						WHERE constraint_name = 'fk_notifications_template_no'
					) THEN
						ALTER TABLE notification.notifications
						ADD CONSTRAINT fk_notifications_template_no
						FOREIGN KEY (template_no) REFERENCES notification.templates(template_no)
						ON DELETE SET NULL;
					END IF;
				END $$;
			`,
		},
		{
			name: "fk_notifications_customer_no",
			query: `
				DO $$ BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM information_schema.table_constraints
						WHERE constraint_name = 'fk_notifications_customer_no'
					) THEN
						ALTER TABLE notification.notifications
						ADD CONSTRAINT fk_notifications_customer_no
						FOREIGN KEY (customer_no) REFERENCES customer.customers(customer_no)
						ON DELETE SET NULL;
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

// ── Step 5: Seed Data ─────────────────────────────────────────────────────────

func seedData(ctx context.Context, db *pgxpool.Pool) error {
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
		{"Employees", "systems", "fa-users", 2},
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
		{sectionNos[0], "Sections", "section", 1},
		{sectionNos[0], "Pages", "pages", 2},
		{sectionNos[1], "Roles", "roles", 1},
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
	hashed, err := bcrypt.GenerateFromPassword([]byte("Khelaghar@786"), bcrypt.DefaultCost)
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

	// ── Notification Templates ────────────────────────────────────────────────
	notificationTemplates := []struct {
		key     string
		channel string
		subject string
		body    string
	}{
		{
			"order_confirmed", "email",
			"Your order {{order_ref}} has been confirmed",
			"Hi {{first_name}}, your order {{order_ref}} worth {{total}} has been confirmed and is being processed.",
		},
		{
			"order_shipped", "email",
			"Your order {{order_ref}} has been shipped",
			"Hi {{first_name}}, your order {{order_ref}} is on its way. Tracking number: {{tracking_no}}.",
		},
		{
			"order_delivered", "email",
			"Your order {{order_ref}} has been delivered",
			"Hi {{first_name}}, your order {{order_ref}} has been delivered. We hope you enjoy your purchase!",
		},
		{
			"order_cancelled", "email",
			"Your order {{order_ref}} has been cancelled",
			"Hi {{first_name}}, your order {{order_ref}} has been cancelled. If this was a mistake please contact support.",
		},
		{
			"order_confirmed", "sms",
			"",
			"Order {{order_ref}} confirmed. Total: {{total}}. Thank you for shopping with us!",
		},
		{
			"order_shipped", "sms",
			"",
			"Order {{order_ref}} shipped. Tracking: {{tracking_no}}.",
		},
		{
			"password_reset", "email",
			"Reset your password",
			"Hi {{first_name}}, click the link to reset your password: {{reset_link}}. Link expires in 30 minutes.",
		},
		{
			"welcome", "email",
			"Welcome to Game n Glam!",
			"Hi {{first_name}}, welcome to Game n Glam! Your account has been created successfully.",
		},
	}

	for _, t := range notificationTemplates {
		_, err = tx.Exec(ctx, `
			INSERT INTO notification.templates (template_id, template_key, channel, subject, body, status)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, uuid.New(), t.key, t.channel, t.subject, t.body, "A")
		if err != nil {
			return fmt.Errorf("seedData: insert notification template [%s/%s]: %w", t.key, t.channel, err)
		}
	}

	// ── Commit ────────────────────────────────────────────────────────────────
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("seedData: commit: %w", err)
	}

	log.Println("seed data inserted successfully")
	return nil
}
