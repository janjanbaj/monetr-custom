# Phase 4: Frontend API Clients & Hooks

This phase covers defining the TypeScript domain models and creating React hooks (TanStack Query) to communicate with the fixed deposit REST endpoints.

## Target Files
1. [FixedDeposit.ts](file:///Users/janeetbajracharya/Desktop/Code/monetr/interface/src/models/FixedDeposit.ts) [NEW]
2. [useFixedDeposits.ts](file:///Users/janeetbajracharya/interface/src/hooks/useFixedDeposits.ts) [NEW]

---

## 1. Domain Model
Create `interface/src/models/FixedDeposit.ts`:

```typescript
import parseDate from '@monetr/interface/util/parseDate';
import type BankAccount from './BankAccount';

export type FixedDepositStatus = 'active' | 'matured' | 'withdrawn';

export default class FixedDeposit {
  fixedDepositId: string;
  sourceBankAccountId: string;
  fixedBankAccountId: string;
  fundingScheduleId?: string;
  name: string;
  amount: number;
  interestRate: number;
  startDate: Date;
  endDate: Date;
  interestFrequency: 'monthly' | 'quarterly' | 'end_of_term';
  interestDestination: 'accumulate' | 'payout';
  interestDestinationAccountId?: string;
  status: FixedDepositStatus;
  createdAt: Date;
  updatedAt: Date;

  // Relations (joined)
  fixedBankAccount?: BankAccount;

  constructor(data?: Partial<FixedDeposit>) {
    if (data) {
      Object.assign(this, {
        ...data,
        startDate: parseDate(data.startDate) || new Date(),
        endDate: parseDate(data.endDate) || new Date(),
        createdAt: parseDate(data.createdAt),
        updatedAt: parseDate(data.updatedAt),
      });
    }
  }
}
```

---

## 2. React Hooks (TanStack Query)
Create `interface/src/hooks/useFixedDeposits.ts`:

```typescript
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
```

---

## Verification Plan
- Compile/transpile the frontend using `tsc` to verify typescript safety.
- Write mock queries and assert request paths.
