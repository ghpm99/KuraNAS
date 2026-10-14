export type PictureInPictureDocument = Document & {
    pictureInPictureEnabled?: boolean;
    pictureInPictureElement?: Element | null;
    exitPictureInPicture?: () => Promise<void>;
};

export type PictureInPictureVideo = HTMLVideoElement & {
    disablePictureInPicture?: boolean;
    requestPictureInPicture?: () => Promise<unknown>;
    webkitSupportsPresentationMode?: (mode: string) => boolean;
    webkitSetPresentationMode?: (mode: string) => void;
    webkitPresentationMode?: string;
};

export const PICTURE_IN_PICTURE_CHANGE_EVENTS = [
    'enterpictureinpicture',
    'leavepictureinpicture',
    'webkitpresentationmodechanged',
] as const;

const WEBKIT_PICTURE_IN_PICTURE_MODE = 'picture-in-picture';
const WEBKIT_INLINE_MODE = 'inline';

const isStandardPictureInPictureSupported = (
    pictureInPictureDocument: PictureInPictureDocument,
    video: PictureInPictureVideo
) =>
    pictureInPictureDocument.pictureInPictureEnabled === true &&
    video.disablePictureInPicture !== true &&
    typeof video.requestPictureInPicture === 'function';

const isWebkitPictureInPictureSupported = (video: PictureInPictureVideo) =>
    typeof video.webkitSetPresentationMode === 'function' &&
    video.webkitSupportsPresentationMode?.(WEBKIT_PICTURE_IN_PICTURE_MODE) === true;

export const isPictureInPictureSupported = (
    pictureInPictureDocument: PictureInPictureDocument,
    video: PictureInPictureVideo | null
) =>
    video !== null &&
    (isStandardPictureInPictureSupported(pictureInPictureDocument, video) ||
        isWebkitPictureInPictureSupported(video));

export const isVideoInPictureInPicture = (
    pictureInPictureDocument: PictureInPictureDocument,
    video: PictureInPictureVideo | null
) =>
    video !== null &&
    (pictureInPictureDocument.pictureInPictureElement === video ||
        video.webkitPresentationMode === WEBKIT_PICTURE_IN_PICTURE_MODE);

export const requestPictureInPicture = async (video: PictureInPictureVideo) => {
    if (typeof video.requestPictureInPicture === 'function') {
        await video.requestPictureInPicture();
        return;
    }
    video.webkitSetPresentationMode?.(WEBKIT_PICTURE_IN_PICTURE_MODE);
};

export const exitPictureInPicture = async (
    pictureInPictureDocument: PictureInPictureDocument,
    video: PictureInPictureVideo
) => {
    if (
        pictureInPictureDocument.pictureInPictureElement === video &&
        pictureInPictureDocument.exitPictureInPicture
    ) {
        await pictureInPictureDocument.exitPictureInPicture();
        return;
    }
    video.webkitSetPresentationMode?.(WEBKIT_INLINE_MODE);
};
