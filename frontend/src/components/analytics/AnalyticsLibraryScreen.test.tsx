import { fireEvent, render, screen } from '@testing-library/react';
import AnalyticsLibraryScreen from './AnalyticsLibraryScreen';
import type { AnalyticsScreenState } from './useAnalyticsScreenState';
import type { AnalyticsOverview } from '@/types/analytics';

const mockNavigate = jest.fn();

jest.mock('react-router-dom', () => ({
    useNavigate: () => mockNavigate,
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

const createOverview = (overrides?: Partial<AnalyticsOverview>): AnalyticsOverview => ({
    period: '7d',
    generated_at: '2026-03-16T11:50:00Z',
    storage: {
        total_bytes: 1000,
        used_bytes: 400,
        free_bytes: 600,
        growth_bytes: 50,
    },
    counts: {
        files_total: 200,
        files_added: 15,
        folders: 10,
    },
    time_series: [],
    types: [],
    extensions: [],
    hot_folders: [],
    top_folders: [],
    recent_files: [
        {
            id: 1,
            name: 'song.mp3',
            path: '/music/song.mp3',
            parent_path: '/music',
            format: 'audio/mp3',
            size_bytes: 5000,
            created_at: '2026-03-16T10:00:00Z',
            updated_at: '2026-03-16T10:00:00Z',
        },
    ],
    duplicates: {
        groups: 0,
        files: 0,
        reclaimable_size: 0,
        top_groups: [],
    },
    library: {
        categorized_media: 50,
        audio_with_metadata: 30,
        video_with_metadata: 15,
        image_with_metadata: 10,
        image_classified: 5,
    },
    processing: {
        metadata_pending: 3,
        metadata_failed: 1,
        thumbnail_pending: 2,
        thumbnail_failed: 4,
    },
    health: {
        status: 'ok',
        last_scan_at: '2026-03-16T11:40:00Z',
        last_scan_seconds: 12,
        indexed_files: 200,
        errors_last_24h: 0,
        recent_errors: [],
    },
    ...overrides,
});

const createState = (overrides: Partial<AnalyticsScreenState> = {}): AnalyticsScreenState =>
    ({
        t: (key: string) => key,
        period: '7d',
        setPeriod: jest.fn(),
        data: createOverview(),
        loading: false,
        error: '',
        refresh: jest.fn().mockResolvedValue(undefined),
        formatBytes: (n: number) => `${n} B`,
        formatPercent: (n: number) => `${n}%`,
        formatDate: (s: string) => s || '-',
        usedPercent: 40,
        reclaimablePercent: 10,
        healthStatusLabel: 'Healthy',
        healthStatusColor: 'success',
        processingFailureTotal: 5,
        updatedMinutes: '10',
        ...overrides,
    }) as unknown as AnalyticsScreenState;

describe('AnalyticsLibraryScreen', () => {
    beforeEach(() => {
        mockNavigate.mockReset();
    });

    it('renders all 6 KPI cards', () => {
        const state = createState();
        render(<AnalyticsLibraryScreen state={state} />);

        expect(screen.getAllByText('ANALYTICS_INDEXED_FILES').length).toBeGreaterThan(0);
        expect(screen.getByText('ANALYTICS_CATEGORIZED_MEDIA')).toBeInTheDocument();
        expect(screen.getAllByText('ANALYTICS_IMAGES_CLASSIFIED').length).toBeGreaterThan(0);
        expect(screen.getAllByText('ANALYTICS_METADATA_PENDING').length).toBeGreaterThan(0);
        expect(screen.getAllByText('ANALYTICS_THUMBNAIL_PENDING').length).toBeGreaterThan(0);
        expect(screen.getByText('ANALYTICS_PROCESSING_FAILURES')).toBeInTheDocument();
    });

    it('renders KPI values from data', () => {
        const state = createState();
        render(<AnalyticsLibraryScreen state={state} />);

        expect(screen.getByText('200')).toBeInTheDocument(); // indexed_files
        expect(screen.getByText('50')).toBeInTheDocument(); // categorized_media
    });

    it('renders processing failure total', () => {
        const state = createState({ processingFailureTotal: 99 });
        render(<AnalyticsLibraryScreen state={state} />);

        expect(screen.getByText('99')).toBeInTheDocument();
    });

    it('renders media coverage section', () => {
        const state = createState();
        render(<AnalyticsLibraryScreen state={state} />);

        expect(screen.getByText('ANALYTICS_MEDIA_COVERAGE')).toBeInTheDocument();
        expect(screen.getByText('NAV_MUSIC')).toBeInTheDocument();
        expect(screen.getByText('NAV_VIDEOS')).toBeInTheDocument();
        expect(screen.getByText('NAV_IMAGES')).toBeInTheDocument();
    });

    it('renders processing queue section', () => {
        const state = createState();
        render(<AnalyticsLibraryScreen state={state} />);

        expect(screen.getByText('ANALYTICS_PROCESSING_QUEUE')).toBeInTheDocument();
        expect(screen.getByText('ANALYTICS_METADATA_LABEL')).toBeInTheDocument();
        expect(screen.getByText('ANALYTICS_THUMBNAIL_LABEL')).toBeInTheDocument();
    });

    it('renders index health section with chip', () => {
        const state = createState();
        render(<AnalyticsLibraryScreen state={state} />);

        expect(screen.getByText('ANALYTICS_INDEX_HEALTH')).toBeInTheDocument();
        expect(screen.getByText('Healthy')).toBeInTheDocument();
    });

    it('renders no errors message when recent_errors is empty', () => {
        const state = createState();
        render(<AnalyticsLibraryScreen state={state} />);

        expect(screen.getByText('ANALYTICS_NO_ERRORS')).toBeInTheDocument();
    });

    it('renders recent errors when present', () => {
        const overview = createOverview({
            health: {
                status: 'error',
                last_scan_at: '2026-03-16T11:40:00Z',
                last_scan_seconds: 12,
                indexed_files: 200,
                errors_last_24h: 2,
                recent_errors: ['read error', 'parse error'],
            },
        });
        const state = createState({
            data: overview,
            healthStatusLabel: 'Error',
            healthStatusColor: 'error',
        });
        render(<AnalyticsLibraryScreen state={state} />);

        expect(screen.getByText('read error')).toBeInTheDocument();
        expect(screen.getByText('parse error')).toBeInTheDocument();
    });

    it('renders recent files section with clickable rows', () => {
        const state = createState();
        render(<AnalyticsLibraryScreen state={state} />);

        expect(screen.getByText('ANALYTICS_RECENT_FILES')).toBeInTheDocument();
        expect(screen.getByText('song.mp3')).toBeInTheDocument();

        const row = screen.getByText('song.mp3').closest('tr');
        fireEvent.click(row!);
        expect(mockNavigate).toHaveBeenCalled();
    });

    it('handles null data gracefully', () => {
        const state = createState({ data: null });
        render(<AnalyticsLibraryScreen state={state} />);

        expect(screen.getAllByText('ANALYTICS_INDEXED_FILES').length).toBeGreaterThan(0);
        expect(screen.getByText('ANALYTICS_CATEGORIZED_MEDIA')).toBeInTheDocument();
    });

    it('renders title and description', () => {
        const state = createState();
        render(<AnalyticsLibraryScreen state={state} />);

        expect(screen.getByText('ANALYTICS_LIBRARY_TITLE')).toBeInTheDocument();
        expect(screen.getByText('ANALYTICS_LIBRARY_DESCRIPTION')).toBeInTheDocument();
    });
});
