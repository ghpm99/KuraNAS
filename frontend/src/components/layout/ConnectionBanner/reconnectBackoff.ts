const BASE_DELAY_MS = 2000;
const MAX_DELAY_MS = 30000;

export const getReconnectDelay = (attemptIndex: number): number =>
    Math.min(BASE_DELAY_MS * 2 ** attemptIndex, MAX_DELAY_MS);
