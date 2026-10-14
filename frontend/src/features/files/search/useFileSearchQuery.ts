import { useCallback, useEffect, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import useDebouncedValue from '@/components/hooks/useDebouncedValue/useDebouncedValue';

const searchQueryParam = 'q';
const minSearchQueryLength = 2;
const searchDebounceMs = 300;

interface SearchQueryOptions {
    minQueryLength?: number;
    shouldAlwaysReplaceHistory?: boolean;
}

const useFileSearchQuery = ({
    minQueryLength = minSearchQueryLength,
    shouldAlwaysReplaceHistory = false,
}: SearchQueryOptions = {}) => {
    const [searchParams, setSearchParams] = useSearchParams();
    const urlQuery = searchParams.get(searchQueryParam) ?? '';
    const [inputValue, setInputValue] = useState(urlQuery);
    const [observedUrlQuery, setObservedUrlQuery] = useState(urlQuery);
    const [lastWrittenQuery, setLastWrittenQuery] = useState(urlQuery);
    const debouncedInputValue = useDebouncedValue(inputValue, searchDebounceMs);

    if (observedUrlQuery !== urlQuery) {
        setObservedUrlQuery(urlQuery);
        if (urlQuery !== lastWrittenQuery) {
            setLastWrittenQuery(urlQuery);
            setInputValue(urlQuery);
        }
    }

    const writeQueryToUrl = useCallback(
        (nextQuery: string) => {
            setLastWrittenQuery(nextQuery);
            setSearchParams(
                (currentParams) => {
                    const nextParams = new URLSearchParams(currentParams);
                    if (nextQuery) {
                        nextParams.set(searchQueryParam, nextQuery);
                    } else {
                        nextParams.delete(searchQueryParam);
                    }
                    return nextParams;
                },
                { replace: shouldAlwaysReplaceHistory || urlQuery !== '' }
            );
        },
        [setSearchParams, shouldAlwaysReplaceHistory, urlQuery]
    );

    useEffect(() => {
        if (debouncedInputValue !== inputValue) {
            return;
        }
        const settledQuery = debouncedInputValue.trim();
        if (settledQuery === urlQuery) {
            return;
        }
        let isCancelled = false;
        queueMicrotask(() => {
            if (!isCancelled) {
                writeQueryToUrl(settledQuery);
            }
        });
        return () => {
            isCancelled = true;
        };
    }, [debouncedInputValue, inputValue, urlQuery, writeQueryToUrl]);

    const clearQuery = useCallback(() => {
        setInputValue('');
        writeQueryToUrl('');
    }, [writeQueryToUrl]);

    const trimmedUrlQuery = urlQuery.trim();
    const activeQuery = trimmedUrlQuery.length >= minQueryLength ? trimmedUrlQuery : '';

    return { inputValue, setInputValue, activeQuery, clearQuery };
};

export default useFileSearchQuery;
