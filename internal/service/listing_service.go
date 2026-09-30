package service

import (
	"errors"

	"github.com/google/uuid"
	"github.com/shikagari/api/internal/domain"
	"github.com/shikagari/api/internal/dto"
	"github.com/shikagari/api/internal/repository/interfaces"
	"github.com/shikagari/api/pkg/response"
)

// ListingService handles vehicle listing creation, updates, search, and deletion.
type ListingService struct {
	listingRepo interfaces.ListingRepository
	dealerRepo  interfaces.DealerRepository
	sellerRepo  interfaces.PrivateSellerRepository
}

// NewListingService constructs a ListingService with its dependencies.
func NewListingService(
	listingRepo interfaces.ListingRepository,
	dealerRepo interfaces.DealerRepository,
	sellerRepo interfaces.PrivateSellerRepository,
) *ListingService {
	return &ListingService{
		listingRepo: listingRepo,
		dealerRepo:  dealerRepo,
		sellerRepo:  sellerRepo,
	}
}

// Create creates a new vehicle listing.
// The seller must have an approved dealer or private seller profile.
func (s *ListingService) Create(
	userID uuid.UUID,
	req dto.CreateListingRequest,
) (*dto.ListingResponse, error) {
	// ── 1. Resolve seller type and verify approval ────────────────────────────
	sellerType, err := s.resolveApprovedSellerType(userID)
	if err != nil {
		return nil, err
	}

	// ── 2. Build and persist the listing ─────────────────────────────────────
	listing := &domain.Listing{
		UserID:       userID,
		SellerType:   sellerType,
		Title:        req.Title,
		Description:  req.Description,
		PriceKES:     req.PriceKES,
		Location:     req.Location,
		BodyType:     domain.BodyType(req.BodyType),
		Make:         req.Make,
		Model:        req.Model,
		Year:         req.Year,
		Mileage:      req.Mileage,
		FuelType:     domain.FuelType(req.FuelType),
		Transmission: domain.Transmission(req.Transmission),
		Drivetrain:   req.Drivetrain,
		EngineSize:   req.EngineSize,
		Color:        req.Color,
		Status:       domain.ListingActive,
		Images:       []string{},
	}
	if req.Doors != nil {
		listing.Doors = *req.Doors
	}

	if err := s.listingRepo.Create(listing); err != nil {
		return nil, errors.New("failed to create listing")
	}

	// ── 3. Reload with full associations for response ─────────────────────────
	listing, err = s.listingRepo.FindByID(listing.ID)
	if err != nil || listing == nil {
		return nil, errors.New("listing created but failed to load")
	}

	res := dto.ToListingResponse(*listing)
	return &res, nil
}

// GetByID retrieves a single listing by UUID.
// Increments the view counter on every public access.
func (s *ListingService) GetByID(id uuid.UUID) (*dto.ListingResponse, error) {
	listing, err := s.listingRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("failed to retrieve listing")
	}
	if listing == nil {
		return nil, errors.New("listing not found")
	}

	// Fire-and-forget view increment — failure is non-critical
	_ = s.listingRepo.IncrementViewCount(id)

	res := dto.ToListingResponse(*listing)
	return &res, nil
}

// Search returns a paginated list of active listings matching the given filters.
func (s *ListingService) Search(
	filters dto.ListingFilterRequest,
) ([]dto.ListingCardResponse, int64, error) {
	// Enforce sane pagination defaults
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.PerPage < 1 || filters.PerPage > 50 {
		filters.PerPage = 20
	}

	listings, total, err := s.listingRepo.Search(filters)
	if err != nil {
		return nil, 0, errors.New("failed to search listings")
	}

	var cards []dto.ListingCardResponse
	for _, l := range listings {
		cards = append(cards, dto.ToListingCardResponse(l))
	}

	return cards, total, nil
}

// GetMyListings returns all listings created by the authenticated seller.
func (s *ListingService) GetMyListings(
	userID uuid.UUID,
	page, perPage int,
) ([]dto.ListingCardResponse, int64, error) {
	listings, total, err := s.listingRepo.FindByUserID(userID, page, perPage)
	if err != nil {
		return nil, 0, errors.New("failed to retrieve your listings")
	}

	var cards []dto.ListingCardResponse
	for _, l := range listings {
		cards = append(cards, dto.ToListingCardResponse(l))
	}
	return cards, total, nil
}

// Update applies partial updates to an existing listing.
// Only the listing owner or an admin may update.
func (s *ListingService) Update(
	listingID uuid.UUID,
	userID uuid.UUID,
	isAdmin bool,
	req dto.UpdateListingRequest,
) (*dto.ListingResponse, error) {
	listing, err := s.listingRepo.FindByID(listingID)
	if err != nil || listing == nil {
		return nil, errors.New("listing not found")
	}

	// ── Ownership check ───────────────────────────────────────────────────────
	if !isAdmin {
		owned, err := s.listingRepo.BelongsToUser(listingID, userID)
		if err != nil {
			return nil, errors.New("failed to verify ownership")
		}
		if !owned {
			return nil, errors.New("you do not have permission to update this listing")
		}
	}

	// ── Apply partial updates ─────────────────────────────────────────────────
	if req.Title != nil {
		listing.Title = *req.Title
	}
	if req.Description != nil {
		listing.Description = *req.Description
	}
	if req.PriceKES != nil {
		listing.PriceKES = *req.PriceKES
	}
	if req.Location != nil {
		listing.Location = *req.Location
	}
	if req.BodyType != nil {
		listing.BodyType = domain.BodyType(*req.BodyType)
	}
	if req.Make != nil {
		listing.Make = *req.Make
	}
	if req.Model != nil {
		listing.Model = *req.Model
	}
	if req.Year != nil {
		listing.Year = *req.Year
	}
	if req.Mileage != nil {
		listing.Mileage = *req.Mileage
	}
	if req.FuelType != nil {
		listing.FuelType = domain.FuelType(*req.FuelType)
	}
	if req.Transmission != nil {
		listing.Transmission = domain.Transmission(*req.Transmission)
	}
	if req.Drivetrain != nil {
		listing.Drivetrain = *req.Drivetrain
	}
	if req.EngineSize != nil {
		listing.EngineSize = *req.EngineSize
	}
	if req.Doors != nil {
		listing.Doors = *req.Doors
	}
	if req.Color != nil {
		listing.Color = *req.Color
	}
	if req.Status != nil {
		listing.Status = domain.ListingStatus(*req.Status)
	}

	if err := s.listingRepo.Update(listing); err != nil {
		return nil, errors.New("failed to update listing")
	}

	listing, err = s.listingRepo.FindByID(listingID)
	if err != nil || listing == nil {
		return nil, errors.New("listing updated but failed to reload")
	}

	res := dto.ToListingResponse(*listing)
	return &res, nil
}

// Delete soft-deletes a listing.
// Only the listing owner or an admin may delete.
func (s *ListingService) Delete(
	listingID uuid.UUID,
	userID uuid.UUID,
	isAdmin bool,
) error {
	listing, err := s.listingRepo.FindByID(listingID)
	if err != nil || listing == nil {
		return errors.New("listing not found")
	}

	if !isAdmin {
		owned, err := s.listingRepo.BelongsToUser(listingID, userID)
		if err != nil {
			return errors.New("failed to verify ownership")
		}
		if !owned {
			return errors.New("you do not have permission to delete this listing")
		}
	}

	return s.listingRepo.Delete(listingID)
}

// AddImages appends new image URLs to a listing's image array.
// Only the listing owner may add images.
func (s *ListingService) AddImages(
	listingID uuid.UUID,
	userID uuid.UUID,
	newImageURLs []string,
) (*dto.ListingResponse, error) {
	listing, err := s.listingRepo.FindByID(listingID)
	if err != nil || listing == nil {
		return nil, errors.New("listing not found")
	}

	owned, err := s.listingRepo.BelongsToUser(listingID, userID)
	if err != nil || !owned {
		return nil, errors.New("you do not have permission to add images to this listing")
	}

	// Enforce a maximum of 10 images per listing
	combined := append([]string(listing.Images), newImageURLs...)
	if len(combined) > 10 {
		return nil, errors.New("a listing may not have more than 10 images")
	}

	if err := s.listingRepo.UpdateImages(listingID, combined); err != nil {
		return nil, errors.New("failed to update listing images")
	}

	listing, _ = s.listingRepo.FindByID(listingID)
	res := dto.ToListingResponse(*listing)
	return &res, nil
}

// BuildSearchMeta constructs pagination metadata for search responses.
func (s *ListingService) BuildSearchMeta(
	filters dto.ListingFilterRequest,
	total int64,
) *response.Meta {
	return buildMeta(filters.Page, filters.PerPage, total)
}

// ── Private Helpers ───────────────────────────────────────────────────────────

// resolveApprovedSellerType checks whether the user has an approved dealer
// or private seller profile, and returns the corresponding SellerType.
func (s *ListingService) resolveApprovedSellerType(userID uuid.UUID) (domain.SellerType, error) {
	// Check dealer profile first
	dealer, err := s.dealerRepo.FindByUserID(userID)
	if err != nil {
		return "", errors.New("failed to validate seller profile")
	}
	if dealer != nil {
		if dealer.ApprovalStatus != domain.ApprovalApproved {
			return "", errors.New("your dealer profile is pending admin approval")
		}
		return domain.SellerTypeDealer, nil
	}

	// Fall back to private seller profile
	seller, err := s.sellerRepo.FindByUserID(userID)
	if err != nil {
		return "", errors.New("failed to validate seller profile")
	}
	if seller != nil {
		if seller.ApprovalStatus != domain.ApprovalApproved {
			return "", errors.New("your private seller profile is pending admin approval")
		}
		return domain.SellerTypePrivate, nil
	}

	return "", errors.New("you must create and have an approved dealer or private seller profile before listing vehicles")
}
