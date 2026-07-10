import axios from 'axios';

const BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || 'https://koola10.fly.dev';

export const api = axios.create({
  baseURL: BASE_URL,
});

export const fetcher = (url: string) => api.get(url).then((res) => res.data);

export const postRequest = (url: string, data?: unknown) => api.post(url, data).then((res) => res.data);
