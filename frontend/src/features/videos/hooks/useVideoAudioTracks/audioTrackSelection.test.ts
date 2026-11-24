import type { VideoAudioTrackDto } from '@/service/videoPlayback';
import { findAudioStreamIndexToStartWith, findDefaultAudioTrack } from './audioTrackSelection';

const buildTrack = (overrides: Partial<VideoAudioTrackDto>): VideoAudioTrackDto => ({
    stream_index: 1,
    language: 'ja',
    title: '',
    codec: 'aac',
    channels: 2,
    is_default: false,
    ...overrides,
});

const tracks = [
    buildTrack({ stream_index: 1, language: 'ja', is_default: true }),
    buildTrack({ stream_index: 2, language: 'pt' }),
];

describe('useVideoAudioTracks/audioTrackSelection', () => {
    it('finds the default track or falls back to the first', () => {
        expect(findDefaultAudioTrack(tracks)?.stream_index).toBe(1);
        expect(findDefaultAudioTrack([buildTrack({ stream_index: 5 })])?.stream_index).toBe(5);
        expect(findDefaultAudioTrack([])).toBeUndefined();
    });

    it('picks the track in the remembered language when it is not the default', () => {
        expect(findAudioStreamIndexToStartWith(tracks, 'pt-BR')).toBe(2);
    });

    it('keeps the default when the remembered language already is the default', () => {
        expect(findAudioStreamIndexToStartWith(tracks, 'ja')).toBeNull();
    });

    it('keeps the default when no track matches, there is no memory or one track only', () => {
        expect(findAudioStreamIndexToStartWith(tracks, 'fr')).toBeNull();
        expect(findAudioStreamIndexToStartWith(tracks, null)).toBeNull();
        expect(findAudioStreamIndexToStartWith([tracks[1]!], 'pt')).toBeNull();
    });
});
