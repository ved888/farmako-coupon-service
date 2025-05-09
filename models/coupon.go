package models

import "time"

type Coupon struct {
	ID                    string    `json:"id" db:"id"`
	CouponCode            string    `json:"coupon_code" db:"coupon_code"`
	ExpiryDate            time.Time `json:"expiry_date" db:"expiry_date"`
	UsageType             string    `json:"usage_type" db:"usage_type"`
	ApplicableMedicineIDs []string  `json:"applicable_medicine_ids"`
	ApplicableCategories  []string  `json:"applicable_categories"`
	MinOrderValue         float64   `json:"min_order_value" db:"min_order_value"`
	ValidFrom             time.Time `json:"valid_from" db:"valid_from"`
	ValidTo               time.Time `json:"valid_to" db:"valid_to"`
	Terms                 string    `json:"terms_and_conditions" db:"terms_and_conditions"`
	DiscountType          string    `json:"discount_type" db:"discount_type"`
	DiscountValue         float64   `json:"discount_value" db:"discount_value"`
	MaxUsagePerUser       int       `json:"max_usage_per_user" db:"max_usage_per_user"`
	Target                string    `json:"target" db:"target"`
}

type CartItem struct {
	ID       string `json:"id"`
	Category string `json:"category"`
}

type ApplicableCoupon struct {
	CouponCode    string  `json:"coupon_code" db:"coupon_code"`
	DiscountValue float64 `json:"discount_value" db:"discount_value"`
}

type ValidateCouponRequest struct {
	CouponCode string     `json:"coupon_code" db:"coupon_code"`
	UserID     string     `json:"user_id db:"user_id`
	CartItems  []CartItem `json:"cart_items" db:"cart_item"`
	OrderTotal float64    `json:"order_total" db:"order_total"`
	Timestamp  time.Time  `json:"timestamp" db:"timestamp"`
}

type DiscountBreakdown struct {
	ItemsDiscount   float64 `json:"items_discount"`
	ChargesDiscount float64 `json:"charges_discount"`
}

type ValidationResult struct {
	IsValid  bool              `json:"is_valid"`
	Discount DiscountBreakdown `json:"discount"`
	Message  string            `json:"message"`
}
