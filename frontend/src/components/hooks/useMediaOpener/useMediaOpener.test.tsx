import { renderHook } from '@testing-library/react';
import useMediaOpener from './useMediaOpener';

const mockNavigate = jest.fn();
const mockReplaceQueue = jest.fn();
const mockUseLocation = jest.fn();

jest.mock('@/components/providers/GlobalMusicProvider', () => ({
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

		expect(result.current.openMediaItem({ id: 4, name: 'cover.jpg', format: '.jpg' })).toBe(true);
		expect(mockNavigate).toHaveBeenNthCalledWith(1, {
			pathname: '/images',
			search: '?image=4',
		}, {
			state: { from: '/files?filter=recent' },
		});

		expect(result.current.openMediaItem({ id: 9, name: 'episode.mp4', format: '.mp4' })).toBe(true);
		expect(mockNavigate).toHaveBeenNthCalledWith(2, '/video/9', {
			state: { from: '/files?filter=recent' },
		});
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
			}),
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
			}),
		);
		expect(mockNavigate).toHaveBeenCalledWith('/music', {
			state: { from: '/files?filter=recent' },
		});

		expect(result.current.openMediaItem({ id: 8, name: 'notes.pdf', format: '.pdf' })).toBe(false);
	});
});
