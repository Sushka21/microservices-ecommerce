// Для dev (npm run dev) оставь пусто — прокси разведёт по портам.
// Для preview/prod задай в .env: VITE_CART_URL=http://localhost:8080, VITE_LOMS_URL=http://localhost:8081
const CART_BASE = import.meta.env.VITE_CART_URL ?? ''
const LOMS_BASE = import.meta.env.VITE_LOMS_URL ?? ''

async function jsonFetch<T>(base: string, path: string, body?: object): Promise<T> {
  const res = await fetch(`${base}${path}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: body ? JSON.stringify(body) : undefined,
  })
  if (!res.ok) {
    const t = await res.text()
    let msg = t
    try {
      const j = JSON.parse(t)
      msg = j.message ?? j.error ?? t
    } catch {
      // use t
    }
    throw new Error(msg || `HTTP ${res.status}`)
  }
  const text = await res.text()
  if (!text.trim()) return undefined as T
  return JSON.parse(text) as T
}

// Cart (gateway 8080)
export const cartApi = {
  addItem(userId: number, sku: number, count: number) {
    return jsonFetch<void>(CART_BASE, '/v1/cart/item/add', { user_id: userId, sku, count })
  },
  deleteItem(userId: number, sku: number) {
    return jsonFetch<void>(CART_BASE, '/v1/cart/item/delete', { user_id: userId, sku })
  },
  async listCart(userId: number): Promise<{ items?: Array<{ sku: number; count: number; name: string; price: number }>; total_price?: number }> {
    const res = await fetch(`${CART_BASE}/v1/cart/list`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ user_id: userId }),
    })
    if (!res.ok) {
      const t = await res.text()
      throw new Error(t || `HTTP ${res.status}`)
    }
    const text = await res.text()
    if (!text.trim()) return { items: [], total_price: 0 }
    const data = JSON.parse(text)
    const r = data?.result ?? data
    const items = r?.items ?? []
    const totalPrice = r?.totalPrice ?? r?.total_price ?? 0
    return { items, total_price: totalPrice }
  },
  clearCart(userId: number) {
    return jsonFetch<void>(CART_BASE, '/v1/cart/clear', { user_id: userId })
  },
  async checkout(userId: number): Promise<{ order_id: number }> {
    const controller = new AbortController()
    const timeout = setTimeout(() => controller.abort(), 15000)
    const res = await fetch(`${CART_BASE}/v1/cart/checkout`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ user_id: userId }),
      signal: controller.signal,
    })
    clearTimeout(timeout)
    if (!res.ok) {
      const t = await res.text()
      throw new Error(t || `HTTP ${res.status}`)
    }
    const text = await res.text()
    if (!text.trim()) {
      throw new Error('Пустой ответ при оформлении заказа. Проверьте, что Cart и LOMS запущены.')
    }
    const data = JSON.parse(text)
    const r = data?.result ?? data
    const orderId = r?.orderId ?? r?.order_id ?? 0
    return { order_id: orderId }
  },
}

// LOMS orders (gateway 8081)
export const orderApi = {
  create(userId: number, items: Array<{ sku: number; count: number }>) {
    return jsonFetch<{ order_id: number }>(LOMS_BASE, '/v1/order/create', { user_id: userId, items })
  },
  async get(orderId: number): Promise<{
    status: number;
    user_id: number;
    items: Array<{ sku: number; count: number }>;
    created_at?: string;
    updated_at?: string;
  }> {
    const data = await jsonFetch<{
      status?: number | string;
      user_id?: number;
      userId?: number;
      items?: Array<{ sku: number; count: number }>;
      result?: { status?: number | string; user_id?: number; userId?: number; items?: Array<{ sku: number; count: number }> };
    }>(LOMS_BASE, '/v1/order/info', { order_id: orderId })
    const r = data?.result ?? data
    let status = r?.status
    if (typeof status === 'string') {
      const map: Record<string, number> = {
        ORDER_STATUS_UNSPECIFIED: 0,
        ORDER_STATUS_NEW: 1,
        ORDER_STATUS_AWAITING_PAYMENT: 2,
        ORDER_STATUS_FAILED: 3,
        ORDER_STATUS_PAID: 4,
        ORDER_STATUS_CANCELLED: 5,
      }
      status = map[status] ?? 0
    }
    const userId = r?.user_id ?? r?.userId ?? 0
    return {
      status: (status as number) ?? 0,
      user_id: userId,
      items: r?.items ?? [],
      created_at: (r as { created_at?: string }).created_at,
      updated_at: (r as { updated_at?: string }).updated_at,
    }
  },
  pay(orderId: number) {
    return jsonFetch<void>(LOMS_BASE, '/v1/order/pay', { order_id: orderId })
  },
  cancel(orderId: number) {
    return jsonFetch<void>(LOMS_BASE, '/v1/order/cancel', { order_id: orderId })
  },
}

// Product (LOMS 8081)
export const productApi = {
  getProduct(sku: number): Promise<{ name: string; price: number }> {
    return jsonFetch(LOMS_BASE, '/v1/product/info', { sku })
  },
  async createProduct(name: string, price: number): Promise<{ sku: number }> {
    const data = await jsonFetch<{ sku?: number; result?: { sku?: number } }>(LOMS_BASE, '/v1/product/create', { name, price })
    const r = data?.result ?? data
    const sku = r?.sku ?? 0
    return { sku }
  },
}

// Stock (LOMS 8081)
export const stockApi = {
  get(sku: number): Promise<{ count: number }> {
    return jsonFetch(LOMS_BASE, '/v1/stock/info', { sku })
  },
  setStock(sku: number, count: number): Promise<void> {
    return jsonFetch<void>(LOMS_BASE, '/v1/stock/set', { sku, count })
  },
}
