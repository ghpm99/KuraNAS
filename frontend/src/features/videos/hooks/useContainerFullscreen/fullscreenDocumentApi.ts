export type FullscreenDocument = Document & {
    webkitFullscreenElement?: Element | null;
    webkitExitFullscreen?: () => Promise<void> | void;
};

export type FullscreenContainer = HTMLElement & {
    webkitRequestFullscreen?: () => Promise<void> | void;
};

export type FullscreenVideo = HTMLVideoElement & {
    webkitEnterFullscreen?: () => void;
    webkitExitFullscreen?: () => void;
    webkitDisplayingFullscreen?: boolean;
};

export const FULLSCREEN_CHANGE_EVENTS = ['fullscreenchange', 'webkitfullscreenchange'] as const;

export const readFullscreenElement = (fullscreenDocument: FullscreenDocument): Element | null =>
    fullscreenDocument.fullscreenElement ?? fullscreenDocument.webkitFullscreenElement ?? null;

export const exitDocumentFullscreen = (fullscreenDocument: FullscreenDocument) => {
    const exitRequest =
        fullscreenDocument.exitFullscreen?.() ?? fullscreenDocument.webkitExitFullscreen?.();
    void Promise.resolve(exitRequest).catch(() => undefined);
};

export const requestContainerFullscreen = (
    container: FullscreenContainer,
    video: FullscreenVideo | null
) => {
    if (container.requestFullscreen) {
        void container.requestFullscreen().catch(() => undefined);
        return;
    }
    if (container.webkitRequestFullscreen) {
        void Promise.resolve(container.webkitRequestFullscreen()).catch(() => undefined);
        return;
    }
    video?.webkitEnterFullscreen?.();
};
