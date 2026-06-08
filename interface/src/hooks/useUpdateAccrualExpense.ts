import { useMutation } from '@tanstack/react-query';

import AccrualExpense from '@monetr/interface/models/AccrualExpense';
import request from '@monetr/interface/util/request';

export function useUpdateAccrualExpense(): (_accrualExpense: AccrualExpense) => Promise<AccrualExpense> {
  const { mutateAsync } = useMutation({
    mutationFn: async (accrualExpense: AccrualExpense): Promise<AccrualExpense> =>
      request<Partial<AccrualExpense>>({
        method: 'PUT',
        url: `/api/bank_accounts/${accrualExpense.bankAccountId}/accrual/${accrualExpense.accrualExpenseId}`,
        data: accrualExpense,
      }).then(result => new AccrualExpense(result?.data)),
    onSuccess: (updated: AccrualExpense, _variables, _onMutateResult, { client: queryClient }) =>
      Promise.all([
        queryClient.setQueryData(
          [`/api/bank_accounts/${updated.bankAccountId}/accrual`],
          (previous: Array<Partial<AccrualExpense>>) =>
            (previous ?? []).map(item => (item.accrualExpenseId === updated.accrualExpenseId ? updated : item)),
        ),
        queryClient.setQueryData(
          [`/api/bank_accounts/${updated.bankAccountId}/accrual/${updated.accrualExpenseId}`],
          updated,
        ),
      ]),
  });

  return mutateAsync;
}
