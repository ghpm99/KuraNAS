import { act, renderHook } from '@testing-library/react';
import usePictureInPicture, {
    PICTURE_IN_PICTURE_ERROR_DURATION_MS,
    PICTURE_IN_PICTURE_ERROR_MESSAGE_KEY,
} from './usePictureInPicture';

type FakeDocument = Document & { pictureInPictureElement: any; exitPictureInPicture: any };
type FakeVideo = HTMLVideoElement & Record<string, any>;

const buildPictureInPictureDocument = (overrides: Record<string, unknown> = {}) =>
    Object.assign(document.implementation.createHTMLDocument('pip'), {
        pictureInPictureEnabled: true,
        pictureInPictureElement: null,
        exitPictureInPicture: jest.fn(() => Promise.resolve()),
        ...overrides,
    }) as FakeDocument;

const buildVideo = (overrides: Record<string, unknown> = {}) =>
    Object.assign(document.createElement('video'), {
        requestPictureInPicture: jest.fn(() => Promise.resolve({})),
        ...overrides,
    }) as FakeVideo;

const renderPictureInPicture = ({
    video = buildVideo(),
    pictureInPictureDocument = buildPictureInPictureDocument(),
    isFullscreen = false,
    exitFullscreen = jest.fn(),
} = {}) => {
    const rendered = renderHook(() =>
        usePictureInPicture({
            videoRef: { current: video },
            isFullscreen,
            exitFullscreen,
            pictureInPictureDocument,
        })
    );
    return { ...rendered, video, pictureInPictureDocument, exitFullscreen };
};

describe('hooks/usePictureInPicture', () => {
    it('mounts without a video and reports no support', async () => {
        const { result } = renderHook(() =>
            usePictureInPicture({
                videoRef: { current: null },
                isFullscreen: false,
                exitFullscreen: jest.fn(),
                pictureInPictureDocument: buildPictureInPictureDocument(),
            })
        );

        expect(result.current.isSupported).toBe(false);
        expect(result.current.isPictureInPicture).toBe(false);
        await act(() => result.current.togglePictureInPicture());
        expect(result.current.requestError).toBeNull();
    });

    it('is supported when the document enables it and the video allows it', () => {
        const { result } = renderPictureInPicture();

        expect(result.current.isSupported).toBe(true);
    });

    it.each([
        [
            'the document disables it',
            {
                video: buildVideo(),
                pictureInPictureDocument: buildPictureInPictureDocument({
                    pictureInPictureEnabled: false,
                }),
            },
        ],
        [
            'the video sets disablePictureInPicture',
            { video: buildVideo({ disablePictureInPicture: true }) },
        ],
        [
            'the video cannot request it',
            { video: buildVideo({ requestPictureInPicture: undefined }) },
        ],
    ])('is unsupported when %s', (_label, params) => {
        const { result } = renderPictureInPicture(params);

        expect(result.current.isSupported).toBe(false);
    });

    it('does nothing when toggled while unsupported', async () => {
        const video = buildVideo({ disablePictureInPicture: true });
        const { result } = renderPictureInPicture({ video });

        await act(() => result.current.togglePictureInPicture());

        expect(video.requestPictureInPicture).not.toHaveBeenCalled();
    });

    it('requests picture in picture on the video', async () => {
        const { result, video } = renderPictureInPicture();

        await act(() => result.current.togglePictureInPicture());

        expect(video.requestPictureInPicture).toHaveBeenCalledTimes(1);
    });

    it('derives its state from the enter and leave events', () => {
        const { result, video, pictureInPictureDocument } = renderPictureInPicture();

        pictureInPictureDocument.pictureInPictureElement = video;
        act(() => {
            video.dispatchEvent(new Event('enterpictureinpicture'));
        });
        expect(result.current.isPictureInPicture).toBe(true);

        pictureInPictureDocument.pictureInPictureElement = null;
        act(() => {
            video.dispatchEvent(new Event('leavepictureinpicture'));
        });
        expect(result.current.isPictureInPicture).toBe(false);
    });

    it('exits through the document when already in picture in picture', async () => {
        const { result, video, pictureInPictureDocument } = renderPictureInPicture();
        pictureInPictureDocument.pictureInPictureElement = video;

        await act(() => result.current.togglePictureInPicture());

        expect(pictureInPictureDocument.exitPictureInPicture).toHaveBeenCalledTimes(1);
        expect(video.requestPictureInPicture).not.toHaveBeenCalled();
    });

    it('leaves container fullscreen before entering', async () => {
        const callOrder: string[] = [];
        const exitFullscreen = jest.fn(() => callOrder.push('exitFullscreen'));
        const video = buildVideo({
            requestPictureInPicture: jest.fn(() => {
                callOrder.push('requestPictureInPicture');
                return Promise.resolve({});
            }),
        });
        const { result } = renderPictureInPicture({ video, isFullscreen: true, exitFullscreen });

        await act(() => result.current.togglePictureInPicture());

        expect(callOrder).toEqual(['exitFullscreen', 'requestPictureInPicture']);
    });

    it('does not touch fullscreen when not fullscreen', async () => {
        const { result, exitFullscreen } = renderPictureInPicture();

        await act(() => result.current.togglePictureInPicture());

        expect(exitFullscreen).not.toHaveBeenCalled();
    });

    describe('Safari presentation mode', () => {
        const buildWebkitVideo = (presentationMode = 'inline') =>
            buildVideo({
                requestPictureInPicture: undefined,
                webkitSupportsPresentationMode: jest.fn(
                    (mode: string) => mode === 'picture-in-picture'
                ),
                webkitSetPresentationMode: jest.fn(),
                webkitPresentationMode: presentationMode,
            });
        const webkitDocument = () =>
            buildPictureInPictureDocument({ pictureInPictureEnabled: undefined });

        it('is supported through webkitSupportsPresentationMode and enters via webkitSetPresentationMode', async () => {
            const video = buildWebkitVideo();
            const { result } = renderPictureInPicture({
                video,
                pictureInPictureDocument: webkitDocument(),
            });
            expect(result.current.isSupported).toBe(true);

            await act(() => result.current.togglePictureInPicture());

            expect(video.webkitSetPresentationMode).toHaveBeenCalledWith('picture-in-picture');
        });

        it('derives its state from webkitpresentationmodechanged and leaves to inline', async () => {
            const video = buildWebkitVideo();
            const { result } = renderPictureInPicture({
                video,
                pictureInPictureDocument: webkitDocument(),
            });

            video.webkitPresentationMode = 'picture-in-picture';
            act(() => {
                video.dispatchEvent(new Event('webkitpresentationmodechanged'));
            });
            expect(result.current.isPictureInPicture).toBe(true);

            await act(() => result.current.togglePictureInPicture());
            expect(video.webkitSetPresentationMode).toHaveBeenCalledWith('inline');
        });

        it('is unsupported when the presentation mode is not available', () => {
            const video = buildWebkitVideo();
            video.webkitSupportsPresentationMode = jest.fn(() => false);
            const { result } = renderPictureInPicture({
                video,
                pictureInPictureDocument: webkitDocument(),
            });

            expect(result.current.isSupported).toBe(false);
        });
    });

    describe('request errors', () => {
        beforeEach(() => {
            jest.useFakeTimers();
        });

        afterEach(() => {
            jest.useRealTimers();
        });

        it('reports a discreet error and clears it afterwards', async () => {
            const video = buildVideo({
                requestPictureInPicture: jest.fn(() =>
                    Promise.reject(new DOMException('blocked', 'NotAllowedError'))
                ),
            });
            const { result } = renderPictureInPicture({ video });

            await act(() => result.current.togglePictureInPicture());
            expect(result.current.requestError).toEqual({
                id: 1,
                messageKey: PICTURE_IN_PICTURE_ERROR_MESSAGE_KEY,
            });

            act(() => {
                jest.advanceTimersByTime(PICTURE_IN_PICTURE_ERROR_DURATION_MS);
            });
            expect(result.current.requestError).toBeNull();
        });
    });

    it('exits picture in picture when unmounted while active', () => {
        const { unmount, video, pictureInPictureDocument } = renderPictureInPicture();
        pictureInPictureDocument.pictureInPictureElement = video;

        unmount();

        expect(pictureInPictureDocument.exitPictureInPicture).toHaveBeenCalledTimes(1);
    });

    it('swallows a failing exit on unmount', async () => {
        const { unmount, video, pictureInPictureDocument } = renderPictureInPicture();
        pictureInPictureDocument.pictureInPictureElement = video;
        pictureInPictureDocument.exitPictureInPicture = jest.fn(() =>
            Promise.reject(new Error('x'))
        );

        expect(() => unmount()).not.toThrow();
        await Promise.resolve();
    });

    it('does not exit on unmount when not active', () => {
        const { unmount, pictureInPictureDocument } = renderPictureInPicture();

        unmount();

        expect(pictureInPictureDocument.exitPictureInPicture).not.toHaveBeenCalled();
    });

    it('removes its listeners on unmount', () => {
        const { unmount, video } = renderPictureInPicture();
        const removeSpy = jest.spyOn(video, 'removeEventListener');

        unmount();

        expect(removeSpy).toHaveBeenCalledWith('enterpictureinpicture', expect.any(Function));
        expect(removeSpy).toHaveBeenCalledWith('leavepictureinpicture', expect.any(Function));
        expect(removeSpy).toHaveBeenCalledWith(
            'webkitpresentationmodechanged',
            expect.any(Function)
        );
    });
});
