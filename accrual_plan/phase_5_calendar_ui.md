# Phase 5: Interactive Gantt-style Calendar UI

This phase sets up wouter routing, sidebar entries, and lays out the visual calendar grid (horizontal Gantt timeline) for the accrual calendar.

## Context Directories
- [interface/src/](file:///Users/janeetbajracharya/Desktop/Code/monetr/interface/src)
- [interface/src/pages/](file:///Users/janeetbajracharya/Desktop/Code/monetr/interface/src/pages)

---

## 1. Sidebar Navigation Link

#### [MODIFY] [BudgetingSidebar.tsx](file:///Users/janeetbajracharya/Desktop/Code/monetr/interface/src/components/Layout/BudgetingSidebar.tsx)
Around line 84, insert the accrual calendar navigation link with the `Calendar` icon:
```typescript
          <NavigationItem to={`/bank/${bankAccount?.bankAccountId}/accrual`}>
            <CalendarSync />
            <Typography color='inherit' ellipsis size='lg' weight='medium'>
              Accrual Calendar
            </Typography>
          </NavigationItem>
```

---

## 2. Route Configuration

#### [MODIFY] [monetr.tsx](file:///Users/janeetbajracharya/Desktop/Code/monetr/interface/src/monetr.tsx)
Add import:
```typescript
import AccrualCalendar from '@monetr/interface/pages/accrual/AccrualCalendar';
```
And add route inside the `BudgetingLayout` Switch:
```typescript
              <Route component={AccrualCalendar} path='/bank/:bankAccountId/accrual' />
```

---

## 3. Base Calendar Grid Component

#### [NEW] [AccrualCalendar.tsx](file:///Users/janeetbajracharya/Desktop/Code/monetr/interface/src/pages/accrual/AccrualCalendar.tsx)
```typescript
import { useState } from 'react';
import { CalendarRange, Plus } from 'lucide-react';
import { useLocation } from 'wouter';

import { Button } from '@monetr/interface/components/Button';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import Typography from '@monetr/interface/components/Typography';
import { useAccrualExpenses } from '@monetr/interface/hooks/useAccrualExpenses';
import { useSelectedBankAccount } from '@monetr/interface/hooks/useSelectedBankAccount';

import styles from './AccrualCalendar.module.scss';

export default function AccrualCalendar(): JSX.Element {
  const [, setLocation] = useLocation();
  const { data: bankAccount } = useSelectedBankAccount();
  const { data: accruals, isLoading, isError } = useAccrualExpenses(bankAccount?.bankAccountId);

  // Focus range - default to the current month
  const [currentDate, setCurrentDate] = useState(new Date());

  if (isLoading) {
    return <div className={styles.loading}>Loading Accrual System...</div>;
  }

  if (isError || !bankAccount) {
    return <div className={styles.error}>Could not load bank account details.</div>;
  }

  return (
    <div className={styles.root}>
      <MTopNavigation icon={CalendarRange} title="Accrual Accounting">
        <Button onClick={() => {}} variant="primary">
          <Plus />
          Add Accrual
        </Button>
      </MTopNavigation>

      <div className={styles.layout}>
        <div className={styles.calendarContainer}>
          <CalendarHeader currentDate={currentDate} onChangeDate={setCurrentDate} />
          <TimelineGrid currentDate={currentDate} items={accruals || []} />
        </div>
      </div>
    </div>
  );
}

function CalendarHeader({ currentDate, onChangeDate }: { currentDate: Date; onChangeDate: (d: Date) => void }): JSX.Element {
  const monthName = currentDate.toLocaleString('default', { month: 'long', year: 'numeric' });

  const prevMonth = () => {
    const d = new Date(currentDate);
    d.setMonth(d.getMonth() - 1);
    onChangeDate(d);
  };

  const nextMonth = () => {
    const d = new Date(currentDate);
    d.setMonth(d.getMonth() + 1);
    onChangeDate(d);
  };

  return (
    <div className={styles.header}>
      <Typography size="2xl" weight="bold">{monthName}</Typography>
      <div className={styles.navButtons}>
        <Button onClick={prevMonth} variant="secondary">Prev</Button>
        <Button onClick={nextMonth} variant="secondary">Next</Button>
      </div>
    </div>
  );
}

function TimelineGrid({ currentDate, items }: { currentDate: Date; items: any[] }): JSX.Element {
  // Generate days for the current month
  const year = currentDate.getFullYear();
  const month = currentDate.getMonth();
  const daysInMonth = new Date(year, month + 1, 0).getDate();
  const days = Array.from({ length: daysInMonth }, (_, i) => new Date(year, month, i + 1));

  return (
    <div className={styles.grid}>
      <div className={styles.daysHeader}>
        {days.map(d => (
          <div key={d.getDate()} className={styles.dayCol}>
            <span className={styles.dayLabel}>{d.getDate()}</span>
            <span className={styles.dayWeek}>{d.toLocaleDateString('default', { weekday: 'narrow' })}</span>
          </div>
        ))}
      </div>

      <div className={styles.timelineRows}>
        {items.length === 0 ? (
          <div className={styles.emptyGrid}>No accrual items in this period. Drag a transaction or click Add to create one.</div>
        ) : (
          items.map(item => (
            <div key={item.accrualExpenseId} className={styles.row}>
              <div className={styles.rowLabel}>{item.name}</div>
              <div className={styles.rowTracks}>
                {/* Horizontal tracks with day divisions */}
                {days.map(d => (
                  <div key={d.getDate()} className={styles.trackCell} />
                ))}
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
}
```

---

## 4. Initial Styles

#### [NEW] [AccrualCalendar.module.scss](file:///Users/janeetbajracharya/Desktop/Code/monetr/interface/src/pages/accrual/AccrualCalendar.module.scss)
```scss
@use '@monetr/interface/variables';

.root {
  display: flex;
  flex-direction: column;
  flex: 1;
  height: 100%;
  overflow: hidden;
}

.loading, .error {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
  font-size: 1.5rem;
  color: var(--content-subtle);
}

.layout {
  display: flex;
  flex: 1;
  overflow: hidden;
}

.calendarContainer {
  display: flex;
  flex-direction: column;
  flex: 1;
  padding: 1.5rem;
  overflow: auto;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.navButtons {
  display: flex;
  gap: 0.5rem;
}

.grid {
  display: flex;
  flex-direction: column;
  border: 1px solid var(--border);
  background-color: var(--background-darker);
  border-radius: var(--border-radius);
  overflow: hidden;
}

.daysHeader {
  display: flex;
  border-bottom: 1px solid var(--border);
  background-color: var(--background-subtle);
}

.dayCol {
  flex: 1;
  min-width: 40px;
  text-align: center;
  padding: 0.5rem 0;
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;

  &:last-child {
    border-right: none;
  }
}

.dayLabel {
  font-weight: var(--bold);
  color: var(--content-emphasis);
}

.dayWeek {
  font-size: 0.75rem;
  color: var(--content-subtle);
}

.timelineRows {
  display: flex;
  flex-direction: column;
  min-height: 200px;
}

.emptyGrid {
  display: flex;
  justify-content: center;
  align-items: center;
  flex: 1;
  padding: 2rem;
  color: var(--content-subtle);
}

.row {
  display: flex;
  border-bottom: 1px solid var(--border);
  align-items: center;
  position: relative;
  height: 60px;

  &:last-child {
    border-bottom: none;
  }
}

.rowLabel {
  width: 150px;
  padding: 0 1rem;
  font-weight: var(--medium);
  color: var(--content-emphasis);
  border-right: 1px solid var(--border);
  flex-shrink: 0;
}

.rowTracks {
  display: flex;
  flex: 1;
  position: relative;
  height: 100%;
}

.trackCell {
  flex: 1;
  min-width: 40px;
  border-right: 1px solid var(--border);
  height: 100%;

  &:last-child {
    border-right: none;
  }
}
```

---

## Verification Plan

### Manual Verification
1. Run `make develop` to start the frontend server.
2. Select a bank account.
3. Click "Accrual Calendar" in the sidebar.
4. Verify the monthly calendar layout loads, shows days from 1 to the end of the month, and navigates forwards/backwards.
