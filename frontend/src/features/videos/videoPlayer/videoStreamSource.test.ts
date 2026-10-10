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
