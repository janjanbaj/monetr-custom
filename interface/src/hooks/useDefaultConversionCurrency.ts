import { useCallback, useEffect, useState } from 'react';

const STORAGE_KEY = 'monetr_default_conversion_currency';
const EVENT_NAME = 'monetr-default-currency-change';

export function useDefaultConversionCurrency() {
  const [currency, setCurrency] = useState<string>(() => {
    return localStorage.getItem(STORAGE_KEY) || 'NPR';
  });

  const changeCurrency = useCallback((newCurrency: string) => {
    localStorage.setItem(STORAGE_KEY, newCurrency);
    setCurrency(newCurrency);
    // Dispatch a custom event to notify other hook instances on the same page
    window.dispatchEvent(new CustomEvent(EVENT_NAME, { detail: newCurrency }));
  }, []);

  useEffect(() => {
    const handleEvent = (event: Event) => {
      const customEvent = event as CustomEvent<string>;
      setCurrency(customEvent.detail);
    };

    window.addEventListener(EVENT_NAME, handleEvent);

    // Sync across browser tabs/windows
    const handleStorage = (event: StorageEvent) => {
      if (event.key === STORAGE_KEY && event.newValue) {
        setCurrency(event.newValue);
      }
    };
    window.addEventListener('storage', handleStorage);

    return () => {
      window.removeEventListener(EVENT_NAME, handleEvent);
      window.removeEventListener('storage', handleStorage);
    };
  }, []);

  return [currency, changeCurrency] as const;
}
