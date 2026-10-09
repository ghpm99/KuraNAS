import {
    buildAudioRows,
    buildImageRows,
    buildVideoRows,
    formatClock,
} from './fileTypeMetadataRows';

describe('formatClock', () => {
    it('formats minutes and hours', () => {
        expect(formatClock(5)).toBe('0:05');
        expect(formatClock(185.4)).toBe('3:05');
        expect(formatClock(3725)).toBe('1:02:05');
    });
});

describe('buildImageRows', () => {
    it('skips empty and zero values', () => {
        expect(buildImageRows({} as any)).toEqual([]);
    });

    it('formats long exposures and joins camera make and model', () => {
        const rows = buildImageRows({
            make: 'Nikon',
            model: '',
            exposure_time: 2,
            f_number: 0,
        } as any);

        expect(rows).toEqual([
            { labelKey: 'FILE_METADATA_CAMERA', value: 'Nikon' },
            { labelKey: 'FILE_METADATA_EXPOSURE', value: '2 s' },
        ]);
    });
});

describe('buildAudioRows', () => {
    it('treats small bitrates as kbps and large ones as bps', () => {
        expect(buildAudioRows({ bitrate: 192 } as any)).toEqual([
            { labelKey: 'FILE_METADATA_BITRATE', value: '192 kbps' },
        ]);
        expect(buildAudioRows({ bitrate: 256000 } as any)).toEqual([
            { labelKey: 'FILE_METADATA_BITRATE', value: '256 kbps' },
        ]);
    });
});

describe('buildVideoRows', () => {
    it('ignores non numeric duration and bitrate text', () => {
        expect(
            buildVideoRows({ duration: 'N/A', bit_rate: '', codec_name: ' hevc ' } as any)
        ).toEqual([{ labelKey: 'FILE_METADATA_VIDEO_CODEC', value: 'hevc' }]);
    });
});
