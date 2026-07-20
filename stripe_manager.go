package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/price"
	"github.com/stripe/stripe-go/v76/product"
)

// ──────────────────────────────────────────────────────────────────────────────
// Create Product + Price
// ──────────────────────────────────────────────────────────────────────────────

func handleCreateProduct(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Price       int64  `json:"price"` // in cents
		Mode        string `json:"mode"`  // "one_time" or "recurring"
		Interval    string `json:"interval"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	// Create the product
	p, err := product.New(&stripe.ProductParams{
		Name:        stripe.String(req.Name),
		Description: stripe.String(req.Description),
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("product create failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Create the price
	priceParams := &stripe.PriceParams{
		Product:    stripe.String(p.ID),
		UnitAmount: stripe.Int64(req.Price),
		Currency:   stripe.String("usd"),
	}
	if req.Mode == "recurring" && req.Interval != "" {
		priceParams.Recurring = &stripe.PriceRecurringParams{
			Interval: stripe.String(req.Interval),
		}
	}

	pr, err := price.New(priceParams)
	if err != nil {
		http.Error(w, fmt.Sprintf("price create failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"product_id": p.ID,
		"price_id":   pr.ID,
		"name":       p.Name,
		"amount":     pr.UnitAmount,
		"currency":   pr.Currency,
		"mode":       req.Mode,
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// Checkout Session
// ──────────────────────────────────────────────────────────────────────────────

func handleCreateCheckout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PriceID      string `json:"price_id"`
		CustomerEmail string `json:"customer_email"`
		SuccessURL   string `json:"success_url"`
		CancelURL    string `json:"cancel_url"`
		Mode         string `json:"mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.PriceID == "" {
		http.Error(w, "price_id is required", http.StatusBadRequest)
		return
	}

	if req.SuccessURL == "" {
		req.SuccessURL = "https://koola10.ai/thanks"
	}
	if req.CancelURL == "" {
		req.CancelURL = "https://koola10.ai/pricing"
	}

	mode := stripe.CheckoutSessionModePayment
	if req.Mode == "subscription" {
		mode = stripe.CheckoutSessionModeSubscription
	}

	modeStr := string(mode)

	sess, err := session.New(&stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(req.PriceID),
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       &modeStr,
		SuccessURL: stripe.String(req.SuccessURL),
		CancelURL:  stripe.String(req.CancelURL),
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("checkout create failed: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         sess.ID,
		"url":        sess.URL,
		"status":     sess.Status,
		"mode":       sess.Mode,
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// Route Registration
// ──────────────────────────────────────────────────────────────────────────────

func registerStripeManagerRoutes(r chi.Router) {
	r.Route("/admin/stripe", func(r chi.Router) {
		r.Use(corsMiddleware)
		r.Use(authMiddleware)

		r.Post("/products", handleCreateProduct)
		r.Post("/checkout", handleCreateCheckout)
	})
}
