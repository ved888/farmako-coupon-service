package server

import (
	"farmako-coupon-service/handler"

	"github.com/go-chi/chi"
)

func AdminRoutes(admin chi.Router) {
	admin.Post("/coupons", handler.CreateCoupon)
}
