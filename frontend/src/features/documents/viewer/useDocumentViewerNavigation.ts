import { useEffect, useMemo } from 'react';
import type { FileData } from '@/features/files/providers/fileProvider/fileContext';
import { documentToFileData, searchResultToFileData } from '../documentFileAdapter';
import { resolveDocumentsListScope } from '../documentsListScope';
import useDocumentsList from '../useDocumentsList';
import useDocumentsSearch from '../useDocumentsSearch';

export type DocumentViewerEntry = {
    file: FileData;
    pageNumber?: number;
};

const maxEntriesScannedForCurrentDocument = 500;

const useDocumentViewerNavigation = (fileId: number, listSearch: string) => {
    const scope = useMemo(() => resolveDocumentsListScope(listSearch), [listSearch]);
    const isSearching = scope.searchQuery !== '';
    const list = useDocumentsList(scope.filters, scope.sort, !isSearching);
    const search = useDocumentsSearch(scope.searchQuery, scope.filters);

    const entries = useMemo<DocumentViewerEntry[]>(
        () =>
            isSearching
                ? search.results.map((result) => ({
                      file: searchResultToFileData(result),
                      pageNumber: result.page ?? undefined,
                  }))
                : list.documents.map((document) => ({ file: documentToFileData(document) })),
        [isSearching, search.results, list.documents]
    );

    const source = isSearching ? search : list;
    const currentIndex = entries.findIndex((entry) => entry.file.id === fileId);
    const isCurrentDocumentLoaded = currentIndex !== -1;
    const isAtLoadedEdge = isCurrentDocumentLoaded && currentIndex >= entries.length - 1;
    const isStillLookingForCurrent =
        !isCurrentDocumentLoaded &&
        entries.length > 0 &&
        entries.length < maxEntriesScannedForCurrentDocument;
    const shouldLoadMore =
        (isAtLoadedEdge || isStillLookingForCurrent) &&
        source.hasNextPage &&
        !source.isFetchingNextPage &&
        !source.isNextPageError;
    const { fetchNextPage } = source;

    useEffect(() => {
        if (shouldLoadMore) fetchNextPage();
    }, [shouldLoadMore, fetchNextPage, entries.length]);

    return {
        currentEntry: isCurrentDocumentLoaded ? entries[currentIndex] : undefined,
        previousEntry: isCurrentDocumentLoaded ? entries[currentIndex - 1] : undefined,
        nextEntry: isCurrentDocumentLoaded ? entries[currentIndex + 1] : undefined,
    };
};

export default useDocumentViewerNavigation;
