import {
    buildTranscodedStreamUrl,
    getAudioMimeType,
    isFormatPlayable,
    resolveTrackStreamSource,
} from './trackStreamSource';

const canPlayNothing = () => '';
const canPlayEverything = () => 'maybe';

describe('trackStreamSource', () => {
    it('resolves without throwing for a track with no format or metadata', () => {
        expect(() =>
            resolveTrackStreamSource({ id: 1, format: '' } as never, canPlayEverything)
        ).not.toThrow();
        expect(
            resolveTrackStreamSource({ id: 1, format: '' } as never, canPlayNothing).url
        ).toContain('/files/stream/1');
    });

    it('maps extensions to mime types ignoring dot and case', () => {
        expect(getAudioMimeType('.WMA')).toBe('audio/x-ms-wma');
        expect(getAudioMimeType('flac')).toBe('audio/flac');
        expect(getAudioMimeType('xyz')).toBeUndefined();
        expect(getAudioMimeType(undefined)).toBeUndefined();
    });

    it('uses the direct stream when the browser can play the format', () => {
        const source = resolveTrackStreamSource(
            { id: 7, format: 'mp3', metadata: undefined },
            canPlayEverything
        );
        expect(source.url).toMatch(/\/files\/stream\/7$/);
        expect(source.transcodedStream).toBeUndefined();
    });

    it('uses the transcode URL when canPlayType rejects the format', () => {
        const canPlayType = jest.fn(canPlayNothing);
        const source = resolveTrackStreamSource(
            { id: 7, format: '.wma', metadata: { length: 215 } as never },
            canPlayType
        );
        expect(canPlayType).toHaveBeenCalledWith('audio/x-ms-wma');
        expect(source.url).toMatch(/\/music\/tracks\/7\/stream\?format=mp3$/);
        expect(source.transcodedStream?.durationSeconds).toBe(215);
        expect(source.transcodedStream?.fallbackUrl).toMatch(/\/files\/stream\/7$/);
        expect(source.transcodedStream?.buildUrl(61.5)).toMatch(/format=mp3&start=61\.500$/);
    });

    it('plays unknown extensions directly', () => {
        expect(isFormatPlayable('weird', canPlayNothing)).toBe(true);
    });

    it('omits start for position zero', () => {
        expect(buildTranscodedStreamUrl(3, 0)).not.toContain('start=');
    });
});
