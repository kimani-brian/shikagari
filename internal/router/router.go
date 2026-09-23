package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shikagari/api/config"
	"github.com/shikagari/api/internal/domain"
	"github.com/shikagari/api/internal/handler"
	"github.com/shikagari/api/internal/middleware"
	"github.com/shikagari/api/internal/repository/postgres"
	"github.com/shikagari/api/internal/service"
	"github.com/shikagari/api/pkg/hash"
	jwtpkg "github.com/shikagari/api/pkg/jwt"
	"github.com/shikagari/api/pkg/response"
	"gorm.io/gorm"
)

// Setup builds and returns the fully configured Gin engine.
// All dependencies are wired here: repositories → services → handlers → routes.
func Setup(cfg *config.Config, db *gorm.DB) *gin.Engine {
	// ── 1. Engine setup ───────────────────────────────────────────────────────
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New() // Use gin.New() — we register our own logger

	// ── 2. Global middleware ──────────────────────────────────────────────────
	r.Use(middleware.Logger())
	r.Use(middleware.CORS(cfg.App.AllowedOrigins))
	r.Use(gin.Recovery()) // Recover from panics; return 500

	// ── 3. Static file serving for uploaded images ────────────────────────────
	r.Static("/uploads", cfg.Upload.Dir)

	// ── 4. Health check (infrastructure/load balancer use) ───────────────────
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": cfg.App.Name,
			"env":     cfg.App.Env,
		})
	})

	// ── 5. Wire dependencies ──────────────────────────────────────────────────
	deps := wireDependencies(cfg, db)

	// ── 6. Register all route groups ─────────────────────────────────────────
	api := r.Group("/api/v1")
	{
		registerAuthRoutes(api, deps)
		registerUserRoutes(api, deps)
		registerDealerRoutes(api, deps)
		registerPrivateSellerRoutes(api, deps)
		registerListingRoutes(api, deps, deps.inquiryHandler)
		registerFavoriteRoutes(api, deps)
		registerInquiryRoutes(api, deps)
		registerAdminRoutes(api, deps)
	}

	// ── 7. 404 handler ────────────────────────────────────────────────────────
	r.NoRoute(func(c *gin.Context) {
		response.NotFound(c, "the requested route does not exist")
	})

	return r
}

// ── Dependency Container ──────────────────────────────────────────────────────

// deps holds all initialised handlers and middleware for route registration.
type deps struct {
	// Middleware
	jwtManager *jwtpkg.Manager
	authMW     gin.HandlerFunc // strict — blocks if no token
	optAuthMW  gin.HandlerFunc // soft — enriches if token present

	// Handlers
	authHandler          *handler.AuthHandler
	userHandler          *handler.UserHandler
	dealerHandler        *handler.DealerHandler
	privateSellerHandler *handler.PrivateSellerHandler
	listingHandler       *handler.ListingHandler
	favoriteHandler      *handler.FavoriteHandler
	inquiryHandler       *handler.InquiryHandler
}

// wireDependencies constructs all repositories, services, and handlers
// following the dependency injection pattern.
func wireDependencies(cfg *config.Config, db *gorm.DB) *deps {
	// ── Packages ──────────────────────────────────────────────────────────────
	hasher := hash.New()
	jwtManager := jwtpkg.New(cfg.JWT.Secret, cfg.JWT.ExpiryHours)
	uploadSvc := service.NewUploadService(cfg)

	// ── Repositories ─────────────────────────────────────────────────────────
	userRepo := postgres.NewUserRepository(db)
	resetTokenRepo := postgres.NewPasswordResetTokenRepository(db)
	sessionRepo := postgres.NewSessionRepository(db)
	securityEventRepo := postgres.NewSecurityEventRepository(db)
	dealerRepo := postgres.NewDealerRepository(db)
	privateSellerRepo := postgres.NewPrivateSellerRepository(db)
	listingRepo := postgres.NewListingRepository(db)
	favoriteRepo := postgres.NewFavoriteRepository(db)
	inquiryRepo := postgres.NewInquiryRepository(db)

	// ── Services ──────────────────────────────────────────────────────────────
	authSvc := service.NewAuthService(userRepo, resetTokenRepo, sessionRepo, securityEventRepo, hasher, jwtManager)
	userSvc := service.NewUserService(userRepo, hasher)
	dealerSvc := service.NewDealerService(dealerRepo, userRepo)
	privateSellerSvc := service.NewPrivateSellerService(privateSellerRepo, userRepo)
	listingSvc := service.NewListingService(listingRepo, dealerRepo, privateSellerRepo)
	favoriteSvc := service.NewFavoriteService(favoriteRepo, listingRepo)
	inquirySvc := service.NewInquiryService(inquiryRepo, listingRepo)

	// ── Handlers ──────────────────────────────────────────────────────────────
	authH := handler.NewAuthHandler(authSvc)
	userH := handler.NewUserHandler(userSvc)
	dealerH := handler.NewDealerHandler(dealerSvc, uploadSvc)
	privateSellerH := handler.NewPrivateSellerHandler(privateSellerSvc, uploadSvc)
	listingH := handler.NewListingHandler(listingSvc, uploadSvc)
	favoriteH := handler.NewFavoriteHandler(favoriteSvc)
	inquiryH := handler.NewInquiryHandler(inquirySvc)

	return &deps{
		jwtManager:           jwtManager,
		authMW:               middleware.Authenticate(jwtManager, sessionRepo),
		optAuthMW:            middleware.OptionalAuthenticate(jwtManager, sessionRepo),
		authHandler:          authH,
		userHandler:          userH,
		dealerHandler:        dealerH,
		privateSellerHandler: privateSellerH,
		listingHandler:       listingH,
		favoriteHandler:      favoriteH,
		inquiryHandler:       inquiryH,
	}
}

// ── Route Registration Functions ──────────────────────────────────────────────

// registerAuthRoutes mounts public authentication endpoints.
//
//	POST   /api/v1/auth/register
//	POST   /api/v1/auth/login
//	POST   /api/v1/auth/password/reset-request
//	POST   /api/v1/auth/password/reset
//	GET    /api/v1/auth/me                 [auth required]
//	GET    /api/v1/auth/security-events    [auth required]
//	GET    /api/v1/auth/sessions           [auth required]
//	DELETE /api/v1/auth/sessions/:id       [auth required]
//	POST   /api/v1/auth/logout             [auth required]
func registerAuthRoutes(rg *gin.RouterGroup, d *deps) {
	auth := rg.Group("/auth")
	{
		auth.POST("/register", d.authHandler.Register)
		auth.POST("/login", d.authHandler.Login)
		auth.POST("/password/reset-request", d.authHandler.RequestPasswordReset)
		auth.POST("/password/reset", d.authHandler.ResetPassword)

		// Protected: returns current user info from JWT
		auth.GET("/me", d.authMW, d.authHandler.Me)
		auth.GET("/security-events", d.authMW, d.authHandler.ListSecurityEvents)
		auth.GET("/sessions", d.authMW, d.authHandler.ListSessions)
		auth.DELETE("/sessions/:id", d.authMW, d.authHandler.RevokeSession)
		auth.POST("/logout", d.authMW, d.authHandler.Logout)
	}
}

// registerUserRoutes mounts user profile management endpoints.
//
//	GET    /api/v1/users/me              [auth required]
//	PATCH  /api/v1/users/me              [auth required]
//	PATCH  /api/v1/users/me/password     [auth required]
func registerUserRoutes(rg *gin.RouterGroup, d *deps) {
	users := rg.Group("/users", d.authMW)
	{
		users.GET("/me", d.userHandler.GetMyProfile)
		users.PATCH("/me", d.userHandler.UpdateMyProfile)
		users.PATCH("/me/password", d.userHandler.ChangePassword)
	}
}

// registerDealerRoutes mounts dealer profile endpoints.
//
//	GET    /api/v1/dealers                    [public] list approved dealers
//	GET    /api/v1/dealers/:id                [public]
//	POST   /api/v1/dealers/profile            [auth + seller role]
//	GET    /api/v1/dealers/profile            [auth required]
//	PATCH  /api/v1/dealers/profile            [auth required]
//	POST   /api/v1/dealers/profile/logo       [auth required]
func registerDealerRoutes(rg *gin.RouterGroup, d *deps) {
	// Public: list approved dealers and view profiles
	rg.GET("/dealers", d.dealerHandler.ListApproved)
	rg.GET("/dealers/:id", d.dealerHandler.GetProfileByID)

	// Authenticated: manage own dealer profile
	dealers := rg.Group("/dealers", d.authMW)
	{
		dealers.POST("/profile", middleware.RequireSeller(), d.dealerHandler.CreateProfile)
		dealers.GET("/profile", d.dealerHandler.GetMyProfile)
		dealers.PATCH("/profile", d.dealerHandler.UpdateProfile)
		dealers.POST("/profile/logo", d.dealerHandler.UploadLogo)
	}
}

// registerPrivateSellerRoutes mounts private seller profile endpoints.
//
//	POST   /api/v1/sellers/profile            [auth + seller role]
//	GET    /api/v1/sellers/profile            [auth required]
//	PATCH  /api/v1/sellers/profile            [auth required]
//	POST   /api/v1/sellers/profile/photo      [auth required]
//	GET    /api/v1/sellers/:id                [public]
func registerPrivateSellerRoutes(rg *gin.RouterGroup, d *deps) {
	// Public: view a private seller's profile by ID
	rg.GET("/sellers/:id", d.privateSellerHandler.GetProfileByID)

	// Authenticated: manage own private seller profile
	sellers := rg.Group("/sellers", d.authMW)
	{
		sellers.POST("/profile", middleware.RequireRole(domain.RoleBuyer, domain.RoleSeller, domain.RoleAdmin), d.privateSellerHandler.CreateProfile)
		sellers.GET("/profile", d.privateSellerHandler.GetMyProfile)
		sellers.PATCH("/profile", d.privateSellerHandler.UpdateProfile)
		sellers.POST("/profile/photo", d.privateSellerHandler.UploadProfilePhoto)
	}
}

// registerListingRoutes mounts vehicle listing endpoints.
//
//	GET    /api/v1/listings                   [public]
//	GET    /api/v1/listings/me                [auth required]
//	GET    /api/v1/listings/:id               [public]
//	POST   /api/v1/listings                   [auth + seller role]
//	PATCH  /api/v1/listings/:id               [auth required — owner or admin]
//	DELETE /api/v1/listings/:id               [auth required — owner or admin]
//	POST   /api/v1/listings/:id/images        [auth required — owner only]
//	POST   /api/v1/listings/:id/inquiries     [auth required]
func registerListingRoutes(rg *gin.RouterGroup, d *deps, inquiryH *handler.InquiryHandler) {
	// Public: search and view listings
	rg.GET("/listings", d.listingHandler.Search)
	rg.GET("/listings/:id", d.listingHandler.GetByID)

	// Authenticated listing management
	listings := rg.Group("/listings", d.authMW)
	{
		// Seller-only: create a listing
		listings.POST("", middleware.RequireSeller(), d.listingHandler.Create)

		// Owner or admin: edit and delete
		listings.PATCH("/:id", d.listingHandler.Update)
		listings.DELETE("/:id", d.listingHandler.Delete)

		// Owner only: manage images
		listings.POST("/:id/images", d.listingHandler.UploadImages)

		// Any authenticated user: send inquiry on a listing
		listings.POST("/:id/inquiries", inquiryH.Send)

		// My listings (seller dashboard)
		listings.GET("/me", d.listingHandler.GetMyListings)
	}
}

// registerFavoriteRoutes mounts wishlist endpoints.
//
//	GET    /api/v1/favorites                  [auth required]
//	POST   /api/v1/favorites/:id/toggle       [auth required]
func registerFavoriteRoutes(rg *gin.RouterGroup, d *deps) {
	favorites := rg.Group("/favorites", d.authMW)
	{
		favorites.GET("", d.favoriteHandler.GetMyFavorites)
		favorites.POST("/:id/toggle", d.favoriteHandler.Toggle)
	}
}

// registerInquiryRoutes mounts buyer/seller inquiry management endpoints.
//
//	GET    /api/v1/inquiries/sent             [auth required — buyer]
//	GET    /api/v1/inquiries/inbox            [auth required — seller]
//	GET    /api/v1/inquiries/:id              [auth required — buyer or seller]
//	PATCH  /api/v1/inquiries/:id/reply        [auth required — seller only]
//	PATCH  /api/v1/inquiries/:id/status       [auth required — seller only]
func registerInquiryRoutes(rg *gin.RouterGroup, d *deps) {
	inquiries := rg.Group("/inquiries", d.authMW)
	{
		inquiries.GET("/sent", d.inquiryHandler.GetMyInquiries)
		inquiries.GET("/inbox", d.inquiryHandler.GetMyInbox)
		inquiries.GET("/:id", d.inquiryHandler.GetByID)
		inquiries.PATCH("/:id/reply", d.inquiryHandler.Reply)
		inquiries.PATCH("/:id/status", d.inquiryHandler.UpdateStatus)
	}
}

// registerAdminRoutes mounts all admin-only management endpoints.
// Every route in this group requires a valid JWT with the admin role.
//
//	── Users ──
//	GET    /api/v1/admin/users
//	GET    /api/v1/admin/users/:id
//	PATCH  /api/v1/admin/users/:id
//	DELETE /api/v1/admin/users/:id
//
//	── Dealer profiles ──
//	GET    /api/v1/admin/dealers
//	PATCH  /api/v1/admin/dealers/:id/review
//
//	── Private seller profiles ──
//	GET    /api/v1/admin/sellers
//	PATCH  /api/v1/admin/sellers/:id/review
func registerAdminRoutes(rg *gin.RouterGroup, d *deps) {
	admin := rg.Group("/admin", d.authMW, middleware.RequireAdmin())
	{
		// ── User management ───────────────────────────────────────────────────
		adminUsers := admin.Group("/users")
		{
			adminUsers.GET("", d.userHandler.AdminListUsers)
			adminUsers.GET("/:id", d.userHandler.AdminGetUser)
			adminUsers.PATCH("/:id", d.userHandler.AdminUpdateUser)
			adminUsers.DELETE("/:id", d.userHandler.AdminDeleteUser)
		}

		// ── Dealer profile review ─────────────────────────────────────────────
		adminDealers := admin.Group("/dealers")
		{
			adminDealers.GET("", d.dealerHandler.AdminListProfiles)
			adminDealers.PATCH("/:id/review", d.dealerHandler.AdminReviewProfile)
		}

		// ── Private seller profile review ─────────────────────────────────────
		adminSellers := admin.Group("/sellers")
		{
			adminSellers.GET("", d.privateSellerHandler.AdminListProfiles)
			adminSellers.PATCH("/:id/review", d.privateSellerHandler.AdminReviewProfile)
		}
	}
}
