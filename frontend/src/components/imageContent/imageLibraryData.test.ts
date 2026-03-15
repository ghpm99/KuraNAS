import {
	buildAutomaticAlbumCollections,
	buildFolderCollections,
	getAutomaticAlbumKey,
	getCollectionTitleFromPath,
	getImageDirectoryPath,
	getImageDate,
	isRecentImage,
	matchesImageSearch,
	matchesImageSection,
} from './imageLibraryData';

const createImage = (overrides: Record<string, any> = {}) => ({
	id: 1,
	name: 'image.jpg',
	path: '/library/trips/image.jpg',
	format: '.jpg',
	size: 1024,
	updated_at: '2026-03-01T10:00:00Z',
	created_at: '2026-03-01T10:00:00Z',
	metadata: {
		width: 2000,
		height: 1200,
		datetime_original: '2026-03-01T10:00:00Z',
		classification: {
			category: 'photo',
			confidence: 0.98,
		},
		...overrides.metadata,
	},
	...overrides,
});

describe('imageLibraryData', () => {
	it('derives image directory paths from file paths and folder paths', () => {
		expect(getImageDirectoryPath(createImage())).toBe('/library/trips');
		expect(getImageDirectoryPath(createImage({ path: '/library/camera' }))).toBe('/library/camera');
		expect(getImageDirectoryPath(createImage({ path: 'C:\\library\\trips\\image.jpg' }))).toBe('C:/library/trips');
	});

	it('reads dates and section matching from persisted metadata', () => {
		const recentDate = getImageDate(createImage());
		const oldDate = new Date('2025-01-01T10:00:00Z');

		expect(isRecentImage(recentDate, new Date('2026-03-14T12:00:00Z').getTime())).toBe(true);
		expect(isRecentImage(oldDate, new Date('2026-03-14T12:00:00Z').getTime())).toBe(false);
		expect(matchesImageSection('captures', createImage({ metadata: { classification: { category: 'capture', confidence: 0.91 } } }), recentDate)).toBe(true);
		expect(matchesImageSection('photos', createImage(), recentDate)).toBe(true);
		expect(matchesImageSection('recent', createImage(), new Date())).toBe(true);
		expect(matchesImageSection('folders', createImage(), null)).toBe(true);
	});

	it('returns fallback values for missing directory and date metadata', () => {
		expect(getImageDirectoryPath(createImage({ path: '' }))).toBe('/');
		expect(getCollectionTitleFromPath('/')).toBe('/');
		expect(
			getImageDate(
				createImage({
					updated_at: '',
					created_at: '',
					metadata: {
						datetime_original: '',
						datetime: '',
						createdAt: '',
					},
				}),
			),
		).toBeNull();
	});

	it('builds folder collections sorted by latest image date', () => {
		const collections = buildFolderCollections([
			createImage({ id: 1, path: '/library/a/one.jpg', created_at: '2026-03-01T10:00:00Z', updated_at: '2026-03-01T10:00:00Z' }),
			createImage({ id: 2, path: '/library/b/two.jpg', created_at: '2026-03-10T10:00:00Z', updated_at: '2026-03-10T10:00:00Z', metadata: { datetime_original: '2026-03-10T10:00:00Z' } }),
		]);

		expect(collections).toHaveLength(2);
		expect(collections[0]?.id).toBe('/library/b');
		expect(collections[0]?.cover?.id).toBe(2);
		expect(collections[1]?.id).toBe('/library/a');
	});

	it('classifies automatic albums with deterministic heuristics', () => {
		expect(getAutomaticAlbumKey(createImage({ name: 'beach-day.jpg', path: '/library/travel/beach-day.jpg' }))).toBe('travel');
		expect(getAutomaticAlbumKey(createImage({ name: 'receipt.jpg', path: '/library/docs/receipt.jpg', metadata: { width: 900, height: 1400 } }))).toBe('documents');
		expect(getAutomaticAlbumKey(createImage({ name: 'desktop.jpg', path: '/library/walls/desktop.jpg', metadata: { width: 3840, height: 2160 } }))).toBe('wallpapers');
		expect(getAutomaticAlbumKey(createImage({ name: 'party-meme.png', path: '/library/memes/party-meme.png', metadata: { classification: { category: 'capture', confidence: 0.8 } } }))).toBe('memes');
		expect(getAutomaticAlbumKey(createImage({ name: 'misc.jpg', path: '/library/random/misc.jpg', metadata: { width: 1200, height: 900 } }))).toBe('others');

		const collections = buildAutomaticAlbumCollections([
			createImage({ id: 1, name: 'beach-day.jpg', path: '/library/travel/beach-day.jpg' }),
			createImage({ id: 2, name: 'receipt.jpg', path: '/library/docs/receipt.jpg', metadata: { width: 900, height: 1400 } }),
			createImage({ id: 3, name: 'desktop.jpg', path: '/library/walls/desktop.jpg', metadata: { width: 3840, height: 2160 } }),
			createImage({ id: 4, name: 'party-meme.png', path: '/library/memes/party-meme.png', metadata: { classification: { category: 'capture', confidence: 0.8 } } }),
			createImage({ id: 5, name: 'misc.jpg', path: '/library/random/misc.jpg', metadata: { width: 1200, height: 900 } }),
		]);

		expect(collections.map((collection) => collection.id)).toEqual(['travel', 'documents', 'wallpapers', 'memes', 'others']);
		expect(collections.find((collection) => collection.id === 'travel')?.images).toHaveLength(1);
		expect(collections.find((collection) => collection.id === 'others')?.images).toHaveLength(1);
	});

	it('matches search across names, paths, and metadata', () => {
		const image = createImage({
			name: 'Family dinner.jpg',
			path: '/library/home/family-dinner.jpg',
			metadata: {
				make: 'Canon',
				model: 'R6',
				image_description: 'Sunday dinner',
			},
		});

		expect(matchesImageSearch(image, '')).toBe(true);
		expect(matchesImageSearch(image, 'canon')).toBe(true);
		expect(matchesImageSearch(image, 'family')).toBe(true);
		expect(matchesImageSearch(image, 'not-found')).toBe(false);
	});
});
