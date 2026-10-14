import type { VideoFileDto } from '@/service/videoPlayback';
import { buildVideoTechnicalRows } from './videoTechnicalRows';

const bareVideo: VideoFileDto = {
    id: 1,
    name: 'a.mp4',
    path: '/a.mp4',
    parent_path: '/',
    format: 'mp4',
    size: 10,
};

const labelsOf = (rows: ReturnType<typeof buildVideoTechnicalRows>) =>
    rows.map((row) => row.labelKey);

describe('videoTechnicalRows', () => {
    it('returns no rows for a bare listing video without summary', () => {
        expect(buildVideoTechnicalRows(bareVideo, undefined)).toEqual([]);
    });

    it('falls back to the listing duration, resolution and codec', () => {
        const rows = buildVideoTechnicalRows(
            { ...bareVideo, duration_seconds: 125, width: 1920, height: 1080, video_codec: 'h264' },
            undefined
        );

        expect(rows).toEqual([
            { labelKey: 'FILE_METADATA_DURATION', value: '2:05' },
            { labelKey: 'FILE_METADATA_DIMENSIONS', value: '1920 × 1080', resolutionClass: 'fhd' },
            { labelKey: 'FILE_METADATA_VIDEO_CODEC', value: 'h264' },
        ]);
    });

    it('prefers the summary and omits empty or zero fields', () => {
        const rows = buildVideoTechnicalRows(
            { ...bareVideo, height: 480, width: 640, video_codec: 'mpeg4' },
            {
                duration: '60',
                width: 3840,
                height: 2160,
                codec_name: 'hevc',
                frame_rate: 29.97,
                bit_rate: '4500000',
                audio_codec: '',
                format_name: 'matroska,webm',
            }
        );

        expect(labelsOf(rows)).toEqual([
            'FILE_METADATA_DURATION',
            'FILE_METADATA_DIMENSIONS',
            'FILE_METADATA_VIDEO_CODEC',
            'FILE_METADATA_FRAME_RATE',
            'FILE_METADATA_BITRATE',
            'FILE_METADATA_CONTAINER',
        ]);
        expect(rows[1]).toEqual({
            labelKey: 'FILE_METADATA_DIMENSIONS',
            value: '3840 × 2160',
            resolutionClass: 'uhd',
        });
        expect(rows[2]?.value).toBe('hevc');
        expect(rows[4]?.value).toBe('4500 kbps');
    });
});
