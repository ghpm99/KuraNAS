import type { IMusicData } from '@/features/music/providers/musicProvider/musicProvider';
import type { MusicQueue, MusicQueueEntry } from '@/types/music';

const AUDIO_FILE_TYPE = 2;

export const queueEntryToTrack = (entry: MusicQueueEntry): IMusicData => ({
    id: entry.file_id,
    name: entry.name,
    path: entry.path,
    type: AUDIO_FILE_TYPE,
    format: entry.format,
    size: 0,
    updated_at: '',
    created_at: '',
    deleted_at: '',
    last_interaction: '',
    last_backup: '',
    check_sum: '',
    directory_content_count: 0,
    starred: false,
    metadata: {
        title: entry.title,
        artist: entry.artist,
        album: entry.album,
        length: entry.length,
    },
});

export const queueToTracks = (queue: Partial<MusicQueue> | undefined): IMusicData[] =>
    (queue?.items ?? []).map(queueEntryToTrack);

export const findStartIndex = (tracks: IMusicData[], trackId?: number) =>
    trackId === undefined
        ? 0
        : Math.max(
              tracks.findIndex((track) => track.id === trackId),
              0
          );
