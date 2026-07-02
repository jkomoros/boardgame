package api

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
)

// qrcodeHandler implements GET /api/game/:name/:id/qrcode.png.
//
// Returns a 240×240 PNG QR code encoding the join URL for this game's
// room code. The URL points at the static-frontend host (NOT the api
// host) so a phone scanning the QR lands on the /join page directly.
//
// Self-hosted replacement for the qrserver.com proxy the V1 Table view
// pointed at; eliminates the cross-origin fetch and keeps room codes
// from leaking to a third party. PNG (not SVG) because Safari mobile
// renders the dataURL-served PNG reliably across versions; SVG works in
// modern browsers but the difference is invisible at QR-display sizes
// and PNG has fewer edge cases.
//
// Gated on companion mode (no room code → 404) so a curious user can't
// generate QRs for arbitrary games.
func (s *Server) qrcodeHandler(c *gin.Context) {
	game := s.getGame(c)
	if game == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no such game"})
		return
	}

	eGame, err := s.storage.ExtendedGame(game.ID())
	if err != nil || eGame == nil || eGame.CompanionRoomCode == "" {
		// Not a companion-mode game; no QR to render.
		c.Status(http.StatusNotFound)
		return
	}

	// Build the /join URL the phone should land on. We honor an explicit
	// origin query param so the Table view (which may know its
	// effective origin better than the api binary in some deploys) can
	// supply it. Otherwise we synthesize from the api request's Host
	// header — usually correct but can be wrong behind a reverse proxy.
	origin := c.Query("origin")
	if origin == "" {
		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
		origin = scheme + "://" + c.Request.Host
	}
	joinURL := origin + "/join?code=" + url.QueryEscape(eGame.CompanionRoomCode)

	// qrcode.Low keeps the modules dense and the image small; spec uses
	// 240×240 px (matches the Table view's CSS .room-code-qr size). Low
	// error correction is fine for a code typed-by-hand-as-fallback —
	// damage tolerance is mostly cosmetic.
	png, err := qrcode.Encode(joinURL, qrcode.Low, 240)
	if err != nil {
		s.logger.Warnln("Failed to encode QR:", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	// Aggressive cache: the QR encodes a code that's stable for the
	// lifetime of the game. 5 minutes is generous without making the
	// browser refetch on every Table view re-render.
	c.Header("Cache-Control", "public, max-age=300")
	c.Data(http.StatusOK, "image/png", png)
}
