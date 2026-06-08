import { useMutation, useQueryClient } from '@tanstack/react-query';

import { useSelectedBankAccountId } from '@monetr/interface/hooks/useSelectedBankAccountId';
import request from '@monetr/interface/util/request';

export function useDeleteAccrualUsageLog(): (_accrualUsageLogId: string) => Promise<unknown> {
  const queryClient = useQueryClient();
  const selectedBankAccountId = useSelectedBankAccountId();

  async function deleteUsageLog(accrualUsageLogId: string): Promise<unknown> {
    return request({
      method: 'DELETE',
      url: `/api/bank_accounts/${selectedBankAccountId}/usage_logs/${accrualUsageLogId}`,
    });
  }

  const mutation = useMutation({
    mutationFn: deleteUsageLog,
    onSuccess: () =>
      queryClient.invalidateQueries({
        queryKey: [`/api/bank_accounts/${selectedBankAccountId}/accrual`],
      }),
  });

  return mutation.mutateAsync;
}
