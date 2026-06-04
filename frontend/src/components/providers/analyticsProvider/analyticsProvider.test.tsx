import { render, screen, renderHook, act } from '@testing-library/react';
import { AnalyticsProvider } from './index';
import { useAnalyticsOverview } from './analyticsContext';

const storageData = {
    storage: { total_bytes: 1000, used_bytes: 400, free_bytes: 600, growth_bytes: 50 },
    counts: { files_total: 200, files_added: 15, folders: 10 },
};

let lastQueriesArg: any;
const refetchSpy = jest.fn().mockResolvedValue({});

// Controls per-query state for the 12 slice queries the provider issues.
let mode: 'loaded' | 'loading' | 'storage-error' = 'loaded';

jest.mock('@tanstack/react-query', () => ({
    useQueries: (arg: any) => {
        lastQueriesArg = arg;
        return arg.queries.map((_query: any, index: number) => {
            const isStorage = index === 0;
            if (mode === 'loading') {
                return {
                    data: undefined,
                    isLoading: true,
                    isError: false,
                    dataUpdatedAt: 0,
                    refetch: refetchSpy,
                };
            }
            if (mode === 'storage-error' && isStorage) {
                return {
                    data: undefined,
                    isLoading: false,
                    isError: true,
                    dataUpdatedAt: 0,
                    refetch: refetchSpy,
                };
            }
            return {
                data: isStorage ? storageData : [],
                isLoading: false,
                isError: false,
                dataUpdatedAt: 1_700_000_000_000,
                refetch: refetchSpy,
            };
        });
    },
}));

function ConsumerComponent() {
    const ctx = useAnalyticsOverview();
    return (
        <div>
            <span data-testid="period">{ctx.period}</span>
            <span data-testid="loading">{String(ctx.loading)}</span>
            <span data-testid="error">{ctx.error || 'none'}</span>
            <span data-testid="has-data">{ctx.data ? 'yes' : 'no'}</span>
            <span data-testid="files">{ctx.data?.counts.files_total ?? 0}</span>
            <button data-testid="refresh" onClick={() => void ctx.refresh()}>
                Refresh
            </button>
            <button data-testid="set-period" onClick={() => ctx.setPeriod('30d')}>
                Set 30d
            </button>
        </div>
    );
}

describe('AnalyticsProvider', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        lastQueriesArg = undefined;
        mode = 'loaded';
    });

    it('composes overview data from slice queries', () => {
        render(
            <AnalyticsProvider>
                <ConsumerComponent />
            </AnalyticsProvider>
        );

        expect(screen.getByTestId('period').textContent).toBe('7d');
        expect(screen.getByTestId('has-data').textContent).toBe('yes');
        expect(screen.getByTestId('files').textContent).toBe('200');
        expect(screen.getByTestId('error').textContent).toBe('none');
    });

    it('keys the period-dependent queries with the current period', () => {
        render(
            <AnalyticsProvider>
                <ConsumerComponent />
            </AnalyticsProvider>
        );

        expect(lastQueriesArg.queries[0].queryKey).toEqual(['analytics', 'storage', '7d']);
    });

    it('reports loading while slices resolve', () => {
        mode = 'loading';
        render(
            <AnalyticsProvider>
                <ConsumerComponent />
            </AnalyticsProvider>
        );

        expect(screen.getByTestId('loading').textContent).toBe('true');
        expect(screen.getByTestId('has-data').textContent).toBe('no');
    });

    it('surfaces an error key when the storage slice fails', () => {
        mode = 'storage-error';
        render(
            <AnalyticsProvider>
                <ConsumerComponent />
            </AnalyticsProvider>
        );

        expect(screen.getByTestId('error').textContent).toBe('ANALYTICS_ERROR_LOAD_BLOCK');
        expect(screen.getByTestId('has-data').textContent).toBe('no');
    });

    it('refreshes every slice and updates the period', () => {
        render(
            <AnalyticsProvider>
                <ConsumerComponent />
            </AnalyticsProvider>
        );

        act(() => {
            screen.getByTestId('refresh').click();
        });
        expect(refetchSpy).toHaveBeenCalled();

        act(() => {
            screen.getByTestId('set-period').click();
        });
        expect(screen.getByTestId('period').textContent).toBe('30d');
        expect(lastQueriesArg.queries[0].queryKey).toEqual(['analytics', 'storage', '30d']);
    });
});

describe('useAnalyticsOverview', () => {
    it('throws when used outside the provider', () => {
        jest.spyOn(console, 'error').mockImplementation(() => {});

        expect(() => {
            renderHook(() => useAnalyticsOverview());
        }).toThrow('useAnalyticsOverview must be used within an AnalyticsProvider');

        (console.error as jest.Mock).mockRestore();
    });
});
