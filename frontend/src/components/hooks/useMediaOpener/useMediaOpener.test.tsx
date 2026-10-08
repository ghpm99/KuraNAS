import { renderHook } from '@testing-library/react';
import useMediaOpener from './useMediaOpener';

const mockNavigate = jest.fn();
const mockReplaceQueue = jest.fn();
const mockUseLocation = jest.fn();

jest.mock('@/features/music/providers/GlobalMusicProvider', () => ({
    useGlobalMusic: () => ({
        replaceQueue: mockReplaceQueue,
    }),
}));

jest.mock('react-router-dom', () => ({
    useNavigate: () => mockNavigate,
    useLocation: () => mockUseLocation(),
}));

describe('components/hooks/useMediaOpener', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockUseLocation.mockReturnValue({
            pathname: '/files',
            search: '?filter=recent',
        });
    });

    it('routes image and video files to their dedicated viewers', () => {
        const { result } = renderHook(() => useMediaOpener());

        expect(
            result.current.openMediaItem({
                id: 4,
                name: 'cover.jpg',
                format: '.jpg',
            })
        ).toBe(true);
        expect(mockNavigate).toHaveBeenNthCalledWith(
            1,
            {
                pathname: '/images',
                search: '?image=4',
            },
            {
                state: { from: '/files?filter=recent' },
            }
        );

        expect(
            result.current.openMediaItem({
                id: 9,
                name: 'episode.mp4',
                format: '.mp4',
            })
        ).toBe(true);
        expect(mockNavigate).toHaveBeenNthCalledWith(2, '/video/9', {
            state: { from: '/files?filter=recent' },
        });
    });

    it('keeps image path in the route and ignores directories', () => {
        const { result } = renderHook(() => useMediaOpener());

        expect(
            result.current.openMediaItem({
                id: 10,
                name: 'detail.jpg',
                format: '.jpg',
                path: '/images/detail.jpg',
            })
        ).toBe(true);
        expect(mockNavigate).toHaveBeenCalledWith(
            {
                pathname: '/images',
                search: '?image=10&imagePath=%2Fimages%2Fdetail.jpg',
            },
            {
                state: { from: '/files?filter=recent' },
            }
        );

        expect(
            result.current.openMediaItem({
                id: 11,
                name: 'folder',
                format: '',
                type: 1,
            })
        ).toBe(false);
    });

    it('sends audio files to the global player and ignores unsupported formats', () => {
        const { result } = renderHook(() => useMediaOpener());

        expect(
            result.current.openMediaItem({
                id: 7,
                name: 'song.mp3',
                format: '.mp3',
                path: '/music/song.mp3',
                size: 1024,
            })
        ).toBe(true);
        expect(mockReplaceQueue).toHaveBeenCalledWith(
            [
                expect.objectContaining({
                    id: 7,
                    name: 'song.mp3',
                    path: '/music/song.mp3',
                    format: '.mp3',
                }),
            ],
            0,
            expect.objectContaining({
                href: '/files?filter=recent',
                labelKey: 'FILES',
            })
        );
        expect(mockNavigate).toHaveBeenCalledWith('/music', {
            state: { from: '/files?filter=recent' },
        });

        expect(
            result.current.openMediaItem({
                id: 8,
                name: 'notes.pdf',
                format: '.pdf',
            })
        ).toBe(false);
    });

    it('enqueues the listed audio files starting at the clicked one', () => {
        const { result } = renderHook(() => useMediaOpener());
        const listedFiles = [
            { id: 1, name: 'a.mp3', format: '.mp3', path: '/m/a.mp3' },
            { id: 2, name: 'cover.jpg', format: '.jpg', path: '/m/cover.jpg' },
            { id: 3, name: 'b.flac', format: '.flac', path: '/m/b.flac' },
            { id: 4, name: 'sub', format: '', type: 1, path: '/m/sub' },
            { id: 5, name: 'c.m4a', format: '.m4a', path: '/m/c.m4a' },
            { id: 6, name: 'd.wav', format: '.wav', path: '/m/d.wav' },
        ];

        expect(result.current.openMediaItem(listedFiles[2]!, listedFiles)).toBe(true);

        expect(mockReplaceQueue).toHaveBeenCalledWith(
            [
                expect.objectContaining({ id: 1 }),
                expect.objectContaining({ id: 3 }),
                expect.objectContaining({ id: 6 }),
            ],
            1,
            expect.objectContaining({ href: '/files?filter=recent' })
        );
        expect(mockNavigate).toHaveBeenCalledWith('/music', {
            state: { from: '/files?filter=recent' },
        });
    });

    it('falls back to a single-item queue when the clicked file is not in the listing', () => {
        const { result } = renderHook(() => useMediaOpener());

        result.current.openMediaItem({ id: 9, name: 'solo.mp3', format: '.mp3' }, [
            { id: 1, name: 'a.mp3', format: '.mp3' },
        ]);

        expect(mockReplaceQueue).toHaveBeenCalledWith(
            [expect.objectContaining({ id: 9 })],
            0,
            expect.anything()
        );
    });

    it('leaves formats without a dedicated screen to the in-place viewer', () => {
        const { result } = renderHook(() => useMediaOpener());

        for (const format of ['.m4a', '.opus', '.m4v', '.md']) {
            expect(result.current.openMediaItem({ id: 1, name: `f${format}`, format })).toBe(false);
        }
        expect(mockNavigate).not.toHaveBeenCalled();
        expect(mockReplaceQueue).not.toHaveBeenCalled();
    });

    it.each(['.heic', '.avif', '.tiff', '.jfif', '.cr2', '.nef', '.dng'])(
        'opens %s on the images screen',
        (format) => {
            const { result } = renderHook(() => useMediaOpener());

            expect(result.current.openMediaItem({ id: 5, name: `photo${format}`, format })).toBe(
                true
            );
            expect(mockNavigate).toHaveBeenCalledWith(
                expect.objectContaining({ search: '?image=5' }),
                expect.anything()
            );
        }
    );
});
