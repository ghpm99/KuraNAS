import type { VideoSubtitleDto } from '@/service/videoPlayback';
import { chooseInitialSubtitleId, getPrimaryLanguage } from './subtitleSelection';

const buildSubtitle = (overrides: Partial<VideoSubtitleDto>): VideoSubtitleDto => ({
    id: 1,
    label: 'sub',
    language: 'en',
    format: 'srt',
    is_forced: false,
    ...overrides,
});

describe('subtitleSelection', () => {
    it('extracts the lowercase primary language', () => {
        expect(getPrimaryLanguage('pt-BR')).toBe('pt');
        expect(getPrimaryLanguage('EN')).toBe('en');
        expect(getPrimaryLanguage('')).toBe('');
    });

    it('selects nothing by default', () => {
        const subtitles = [buildSubtitle({ id: 1 }), buildSubtitle({ id: 2, language: 'pt-BR' })];

        expect(chooseInitialSubtitleId(subtitles, null, 'pt-BR')).toBeNull();
    });

    it('turns on the forced subtitle in the interface language', () => {
        const subtitles = [
            buildSubtitle({ id: 1, is_forced: true }),
            buildSubtitle({ id: 2, language: 'pt', is_forced: true }),
        ];

        expect(chooseInitialSubtitleId(subtitles, null, 'pt-BR')).toBe(2);
    });

    it('ignores forced subtitles in other languages and blank interface language', () => {
        const subtitles = [buildSubtitle({ id: 1, language: '', is_forced: true })];

        expect(chooseInitialSubtitleId(subtitles, null, 'pt-BR')).toBeNull();
        expect(chooseInitialSubtitleId(subtitles, null, '')).toBeNull();
    });

    it('prefers the remembered language over the forced one, favoring non-forced tracks', () => {
        const subtitles = [
            buildSubtitle({ id: 1, language: 'pt', is_forced: true }),
            buildSubtitle({ id: 2, language: 'en', is_forced: true }),
            buildSubtitle({ id: 3, language: 'en' }),
        ];

        expect(chooseInitialSubtitleId(subtitles, 'en', 'pt-BR')).toBe(3);
    });

    it('falls back to a forced-only remembered language track', () => {
        const subtitles = [buildSubtitle({ id: 2, language: 'en', is_forced: true })];

        expect(chooseInitialSubtitleId(subtitles, 'en-US', 'fr')).toBe(2);
    });

    it('falls back to the forced interface subtitle when the remembered language is missing', () => {
        const subtitles = [buildSubtitle({ id: 1, language: 'pt', is_forced: true })];

        expect(chooseInitialSubtitleId(subtitles, 'es', 'pt-BR')).toBe(1);
    });
});
