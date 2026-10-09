import type { ReactNode } from 'react';
import { GlobalSearchContext } from './useGlobalSearch';
import GlobalSearchDialog from './GlobalSearchDialog';
import useGlobalSearchProvider from './useGlobalSearchProvider';

export const GlobalSearchProvider = ({ children }: { children: ReactNode }) => {
    const {
        open,
        query,
        sections,
        isFetching,
        isUpdating,
        hasSearchError,
        searchErrorMessage,
        retrySearch,
        suggestion,
        isFuzzyResult,
        recentSearches,
        rerunRecentSearch,
        forgetRecentSearch,
        clearRecentSearches,
        activeItemId,
        shortcut,
        showEmptyState,
        openSearch,
        closeSearch,
        setQuery,
        setActiveIndex,
        handleInputKeyDown,
        activateItem,
        activateSecondaryAction,
    } = useGlobalSearchProvider();

    const flattenedItems = sections.flatMap((section) => section.items);

    return (
        <GlobalSearchContext.Provider value={{ openSearch, shortcut }}>
            {children}
            <GlobalSearchDialog
                open={open}
                query={query}
                sections={sections}
                isFetching={isFetching}
                isUpdating={isUpdating}
                hasSearchError={hasSearchError}
                searchErrorMessage={searchErrorMessage}
                onRetry={retrySearch}
                suggestion={suggestion}
                isFuzzyResult={isFuzzyResult}
                recentSearches={recentSearches}
                onRecentSearchSelect={rerunRecentSearch}
                onRecentSearchRemove={forgetRecentSearch}
                onRecentSearchesClear={clearRecentSearches}
                activeItemId={activeItemId}
                shortcut={shortcut}
                showEmptyState={showEmptyState}
                onClose={closeSearch}
                onQueryChange={setQuery}
                onInputKeyDown={handleInputKeyDown}
                onItemHover={(itemId) => {
                    const nextIndex = flattenedItems.findIndex((item) => item.id === itemId);
                    if (nextIndex >= 0) {
                        setActiveIndex(nextIndex);
                    }
                }}
                onItemSelect={activateItem}
                onItemSecondaryAction={activateSecondaryAction}
            />
        </GlobalSearchContext.Provider>
    );
};

export default GlobalSearchProvider;
