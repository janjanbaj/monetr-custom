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
