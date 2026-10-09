import { fireEvent, render as rtlRender, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { ReactElement } from 'react';
import FileDetails from './fileDetails';

const mockUseFile = jest.fn();
const mockEnqueueSnackbar = jest.fn();

const render = (ui: ReactElement) =>
    rtlRender(
        <QueryClientProvider
            client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
        >
            {ui}
        </QueryClientProvider>
    );

jest.mock('@/features/files/providers/fileProvider/fileContext', () => ({
    __esModule: true,
    default: () => mockUseFile(),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (k: string) => k }),
}));

jest.mock('notistack', () => ({
    useSnackbar: () => ({ enqueueSnackbar: mockEnqueueSnackbar }),
}));

const baseFile = {
    id: 2,
    name: 'a.mp3',
    type: 2,
    format: '.mp3',
    size: 1024,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-02T00:00:00Z',
    path: '/music/a.mp3',
} as any;

const baseContext = {
    isLoadingAccessData: false,
    recentAccessFiles: [],
    handleSelectItem: jest.fn(),
};

describe('fileDetails', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockUseFile.mockReturnValue(baseContext);
    });

    it('mounts for a partial payload without any service mock', () => {
        mockUseFile.mockReturnValue({});

        render(<FileDetails file={{ id: 1 } as any} onClose={jest.fn()} />);

        expect(screen.getByText('FILE_DETAILS_TITLE')).toBeInTheDocument();
        expect(screen.getByText('FILE_DETAILS_CHECKSUM_CALCULATING')).toBeInTheDocument();
    });

    it('renders file details and recent access list', () => {
        mockUseFile.mockReturnValue({
            ...baseContext,
            recentAccessFiles: [
                {
                    id: 10,
                    ip_address: '127.0.0.1',
                    file_id: 2,
                    accessed_at: '2026-01-03T00:00:00Z',
                },
                {
                    id: 11,
                    ip_address: '10.0.0.9',
                    file_id: 77,
                    accessed_at: '2026-01-03T00:00:00Z',
                },
            ],
        });

        render(<FileDetails file={baseFile} onClose={jest.fn()} />);

        expect(screen.getByText('FILE_DETAILS_TITLE')).toBeInTheDocument();
        expect(screen.getByText('/music/a.mp3')).toBeInTheDocument();
        expect(screen.getByText('AUDIO_MP3')).toBeInTheDocument();
        expect(screen.getByText('127.0.0.1')).toBeInTheDocument();
        expect(screen.queryByText('10.0.0.9')).toBeNull();
    });

    it('calls onClose from the close button', () => {
        const onClose = jest.fn();
        render(<FileDetails file={baseFile} onClose={onClose} />);

        fireEvent.click(screen.getByRole('button', { name: 'CLOSE' }));

        expect(onClose).toHaveBeenCalledTimes(1);
    });

    it('shows the checksum in monospace with a copy button', async () => {
        const writeText = jest.fn().mockResolvedValue(undefined);
        Object.assign(navigator, { clipboard: { writeText } });

        render(<FileDetails file={{ ...baseFile, check_sum: 'abc123def' }} onClose={jest.fn()} />);

        const checksum = screen.getByText('abc123def');
        expect(checksum.tagName).toBe('CODE');
        expect(getComputedStyle(checksum).fontFamily).toContain('monospace');
        expect(screen.queryByText('FILE_DETAILS_CHECKSUM_CALCULATING')).toBeNull();

        fireEvent.click(screen.getByRole('button', { name: 'FILE_DETAILS_CHECKSUM_COPY' }));

        expect(writeText).toHaveBeenCalledWith('abc123def');
        await screen.findByRole('button', { name: 'FILE_DETAILS_CHECKSUM_COPY' });
        expect(mockEnqueueSnackbar).toHaveBeenCalledWith('FILE_DETAILS_CHECKSUM_COPIED', {
            variant: 'success',
        });
    });

    it('reports a failed checksum copy', async () => {
        Object.assign(navigator, {
            clipboard: { writeText: jest.fn().mockRejectedValue(new Error('denied')) },
        });
        render(<FileDetails file={{ ...baseFile, check_sum: 'abc' }} onClose={jest.fn()} />);

        fireEvent.click(screen.getByRole('button', { name: 'FILE_DETAILS_CHECKSUM_COPY' }));

        await screen.findByRole('button', { name: 'FILE_DETAILS_CHECKSUM_COPY' });
        await Promise.resolve();
        expect(mockEnqueueSnackbar).toHaveBeenCalledWith('FILE_DETAILS_CHECKSUM_COPY_FAILED', {
            variant: 'error',
        });
    });

    it('shows calculating and no copy button while the checksum is empty', () => {
        render(<FileDetails file={{ ...baseFile, check_sum: '' }} onClose={jest.fn()} />);

        expect(screen.getByText('FILE_DETAILS_CHECKSUM_CALCULATING')).toBeInTheDocument();
        expect(screen.queryByRole('button', { name: 'FILE_DETAILS_CHECKSUM_COPY' })).toBeNull();
    });

    it('shows last interaction and last backup from the backend optional wrapper', () => {
        render(
            <FileDetails
                file={{
                    ...baseFile,
                    last_interaction: { Value: '2026-02-01T10:00:00Z', HasValue: true },
                    last_backup: { Value: '0001-01-01T00:00:00Z', HasValue: false },
                }}
                onClose={jest.fn()}
            />
        );

        expect(screen.getByText('FILE_DETAILS_LAST_INTERACTION')).toBeInTheDocument();
        expect(screen.getByText('FILE_DETAILS_LAST_BACKUP')).toBeInTheDocument();
        expect(screen.getAllByText('FILE_DETAILS_NEVER')).toHaveLength(1);
    });

    it('shows never for missing interaction and backup', () => {
        render(<FileDetails file={baseFile} onClose={jest.fn()} />);

        expect(screen.getAllByText('FILE_DETAILS_NEVER')).toHaveLength(2);
    });

    it('shows the cold-tier badge for a migrated file', () => {
        render(<FileDetails file={{ ...baseFile, tier: 'cold' }} onClose={jest.fn()} />);

        expect(screen.getByText('FILE_TIER_COLD')).toBeInTheDocument();
    });

    it('shows the hot-tier badge for a hot file and no badge without a tier', () => {
        const hot = render(<FileDetails file={{ ...baseFile, tier: 'hot' }} onClose={jest.fn()} />);
        expect(screen.getByText('FILE_TIER_HOT')).toBeInTheDocument();
        expect(screen.queryByText('FILE_TIER_COLD')).toBeNull();
        hot.unmount();

        render(<FileDetails file={baseFile} onClose={jest.fn()} />);
        expect(screen.queryByText('FILE_TIER_HOT')).toBeNull();
        expect(screen.queryByText('FILE_TIER_COLD')).toBeNull();
    });

    it('renders loading spinner for recent activity', () => {
        mockUseFile.mockReturnValue({ ...baseContext, isLoadingAccessData: true });

        render(<FileDetails file={baseFile} onClose={jest.fn()} />);

        expect(screen.getByRole('progressbar')).toBeInTheDocument();
    });

    it('renders a folder with name, path and dates but no file-only fields', () => {
        render(
            <FileDetails
                file={
                    {
                        id: 5,
                        name: 'photos',
                        type: 1,
                        format: '',
                        size: 0,
                        created_at: '2026-01-01T00:00:00Z',
                        updated_at: '2026-01-02T00:00:00Z',
                        path: '/media/photos',
                    } as any
                }
                onClose={jest.fn()}
            />
        );

        expect(screen.getByText('photos')).toBeInTheDocument();
        expect(screen.getByText('/media/photos')).toBeInTheDocument();
        expect(screen.getByText('FOLDER')).toBeInTheDocument();
        expect(screen.queryByText('FILE_DETAILS_CHECKSUM')).toBeNull();
        expect(screen.queryByText('SIZE')).toBeNull();
        expect(screen.queryByText('RECENT_ACTIVITY')).toBeNull();
        expect(
            screen.getByText('FOLDER_STATS_CALCULATING', { selector: 'span' })
        ).toBeInTheDocument();
    });
});
