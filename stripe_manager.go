package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/price"
	"github.com/stripe/stripe-go/v76/product"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/subscription"
	"github.com/stripe/stripe-go/v76/webhookendpoint"
)

func initStripeManager() {
	stripe.Key = os.Getenv("STRIPE_API_KEY")
}

// ──────────────────────────────────────────────────────────────────────────────
// Products
// ──────────────────────────────────────────────────────────────────────────────

func handleListProducts(w http.ResponseWriter, r *http.Request) {
	params := &stripe.ProductListParams{}
	params.Limit = 100
	result := product.List(params)

	var products []map[string]interface{}
	for result.Next() {
		p := result.Product()
		products = append(products, map[string]interface{}{
			"id":          p.ID,
			"name":        p.Name,
			"description": p.Description,
			"active":      p.Active,
			"created":     p.Created,
			"images":      p.Images,
			"metadata":    p.Metadata,
		})
	}
	if products == nil {
		products = []map[string]interface{}{}
	}
	json.NewEncoder(w).Encode(products)
}

func handleCreateProduct(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string            `json:"name"`
		Description string            `json:"description"`
		Price       int64             `json:"price"` // in cents
		Currency    string            `json:"currency"`
		Mode        string            `json:"mode"` // "one_time" or "recurring"
		Interval    string            `json:"interval"` // "month", "year", etc.
		Metadata    map[string]string `json:"metadata"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create the product
	productParams := &stripe.ProductParams{
		Name:        stripe.String(req.Name),
		Description: stripe.String(req.Description),
	}
	if req.Metadata != nil {
		for k, v := range req.Metadata {
			productParams.AddMetadata(k, v)
		}
	}
	p, err := product.New(productParams)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create product: %v", err), http.StatusInternalServerError)
		return
	}

	// Create the price if provided
	var pr *stripe.Price
	if req.Price > 0 {
		priceParams := &stripe.PriceParams{
			Product:    stripe.String(p.ID),
			UnitAmount: stripe.Int64(req.Price),
			Currency:   stripe.String(string(stripe.CurrencyUSD)),
		}
		if req.Currency != "" {
			priceParams.Currency = stripe.String(req.Currency)
		}
		if req.Mode == "recurring" {
			priceParams.Recurring = &stripe.PriceRecurringParams{
				Interval: stripe.String(req.Interval),
			}
		}
		pr, err = price.New(priceParams)
		if err != nil {
			http.Error(w, fmt.Sprintf("Product created but price failed: %v", err), http.StatusInternalServerError)
			return
		}
	}

	result := map[string]interface{}{
		"product": map[string]interface{}{
			"id":          p.ID,
			"name":        p.Name,
			"description": p.Description,
			"active":      p.Active,
		},
	}
	if pr != nil {
		result["price"] = map[string]interface{}{
			"id":         pr.ID,
			"unit_amount": pr.UnitAmount,
			"currency":   pr.Currency,
			"recurring":  pr.Recurring != nil,
		}
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

func handleGetProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	p, err := product.Get(id, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("Product not found: %v", err), http.StatusNotFound)
		return
	}

	// Get prices for this product
	params := &stripe.PriceListParams{}
	params.Product = stripe.String(id)
	priceResult := price.List(params)

	var prices []map[string]interface{}
	for priceResult.Next() {
		pr := priceResult.Price()
		prices = append(prices, map[string]interface{}{
			"id":          pr.ID,
			"unit_amount": pr.UnitAmount,
			"currency":    pr.Currency,
			"active":      pr.Active,
			"recurring":   pr.Recurring != nil,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":          p.ID,
		"name":        p.Name,
		"description": p.Description,
		"active":      p.Active,
		"created":     p.Created,
		"images":      p.Images,
		"metadata":    p.Metadata,
		"prices":      prices,
	})
}

func handleUpdateProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Active      *bool  `json:"active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	params := &stripe.ProductParams{}
	if req.Name != "" {
		params.Name = stripe.String(req.Name)
	}
	if req.Description != "" {
		params.Description = stripe.String(req.Description)
	}
	if req.Active != nil {
		params.Active = stripe.Bool(*req.Active)
	}

	p, err := product.Update(id, params)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update product: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":          p.ID,
		"name":        p.Name,
		"description": p.Description,
		"active":      p.Active,
	})
}

func handleDeleteProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := product.Del(id, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete product: %v", err), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted", "id": id})
}

// ──────────────────────────────────────────────────────────────────────────────
// Customers
// ──────────────────────────────────────────────────────────────────────────────

func handleListCustomers(w http.ResponseWriter, r *http.Request) {
	params := &stripe.CustomerListParams{}
	params.Limit = 100
	result := customer.List(params)

	var customers []map[string]interface{}
	for result.Next() {
		c := result.Customer()
		customers = append(customers, map[string]interface{}{
			"id":    c.ID,
			"email": c.Email,
			"name":  c.Name,
			"phone": c.Phone,
			"created": c.Created,
		})
	}
	if customers == nil {
		customers = []map[string]interface{}{}
	}
	json.NewEncoder(w).Encode(customers)
}

func handleGetCustomer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	c, err := customer.Get(id, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("Customer not found: %v", err), http.StatusNotFound)
		return
	}

	// Get subscriptions for this customer
	subParams := &stripe.SubscriptionListParams{}
	subParams.Customer = stripe.String(id)
	subResult := subscription.List(subParams)

	var subscriptions []map[string]interface{}
	for subResult.Next() {
		s := subResult.Subscription()
		subscriptions = append(subscriptions, map[string]interface{}{
			"id":     s.ID,
			"status": s.Status,
			"created": s.Created,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":            c.ID,
		"email":         c.Email,
		"name":          c.Name,
		"phone":         c.Phone,
		"created":       c.Created,
		"subscriptions": subscriptions,
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// Subscriptions
// ──────────────────────────────────────────────────────────────────────────────

func handleListSubscriptions(w http.ResponseWriter, r *http.Request) {
	params := &stripe.SubscriptionListParams{}
	params.Limit = 100
	result := subscription.List(params)

	var subs []map[string]interface{}
	for result.Next() {
		s := result.Subscription()
		subs = append(subs, map[string]interface{}{
			"id":      s.ID,
			"status":  s.Status,
			"current_period_end": s.CurrentPeriodEnd,
			"created": s.Created,
		})
	}
	if subs == nil {
		subs = []map[string]interface{}{}
	}
	json.NewEncoder(w).Encode(subs)
}

func handleCancelSubscription(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(true),
	}
	s, err := sub.Update(id, params)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to cancel subscription: %v", err), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":     s.ID,
		"status": s.Status,
		"cancel_at_period_end": s.CancelAtPeriodEnd,
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// Payments (Payment Intents)
// ──────────────────────────────────────────────────────────────────────────────

func handleListPayments(w http.ResponseWriter, r *http.Request) {
	params := &stripe.PaymentIntentListParams{}
	params.Limit = 100
	result := paymentintent.List(params)

	var payments []map[string]interface{}
	for result.Next() {
		pi := result.PaymentIntent()
		payments = append(payments, map[string]interface{}{
			"id":       pi.ID,
			"amount":   pi.Amount,
			"currency": pi.Currency,
			"status":   pi.Status,
			"created":  pi.Created,
		})
	}
	if payments == nil {
		payments = []map[string]interface{}{}
	}
	json.NewEncoder(w).Encode(payments)
}

// ──────────────────────────────────────────────────────────────────────────────
// Checkout Sessions
// ──────────────────────────────────────────────────────────────────────────────

func handleCreateCheckout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PriceID      string `json:"price_id"`
		CustomerEmail string `json:"customer_email"`
		SuccessURL   string `json:"success_url"`
		CancelURL    string `json:"cancel_url"`
		Mode         string `json:"mode"` // "payment", "subscription", "setup"
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.PriceID == "" || req.SuccessURL == "" || req.CancelURL == "" {
		http.Error(w, "price_id, success_url, and cancel_url are required", http.StatusBadRequest)
		return
	}

	mode := stripe.CheckoutSessionModePayment
	if req.Mode == "subscription" {
		mode = stripe.CheckoutSessionModeSubscription
	}

	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(req.PriceID),
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(mode)),
		SuccessURL: stripe.String(req.SuccessURL),
		CancelURL:  stripe.String(req.CancelURL),
	}

	if req.CustomerEmail != "" {
		params.CustomerEmail = stripe.String(req.CustomerEmail)
	}

	s, err := session.New(params)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create checkout session: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         s.ID,
		"url":        s.URL,
		"status":     s.Status,
		"mode":       s.Mode,
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// Webhooks
// ──────────────────────────────────────────────────────────────────────────────

func handleListWebhooks(w http.ResponseWriter, r *http.Request) {
	params := &stripe.WebhookEndpointListParams{}
	params.Limit = 100
	result := webhookendpoint.List(params)

	var webhooks []map[string]interface{}
	for result.Next() {
		we := result.WebhookEndpoint()
		webhooks = append(webhooks, map[string]interface{}{
			"id":      we.ID,
			"url":     we.URL,
			"status":  we.Status,
			"enabled_events": we.EnabledEvents,
			"created": we.Created,
		})
	}
	if webhooks == nil {
		webhooks = []map[string]interface{}{}
	}
	json.NewEncoder(w).Encode(webhooks)
}

func handleCreateWebhook(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL           string   `json:"url"`
		EnabledEvents []string `json:"enabled_events"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}
	if len(req.EnabledEvents) == 0 {
		req.EnabledEvents = []string{"checkout.session.completed", "invoice.payment_succeeded", "customer.subscription.deleted"}
	}

	params := &stripe.WebhookEndpointParams{
		URL:           stripe.String(req.URL),
		EnabledEvents: stripe.StringSlice(req.EnabledEvents),
	}
	we, err := webhookendpoint.New(params)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create webhook: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":            we.ID,
		"url":           we.URL,
		"status":        we.Status,
		"secret":        we.Secret,
		"enabled_events": we.EnabledEvents,
	})
}

func handleDeleteWebhook(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := webhookendpoint.Del(id, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete webhook: %v", err), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted", "id": id})
}

// ──────────────────────────────────────────────────────────────────────────────
// Revenue Summary
// ──────────────────────────────────────────────────────────────────────────────

func handleRevenueSummary(w http.ResponseWriter, r *http.Request) {
	// Count products
	productParams := &stripe.ProductListParams{}
	productParams.Limit = 100
	productResult := product.List(productParams)
	productCount := 0
	for productResult.Next() {
		productCount++
	}

	// Count customers
	customerParams := &stripe.CustomerListParams{}
	customerParams.Limit = 100
	customerResult := customer.List(customerParams)
	customerCount := 0
	for customerResult.Next() {
		customerCount++
	}

	// Count active subscriptions
	subParams := &stripe.SubscriptionListParams{}
	subParams.Limit = 100
	subResult := subscription.List(subParams)
	activeSubscriptions := 0
	totalMRR := int64(0)
	for subResult.Next() {
		s := subResult.Subscription()
		if s.Status == stripe.SubscriptionStatusActive {
			activeSubscriptions++
			if len(s.Items.Data) > 0 {
				totalMRR += s.Items.Data[0].Price.UnitAmount
			}
		}
	}

	// Count payments
	paymentParams := &stripe.PaymentIntentListParams{}
	paymentParams.Limit = 100
	paymentResult := paymentintent.List(paymentParams)
	paymentCount := 0
	totalRevenue := int64(0)
	for paymentResult.Next() {
		pi := paymentResult.PaymentIntent()
		paymentCount++
		if pi.Status == stripe.PaymentIntentStatusSucceeded {
			totalRevenue += pi.Amount
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"products":            productCount,
		"customers":           customerCount,
		"active_subscriptions": activeSubscriptions,
		"total_mrr":           totalMRR,
		"total_revenue":       totalRevenue,
		"total_payments":      paymentCount,
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// Route Registration
// ──────────────────────────────────────────────────────────────────────────────

func registerStripeManagerRoutes(r chi.Router) {
	initStripeManager()

	r.Route("/admin/stripe", func(r chi.Router) {
		r.Use(corsMiddleware)
		r.Use(authMiddleware)

		// Products
		r.Get("/products", handleListProducts)
		r.Post("/products", handleCreateProduct)
		r.Get("/products/{id}", handleGetProduct)
		r.Put("/products/{id}", handleUpdateProduct)
		r.Delete("/products/{id}", handleDeleteProduct)

		// Customers
		r.Get("/customers", handleListCustomers)
		r.Get("/customers/{id}", handleGetCustomer)

		// Subscriptions
		r.Get("/subscriptions", handleListSubscriptions)
		r.Post("/subscriptions/{id}/cancel", handleCancelSubscription)

		// Payments
		r.Get("/payments", handleListPayments)

		// Checkout
		r.Post("/checkout", handleCreateCheckout)

		// Webhooks
		r.Get("/webhooks", handleListWebhooks)
		r.Post("/webhooks", handleCreateWebhook)
		r.Delete("/webhooks/{id}", handleDeleteWebhook)

		// Revenue
		r.Get("/revenue", handleRevenueSummary)
	})

	log.Println("[Stripe Manager] Routes registered at /admin/stripe/*")
}
