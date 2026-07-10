import useSWR from 'swr';
import { fetcher } from '@/lib/api';

const REFRESH_INTERVAL = 60000; // 60 seconds

export function useHealth() {
  const { data, error, isLoading, mutate } = useSWR('/health', fetcher, {
    refreshInterval: REFRESH_INTERVAL,
  });
  return { data, error, isLoading, mutate };
}

export function useRevenue() {
  const { data, error, isLoading, mutate } = useSWR('/swarm/revenue', fetcher, {
    refreshInterval: REFRESH_INTERVAL,
  });
  return { data, error, isLoading, mutate };
}

export function useAgentStatus() {
  const { data, error, isLoading, mutate } = useSWR('/swarm/status', fetcher, {
    refreshInterval: REFRESH_INTERVAL,
  });
  return { data, error, isLoading, mutate };
}

export function useFinancialHealth() {
  const { data, error, isLoading, mutate } = useSWR('/financial/status', fetcher, {
    refreshInterval: REFRESH_INTERVAL,
  });
  return { data, error, isLoading, mutate };
}

export function useDailyReport() {
  const { data, error, isLoading, mutate } = useSWR('/swarm/report', fetcher, {
    refreshInterval: REFRESH_INTERVAL,
  });
  return { data, error, isLoading, mutate };
}
