import { useMemo } from 'react';
import { type VideoPlaylistDto, type VideoPlaylistItemDto } from '@/service/videoPlayback';

export type VideoDetailItem = VideoPlaylistItemDto & {
	displayTitle: string;
	sequenceLabel: string;
	seasonNumber: number | null;
	episodeNumber: number | null;
};

export type VideoSeasonGroup = {
	key: string;
	label: string;
	items: VideoDetailItem[];
};

type ParsedEpisode = {
	displayTitle: string;
	seasonNumber: number | null;
	episodeNumber: number | null;
	sequenceLabel: string;
};

const extensionPattern = /\.[^/.]+$/;
const whitespacePattern = /[._-]+/g;
const spaceCollapsePattern = /\s+/g;

const parseEpisode = (name: string): ParsedEpisode => {
	const cleanName = name.replace(extensionPattern, '');

	const seasonEpisodeMatch = cleanName.match(/(\d{1,2})x(\d{1,2})/i) ?? cleanName.match(/s(\d{1,2})e(\d{1,2})/i);
	if (seasonEpisodeMatch) {
		const seasonNumber = Number(seasonEpisodeMatch[1] ?? '1');
		const episodeNumber = Number(seasonEpisodeMatch[2] ?? '0');
		const displayTitle = cleanName
			.replace(seasonEpisodeMatch[0], ' ')
			.replace(whitespacePattern, ' ')
			.replace(spaceCollapsePattern, ' ')
			.trim();

		return {
			displayTitle: displayTitle || cleanName,
			seasonNumber: Number.isFinite(seasonNumber) ? seasonNumber : null,
			episodeNumber: Number.isFinite(episodeNumber) && episodeNumber > 0 ? episodeNumber : null,
			sequenceLabel:
				Number.isFinite(episodeNumber) && episodeNumber > 0
					? `S${String(seasonNumber).padStart(2, '0')}E${String(episodeNumber).padStart(2, '0')}`
					: '',
		};
	}

	const episodeOnlyMatch = cleanName.match(
		/(?:(?:ep\.?\s*)|(?:episode\s*)|(?:epis[oó]dio\s*)|(?:cap[ií]tulo\s*))(\d{1,3})/i,
	);
	if (episodeOnlyMatch) {
		const episodeNumber = Number(episodeOnlyMatch[1] ?? '0');
		const displayTitle = cleanName
			.replace(episodeOnlyMatch[0], ' ')
			.replace(whitespacePattern, ' ')
			.replace(spaceCollapsePattern, ' ')
			.trim();

		return {
			displayTitle: displayTitle || cleanName,
			seasonNumber: 1,
			episodeNumber: Number.isFinite(episodeNumber) && episodeNumber > 0 ? episodeNumber : null,
			sequenceLabel:
				Number.isFinite(episodeNumber) && episodeNumber > 0
					? `S01E${String(episodeNumber).padStart(2, '0')}`
					: '',
		};
	}

	return {
		displayTitle: cleanName.replace(whitespacePattern, ' ').replace(spaceCollapsePattern, ' ').trim() || cleanName,
		seasonNumber: null,
		episodeNumber: null,
		sequenceLabel: '',
	};
};

export const useVideoPlaylistDetail = (playlist: VideoPlaylistDto) => {
	return useMemo(() => {
		const orderedItems = [...playlist.items]
			.sort((a, b) => a.order_index - b.order_index || a.id - b.id)
			.map((item) => {
				const parsed = parseEpisode(item.video.name);
				return {
					...item,
					...parsed,
				};
			});

		const completedCount = orderedItems.filter((item) => item.status === 'completed').length;
		const inProgressItem = orderedItems.find((item) => item.status === 'in_progress') ?? null;
		const resumeItem = inProgressItem ?? orderedItems.find((item) => item.status !== 'completed') ?? orderedItems[0] ?? null;
		const hasEpisodeData = orderedItems.some((item) => item.episodeNumber !== null);

		const groupedSeasons = hasEpisodeData
			? orderedItems.reduce<VideoSeasonGroup[]>((groups, item) => {
					const seasonNumber = item.seasonNumber ?? 1;
					const key = `season-${seasonNumber}`;
					const existingGroup = groups.find((group) => group.key === key);
					if (existingGroup) {
						existingGroup.items.push(item);
						return groups;
					}

					return [
						...groups,
						{
							key,
							label: String(seasonNumber),
							items: [item],
						},
					];
			  }, [])
			: [];

		return {
			orderedItems,
			groupedSeasons,
			completedCount,
			hasEpisodeData,
			resumeItem,
		};
	}, [playlist]);
};
