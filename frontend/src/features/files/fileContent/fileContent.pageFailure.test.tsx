import { act, fireEvent, render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import FileContent from './fileContent';
import FileProvider from '../providers/fileProvider';
import useFileSearchResults from '../search/useFileSearchResults';
import { createTestFile } from '../selection/testFileFactory';
import { apiBase } from '@/service';
import { buildServerPage } from '@/shared/test/buildServerPage';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn(), post: jest.fn(), put: jest.fn(), delete: jest.fn() },
}));
jest.mock('react-router-dom', () => ({
    ...jest.requireActual('react-router-dom'),
    useLocation: () => ({ pathname: '/files', search: '', hash: '', state: null, key: 'default' }),
    useNavigate: () => jest.fn(),
}));
jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));
jest.mock('@/components/hooks/useMediaOpener/useMediaOpener', () => ({
    __esModule: true,
    default: () => ({ openMediaItem: () => false }),
}));
jest.mock('@/components/loadMoreSentinel/loadMoreSentinel', () => ({
    __esModule: true,
    default: ({
        hasNextPage,
        fetchNextPage,
    }: {
        hasNextPage: boolean;
        fetchNextPage: () => void;
    }) => (hasNextPage ? <button onClick={fetchNextPage}>load-more</button> : null),
}));

const mockedGet = (apiBase as unknown as { get: jest.Mock }).get;

const firstPageFile = createTestFile(1, { name: 'primeira.txt' });
const secondPageFile = createTestFile(2, { name: 'segunda.txt' });
const firstPage = buildServerPage([firstPageFile], { page: 1, pageSize: 60, hasNext: true });
const secondPage = buildServerPage([secondPageFile], { page: 2, pageSize: 60, hasNext: false });

const emptyPage = buildServerPage([], { page: 1, pageSize: 60, hasNext: false });

const pageOfRequest = ([, config]: [string, { params: { page: number } }]) => config.params.page;

const requestedPages = (listingUrl: string): number[] =>
    mockedGet.mock.calls.filter(([url]) => url === listingUrl).map(pageOfRequest);

const respondWithSecondPageFailingOnce = (listingUrl: string) => {
    let hasSecondPageFailed = false;
    mockedGet.mockImplementation((url: string, config: { params: { page: number } }) => {
        if (url !== listingUrl) return Promise.resolve({ data: emptyPage });
        if (config.params.page === 1) return Promise.resolve({ data: firstPage });
        if (!hasSecondPageFailed) {
            hasSecondPageFailed = true;
            return Promise.reject(new Error('falha na página 2'));
        }
        return Promise.resolve({ data: secondPage });
    });
};

const SearchedFiles = () => {
    const searchListing = useFileSearchResults({ query: 'txt', isRecursive: true });
    return <FileContent searchListing={searchListing} />;
};

const renderInProvider = (content: React.ReactNode) => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const rendered = render(
        <QueryClientProvider client={queryClient}>
            <FileProvider>{content}</FileProvider>
        </QueryClientProvider>
    );
    return { ...rendered, queryClient };
};

const failSecondPageLoad = async () => {
    fireEvent.click(await screen.findByRole('button', { name: 'load-more' }));
    await screen.findByRole('alert');
};

const expectBothPagesWithoutError = async () => {
    expect(await screen.findByText('segunda.txt')).toBeInTheDocument();
    expect(screen.getByText('primeira.txt')).toBeInTheDocument();
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
};

describe('FileContent when a later page fails', () => {
    beforeEach(() => jest.clearAllMocks());

    it('keeps the first page, shows the error below and retries page 2 in the folder listing', async () => {
        respondWithSecondPageFailingOnce('/files/tree');
        renderInProvider(<FileContent />);

        await failSecondPageLoad();
        expect(screen.getByText('primeira.txt')).toBeInTheDocument();
        expect(screen.queryByRole('button', { name: 'load-more' })).not.toBeInTheDocument();

        fireEvent.click(screen.getByRole('button', { name: 'TRY_AGAIN' }));

        await expectBothPagesWithoutError();
        expect(requestedPages('/files/tree')).toEqual([1, 2, 2]);
    });

    it('keeps the first page, shows the error below and retries page 2 in the folder search', async () => {
        respondWithSecondPageFailingOnce('/files/search');
        renderInProvider(<SearchedFiles />);

        await failSecondPageLoad();
        expect(screen.getByText('primeira.txt')).toBeInTheDocument();

        fireEvent.click(screen.getByRole('button', { name: 'TRY_AGAIN' }));

        await expectBothPagesWithoutError();
        expect(requestedPages('/files/search')).toEqual([1, 2, 2]);
    });

    it('keeps the first page without an error when a background revalidation fails', async () => {
        respondWithSecondPageFailingOnce('/files/tree');
        const { queryClient } = renderInProvider(<FileContent />);
        await screen.findByText('primeira.txt');

        mockedGet.mockClear();
        mockedGet.mockRejectedValue(new Error('servidor fora'));
        await act(() => queryClient.refetchQueries());

        expect(mockedGet).toHaveBeenCalled();
        expect(screen.getByText('primeira.txt')).toBeInTheDocument();
        expect(screen.queryByRole('alert')).not.toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'load-more' })).toBeInTheDocument();
    });

    it('shows the error state, not the empty one, when the server is down from the start', async () => {
        mockedGet.mockImplementation((url: string) =>
            url === '/files/tree'
                ? Promise.reject(new Error('servidor fora'))
                : Promise.resolve({ data: emptyPage })
        );
        renderInProvider(<FileContent />);

        expect(await screen.findByRole('alert')).toHaveTextContent('FILES_LISTING_ERROR_TITLE');
        expect(screen.queryByText('FILES_EMPTY_FOLDER_TITLE')).not.toBeInTheDocument();
    });
});
