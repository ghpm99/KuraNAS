import { act, renderHook } from '@testing-library/react';
import { useImageViewer } from './useImageViewer';

describe('useImageViewer', () => {
	beforeEach(() => {
		jest.useFakeTimers();
	});

	afterEach(() => {
		jest.clearAllTimers();
		jest.useRealTimers();
	});

	it('advances automatically while slideshow is active', () => {
		const { result } = renderHook(() =>
			useImageViewer([
				{ id: 1 },
				{ id: 2 },
				{ id: 3 },
			]),
		);

		act(() => {
			result.current.openImage(1);
			result.current.toggleSlideshow();
		});

		expect(result.current.activeImage?.id).toBe(1);
		expect(result.current.isSlideshowPlaying).toBe(true);

		act(() => {
			jest.advanceTimersByTime(3500);
		});

		expect(result.current.activeImage?.id).toBe(2);
	});

	it('toggles details and filmstrip with keyboard shortcuts and stops slideshow on close', () => {
		const { result } = renderHook(() =>
			useImageViewer([
				{ id: 1 },
				{ id: 2 },
			]),
		);

		act(() => {
			result.current.openImage(1);
			result.current.toggleSlideshow();
		});

		act(() => {
			window.dispatchEvent(new KeyboardEvent('keydown', { key: 'i' }));
			window.dispatchEvent(new KeyboardEvent('keydown', { key: 'f' }));
		});

		expect(result.current.showDetails).toBe(false);
		expect(result.current.showFilmstrip).toBe(false);

		act(() => {
			result.current.closeViewer();
		});

		expect(result.current.activeImage).toBeNull();
		expect(result.current.isSlideshowPlaying).toBe(false);
	});

	it('wraps to the previous image and clamps zoom boundaries', () => {
		const { result } = renderHook(() =>
			useImageViewer([
				{ id: 1 },
				{ id: 2 },
			]),
		);

		act(() => {
			result.current.openImage(1);
		});

		act(() => {
			result.current.goPrevious();
		});

		expect(result.current.activeImage?.id).toBe(2);

		act(() => {
			for (let step = 0; step < 30; step += 1) {
				result.current.increaseZoom();
			}
		});
		expect(result.current.zoom).toBe(5);

		act(() => {
			for (let step = 0; step < 40; step += 1) {
				result.current.decreaseZoom();
			}
		});
		expect(result.current.zoom).toBe(0.5);

		act(() => {
			result.current.resetZoom();
		});
		expect(result.current.zoom).toBe(1);
	});

	it('does not start slideshow when there is only one image', () => {
		const { result } = renderHook(() => useImageViewer([{ id: 1 }]));

		act(() => {
			result.current.openImage(1);
			result.current.toggleSlideshow();
			result.current.goNext();
			result.current.goPrevious();
		});

		expect(result.current.isSlideshowPlaying).toBe(false);
		expect(result.current.activeImage?.id).toBe(1);
	});
});
