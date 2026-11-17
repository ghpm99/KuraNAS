import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import AnalyticsDuplicatesScreen from './AnalyticsDuplicatesScreen';
import { getDuplicateGroupsPage } from '@/service/analytics';
import { deleteFile } from '@/service/files';
import { buildServerPage } from '@/shared/test/buildServerPage';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import { buildDuplicateCopy, buildDuplicateGroup } from './duplicateGroupFixtures';
import { useLocation } from 'react-router-dom';

jest.mock('@/service/analytics', () => ({ getDuplicateGroupsPage: jest.fn() }));
jest.mock('@/service/files', () => ({ deleteFile: jest.fn() }));
jest.mock('@/components/i18n/provider/i18nContext', () => ({
    ...jest.requireActual('@/components/i18n/provider/i18nContext'),
    __esModule: true,
    default: () => ({
        t: (key: string, variables?: Record<string, string>) =>
            variables ? `${key} ${JSON.stringify(variables)}` : key,
    }),
}));

const mockedGetGroupsPage = getDuplicateGroupsPage as jest.Mock;
const mockedDeleteFile = deleteFile as jest.Mock;

const firstCopies = [
    buildDuplicateCopy(1, { path: '/Fotos/a/foto.jpg', tier: 'hot' }),
    buildDuplicateCopy(2, { path: '/Fotos/b/foto.jpg', tier: 'cold' }),
    buildDuplicateCopy(3, { path: '/Fotos/c/foto.jpg' }),
];
const secondCopies = [
    buildDuplicateCopy(4, { path: '/Docs/x.pdf' }),
    buildDuplicateCopy(5, { path: '/Docs/y.pdf' }),
];
const firstGroup = buildDuplicateGroup('sig-first', firstCopies);
const secondGroup = buildDuplicateGroup('sig-second', secondCopies);

const SearchProbe = () => <span data-testid="search">{useLocation().search}</span>;

const renderScreen = (entry = '/analytics/duplicates') =>
    renderWithoutBackend(
        <>
            <AnalyticsDuplicatesScreen />
            <SearchProbe />
        </>,
        [entry]
    );

const onePage = (groups = [firstGroup, secondGroup]) =>
    buildServerPage(groups, { page: 1, pageSize: 50, hasNext: false });

const expandGroupWithCopies = async (copyCount: number) => {
    const header = await screen.findByRole('button', {
        name: new RegExp(`"copies":"${copyCount}"`),
    });
    fireEvent.click(header);
};

const confirmDialog = () => screen.getByRole('dialog');

describe('AnalyticsDuplicatesScreen', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedGetGroupsPage.mockResolvedValue(onePage());
        mockedDeleteFile.mockResolvedValue(undefined);
    });

    it('lists the groups with size, copies and reclaimable space and expands the copies', async () => {
        renderScreen();

        const header = await screen.findByRole('button', { name: /"copies":"3"/ });
        expect(header).toHaveTextContent('"size":"4.00 KB"');
        expect(header).toHaveTextContent('"reclaimable":"8.00 KB"');
        expect(header).toHaveAttribute('aria-expanded', 'false');

        fireEvent.click(header);

        expect(header).toHaveAttribute('aria-expanded', 'true');
        expect(screen.getByText('/Fotos/a/foto.jpg')).toBeInTheDocument();
        expect(screen.getByText('FILE_TIER_COLD')).toBeInTheDocument();
        expect(mockedGetGroupsPage).toHaveBeenCalledWith({
            page: 1,
            pageSize: 50,
            type: undefined,
        });
    });

    it('opens the copy and its folder in Files', async () => {
        renderScreen();
        await expandGroupWithCopies(3);

        expect(screen.getByRole('link', { name: '/Fotos/b/foto.jpg' })).toHaveAttribute(
            'href',
            '/files/Fotos/b/foto.jpg'
        );
        expect(
            screen.getByRole('link', {
                name: 'ANALYTICS_DUPLICATES_OPEN_FOLDER_OF {"path":"/Fotos/b/foto.jpg"}',
            })
        ).toHaveAttribute('href', '/files/Fotos/b');
    });

    it('shows the empty state', async () => {
        mockedGetGroupsPage.mockResolvedValue(onePage([]));

        renderScreen();

        expect(await screen.findByText('ANALYTICS_DUPLICATES_EMPTY')).toBeInTheDocument();
    });

    it('loads the next page while has_next is true', async () => {
        mockedGetGroupsPage
            .mockResolvedValueOnce(
                buildServerPage([firstGroup], { page: 1, pageSize: 50, hasNext: true })
            )
            .mockResolvedValueOnce(
                buildServerPage([secondGroup], { page: 2, pageSize: 50, hasNext: false })
            );

        renderScreen();
        fireEvent.click(await screen.findByRole('button', { name: 'LOAD_MORE' }));

        expect(await screen.findByRole('button', { name: /"copies":"2"/ })).toBeInTheDocument();
        expect(mockedGetGroupsPage).toHaveBeenLastCalledWith({
            page: 2,
            pageSize: 50,
            type: undefined,
        });
        expect(screen.getByRole('button', { name: /"copies":"3"/ })).toBeInTheDocument();
    });

    it('keeps the loaded groups when the next page fails and retries that page', async () => {
        mockedGetGroupsPage
            .mockResolvedValueOnce(
                buildServerPage([firstGroup], { page: 1, pageSize: 50, hasNext: true })
            )
            .mockRejectedValueOnce({ response: { data: { error: 'falha do servidor' } } })
            .mockResolvedValueOnce(
                buildServerPage([secondGroup], { page: 2, pageSize: 50, hasNext: false })
            );

        renderScreen();
        fireEvent.click(await screen.findByRole('button', { name: 'LOAD_MORE' }));

        expect(await screen.findByText('falha do servidor')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /"copies":"3"/ })).toBeInTheDocument();

        fireEvent.click(screen.getByRole('button', { name: 'TRY_AGAIN' }));

        expect(await screen.findByRole('button', { name: /"copies":"2"/ })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /"copies":"3"/ })).toBeInTheDocument();
    });

    it('reads the type filter from the URL and writes it back when changed', async () => {
        renderScreen('/analytics/duplicates?type=audio');

        await waitFor(() =>
            expect(mockedGetGroupsPage).toHaveBeenCalledWith({
                page: 1,
                pageSize: 50,
                type: 'audio',
            })
        );

        fireEvent.mouseDown(screen.getByRole('combobox'));
        fireEvent.click(
            await screen.findByRole('option', { name: 'ANALYTICS_DUPLICATE_TYPE_VIDEO' })
        );

        expect(screen.getByTestId('search')).toHaveTextContent('?type=video');
        await waitFor(() =>
            expect(mockedGetGroupsPage).toHaveBeenLastCalledWith({
                page: 1,
                pageSize: 50,
                type: 'video',
            })
        );

        fireEvent.mouseDown(screen.getByRole('combobox'));
        fireEvent.click(
            await screen.findByRole('option', { name: 'ANALYTICS_DUPLICATE_TYPE_ALL' })
        );

        expect(screen.getByTestId('search')).toBeEmptyDOMElement();
    });

    it('deletes one copy after a plain confirmation, shows the result and refetches the list', async () => {
        renderScreen();
        await expandGroupWithCopies(3);

        fireEvent.click(
            screen.getByRole('button', {
                name: 'ANALYTICS_DUPLICATES_DELETE_COPY {"path":"/Fotos/a/foto.jpg"}',
            })
        );
        expect(within(confirmDialog()).getByText('CONFIRM_DELETE')).toBeInTheDocument();
        expect(
            screen.queryByText(/ANALYTICS_DUPLICATES_DELETE_ALL_COPIES_WARNING/)
        ).not.toBeInTheDocument();

        fireEvent.click(within(confirmDialog()).getByRole('button', { name: 'DELETE' }));

        expect(
            await screen.findByText('ANALYTICS_DUPLICATES_RESULT_TRASHED {"count":"1"}')
        ).toBeInTheDocument();
        expect(mockedDeleteFile).toHaveBeenCalledTimes(1);
        expect(mockedDeleteFile).toHaveBeenCalledWith(1);
        await waitFor(() => expect(mockedGetGroupsPage).toHaveBeenCalledTimes(2));
    });

    it('deletes the selected copies in batch with a plain confirmation and lists each failure with its reason', async () => {
        mockedDeleteFile.mockImplementation((copyId: number) =>
            copyId === 2
                ? Promise.reject({ response: { data: { error: 'arquivo em uso' } } })
                : Promise.resolve()
        );
        renderScreen();
        await expandGroupWithCopies(3);

        fireEvent.click(
            screen.getByRole('checkbox', {
                name: 'ANALYTICS_DUPLICATES_SELECT_COPY {"path":"/Fotos/a/foto.jpg"}',
            })
        );
        fireEvent.click(
            screen.getByRole('checkbox', {
                name: 'ANALYTICS_DUPLICATES_SELECT_COPY {"path":"/Fotos/b/foto.jpg"}',
            })
        );
        expect(screen.getByText('FILES_SELECTION_COUNT {"count":"2"}')).toBeInTheDocument();

        fireEvent.click(
            within(screen.getByRole('toolbar')).getByRole('button', { name: 'DELETE' })
        );
        expect(
            within(confirmDialog()).getByText('FILES_CONFIRM_DELETE_MANY {"count":"2"}')
        ).toBeInTheDocument();
        fireEvent.click(within(confirmDialog()).getByRole('button', { name: 'DELETE' }));

        expect(
            await screen.findByText('ANALYTICS_DUPLICATES_RESULT_TRASHED {"count":"1"}')
        ).toBeInTheDocument();
        expect(
            screen.getByText('ANALYTICS_DUPLICATES_RESULT_FAILED {"count":"1"}')
        ).toBeInTheDocument();
        expect(
            screen.getByText(
                'ANALYTICS_DUPLICATES_RESULT_FAILURE_ITEM {"path":"/Fotos/b/foto.jpg","message":"arquivo em uso"}'
            )
        ).toBeInTheDocument();
        expect(screen.getByText('FILES_SELECTION_COUNT {"count":"1"}')).toBeInTheDocument();
    });

    it('asks for an extra confirmation when every copy of a group is selected', async () => {
        renderScreen();
        await expandGroupWithCopies(2);

        fireEvent.click(
            screen.getByRole('checkbox', {
                name: 'ANALYTICS_DUPLICATES_SELECT_COPY {"path":"/Docs/x.pdf"}',
            })
        );
        fireEvent.click(
            screen.getByRole('checkbox', {
                name: 'ANALYTICS_DUPLICATES_SELECT_COPY {"path":"/Docs/y.pdf"}',
            })
        );
        fireEvent.click(
            within(screen.getByRole('toolbar')).getByRole('button', { name: 'DELETE' })
        );
        fireEvent.click(within(confirmDialog()).getByRole('button', { name: 'DELETE' }));

        expect(
            within(confirmDialog()).getByText(
                'ANALYTICS_DUPLICATES_DELETE_ALL_COPIES_WARNING {"count":"1"}'
            )
        ).toBeInTheDocument();
        expect(mockedDeleteFile).not.toHaveBeenCalled();

        fireEvent.click(
            within(confirmDialog()).getByRole('button', {
                name: 'ANALYTICS_DUPLICATES_DELETE_ANYWAY',
            })
        );

        await waitFor(() => expect(mockedDeleteFile).toHaveBeenCalledTimes(2));
        expect(mockedDeleteFile).toHaveBeenCalledWith(4);
        expect(mockedDeleteFile).toHaveBeenCalledWith(5);
    });

    it('cancelling the confirmation deletes nothing', async () => {
        renderScreen();
        await expandGroupWithCopies(3);

        fireEvent.click(
            screen.getByRole('button', {
                name: 'ANALYTICS_DUPLICATES_DELETE_COPY {"path":"/Fotos/c/foto.jpg"}',
            })
        );
        fireEvent.click(within(confirmDialog()).getByRole('button', { name: 'ACTION_CANCEL' }));

        await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
        expect(mockedDeleteFile).not.toHaveBeenCalled();
    });
});
