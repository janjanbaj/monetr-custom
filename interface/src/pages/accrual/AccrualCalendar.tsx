import { Fragment, useRef, useState } from 'react';
import { CalendarRange, HeartCrack, Plus, ArrowRight } from 'lucide-react';
import { useQuery } from '@tanstack/react-query';

import { Button } from '@monetr/interface/components/Button';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import Typography from '@monetr/interface/components/Typography';
import { useAccrualExpenses } from '@monetr/interface/hooks/useAccrualExpenses';
import { useUpdateAccrualExpense } from '@monetr/interface/hooks/useUpdateAccrualExpense';
import { useCreateAccrualExpense } from '@monetr/interface/hooks/useCreateAccrualExpense';
import { useSelectedBankAccount } from '@monetr/interface/hooks/useSelectedBankAccount';
import AccrualExpense from '@monetr/interface/models/AccrualExpense';
import Transaction from '@monetr/interface/models/Transaction';
import { showNewAccrualModal } from '@monetr/interface/modals/NewAccrualModal';
import { showAccrualExpenseModal } from '@monetr/interface/modals/AccrualExpenseModal';
import request from '@monetr/interface/util/request';

import styles from './AccrualCalendar.module.scss';

// Hook to retrieve bank transactions (with a high limit or filter search)
function useUnallocatedTransactions(bankAccountId?: string) {
  return useQuery<Transaction[]>({
    queryKey: ['bank_account', bankAccountId, 'transactions_unallocated'],
    queryFn: () =>
      request<Partial<Transaction>[]>({
        method: 'GET',
        url: `/api/bank_accounts/${bankAccountId}/transactions?limit=100`,
      })
        .then(res => (res.data || []))
        .then(list => list.map(item => new Transaction(item))),
    enabled: Boolean(bankAccountId),
  });
}

export default function AccrualCalendar(): JSX.Element {
  const { data: bankAccount } = useSelectedBankAccount();
  const { data: accruals, isLoading, isError } = useAccrualExpenses();
  const { data: transactions } = useUnallocatedTransactions(bankAccount?.bankAccountId);
  const createAccrual = useCreateAccrualExpense();
  const [currentDate, setCurrentDate] = useState(new Date());

  const handleDragStart = (e: React.DragEvent, txn: Transaction) => {
    e.dataTransfer.setData(
      'text/plain',
      JSON.stringify({
        transactionId: txn.transactionId,
        name: txn.name || txn.originalName,
        amount: txn.amount,
      })
    );
  };

  const handleDropOnDay = (e: React.DragEvent, targetDate: Date) => {
    e.preventDefault();
    if (!bankAccount) return;
    try {
      const raw = e.dataTransfer.getData('text/plain');
      if (!raw) return;
      const data = JSON.parse(raw);

      // Default duration is 30 days starting on the dropped day
      const endDate = new Date(targetDate);
      endDate.setDate(endDate.getDate() + 30);

      const newAccrual = new AccrualExpense({
        bankAccountId: bankAccount.bankAccountId,
        transactionId: data.transactionId,
        name: data.name,
        amount: data.amount,
        startDate: targetDate,
        endDate: endDate,
      });

      void createAccrual(newAccrual);
    } catch (err) {
      console.error('Failed to parse dropped transaction data', err);
    }
  };

  if (isLoading) {
    return (
      <div className={styles.loading}>
        <Typography size='5xl'>One moment...</Typography>
      </div>
    );
  }

  if (isError) {
    return (
      <div className={styles.error}>
        <HeartCrack className={styles.error} />
        <Typography size='5xl'>Something isn't right...</Typography>
        <Typography size='2xl'>We weren't able to retrieve accrual data at this time...</Typography>
      </div>
    );
  }

  // Client-side filter: only show transactions that haven't been allocated to an accrual
  const allocatedTxnIds = new Set(
    (accruals || []).map(a => a.transactionId).filter(Boolean)
  );
  const unallocatedTransactions = (transactions || []).filter(
    t => !allocatedTxnIds.has(t.transactionId)
  );

  return (
    <Fragment>
      <MTopNavigation icon={CalendarRange} title='Accrual Calendar'>
        <Button onClick={showNewAccrualModal} variant='primary'>
          <Plus />
          Add Accrual
        </Button>
      </MTopNavigation>
      <div className={styles.layout}>
        <div className={styles.calendarContainer}>
          <CalendarHeader currentDate={currentDate} onChangeDate={setCurrentDate} />
          <TimelineGrid
            currentDate={currentDate}
            items={accruals || []}
            onDropOnDay={handleDropOnDay}
          />
          <AnalyticsSummary
            items={accruals || []}
            transactions={transactions || []}
            currentDate={currentDate}
          />
        </div>

        <div className={styles.sidebar}>
          <Typography className={styles.sidebarTitle} size='lg' weight='bold'>
            Recent Transactions
          </Typography>
          <Typography className={styles.sidebarSubtitle} size='sm' color='subtle'>
            Drag and drop a transaction onto any calendar day to establish a new accrual.
          </Typography>
          <div className={styles.txnList}>
            {unallocatedTransactions.length === 0 ? (
              <div className={styles.emptySidebar}>
                All recent transactions are allocated.
              </div>
            ) : (
              unallocatedTransactions.map(txn => (
                <div
                  key={txn.transactionId}
                  className={styles.txnCard}
                  draggable
                  onDragStart={e => handleDragStart(e, txn)}
                >
                  <div className={styles.txnName}>{txn.name || txn.originalName}</div>
                  <div className={styles.txnMeta}>
                    <span className={styles.txnAmount}>
                      ${(txn.amount / 100).toFixed(2)}
                    </span>
                    <ArrowRight className={styles.txnDragIcon} size={14} />
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      </div>
    </Fragment>
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
      <Typography size='2xl' weight='bold'>{monthName}</Typography>
      <div className={styles.navButtons}>
        <Button onClick={prevMonth} variant='secondary'>Prev</Button>
        <Button onClick={nextMonth} variant='secondary'>Next</Button>
      </div>
    </div>
  );
}

interface DragState {
  expense: AccrualExpense;
  type: 'move' | 'resize-start' | 'resize-end';
  startX: number;
  originalStartDate: Date;
  originalEndDate: Date;
}

interface TimelineGridProps {
  currentDate: Date;
  items: AccrualExpense[];
  onDropOnDay: (e: React.DragEvent, d: Date) => void;
}

function TimelineGrid({ currentDate, items, onDropOnDay }: TimelineGridProps): JSX.Element {
  const updateExpense = useUpdateAccrualExpense();
  const year = currentDate.getFullYear();
  const month = currentDate.getMonth();
  const daysInMonth = new Date(year, month + 1, 0).getDate();
  const days = Array.from({ length: daysInMonth }, (_, i) => new Date(year, month, i + 1));

  const trackRef = useRef<HTMLDivElement>(null);
  const dragStateRef = useRef<DragState | null>(null);
  const [, forceRender] = useState(0);

  // Drag-over hover state for column dropping
  const [hoveredDateStr, setHoveredDateStr] = useState<string | null>(null);

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

    forceRender(n => n + 1);
  };

  const handlePointerUp = (e: React.PointerEvent) => {
    if (!dragStateRef.current) return;
    const { expense } = dragStateRef.current;
    (e.target as HTMLElement).releasePointerCapture(e.pointerId);
    dragStateRef.current = null;

    updateExpense(expense);
  };

  const handleDragOverDay = (e: React.DragEvent, d: Date) => {
    e.preventDefault();
    const key = d.toISOString().split('T')[0];
    if (hoveredDateStr !== key) {
      setHoveredDateStr(key);
    }
  };

  const handleDragLeaveDay = () => {
    setHoveredDateStr(null);
  };

  const handleDropOnDayLocal = (e: React.DragEvent, d: Date) => {
    setHoveredDateStr(null);
    onDropOnDay(e, d);
  };

  return (
    <div
      className={styles.grid}
      onPointerMove={handlePointerMove}
      onPointerUp={handlePointerUp}
    >
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
          <div className={styles.emptyContainer}>
            <div className={styles.emptyGrid}>No accrual items in this period. Drag a transaction here to start tracking.</div>
            <div className={styles.ghostRow}>
              <div className={styles.rowLabelSpacer} />
              <div className={styles.rowTracks}>
                {days.map(d => {
                  const dateKey = d.toISOString().split('T')[0];
                  const isHovered = hoveredDateStr === dateKey;
                  return (
                    <div
                      key={d.getDate()}
                      className={`${styles.trackCell} ${isHovered ? styles.trackCellDragOver : ''}`}
                      onDragOver={e => handleDragOverDay(e, d)}
                      onDragLeave={handleDragLeaveDay}
                      onDrop={e => handleDropOnDayLocal(e, d)}
                    />
                  );
                })}
              </div>
            </div>
          </div>
        ) : (
          items.map(item => {
            const itemStartMs = item.startDate.getTime();
            const itemEndMs = item.endDate.getTime();
            const monthStartMs = new Date(year, month, 1).getTime();
            const monthEndMs = new Date(year, month, daysInMonth, 23, 59, 59).getTime();

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
                  {days.map(d => {
                    const dateKey = d.toISOString().split('T')[0];
                    const isHovered = hoveredDateStr === dateKey;
                    return (
                      <div
                        key={d.getDate()}
                        className={`${styles.trackCell} ${isHovered ? styles.trackCellDragOver : ''}`}
                        onDragOver={e => handleDragOverDay(e, d)}
                        onDragLeave={handleDragLeaveDay}
                        onDrop={e => handleDropOnDayLocal(e, d)}
                      />
                    );
                  })}

                  <div
                    className={styles.accrualBlock}
                    style={{ left: `${leftPercent}%`, width: `${Math.max(widthPercent, 1)}%` }}
                    onPointerDown={e => handlePointerDown(e, item, 'move')}
                    onDoubleClick={() => showAccrualExpenseModal(item)}
                    title="Double click to edit details and log daily consumption"
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

function AnalyticsSummary({
  items,
  transactions,
  currentDate,
}: {
  items: AccrualExpense[];
  transactions: Transaction[];
  currentDate: Date;
}): JSX.Element {
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
          <span>
            Accrual adjusted spending is lower by <strong>${((totalCashOutflow - totalAccrualExpense) / 100).toFixed(2)}</strong> than cash spent due to bulk items/subscriptions spread over time.
          </span>
        ) : (
          <span>
            Accrual adjusted spending is higher by <strong>${((totalAccrualExpense - totalCashOutflow) / 100).toFixed(2)}</strong> due to utilization of previously purchased bulk assets.
          </span>
        )}
      </div>
    </div>
  );
}


