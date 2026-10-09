import { useCallback, useEffect, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { getMusicRoute } from '@/app/routes';
import useDebouncedValue from '@/components/hooks/useDebouncedValue/useDebouncedValue';

export const MUSIC_SEARCH_MIN_LENGTH = 2;
export const MUSIC_SEARCH_DEBOUNCE_MS = 300;

type MusicSearchLocationState = { musicSearchReturnTo?: string } | null;

const musicSearchPath = getMusicRoute('search');

export const useMusicSearchField = () => {
    const location = useLocation();
    const navigate = useNavigate();
    const isSearchView = location.pathname === musicSearchPath;
    const currentTerm = isSearchView ? (new URLSearchParams(location.search).get('q') ?? '') : '';
    const returnTo = (location.state as MusicSearchLocationState)?.musicSearchReturnTo;

    const [inputText, setInputText] = useState(currentTerm);
    const [observedTerm, setObservedTerm] = useState(currentTerm);
    const debouncedTerm = useDebouncedValue(inputText.trim(), MUSIC_SEARCH_DEBOUNCE_MS);

    if (currentTerm !== observedTerm) {
        setObservedTerm(currentTerm);
        if (currentTerm !== inputText.trim()) {
            setInputText(currentTerm);
        }
    }

    const goToSearch = useCallback(
        (term: string, shouldReplaceHistory: boolean) => {
            const musicSearchReturnTo = isSearchView
                ? returnTo
                : `${location.pathname}${location.search}`;
            navigate(`${musicSearchPath}?q=${encodeURIComponent(term)}`, {
                replace: shouldReplaceHistory,
                state: { musicSearchReturnTo },
            });
        },
        [isSearchView, returnTo, location.pathname, location.search, navigate]
    );

    useEffect(() => {
        const isStale = debouncedTerm !== inputText.trim();
        const isTermTooShort = debouncedTerm.length < MUSIC_SEARCH_MIN_LENGTH;
        const isAlreadyShown = isSearchView && debouncedTerm === currentTerm;
        if (isStale || isTermTooShort || isAlreadyShown) {
            return;
        }
        goToSearch(debouncedTerm, isSearchView);
    }, [debouncedTerm, inputText, isSearchView, currentTerm, goToSearch]);

    const submit = () => {
        const term = inputText.trim();
        if (term === '') {
            return;
        }
        goToSearch(term, false);
    };

    const clear = () => {
        setInputText('');
        if (isSearchView) {
            navigate(returnTo ?? getMusicRoute('home'));
        }
    };

    return { inputText, setInputText, submit, clear };
};
