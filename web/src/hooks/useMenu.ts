import { useQuery } from '@tanstack/react-query';
import { menuAPI } from '../api/client';

export function useMenu(branchId?: string, category?: string, search?: string) {
  return useQuery({
    queryKey: ['menu', branchId, category, search],
    queryFn: async () => {
      const res = await menuAPI.list({ branch_id: branchId, category, search, limit: 100 });
      return res.data.data;
    },
  });
}

export function useMenuStats(branchId?: string) {
  return useQuery({
    queryKey: ['menuStats', branchId],
    queryFn: async () => {
      const res = await menuAPI.stats(branchId);
      return res.data.data;
    },
  });
}
