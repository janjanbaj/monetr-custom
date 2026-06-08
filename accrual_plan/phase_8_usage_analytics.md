# Phase 8: Usage Log Entry Modal & Amortization Analytics

This phase implements the detail modal for managing custom daily usage events and displays analytics comparing cash-flow expenditures versus accrual-recognized expenses.

## Context Directories
- [interface/src/modals/](file:///Users/janeetbajracharya/Desktop/Code/monetr/interface/src/modals)
- [interface/src/pages/accrual/](file:///Users/janeetbajracharya/Desktop/Code/monetr/interface/src/pages/accrual)

---

## 1. Accrual Edit & Usage Modal

#### [NEW] [AccrualExpenseModal.tsx](file:///Users/janeetbajracharya/Desktop/Code/monetr/interface/src/modals/AccrualExpenseModal.tsx)
Creates a modal enabling users to view the details of an accrual item, log manual consumption values, and adjust metadata fields:

```typescript
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { useRef } from 'react';
import { Button } from '@monetr/interface/components/Button';
import Typography from '@monetr/interface/components/Typography';
import MModal, { type MModalRef } from '@monetr/interface/components/MModal';
import AccrualExpense from '@monetr/interface/models/AccrualExpense';
import { useDeleteAccrualExpense } from '@monetr/interface/hooks/useAccrualExpenses';

import styles from './AccrualExpenseModal.module.scss';

interface AccrualExpenseModalProps {
  expense: AccrualExpense;
}

function AccrualExpenseModal({ expense }: AccrualExpenseModalProps): JSX.Element {
  const modal = useModal();
  const deleteExpense = useDeleteAccrualExpense();
  const ref = useRef<MModalRef>(null);

  const handleDelete = () => {
    deleteExpense.mutate({
      bankAccountId: expense.bankAccountId,
      accrualExpenseId: expense.accrualExpenseId,
    }, {
      onSuccess: () => {
        modal.resolve();
        modal.remove();
      }
    });
  };

  return (
    <MModal className={styles.modal} open={modal.visible} ref={ref}>
      <div className={styles.body}>
        <Typography className={styles.heading} size="xl" weight="bold">
          {expense.name}
        </Typography>
        <Typography color="subtle" size="sm">
          Total Purchase: ${(expense.amount / 100).toFixed(2)}
        </Typography>

        {/* Dynamic Usage Section */}
        <div className={styles.usageSection}>
          <Typography size="lg" weight="semibold">Log Daily Consumption</Typography>
          <p className={styles.description}>
            Log specific usage amounts to track faster depletion. Unlogged days automatically divide the remaining balance linearly.
          </p>
          {/* Form fields to add new UsageLog entries using mutation */}
        </div>

        <div className={styles.actions}>
          <Button onClick={handleDelete} variant="destructive">
            Delete Accrual
          </Button>
          <Button onClick={modal.remove} variant="secondary">
            Close
          </Button>
        </div>
      </div>
    </MModal>
  );
}

const accrualExpenseModal = NiceModal.create(AccrualExpenseModal);
export default accrualExpenseModal;

export function showAccrualExpenseModal(expense: AccrualExpense): Promise<void> {
  return NiceModal.show(accrualExpenseModal, { expense });
}
```

---

## 2. Analytics Dashboard

#### [MODIFY] [AccrualCalendar.tsx](file:///Users/janeetbajracharya/Desktop/Code/monetr/interface/src/pages/accrual/AccrualCalendar.tsx)
Add an Analytics Bar at the bottom of the page compiling Cash Spent vs Accrual Recognized in the currently viewed month, day or selected time-frame. :

```typescript
function AnalyticsSummary({ items, transactions, currentDate }: {
  items: AccrualExpense[];
  transactions: Transaction[];
  currentDate: Date;
}) {
  const year = currentDate.getFullYear();
  const month = currentDate.getMonth();

  // 1. Calculate Cash outflow from linked transactions
  const totalCashOutflow = transactions
    .filter(t => t.date.getFullYear() === year && t.date.getMonth() === month && t.amount > 0)
    .reduce((sum, t) => sum + t.amount, 0);

  // 2. Calculate Accrual recognized consumption for this month
  let totalAccrualExpense = 0;
  for (const item of items) {
    const dailyMap = item.getDailyUsageMap();
    for (const [dateStr, amount] of dailyMap.entries()) {
      const date = new Date(dateStr + 'T00:00:00');
      if (date.getFullYear() === year && date.getMonth() === month) {
        totalAccrualExpense += amount;
      }
    }
  }

  return (
    <div className={styles.analyticsBar}>
      <div className={styles.metricCard}>
        <div className={styles.metricLabel}>Cash Outflow (This Month)</div>
        <div className={styles.metricValue}>
          ${(totalCashOutflow / 100).toFixed(2)}
        </div>
      </div>
      <div className={styles.metricCard}>
        <div className={styles.metricLabel}>Accrual Consumption (Recognized)</div>
        <div className={`${styles.metricValue} ${styles.accrualValue}`}>
          ${(totalAccrualExpense / 100).toFixed(2)}
        </div>
      </div>
      <div className={styles.metricExplanation}>
        {totalCashOutflow > totalAccrualExpense ? (
          <span>Accrual adjusted spending is lower by <strong>${((totalCashOutflow - totalAccrualExpense)/100).toFixed(2)}</strong> than cash spent due to bulk items/subscriptions spread over time.</span>
        ) : (
          <span>Accrual adjusted spending is higher by <strong>${((totalAccrualExpense - totalCashOutflow)/100).toFixed(2)}</strong> due to utilization of previously purchased bulk assets.</span>
        )}
      </div>
    </div>
  );
}
```

Add the `<AnalyticsSummary>` component at the bottom of the main calendar page layout in `AccrualCalendar`.

---

## Verification Plan

### Manual Verification
1. Run `make develop` and open the app.
2. Select the Accrual Calendar page.
3. Double click on a calendar block to trigger `showAccrualExpenseModal`.
4. Log manual consumption entries on specific days (e.g. $10 used today). Verify that the block recalculates linear rates for all remaining days and updates correctly.
5. Review the Analytics Card showing Cash Spent vs Accrual recognized to confirm that values align mathematically with transactions and amortization maps.
