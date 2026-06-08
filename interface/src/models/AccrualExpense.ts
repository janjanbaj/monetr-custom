import parseDate from '@monetr/interface/util/parseDate';

export interface AccrualUsageLog {
  accrualUsageLogId: string;
  accrualExpenseId: string;
  amount: number;
  date: Date;
  createdAt: Date;
  updatedAt: Date;
}

export default class AccrualExpense {
  accrualExpenseId: string;
  bankAccountId: string;
  transactionId?: string;
  name: string;
  description?: string;
  amount: number;
  startDate: Date;
  endDate: Date;
  createdAt: Date;
  updatedAt: Date;
  usageLogs?: AccrualUsageLog[];

  constructor(data?: Partial<AccrualExpense>) {
    if (data) {
      Object.assign(this, {
        ...data,
        startDate: parseDate(data.startDate) || new Date(),
        endDate: parseDate(data.endDate) || new Date(),
        createdAt: parseDate(data.createdAt),
        updatedAt: parseDate(data.updatedAt),
        usageLogs: data.usageLogs?.map(log => ({
          ...log,
          date: parseDate(log.date) || new Date(),
          createdAt: parseDate(log.createdAt) || new Date(),
          updatedAt: parseDate(log.updatedAt) || new Date(),
        })) || [],
      });
    }
  }

  /**
   * Calculates daily recognized consumption.
   * Leverages custom Usage Logs if present; defaults to linear depreciation.
   */
  getDailyUsageMap(): Map<string, number> {
    const usage = new Map<string, number>();
    const totalDays = Math.ceil((this.endDate.getTime() - this.startDate.getTime()) / (1000 * 60 * 60 * 24)) + 1;
    if (totalDays <= 0) return usage;

    // 1. Identify logged dates and total manual amount
    const logs = this.usageLogs || [];
    let loggedAmount = 0;
    const loggedDates = new Set<string>();

    for (const log of logs) {
      const dateKey = log.date.toISOString().split('T')[0];
      loggedDates.add(dateKey);
      loggedAmount += log.amount;
      usage.set(dateKey, (usage.get(dateKey) || 0) + log.amount);
    }

    // 2. Linear depreciation for remaining days
    const remainingAmount = Math.max(0, this.amount - loggedAmount);
    const unloggedDaysCount = totalDays - loggedDates.size;

    if (remainingAmount > 0 && unloggedDaysCount > 0) {
      const dailyLinearAmount = Math.floor(remainingAmount / unloggedDaysCount);
      let remainder = remainingAmount % unloggedDaysCount;

      for (let i = 0; i < totalDays; i++) {
        const d = new Date(this.startDate);
        d.setDate(d.getDate() + i);
        const dateKey = d.toISOString().split('T')[0];

        if (!loggedDates.has(dateKey)) {
          let dayValue = dailyLinearAmount;
          if (remainder > 0) {
            dayValue += 1;
            remainder -= 1;
          }
          usage.set(dateKey, dayValue);
        }
      }
    } else if (remainingAmount === 0) {
      // If depleted early by manual logs, remaining days get 0
      for (let i = 0; i < totalDays; i++) {
        const d = new Date(this.startDate);
        d.setDate(d.getDate() + i);
        const dateKey = d.toISOString().split('T')[0];
        if (!usage.has(dateKey)) {
          usage.set(dateKey, 0);
        }
      }
    }

    return usage;
  }
}
