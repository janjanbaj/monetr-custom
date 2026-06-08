# Phase 5: User Interface and Modals

This phase covers building the user interface to manage, track, and configure Fixed Deposits.

## Target Files
1. [BudgetingSidebar.tsx](file:///Users/janeetbajracharya/Desktop/Code/monetr/interface/src/components/Layout/BudgetingSidebar.tsx) [MODIFY]
2. [monetr.tsx](file:///Users/janeetbajracharya/Desktop/Code/monetr/interface/src/monetr.tsx) [MODIFY]
3. [fixed_deposits.tsx](file:///Users/janeetbajracharya/Desktop/Code/monetr/interface/src/pages/fixed_deposits.tsx) [NEW]
4. [fixed_deposits.module.scss](file:///Users/janeetbajracharya/Desktop/Code/monetr/interface/src/pages/fixed_deposits.module.scss) [NEW]
5. [NewFixedDepositModal.tsx](file:///Users/janeetbajracharya/Desktop/Code/monetr/interface/src/modals/NewFixedDepositModal.tsx) [NEW]
6. [NewFixedDepositModal.module.scss](file:///Users/janeetbajracharya/Desktop/Code/monetr/interface/src/modals/NewFixedDepositModal.module.scss) [NEW]

---

## 1. Sidebar Link
In `interface/src/components/Layout/BudgetingSidebar.tsx`, add the NavigationItem in `navList` using the Lucide `Lock` icon:
```typescript
          <NavigationItem to={`/bank/${bankAccount?.bankAccountId}/fixed_deposits`}>
            <Lock />
            <Typography color='inherit' ellipsis size='lg' weight='medium'>
              Fixed Deposits
            </Typography>
          </NavigationItem>
```

---

## 2. Main Page Layout
Create `interface/src/pages/fixed_deposits.tsx`. This page displays:
- Active fixed deposits cards with nice progress bars showing maturity percentage.
- Interest rates, expected maturity date, and frequency details.
- "Withdraw Early" button which opens a confirmation alert, cancels future schedules, and returns funds.
- List of completed/matured deposits in a secondary list.

```typescript
import { Fragment } from 'react';
import { Lock, PlusCircle, Calendar, Coins, ArrowRightLeft } from 'lucide-react';
import { useSnackbar } from '@monetr/notify';

import { Button } from '@monetr/interface/components/Button';
import Flex from '@monetr/interface/components/Flex';
import { layoutVariants } from '@monetr/interface/components/Layout';
import Typography from '@monetr/interface/components/Typography';
import Badge from '@monetr/interface/components/Badge';
import { useSelectedBankAccount } from '@monetr/interface/hooks/useSelectedBankAccount';
import { useFixedDeposits, useWithdrawFixedDeposit } from '@monetr/interface/hooks/useFixedDeposits';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { AmountType } from '@monetr/interface/util/amounts';
import { showNewFixedDepositModal } from '@monetr/interface/modals/NewFixedDepositModal';

import styles from './fixed_deposits.module.scss';

export default function FixedDepositsPage(): JSX.Element {
  const { data: bankAccount } = useSelectedBankAccount();
  const { data: deposits, isLoading } = useFixedDeposits(bankAccount?.bankAccountId || '');
  const withdrawDeposit = useWithdrawFixedDeposit(bankAccount?.bankAccountId || '');
  const { data: locale } = useLocaleCurrency();
  const { enqueueSnackbar } = useSnackbar();

  if (isLoading || !bankAccount) {
    return <div>Loading fixed deposits...</div>;
  }

  const activeDeposits = deposits?.filter(d => d.status === 'active') || [];
  const historicalDeposits = deposits?.filter(d => d.status !== 'active') || [];

  const handleWithdraw = async (depositId: string, name: string) => {
    if (confirm(`Are you sure you want to withdraw "${name}" early? The remaining interest schedule will be cancelled, and all funds in the deposit will be returned. Log any penalty manually.`)) {
      try {
        await withdrawDeposit.mutateAsync(depositId);
        enqueueSnackbar('Fixed deposit successfully withdrawn.', { variant: 'success' });
      } catch (err: any) {
        enqueueSnackbar(err?.response?.data?.error || 'Failed to withdraw fixed deposit', { variant: 'error' });
      }
    }
  };

  return (
    <div className={styles.root}>
      <Flex align='center' justify='space-between'>
        <Flex align='center' gap='sm'>
          <Lock className={styles.headerIcon} />
          <Typography size='xxl' weight='bold'>Fixed Deposits & CDs</Typography>
        </Flex>
        <Button onClick={() => showNewFixedDepositModal()} variant='primary'>
          <PlusCircle className={styles.btnIcon} /> Open New Deposit
        </Button>
      </Flex>

      {/* Active Deposits */}
      <div className={styles.section}>
        <Typography className={styles.sectionTitle} size='lg' weight='semibold'>Active Deposits</Typography>
        {activeDeposits.length === 0 ? (
          <div className={styles.emptyState}>
            <Coins className={styles.emptyIcon} />
            <Typography color='secondary'>No active Fixed Deposits or CDs found.</Typography>
          </div>
        ) : (
          <div className={styles.grid}>
            {activeDeposits.map(d => {
              const totalDays = Math.ceil((d.endDate.getTime() - d.startDate.getTime()) / (1000 * 60 * 60 * 24));
              const elapsedDays = Math.max(0, Math.ceil((new Date().getTime() - d.startDate.getTime()) / (1000 * 60 * 60 * 24)));
              const progressPct = Math.min(100, Math.round((elapsedDays / totalDays) * 100));

              return (
                <div className={styles.card} key={d.fixedDepositId}>
                  <Flex align='center' justify='space-between'>
                    <Typography size='lg' weight='semibold'>{d.name}</Typography>
                    <Badge variant='success'>Active</Badge>
                  </Flex>

                  <div className={styles.amountText}>
                    {locale.formatAmount(d.amount, AmountType.Stored)}
                  </div>

                  <Flex gap='md' className={styles.metaRow}>
                    <div>
                      <Typography color='secondary' size='sm'>Interest Rate</Typography>
                      <Typography weight='medium'>{d.interestRate}% ({d.interestFrequency})</Typography>
                    </div>
                    <div>
                      <Typography color='secondary' size='sm'>Maturity Date</Typography>
                      <Typography weight='medium'>{d.endDate.toLocaleDateString()}</Typography>
                    </div>
                  </Flex>

                  {/* Progress Bar */}
                  <div className={styles.progressContainer}>
                    <Flex justify='space-between' className={styles.progressLabel}>
                      <Typography size='sm' color='secondary'>{progressPct}% Completed</Typography>
                      <Typography size='sm' color='secondary'>{elapsedDays} / {totalDays} days</Typography>
                    </Flex>
                    <div className={styles.progressBar}>
                      <div className={styles.progressFill} style={{ width: `${progressPct}%` }} />
                    </div>
                  </div>

                  <Button
                    className={styles.withdrawBtn}
                    onClick={() => handleWithdraw(d.fixedDepositId, d.name)}
                    variant='secondary'
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
          <Typography className={styles.sectionTitle} size='lg' weight='semibold'>History</Typography>
          <div className={styles.historyList}>
            {historicalDeposits.map(d => (
              <Flex align='center' justify='space-between' className={styles.historyRow} key={d.fixedDepositId}>
                <Flex align='center' gap='md'>
                  <ArrowRightLeft className={styles.historyIcon} />
                  <div>
                    <Typography weight='semibold'>{d.name}</Typography>
                    <Typography color='secondary' size='sm'>
                      Term: {d.startDate.toLocaleDateString()} - {d.endDate.toLocaleDateString()}
                    </Typography>
                  </div>
                </Flex>
                <Flex align='center' gap='md'>
                  <Typography weight='semibold'>{locale.formatAmount(d.amount, AmountType.Stored)}</Typography>
                  <Badge variant={d.status === 'matured' ? 'info' : 'secondary'}>
                    {d.status.toUpperCase()}
                  </Badge>
                </Flex>
              </Flex>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
```

---

## 3. New Fixed Deposit Modal
Create `interface/src/modals/NewFixedDepositModal.tsx`. A modal form allows the user to specify FD details and select checking/savings accounts to fund it.

```typescript
import { Fragment, useRef } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import type { FormikHelpers } from 'formik';
import { useSnackbar } from '@monetr/notify';

import { Button } from '@monetr/interface/components/Button';
import Flex from '@monetr/interface/components/Flex';
import FormAmountField from '@monetr/interface/components/FormAmountField';
import FormButton from '@monetr/interface/components/FormButton';
import FormTextField from '@monetr/interface/components/FormTextField';
import MForm from '@monetr/interface/components/MForm';
import MModal, { type MModalRef } from '@monetr/interface/components/MModal';
import Typography from '@monetr/interface/components/Typography';
import { useSelectedBankAccount } from '@monetr/interface/hooks/useSelectedBankAccount';
import { useBankAccounts } from '@monetr/interface/hooks/useBankAccounts';
import { useCreateFixedDeposit } from '@monetr/interface/hooks/useFixedDeposits';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';

import styles from './NewFixedDepositModal.module.scss';

interface FormValues {
  name: string;
  amount: number;
  interestRate: number;
  startDate: string;
  termMonths: number;
  interestFrequency: 'monthly' | 'quarterly' | 'end_of_term';
  interestDestination: 'accumulate' | 'payout';
  interestDestinationAccountId: string;
}

function NewFixedDepositModal(): JSX.Element {
  const modal = useModal();
  const { enqueueSnackbar } = useSnackbar();
  const { data: locale } = useLocaleCurrency();
  const { data: bankAccount } = useSelectedBankAccount();
  const { data: bankAccounts } = useBankAccounts();
  const createDeposit = useCreateFixedDeposit(bankAccount?.bankAccountId || '');
  const ref = useRef<MModalRef>(null);

  if (!bankAccount) return null;

  const initialValues: FormValues = {
    name: '',
    amount: 1000,
    interestRate: 5.0,
    startDate: new Date().toISOString().split('T')[0],
    termMonths: 12,
    interestFrequency: 'monthly',
    interestDestination: 'accumulate',
    interestDestinationAccountId: bankAccount.bankAccountId,
  };

  const submit = async (values: FormValues, helper: FormikHelpers<FormValues>): Promise<void> => {
    helper.setSubmitting(true);
    const startIso = new Date(values.startDate).toISOString();
    return await createDeposit.mutateAsync({
      name: values.name,
      amount: locale.friendlyToAmount(values.amount),
      interestRate: Number(values.interestRate),
      startDate: startIso,
      termMonths: Number(values.termMonths),
      interestFrequency: values.interestFrequency,
      interestDestination: values.interestDestination,
      interestDestinationAccountId: values.interestDestination === 'payout' ? values.interestDestinationAccountId : undefined,
    })
      .then(() => {
        enqueueSnackbar('Fixed deposit successfully created.', { variant: 'success' });
        modal.remove();
      })
      .catch((error: any) => {
        enqueueSnackbar(error?.response?.data?.error || 'Failed to create fixed deposit.', { variant: 'error' });
      })
      .finally(() => helper.setSubmitting(false));
  };

  return (
    <MModal open={modal.visible} ref={ref}>
      <MForm className={styles.form} onSubmit={submit} initialValues={initialValues}>
        {({ isSubmitting, values, setFieldValue }) => (
          <Fragment>
            <Typography size='xl' weight='bold' className={styles.heading}>
              Open New Fixed Deposit / CD
            </Typography>

            <Flex orientation='column' gap='md'>
              <FormTextField label='Deposit Name' name='name' placeholder='e.g., 1 Year CD at Wells Fargo' required />

              <FormAmountField label='Principal Amount' name='amount' currency={bankAccount.currency} required />

              <FormTextField label='Annual Interest Rate (%)' name='interestRate' type='number' required />

              <FormTextField label='Start Date' name='startDate' type='date' required />

              <div>
                <label className={styles.fieldLabel}>Term Duration (Months)</label>
                <select
                  name='termMonths'
                  value={values.termMonths}
                  onChange={e => void setFieldValue('termMonths', Number(e.target.value))}
                  className={styles.select}
                >
                  <option value={3}>3 Months</option>
                  <option value={6}>6 Months</option>
                  <option value={12}>12 Months (1 Year)</option>
                  <option value={24}>24 Months (2 Years)</option>
                  <option value={36}>36 Months (3 Years)</option>
                  <option value={60}>60 Months (5 Years)</option>
                </select>
              </div>

              <div>
                <label className={styles.fieldLabel}>Interest Payout Frequency</label>
                <select
                  name='interestFrequency'
                  value={values.interestFrequency}
                  onChange={e => void setFieldValue('interestFrequency', e.target.value)}
                  className={styles.select}
                >
                  <option value='monthly'>Monthly</option>
                  <option value='quarterly'>Quarterly</option>
                  <option value='end_of_term'>End of Term</option>
                </select>
              </div>

              <div>
                <label className={styles.fieldLabel}>Interest Destination</label>
                <select
                  name='interestDestination'
                  value={values.interestDestination}
                  onChange={e => void setFieldValue('interestDestination', e.target.value)}
                  className={styles.select}
                >
                  <option value='accumulate'>Accumulate in Deposit Account</option>
                  <option value='payout'>Pay Out to External Account</option>
                </select>
              </div>

              {values.interestDestination === 'payout' && (
                <div>
                  <label className={styles.fieldLabel}>Destination Payout Account</label>
                  <select
                    name='interestDestinationAccountId'
                    value={values.interestDestinationAccountId}
                    onChange={e => void setFieldValue('interestDestinationAccountId', e.target.value)}
                    className={styles.select}
                  >
                    {bankAccounts
                      ?.filter(acc => acc.bankAccountId !== bankAccount.bankAccountId)
                      .map(acc => (
                        <option value={acc.bankAccountId} key={acc.bankAccountId}>
                          {acc.name}
                        </option>
                      ))}
                  </select>
                </div>
              )}
            </Flex>

            <Flex gap='md' justify='end' className={styles.btnRow}>
              <Button disabled={isSubmitting} onClick={modal.remove} variant='secondary'>
                Cancel
              </Button>
              <FormButton type='submit' variant='primary' disabled={isSubmitting}>
                Create Deposit
              </FormButton>
            </Flex>
          </Fragment>
        )}
      </MForm>
    </MModal>
  );
}

const newFixedDepositModal = NiceModal.create(NewFixedDepositModal);
export default newFixedDepositModal;

export function showNewFixedDepositModal(): Promise<void> {
  return NiceModal.show<void, any, unknown>(newFixedDepositModal);
}
```

## 4. Analytics Page:
Show Combined networth for a selected bank account. Show the total number of fixed deposits that are linked to this bank account. 

Show the free balance and the total balance which is the sum of the free balance and fixed deposit. 

Using the fixesd deposit interest payment cycle calculate the total expected interest. Show the projected worth of the entire account based on the accured interest.
---

## Verification Plan
- Deploy code locally, open browser subagent, create manual bank accounts, fund a fixed deposit, and verify UI components load with correct calculations.
