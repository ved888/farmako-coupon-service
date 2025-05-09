package dbhelper

import (
	"farmako-coupon-service/models"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

func CreateCouponWithTx(tx *sqlx.Tx, coupon *models.Coupon) (string, error) {
	var couponID string

	query := `
		INSERT INTO coupons (
			coupon_code, expiry_date, usage_type, min_order_value, valid_from, valid_to,
			terms_and_conditions, discount_type, discount_value, max_usage_per_user, target
		) VALUES (
			:coupon_code, :expiry_date, :usage_type, :min_order_value, :valid_from, :valid_to,
			:terms_and_conditions, :discount_type, :discount_value, :max_usage_per_user, :target
		) RETURNING id
	`
	rows, err := tx.NamedQuery(query, coupon)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&couponID); err != nil {
			return "", err
		}
	}

	return couponID, nil
}

func InsertCouponApplicableMedicines(tx *sqlx.Tx, couponID string, medicineIDs []string) error {
	for _, medID := range medicineIDs {
		_, err := tx.Exec(`INSERT INTO coupon_applicable_medicines (coupon_id, medicine_id) VALUES ($1, $2)`, couponID, medID)
		if err != nil {
			return err
		}
	}
	return nil
}

func InsertCouponApplicableCategories(tx *sqlx.Tx, couponID string, categories []string) error {
	for _, category := range categories {
		_, err := tx.Exec(`INSERT INTO coupon_applicable_categories (coupon_id, category) VALUES ($1, $2)`, couponID, category)
		if err != nil {
			return err
		}
	}
	return nil
}

func FetchApplicableCoupons(db *sqlx.DB, cartItems []models.CartItem, orderTotal float64, ts time.Time) ([]models.ApplicableCoupon, error) {
	query := `
		SELECT c.coupon_code, c.discount_value
		FROM coupons c
		WHERE c.expiry_date > $1 AND $2 >= c.min_order_value
	`
	rows, err := db.Queryx(query, ts, orderTotal)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var coupons []models.ApplicableCoupon
	for rows.Next() {
		var c models.ApplicableCoupon
		if err := rows.StructScan(&c); err != nil {
			return nil, err
		}
		coupons = append(coupons, c)
	}

	return coupons, nil
}

func ValidateCoupon(db *sqlx.DB, req models.ValidateCouponRequest) (*models.ValidationResult, error) {
	// Validate coupon details (expiry, discount type, value)
	validationResult, err := ValidateCouponDetails(db, req.CouponCode, req.Timestamp)
	if err != nil {
		return nil, err
	}

	// If the coupon is invalid, return the validation result
	if !validationResult.IsValid {
		return validationResult, nil
	}

	// Calculate the items discount based on discount type
	var itemsDiscount float64
	if req.OrderTotal > 0 {
		// Only calculate if the order total is more than zero
		if req.OrderTotal > 0 && validationResult.Discount.ItemsDiscount > 0 {
			if validationResult.Discount.ItemsDiscount > req.OrderTotal {
				itemsDiscount = req.OrderTotal
			} else {
				itemsDiscount = validationResult.Discount.ItemsDiscount
			}
		}
	}

	// Return the final validation result with calculated items discount
	return &models.ValidationResult{
		IsValid: true,
		Message: "coupon applied successfully",
		Discount: models.DiscountBreakdown{
			ItemsDiscount:   itemsDiscount,
			ChargesDiscount: 0,
		},
	}, nil
}

// ValidateCouponDetails will check the validity of the coupon based on the coupon code, expiry date, and discount type.
func ValidateCouponDetails(db *sqlx.DB, couponCode string, timestamp time.Time) (*models.ValidationResult, error) {
	query := `
		SELECT discount_value, discount_type, expiry_date
		FROM coupons
		WHERE coupon_code = $1
	`
	var discountValue float64
	var discountType string
	var expiry time.Time

	err := db.QueryRow(query, couponCode).Scan(&discountValue, &discountType, &expiry)
	if err != nil {
		return nil, fmt.Errorf("coupon not found or expired")
	}

	if timestamp.After(expiry) {
		return &models.ValidationResult{
			IsValid: false,
			Message: "coupon expired or not applicable",
		}, nil
	}

	// Return only necessary fields to validate the coupon
	return &models.ValidationResult{
		IsValid: true,
		Message: "coupon applied successfully",
		Discount: models.DiscountBreakdown{
			ItemsDiscount:   discountValue, // Set this field with discount value
			ChargesDiscount: 0,             // Set ChargesDiscount to 0 for now as no charges discount is given
		},
	}, nil
}

func RecordCouponUsage(db *sqlx.DB, couponCode string, userID string) error {
	// Check if coupon has already been used by this user
	var count int
	err := db.Get(&count, `
		SELECT COUNT(*) FROM coupon_usages
		WHERE coupon_id = (SELECT id FROM coupons WHERE coupon_code = $1) AND user_id = $2
	`, couponCode, userID)
	if err != nil {
		return err
	}

	// Validate against max usage per user
	if count > 0 {
		return fmt.Errorf("max usage limit reached for this user")
	}

	// Insert coupon usage
	_, err = db.Exec(`
		INSERT INTO coupon_usages (coupon_id, user_id)
		VALUES ((SELECT id FROM coupons WHERE coupon_code = $1), $2)
	`, couponCode, userID)
	if err != nil {
		// Handle unique constraint violation (if concurrent insert)
		if strings.Contains(err.Error(), "duplicate key value") {
			return fmt.Errorf("coupon already used by this user")
		}
		return err
	}

	return nil
}
