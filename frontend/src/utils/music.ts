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
