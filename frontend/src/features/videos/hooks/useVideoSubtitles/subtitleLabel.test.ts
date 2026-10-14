import { buildSubtitleLabel } from './subtitleLabel';

const translate = (key: string) => `[${key}]`;

describe('subtitleLabel', () => {
    it('uses the language and the upper-cased format', () => {
        expect(
            buildSubtitleLabel(
                { id: 1, label: 'x', language: 'pt-BR', format: 'srt', is_forced: false },
                translate
            )
        ).toBe('pt-BR (SRT)');
    });

    it('falls back to the label and marks forced tracks', () => {
        expect(
            buildSubtitleLabel(
                { id: 1, label: 'Commentary', language: '', format: 'vtt', is_forced: true },
                translate
            )
        ).toBe('Commentary (VTT) · [VIDEO_SUBTITLES_FORCED]');
    });
});
