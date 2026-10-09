import { getTranslationRetryDelay } from './translationRetry';

describe('i18n/provider/translationRetry', () => {
    it('doubles the delay per attempt up to a ceiling', () => {
        expect(getTranslationRetryDelay(0)).toBe(1000);
        expect(getTranslationRetryDelay(2)).toBe(4000);
        expect(getTranslationRetryDelay(10)).toBe(15000);
    });
});
