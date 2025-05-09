package handler

import (
	"farmako-coupon-service/cache"
	"farmako-coupon-service/database"
	"farmako-coupon-service/dbhelper"
	"farmako-coupon-service/models"
	"farmako-coupon-service/utils"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	pcache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
)

// CreateCoupon godoc
//
//	@Summary		Create a new coupon
//	@Description	Admin creates a coupon with the required fields.
//	@Tags			Admin
//	@Param			coupon	body	models.Coupon	true	"Coupon Payload"
//	@Accept			json
//	@Produce		json
//	@Success		201
//	@Failure		400
//	@Failure		500
//	@Router			/v1/admin/coupons   [post]
func CreateCoupon(w http.ResponseWriter, r *http.Request) {
	var coupon models.Coupon
	if err := utils.ParseBody(r.Body, &coupon); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	txErr := database.Tx(func(tx *sqlx.Tx) error {
		// Insert core coupon
		couponID, err := dbhelper.CreateCouponWithTx(tx, &coupon)
		if err != nil {
			return errors.Wrapf(err, "CreateCoupon: Failed to create coupon")
		}
		// Insert medicines
		if err := dbhelper.InsertCouponApplicableMedicines(tx, couponID, coupon.ApplicableMedicineIDs); err != nil {
			return errors.Wrapf(err, "CreateCoupon: Failed to insert applicable medicines")

		}

		// Insert categories
		if err := dbhelper.InsertCouponApplicableCategories(tx, couponID, coupon.ApplicableCategories); err != nil {
			return errors.Wrapf(err, "CreateCoupon: Failed to insert applicable categories")
		}

		utils.RespondJSON(w, http.StatusCreated, map[string]string{"coupon_id": couponID})
		return nil
	})
	if txErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, txErr,
			"CreateCoupon: failed to create entry for the coupon",
		)
		return
	}
}

// GetApplicableCoupons godoc
// @Summary            Get applicable coupons
// @Description        Returns applicable coupons based on order/cart
// @Tags               Public
// @Accept             json
// @Produce            json
// @Param              request        body   models.ValidateCouponRequest   true "Order info"
// @Success            200
// @Failure            400
// @Router             /v1/public/coupons/applicable [post]
func GetApplicableCoupons(w http.ResponseWriter, r *http.Request) {
	var req models.ValidateCouponRequest
	if err := utils.ParseBody(r.Body, &req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	cacheKey := fmt.Sprintf("coupons-%v-%v", req.OrderTotal, req.Timestamp.Unix())
	if cached, found := cache.CouponCache.Get(cacheKey); found {
		coupons := cached.([]models.ApplicableCoupon)
		utils.RespondJSON(w, http.StatusOK, map[string]interface{}{"applicable_coupons": coupons})
		return
	}

	coupons, err := dbhelper.FetchApplicableCoupons(database.FCS, req.CartItems, req.OrderTotal, req.Timestamp)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch applicable coupons")
		return
	}

	cache.CouponCache.Set(cacheKey, coupons, pcache.DefaultExpiration)
	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{"applicable_coupons": coupons})
}

// ValidateCoupon godoc
// @Summary               Validate a coupon
// @Description           Validates a coupon code against a cart and returns the discount
// @Tags                  Coupons
// @Accept                json
// @Produce               json
// @Param                 request    body      models.ValidateCouponRequest   true "Coupon validation request"
// @Success               200        {object}  models.ValidationResult
// @Failure               400
// @Router                /v1/public/coupons/validate [post]
func ValidateCoupon(w http.ResponseWriter, r *http.Request) {
	var req models.ValidateCouponRequest
	if err := utils.ParseBody(r.Body, &req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	// Channel to receive validation result
	resultChan := make(chan *models.ValidationResult)
	errorChan := make(chan error)
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		result, err := dbhelper.ValidateCoupon(database.FCS, req)
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- result
	}()

	// Close channels once goroutine finishes
	go func() {
		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()

	// Wait for the result or error
	select {
	case res := <-resultChan:
		// Start a transaction for the coupon usage check and insert
		tx, err := database.FCS.Beginx()
		if err != nil {
			utils.RespondError(w, http.StatusInternalServerError, err, "Failed to start transaction")
			return
		}
		defer tx.Rollback()

		// 1. Check if the coupon is valid and hasn't been used already (record usage)
		err = dbhelper.RecordCouponUsage(database.FCS, req.CouponCode, req.UserID)
		if err != nil {
			utils.RespondError(w, http.StatusConflict, err, "Failed to apply coupon")
			return
		}

		// Commit the transaction if everything is fine
		err = tx.Commit()
		if err != nil {
			utils.RespondError(w, http.StatusInternalServerError, err, "Failed to commit transaction")
			return
		}

		// Respond with validation result
		utils.RespondJSON(w, http.StatusOK, res)

	case err := <-errorChan:
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to validate coupon")
	case <-time.After(2 * time.Second):
		utils.RespondError(w, http.StatusRequestTimeout, fmt.Errorf("timeout"), "Validation took too long")
	}
}

func ValidateCoupon0(w http.ResponseWriter, r *http.Request) {
	var req models.ValidateCouponRequest
	if err := utils.ParseBody(r.Body, &req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	// Channel to receive validation result
	resultChan := make(chan *models.ValidationResult)
	errorChan := make(chan error)
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		result, err := dbhelper.ValidateCoupon(database.FCS, req)
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- result
	}()

	// Close channels once goroutine finishes
	go func() {
		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()

	// Wait for the result or error
	select {
	case res := <-resultChan:
		utils.RespondJSON(w, http.StatusOK, res)
	case err := <-errorChan:
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to validate coupon")
	case <-time.After(2 * time.Second):
		utils.RespondError(w, http.StatusRequestTimeout, fmt.Errorf("timeout"), "Validation took too long")
	}
}
