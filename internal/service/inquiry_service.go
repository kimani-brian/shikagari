package service

import (
	"errors"

	"github.com/google/uuid"
	"github.com/shikagari/api/internal/domain"
	"github.com/shikagari/api/internal/dto"
	"github.com/shikagari/api/internal/repository/interfaces"
)

// InquiryService handles buyer-seller messaging around active listings.
type InquiryService struct {
	inquiryRepo interfaces.InquiryRepository
	listingRepo interfaces.ListingRepository
}

// NewInquiryService constructs an InquiryService with its dependencies.
func NewInquiryService(
	inquiryRepo interfaces.InquiryRepository,
	listingRepo interfaces.ListingRepository,
) *InquiryService {
	return &InquiryService{
		inquiryRepo: inquiryRepo,
		listingRepo: listingRepo,
	}
}

// Send creates a new inquiry from a buyer to a listing seller.
func (s *InquiryService) Send(
	buyerID uuid.UUID,
	listingID uuid.UUID,
	req dto.CreateInquiryRequest,
) (*dto.InquiryResponse, error) {
	listing, err := s.listingRepo.FindByID(listingID)
	if err != nil {
		return nil, errors.New("failed to retrieve listing")
	}
	if listing == nil {
		return nil, errors.New("listing not found")
	}
	if listing.Status != domain.ListingActive {
		return nil, errors.New("this listing is no longer active")
	}
	if listing.UserID == buyerID {
		return nil, errors.New("you cannot send an inquiry to your own listing")
	}

	inquiry := &domain.Inquiry{
		ListingID: listing.ID,
		BuyerID:   buyerID,
		SellerID:  listing.UserID,
		Message:   req.Message,
		Status:    domain.InquiryStatusOpen,
	}
	if err := s.inquiryRepo.Create(inquiry); err != nil {
		return nil, errors.New("failed to create inquiry")
	}

	created, err := s.inquiryRepo.FindByID(inquiry.ID)
	if err != nil {
		return nil, errors.New("failed to retrieve created inquiry")
	}
	if created == nil {
		return nil, errors.New("created inquiry could not be found")
	}

	result := dto.ToInquiryResponse(*created)
	return &result, nil
}

// GetByID returns a single inquiry if the caller is authorized to view it.
func (s *InquiryService) GetByID(
	inquiryID uuid.UUID,
	userID uuid.UUID,
	isAdmin bool,
) (*dto.InquiryResponse, error) {
	inquiry, err := s.inquiryRepo.FindByID(inquiryID)
	if err != nil {
		return nil, errors.New("failed to retrieve inquiry")
	}
	if inquiry == nil {
		return nil, errors.New("inquiry not found")
	}

	if !isAdmin && inquiry.BuyerID != userID && inquiry.SellerID != userID {
		return nil, errors.New("you are not allowed to view this inquiry")
	}

	result := dto.ToInquiryResponse(*inquiry)
	return &result, nil
}

// GetMyInquiries returns inquiries sent by the authenticated buyer.
func (s *InquiryService) GetMyInquiries(
	buyerID uuid.UUID,
	page, perPage int,
) ([]dto.InquirySummary, int64, error) {
	inquiries, total, err := s.inquiryRepo.FindByBuyerID(buyerID, page, perPage)
	if err != nil {
		return nil, 0, errors.New("failed to retrieve inquiries")
	}

	result := make([]dto.InquirySummary, 0, len(inquiries))
	for _, inquiry := range inquiries {
		result = append(result, dto.ToInquirySummary(inquiry))
	}

	return result, total, nil
}

// GetMyInbox returns inquiries received by the authenticated seller.
func (s *InquiryService) GetMyInbox(
	sellerID uuid.UUID,
	page, perPage int,
) ([]dto.InquirySummary, int64, error) {
	inquiries, total, err := s.inquiryRepo.FindBySellerID(sellerID, page, perPage)
	if err != nil {
		return nil, 0, errors.New("failed to retrieve inbox")
	}

	result := make([]dto.InquirySummary, 0, len(inquiries))
	for _, inquiry := range inquiries {
		result = append(result, dto.ToInquirySummary(inquiry))
	}

	return result, total, nil
}

// Reply records a seller response on an inquiry and marks it replied.
func (s *InquiryService) Reply(
	inquiryID uuid.UUID,
	sellerID uuid.UUID,
	req dto.ReplyInquiryRequest,
) (*dto.InquiryResponse, error) {
	inquiry, err := s.inquiryRepo.FindByID(inquiryID)
	if err != nil {
		return nil, errors.New("failed to retrieve inquiry")
	}
	if inquiry == nil {
		return nil, errors.New("inquiry not found")
	}
	if inquiry.SellerID != sellerID {
		return nil, errors.New("you can only reply to inquiries addressed to you")
	}
	if inquiry.Status == domain.InquiryStatusClosed {
		return nil, errors.New("closed inquiries cannot be replied to")
	}

	if err := s.inquiryRepo.UpdateReply(inquiryID, req.Reply); err != nil {
		return nil, errors.New("failed to save reply")
	}

	updated, err := s.inquiryRepo.FindByID(inquiryID)
	if err != nil {
		return nil, errors.New("failed to retrieve updated inquiry")
	}
	if updated == nil {
		return nil, errors.New("updated inquiry could not be found")
	}

	result := dto.ToInquiryResponse(*updated)
	return &result, nil
}

// UpdateStatus changes the seller-visible status of an inquiry.
func (s *InquiryService) UpdateStatus(
	inquiryID uuid.UUID,
	sellerID uuid.UUID,
	req dto.UpdateInquiryStatusRequest,
) (*dto.InquiryResponse, error) {
	inquiry, err := s.inquiryRepo.FindByID(inquiryID)
	if err != nil {
		return nil, errors.New("failed to retrieve inquiry")
	}
	if inquiry == nil {
		return nil, errors.New("inquiry not found")
	}
	if inquiry.SellerID != sellerID {
		return nil, errors.New("you can only update inquiries addressed to you")
	}

	status := domain.InquiryStatus(req.Status)
	switch status {
	case domain.InquiryStatusOpen, domain.InquiryStatusReplied, domain.InquiryStatusClosed:
	default:
		return nil, errors.New("invalid inquiry status")
	}

	if err := s.inquiryRepo.UpdateStatus(inquiryID, status); err != nil {
		return nil, errors.New("failed to update inquiry status")
	}

	updated, err := s.inquiryRepo.FindByID(inquiryID)
	if err != nil {
		return nil, errors.New("failed to retrieve updated inquiry")
	}
	if updated == nil {
		return nil, errors.New("updated inquiry could not be found")
	}

	result := dto.ToInquiryResponse(*updated)
	return &result, nil
}
