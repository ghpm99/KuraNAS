import type { VideoLibraryResolution } from '@/service/videoPlayback';

const hdMinimumHeight = 720;
const fullHdMinimumHeight = 1080;
const ultraHdMinimumHeight = 2160;

export const classifyVideoResolution = (
    height: number | null | undefined
): VideoLibraryResolution | null => {
    if (typeof height !== 'number' || !Number.isFinite(height) || height <= 0) {
        return null;
    }
    if (height >= ultraHdMinimumHeight) return 'uhd';
    if (height >= fullHdMinimumHeight) return 'fhd';
    if (height >= hdMinimumHeight) return 'hd';
    return 'sd';
};
