import { useMutation, useQueryClient } from '@tanstack/react-query';

import AccrualExpense from '@monetr/interface/models/AccrualExpense';
import request from '@monetr/interface/util/request';

export function useCreateAccrualExpense(): (_accrualExpense: AccrualExpense) => Promise<AccrualExpense> {
  const queryClient = useQueryClient();

  async function createAccrualExpense(accrualExpense: AccrualExpense): Promise<AccrualExpense> {
    return request<Partial<AccrualExpense>>({
      method: 'POST',
      url: `/api/bank_accounts/${accrualExpense.bankAccountId}/accrual`,
      data: accrualExpense,
    }).then(result => new AccrualExpense(result?.data));
  }

  const mutation = useMutation({
    mutationFn: createAccrualExpense,
    onSuccess: (created: AccrualExpense) =>
      Promise.all([
        queryClient.setQueryData(
          [`/api/bank_accounts/${created.bankAccountId}/accrual`],
          (previous: Array<Partial<AccrualExpense>>) => (previous || []).concat(created),
        ),
        queryClient.setQueryData(
          [`/api/bank_accounts/${created.bankAccountId}/accrual/${created.accrualExpenseId}`],
          created,
        ),
      ]),
  });

  return mutation.mutateAsync;
}
