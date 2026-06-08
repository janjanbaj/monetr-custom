import parseDate from '@monetr/interface/util/parseDate';
import type BankAccount from './BankAccount';

export type FixedDepositStatus = 'active' | 'matured' | 'withdrawn';

export default class FixedDeposit {
  fixedDepositId: string;
  sourceBankAccountId: string;
  fixedBankAccountId: string;
  fundingScheduleId?: string;
  name: string;
  amount: number;
  interestRate: number;
  startDate: Date;
  endDate: Date;
  interestFrequency: 'monthly' | 'quarterly' | 'end_of_term';
  interestDestination: 'accumulate' | 'payout';
  interestDestinationAccountId?: string;
  status: FixedDepositStatus;
  createdAt: Date;
  updatedAt: Date;

  // Relations (joined)
  fixedBankAccount?: BankAccount;

  constructor(data?: Partial<FixedDeposit>) {
    if (data) {
      Object.assign(this, {
        ...data,
        startDate: parseDate(data.startDate) || new Date(),
        endDate: parseDate(data.endDate) || new Date(),
        createdAt: parseDate(data.createdAt),
        updatedAt: parseDate(data.updatedAt),
      });
    }
  }
}
