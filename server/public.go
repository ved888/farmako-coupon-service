package server

import (
	"farmako-coupon-service/handler"

	"github.com/go-chi/chi"
)

func PublicRoutes(public chi.Router) {
	// Public coupon routes
	public.Post("/coupons/applicable", handler.GetApplicableCoupons)
	public.Post("/coupons/validate", handler.ValidateCoupon)
}
