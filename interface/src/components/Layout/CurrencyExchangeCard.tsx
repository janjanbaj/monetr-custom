import React from 'react';
import { useQuery } from '@tanstack/react-query';

import Select, { type SelectOption } from '@monetr/interface/components/Select';
import Typography from '@monetr/interface/components/Typography';
import { useDefaultConversionCurrency } from '@monetr/interface/hooks/useDefaultConversionCurrency';
import { useInstalledCurrencies } from '@monetr/interface/hooks/useInstalledCurrencies';
import request from '@monetr/interface/util/request';

import styles from './CurrencyExchangeCard.module.scss';

interface ExchangeRateResponse {
  from: string;
  to: string;
  rate: number;
}

export default function CurrencyExchangeCard(): React.JSX.Element {
  const { data: currencies, isLoading: currenciesLoading } = useInstalledCurrencies();
  const [defaultCurrency, setDefaultCurrency] = useDefaultConversionCurrency();

  // Fetch rate from defaultCurrency -> USD
  const { data: rateToUsd, isLoading: loadingToUsd } = useQuery<ExchangeRateResponse>({
    queryKey: ['/api/exchange/rate', defaultCurrency, 'USD'],
    queryFn: async () => {
      const resp = await request<ExchangeRateResponse>({
        method: 'GET',
        url: '/api/exchange/rate',
        params: { from: defaultCurrency, to: 'USD' },
      });
      return resp.data;
    },
    enabled: defaultCurrency !== 'USD',
    staleTime: 5 * 60 * 1000,
  });

  // Fetch rate from USD -> defaultCurrency
  const { data: rateFromUsd, isLoading: loadingFromUsd } = useQuery<ExchangeRateResponse>({
    queryKey: ['/api/exchange/rate', 'USD', defaultCurrency],
    queryFn: async () => {
      const resp = await request<ExchangeRateResponse>({
        method: 'GET',
        url: '/api/exchange/rate',
        params: { from: 'USD', to: defaultCurrency },
      });
      return resp.data;
    },
    enabled: defaultCurrency !== 'USD',
    staleTime: 5 * 60 * 1000,
  });

  const options = (currencies ?? []).map(c => ({ label: c, value: c }));
  const selectedOption = options.find(o => o.value === defaultCurrency) || { label: defaultCurrency, value: defaultCurrency };

  const handleCurrencyChange = (option: SelectOption<string>) => {
    setDefaultCurrency(option.value);
  };

  const showRates = defaultCurrency !== 'USD';
  const loading = loadingToUsd || loadingFromUsd;

  return (
    <div className={styles.card}>
      <Typography className={styles.title} weight='semibold'>
        Exchange Rates
      </Typography>

      {showRates ? (
        <div className={styles.rates}>
          {loading ? (
            <span className={styles.loadingText}>Fetching Google Finance rates...</span>
          ) : (
            <>
              <div className={styles.rateItem}>
                <span>1 {defaultCurrency} =</span>
                <span className={styles.rateValue}>
                  {rateToUsd?.rate?.toFixed(4) || '—'} USD
                </span>
              </div>
              <div className={styles.rateItem}>
                <span>1 USD =</span>
                <span className={styles.rateValue}>
                  {rateFromUsd?.rate?.toFixed(4) || '—'} {defaultCurrency}
                </span>
              </div>
            </>
          )}
        </div>
      ) : (
        <div className={styles.rates}>
          <span className={styles.loadingText}>USD (Base Currency)</span>
        </div>
      )}

      <div className={styles.selectWrapper}>
        <Select
          disabled={currenciesLoading}
          isLoading={currenciesLoading}
          label='Default Currency'
          onChange={handleCurrencyChange}
          options={options}
          value={selectedOption}
        />
      </div>
    </div>
  );
}
