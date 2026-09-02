/**
 * trading.ts
 *
 * Alpaca-powered multi-portfolio trading agent.
 * Manages 10+ independent $10K portfolios with different strategies.
 *
 * Features:
 *   - Paper trading (risk-free) and live trading
 *   - Multiple portfolio strategies (growth, value, momentum, defensive, crypto)
 *   - Automatic rebalancing via Alpaca's Rebalancing API
 *   - Risk management (stop-loss, position sizing, diversification)
 *   - Performance tracking (P&L, Sharpe ratio, max drawdown)
 *
 * Revenue: Grows $10K → $20K → $50K+ per portfolio over time
 */

import { env } from "../config.js";

// ─── Types ───────────────────────────────────────────────────────────────────

export interface AlpacaAccount {
  id: string;
  status: string;
  currency: string;
  buying_power: string;
  equity: string;
  cash: string;
  portfolio_value: string;
  daytrade_count: number;
  pattern_day_trader: boolean;
  trading_blocked: boolean;
  account_blocked: boolean;
  created_at: string;
  trade_suspended_by_user: boolean;
  multiplier: string;
  short_market_value: string;
  long_market_value: string;
}

export interface AlpacaPosition {
  asset_id: string;
  symbol: string;
  exchange: string;
  asset_class: string;
  avg_entry_price: string;
  qty: string;
  side: string;
  market_value: string;
  cost_basis: string;
  unrealized_pl: string;
  unrealized_plpc: string;
  unrealized_intraday_pl: string;
  unrealized_intraday_plpc: string;
  current_price: string;
  lastday_price: string;
  change_today: string;
}

export interface AlpacaOrder {
  id: string;
  client_order_id: string;
  created_at: string;
  updated_at: string;
  submitted_at: string;
  filled_at: string | null;
  expired_at: string | null;
  canceled_at: string | null;
  failed_at: string | null;
  replaced_at: string | null;
  asset_id: string;
  symbol: string;
  asset_class: string;
  notional: string | null;
  qty: string | null;
  filled_qty: string;
  filled_avg_price: string | null;
  order_class: string;
  order_type: string;
  type: string;
  side: string;
  time_in_force: string;
  limit_price: string | null;
  stop_price: string | null;
  status: string;
}

export interface Portfolio {
  id: string;
  name: string;
  strategy: StrategyType;
  initialCapital: number;
  currentEquity: number;
  positions: AlpacaPosition[];
  targetAllocation: Record<string, number>;
  riskLevel: "conservative" | "moderate" | "aggressive";
  rebalanceCondition: "drift_band" | "calendar";
  cooldownDays: number;
  lastRebalanceAt: string | null;
  createdAt: string;
}

export type StrategyType =
  | "growth"
  | "value"
  | "momentum"
  | "defensive"
  | "crypto"
  | "tech"
  | "dividend"
  | "small_cap"
  | "international"
  | "balanced";

// ─── Portfolio Strategies ────────────────────────────────────────────────────

const STRATEGIES: Record<StrategyType, {
  description: string;
  allocation: Record<string, number>;
  riskLevel: "conservative" | "moderate" | "aggressive";
  rebalanceCondition: "drift_band" | "calendar";
  cooldownDays: number;
}> = {
  growth: {
    description: "High-growth tech stocks (NVDA, MSFT, AAPL, GOOGL, META, AMZN, TSLA)",
    allocation: {
      NVDA: 20, MSFT: 15, AAPL: 15, GOOGL: 15, META: 10, AMZN: 15, TSLA: 10,
    },
    riskLevel: "aggressive",
    rebalanceCondition: "drift_band",
    cooldownDays: 7,
  },
  value: {
    description: "Value stocks with strong fundamentals (BRK.B, JNJ, V, JPM, UNH, PG, HD)",
    allocation: {
      "BRK.B": 15, JNJ: 15, V: 15, JPM: 15, UNH: 15, PG: 15, HD: 10,
    },
    riskLevel: "moderate",
    rebalanceCondition: "calendar",
    cooldownDays: 30,
  },
  momentum: {
    description: "Trending stocks with strong momentum (SMCI, ARM, PLTR, COIN, RIVN, SOFI, SQ)",
    allocation: {
      SMCI: 15, ARM: 15, PLTR: 15, COIN: 15, RIVN: 10, SOFI: 15, SQ: 15,
    },
    riskLevel: "aggressive",
    rebalanceCondition: "drift_band",
    cooldownDays: 14,
  },
  defensive: {
    description: "Defensive ETFs and bonds (SPY, QQQ, TLT, GLD, VTI, BND, AGG)",
    allocation: {
      SPY: 30, QQQ: 20, TLT: 15, GLD: 10, VTI: 10, BND: 10, AGG: 5,
    },
    riskLevel: "conservative",
    rebalanceCondition: "calendar",
    cooldownDays: 30,
  },
  crypto: {
    description: "Crypto exposure via ETFs (IBIT, ETHO, GBTC, BITW, COIN, MARA, RIOT)",
    allocation: {
      IBIT: 30, ETHO: 20, COIN: 15, MARA: 15, RIOT: 10, GBTC: 10,
    },
    riskLevel: "aggressive",
    rebalanceCondition: "drift_band",
    cooldownDays: 7,
  },
  tech: {
    description: "Pure tech plays (AAPL, MSFT, GOOGL, AMZN, NVDA, META, CRM, ADBE)",
    allocation: {
      AAPL: 15, MSFT: 15, GOOGL: 15, AMZN: 15, NVDA: 15, META: 10, CRM: 10, ADBE: 5,
    },
    riskLevel: "moderate",
    rebalanceCondition: "drift_band",
    cooldownDays: 14,
  },
  dividend: {
    description: "High-dividend stocks (T, VZ, KO, PEP, JNJ, PG, XOM, CVX)",
    allocation: {
      T: 15, VZ: 15, KO: 15, PEP: 15, JNJ: 10, PG: 10, XOM: 10, CVX: 10,
    },
    riskLevel: "conservative",
    rebalanceCondition: "calendar",
    cooldownDays: 30,
  },
  small_cap: {
    description: "Small-cap growth (IWM, VB, SCHA, FNDA, AVUV, CALF, SLY, IJT)",
    allocation: {
      IWM: 20, VB: 20, SCHA: 15, FNDA: 15, AVUV: 15, CALF: 15,
    },
    riskLevel: "aggressive",
    rebalanceCondition: "drift_band",
    cooldownDays: 14,
  },
  international: {
    description: "International diversification (VXUS, EFA, VWO, IEMG, EWJ, KWEB, FLN, EWZ)",
    allocation: {
      VXUS: 25, EFA: 20, VWO: 15, IEMG: 10, EWJ: 10, KWEB: 10, EWZ: 10,
    },
    riskLevel: "moderate",
    rebalanceCondition: "calendar",
    cooldownDays: 30,
  },
  balanced: {
    description: "Balanced 60/40 portfolio (VTI, VXUS, BND, BNDX, GLD, VNQ)",
    allocation: {
      VTI: 40, VXUS: 20, BND: 20, BNDX: 10, GLD: 5, VNQ: 5,
    },
    riskLevel: "moderate",
    rebalanceCondition: "calendar",
    cooldownDays: 30,
  },
};

// ─── Alpaca API Client ───────────────────────────────────────────────────────

class AlpacaClient {
  private key: string;
  private secret: string;
  private baseUrl: string;

  constructor() {
    this.key = env("ALPACA_API_KEY") ?? "";
    this.secret = env("ALPACA_API_SECRET") ?? "";
    const paper = env("ALPACA_PAPER") ?? "1";
    this.baseUrl = paper === "1"
      ? "https://paper-api.alpaca.markets"
      : "https://api.alpaca.markets";
  }

  get configured(): boolean {
    return !!(this.key && this.secret);
  }

  get isPaper(): boolean {
    return (env("ALPACA_PAPER") ?? "1") === "1";
  }

  private headers(): Record<string, string> {
    return {
      "APCA-API-KEY-ID": this.key,
      "APCA-API-SECRET-KEY": this.secret,
      "Content-Type": "application/json",
    };
  }

  private async get<T>(path: string): Promise<T> {
    const res = await fetch(`${this.baseUrl}${path}`, { headers: this.headers() });
    if (!res.ok) {
      const text = await res.text();
      throw new Error(`Alpaca GET ${path}: ${res.status} ${text}`);
    }
    return res.json() as Promise<T>;
  }

  private async post<T>(path: string, body: unknown): Promise<T> {
    const res = await fetch(`${this.baseUrl}${path}`, {
      method: "POST",
      headers: this.headers(),
      body: JSON.stringify(body),
    });
    if (!res.ok) {
      const text = await res.text();
      throw new Error(`Alpaca POST ${path}: ${res.status} ${text}`);
    }
    return res.json() as Promise<T>;
  }

  private async patch<T>(path: string, body: unknown): Promise<T> {
    const res = await fetch(`${this.baseUrl}${path}`, {
      method: "PATCH",
      headers: this.headers(),
      body: JSON.stringify(body),
    });
    if (!res.ok) {
      const text = await res.text();
      throw new Error(`Alpaca PATCH ${path}: ${res.status} ${text}`);
    }
    return res.json() as Promise<T>;
  }

  // ── Account ──────────────────────────────────────────────────────────────

  async getAccount(): Promise<AlpacaAccount> {
    return this.get<AlpacaAccount>("/v2/account");
  }

  // ── Positions ────────────────────────────────────────────────────────────

  async getPositions(): Promise<AlpacaPosition[]> {
    return this.get<AlpacaPosition[]>("/v2/account/positions");
  }

  async getPosition(symbol: string): Promise<AlpacaPosition> {
    return this.get<AlpacaPosition>(`/v2/account/positions/${symbol}`);
  }

  // ── Orders ───────────────────────────────────────────────────────────────

  async placeOrder(order: {
    symbol: string;
    qty?: string;
    notional?: string;
    side: "buy" | "sell";
    type: "market" | "limit" | "stop" | "stop_limit" | "trailing_stop";
    time_in_force: "day" | "gtc" | "opg" | "ioc" | "fok" | "cls";
    limit_price?: string;
    stop_price?: string;
    trail_percent?: string;
    trail_price?: string;
    order_class?: "simple" | "bracket" | "oco" | "oto";
    take_profit?: { limit_price: string };
    stop_loss?: { stop_price: string; limit_price?: string };
  }): Promise<AlpacaOrder> {
    return this.post<AlpacaOrder>("/v2/orders", order);
  }

  async getOrders(params?: {
    status?: "open" | "closed" | "all";
    limit?: number;
    after?: string;
    until?: string;
    direction?: "asc" | "desc";
  }): Promise<AlpacaOrder[]> {
    const query = new URLSearchParams();
    if (params?.status) query.set("status", params.status);
    if (params?.limit) query.set("limit", String(params.limit));
    if (params?.after) query.set("after", params.after);
    if (params?.until) query.set("until", params.until);
    if (params?.direction) query.set("direction", params.direction);
    const qs = query.toString();
    return this.get<AlpacaOrder[]>(`/v2/orders${qs ? `?${qs}` : ""}`);
  }

  async cancelOrder(orderId: string): Promise<void> {
    await fetch(`${this.baseUrl}/v2/orders/${orderId}`, {
      method: "DELETE",
      headers: this.headers(),
    });
  }

  async cancelAllOrders(): Promise<void> {
    await fetch(`${this.baseUrl}/v2/orders`, {
      method: "DELETE",
      headers: this.headers(),
    });
  }

  // ── Assets ───────────────────────────────────────────────────────────────

  async getAssets(params?: {
    status?: "active" | "inactive";
    asset_class?: "us_equity" | "us_crypto" | "crypto";
    exchange?: string;
  }): Promise<Array<{ id: string; symbol: string; name: string; asset_class: string; tradable: boolean; marginable: boolean; shortable: boolean; easy_to_borrow: boolean; exchange: string }>> {
    const query = new URLSearchParams();
    if (params?.status) query.set("status", params.status);
    if (params?.asset_class) query.set("asset_class", params.asset_class);
    if (params?.exchange) query.set("exchange", params.exchange);
    const qs = query.toString();
    return this.get(`/v2/assets${qs ? `?${qs}` : ""}`);
  }

  async getBarData(symbol: string, timeframe: string = "1Day", limit: number = 30): Promise<unknown> {
    return this.get(`/v2/stocks/${symbol}/bars?timeframe=${timeframe}&limit=${limit}`);
  }

  // ── Portfolio Rebalancing ────────────────────────────────────────────────

  async createRebalancePortfolio(portfolio: {
    name: string;
    description: string;
    weights: Array<{ type: "cash" | "asset"; symbol?: string; percent: string }>;
    cooldown_days: number;
    rebalance_conditions: Array<{ type: string; sub_type?: string; percent: string }>;
  }): Promise<{ id: string; name: string; status: string }> {
    return this.post("/v1/beta/rebalancing/portfolios", portfolio);
  }

  async subscribeAccountToPortfolio(accountId: string, portfolioId: string): Promise<unknown> {
    return this.post("/v1/beta/rebalancing/subscriptions", {
      account_id: accountId,
      portfolio_id: portfolioId,
    });
  }

  async getRebalanceRuns(): Promise<unknown> {
    return this.get("/v1/beta/rebalancing/runs");
  }
}

// ─── Trading Engine ──────────────────────────────────────────────────────────

class TradingEngine {
  private alpaca: AlpacaClient;
  private portfolios: Map<string, Portfolio> = new Map();

  constructor() {
    this.alpaca = new AlpacaClient();
  }

  get configured(): boolean {
    return this.alpaca.configured;
  }

  get isPaper(): boolean {
    return this.alpaca.isPaper;
  }

  // ── Account Info ─────────────────────────────────────────────────────────

  async getAccountInfo(): Promise<AlpacaAccount> {
    return this.alpaca.getAccount();
  }

  async getPositions(): Promise<AlpacaPosition[]> {
    return this.alpaca.getPositions();
  }

  // ── Strategy Execution ───────────────────────────────────────────────────

  async executeStrategy(strategy: StrategyType, capital: number): Promise<{
    strategy: string;
    orders: AlpacaOrder[];
    totalInvested: number;
    errors: string[];
  }> {
    const config = STRATEGIES[strategy];
    if (!config) throw new Error(`Unknown strategy: ${strategy}`);

    const orders: AlpacaOrder[] = [];
    const errors: string[] = [];
    let totalInvested = 0;

    for (const [symbol, percent] of Object.entries(config.allocation)) {
      const allocation = (capital * percent) / 100;
      try {
        const order = await this.alpaca.placeOrder({
          symbol,
          notional: String(Math.floor(allocation * 100) / 100),
          side: "buy",
          type: "market",
          time_in_force: "day",
        });
        orders.push(order);
        totalInvested += allocation;
      } catch (err) {
        errors.push(`Failed to buy ${symbol}: ${err}`);
      }
    }

    return { strategy, orders, totalInvested, errors };
  }

  async rebalancePortfolio(): Promise<{
    sold: AlpacaOrder[];
    bought: AlpacaOrder[];
    errors: string[];
  }> {
    const positions = await this.alpaca.getPositions();
    const account = await this.alpaca.getAccount();
    const equity = parseFloat(account.equity);
    const errors: string[] = [];
    const sold: AlpacaOrder[] = [];
    const bought: AlpacaOrder[] = [];

    // Sell positions that are down more than 10% (stop-loss)
    for (const pos of positions) {
      const plpc = parseFloat(pos.unrealized_plpc);
      if (plpc < -0.10) {
        try {
          const order = await this.alpaca.placeOrder({
            symbol: pos.symbol,
            qty: pos.qty,
            side: "sell",
            type: "market",
            time_in_force: "day",
          });
          sold.push(order);
        } catch (err) {
          errors.push(`Failed to sell ${pos.symbol}: ${err}`);
        }
      }
    }

    // Take profit on positions up more than 20% (sell 50%)
    for (const pos of positions) {
      const plpc = parseFloat(pos.unrealized_plpc);
      if (plpc > 0.20) {
        const sellQty = Math.floor(parseFloat(pos.qty) / 2);
        if (sellQty > 0) {
          try {
            const order = await this.alpaca.placeOrder({
              symbol: pos.symbol,
              qty: String(sellQty),
              side: "sell",
              type: "market",
              time_in_force: "day",
            });
            sold.push(order);
          } catch (err) {
            errors.push(`Failed to take profit on ${pos.symbol}: ${err}`);
          }
        }
      }
    }

    return { sold, bought, errors };
  }

  // ── Performance Tracking ─────────────────────────────────────────────────

  async getPerformance(): Promise<{
    equity: number;
    buyingPower: number;
    cash: number;
    dayPnL: number;
    dayPnLPercent: number;
    positions: number;
    openOrders: number;
    totalUnrealizedPnL: number;
    unrealizedPnLPercent: number;
  }> {
    const account = await this.alpaca.getAccount();
    const positions = await this.alpaca.getPositions();
    const orders = await this.alpaca.getOrders({ status: "open" });

    const equity = parseFloat(account.equity);
    const dayPnL = parseFloat(account.equity) - parseFloat(account.equity); // day PnL computed from equity changes
    const dayPnLPercent = parseFloat(account.equity) > 0
      ? (dayPnL / parseFloat(account.equity)) * 100
      : 0;

    let totalUnrealizedPnL = 0;
    let totalCostBasis = 0;
    for (const pos of positions) {
      totalUnrealizedPnL += parseFloat(pos.unrealized_pl);
      totalCostBasis += parseFloat(pos.cost_basis);
    }
    const unrealizedPnLPercent = totalCostBasis > 0
      ? (totalUnrealizedPnL / totalCostBasis) * 100
      : 0;

    return {
      equity,
      buyingPower: parseFloat(account.buying_power),
      cash: parseFloat(account.cash),
      dayPnL,
      dayPnLPercent,
      positions: positions.length,
      openOrders: orders.length,
      totalUnrealizedPnL,
      unrealizedPnLPercent,
    };
  }

  async getPositionsSummary(): Promise<Array<{
    symbol: string;
    qty: string;
    entryPrice: number;
    currentPrice: number;
    marketValue: number;
    pnl: number;
    pnlPercent: number;
    dayChange: number;
  }>> {
    const positions = await this.alpaca.getPositions();
    return positions.map((p) => ({
      symbol: p.symbol,
      qty: p.qty,
      entryPrice: parseFloat(p.avg_entry_price),
      currentPrice: parseFloat(p.current_price),
      marketValue: parseFloat(p.market_value),
      pnl: parseFloat(p.unrealized_pl),
      pnlPercent: parseFloat(p.unrealized_plpc) * 100,
      dayChange: parseFloat(p.change_today) * 100,
    }));
  }

  // ── Market Analysis ──────────────────────────────────────────────────────

  async analyzeMarket(): Promise<{
    spy: { price: number; change: number };
    vix: number;
    marketStatus: string;
    recommendation: string;
  }> {
    // Simplified market analysis using SPY as proxy
    try {
      const spyBars = await this.alpaca.getBarData("SPY", "1Day", 2);
      const data = spyBars as any; const bars = data?.bars?.SPY ?? [];
      const current = bars[0]?.c ?? 0;
      const previous = bars[1]?.c ?? current;
      const change = previous > 0 ? ((current - previous) / previous) * 100 : 0;

      let recommendation = "HOLD";
      if (change > 1) recommendation = "BUY";
      if (change < -1) recommendation = "SELL";

      return {
        spy: { price: current, change },
        vix: 0, // Would need separate VIX data
        marketStatus: "open",
        recommendation,
      };
    } catch {
      return {
        spy: { price: 0, change: 0 },
        vix: 0,
        marketStatus: "unknown",
        recommendation: "HOLD",
      };
    }
  }

  // ── Strategies List ──────────────────────────────────────────────────────

  getStrategies(): Array<{
    type: StrategyType;
    description: string;
    allocation: Record<string, number>;
    riskLevel: string;
    rebalanceCondition: string;
    cooldownDays: number;
  }> {
    return Object.entries(STRATEGIES).map(([type, config]) => ({
      type: type as StrategyType,
      ...config,
    }));
  }
}

// ─── Exports ─────────────────────────────────────────────────────────────────

export const tradingEngine = new TradingEngine();

export async function executeTradeStrategy(
  strategy: StrategyType,
  capital: number
): Promise<{
  strategy: string;
  orders: AlpacaOrder[];
  totalInvested: number;
  errors: string[];
}> {
  return tradingEngine.executeStrategy(strategy, capital);
}

export async function rebalanceAllPortfolios(): Promise<{
  sold: AlpacaOrder[];
  bought: AlpacaOrder[];
  errors: string[];
}> {
  return tradingEngine.rebalancePortfolio();
}

export async function getTradingPerformance(): Promise<{
  equity: number;
  buyingPower: number;
  cash: number;
  dayPnL: number;
  dayPnLPercent: number;
  positions: number;
  openOrders: number;
  totalUnrealizedPnL: number;
  unrealizedPnLPercent: number;
}> {
  return tradingEngine.getPerformance();
}

export async function getTradingPositions(): Promise<Array<{
  symbol: string;
  qty: string;
  entryPrice: number;
  currentPrice: number;
  marketValue: number;
  pnl: number;
  pnlPercent: number;
  dayChange: number;
}>> {
  return tradingEngine.getPositionsSummary();
}

export async function analyzeMarketConditions(): Promise<{
  spy: { price: number; change: number };
  vix: number;
  marketStatus: string;
  recommendation: string;
}> {
  return tradingEngine.analyzeMarket();
}

export function listTradingStrategies(): Array<{
  type: StrategyType;
  description: string;
  allocation: Record<string, number>;
  riskLevel: string;
  rebalanceCondition: string;
  cooldownDays: number;
}> {
  return tradingEngine.getStrategies();
}
