import { useCallback, useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import { AnalyticsPeriod } from '@/types/analytics';

const periodParamName = 'period';
const defaultPeriod: AnalyticsPeriod = '7d';
const validPeriods: readonly AnalyticsPeriod[] = ['24h', '7d', '30d', '90d'];

const isAnalyticsPeriod = (rawPeriod: string | null): rawPeriod is AnalyticsPeriod =>
    validPeriods.includes(rawPeriod as AnalyticsPeriod);

export const useAnalyticsPeriodParam = () => {
    const [searchParams, setSearchParams] = useSearchParams();
    const rawPeriod = searchParams.get(periodParamName);
    const hasInvalidPeriod = rawPeriod !== null && !isAnalyticsPeriod(rawPeriod);
    const period = isAnalyticsPeriod(rawPeriod) ? rawPeriod : defaultPeriod;

    const setPeriod = useCallback(
        (nextPeriod: AnalyticsPeriod) => {
            setSearchParams(
                (currentParams) => {
                    const nextParams = new URLSearchParams(currentParams);
                    nextParams.set(periodParamName, nextPeriod);
                    return nextParams;
                },
                { replace: true }
            );
        },
        [setSearchParams]
    );

    useEffect(() => {
        if (hasInvalidPeriod) {
            setPeriod(defaultPeriod);
        }
    }, [hasInvalidPeriod, setPeriod]);

    return { period, setPeriod };
};
