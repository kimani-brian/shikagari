package database

import (
	"log"

	"github.com/shikagari/api/config"
	"github.com/shikagari/api/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the shared GORM database instance used across the application.
var DB *gorm.DB

// Connect establishes a connection to PostgreSQL using the provided config
// and runs GORM AutoMigrate to keep the schema in sync with domain models.
func Connect(cfg *config.Config) *gorm.DB {
	logLevel := logger.Info
	if cfg.App.Env == "production" {
		logLevel = logger.Error // Reduce noise in production logs
	}

	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		// Disable automatic pluralisation of table names to match our migrations
		NamingStrategy: nil,
	})
	if err != nil {
		log.Fatalf("[database] FATAL: failed to connect to PostgreSQL: %v", err)
	}

	// Configure connection pool for production workloads
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("[database] FATAL: failed to get underlying sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(25)   // Maximum concurrent connections
	sqlDB.SetMaxIdleConns(10)   // Connections kept alive when idle
	sqlDB.SetConnMaxLifetime(0) // No connection lifetime limit

	log.Println("[database] Connected to PostgreSQL successfully")

	// Run schema migrations
	runMigrations(db)

	DB = db
	return db
}

// runMigrations applies GORM AutoMigrate for all domain models.
// AutoMigrate only adds new columns/tables — it never drops existing ones.
func runMigrations(db *gorm.DB) {
	log.Println("[database] Running schema migrations...")

	err := db.AutoMigrate(
		&domain.User{},
		&domain.DealerProfile{},
		&domain.PrivateSellerProfile{},
		&domain.Listing{},
		&domain.Favorite{},
		&domain.Inquiry{},
		&domain.PasswordResetToken{},
		&domain.SecurityEvent{},
		&domain.UserSession{},
	)
	if err != nil {
		log.Fatalf("[database] FATAL: migration failed: %v", err)
	}

	// Apply composite unique index on favorites (user_id, listing_id)
	// Prevents a user from saving the same listing more than once
	if err := applyCustomIndexes(db); err != nil {
		log.Fatalf("[database] FATAL: failed to apply custom indexes: %v", err)
	}

	log.Println("[database] Migrations completed successfully")
}

// applyCustomIndexes creates indexes that GORM AutoMigrate cannot infer
// from struct tags alone (e.g. composite unique constraints).
func applyCustomIndexes(db *gorm.DB) error {
	indexes := []string{
		// Prevent duplicate favorites
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_favorites_user_listing
         ON favorites (user_id, listing_id)`,

		// Speed up listing search queries
		`CREATE INDEX IF NOT EXISTS idx_listings_make_model
         ON listings (make, model)`,

		`CREATE INDEX IF NOT EXISTS idx_listings_price
         ON listings (price_kes)`,

		`CREATE INDEX IF NOT EXISTS idx_listings_year
         ON listings (year)`,

		`CREATE INDEX IF NOT EXISTS idx_listings_status
         ON listings (status)`,

		// Account security analytics
		`CREATE INDEX IF NOT EXISTS idx_security_events_user_id
         ON security_events (user_id)`,

		`CREATE INDEX IF NOT EXISTS idx_security_events_created_at
         ON security_events (created_at DESC)`,

		// Session lifecycle management
		`CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id
         ON user_sessions (user_id)`,

		`CREATE INDEX IF NOT EXISTS idx_user_sessions_revoked_at
         ON user_sessions (revoked_at)`,

		`CREATE INDEX IF NOT EXISTS idx_user_sessions_last_active
         ON user_sessions (last_active DESC)`,
	}

	for _, query := range indexes {
		if err := db.Exec(query).Error; err != nil {
			return err
		}
	}
	return nil
}
