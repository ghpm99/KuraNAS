import { act, renderHook } from '@testing-library/react';
import useContainerFullscreen from './useContainerFullscreen';

type FakeDocument = Document & Record<string, any>;

const buildFullscreenDocument = (overrides: Record<string, unknown> = {}) => {
    const fullscreenDocument = document.implementation.createHTMLDocument('fullscreen');
    Object.assign(fullscreenDocument, overrides);
    return fullscreenDocument as FakeDocument;
};

const setFullscreenElement = (
    fullscreenDocument: FakeDocument,
    propertyName: string,
    element: Element | null
) => {
    Object.defineProperty(fullscreenDocument, propertyName, {
        value: element,
        configurable: true,
    });
};

const buildContainer = (overrides: Record<string, unknown> = {}) =>
    Object.assign(document.createElement('div'), overrides) as HTMLElement & Record<string, any>;

const buildVideoRef = (video: Partial<HTMLVideoElement> & Record<string, any> = {}) => ({
    current: Object.assign(document.createElement('video'), video),
});

describe('hooks/useContainerFullscreen', () => {
    it('mounts without a container and ignores the toggle', () => {
        const { result } = renderHook(() =>
            useContainerFullscreen({
                container: null,
                videoRef: { current: null },
                fullscreenDocument: buildFullscreenDocument(),
            })
        );

        expect(result.current.isFullscreen).toBe(false);
        act(() => result.current.toggleFullscreen());
        expect(result.current.isFullscreen).toBe(false);
    });

    it('requests fullscreen on the container, not on the video', () => {
        const fullscreenDocument = buildFullscreenDocument();
        const container = buildContainer({ requestFullscreen: jest.fn(() => Promise.resolve()) });
        const videoRef = buildVideoRef({ requestFullscreen: jest.fn() });
        const { result } = renderHook(() =>
            useContainerFullscreen({ container, videoRef, fullscreenDocument })
        );

        act(() => result.current.toggleFullscreen());

        expect(container.requestFullscreen).toHaveBeenCalledTimes(1);
        expect(videoRef.current.requestFullscreen).not.toHaveBeenCalled();
    });

    it('derives its state from fullscreenchange, including leaving with Escape', () => {
        const fullscreenDocument = buildFullscreenDocument();
        const container = buildContainer({ requestFullscreen: jest.fn(() => Promise.resolve()) });
        const { result } = renderHook(() =>
            useContainerFullscreen({
                container,
                videoRef: buildVideoRef(),
                fullscreenDocument,
            })
        );

        act(() => result.current.toggleFullscreen());
        expect(result.current.isFullscreen).toBe(false);

        setFullscreenElement(fullscreenDocument, 'fullscreenElement', container);
        act(() => {
            fullscreenDocument.dispatchEvent(new Event('fullscreenchange'));
        });
        expect(result.current.isFullscreen).toBe(true);

        setFullscreenElement(fullscreenDocument, 'fullscreenElement', null);
        act(() => {
            fullscreenDocument.dispatchEvent(new Event('fullscreenchange'));
        });
        expect(result.current.isFullscreen).toBe(false);
    });

    it('is not fullscreen when another element owns the fullscreen', () => {
        const fullscreenDocument = buildFullscreenDocument();
        const container = buildContainer();
        const { result } = renderHook(() =>
            useContainerFullscreen({ container, videoRef: buildVideoRef(), fullscreenDocument })
        );

        setFullscreenElement(fullscreenDocument, 'fullscreenElement', document.createElement('p'));
        act(() => {
            fullscreenDocument.dispatchEvent(new Event('fullscreenchange'));
        });

        expect(result.current.isFullscreen).toBe(false);
    });

    it('exits fullscreen through the document when something is fullscreen', () => {
        const exitFullscreen = jest.fn(() => Promise.resolve());
        const fullscreenDocument = buildFullscreenDocument({ exitFullscreen });
        const container = buildContainer({ requestFullscreen: jest.fn() });
        setFullscreenElement(fullscreenDocument, 'fullscreenElement', container);
        const { result } = renderHook(() =>
            useContainerFullscreen({ container, videoRef: buildVideoRef(), fullscreenDocument })
        );

        act(() => result.current.toggleFullscreen());

        expect(exitFullscreen).toHaveBeenCalledTimes(1);
        expect(container.requestFullscreen).not.toHaveBeenCalled();
    });

    it('uses the webkit prefixed API when the standard one is missing', () => {
        const webkitExitFullscreen = jest.fn();
        const fullscreenDocument = buildFullscreenDocument({ webkitExitFullscreen });
        const container = buildContainer({ webkitRequestFullscreen: jest.fn() });
        const { result } = renderHook(() =>
            useContainerFullscreen({ container, videoRef: buildVideoRef(), fullscreenDocument })
        );

        act(() => result.current.toggleFullscreen());
        expect(container.webkitRequestFullscreen).toHaveBeenCalledTimes(1);

        setFullscreenElement(fullscreenDocument, 'webkitFullscreenElement', container);
        act(() => {
            fullscreenDocument.dispatchEvent(new Event('webkitfullscreenchange'));
        });
        expect(result.current.isFullscreen).toBe(true);

        act(() => result.current.toggleFullscreen());
        expect(webkitExitFullscreen).toHaveBeenCalledTimes(1);
    });

    it('falls back to the video element on iOS Safari and tracks its native events', () => {
        const fullscreenDocument = buildFullscreenDocument();
        const container = buildContainer();
        const videoRef = buildVideoRef({ webkitEnterFullscreen: jest.fn() });
        const { result } = renderHook(() =>
            useContainerFullscreen({ container, videoRef, fullscreenDocument })
        );

        act(() => result.current.toggleFullscreen());
        expect(videoRef.current.webkitEnterFullscreen).toHaveBeenCalledTimes(1);

        act(() => {
            videoRef.current.dispatchEvent(new Event('webkitbeginfullscreen'));
        });
        expect(result.current.isFullscreen).toBe(true);

        act(() => {
            videoRef.current.dispatchEvent(new Event('webkitendfullscreen'));
        });
        expect(result.current.isFullscreen).toBe(false);
    });

    it('leaves the iOS native fullscreen through the video element', () => {
        const fullscreenDocument = buildFullscreenDocument();
        const videoRef = buildVideoRef({
            webkitDisplayingFullscreen: true,
            webkitExitFullscreen: jest.fn(),
        });
        const { result } = renderHook(() =>
            useContainerFullscreen({
                container: buildContainer(),
                videoRef,
                fullscreenDocument,
            })
        );

        act(() => result.current.toggleFullscreen());

        expect(videoRef.current.webkitExitFullscreen).toHaveBeenCalledTimes(1);
    });

    it('does nothing when no fullscreen API exists at all', () => {
        const { result } = renderHook(() =>
            useContainerFullscreen({
                container: buildContainer(),
                videoRef: { current: null },
                fullscreenDocument: buildFullscreenDocument(),
            })
        );

        expect(() => act(() => result.current.toggleFullscreen())).not.toThrow();
    });

    it('stops listening after unmount', () => {
        const fullscreenDocument = buildFullscreenDocument();
        const removeListener = jest.spyOn(fullscreenDocument, 'removeEventListener');
        const { unmount } = renderHook(() =>
            useContainerFullscreen({
                container: buildContainer(),
                videoRef: buildVideoRef(),
                fullscreenDocument,
            })
        );

        unmount();

        expect(removeListener).toHaveBeenCalledWith('fullscreenchange', expect.any(Function));
        expect(removeListener).toHaveBeenCalledWith('webkitfullscreenchange', expect.any(Function));
    });
});
