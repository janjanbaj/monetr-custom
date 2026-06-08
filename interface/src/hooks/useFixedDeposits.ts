import { useMutation, useQuery, useQueryClient, type UseQueryResult } from '@tanstack/react-query';
import FixedDeposit from '@monetr/interface/models/FixedDeposit';
import request from '@monetr/interface/util/request';

export function useFixedDeposits(bankAccountId: string): UseQueryResult<Array<FixedDeposit>, unknown> {
  return useQuery<Array<Partial<FixedDeposit>>, unknown, Array<FixedDeposit>>({
    queryKey: [`/api/bank_accounts/${bankAccountId}/fixed_deposits`],
    enabled: !!bankAccountId,
    select: data => data.map(item => new FixedDeposit(item)),
  });
}

export interface CreateFixedDepositPayload {
  name: string;
  amount: number;
  interestRate: number;
  startDate: string; // ISO string
  termMonths: number;
  interestFrequency: 'monthly' | 'quarterly' | 'end_of_term';
  interestDestination: 'accumulate' | 'payout';
  interestDestinationAccountId?: string;
}

export function useCreateFixedDeposit(bankAccountId: string) {
  const queryClient = useQueryClient();

  return useMutation<FixedDeposit, Error, CreateFixedDepositPayload>({
    mutationFn: async payload => {
      const { data } = await request<FixedDeposit, CreateFixedDepositPayload>({
        method: 'POST',
        url: `/api/bank_accounts/${bankAccountId}/fixed_deposits`,
        data: payload,
      });
      return new FixedDeposit(data);
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: [`/api/bank_accounts/${bankAccountId}/fixed_deposits`] });
      void queryClient.invalidateQueries({ queryKey: ['/api/bank_accounts'] }); // refresh available balance
    },
  });
}

export function useWithdrawFixedDeposit(bankAccountId: string) {
  const queryClient = useQueryClient();

  return useMutation<FixedDeposit, Error, string>({
    mutationFn: async (fixedDepositId: string) => {
      const { data } = await request<FixedDeposit>({
        method: 'POST',
        url: `/api/bank_accounts/${bankAccountId}/fixed_deposits/${fixedDepositId}/withdraw`,
      });
      return new FixedDeposit(data);
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: [`/api/bank_accounts/${bankAccountId}/fixed_deposits`] });
      void queryClient.invalidateQueries({ queryKey: ['/api/bank_accounts'] });
    },
  });
}

export function useDeleteFixedDeposit(bankAccountId: string) {
  const queryClient = useQueryClient();

  return useMutation<void, Error, string>({
    mutationFn: async (fixedDepositId: string) => {
      await request<void>({
        method: 'DELETE',
        url: `/api/bank_accounts/${bankAccountId}/fixed_deposits/${fixedDepositId}`,
      });
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: [`/api/bank_accounts/${bankAccountId}/fixed_deposits`] });
    },
  });
}
