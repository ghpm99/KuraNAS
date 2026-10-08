import { appRoutes } from '@/app/routes';
import { createRouteMusicPlaybackContext } from '@/features/music/components/playbackContext';
import { useGlobalMusic } from '@/features/music/providers/GlobalMusicProvider';
import type { IImageMetadata } from '@/components/providers/imageProvider/imageProvider';
import type {
    IMusicData,
    IMusicMetadata,
} from '@/features/music/providers/musicProvider/musicProvider';
import { FileType, getFileTypeInfo, hasDedicatedMediaScreen } from '@/utils';
import { useCallback } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';

type OpenableMediaFile = {
    id: number;
    name: string;
    format: string;
    type?: number;
    path?: string;
    size?: number;
    updated_at?: string;
    created_at?: string;
    deleted_at?: string;
    last_interaction?: string;
    last_backup?: string;
    check_sum?: string;
    directory_content_count?: number;
    starred?: boolean;
    metadata?: IMusicMetadata | IImageMetadata;
};

const getCurrentRoute = (pathname: string, search: string) => `${pathname}${search}`;

const toMusicTrack = (file: OpenableMediaFile): IMusicData => ({
    id: file.id,
    name: file.name,
    path: file.path ?? '',
    type: file.type ?? FileType.File,
    format: file.format,
    size: file.size ?? 0,
    updated_at: file.updated_at ?? '',
    created_at: file.created_at ?? '',
    deleted_at: file.deleted_at ?? '',
    last_interaction: file.last_interaction ?? '',
    last_backup: file.last_backup ?? '',
    check_sum: file.check_sum ?? '',
    directory_content_count: file.directory_content_count ?? 0,
    starred: file.starred ?? false,
    metadata: isMusicMetadata(file.metadata) ? file.metadata : undefined,
});

const isMusicMetadata = (metadata: OpenableMediaFile['metadata']): metadata is IMusicMetadata => {
    if (!metadata) {
        return false;
    }

    return 'duration' in metadata || 'album' in metadata || 'artist' in metadata;
};

const isQueueableAudioFile = (file: OpenableMediaFile): boolean =>
    file.type !== FileType.Directory &&
    hasDedicatedMediaScreen(file.format) &&
    getFileTypeInfo(file.format).type === 'audio';

const buildAudioQueue = (
    clickedFile: OpenableMediaFile,
    listedFiles: OpenableMediaFile[]
): { tracks: IMusicData[]; startIndex: number } => {
    const listedAudioFiles = listedFiles.filter(isQueueableAudioFile);
    const clickedIndex = listedAudioFiles.findIndex((audioFile) => audioFile.id === clickedFile.id);
    if (clickedIndex === -1) {
        return { tracks: [toMusicTrack(clickedFile)], startIndex: 0 };
    }
    return { tracks: listedAudioFiles.map(toMusicTrack), startIndex: clickedIndex };
};

export default function useMediaOpener() {
    const navigate = useNavigate();
    const location = useLocation();
    const { replaceQueue } = useGlobalMusic();

    const enqueueAudioFiles = useCallback(
        (clickedFile: OpenableMediaFile, listedFiles: OpenableMediaFile[]) => {
            const { tracks, startIndex } = buildAudioQueue(clickedFile, listedFiles);
            replaceQueue(
                tracks,
                startIndex,
                createRouteMusicPlaybackContext(location.pathname, location.search)
            );
        },
        [location.pathname, location.search, replaceQueue]
    );

    const openMediaItem = useCallback(
        (file: OpenableMediaFile, listedFiles: OpenableMediaFile[] = []) => {
            if (file.type === FileType.Directory || !hasDedicatedMediaScreen(file.format)) {
                return false;
            }

            const currentRoute = getCurrentRoute(location.pathname, location.search);
            const fileType = getFileTypeInfo(file.format);

            switch (fileType.type) {
                case 'video':
                    navigate(`${appRoutes.videoPlayerBase}/${file.id}`, {
                        state: { from: currentRoute },
                    });
                    return true;
                case 'image':
                    navigate(
                        {
                            pathname: appRoutes.images,
                            search: file.path
                                ? `?image=${file.id}&imagePath=${encodeURIComponent(file.path)}`
                                : `?image=${file.id}`,
                        },
                        {
                            state: { from: currentRoute },
                        }
                    );
                    return true;
                case 'audio':
                    enqueueAudioFiles(file, listedFiles);
                    navigate(appRoutes.music, {
                        state: { from: currentRoute },
                    });
                    return true;
                default:
                    return false;
            }
        },
        [location.pathname, location.search, navigate, enqueueAudioFiles]
    );

    return {
        openMediaItem,
    };
}
