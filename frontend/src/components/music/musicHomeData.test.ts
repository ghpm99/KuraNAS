import type { IMusicData } from '@/components/providers/musicProvider/musicProvider';
import { buildMusicAlbumHighlights, buildMusicArtistHighlights } from './musicHomeData';

const track = (overrides: Partial<IMusicData>): IMusicData =>
	({
		id: 1,
		name: 'track.mp3',
		path: '/library/track.mp3',
		type: 1,
		format: 'mp3',
		size: 100,
		updated_at: '',
		created_at: '',
		deleted_at: '',
		last_interaction: '',
		last_backup: '',
		check_sum: '',
		directory_content_count: 0,
		starred: false,
		metadata: {
			id: 1,
			fileId: 1,
			path: '/library/track.mp3',
			format: 'mp3',
			title: 'Track',
			artist: '',
			album: '',
			year: 2024,
			genre: 'Jazz',
			track: 1,
			disc: 1,
			duration: 180,
			bitrate: 320,
			sampleRate: 44100,
			channels: 2,
			createdAt: '',
		},
		...overrides,
	}) as IMusicData;

describe('music home data helpers', () => {
	it('groups artists by latest activity and counts unique albums', () => {
		const artists = buildMusicArtistHighlights(
			[
				track({ id: 1, created_at: '2026-03-10T10:00:00Z', metadata: { ...track({}).metadata!, artist: 'Artist B', album: 'Album 1' } }),
				track({ id: 2, created_at: '2026-03-11T10:00:00Z', metadata: { ...track({}).metadata!, artist: 'Artist A', album: 'Album 1' } }),
				track({ id: 3, created_at: '2026-03-12T10:00:00Z', metadata: { ...track({}).metadata!, artist: 'Artist A', album: 'Album 2' } }),
				track({ id: 4, created_at: '2026-03-09T10:00:00Z', metadata: { ...track({}).metadata!, artist: '', album: 'Ignored' } }),
			],
			5,
		);

		expect(artists).toEqual([
			{ artist: 'Artist A', trackCount: 2, albumCount: 2 },
			{ artist: 'Artist B', trackCount: 1, albumCount: 1 },
		]);
	});

	it('groups albums by artist and sorts the newest ones first', () => {
		const albums = buildMusicAlbumHighlights(
			[
				track({
					id: 1,
					created_at: '2026-03-10T10:00:00Z',
					metadata: { ...track({}).metadata!, artist: 'Artist A', album: 'Album 1', year: 2022 },
				}),
				track({
					id: 2,
					created_at: '2026-03-12T10:00:00Z',
					metadata: { ...track({}).metadata!, artist: 'Artist B', album: 'Album 2', year: 2025 },
				}),
				track({
					id: 3,
					created_at: '2026-03-11T10:00:00Z',
					metadata: { ...track({}).metadata!, artist: 'Artist A', album: 'Album 1', year: 2022 },
				}),
				track({
					id: 4,
					created_at: '2026-03-09T10:00:00Z',
					metadata: { ...track({}).metadata!, artist: '', album: 'Ignored', year: 2024 },
				}),
			],
			5,
		);

		expect(albums).toEqual([
			{ album: 'Album 2', artist: 'Artist B', year: '2025', trackCount: 1 },
			{ album: 'Album 1', artist: 'Artist A', year: '2022', trackCount: 2 },
		]);
	});

	it('uses timestamp fallbacks, trims values, ignores invalid artists, and respects the limit', () => {
		const artists = buildMusicArtistHighlights(
			[
				track({
					id: 1,
					updated_at: '2026-03-10T10:00:00Z',
					metadata: { ...track({}).metadata!, artist: '  Artist C  ', album: '  Album C  ' },
				}),
				track({
					id: 2,
					last_interaction: '2026-03-12T10:00:00Z',
					metadata: { ...track({}).metadata!, artist: 'Artist B', album: '   ' },
				}),
				track({
					id: 3,
					created_at: 'invalid-date',
					metadata: { ...track({}).metadata!, artist: 'Artist A', album: 'Album A' },
				}),
				track({
					id: 4,
					metadata: { ...track({}).metadata!, artist: 'Artist B', album: 'Album B' },
				}),
				track({
					id: 5,
					metadata: undefined as any,
				}),
			],
			2,
		);

		expect(artists).toEqual([
			{ artist: 'Artist B', trackCount: 2, albumCount: 1 },
			{ artist: 'Artist C', trackCount: 1, albumCount: 1 },
		]);
	});

	it('uses album timestamp fallbacks, handles missing year, and sorts ties alphabetically', () => {
		const albums = buildMusicAlbumHighlights(
			[
				track({
					id: 1,
					updated_at: '2026-03-11T10:00:00Z',
					metadata: { ...track({}).metadata!, artist: ' Artist A ', album: ' Album B ', year: 2020 },
				}),
				track({
					id: 2,
					created_at: '2026-03-11T10:00:00Z',
					metadata: { ...track({}).metadata!, artist: 'Artist A', album: 'Album A', year: 2020 },
				}),
				track({
					id: 3,
					last_interaction: '2026-03-12T10:00:00Z',
					metadata: { ...track({}).metadata!, artist: 'Artist B', album: 'Album Y', year: 0 },
				}),
				track({
					id: 4,
					metadata: { ...track({}).metadata!, artist: 'Artist B', album: 'Album Y', year: 0 },
				}),
				track({
					id: 5,
					created_at: 'invalid-date',
					metadata: { ...track({}).metadata!, artist: 'Artist C', album: 'Album Z', year: 2021 },
				}),
				track({
					id: 6,
					metadata: { ...track({}).metadata!, artist: '', album: 'Ignored', year: 2024 },
				}),
			],
			3,
		);

		expect(albums).toEqual([
			{ album: 'Album Y', artist: 'Artist B', year: undefined, trackCount: 2 },
			{ album: 'Album A', artist: 'Artist A', year: '2020', trackCount: 1 },
			{ album: 'Album B', artist: 'Artist A', year: '2020', trackCount: 1 },
		]);
	});
});
