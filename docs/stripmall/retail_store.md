# Store #1 – Discount Retail (Dollar General / Family Dollar)

## Hardware and Software Requirements
- **POS**: Square/Clover with Custom API integration
- **Inventory**: RFID scanners, IoT-enabled shelving, auto-replenishment backend
- **Security**: AI-powered CCTV with object detection and loss prevention logic
- **Signage**: 4K Digital displays with Solara agent integration

## AI Agent Configuration
- **Nova (Inventory)**: Auto-order based on sales data, seasonality, and supplier catalogues.
- **Sterling (Pricing)**: Dynamic pricing based on competitors and stock levels.
- **Forge (Staffing)**: AI-driven scheduling from foot traffic predictions.
- **Solara (Signage)**: In-store digital signage with personalised offers.
- **Sage (Security)**: AI camera monitoring for loss prevention.

## Data Schema (Retail Sub-ledger)
```json
{
  "timestamp": "ISO8601",
  "type": "revenue | cost",
  "category": "inventory | labor | sales | security",
  "profit_center": "retail",
  "amount": "float64",
  "description": "string"
}
```

## Customer Experience Flow
1. **Entry**: Customer is greeted by digital signage displaying personalized offers (Solara).
2. **Shopping**: RFID-enabled items track interest; inventory levels are monitored in real-time (Nova).
3. **Checkout**: POS logs transaction, updates Mirror loyalty points, and routes 70% revenue to Spiral.
4. **Exit**: Loss prevention (Sage) verifies checkout; personalized coupon issued for next visit.

## Staff Training Guide
- **AI Staff**: Monitors IoT sensors and executes autonomous orders. Reports anomalies to human supervisors via AgentMail.
- **Human Staff**: Focuses on customer service and stocking shelves based on Nova's AI-generated priority lists.
