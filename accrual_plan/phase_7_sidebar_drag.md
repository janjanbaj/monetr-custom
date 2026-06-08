# Phase 7: Transaction Sidebar Panel & Drag-to-Create

This phase builds the Sidebar Panel, retrieves transactions, and implements HTML5 drag-and-drop, allowing transactions to be dropped directly onto the timeline grid to generate new accrual schedules.

## Context Directories
- [interface/src/pages/accrual/](file:///Users/janeetbajracharya/Desktop/Code/monetr/interface/src/pages/accrual)

---

## 1. Sidebar Implementation & Drag Triggers

#### [MODIFY] [AccrualCalendar.tsx](file:///Users/janeetbajracharya/Desktop/Code/monetr/interface/src/pages/accrual/AccrualCalendar.tsx)
Integrate the sidebar panel list showing transactions and drop zones:

```typescript
import { useQuery } from '@tanstack/react-query';
import request from '@monetr/interface/util/request';
import Transaction from '@monetr/interface/models/Transaction';
import { useCreateAccrualExpense } from '@monetr/interface/hooks/useAccrualExpenses';

// Hook to retrieve bank transactions (with a high limit or filter search)
function useUnallocatedTransactions(bankAccountId?: string) {
  return useQuery<Transaction[]>({
    queryKey: ['bank_account', bankAccountId, 'transactions_unallocated'],
    queryFn: () =>
      request(`/bank_accounts/${bankAccountId}/transactions?limit=100`)
        .then(res => res.data as Partial<Transaction>[])
        .then(list => list.map(item => new Transaction(item))),
    enabled: Boolean(bankAccountId),
  });
}

export default function AccrualCalendar(): JSX.Element {
  const { data: bankAccount } = useSelectedBankAccount();
  const { data: accruals, isLoading } = useAccrualExpenses(bankAccount?.bankAccountId);
  const { data: transactions } = useUnallocatedTransactions(bankAccount?.bankAccountId);
  const createAccrual = useCreateAccrualExpense();
  const [currentDate, setCurrentDate] = useState(new Date());

  const handleDragStart = (e: React.DragEvent, txn: Transaction) => {
    e.dataTransfer.setData('text/plain', JSON.stringify({
      transactionId: txn.transactionId,
      name: txn.name || txn.originalName,
      amount: txn.amount,
    }));
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault(); // Required to allow drop!
  };

  const handleDropOnDay = (e: React.DragEvent, targetDate: Date) => {
    e.preventDefault();
    try {
      const data = JSON.parse(e.dataTransfer.getData('text/plain'));
      
      // Default duration is 30 days starting on the dropped day
      const endDate = new Date(targetDate);
      endDate.setDate(endDate.getDate() + 30);

      createAccrual.mutate({
        bankAccountId: bankAccount!.bankAccountId,
        transactionId: data.transactionId,
        name: data.name,
        amount: data.amount,
        startDate: targetDate,
        endDate: endDate,
      });
    } catch (err) {
      console.error('Failed to parse dropped transaction data', err);
    }
  };

  if (isLoading) return <div className={styles.loading}>Loading...</div>;

  return (
    <div className={styles.root}>
      <MTopNavigation icon={CalendarRange} title="Accrual Accounting" />

      <div className={styles.layout}>
        {/* Timeline Grid */}
        <div className={styles.calendarContainer}>
          <CalendarHeader currentDate={currentDate} onChangeDate={setCurrentDate} />
          <TimelineGrid
            currentDate={currentDate}
            items={accruals || []}
            onDragOver={handleDragOver}
            onDropOnDay={handleDropOnDay}
          />
        </div>

        {/* Sidebar Transactions List */}
        <div className={styles.sidebar}>
          <Typography className={styles.sidebarTitle} size="lg" weight="bold">
            Recent Transactions
          </Typography>
          <div className={styles.txnList}>
            {transactions?.map(txn => (
              <div
                key={txn.transactionId}
                className={styles.txnCard}
                draggable
                onDragStart={e => handleDragStart(e, txn)}
              >
                <div className={styles.txnName}>{txn.name || txn.originalName}</div>
                <div className={styles.txnAmount}>
                  ${(txn.amount / 100).toFixed(2)}
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
```

Wait, ensure `TimelineGrid` calls `onDragOver` and `onDropOnDay`:
```typescript
// Add these props to TimelineGrid:
// onDragOver: (e: React.DragEvent) => void
// onDropOnDay: (e: React.DragEvent, d: Date) => void

// Inside TimelineGrid rendering of track cells, attach drop handlers:
<div 
  key={d.getDate()} 
  className={styles.trackCell} 
  onDragOver={onDragOver}
  onDrop={e => onDropOnDay(e, d)}
/>
```

---

## 2. Layout & Card Styles

#### [MODIFY] [AccrualCalendar.module.scss](file:///Users/janeetbajracharya/Desktop/Code/monetr/interface/src/pages/accrual/AccrualCalendar.module.scss)
Append styles for the split sidebar layout:

```scss
.sidebar {
  width: 280px;
  border-left: 1px solid var(--border);
  background-color: var(--background-darker);
  display: flex;
  flex-direction: column;
  padding: 1rem;
  overflow-y: auto;
}

.sidebarTitle {
  margin-bottom: 1rem;
  color: var(--content-emphasis);
}

.txnList {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.txnCard {
  background-color: var(--background-subtle);
  border: 1px solid var(--border);
  border-radius: var(--border-radius);
  padding: 0.75rem;
  cursor: grab;
  user-select: none;
  transition: transform 0.2s ease, border-color 0.2s ease;

  &:hover {
    border-color: var(--brand-muted);
    transform: translateY(-2px);
  }

  &:active {
    cursor: grabbing;
  }
}

.txnName {
  font-weight: var(--medium);
  color: var(--content-emphasis);
  font-size: 0.85rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.txnAmount {
  font-size: 0.8rem;
  color: var(--content-subtle);
  margin-top: 0.25rem;
}
```

---

## 3. Extra Features for Accurial Recognition:
- Allow for certain days for Accurial items to be exempted from. For example if I eat out some day or do not use the rice some day, I can click to skip the recognition of the expense that day. 


## 4. Optional Auto-Transaction:
- When I recognize the expense today, I can have an option to create a transaction for it that ties into the default system of monetr.

## Verification Plan

### Manual Verification
1. Open the Accrual Calendar page.
2. Verify that the right sidebar shows the scrollable list of recent transactions.
3. Drag a transaction from the sidebar, hover it over a day column on the timeline, and drop it.
4. Verify a POST request is sent to `/bank_accounts/:bankAccountId/accrual` with the transaction ID, name, amount, and the drop date as `startDate`. Verify the grid reloads showing the new calendar block.
