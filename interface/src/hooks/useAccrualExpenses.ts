import { type UseQueryResult, useQuery } from '@tanstack/react-query';

import { useSelectedBankAccountId } from '@monetr/interface/hooks/useSelectedBankAccountId';
import AccrualExpense from '@monetr/interface/models/AccrualExpense';

export function useAccrualExpenses(): UseQueryResult<Array<AccrualExpense>, unknown> {
  const selectedBankAccountId = useSelectedBankAccountId();
  return useQuery<Array<Partial<AccrualExpense>>, unknown, Array<AccrualExpense>>({
    queryKey: [`/api/bank_accounts/${selectedBankAccountId}/accrual`],
    enabled: Boolean(selectedBankAccountId),
    select: data => (data || []).map(item => new AccrualExpense(item)),
  });
}
