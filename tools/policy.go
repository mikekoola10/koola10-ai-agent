package tools

import (
	"fmt"
)

// RiskLevel represents the severity of an action
type RiskLevel string

const (
	GREEN  RiskLevel = "GREEN"  // Safe, read-only operations
	YELLOW RiskLevel = "YELLOW" // Requires caution, write operations with moderate impact
	RED    RiskLevel = "RED"    // High-risk, potential for data loss or security issues
)

// StripeActionType identifies specific Stripe actions
type StripeActionType string

const (
	// GREEN - Read-only operations
	ListProducts       StripeActionType = "list_products"
	ListCustomers      StripeActionType = "list_customers"
	ListSubscriptions  StripeActionType = "list_subscriptions"
	ListPayments       StripeActionType = "list_payments"
	GetProduct         StripeActionType = "get_product"
	GetCustomer        StripeActionType = "get_customer"
	GetSubscription    StripeActionType = "get_subscription"
	GetPayment         StripeActionType = "get_payment"

	// YELLOW - Moderate-risk write operations
	CreateCheckoutSession StripeActionType = "create_checkout_session"
	CreateProduct         StripeActionType = "create_product"
	CancelSubscription    StripeActionType = "cancel_subscription"

	// RED - High-risk operations (blocked)
	CreateRefund      StripeActionType = "create_refund"
	TransferFunds     StripeActionType = "transfer_funds"
	DeleteCustomer    StripeActionType = "delete_customer"
	ExposeSecretKey   StripeActionType = "expose_secret_key"
	UpdatePaymentInfo StripeActionType = "update_payment_info"
)

// PolicyDecision represents the result of a policy evaluation
type PolicyDecision struct {
	Action      string    `json:"action"`
	RiskLevel   RiskLevel `json:"risk_level"`
	Allowed     bool      `json:"allowed"`
	Reason      string    `json:"reason"`
	RequiresLog bool      `json:"requires_log"`
}

// EvaluateStripeAction evaluates the risk level of a Stripe action
func EvaluateStripeAction(action string) PolicyDecision {
	stripeAction := StripeActionType(action)

	// GREEN - Read-only operations (always allowed)
	switch stripeAction {
	case ListProducts, ListCustomers, ListSubscriptions, ListPayments,
		GetProduct, GetCustomer, GetSubscription, GetPayment:
		return PolicyDecision{
			Action:      action,
			RiskLevel:   GREEN,
			Allowed:     true,
			Reason:      "Read-only operation",
			RequiresLog: false,
		}
	}

	// YELLOW - Moderate-risk operations (allowed but logged)
	switch stripeAction {
	case CreateCheckoutSession:
		return PolicyDecision{
			Action:      action,
			RiskLevel:   YELLOW,
			Allowed:     true,
			Reason:      "Checkout session creation requires audit trail",
			RequiresLog: true,
		}
	case CreateProduct:
		return PolicyDecision{
			Action:      action,
			RiskLevel:   YELLOW,
			Allowed:     true,
			Reason:      "Product creation requires audit trail",
			RequiresLog: true,
		}
	case CancelSubscription:
		return PolicyDecision{
			Action:      action,
			RiskLevel:   YELLOW,
			Allowed:     true,
			Reason:      "Subscription cancellation requires audit trail",
			RequiresLog: true,
		}
	}

	// RED - High-risk operations (blocked)
	switch stripeAction {
	case CreateRefund:
		return PolicyDecision{
			Action:      action,
			RiskLevel:   RED,
			Allowed:     false,
			Reason:      "Refund operations require explicit approval",
			RequiresLog: true,
		}
	case TransferFunds:
		return PolicyDecision{
			Action:      action,
			RiskLevel:   RED,
			Allowed:     false,
			Reason:      "Fund transfers blocked - requires manual review",
			RequiresLog: true,
		}
	case DeleteCustomer:
		return PolicyDecision{
			Action:      action,
			RiskLevel:   RED,
			Allowed:     false,
			Reason:      "Customer deletion blocked - irreversible operation",
			RequiresLog: true,
		}
	case ExposeSecretKey:
		return PolicyDecision{
			Action:      action,
			RiskLevel:   RED,
			Allowed:     false,
			Reason:      "Secret key exposure blocked - security violation",
			RequiresLog: true,
		}
	case UpdatePaymentInfo:
		return PolicyDecision{
			Action:      action,
			RiskLevel:   RED,
			Allowed:     false,
			Reason:      "Payment info updates blocked - PCI compliance concern",
			RequiresLog: true,
		}
	}

	// Unknown action defaults to RED
	return PolicyDecision{
		Action:      action,
		RiskLevel:   RED,
		Allowed:     false,
		Reason:      fmt.Sprintf("Unknown or unclassified action: %s", action),
		RequiresLog: true,
	}
}

// EvaluateAction is the main entry point for policy evaluation
// It currently only handles Stripe actions, but can be extended for other tools
func EvaluateAction(toolName string, action string) PolicyDecision {
	switch toolName {
	case "stripe":
		return EvaluateStripeAction(action)
	default:
		// Default: allow unknown tools but log
		return PolicyDecision{
			Action:      action,
			RiskLevel:   GREEN,
			Allowed:     true,
			Reason:      fmt.Sprintf("No policy defined for tool: %s", toolName),
			RequiresLog: false,
		}
	}
}
