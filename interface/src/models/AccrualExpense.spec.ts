import AccrualExpense from '@monetr/interface/models/AccrualExpense';

describe('AccrualExpense', () => {
  it('will construct with date fields parsed', () => {
    const expense = new AccrualExpense({
      accrualExpenseId: 'ae_1',
      bankAccountId: 'ba_1',
      name: 'Annual Insurance',
      amount: 120000,
      startDate: '2026-01-01T00:00:00Z' as any,
      endDate: '2026-12-31T00:00:00Z' as any,
      createdAt: '2026-01-01T00:00:00Z' as any,
      updatedAt: '2026-01-01T00:00:00Z' as any,
    });
    expect(expense.startDate).toBeInstanceOf(Date);
    expect(expense.endDate).toBeInstanceOf(Date);
    expect(expense.createdAt).toBeInstanceOf(Date);
    expect(expense.name).toBe('Annual Insurance');
  });

  it('will construct with usage logs parsed', () => {
    const expense = new AccrualExpense({
      accrualExpenseId: 'ae_2',
      bankAccountId: 'ba_1',
      name: 'Subscription',
      amount: 3000,
      startDate: '2026-06-01T00:00:00Z' as any,
      endDate: '2026-06-03T00:00:00Z' as any,
      createdAt: '2026-06-01T00:00:00Z' as any,
      updatedAt: '2026-06-01T00:00:00Z' as any,
      usageLogs: [
        {
          accrualUsageLogId: 'ul_1',
          accrualExpenseId: 'ae_2',
          amount: 1000,
          date: '2026-06-01T00:00:00Z' as any,
          createdAt: '2026-06-01T00:00:00Z' as any,
          updatedAt: '2026-06-01T00:00:00Z' as any,
        },
      ],
    });
    expect(expense.usageLogs).toHaveLength(1);
    expect(expense.usageLogs![0].date).toBeInstanceOf(Date);
  });

  describe('getDailyUsageMap', () => {
    it('distributes amount linearly across days when no usage logs', () => {
      const expense = new AccrualExpense({
        accrualExpenseId: 'ae_3',
        bankAccountId: 'ba_1',
        name: 'Test',
        amount: 300,
        startDate: new Date('2026-06-01T00:00:00Z'),
        endDate: new Date('2026-06-03T00:00:00Z'),
        createdAt: new Date(),
        updatedAt: new Date(),
      });

      const usage = expense.getDailyUsageMap();
      // 3 days total (June 1, 2, 3), 300 / 3 = 100 per day
      expect(usage.size).toBe(3);
      expect(usage.get('2026-06-01')).toBe(100);
      expect(usage.get('2026-06-02')).toBe(100);
      expect(usage.get('2026-06-03')).toBe(100);
    });

    it('distributes remainder cents to early days', () => {
      const expense = new AccrualExpense({
        accrualExpenseId: 'ae_4',
        bankAccountId: 'ba_1',
        name: 'Test',
        amount: 100, // 100 cents across 3 days = 33 + 33 + 34
        startDate: new Date('2026-06-01T00:00:00Z'),
        endDate: new Date('2026-06-03T00:00:00Z'),
        createdAt: new Date(),
        updatedAt: new Date(),
      });

      const usage = expense.getDailyUsageMap();
      const values = Array.from(usage.values());
      // Total should equal 100
      expect(values.reduce((a, b) => a + b, 0)).toBe(100);
      // First day(s) get the remainder (+1 each)
      expect(usage.get('2026-06-01')).toBe(34);
      expect(usage.get('2026-06-02')).toBe(33);
      expect(usage.get('2026-06-03')).toBe(33);
    });

    it('respects usage logs and distributes remaining linearly', () => {
      const expense = new AccrualExpense({
        accrualExpenseId: 'ae_5',
        bankAccountId: 'ba_1',
        name: 'Test',
        amount: 300,
        startDate: new Date('2026-06-01T00:00:00Z'),
        endDate: new Date('2026-06-03T00:00:00Z'),
        createdAt: new Date(),
        updatedAt: new Date(),
        usageLogs: [
          {
            accrualUsageLogId: 'ul_1',
            accrualExpenseId: 'ae_5',
            amount: 150,
            date: new Date('2026-06-01T00:00:00Z'),
            createdAt: new Date(),
            updatedAt: new Date(),
          },
        ],
      });

      const usage = expense.getDailyUsageMap();
      // Day 1: 150 (from log)
      // Remaining: 300 - 150 = 150 across 2 unlogged days = 75 each
      expect(usage.get('2026-06-01')).toBe(150);
      expect(usage.get('2026-06-02')).toBe(75);
      expect(usage.get('2026-06-03')).toBe(75);
    });

    it('handles early depletion (all amount consumed by logs)', () => {
      const expense = new AccrualExpense({
        accrualExpenseId: 'ae_6',
        bankAccountId: 'ba_1',
        name: 'Test',
        amount: 200,
        startDate: new Date('2026-06-01T00:00:00Z'),
        endDate: new Date('2026-06-03T00:00:00Z'),
        createdAt: new Date(),
        updatedAt: new Date(),
        usageLogs: [
          {
            accrualUsageLogId: 'ul_1',
            accrualExpenseId: 'ae_6',
            amount: 200,
            date: new Date('2026-06-01T00:00:00Z'),
            createdAt: new Date(),
            updatedAt: new Date(),
          },
        ],
      });

      const usage = expense.getDailyUsageMap();
      expect(usage.get('2026-06-01')).toBe(200);
      expect(usage.get('2026-06-02')).toBe(0);
      expect(usage.get('2026-06-03')).toBe(0);
    });

    it('returns empty map for zero or negative duration', () => {
      const expense = new AccrualExpense({
        accrualExpenseId: 'ae_7',
        bankAccountId: 'ba_1',
        name: 'Test',
        amount: 100,
        startDate: new Date('2026-06-03T00:00:00Z'),
        endDate: new Date('2026-06-01T00:00:00Z'), // end before start
        createdAt: new Date(),
        updatedAt: new Date(),
      });

      const usage = expense.getDailyUsageMap();
      expect(usage.size).toBe(0);
    });
  });
});
