import { useCallback, useEffect, useRef, useState, type RefObject } from 'react';
import {
    exitPictureInPicture,
    isPictureInPictureSupported,
    isVideoInPictureInPicture,
    PICTURE_IN_PICTURE_CHANGE_EVENTS,
    requestPictureInPicture,
    type PictureInPictureDocument,
    type PictureInPictureVideo,
} from './pictureInPictureDocumentApi';

export const PICTURE_IN_PICTURE_ERROR_MESSAGE_KEY = 'VIDEO_PICTURE_IN_PICTURE_ERROR';
export const PICTURE_IN_PICTURE_ERROR_DURATION_MS = 2500;

export interface PictureInPictureError {
    id: number;
    messageKey: string;
}

interface UsePictureInPictureParams {
    videoRef: RefObject<HTMLVideoElement | null>;
    isFullscreen: boolean;
    exitFullscreen: () => void;
    pictureInPictureDocument?: PictureInPictureDocument;
}

const usePictureInPicture = ({
    videoRef,
    isFullscreen,
    exitFullscreen,
    pictureInPictureDocument = document,
}: UsePictureInPictureParams) => {
    const [isSupported, setIsSupported] = useState(false);
    const [isPictureInPicture, setIsPictureInPicture] = useState(false);
    const [requestError, setRequestError] = useState<PictureInPictureError | null>(null);
    const errorCounterRef = useRef(0);

    useEffect(() => {
        const video = videoRef.current as PictureInPictureVideo | null;
        setIsSupported(isPictureInPictureSupported(pictureInPictureDocument, video));
        if (!video) return;

        const syncWithVideo = () =>
            setIsPictureInPicture(isVideoInPictureInPicture(pictureInPictureDocument, video));
        PICTURE_IN_PICTURE_CHANGE_EVENTS.forEach((eventName) =>
            video.addEventListener(eventName, syncWithVideo)
        );
        syncWithVideo();

        return () => {
            PICTURE_IN_PICTURE_CHANGE_EVENTS.forEach((eventName) =>
                video.removeEventListener(eventName, syncWithVideo)
            );
            if (isVideoInPictureInPicture(pictureInPictureDocument, video)) {
                void exitPictureInPicture(pictureInPictureDocument, video).catch(() => undefined);
            }
        };
    }, [pictureInPictureDocument, videoRef]);

    useEffect(() => {
        if (!requestError) return;
        const timeoutId = setTimeout(
            () => setRequestError(null),
            PICTURE_IN_PICTURE_ERROR_DURATION_MS
        );
        return () => clearTimeout(timeoutId);
    }, [requestError]);

    const reportRequestError = useCallback(() => {
        errorCounterRef.current += 1;
        setRequestError({
            id: errorCounterRef.current,
            messageKey: PICTURE_IN_PICTURE_ERROR_MESSAGE_KEY,
        });
    }, []);

    const togglePictureInPicture = useCallback(async () => {
        const video = videoRef.current as PictureInPictureVideo | null;
        if (!video || !isPictureInPictureSupported(pictureInPictureDocument, video)) return;

        try {
            if (isVideoInPictureInPicture(pictureInPictureDocument, video)) {
                await exitPictureInPicture(pictureInPictureDocument, video);
                return;
            }
            if (isFullscreen) exitFullscreen();
            await requestPictureInPicture(video);
        } catch {
            reportRequestError();
        }
    }, [exitFullscreen, isFullscreen, pictureInPictureDocument, reportRequestError, videoRef]);

    return { isSupported, isPictureInPicture, togglePictureInPicture, requestError };
};

export default usePictureInPicture;
