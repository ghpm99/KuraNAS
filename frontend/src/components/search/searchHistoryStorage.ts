export const searchHistoryStorageKey = 'kuranas.globalSearch.recentQueries';
export const searchHistoryMaxEntries = 10;
export const searchHistoryMinCharacters = 2;

const normalizeHistoryEntry = (entry: string) => entry.trim().toLowerCase();

const isHistoryEntry = (entry: unknown): entry is string =>
    typeof entry === 'string' && entry.trim().length >= searchHistoryMinCharacters;

export const readSearchHistory = (): string[] => {
    try {
        const storedValue = window.localStorage.getItem(searchHistoryStorageKey);
        if (!storedValue) {
            return [];
        }
        const parsedValue: unknown = JSON.parse(storedValue);
        if (!Array.isArray(parsedValue)) {
            return [];
        }
        return parsedValue.filter(isHistoryEntry).slice(0, searchHistoryMaxEntries);
    } catch {
        return [];
    }
};

const writeSearchHistory = (entries: string[]) => {
    try {
        window.localStorage.setItem(searchHistoryStorageKey, JSON.stringify(entries));
    } catch {
        return;
    }
};

export const addSearchHistoryEntry = (entries: string[], query: string): string[] => {
    const trimmedQuery = query.trim();
    if (trimmedQuery.length < searchHistoryMinCharacters) {
        return entries;
    }
    const normalizedQuery = normalizeHistoryEntry(trimmedQuery);
    const otherEntries = entries.filter(
        (entry) => normalizeHistoryEntry(entry) !== normalizedQuery
    );
    const updatedEntries = [trimmedQuery, ...otherEntries].slice(0, searchHistoryMaxEntries);
    writeSearchHistory(updatedEntries);
    return updatedEntries;
};

export const removeSearchHistoryEntry = (entries: string[], query: string): string[] => {
    const updatedEntries = entries.filter((entry) => entry !== query);
    writeSearchHistory(updatedEntries);
    return updatedEntries;
};

export const clearSearchHistory = (): string[] => {
    writeSearchHistory([]);
    return [];
};
