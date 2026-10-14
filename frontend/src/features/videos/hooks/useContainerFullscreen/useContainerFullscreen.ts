import { useCallback, useEffect, useState, type RefObject } from 'react';
import {
    exitDocumentFullscreen,
    FULLSCREEN_CHANGE_EVENTS,
    readFullscreenElement,
    requestContainerFullscreen,
    type FullscreenContainer,
    type FullscreenDocument,
    type FullscreenVideo,
} from './fullscreenDocumentApi';

interface UseContainerFullscreenParams {
    container: HTMLElement | null;
    videoRef: RefObject<HTMLVideoElement | null>;
    fullscreenDocument?: FullscreenDocument;
}

const useContainerFullscreen = ({
    container,
    videoRef,
    fullscreenDocument = document,
}: UseContainerFullscreenParams) => {
    const [isFullscreen, setIsFullscreen] = useState(false);

    useEffect(() => {
        if (!container) return;
        const video = videoRef.current as FullscreenVideo | null;
        const syncWithDocument = () =>
            setIsFullscreen(readFullscreenElement(fullscreenDocument) === container);
        const markVideoAsFullscreen = () => setIsFullscreen(true);
        const markVideoAsWindowed = () => setIsFullscreen(false);

        FULLSCREEN_CHANGE_EVENTS.forEach((eventName) =>
            fullscreenDocument.addEventListener(eventName, syncWithDocument)
        );
        video?.addEventListener('webkitbeginfullscreen', markVideoAsFullscreen);
        video?.addEventListener('webkitendfullscreen', markVideoAsWindowed);

        return () => {
            FULLSCREEN_CHANGE_EVENTS.forEach((eventName) =>
                fullscreenDocument.removeEventListener(eventName, syncWithDocument)
            );
            video?.removeEventListener('webkitbeginfullscreen', markVideoAsFullscreen);
            video?.removeEventListener('webkitendfullscreen', markVideoAsWindowed);
        };
    }, [container, fullscreenDocument, videoRef]);

    const toggleFullscreen = useCallback(() => {
        const video = videoRef.current as FullscreenVideo | null;
        if (readFullscreenElement(fullscreenDocument)) {
            exitDocumentFullscreen(fullscreenDocument);
            return;
        }
        if (video?.webkitDisplayingFullscreen) {
            video.webkitExitFullscreen?.();
            return;
        }
        if (!container) return;
        requestContainerFullscreen(container as FullscreenContainer, video);
    }, [container, fullscreenDocument, videoRef]);

    return { isFullscreen, toggleFullscreen };
};

export default useContainerFullscreen;
