import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import FilePage from './index';

const mockFilterSync = jest.fn();

jest.mock('@/features/files/files/filesLayout', () => ({
    __esModule: true,
    default: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));
jest.mock('@/features/files/files/FilesExplorerScreen', () => ({
    __esModule: true,
    default: () => <div>explorer</div>,
}));
jest.mock('@/features/files/files/FileListFilterSync', () => ({
    __esModule: true,
    default: ({ filter }: { filter: string }) => {
        mockFilterSync(filter);
        return null;
    },
}));

const renderAt = (entry: string) =>
    render(
        <MemoryRouter initialEntries={[entry]}>
            <FilePage />
        </MemoryRouter>
    );

describe('FilePage', () => {
    beforeEach(() => jest.clearAllMocks());

    it('lists every file by default', () => {
        renderAt('/files');

        expect(screen.getByText('explorer')).toBeInTheDocument();
        expect(mockFilterSync).toHaveBeenLastCalledWith('all');
    });

    it('opens the recent listing when the URL asks for it', () => {
        renderAt('/files?filter=recent');

        expect(mockFilterSync).toHaveBeenLastCalledWith('recent');
    });
});
