import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import AnalyticsContent from './AnalyticsContent';

const mockUseAnalyticsOverview = jest.fn();

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string, values?: Record<string, string>) =>
            values?.minutes ? `${key}:${values.minutes}` : key,
    }),
}));

jest.mock('@/components/providers/analyticsProvider/analyticsContext', () => ({
    useAnalyticsOverview: () => mockUseAnalyticsOverview(),
}));

jest.mock('@/components/hooks/useAnalyticsFormatters/useAnalyticsFormatters', () => ({
    useAnalyticsFormatters: () => ({
        formatBytes: (value: number) => `${value} B`,
        formatPercent: (value: number) => `${Math.round(value)}%`,
        formatDate: (value: string) => value || '-',
    }),
}));

jest.mock('@/components/hooks/useAnalyticsDerived/useAnalyticsDerived', () => ({
    useAnalyticsDerived: () => ({ usedPercent: 50, reclaimablePercent: 10 }),
}));

describe('AnalyticsContent', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockUseAnalyticsOverview.mockReturnValue({
            period: '7d',
            setPeriod: jest.fn(),
            loading: false,
            error: '',
            refresh: jest.fn(),
            data: {
                generated_at: '2026-01-01T00:00:00Z',
                storage: {
                    used_bytes: 1024,
                    total_bytes: 2048,
                    free_bytes: 1024,
                    growth_bytes: 128,
                },
                counts: { files_added: 2, files_total: 20, folders: 4 },
                time_series: [{ date: '2026-01-01', used_bytes: 1024 }],
                types: [{ type: 'video', count: 1, bytes: 1024 }],
                extensions: [{ ext: '.mp4', count: 1, bytes: 1024 }],
                hot_folders: [
                    {
                        path: '/media',
                        new_files: 2,
                        added_bytes: 10,
                        last_event: '2026-01-01T00:00:00Z',
                    },
                ],
                top_folders: [
                    {
                        path: '/media/videos',
                        files: 1,
                        bytes: 1024,
                        last_modified: '2026-01-01T00:00:00Z',
                    },
                ],
                recent_files: [
                    {
                        id: 1,
                        name: 'movie.mp4',
                        path: '/media/videos/movie.mp4',
                        parent_path: '/media/videos',
                        format: '.mp4',
                        size_bytes: 1024,
                        created_at: '2026-01-01T00:00:00Z',
                        updated_at: '2026-01-01T00:00:00Z',
                    },
                ],
                duplicates: {
                    groups: 1,
                    files: 2,
                    reclaimable_size: 256,
                    top_groups: [
                        {
                            signature: 'abcdef1234567890',
                            copies: 2,
                            size_bytes: 128,
                            reclaimable_size: 128,
                            paths: ['/media/videos/movie.mp4'],
                        },
                    ],
                },
                library: {
                    categorized_media: 8,
                    audio_with_metadata: 3,
                    video_with_metadata: 3,
                    image_with_metadata: 2,
                    image_classified: 2,
                },
                processing: {
                    metadata_pending: 1,
                    metadata_failed: 0,
                    thumbnail_pending: 2,
                    thumbnail_failed: 1,
                },
                health: {
                    status: 'ok',
                    last_scan_at: '2026-01-01T00:00:00Z',
                    last_scan_seconds: 60,
                    indexed_files: 20,
                    errors_last_24h: 0,
                    recent_errors: [],
                },
            },
        });
    });

    it('renders the overview section by default', () => {
        render(
            <MemoryRouter initialEntries={['/analytics']}>
                <AnalyticsContent />
            </MemoryRouter>
        );

        expect(screen.getByText('ANALYTICS_OVERVIEW_TITLE')).toBeInTheDocument();
        expect(screen.getByText('ANALYTICS_STORAGE_TREND')).toBeInTheDocument();
        expect(screen.getByText('ANALYTICS_SECTION_LIBRARY')).toBeInTheDocument();
    });

    it('renders the library and indexing section for the library route', () => {
        render(
            <MemoryRouter initialEntries={['/analytics/library']}>
                <AnalyticsContent />
            </MemoryRouter>
        );

        expect(screen.getByText('ANALYTICS_LIBRARY_TITLE')).toBeInTheDocument();
        expect(screen.getByText('ANALYTICS_PROCESSING_QUEUE')).toBeInTheDocument();
        expect(screen.getByText('ANALYTICS_CATEGORIZED_MEDIA')).toBeInTheDocument();
    });
});
