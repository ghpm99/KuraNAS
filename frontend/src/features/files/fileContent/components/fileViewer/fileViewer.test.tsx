import { fireEvent, render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { ReactElement } from 'react';
import FileViewer from './fileViewer';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (k: string) => k,
    }),
}));

const renderWithQuery = (ui: ReactElement) =>
    render(
        <QueryClientProvider
            client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
        >
            {ui}
        </QueryClientProvider>
    );

describe('fileViewer', () => {
    const base = { id: 1, name: 'file', format: '.x', size: 10 } as any;

    it('mounts for a partial payload without any service mock', () => {
        const { container } = renderWithQuery(<FileViewer file={{ id: 3 } as any} />);

        expect(container).not.toBeEmptyDOMElement();
        expect(screen.getByRole('link', { name: 'DOWNLOAD' })).toBeInTheDocument();
    });

    it('renders image/audio/video/document branches', () => {
        const { rerender, container } = renderWithQuery(
            <FileViewer file={{ ...base, format: '.jpg' }} />
        );
        expect(screen.getByRole('img')).toHaveAttribute(
            'src',
            expect.stringContaining('/files/blob/1')
        );

        rerender(
            <QueryClientProvider client={new QueryClient()}>
                <FileViewer file={{ ...base, format: '.mp3' }} />
            </QueryClientProvider>
        );
        expect(screen.getByText('AUDIO_NOT_SUPPORTED')).toBeInTheDocument();
        expect(container.querySelector('audio')).not.toBeNull();

        rerender(
            <QueryClientProvider client={new QueryClient()}>
                <FileViewer file={{ ...base, format: '.mkv' }} />
            </QueryClientProvider>
        );
        expect(container.querySelector('video')).not.toBeNull();

        rerender(
            <QueryClientProvider client={new QueryClient()}>
                <FileViewer file={{ ...base, format: '.pdf', name: 'doc' }} />
            </QueryClientProvider>
        );
        expect(screen.getByTitle('doc')).toBeInTheDocument();
    });

    it('shows only the thumbnail for formats the browser cannot decode', () => {
        renderWithQuery(<FileViewer file={{ ...base, format: '.heic' }} />);

        expect(screen.getByRole('img')).toHaveAttribute(
            'src',
            expect.stringContaining('/files/thumbnail/1')
        );
    });

    it.each(['.zip', '.docx', '.xyz'])('falls back to a download card for %s', (format) => {
        renderWithQuery(<FileViewer file={{ ...base, format, name: 'thing' }} />);

        expect(screen.getByText('thing')).toBeInTheDocument();
        expect(screen.getByText('FILE_PREVIEW_UNAVAILABLE')).toBeInTheDocument();
        expect(screen.getByRole('link', { name: 'DOWNLOAD' })).toHaveAttribute(
            'href',
            expect.stringContaining('/files/download/1')
        );
    });

    it('swaps to the fallback card when the browser cannot play the media', () => {
        const { container } = renderWithQuery(
            <FileViewer file={{ ...base, format: '.mkv', name: 'movie.mkv' }} />
        );

        fireEvent.error(container.querySelector('video') as HTMLVideoElement);

        expect(container.querySelector('video')).toBeNull();
        expect(screen.getByText('FILE_PREVIEW_PLAYBACK_FAILED')).toBeInTheDocument();
        expect(screen.getByRole('link', { name: 'DOWNLOAD' })).toBeInTheDocument();
    });

    it('swaps audio to the fallback card on playback error', () => {
        const { container } = renderWithQuery(
            <FileViewer file={{ ...base, format: '.m4a', name: 'song.m4a' }} />
        );

        fireEvent.error(container.querySelector('audio') as HTMLAudioElement);

        expect(container.querySelector('audio')).toBeNull();
        expect(screen.getByText('FILE_PREVIEW_PLAYBACK_FAILED')).toBeInTheDocument();
    });

    it('delegates text and code files to the text viewer', () => {
        renderWithQuery(<FileViewer file={{ ...base, format: '.py', name: 'run.py' }} />);

        expect(screen.getByText('TEXT_VIEWER_LOADING')).toBeInTheDocument();
    });
});
