const MAX_TRANSLATION_RETRIES = 6;
const BASE_DELAY_MS = 1000;
const MAX_DELAY_MS = 15000;

export const translationRetryCount = MAX_TRANSLATION_RETRIES;

export const getTranslationRetryDelay = (attemptIndex: number): number =>
    Math.min(BASE_DELAY_MS * 2 ** attemptIndex, MAX_DELAY_MS);
