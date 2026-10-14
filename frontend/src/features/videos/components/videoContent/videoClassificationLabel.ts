const classificationLabelKeys: Record<string, string> = {
    series: 'VIDEO_CLASS_SERIES',
    movie: 'VIDEO_CLASS_MOVIE',
    anime: 'VIDEO_CLASS_ANIME',
    course: 'VIDEO_CLASS_COURSE',
    personal: 'VIDEO_CLASS_PERSONAL',
    clip: 'VIDEO_CLASS_CLIP',
    music_video: 'VIDEO_CLASS_MUSIC_VIDEO',
    program: 'VIDEO_CLASS_PROGRAM',
};

const fallbackClassificationLabelKey = 'NAV_VIDEOS';

export function getVideoClassificationLabelKey(classification?: string | null): string {
    if (!classification) return fallbackClassificationLabelKey;
    return classificationLabelKeys[classification.toLowerCase()] ?? fallbackClassificationLabelKey;
}
