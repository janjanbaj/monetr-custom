import { useMutation, useQueryClient } from '@tanstack/react-query';

import { useSelectedBankAccountId } from '@monetr/interface/hooks/useSelectedBankAccountId';
import type AccrualExpense from '@monetr/interface/models/AccrualExpense';
import request from '@monetr/interface/util/request';

export function useRemoveAccrualExpense(): (_accrualExpenseId: string) => Promise<unknown> {
  const queryClient = useQueryClient();
  const selectedBankAccountId = useSelectedBankAccountId();

  async function removeAccrualExpense(accrualExpenseId: string): Promise<string> {
    return request({
      method: 'DELETE',
      url: `/api/bank_accounts/${selectedBankAccountId}/accrual/${accrualExpenseId}`,
    }).then(() => accrualExpenseId);
  }

  const mutation = useMutation({
    mutationFn: removeAccrualExpense,
    onSuccess: (removedAccrualExpenseId: string) =>
      Promise.all([
        queryClient.setQueryData(
          [`/api/bank_accounts/${selectedBankAccountId}/accrual`],
          (previous: Array<Partial<AccrualExpense>>) =>
            previous.filter(item => item.accrualExpenseId !== removedAccrualExpenseId),
        ),
        queryClient.removeQueries({
          queryKey: [`/api/bank_accounts/${selectedBankAccountId}/accrual/${removedAccrualExpenseId}`],
        }),
      ]),
  });

  return mutation.mutateAsync;
}
