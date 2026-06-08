# Phase 6: Drag-and-Drop & Resizing Interactions

This phase adds pointer-based drag/resize features to calendar blocks on the horizontal Gantt timeline. It implements reactive drag coordinates (60fps) and triggers database updates when the interaction finishes.

## Context Directories
- [interface/src/pages/accrual/](file:///Users/janeetbajracharya/Desktop/Code/monetr/interface/src/pages/accrual)

---

## 1. Interactive Timeline Row & Block Layout

#### [MODIFY] [AccrualCalendar.tsx](file:///Users/janeetbajracharya/Desktop/Code/monetr/interface/src/pages/accrual/AccrualCalendar.tsx)
Replace `TimelineGrid` with the interactive implementation showing absolute-positioned blocks that handle mouse/touch resizing and sliding:

```typescript
import { useRef } from 'react';
import AccrualExpense from '@monetr/interface/models/AccrualExpense';
import { useUpdateAccrualExpense } from '@monetr/interface/hooks/useAccrualExpenses';

interface DragState {
  expense: AccrualExpense;
  type: 'move' | 'resize-start' | 'resize-end';
  startX: number;
  originalStartDate: Date;
  originalEndDate: Date;
}

function TimelineGrid({ currentDate, items }: { currentDate: Date; items: AccrualExpense[] }): JSX.Element {
  const updateExpense = useUpdateAccrualExpense();
  const year = currentDate.getFullYear();
  const month = currentDate.getMonth();
  const daysInMonth = new Date(year, month + 1, 0).getDate();
  const days = Array.from({ length: daysInMonth }, (_, i) => new Date(year, month, i + 1));

  const trackRef = useRef<HTMLDivElement>(null);
  const dragStateRef = useRef<DragState | null>(null);

  // Helper: map coordinate X to Date
  const getDateFromX = (clientX: number): Date | null => {
    if (!trackRef.current) return null;
    const rect = trackRef.current.getBoundingClientRect();
    const relativeX = clientX - rect.left;
    const colWidth = rect.width / daysInMonth;
    const dayIndex = Math.floor(relativeX / colWidth);
    const clampedIndex = Math.max(0, Math.min(daysInMonth - 1, dayIndex));
    return new Date(year, month, clampedIndex + 1);
  };

  const handlePointerDown = (
    e: React.PointerEvent,
    item: AccrualExpense,
    type: 'move' | 'resize-start' | 'resize-end'
  ) => {
    e.stopPropagation();
    (e.target as HTMLElement).setPointerCapture(e.pointerId);

    dragStateRef.current = {
      expense: item,
      type,
      startX: e.clientX,
      originalStartDate: new Date(item.startDate),
      originalEndDate: new Date(item.endDate),
    };
  };

  const handlePointerMove = (e: React.PointerEvent) => {
    if (!dragStateRef.current || !trackRef.current) return;
    const { expense, type, startX, originalStartDate, originalEndDate } = dragStateRef.current;

    const rect = trackRef.current.getBoundingClientRect();
    const colWidth = rect.width / daysInMonth;
    const deltaDays = Math.round((e.clientX - startX) / colWidth);

    if (deltaDays === 0) return;

    if (type === 'move') {
      const newStart = new Date(originalStartDate);
      newStart.setDate(newStart.getDate() + deltaDays);
      const newEnd = new Date(originalEndDate);
      newEnd.setDate(newEnd.getDate() + deltaDays);

      // Perform local visual updates (re-render item locally if desired, or let state handle)
      expense.startDate = newStart;
      expense.endDate = newEnd;
    } else if (type === 'resize-start') {
      const newStart = new Date(originalStartDate);
      newStart.setDate(newStart.getDate() + deltaDays);
      if (newStart <= originalEndDate) {
        expense.startDate = newStart;
      }
    } else if (type === 'resize-end') {
      const newEnd = new Date(originalEndDate);
      newEnd.setDate(newEnd.getDate() + deltaDays);
      if (newEnd >= originalStartDate) {
        expense.endDate = newEnd;
      }
    }
  };

  const handlePointerUp = (e: React.PointerEvent) => {
    if (!dragStateRef.current) return;
    const { expense } = dragStateRef.current;
    (e.target as HTMLElement).releasePointerCapture(e.pointerId);
    dragStateRef.current = null;

    // Persist changes to database
    updateExpense.mutate({
      bankAccountId: expense.bankAccountId,
      accrualExpenseId: expense.accrualExpenseId,
      startDate: expense.startDate,
      endDate: expense.endDate,
    });
  };

  return (
    <div className={styles.grid}>
      <div className={styles.daysHeader}>
        <div className={styles.rowLabelSpacer} />
        <div className={styles.headerDaysWrapper} ref={trackRef}>
          {days.map(d => (
            <div key={d.getDate()} className={styles.dayCol}>
              <span className={styles.dayLabel}>{d.getDate()}</span>
              <span className={styles.dayWeek}>{d.toLocaleDateString('default', { weekday: 'narrow' })}</span>
            </div>
          ))}
        </div>
      </div>

      <div className={styles.timelineRows}>
        {items.length === 0 ? (
          <div className={styles.emptyGrid}>No accrual items in this period.</div>
        ) : (
          items.map(item => {
            // Compute percentage bounds for absolute positioning
            const itemStartMs = item.startDate.getTime();
            const itemEndMs = item.endDate.getTime();
            const monthStartMs = new Date(year, month, 1).getTime();
            const monthEndMs = new Date(year, month, daysInMonth, 23, 59, 59).getTime();

            // Clamping dates to current month window
            const displayStart = Math.max(monthStartMs, itemStartMs);
            const displayEnd = Math.min(monthEndMs, itemEndMs);

            if (displayEnd < monthStartMs || displayStart > monthEndMs) return null;

            const totalMonthDuration = monthEndMs - monthStartMs;
            const leftPercent = ((displayStart - monthStartMs) / totalMonthDuration) * 100;
            const widthPercent = ((displayEnd - displayStart) / totalMonthDuration) * 100;

            return (
              <div key={item.accrualExpenseId} className={styles.row}>
                <div className={styles.rowLabel}>{item.name}</div>
                <div className={styles.rowTracks}>
                  {days.map(d => (
                    <div key={d.getDate()} className={styles.trackCell} />
                  ))}
                  
                  {/* Absolute positioned block */}
                  <div
                    className={styles.accrualBlock}
                    style={{ left: `${leftPercent}%`, width: `${widthPercent}%` }}
                    onPointerDown={e => handlePointerDown(e, item, 'move')}
                    onPointerMove={handlePointerMove}
                    onPointerUp={handlePointerUp}
                  >
                    <div
                      className={`${styles.resizeHandle} ${styles.handleLeft}`}
                      onPointerDown={e => handlePointerDown(e, item, 'resize-start')}
                    />
                    <div className={styles.blockContent}>
                      <span className={styles.blockTitle}>{item.name}</span>
                      <span className={styles.blockAmount}>
                        ${(item.amount / 100).toFixed(2)}
                      </span>
                    </div>
                    <div
                      className={`${styles.resizeHandle} ${styles.handleRight}`}
                      onPointerDown={e => handlePointerDown(e, item, 'resize-end')}
                    />
                  </div>
                </div>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}
```

---

## 2. Dynamic Styles for Blocks and Handles

#### [MODIFY] [AccrualCalendar.module.scss](file:///Users/janeetbajracharya/Desktop/Code/monetr/interface/src/pages/accrual/AccrualCalendar.module.scss)
Append styles for interactive tracks and absolute dragging positions:

```scss
.rowLabelSpacer {
  width: 150px;
  border-right: 1px solid var(--border);
  flex-shrink: 0;
  background-color: var(--background-darker);
}

.headerDaysWrapper {
  display: flex;
  flex: 1;
}

.accrualBlock {
  position: absolute;
  top: 10%;
  height: 80%;
  background: linear-gradient(135deg, var(--brand) 0%, var(--brand-subtle) 100%);
  border-radius: var(--border-radius);
  border: 1px solid var(--brand-muted);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: space-between;
  cursor: grab;
  user-select: none;
  touch-action: none;
  overflow: hidden;

  &:active {
    cursor: grabbing;
    border-color: var(--brand-bright);
  }
}

.blockContent {
  display: flex;
  flex-direction: column;
  padding: 0 0.5rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.blockTitle {
  font-weight: var(--semibold);
  font-size: 0.85rem;
  color: var(--content-emphasis);
}

.blockAmount {
  font-size: 0.75rem;
  color: var(--brand-bright);
}

.resizeHandle {
  width: 8px;
  height: 100%;
  cursor: ew-resize;
  background-color: rgba(255, 255, 255, 0.1);
  transition: background-color 0.2s ease;

  &:hover {
    background-color: var(--brand-bright);
  }
}

.handleLeft {
  border-top-left-radius: var(--border-radius);
  border-bottom-left-radius: var(--border-radius);
}

.handleRight {
  border-top-right-radius: var(--border-radius);
  border-bottom-right-radius: var(--border-radius);
}
```

---

## Verification Plan

### Manual Verification
1. Open the browser and go to `/bank/:bankAccountId/accrual`.
2. Grab the middle of a block and slide it left or right. Check if dates update in local state.
3. Grab the left edge or right edge and stretch/shrink it.
4. Release the pointer click. Verify that a PUT request is dispatched to `/bank_accounts/:bankAccountId/accrual/:accrualExpenseId` with updated dates, and the query cache invalidates to reload the updated bounds.
