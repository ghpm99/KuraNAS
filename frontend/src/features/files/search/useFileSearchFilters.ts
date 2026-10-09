import { useCallback, useMemo } from 'react';
import { useSearchParams } from 'react-router-dom';
import {
    emptyFileSearchFilters,
    parseFileSearchFilters,
    writeFileSearchFilters,
    type FileSearchFilters,
} from './fileSearchFilters';

const useFileSearchFilters = () => {
    const [searchParams, setSearchParams] = useSearchParams();
    const serializedParams = searchParams.toString();
    const filters = useMemo(
        () => parseFileSearchFilters(new URLSearchParams(serializedParams)),
        [serializedParams]
    );

    const setFilters = useCallback(
        (nextFilters: FileSearchFilters) => {
            setSearchParams(
                (currentParams) => writeFileSearchFilters(currentParams, nextFilters),
                { replace: true }
            );
        },
        [setSearchParams]
    );

    const resetFilters = useCallback(() => setFilters(emptyFileSearchFilters), [setFilters]);

    return { filters, setFilters, resetFilters };
};

export default useFileSearchFilters;
