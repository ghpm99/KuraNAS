import { act, renderHook } from '@testing-library/react';
import { useImageViewer, type ImageViewerPaging } from './useImageViewer';

describe('useImageViewer', () => {
    beforeEach(() => {
        jest.useFakeTimers();
    });

    afterEach(() => {
        jest.clearAllTimers();
        jest.useRealTimers();
    });

    it('advances automatically while slideshow is active', () => {
        const { result } = renderHook(() => useImageViewer([{ id: 1 }, { id: 2 }, { id: 3 }]));

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
        const { result } = renderHook(() => useImageViewer([{ id: 1 }, { id: 2 }]));

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

    it('does not wrap around at either end of the list', () => {
        const { result } = renderHook(() => useImageViewer([{ id: 1 }, { id: 2 }]));

        act(() => {
            result.current.openImage(1);
        });
        expect(result.current.canGoPrevious).toBe(false);
        expect(result.current.canGoNext).toBe(true);

        act(() => {
            result.current.goPrevious();
        });
        expect(result.current.activeImage?.id).toBe(1);

        act(() => {
            result.current.goNext();
            result.current.goNext();
        });
        expect(result.current.activeImage?.id).toBe(2);
        expect(result.current.canGoNext).toBe(false);
    });

    it('clamps zoom boundaries', () => {
        const { result } = renderHook(() => useImageViewer([{ id: 1 }, { id: 2 }]));

        act(() => {
            result.current.openImage(1);
        });

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

describe('useImageViewer paging beyond the loaded images', () => {
    const buildPaging = (overrides: Partial<ImageViewerPaging<{ id: number }>> = {}) => ({
        hasNext: true,
        hasPrevious: false,
        isLoading: false,
        loadNext: jest.fn().mockResolvedValue(undefined),
        loadPrevious: jest.fn().mockResolvedValue(undefined),
        ...overrides,
    });

    const renderViewerAt = (
        images: { id: number }[],
        paging: ImageViewerPaging<{ id: number }>,
        openedId: number
    ) => {
        const hook = renderHook(() => useImageViewer(images, 3500, (image) => image.id, paging));
        act(() => {
            hook.result.current.openImage(openedId);
        });
        return hook;
    };

    it('fetches the next page at the end of the loaded images and continues into it', async () => {
        const loadNext = jest.fn().mockResolvedValue([{ id: 1 }, { id: 2 }, { id: 3 }, { id: 4 }]);
        const { result } = renderViewerAt([{ id: 1 }, { id: 2 }], buildPaging({ loadNext }), 2);
        expect(result.current.canGoNext).toBe(true);

        await act(async () => {
            result.current.goNext();
        });

        expect(loadNext).toHaveBeenCalledTimes(1);
        expect(result.current.viewerImageId).toBe(3);
    });

    it('does not call the loader when there are more loaded images ahead', () => {
        const paging = buildPaging();
        const { result } = renderViewerAt([{ id: 1 }, { id: 2 }], paging, 1);

        act(() => {
            result.current.goNext();
        });

        expect(paging.loadNext).not.toHaveBeenCalled();
        expect(result.current.viewerImageId).toBe(2);
    });

    it('stays put when the next page fails, is empty or the viewer already moved', async () => {
        const failing = buildPaging({ loadNext: jest.fn().mockResolvedValue(undefined) });
        const failedHook = renderViewerAt([{ id: 1 }], failing, 1);
        await act(async () => {
            failedHook.result.current.goNext();
        });
        expect(failedHook.result.current.viewerImageId).toBe(1);

        const empty = buildPaging({ loadNext: jest.fn().mockResolvedValue([{ id: 1 }]) });
        const emptyHook = renderViewerAt([{ id: 1 }], empty, 1);
        await act(async () => {
            emptyHook.result.current.goNext();
        });
        expect(emptyHook.result.current.viewerImageId).toBe(1);

        let finishLoading: (loaded: { id: number }[]) => void = () => undefined;
        const slow = buildPaging({
            loadNext: jest.fn().mockReturnValue(
                new Promise<{ id: number }[]>((resolve) => {
                    finishLoading = resolve;
                })
            ),
        });
        const slowHook = renderViewerAt([{ id: 1 }], slow, 1);
        await act(async () => {
            slowHook.result.current.goNext();
        });
        act(() => {
            slowHook.result.current.closeViewer();
        });
        await act(async () => {
            finishLoading([{ id: 1 }, { id: 2 }]);
        });
        expect(slowHook.result.current.viewerImageId).toBeNull();
    });

    it('ignores navigation while a page is already loading', async () => {
        const paging = buildPaging({ isLoading: true });
        const { result } = renderViewerAt([{ id: 1 }], paging, 1);

        await act(async () => {
            result.current.goNext();
        });

        expect(paging.loadNext).not.toHaveBeenCalled();
    });

    it('disables next at the true end of the list', () => {
        const paging = buildPaging({ hasNext: false });
        const { result } = renderViewerAt([{ id: 1 }, { id: 2 }], paging, 2);

        expect(result.current.canGoNext).toBe(false);
        act(() => {
            result.current.goNext();
        });
        expect(paging.loadNext).not.toHaveBeenCalled();
        expect(result.current.viewerImageId).toBe(2);
    });

    it('fetches previous images before the first loaded image', async () => {
        const loadPrevious = jest.fn().mockResolvedValue([{ id: 8 }, { id: 9 }, { id: 10 }]);
        const { result } = renderViewerAt(
            [{ id: 9 }, { id: 10 }],
            buildPaging({ hasNext: false, hasPrevious: true, loadPrevious }),
            9
        );
        expect(result.current.canGoPrevious).toBe(true);

        await act(async () => {
            result.current.goPrevious();
        });

        expect(loadPrevious).toHaveBeenCalledTimes(1);
        expect(result.current.viewerImageId).toBe(8);
    });

    it('stops the slideshow at the end of the list instead of wrapping', () => {
        jest.useFakeTimers();
        const { result } = renderViewerAt(
            [{ id: 1 }, { id: 2 }],
            buildPaging({ hasNext: false }),
            1
        );
        act(() => {
            result.current.toggleSlideshow();
        });

        act(() => {
            jest.advanceTimersByTime(3500);
        });
        expect(result.current.viewerImageId).toBe(2);

        act(() => {
            jest.advanceTimersByTime(3500);
        });
        expect(result.current.viewerImageId).toBe(2);
        expect(result.current.isSlideshowPlaying).toBe(false);
        jest.useRealTimers();
    });
});

describe('useImageViewer rotation, pan and keyboard', () => {
    const pressKey = (key: string, init: KeyboardEventInit = {}, target: EventTarget = window) =>
        act(() => {
            target.dispatchEvent(new KeyboardEvent('keydown', { key, bubbles: true, ...init }));
        });

    it('rotates in 90 degree steps, wraps after a full turn and resets for the next image', () => {
        const { result } = renderHook(() => useImageViewer([{ id: 1 }, { id: 2 }]));
        act(() => {
            result.current.openImage(1);
        });
        expect(result.current.rotation).toBe(0);

        act(() => {
            result.current.rotateClockwise();
        });
        expect(result.current.rotation).toBe(90);

        act(() => {
            result.current.rotateClockwise();
            result.current.rotateClockwise();
        });
        expect(result.current.rotation).toBe(270);

        act(() => {
            result.current.rotateClockwise();
        });
        expect(result.current.rotation).toBe(0);

        act(() => {
            result.current.rotateClockwise();
            result.current.goNext();
        });
        expect(result.current.rotation).toBe(0);

        act(() => {
            result.current.goPrevious();
        });
        expect(result.current.rotation).toBe(0);
    });

    it('rotates with the r shortcut', () => {
        const { result } = renderHook(() => useImageViewer([{ id: 1 }]));
        act(() => {
            result.current.openImage(1);
        });

        pressKey('r');
        expect(result.current.rotation).toBe(90);
        pressKey('R');
        expect(result.current.rotation).toBe(180);
    });

    it('pans only while zoomed and drops the pan when zoom returns to 1', () => {
        const { result } = renderHook(() => useImageViewer([{ id: 1 }]));
        act(() => {
            result.current.openImage(1);
            result.current.setPan(30, 40);
        });
        expect(result.current.pan).toEqual({ x: 0, y: 0 });

        act(() => {
            result.current.setZoomLevel(2);
        });
        act(() => {
            result.current.setPan(30, 40);
        });
        expect(result.current.zoom).toBe(2);
        expect(result.current.pan).toEqual({ x: 30, y: 40 });

        act(() => {
            result.current.setZoomLevel(1);
        });
        expect(result.current.pan).toEqual({ x: 0, y: 0 });

        act(() => {
            result.current.setZoomLevel(99);
        });
        expect(result.current.zoom).toBe(5);
    });

    it.each([
        ['input', () => document.createElement('input')],
        ['textarea', () => document.createElement('textarea')],
        ['select', () => document.createElement('select')],
        [
            'contenteditable',
            () => {
                const editable = document.createElement('div');
                editable.setAttribute('contenteditable', 'true');
                return editable;
            },
        ],
        [
            'child of a contenteditable',
            () => {
                const editable = document.createElement('div');
                editable.setAttribute('contenteditable', 'true');
                const child = document.createElement('span');
                editable.appendChild(child);
                document.body.appendChild(editable);
                return child;
            },
        ],
    ])('ignores every shortcut while typing in a %s', (_label, buildField) => {
        const { result } = renderHook(() => useImageViewer([{ id: 1 }, { id: 2 }]));
        act(() => {
            result.current.openImage(1);
        });
        const field = buildField();
        if (!field.isConnected) {
            document.body.appendChild(field);
        }

        ['i', 'f', 's', 'r', '+', '-', 'ArrowRight', 'Escape'].forEach((key) =>
            pressKey(key, {}, field)
        );

        expect(result.current.showDetails).toBe(true);
        expect(result.current.showFilmstrip).toBe(true);
        expect(result.current.isSlideshowPlaying).toBe(false);
        expect(result.current.rotation).toBe(0);
        expect(result.current.zoom).toBe(1);
        expect(result.current.viewerImageId).toBe(1);
        document.body.innerHTML = '';
    });

    it('ignores shortcuts while an input holds focus even if the event targets the window', () => {
        const { result } = renderHook(() => useImageViewer([{ id: 1 }]));
        act(() => {
            result.current.openImage(1);
        });
        const searchField = document.createElement('input');
        document.body.appendChild(searchField);
        searchField.focus();

        pressKey('i');
        expect(result.current.showDetails).toBe(true);

        searchField.blur();
        pressKey('i');
        expect(result.current.showDetails).toBe(false);
        document.body.innerHTML = '';
    });

    it('ignores shortcuts combined with ctrl, meta or alt', () => {
        const { result } = renderHook(() => useImageViewer([{ id: 1 }]));
        act(() => {
            result.current.openImage(1);
        });

        pressKey('r', { ctrlKey: true });
        pressKey('r', { metaKey: true });
        pressKey('r', { altKey: true });

        expect(result.current.rotation).toBe(0);
    });
});
