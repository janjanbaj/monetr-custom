import { useMutation, useQueryClient } from '@tanstack/react-query';

import { useSelectedBankAccountId } from '@monetr/interface/hooks/useSelectedBankAccountId';
import type { AccrualUsageLog } from '@monetr/interface/models/AccrualExpense';
import request from '@monetr/interface/util/request';

export interface CreateAccrualUsageLogRequest {
  accrualExpenseId: string;
  amount: number;
  date: Date;
}

export function useCreateAccrualUsageLog(): (_: CreateAccrualUsageLogRequest) => Promise<AccrualUsageLog> {
  const queryClient = useQueryClient();
  const selectedBankAccountId = useSelectedBankAccountId();

  async function createUsageLog(req: CreateAccrualUsageLogRequest): Promise<AccrualUsageLog> {
    return request<AccrualUsageLog>({
      method: 'POST',
      url: `/api/bank_accounts/${selectedBankAccountId}/accrual/${req.accrualExpenseId}/usage_logs`,
      data: {
        amount: req.amount,
        date: req.date,
      },
    }).then(result => result.data);
  }

  const mutation = useMutation({
    mutationFn: createUsageLog,
    onSuccess: () =>
      queryClient.invalidateQueries({
        queryKey: [`/api/bank_accounts/${selectedBankAccountId}/accrual`],
      }),
  });

  return mutation.mutateAsync;
}
