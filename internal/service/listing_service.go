package service

import (
	"errors"
	"time"

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
	userRepo    interfaces.UserRepository
}

// NewListingService constructs a ListingService with its dependencies.
func NewListingService(
	listingRepo interfaces.ListingRepository,
	dealerRepo interfaces.DealerRepository,
	userRepo interfaces.UserRepository,
) *ListingService {
	return &ListingService{
		listingRepo: listingRepo,
		dealerRepo:  dealerRepo,
		userRepo:    userRepo,
	}
}

// Create creates a new vehicle listing.
// Dealers with an approved profile publish immediately. Buyers must supply
// their identity with the request, then upload an NTSA e-logbook via
// AttachELogbook to submit the listing for admin review.
func (s *ListingService) Create(
	userID uuid.UUID,
	req dto.CreateListingRequest,
) (*dto.ListingResponse, error) {
	// ── 1. Resolve seller type and listing visibility ────────────────────────
	sellerType, err := s.resolveSellerType(userID)
	if err != nil {
		return nil, err
	}

	isDealer := sellerType == domain.SellerTypeDealer

	// Buyers must prove identity and ownership before a listing goes live.
	// Identity comes in with the create call; the NTSA e-logbook file is
	// uploaded afterwards (POST /listings/:id/elogbook) because the upload
	// endpoint needs the listing ID. Until that file lands the listing is a
	// "draft" and stays out of the admin review queue.
	if !isDealer {
		if req.VerificationFullName == "" || req.VerificationIDNumber == "" {
			return nil, errors.New("full name and ID number are required to list a car")
		}
	}

	listingStatus := domain.ListingPending
	verification := domain.VerificationDraft
	if isDealer {
		listingStatus = domain.ListingActive
		verification = domain.VerificationApproved
	} else if req.VerificationELogbook != "" {
		verification = domain.VerificationPending
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
		Status:       listingStatus,
		Images:       []string{},

		VerificationStatus:   verification,
		VerificationFullName: req.VerificationFullName,
		VerificationIDNumber: req.VerificationIDNumber,
		VerificationELogbook: req.VerificationELogbook,
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
		next := domain.ListingStatus(*req.Status)
		// Sellers cannot self-publish: a buyer listing only becomes active once
		// an admin has verified its NTSA e-logbook. Admins keep full control.
		if next == domain.ListingActive &&
			!isAdmin &&
			listing.VerificationStatus != domain.VerificationApproved {
			return nil, errors.New("this listing goes live once an admin approves your e-logbook")
		}
		listing.Status = next
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

// SetCoverImage chooses which uploaded photo represents the listing on cards
// and in search results. Only the owner or an admin may change it, and the
// image has to belong to the listing. An empty URL clears the choice so the
// first image is used again.
func (s *ListingService) SetCoverImage(
	listingID uuid.UUID,
	userID uuid.UUID,
	isAdmin bool,
	imageURL string,
) (*dto.ListingResponse, error) {
	listing, err := s.listingRepo.FindByID(listingID)
	if err != nil || listing == nil {
		return nil, errors.New("listing not found")
	}

	if !isAdmin {
		owned, err := s.listingRepo.BelongsToUser(listingID, userID)
		if err != nil {
			return nil, errors.New("failed to verify ownership")
		}
		if !owned {
			return nil, errors.New("you do not have permission to update this listing")
		}
	}

	if imageURL != "" {
		found := false
		for _, img := range listing.Images {
			if img == imageURL {
				found = true
				break
			}
		}
		if !found {
			return nil, errors.New("that image does not belong to this listing")
		}
	}

	listing.CoverImage = imageURL
	if err := s.listingRepo.Update(listing); err != nil {
		return nil, errors.New("failed to update listing cover image")
	}

	listing, err = s.listingRepo.FindByID(listingID)
	if err != nil || listing == nil {
		return nil, errors.New("listing updated but failed to reload")
	}
	res := dto.ToListingResponse(*listing)
	return &res, nil
}

// AttachELogbook stores the NTSA e-logbook URL against a listing.
// Only the listing owner may attach one, and only while the listing is
// still awaiting verification.
func (s *ListingService) AttachELogbook(listingID, userID uuid.UUID, url string) error {
	listing, err := s.listingRepo.FindByID(listingID)
	if err != nil || listing == nil {
		return errors.New("listing not found")
	}

	owned, err := s.listingRepo.BelongsToUser(listingID, userID)
	if err != nil || !owned {
		return errors.New("you do not have permission to update this listing")
	}

	if listing.VerificationStatus == domain.VerificationApproved {
		return errors.New("this listing has already been approved")
	}
	if listing.VerificationFullName == "" || listing.VerificationIDNumber == "" {
		return errors.New("add your full name and ID number before uploading the e-logbook")
	}

	listing.VerificationELogbook = url
	listing.VerificationStatus = domain.VerificationPending
	listing.RejectionReason = ""
	listing.Status = domain.ListingPending

	return s.listingRepo.Update(listing)
}

// BuildSearchMeta constructs pagination metadata for search responses.
func (s *ListingService) BuildSearchMeta(
	filters dto.ListingFilterRequest,
	total int64,
) *response.Meta {
	return buildMeta(filters.Page, filters.PerPage, total)
}

// AdminReviewListing approves or rejects a pending buyer listing.
// Approving flips it to active so it becomes publicly searchable.
func (s *ListingService) AdminReviewListing(
	listingID uuid.UUID,
	adminID uuid.UUID,
	req dto.AdminReviewListingRequest,
) (*dto.ListingResponse, error) {
	listing, err := s.listingRepo.FindByID(listingID)
	if err != nil || listing == nil {
		return nil, errors.New("listing not found")
	}
	if listing.VerificationStatus == domain.VerificationApproved {
		return nil, errors.New("this listing has already been approved")
	}
	if listing.VerificationELogbook == "" {
		return nil, errors.New("this listing has no NTSA e-logbook to verify")
	}

	if domain.VerificationStatus(req.VerificationStatus) == domain.VerificationRejected {
		if req.RejectionReason == "" {
			return nil, errors.New("a rejection reason is required")
		}
		listing.VerificationStatus = domain.VerificationRejected
		listing.RejectionReason = req.RejectionReason
		listing.Status = domain.ListingInactive
	} else {
		listing.VerificationStatus = domain.VerificationApproved
		listing.RejectionReason = ""
		listing.Status = domain.ListingActive
		now := time.Now()
		listing.VerifiedAt = &now
		listing.VerifiedByID = &adminID
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

// ListPendingListings returns buyer listings awaiting verification.
func (s *ListingService) ListPendingListings(page, perPage int) ([]dto.AdminListingResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 20
	}

	listings, total, err := s.listingRepo.FindByVerificationStatus(
		domain.VerificationPending, page, perPage)
	if err != nil {
		return nil, 0, errors.New("failed to retrieve pending listings")
	}

	out := make([]dto.AdminListingResponse, 0, len(listings))
	for _, l := range listings {
		out = append(out, dto.ToAdminListingResponse(l))
	}
	return out, total, nil
}

// ── Private Helpers ───────────────────────────────────────────────────────────

// resolveSellerType determines how a listing should be attributed.
// Dealers with an approved dealer profile list as "dealer"; everyone else
// (buyers) lists as "private" and goes through verification.
func (s *ListingService) resolveSellerType(userID uuid.UUID) (domain.SellerType, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil || user == nil {
		return "", errors.New("user not found")
	}

	dealer, err := s.dealerRepo.FindByUserID(userID)
	if err != nil {
		return "", errors.New("failed to validate seller profile")
	}
	if dealer != nil && dealer.ApprovalStatus == domain.ApprovalApproved {
		return domain.SellerTypeDealer, nil
	}
	if dealer != nil {
		return "", errors.New("your dealer profile is pending admin approval")
	}

	return domain.SellerTypePrivate, nil
}
