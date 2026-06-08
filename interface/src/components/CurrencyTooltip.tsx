import React from 'react';
import { useQuery } from '@tanstack/react-query';

import { Tooltip, TooltipContent, TooltipTrigger } from '@monetr/interface/components/Tooltip';
import { useDefaultConversionCurrency } from '@monetr/interface/hooks/useDefaultConversionCurrency';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { AmountType, formatAmount } from '@monetr/interface/util/amounts';
import request from '@monetr/interface/util/request';

interface CurrencyTooltipProps {
  children: React.ReactNode;
  amount: number; // Stored units (e.g. cents/paisa)
  currency?: string; // Optional, falls back to active currency
}

interface ExchangeRateResponse {
  from: string;
  to: string;
  rate: number;
}

export default function CurrencyTooltip({ children, amount, currency: explicitCurrency }: CurrencyTooltipProps): React.JSX.Element {
  const { data: localeData } = useLocaleCurrency();
  const [defaultCurrency] = useDefaultConversionCurrency();

  const activeCurrency = explicitCurrency || localeData?.currency || 'USD';
  const locale = localeData?.locale || 'en-US';

  const isNonUSD = activeCurrency !== 'USD';

  // Determine target currency for tooltip conversion:
  // - If source is not USD, we convert to USD.
  // - If source is USD, we convert to the user's selected default conversion currency.
  const targetCurrency = isNonUSD ? 'USD' : (defaultCurrency !== 'USD' ? defaultCurrency : null);
  const fromCurrency = activeCurrency;
  const toCurrency = targetCurrency;

  const { data, isLoading, isError } = useQuery<ExchangeRateResponse>({
    queryKey: ['/api/exchange/rate', fromCurrency, toCurrency],
    queryFn: async () => {
      const resp = await request<ExchangeRateResponse>({
        method: 'GET',
        url: '/api/exchange/rate',
        params: { from: fromCurrency, to: toCurrency },
      });
      return resp.data;
    },
    enabled: Boolean(toCurrency && fromCurrency && fromCurrency !== toCurrency),
    staleTime: 5 * 60 * 1000, // 5 minutes stale time
    cacheTime: 10 * 60 * 1000,
  });

  if (!toCurrency || fromCurrency === toCurrency) {
    return <>{children}</>;
  }

  const rate = data?.rate;
  let tooltipContent = '';

  if (isLoading) {
    tooltipContent = 'Loading conversion...';
  } else if (isError || !rate) {
    tooltipContent = 'Conversion unavailable';
  } else {
    // Perform conversion: multiply the raw stored value by rate and format
    const convertedAmount = Math.round(amount * rate);
    const formatted = formatAmount(convertedAmount, AmountType.Stored, locale, toCurrency);
    tooltipContent = `≈ ${formatted}`;
  }

  return (
    <Tooltip delayDuration={300}>
      <TooltipTrigger asChild>
        <span
          style={{
            cursor: 'help',
            borderBottom: '1px dotted rgba(255, 255, 255, 0.35)',
            display: 'inline-flex',
            alignItems: 'center',
          }}
        >
          {children}
        </span>
      </TooltipTrigger>
      <TooltipContent side='top'>{tooltipContent}</TooltipContent>
    </Tooltip>
  );
}
