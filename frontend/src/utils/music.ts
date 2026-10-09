import type { IMusicData } from '@/features/music/providers/musicProvider/musicProvider';
import type { IMusicMetadata } from '@/types/music';
import { formatSize } from '@/utils';

export const getMusicTitle = (music: IMusicData): string => {
    return music.metadata?.title || music.name;
};

export const getMusicArtist = (music: IMusicData): string => {
    return music.metadata?.artist || 'Unknown Artist';
};

export const formatMusicDuration = (seconds: number): string => {
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}:${secs.toString().padStart(2, '0')}`;
};

export const getTrackDurationSeconds = (metadata?: IMusicMetadata): number => {
    const lengthSeconds = metadata?.length;
    return typeof lengthSeconds === 'number' && Number.isFinite(lengthSeconds) && lengthSeconds > 0
        ? lengthSeconds
        : 0;
};

export const musicMetadata = (music: {
    format: string;
    size: number;
    metadata?: IMusicMetadata;
}): string => {
    const format = music.format ? `${music.format} - ` : '';
    const fileSize = formatSize(music.size);
    const durationSeconds = getTrackDurationSeconds(music.metadata);
    const dur = durationSeconds ? formatMusicDuration(durationSeconds) : '';
    return `${format}${fileSize}${dur ? ` - ${dur}` : ''}`;
};

const LEADING_NUMBER_PATTERN = /^\s*(\d{1,9})/;

const parseLeadingNumber = (numberText?: string): number | undefined => {
    const match = LEADING_NUMBER_PATTERN.exec(numberText ?? '');
    return match ? Number(match[1]) : undefined;
};

export const parseTrackNumber = (metadata?: IMusicMetadata): number | undefined =>
    parseLeadingNumber(metadata?.track_number);

export const parseDiscNumber = (metadata?: IMusicMetadata): number =>
    parseLeadingNumber(metadata?.disc_number) ?? 1;

export const formatTotalDuration = (
    totalSeconds: number,
    translate: (key: string, options?: Record<string, string>) => string
): string => {
    const totalMinutes = Math.max(0, Math.round(totalSeconds / 60));
    const hours = Math.floor(totalMinutes / 60);
    const minutes = totalMinutes % 60;
    if (hours === 0) {
        return translate('MUSIC_DURATION_MINUTES', { minutes: String(minutes) });
    }
    return translate('MUSIC_DURATION_HOURS_MINUTES', {
        hours: String(hours),
        minutes: String(minutes).padStart(2, '0'),
    });
};
