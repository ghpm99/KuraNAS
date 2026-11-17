import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import AnalyticsDuplicatesScreen from './AnalyticsDuplicatesScreen';
import { apiBase } from '@/service';
import { buildServerPage } from '@/shared/test/buildServerPage';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import { buildDuplicateCopy, buildDuplicateGroup } from './duplicateGroupFixtures';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn(), delete: jest.fn() },
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; delete: jest.Mock };

const duplicatesPage = buildServerPage(
    [
        buildDuplicateGroup('sig', [
            buildDuplicateCopy(11, { path: '/Fotos/um.jpg' }),
            buildDuplicateCopy(12, { path: '/Fotos/dois.jpg' }),
        ]),
    ],
    { page: 1, pageSize: 50, hasNext: false }
);

const duplicateGroupCalls = () =>
    mockedApi.get.mock.calls.filter(([url]) => url === '/analytics/duplicates/groups');

describe('duplicates deletion seam', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedApi.get.mockImplementation((url: string) =>
            url === '/analytics/duplicates/groups'
                ? Promise.resolve({ data: duplicatesPage })
                : Promise.resolve({ data: {} })
        );
        mockedApi.delete.mockResolvedValue({ data: undefined });
    });

    it('moves the selected copy to the trash with the exact request and refetches the duplicates', async () => {
        renderWithoutBackend(<AnalyticsDuplicatesScreen />, ['/analytics/duplicates']);

        fireEvent.click(await screen.findByRole('button', { expanded: false }));
        fireEvent.click(screen.getAllByRole('checkbox')[1]!);
        fireEvent.click(
            within(screen.getByRole('toolbar')).getByRole('button', { name: /^Excluir$|^DELETE$/ })
        );
        fireEvent.click(
            within(screen.getByRole('dialog')).getByRole('button', { name: /^Excluir$|^DELETE$/ })
        );

        await waitFor(() =>
            expect(mockedApi.delete).toHaveBeenCalledWith('/files/path', { data: { id: 12 } })
        );
        expect(mockedApi.delete).toHaveBeenCalledTimes(1);
        expect(duplicateGroupCalls()[0]?.[1]).toEqual({ params: { page: 1, page_size: 50 } });
        await waitFor(() => expect(duplicateGroupCalls().length).toBeGreaterThanOrEqual(2));
    });
});
