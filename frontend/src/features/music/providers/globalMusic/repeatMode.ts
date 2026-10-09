export type RepeatMode = 'none' | 'all' | 'one';

export const parseRepeatMode = (rawRepeatMode: string | undefined): RepeatMode =>
    rawRepeatMode === 'all' || rawRepeatMode === 'one' ? rawRepeatMode : 'none';
