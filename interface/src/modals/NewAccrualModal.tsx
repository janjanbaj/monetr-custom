import { useRef } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { startOfDay, startOfTomorrow, addMonths } from 'date-fns';
import { type FormikHelpers } from 'formik';

import type { ApiError } from '@monetr/interface/api/client';
import { Button } from '@monetr/interface/components/Button';
import FormAmountField from '@monetr/interface/components/FormAmountField';
import FormButton from '@monetr/interface/components/FormButton';
import FormDatePicker from '@monetr/interface/components/FormDatePicker';
import FormTextField from '@monetr/interface/components/FormTextField';
import MForm from '@monetr/interface/components/MForm';
import MModal, { type MModalRef } from '@monetr/interface/components/MModal';
import Typography from '@monetr/interface/components/Typography';
import { useCreateAccrualExpense } from '@monetr/interface/hooks/useCreateAccrualExpense';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useSelectedBankAccount } from '@monetr/interface/hooks/useSelectedBankAccount';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import AccrualExpense from '@monetr/interface/models/AccrualExpense';
import type { APIError } from '@monetr/interface/util/request';
import type { ExtractProps } from '@monetr/interface/util/typescriptEvils';
import { useSnackbar } from '@monetr/notify';

import styles from './NewAccrualModal.module.scss';

interface NewAccrualValues {
  name: string;
  description: string;
  amount: number;
  startDate: Date;
  endDate: Date;
}

function NewAccrualModal(): JSX.Element {
  const { inTimezone } = useTimezone();
  const {
    data: { friendlyToAmount },
  } = useLocaleCurrency();
  const modal = useModal();
  const { enqueueSnackbar } = useSnackbar();
  const { data: selectedBankAccount } = useSelectedBankAccount();
  const createAccrualExpense = useCreateAccrualExpense();

  const ref = useRef<MModalRef>(null);

  if (!selectedBankAccount) {
    return (
      <MModal className={styles.modal} open={modal.visible} ref={ref}>
        One moment...
      </MModal>
    );
  }

  const initialValues: NewAccrualValues = {
    name: '',
    description: '',
    amount: 0.0,
    startDate: startOfDay(new Date(), {
      in: inTimezone,
    }),
    endDate: startOfDay(addMonths(new Date(), 1), {
      in: inTimezone,
    }),
  };

  async function submit(values: NewAccrualValues, helper: FormikHelpers<NewAccrualValues>): Promise<void> {
    const newAccrual = new AccrualExpense({
      bankAccountId: selectedBankAccount.bankAccountId,
      name: values.name.trim(),
      description: values.description.trim() || undefined,
      amount: friendlyToAmount(values.amount),
      startDate: startOfDay(new Date(values.startDate), {
        in: inTimezone,
      }),
      endDate: startOfDay(new Date(values.endDate), {
        in: inTimezone,
      }),
    });

    helper.setSubmitting(true);
    return createAccrualExpense(newAccrual)
      .then(created => modal.resolve(created))
      .then(() => modal.remove())
      .catch(
        (error: ApiError<APIError>) =>
          void enqueueSnackbar(error.response.data.error, {
            variant: 'error',
            disableWindowBlurListener: true,
          }),
      )
      .finally(() => helper.setSubmitting(false));
  }

  return (
    <MModal className={styles.modal} open={modal.visible} ref={ref}>
      <MForm className={styles.form} data-testid='new-accrual-modal' initialValues={initialValues} onSubmit={submit}>
        <div className={styles.body}>
          <Typography className={styles.heading} size='xl' weight='bold'>
            Create A New Accrual
          </Typography>
          <FormTextField
            autoComplete='off'
            autoFocus
            data-1p-ignore
            label='What are you tracking?'
            name='name'
            placeholder='Rice, Subscription, Toiletries...'
            required
          />
          <FormTextField
            autoComplete='off'
            data-1p-ignore
            label='Description (optional)'
            name='description'
            placeholder='A brief note about this accrual...'
          />
          <FormAmountField
            allowNegative={false}
            label='Total amount'
            name='amount'
            required
          />
          <div className={styles.fieldRow}>
            <FormDatePicker
              className={styles.fieldRowItem}
              label='Start date'
              name='startDate'
              required
            />
            <FormDatePicker
              className={styles.fieldRowItem}
              label='End date'
              min={startOfTomorrow({
                in: inTimezone,
              })}
              name='endDate'
              required
            />
          </div>
        </div>
        <div className={styles.actions}>
          <Button data-testid='close-new-accrual-modal' onClick={modal.remove} variant='secondary'>
            Cancel
          </Button>
          <FormButton type='submit' variant='primary'>
            Create
          </FormButton>
        </div>
      </MForm>
    </MModal>
  );
}

const newAccrualModal = NiceModal.create(NewAccrualModal);

export default newAccrualModal;

export function showNewAccrualModal(): Promise<AccrualExpense | null> {
  return NiceModal.show<AccrualExpense | null, ExtractProps<typeof newAccrualModal>, unknown>(newAccrualModal);
}
