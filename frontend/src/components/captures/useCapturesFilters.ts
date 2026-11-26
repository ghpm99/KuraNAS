import { useCallback, useEffect, useMemo, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import useDebouncedValue from '@/components/hooks/useDebouncedValue/useDebouncedValue';
import type { CaptureStatus } from '@/types/captures';
import {
    emptyCapturesFilters,
    hasActiveCapturesFilters,
    parseCapturesFilters,
    writeCapturesFilters,
    type CapturesFilters,
} from './capturesFilters';

const nameDebounceMs = 300;

export const useCapturesFilters = () => {
    const [searchParams, setSearchParams] = useSearchParams();
    const appliedFilters = useMemo(() => parseCapturesFilters(searchParams), [searchParams]);
    const [nameInput, setNameInput] = useState(appliedFilters.name);
    const debouncedNameInput = useDebouncedValue(nameInput, nameDebounceMs);

    const updateFilters = useCallback(
        (changes: Partial<CapturesFilters>) => {
            setSearchParams(
                (currentParams) =>
                    writeCapturesFilters(currentParams, {
                        ...parseCapturesFilters(currentParams),
                        ...changes,
                    }),
                { replace: true }
            );
        },
        [setSearchParams]
    );

    useEffect(() => {
        if (debouncedNameInput !== nameInput) return;
        const settledName = debouncedNameInput.trim();
        if (settledName === appliedFilters.name) return;
        updateFilters({ name: settledName });
    }, [debouncedNameInput, nameInput, appliedFilters.name, updateFilters]);

    const changeStatus = useCallback(
        (status: CaptureStatus | '') => updateFilters({ status }),
        [updateFilters]
    );

    const clearFilters = useCallback(() => {
        setNameInput('');
        setSearchParams(
            (currentParams) => writeCapturesFilters(currentParams, emptyCapturesFilters),
            {
                replace: true,
            }
        );
    }, [setSearchParams]);

    return {
        appliedFilters,
        nameInput,
        changeNameInput: setNameInput,
        changeStatus,
        clearFilters,
        hasActiveFilters: hasActiveCapturesFilters(appliedFilters),
    };
};
