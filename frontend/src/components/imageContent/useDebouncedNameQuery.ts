import { useEffect, useState } from 'react';

const searchDebounceMs = 300;

export const useDebouncedNameQuery = (
    committedNameQuery: string,
    commitNameQuery: (nameQuery: string) => void
) => {
    const [typedNameQuery, setTypedNameQuery] = useState(committedNameQuery);

    useEffect(() => {
        setTypedNameQuery(committedNameQuery);
    }, [committedNameQuery]);

    useEffect(() => {
        if (typedNameQuery.trim() === committedNameQuery) {
            return;
        }
        const timeoutId = window.setTimeout(
            () => commitNameQuery(typedNameQuery),
            searchDebounceMs
        );
        return () => window.clearTimeout(timeoutId);
    }, [typedNameQuery, committedNameQuery, commitNameQuery]);

    return { typedNameQuery, setTypedNameQuery };
};
