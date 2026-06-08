import { Fragment } from 'react';
import { Lock, PlusCircle, Coins, ArrowRightLeft, Landmark, BarChart2, TrendingUp } from 'lucide-react';
import { useSnackbar } from '@monetr/notify';

import { Button } from '@monetr/interface/components/Button';
import Flex from '@monetr/interface/components/Flex';
import Typography from '@monetr/interface/components/Typography';
import Badge from '@monetr/interface/components/Badge';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import { useSelectedBankAccount } from '@monetr/interface/hooks/useSelectedBankAccount';
import { useFixedDeposits, useWithdrawFixedDeposit } from '@monetr/interface/hooks/useFixedDeposits';
import { useCurrentBalance } from '@monetr/interface/hooks/useCurrentBalance';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { AmountType } from '@monetr/interface/util/amounts';
import FixedDeposit from '@monetr/interface/models/FixedDeposit';
import { showNewFixedDepositModal } from '@monetr/interface/modals/NewFixedDepositModal';

import styles from './fixed_deposits.module.scss';

export default function FixedDepositsPage(): JSX.Element {
  const { data: bankAccount } = useSelectedBankAccount();
  const { data: deposits, isLoading } = useFixedDeposits(bankAccount?.bankAccountId || '');
  const { data: balance } = useCurrentBalance();
  const withdrawDeposit = useWithdrawFixedDeposit(bankAccount?.bankAccountId || '');
  const { data: locale } = useLocaleCurrency();
  const { enqueueSnackbar } = useSnackbar();

  if (isLoading || !bankAccount) {
    return (
      <div style={{ padding: '2rem', textAlign: 'center' }}>
        <Typography size='xl'>Loading fixed deposits...</Typography>
      </div>
    );
  }

  const activeDeposits = deposits?.filter(d => d.status === 'active') || [];
  const historicalDeposits = deposits?.filter(d => d.status !== 'active') || [];

  const handleWithdraw = async (depositId: string, name: string) => {
    if (
      confirm(
        `Are you sure you want to withdraw "${name}" early? The remaining interest schedule will be cancelled, and all funds in the deposit will be returned. Log any penalty manually.`,
      )
    ) {
      try {
        await withdrawDeposit.mutateAsync(depositId);
        enqueueSnackbar('Fixed deposit successfully withdrawn.', { variant: 'success' });
      } catch (err: any) {
        enqueueSnackbar(err?.response?.data?.error || 'Failed to withdraw fixed deposit', { variant: 'error' });
      }
    }
  };

  const getTermMonths = (start: Date, end: Date) => {
    return (end.getFullYear() - start.getFullYear()) * 12 + (end.getMonth() - start.getMonth());
  };

  const getExpectedInterest = (d: FixedDeposit) => {
    const termMonths = getTermMonths(d.startDate, d.endDate);
    let expected = 0;
    if (d.interestRate > 0) {
      if (d.interestFrequency === 'monthly') {
        const single = Math.round(d.amount * (d.interestRate / 100.0) / 12.0);
        expected = single * termMonths;
      } else if (d.interestFrequency === 'quarterly') {
        const single = Math.round(d.amount * (d.interestRate / 100.0) / 4.0);
        const count = Math.max(1, Math.floor(termMonths / 3));
        expected = single * count;
      } else if (d.interestFrequency === 'end_of_term') {
        expected = Math.round(d.amount * (d.interestRate / 100.0) * (termMonths / 12.0));
      }
    }
    return expected;
  };

  // Calculations for Analytics
  const freeBalance = balance?.free ?? bankAccount.availableBalance;
  const activeDepositsTotal = activeDeposits.reduce((acc, d) => acc + d.amount, 0);
  const combinedNetWorth = freeBalance + activeDepositsTotal;
  const totalExpectedInterest = activeDeposits.reduce((acc, d) => acc + getExpectedInterest(d), 0);
  const projectedWorth = combinedNetWorth + totalExpectedInterest;

  return (
    <Fragment>
      <MTopNavigation icon={Lock} title='Fixed Deposits & CDs'>
        <Button onClick={showNewFixedDepositModal} variant='primary'>
          <PlusCircle className={styles.btnIcon} /> Open New Deposit
        </Button>
      </MTopNavigation>

      <div className={styles.root}>
        {/* Analytics Section */}
        <div className={styles.analyticsGrid}>
          <div className={styles.analyticsCard}>
            <Coins className={styles.analyticsCardIcon} />
            <span className={styles.cardTitle}>Free-To-Use Balance</span>
            <span className={styles.cardValue}>{locale.formatAmount(freeBalance, AmountType.Stored)}</span>
          </div>
          <div className={styles.analyticsCard}>
            <Landmark className={styles.analyticsCardIcon} />
            <span className={styles.cardTitle}>Fixed Deposits ({activeDeposits.length} Active)</span>
            <span className={styles.cardValue}>{locale.formatAmount(activeDepositsTotal, AmountType.Stored)}</span>
          </div>
          <div className={styles.analyticsCard}>
            <BarChart2 className={styles.analyticsCardIcon} />
            <span className={styles.cardTitle}>Combined Net Worth</span>
            <span className={styles.cardValue}>{locale.formatAmount(combinedNetWorth, AmountType.Stored)}</span>
          </div>
          <div className={styles.analyticsCard}>
            <TrendingUp className={styles.analyticsCardIcon} />
            <span className={styles.cardTitle}>Projected Worth (w/ Interest)</span>
            <span className={styles.cardValue}>{locale.formatAmount(projectedWorth, AmountType.Stored)}</span>
          </div>
        </div>

        {/* Active Deposits */}
        <div className={styles.section}>
          <Typography className={styles.sectionTitle} size='lg' weight='semibold'>
            Active Deposits
          </Typography>
          {activeDeposits.length === 0 ? (
            <div className={styles.emptyState}>
              <Coins className={styles.emptyIcon} />
              <Typography color='subtle'>No active Fixed Deposits or CDs found.</Typography>
            </div>
          ) : (
            <div className={styles.grid}>
              {activeDeposits.map(d => {
                const totalDays = Math.ceil((d.endDate.getTime() - d.startDate.getTime()) / (1000 * 60 * 60 * 24));
                const elapsedDays = Math.max(
                  0,
                  Math.ceil((new Date().getTime() - d.startDate.getTime()) / (1000 * 60 * 60 * 24)),
                );
                const progressPct = Math.min(100, Math.round((elapsedDays / totalDays) * 100));
                const expectedInterest = getExpectedInterest(d);

                return (
                  <div className={styles.card} key={d.fixedDepositId}>
                    <Flex align='center' justify='between'>
                      <Typography size='lg' weight='semibold'>
                        {d.name}
                      </Typography>
                      <Badge variant='success'>Active</Badge>
                    </Flex>

                    <div className={styles.amountText}>{locale.formatAmount(d.amount, AmountType.Stored)}</div>

                    <div className={styles.metaRow}>
                      <div>
                        <Typography color='subtle' size='sm'>
                          Interest Rate
                        </Typography>
                        <Typography weight='medium'>
                          {d.interestRate}% ({d.interestFrequency})
                        </Typography>
                      </div>
                      <div>
                        <Typography color='subtle' size='sm'>
                          Expected Interest
                        </Typography>
                        <Typography weight='medium' color='emphasis'>
                          +{locale.formatAmount(expectedInterest, AmountType.Stored)}
                        </Typography>
                      </div>
                    </div>

                    <div className={styles.metaRow}>
                      <div>
                        <Typography color='subtle' size='sm'>
                          Term Length
                        </Typography>
                        <Typography weight='medium'>{getTermMonths(d.startDate, d.endDate)} Months</Typography>
                      </div>
                      <div>
                        <Typography color='subtle' size='sm'>
                          Maturity Date
                        </Typography>
                        <Typography weight='medium'>{d.endDate.toLocaleDateString()}</Typography>
                      </div>
                    </div>

                    {/* Progress Bar */}
                    <div className={styles.progressContainer}>
                      <Flex justify='between' className={styles.progressLabel}>
                        <Typography size='sm' color='subtle'>
                          {progressPct}% Completed
                        </Typography>
                        <Typography size='sm' color='subtle'>
                          {elapsedDays} / {totalDays} days
                        </Typography>
                      </Flex>
                      <div className={styles.progressBar}>
                        <div className={styles.progressFill} style={{ width: `${progressPct}%` }} />
                      </div>
                    </div>

                    <Button
                      className={styles.withdrawBtn}
                      onClick={() => handleWithdraw(d.fixedDepositId, d.name)}
                      variant='destructive'
                    >
                      Withdraw Early / Cancel
                    </Button>
                  </div>
                );
              })}
            </div>
          )}
        </div>

        {/* History */}
        {historicalDeposits.length > 0 && (
          <div className={styles.section}>
            <Typography className={styles.sectionTitle} size='lg' weight='semibold'>
              History
            </Typography>
            <div className={styles.historyList}>
              {historicalDeposits.map(d => (
                <div className={styles.historyRow} key={d.fixedDepositId}>
                  <Flex align='center' gap='md'>
                    <ArrowRightLeft className={styles.historyIcon} />
                    <div>
                      <Typography weight='semibold'>{d.name}</Typography>
                      <Typography color='subtle' size='sm'>
                        Term: {d.startDate.toLocaleDateString()} - {d.endDate.toLocaleDateString()} ({getTermMonths(d.startDate, d.endDate)}m, {d.interestRate}%)
                      </Typography>
                    </div>
                  </Flex>
                  <Flex align='center' gap='md'>
                    <Typography weight='semibold'>{locale.formatAmount(d.amount, AmountType.Stored)}</Typography>
                    <Badge variant={d.status === 'matured' ? 'info' : 'brand'}>{d.status.toUpperCase()}</Badge>
                  </Flex>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </Fragment>
  );
}
