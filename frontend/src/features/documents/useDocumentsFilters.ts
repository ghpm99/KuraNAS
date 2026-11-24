import { useCallback, useMemo } from 'react';
import { useSearchParams } from 'react-router-dom';
import {
    countActiveDocumentsFilters,
    defaultDocumentsSort,
    emptyDocumentsFilters,
    parseDocumentsFilters,
    parseDocumentsFolderId,
    parseDocumentsSort,
    toDocumentFilters,
    toDocumentSort,
    writeDocumentsFilters,
    writeDocumentsFolderId,
    writeDocumentsSort,
    type DocumentsFilterState,
    type DocumentsSortState,
} from './documentsFilters';
import { loadDocumentsSort, saveDocumentsSort } from './documentsSortPreference';

const useDocumentsFilters = () => {
    const [searchParams, setSearchParams] = useSearchParams();
    const serializedParams = searchParams.toString();

    const filters = useMemo(
        () => parseDocumentsFilters(new URLSearchParams(serializedParams)),
        [serializedParams]
    );
    const sortState = useMemo(
        () =>
            parseDocumentsSort(new URLSearchParams(serializedParams)) ??
            loadDocumentsSort() ??
            defaultDocumentsSort,
        [serializedParams]
    );
    const folderId = useMemo(
        () => parseDocumentsFolderId(new URLSearchParams(serializedParams)),
        [serializedParams]
    );
    const requestFilters = useMemo(
        () => ({ ...toDocumentFilters(filters), folderId }),
        [filters, folderId]
    );
    const requestSort = useMemo(() => toDocumentSort(sortState), [sortState]);

    const setFilters = useCallback(
        (nextFilters: DocumentsFilterState) => {
            setSearchParams((currentParams) => writeDocumentsFilters(currentParams, nextFilters), {
                replace: true,
            });
        },
        [setSearchParams]
    );

    const setSort = useCallback(
        (nextSort: DocumentsSortState) => {
            saveDocumentsSort(nextSort);
            setSearchParams((currentParams) => writeDocumentsSort(currentParams, nextSort), {
                replace: true,
            });
        },
        [setSearchParams]
    );

    const setFolderId = useCallback(
        (nextFolderId: number | undefined) => {
            setSearchParams(
                (currentParams) => writeDocumentsFolderId(currentParams, nextFolderId),
                { replace: true }
            );
        },
        [setSearchParams]
    );

    const resetFilters = useCallback(() => setFilters(emptyDocumentsFilters), [setFilters]);

    return {
        filters,
        sortState,
        folderId,
        requestFilters,
        requestSort,
        activeFilterCount: countActiveDocumentsFilters(filters),
        setFilters,
        setSort,
        setFolderId,
        resetFilters,
    };
};

export default useDocumentsFilters;
