import { buildAudioTrackLabel } from './audioTrackLabel';

const translate = (key: string) => key;
const baseTrack = {
    stream_index: 1,
    language: 'ja',
    title: '',
    codec: 'aac',
    channels: 2,
    is_default: true,
};

describe('useVideoAudioTracks/audioTrackLabel', () => {
    it('names the language in the interface language and appends the title', () => {
        expect(buildAudioTrackLabel(baseTrack, 'pt-BR', translate)).toBe('japonês');
        expect(buildAudioTrackLabel({ ...baseTrack, title: 'Dub' }, 'en-US', translate)).toBe(
            'Japanese · Dub'
        );
    });

    it('falls back to the code when the language is not a valid tag', () => {
        expect(buildAudioTrackLabel({ ...baseTrack, language: '!!' }, 'en-US', translate)).toBe(
            '!!'
        );
    });

    it('labels a track without language with the translated placeholder', () => {
        expect(buildAudioTrackLabel({ ...baseTrack, language: '' }, 'en-US', translate)).toBe(
            'VIDEO_AUDIO_TRACK_UNKNOWN_LANGUAGE'
        );
    });
});
