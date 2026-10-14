import type { VideoSubtitleDto } from '@/service/videoPlayback';
import {
    chooseInitialSubtitleKey,
    findRememberedLanguageSubtitleKey,
    getPrimaryLanguage,
} from './subtitleSelection';
import type { VideoSubtitlePreferences } from './videoSubtitlePreferences';

const buildSubtitle = (overrides: Partial<VideoSubtitleDto>): VideoSubtitleDto => ({
    id: 1,
    label: 'sub',
    language: 'en',
    format: 'srt',
    is_forced: false,
    ...overrides,
});

const enabledIn = (language: string): VideoSubtitlePreferences => ({ isEnabled: true, language });
const disabledWith = (language: string): VideoSubtitlePreferences => ({
    isEnabled: false,
    language,
});

describe('subtitleSelection', () => {
    it('extracts the lowercase primary language', () => {
        expect(getPrimaryLanguage('pt-BR')).toBe('pt');
        expect(getPrimaryLanguage('EN')).toBe('en');
        expect(getPrimaryLanguage('')).toBe('');
    });

    it('selects nothing by default', () => {
        const subtitles = [buildSubtitle({ id: 1 }), buildSubtitle({ id: 2, language: 'pt-BR' })];

        expect(chooseInitialSubtitleKey(subtitles, null, 'pt-BR')).toBeNull();
    });

    it('turns on the forced subtitle in the interface language', () => {
        const subtitles = [
            buildSubtitle({ id: 1, is_forced: true }),
            buildSubtitle({ id: 2, language: 'pt', is_forced: true }),
        ];

        expect(chooseInitialSubtitleKey(subtitles, null, 'pt-BR')).toBe('external:2');
    });

    it('ignores forced subtitles in other languages and blank interface language', () => {
        const subtitles = [buildSubtitle({ id: 1, language: '', is_forced: true })];

        expect(chooseInitialSubtitleKey(subtitles, null, 'pt-BR')).toBeNull();
        expect(chooseInitialSubtitleKey(subtitles, null, '')).toBeNull();
    });

    it('prefers the remembered language over the forced one, favoring non-forced tracks', () => {
        const subtitles = [
            buildSubtitle({ id: 1, language: 'pt', is_forced: true }),
            buildSubtitle({ id: 2, language: 'en', is_forced: true }),
            buildSubtitle({ id: 3, language: 'en' }),
        ];

        expect(chooseInitialSubtitleKey(subtitles, enabledIn('en'), 'pt-BR')).toBe('external:3');
    });

    it('falls back to a forced-only remembered language track', () => {
        const subtitles = [buildSubtitle({ id: 2, language: 'en', is_forced: true })];

        expect(chooseInitialSubtitleKey(subtitles, enabledIn('en-US'), 'fr')).toBe('external:2');
    });

    it('falls back to the forced interface subtitle when the remembered language is missing', () => {
        const subtitles = [buildSubtitle({ id: 1, language: 'pt', is_forced: true })];

        expect(chooseInitialSubtitleKey(subtitles, enabledIn('es'), 'pt-BR')).toBe('external:1');
    });

    it('does not turn on a regular subtitle when disabled', () => {
        const subtitles = [buildSubtitle({ id: 1, language: 'en' })];

        expect(chooseInitialSubtitleKey(subtitles, disabledWith('en'), 'pt-BR')).toBeNull();
    });

    it('still turns on the forced interface subtitle when disabled', () => {
        const subtitles = [
            buildSubtitle({ id: 1, language: 'en' }),
            buildSubtitle({ id: 2, language: 'pt', is_forced: true }),
        ];

        expect(chooseInitialSubtitleKey(subtitles, disabledWith('en'), 'pt-BR')).toBe('external:2');
    });

    it('finds the remembered language subtitle regardless of enabled state', () => {
        const subtitles = [buildSubtitle({ id: 1, language: 'pt' }), buildSubtitle({ id: 2 })];

        expect(findRememberedLanguageSubtitleKey(subtitles, disabledWith('en'))).toBe('external:2');
        expect(findRememberedLanguageSubtitleKey(subtitles, disabledWith(''))).toBeNull();
        expect(findRememberedLanguageSubtitleKey(subtitles, disabledWith('es'))).toBeNull();
        expect(findRememberedLanguageSubtitleKey(subtitles, null)).toBeNull();
    });

    it('picks embedded subtitles by key and prefers non-forced ones', () => {
        const subtitles = [
            buildSubtitle({ id: 0, source: 'embedded', key: '5:1', is_forced: true }),
            buildSubtitle({ id: 0, source: 'embedded', key: '5:2' }),
        ];

        expect(chooseInitialSubtitleKey(subtitles, enabledIn('en'), 'pt-BR')).toBe('5:2');
    });

    it('turns on an embedded forced subtitle in the interface language', () => {
        const subtitles = [
            buildSubtitle({
                id: 0,
                source: 'embedded',
                key: '5:1',
                language: 'pt',
                is_forced: true,
            }),
        ];

        expect(chooseInitialSubtitleKey(subtitles, null, 'pt-BR')).toBe('5:1');
    });

    it('treats a subtitle without source as external', () => {
        const subtitles = [buildSubtitle({ id: 4 })];

        expect(findRememberedLanguageSubtitleKey(subtitles, enabledIn('en'))).toBe('external:4');
    });
});
