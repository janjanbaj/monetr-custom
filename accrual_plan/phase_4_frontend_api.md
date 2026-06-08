# Phase 4: Frontend API Models & React Hooks

This phase establishes the client-side representation of the Accrual Expense objects and coordinates synchronization with the Go backend through React Query.

## Context Directories
- [interface/src/models/](file:///Users/janeetbajracharya/Desktop/Code/monetr/interface/src/models)
- [interface/src/hooks/](file:///Users/janeetbajracharya/Desktop/Code/monetr/interface/src/hooks)

---

## 1. Frontend Model Class

#### [NEW] [AccrualExpense.ts](file:///Users/janeetbajracharya/Desktop/Code/monetr/interface/src/models/AccrualExpense.ts)
```typescript
import parseDate from '@monetr/interface/util/parseDate';

export interface AccrualUsageLog {
  accrualUsageLogId: string;
  accrualExpenseId: string;
  amount: number;
  date: Date;
  createdAt: Date;
  updatedAt: Date;
}

export default class AccrualExpense {
  accrualExpenseId: string;
  bankAccountId: string;
  transactionId?: string;
  name: string;
  description?: string;
  amount: number;
  startDate: Date;
  endDate: Date;
  createdAt: Date;
  updatedAt: Date;
  usageLogs?: AccrualUsageLog[];

  constructor(data?: Partial<AccrualExpense>) {
    if (data) {
      Object.assign(this, {
        ...data,
        startDate: parseDate(data.startDate) || new Date(),
        endDate: parseDate(data.endDate) || new Date(),
        createdAt: parseDate(data.createdAt),
        updatedAt: parseDate(data.updatedAt),
        usageLogs: data.usageLogs?.map(log => ({
          ...log,
          date: parseDate(log.date) || new Date(),
          createdAt: parseDate(log.createdAt) || new Date(),
          updatedAt: parseDate(log.updatedAt) || new Date(),
        })) || [],
      });
    }
  }

  /**
   * Calculates daily recognized consumption.
   * Leverages custom Usage Logs if present; defaults to linear depreciation.
   */
  getDailyUsageMap(): Map<string, number> {
    const usage = new Map<string, number>();
    const totalDays = Math.ceil((this.endDate.getTime() - this.startDate.getTime()) / (1000 * 60 * 60 * 24)) + 1;
    if (totalDays <= 0) return usage;

    // 1. Identify logged dates and total manual amount
    const logs = this.usageLogs || [];
    let loggedAmount = 0;
    const loggedDates = new Set<string>();

    for (const log of logs) {
      const dateKey = log.date.toISOString().split('T')[0];
      loggedDates.add(dateKey);
      loggedAmount += log.amount;
      usage.set(dateKey, (usage.get(dateKey) || 0) + log.amount);
    }

    // 2. Linear depreciation for remaining days
    const remainingAmount = Math.max(0, this.amount - loggedAmount);
    const unloggedDaysCount = totalDays - loggedDates.size;

    if (remainingAmount > 0 && unloggedDaysCount > 0) {
      const dailyLinearAmount = Math.floor(remainingAmount / unloggedDaysCount);
      let remainder = remainingAmount % unloggedDaysCount;

      for (let i = 0; i < totalDays; i++) {
        const d = new Date(this.startDate);
        d.setDate(d.getDate() + i);
        const dateKey = d.toISOString().split('T')[0];

        if (!loggedDates.has(dateKey)) {
          let dayValue = dailyLinearAmount;
          if (remainder > 0) {
            dayValue += 1;
            remainder -= 1;
          }
          usage.set(dateKey, dayValue);
        }
      }
    } else if (remainingAmount === 0) {
      // If depleted early by manual logs, remaining days get 0
      for (let i = 0; i < totalDays; i++) {
        const d = new Date(this.startDate);
        d.setDate(d.getDate() + i);
        const dateKey = d.toISOString().split('T')[0];
        if (!usage.has(dateKey)) {
          usage.set(dateKey, 0);
        }
      }
    }

    return usage;
  }
}
```

---

## 2. React Query Hooks

#### [NEW] [useAccrualExpenses.ts](file:///Users/janeetbajracharya/Desktop/Code/monetr/interface/src/hooks/useAccrualExpenses.ts)
```typescript
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import request from '@monetr/interface/util/request';
import AccrualExpense from '@monetr/interface/models/AccrualExpense';

export function useAccrualExpenses(bankAccountId?: string) {
  return useQuery<AccrualExpense[]>({
    queryKey: ['bank_account', bankAccountId, 'accrual'],
    queryFn: () =>
      request(`/bank_accounts/${bankAccountId}/accrual`)
        .then(res => res.data as Partial<AccrualExpense>[])
        .then(list => list.map(item => new AccrualExpense(item))),
    enabled: Boolean(bankAccountId),
  });
}

export function useCreateAccrualExpense() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (payload: {
      bankAccountId: string;
      transactionId?: string;
      name: string;
      description?: string;
      amount: number;
      startDate: Date;
      endDate: Date;
    }) =>
      request.post(`/bank_accounts/${payload.bankAccountId}/accrual`, payload).then(res => new AccrualExpense(res.data)),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['bank_account', data.bankAccountId, 'accrual'] });
    },
  });
}

export function useUpdateAccrualExpense() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (payload: {
      bankAccountId: string;
      accrualExpenseId: string;
      name?: string;
      description?: string;
      amount?: number;
      startDate?: Date;
      endDate?: Date;
    }) =>
      request
        .put(`/bank_accounts/${payload.bankAccountId}/accrual/${payload.accrualExpenseId}`, payload)
        .then(res => new AccrualExpense(res.data)),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['bank_account', data.bankAccountId, 'accrual'] });
    },
  });
}

export function useDeleteAccrualExpense() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (payload: { bankAccountId: string; accrualExpenseId: string }) =>
      request.delete(`/bank_accounts/${payload.bankAccountId}/accrual/${payload.accrualExpenseId}`),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['bank_account', variables.bankAccountId, 'accrual'] });
    },
  });
}
```

---

## Verification Plan

### Automated Tests
1. Add front-end TypeScript checks and build test triggers:
   ```bash
   pnpm build
   ```
2. Verify that type imports resolve without issues and compilation finishes successfully.
3. Write a small Jest/Vitest test file `AccrualExpense.spec.ts` verifying that `getDailyUsageMap()` correctly distributes linear and log values, especially checking early depletion edge cases.
