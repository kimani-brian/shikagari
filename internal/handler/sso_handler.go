package handler

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/shikagari/api/pkg/response"
	"github.com/shikagari/api/pkg/sso"
)

// SSOHandler mints auction handoff tickets for the caller.
type SSOHandler struct {
	minter *sso.Minter
}

// NewSSOHandler builds the ticket endpoint. A nil minter (empty secret)
// disables minting entirely rather than signing with nothing.
func NewSSOHandler(secret string) *SSOHandler {
	minter, err := sso.NewMinter(secret)
	if err != nil {
		return &SSOHandler{minter: nil}
	}
	return &SSOHandler{minter: minter}
}

// MintAuctionTicket issues a one-time ticket carrying the caller's identity
// into the auction engine. The caller is already authenticated by middleware;
// the ticket simply re-attests who they are for a different service.
func (h *SSOHandler) MintAuctionTicket(c *gin.Context) {
	if h.minter == nil {
		response.InternalServerError(c, "Auction handoff is not configured.")
		return
	}
	claims := mustGetClaims(c)

	raw, issued, err := h.minter.Mint(
		claims.UserID.String(), string(claims.Role), claims.Email, claims.IsVerified, time.Now(),
	)
	if err != nil {
		response.InternalServerError(c, "Could not issue handoff ticket.")
		return
	}
	response.OK(c, "Handoff ticket issued.", gin.H{
		"ticket":     raw,
		"expires_at": issued.Expiry,
	})
}
