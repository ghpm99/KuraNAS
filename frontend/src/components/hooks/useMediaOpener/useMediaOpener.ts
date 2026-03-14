import { appRoutes } from '@/app/routes';
import { createRouteMusicPlaybackContext } from '@/components/music/playbackContext';
import { useGlobalMusic } from '@/components/providers/GlobalMusicProvider';
import type { IMusicData, IMusicMetadata } from '@/components/providers/musicProvider/musicProvider';
import { FileType, getFileTypeInfo } from '@/utils';
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
	metadata?: IMusicMetadata;
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
	metadata: file.metadata,
});

export default function useMediaOpener() {
	const navigate = useNavigate();
	const location = useLocation();
	const { replaceQueue } = useGlobalMusic();

	const openMediaItem = useCallback(
		(file: OpenableMediaFile) => {
			if (file.type === FileType.Directory) {
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
					navigate({
						pathname: appRoutes.images,
						search: `?image=${file.id}`,
					}, {
						state: { from: currentRoute },
					});
					return true;
				case 'audio':
					replaceQueue([toMusicTrack(file)], 0, createRouteMusicPlaybackContext(location.pathname, location.search));
					navigate(appRoutes.music, {
						state: { from: currentRoute },
					});
					return true;
				default:
					return false;
			}
		},
		[location.pathname, location.search, navigate, replaceQueue],
	);

	return {
		openMediaItem,
	};
}
