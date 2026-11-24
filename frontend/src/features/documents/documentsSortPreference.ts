import {
    documentSortFields,
    documentSortOrders,
    type DocumentsSortState,
} from './documentsFilters';

export const documentsSortStorageKey = 'kuranas.documents.sort';

const isDocumentsSortState = (candidate: unknown): candidate is DocumentsSortState => {
    if (typeof candidate !== 'object' || candidate === null) return false;
    const { sort, order } = candidate as Partial<DocumentsSortState>;
    return (
        documentSortFields.some((field) => field === sort) &&
        documentSortOrders.some((direction) => direction === order)
    );
};

export const loadDocumentsSort = (): DocumentsSortState | null => {
    try {
        const storedSort = window.localStorage.getItem(documentsSortStorageKey);
        if (!storedSort) return null;
        const parsedSort: unknown = JSON.parse(storedSort);
        return isDocumentsSortState(parsedSort) ? parsedSort : null;
    } catch {
        return null;
    }
};

export const saveDocumentsSort = (sort: DocumentsSortState): void => {
    try {
        window.localStorage.setItem(documentsSortStorageKey, JSON.stringify(sort));
    } catch {
        return;
    }
};
