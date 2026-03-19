package dto

// ── Response DTOs ─────────────────────────────────────────────────────────────

// UploadImageResponse is returned after a successful image upload.
type UploadImageResponse struct {
	URL      string `json:"url"`
	FileName string `json:"file_name"`
	SizeKB   int64  `json:"size_kb"`
}

// UploadLogoResponse is returned after a dealer logo upload.
type UploadLogoResponse struct {
	LogoURL string `json:"logo_url"`
}

// UploadProfilePhotoResponse is returned after a private seller photo upload.
type UploadProfilePhotoResponse struct {
	ProfilePhotoURL string `json:"profile_photo_url"`
}
