import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { format, startOfDay } from 'date-fns';
import { AlertCircle, Calendar, Trash2, X, Check, CheckSquare, Square } from 'lucide-react';
import { useState, useRef } from 'react';

import { Button } from '@monetr/interface/components/Button';
import MModal, { type MModalRef } from '@monetr/interface/components/MModal';
import Typography from '@monetr/interface/components/Typography';
import { useCreateAccrualUsageLog } from '@monetr/interface/hooks/useCreateAccrualUsageLog';
import { useDeleteAccrualUsageLog } from '@monetr/interface/hooks/useDeleteAccrualUsageLog';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useRemoveAccrualExpense } from '@monetr/interface/hooks/useRemoveAccrualExpense';
import { useCreateTransaction } from '@monetr/interface/hooks/useCreateTransaction';
import AccrualExpense from '@monetr/interface/models/AccrualExpense';
import { useSnackbar } from '@monetr/notify';

import styles from './AccrualExpenseModal.module.scss';

interface AccrualExpenseModalProps {
  expense: AccrualExpense;
}

function AccrualExpenseModal({ expense }: AccrualExpenseModalProps): JSX.Element {
  const modal = useModal();
  const removeExpense = useRemoveAccrualExpense();
  const createUsageLog = useCreateAccrualUsageLog();
  const deleteUsageLog = useDeleteAccrualUsageLog();
  const createTransaction = useCreateTransaction();
  const { enqueueSnackbar } = useSnackbar();
  
  const {
    data: { friendlyToAmount, amountToFriendly },
  } = useLocaleCurrency();

  const ref = useRef<MModalRef>(null);

  // State to track custom consumption form
  const [editingDate, setEditingDate] = useState<string | null>(null);
  const [customAmount, setCustomAmount] = useState<string>('');
  const [autoCreateTxn, setAutoCreateTxn] = useState<boolean>(false);

  const handleDeleteExpense = () => {
    removeExpense(expense.accrualExpenseId)
      .then(() => {
        void enqueueSnackbar('Accrual expense deleted', { variant: 'success' });
        modal.resolve();
        modal.remove();
      })
      .catch((err: any) => {
        void enqueueSnackbar(err.response?.data?.error || 'Failed to delete accrual', { variant: 'error' });
      });
  };

  const handleExemptDay = (date: Date) => {
    createUsageLog({
      accrualExpenseId: expense.accrualExpenseId,
      amount: 0,
      date,
    })
      .then(() => {
        void enqueueSnackbar('Day exempted successfully', { variant: 'success' });
      })
      .catch((err: any) => {
        void enqueueSnackbar(err.response?.data?.error || 'Failed to exempt day', { variant: 'error' });
      });
  };

  const handleRemoveLog = (logId: string) => {
    deleteUsageLog(logId)
      .then(() => {
        void enqueueSnackbar('Usage log removed', { variant: 'success' });
      })
      .catch((err: any) => {
        void enqueueSnackbar(err.response?.data?.error || 'Failed to remove log', { variant: 'error' });
      });
  };

  const handleSaveCustomUsage = (date: Date, dateStr: string) => {
    const amountVal = parseFloat(customAmount);
    if (isNaN(amountVal) || amountVal <= 0) {
      void enqueueSnackbar('Please enter a valid amount greater than 0', { variant: 'error' });
      return;
    }

    const cents = friendlyToAmount(amountVal);

    createUsageLog({
      accrualExpenseId: expense.accrualExpenseId,
      amount: cents,
      date,
    })
      .then(() => {
        if (autoCreateTxn) {
          return createTransaction({
            name: `${expense.name} (Accrual Usage)`,
            bankAccountId: expense.bankAccountId,
            amount: cents,
            spendingId: null,
            date,
            merchantName: null,
            isPending: false,
            adjustsBalance: true,
          });
        }
        return null;
      })
      .then(() => {
        void enqueueSnackbar('Custom usage logged successfully', { variant: 'success' });
        setEditingDate(null);
        setCustomAmount('');
        setAutoCreateTxn(false);
      })
      .catch((err: any) => {
        void enqueueSnackbar(err.response?.data?.error || 'Failed to log usage', { variant: 'error' });
      });
  };

  // Generate days in accrual range
  const days: Date[] = [];
  const curr = new Date(expense.startDate);
  const end = new Date(expense.endDate);
  while (curr <= end) {
    days.push(new Date(curr));
    curr.setDate(curr.getDate() + 1);
  }

  const dailyMap = expense.getDailyUsageMap();
  const totalLogged = (expense.usageLogs || []).reduce((sum, log) => sum + log.amount, 0);
  const remainingAmount = Math.max(0, expense.amount - totalLogged);

  return (
    <MModal className={styles.modal} open={modal.visible} ref={ref}>
      <div className={styles.body}>
        <div className={styles.header}>
          <Typography className={styles.heading} size='2xl' weight='bold'>
            {expense.name}
          </Typography>
          <Button className={styles.closeBtn} onClick={modal.remove} variant='secondary'>
            <X />
          </Button>
        </div>

        <div className={styles.meta}>
          <div className={styles.metaItem}>
            <span className={styles.metaLabel}>Total Purchase</span>
            <span className={styles.metaValue}>${(expense.amount / 100).toFixed(2)}</span>
          </div>
          <div className={styles.metaItem}>
            <span className={styles.metaLabel}>Remaining</span>
            <span className={styles.metaValue}>${(remainingAmount / 100).toFixed(2)}</span>
          </div>
          <div className={styles.metaItem}>
            <span className={styles.metaLabel}>Period</span>
            <span className={styles.metaValue}>
              {format(expense.startDate, 'MMM d, yyyy')} - {format(expense.endDate, 'MMM d, yyyy')}
            </span>
          </div>
        </div>

        {expense.description && (
          <div className={styles.description}>
            <Typography size='sm' color='subtle'>
              {expense.description}
            </Typography>
          </div>
        )}

        <div className={styles.usageSection}>
          <Typography className={styles.sectionTitle} size='lg' weight='semibold'>
            Daily Breakdown & Recognition
          </Typography>
          <Typography className={styles.sectionSubtitle} size='sm' color='subtle'>
            Exempt specific days (skip recognition) or log custom usage amounts. Unlogged days auto-depreciate remaining balances linearly.
          </Typography>

          <div className={styles.daysList}>
            {days.map(d => {
              const dateStr = d.toISOString().split('T')[0];
              const dailyAmount = dailyMap.get(dateStr) || 0;
              
              const matchingLog = (expense.usageLogs || []).find(
                log => log.date.toISOString().split('T')[0] === dateStr
              );

              const isExempted = matchingLog && matchingLog.amount === 0;
              const isCustom = matchingLog && matchingLog.amount > 0;
              const isEditing = editingDate === dateStr;

              return (
                <div key={dateStr} className={`${styles.dayRow} ${isExempted ? styles.exemptedRow : ''}`}>
                  <div className={styles.dayInfo}>
                    <Calendar className={styles.dayIcon} size={16} />
                    <span className={styles.dayDate}>{format(d, 'EEE, MMM d')}</span>
                    {isExempted && <span className={styles.badgeExempt}>Exempted</span>}
                    {isCustom && <span className={styles.badgeCustom}>Custom Log</span>}
                  </div>

                  {isEditing ? (
                    <div className={styles.editForm}>
                      <div className={styles.inputRow}>
                        <span className={styles.dollarSign}>$</span>
                        <input
                          type='number'
                          step='0.01'
                          placeholder='0.00'
                          value={customAmount}
                          onChange={e => setCustomAmount(e.target.value)}
                          className={styles.amountInput}
                          autoFocus
                        />
                      </div>
                      <div className={styles.checkboxRow} onClick={() => setAutoCreateTxn(!autoCreateTxn)}>
                        {autoCreateTxn ? (
                          <CheckSquare className={styles.checkbox} size={16} />
                        ) : (
                          <Square className={styles.checkbox} size={16} />
                        )}
                        <span className={styles.checkboxLabel}>Create standard Monetr transaction</span>
                      </div>
                      <div className={styles.editActions}>
                        <Button
                          onClick={() => handleSaveCustomUsage(d, dateStr)}
                          variant='primary'
                          className={styles.miniBtn}
                        >
                          <Check size={14} />
                        </Button>
                        <Button
                          onClick={() => setEditingDate(null)}
                          variant='secondary'
                          className={styles.miniBtn}
                        >
                          <X size={14} />
                        </Button>
                      </div>
                    </div>
                  ) : (
                    <div className={styles.dayActions}>
                      <span className={styles.dayAmount}>
                        ${(dailyAmount / 100).toFixed(2)}
                      </span>

                      {matchingLog ? (
                        <Button
                          onClick={() => handleRemoveLog(matchingLog.accrualUsageLogId)}
                          variant='secondary'
                          className={styles.actionBtn}
                        >
                          Revert
                        </Button>
                      ) : (
                        <div className={styles.actionGroup}>
                          <Button
                            onClick={() => {
                              setEditingDate(dateStr);
                              setCustomAmount(amountToFriendly(dailyAmount).toString());
                            }}
                            variant='secondary'
                            className={styles.actionBtn}
                          >
                            Log Custom
                          </Button>
                          <Button
                            onClick={() => handleExemptDay(d)}
                            variant='secondary'
                            className={styles.actionBtn}
                          >
                            Skip
                          </Button>
                        </div>
                      )}
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        </div>

        <div className={styles.actions}>
          <Button onClick={handleDeleteExpense} variant='destructive' className={styles.deleteBtn}>
            <Trash2 size={16} />
            Delete Accrual
          </Button>
          <Button onClick={modal.remove} variant='secondary'>
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
