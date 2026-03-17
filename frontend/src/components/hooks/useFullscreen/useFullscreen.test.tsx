import { act, renderHook } from '@testing-library/react';
import { useFullscreen } from './useFullscreen';

describe('hooks/useFullscreen', () => {
    let fullscreenElementRef: Element | null;

    beforeEach(() => {
        fullscreenElementRef = null;
        Object.defineProperty(document, 'fullscreenElement', {
            configurable: true,
            get: () => fullscreenElementRef,
        });
    });

    it('enters fullscreen with standard API and updates state on event', () => {
        const requestFullscreen = jest.fn();
        Object.defineProperty(document.documentElement, 'requestFullscreen', {
            configurable: true,
            value: requestFullscreen,
        });

        const { result } = renderHook(() => useFullscreen());
        expect(result.current.isFullscreen).toBe(false);

        act(() => {
            result.current.enterFullscreen();
        });
        expect(requestFullscreen).toHaveBeenCalledTimes(1);

        act(() => {
            fullscreenElementRef = document.documentElement;
            document.dispatchEvent(new Event('fullscreenchange'));
        });

        expect(result.current.isFullscreen).toBe(true);
    });

    it('uses webkit fallback when standard enter api is unavailable', () => {
        Object.defineProperty(document.documentElement, 'requestFullscreen', {
            configurable: true,
            value: undefined,
        });
        const webkitRequestFullscreen = jest.fn();
        Object.defineProperty(document.documentElement, 'webkitRequestFullscreen', {
            configurable: true,
            value: webkitRequestFullscreen,
        });

        const { result } = renderHook(() => useFullscreen());
        act(() => {
            result.current.enterFullscreen();
        });

        expect(webkitRequestFullscreen).toHaveBeenCalledTimes(1);
    });

    it('uses moz and ms enter fallbacks', () => {
        Object.defineProperty(document.documentElement, 'requestFullscreen', {
            configurable: true,
            value: undefined,
        });
        Object.defineProperty(document.documentElement, 'webkitRequestFullscreen', {
            configurable: true,
            value: undefined,
        });

        const mozRequestFullScreen = jest.fn();
        Object.defineProperty(document.documentElement, 'mozRequestFullScreen', {
            configurable: true,
            value: mozRequestFullScreen,
        });

        const { result } = renderHook(() => useFullscreen());
        act(() => {
            result.current.enterFullscreen();
        });
        expect(mozRequestFullScreen).toHaveBeenCalledTimes(1);

        Object.defineProperty(document.documentElement, 'mozRequestFullScreen', {
            configurable: true,
            value: undefined,
        });
        const msRequestFullscreen = jest.fn();
        Object.defineProperty(document.documentElement, 'msRequestFullscreen', {
            configurable: true,
            value: msRequestFullscreen,
        });
        act(() => {
            result.current.enterFullscreen();
        });
        expect(msRequestFullscreen).toHaveBeenCalledTimes(1);
    });

    it('exits fullscreen and uses vendor fallbacks', () => {
        fullscreenElementRef = document.documentElement;
        const exitFullscreen = jest.fn();
        Object.defineProperty(document, 'exitFullscreen', {
            configurable: true,
            value: exitFullscreen,
        });

        const { result } = renderHook(() => useFullscreen());
        act(() => {
            result.current.toggleFullscreen();
        });
        expect(exitFullscreen).toHaveBeenCalledTimes(1);

        Object.defineProperty(document, 'exitFullscreen', {
            configurable: true,
            value: undefined,
        });
        const webkitExitFullscreen = jest.fn();
        Object.defineProperty(document, 'webkitExitFullscreen', {
            configurable: true,
            value: webkitExitFullscreen,
        });

        act(() => {
            result.current.exitFullscreen();
        });
        expect(webkitExitFullscreen).toHaveBeenCalledTimes(1);

        Object.defineProperty(document, 'webkitExitFullscreen', {
            configurable: true,
            value: undefined,
        });
        const mozCancelFullScreen = jest.fn();
        Object.defineProperty(document, 'mozCancelFullScreen', {
            configurable: true,
            value: mozCancelFullScreen,
        });
        act(() => {
            result.current.exitFullscreen();
        });
        expect(mozCancelFullScreen).toHaveBeenCalledTimes(1);

        Object.defineProperty(document, 'mozCancelFullScreen', {
            configurable: true,
            value: undefined,
        });
        const msExitFullscreen = jest.fn();
        Object.defineProperty(document, 'msExitFullscreen', {
            configurable: true,
            value: msExitFullscreen,
        });
        act(() => {
            result.current.exitFullscreen();
        });
        expect(msExitFullscreen).toHaveBeenCalledTimes(1);
    });

    it('toggle enters fullscreen when currently not fullscreen', () => {
        const requestFullscreen = jest.fn();
        Object.defineProperty(document.documentElement, 'requestFullscreen', {
            configurable: true,
            value: requestFullscreen,
        });

        const { result } = renderHook(() => useFullscreen());
        act(() => {
            result.current.toggleFullscreen();
        });
        expect(requestFullscreen).toHaveBeenCalledTimes(1);
    });

    it('logs fullscreen errors', () => {
        const errorSpy = jest.spyOn(console, 'error').mockImplementation(() => {});
        renderHook(() => useFullscreen());

        act(() => {
            document.dispatchEvent(new Event('fullscreenerror'));
        });

        expect(errorSpy).toHaveBeenCalled();
        errorSpy.mockRestore();
    });
});
