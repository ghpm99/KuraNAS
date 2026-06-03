import type {
    IMusicData,
    IMusicMetadata,
} from '@/features/music/providers/musicProvider/musicProvider';
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

export const musicMetadata = (music: {
    format: string;
    size: number;
    metadata?: IMusicMetadata;
}): string => {
    const format = music.format ? `${music.format} - ` : '';
    const fileSize = formatSize(music.size);
    const dur = music.metadata?.duration ? formatMusicDuration(music.metadata.duration) : '';
    return `${format}${fileSize}${dur ? ` - ${dur}` : ''}`;
};
