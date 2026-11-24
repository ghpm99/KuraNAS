import {
    buildDirectVideoStreamUrl,
    buildRemuxVideoStreamUrl,
    buildTranscodeVideoStreamUrl,
    isContainerUnplayableByBrowser,
} from './videoStreamSource';

describe('videoPlayer/videoStreamSource', () => {
    it('builds the default stream and remux urls without any service mock', () => {
        expect(buildDirectVideoStreamUrl(7)).toMatch(/\/files\/video-stream\/7$/);
        expect(buildRemuxVideoStreamUrl(7, 0)).toMatch(/\/video\/stream\/7\/remux$/);
    });

    it('adds the start offset to the remux url', () => {
        expect(buildRemuxVideoStreamUrl(7, 12.5)).toMatch(
            /\/video\/stream\/7\/remux\?start=12\.500$/
        );
    });

    it('builds the transcode url with and without the start offset', () => {
        expect(buildTranscodeVideoStreamUrl(7, 0)).toMatch(/\/video\/stream\/7\/transcode$/);
        expect(buildTranscodeVideoStreamUrl(7, 12.5)).toMatch(
            /\/video\/stream\/7\/transcode\?start=12\.500$/
        );
    });

    it('adds the audio stream index alone or combined with the start offset', () => {
        expect(buildRemuxVideoStreamUrl(7, 0, 2)).toMatch(
            /\/video\/stream\/7\/remux\?audio_stream_index=2$/
        );
        expect(buildRemuxVideoStreamUrl(7, 12.5, 2)).toMatch(
            /\/video\/stream\/7\/remux\?start=12\.500&audio_stream_index=2$/
        );
        expect(buildTranscodeVideoStreamUrl(7, 12.5, 0)).toMatch(
            /\/video\/stream\/7\/transcode\?start=12\.500&audio_stream_index=0$/
        );
        expect(buildTranscodeVideoStreamUrl(7, 0, 3)).toMatch(
            /\/video\/stream\/7\/transcode\?audio_stream_index=3$/
        );
    });

    it('flags remux candidate containers the browser cannot play', () => {
        expect(isContainerUnplayableByBrowser('mkv', () => '')).toBe(true);
        expect(isContainerUnplayableByBrowser('.M2TS', () => '')).toBe(true);
    });

    it('does not flag playable, unknown or unverifiable containers', () => {
        expect(isContainerUnplayableByBrowser('mkv', () => 'maybe')).toBe(false);
        expect(isContainerUnplayableByBrowser('mp4', () => '')).toBe(false);
        expect(isContainerUnplayableByBrowser(undefined, () => '')).toBe(false);
        expect(isContainerUnplayableByBrowser('mkv', undefined)).toBe(false);
    });
});
