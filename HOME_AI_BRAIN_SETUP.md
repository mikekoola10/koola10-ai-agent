# Home AI Brain: Koola10/Spiral Integration Guide

This document outlines the hardware and software configuration required to transform your household into a fully intelligent, AI-driven home ecosystem.

## Phase 1: Smart Device Integration

### Recommended Central Hub
**Home Assistant (Local Control / Open Source)**
- **Why:** Offers the best local privacy, vast device support (2000+ integrations), and a robust REST API for Koola10 integration.
- **Hardware:** Raspberry Pi 4 (8GB) or a dedicated NUC.

### Recommended Smart Devices
| Category | Device Recommendation | Protocol |
| :--- | :--- | :--- |
| **Lights** | Philips Hue / Govee | Zigbee / WiFi |
| **Sensors (Motion/Temp)** | Aqara P1 / Sonoff | Zigbee |
| **Thermostat** | Ecobee Premium / Nest | WiFi |
| **Smart Plugs** | TP-Link Kasa / Shelly | WiFi |
| **Smart Locks** | August Wi-Fi / Yale | WiFi / Zigbee |
| **Cameras** | Reolink (PoE) / Eufy | RTSP / Local |
| **Audio** | Sonos / HomePod / Echo | Multi-room |

## Phase 2: Software Configuration

### 1. Hub Connection
Expose Home Assistant to Koola10 via Long-Lived Access Tokens. Koola10 will use the `/home/sensor-update` endpoint to mirror the hub's state.

### 2. Automation Workflows
Workflows are defined in Koola10's `HomeBrain` logic, covering:
- **Welcome Home:** Entry lights, Daily Summary audio, Thermostat adjust.
- **Energy Optimization:** Night-time temperature lowering, lighting shutoff.
- **Health & Safety:** Anomaly detection on bedroom door/motion.
- **Focus Mode:** Dimming, ambient sound, notification silencing.
- **Parent Mode:** UI Simplification, warmer tones, proactive billing prompts.

## Phase 3: Intelligence & Learning
Koola10 uses DeepSeek to analyze patterns in `/data/home_brain.json` every 24 hours to update predictive models.

## Phase 4: Full Integration
- **Digital Mancave:** Accessible via the "Home Control" panel.
- **Ledger:** Energy usage is synced to the `EconomicLedger` for real-time ROI tracking.
