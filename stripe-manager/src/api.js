const API_BASE = '/admin/stripe';
const ADMIN_KEY = 'MzE5OGYzNGEtZmM1ZC00YjY3LWI3ZGMtYjZiOTc5YzdjNzUyYjcwNDczMjYtNjg4Yi00OGIzLTg3NzMtZGQzOTc5NTViZmE0';

async function apiCall(path, options = {}) {
  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${ADMIN_KEY}`,
      ...options.headers,
    },
  });

  if (!res.ok) {
    const text = await res.text();
    throw new Error(`HTTP ${res.status}: ${text}`);
  }

  return res.json();
}

// Products
export const listProducts = () => apiCall('/products');
export const getProduct = (id) => apiCall(`/products/${id}`);
export const createProduct = (data) => apiCall('/products', { method: 'POST', body: JSON.stringify(data) });
export const updateProduct = (id, data) => apiCall(`/products/${id}`, { method: 'PUT', body: JSON.stringify(data) });
export const deleteProduct = (id) => apiCall(`/products/${id}`, { method: 'DELETE' });

// Customers
export const listCustomers = () => apiCall('/customers');
export const getCustomer = (id) => apiCall(`/customers/${id}`);

// Subscriptions
export const listSubscriptions = () => apiCall('/subscriptions');
export const cancelSubscription = (id) => apiCall(`/subscriptions/${id}/cancel`, { method: 'POST' });

// Payments
export const listPayments = () => apiCall('/payments');

// Checkout
export const createCheckout = (data) => apiCall('/checkout', { method: 'POST', body: JSON.stringify(data) });

// Webhooks
export const listWebhooks = () => apiCall('/webhooks');
export const createWebhook = (data) => apiCall('/webhooks', { method: 'POST', body: JSON.stringify(data) });
export const deleteWebhook = (id) => apiCall(`/webhooks/${id}`, { method: 'DELETE' });

// Revenue
export const getRevenue = () => apiCall('/revenue');
