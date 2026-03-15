import { getVideoDetailRoute, getVideoDetailSlugFromPath, getVideoSectionFromPath, getVideoSectionMeta, videoNavigationItems } from './navigation';

describe('video navigation helpers', () => {
	it('resolves the current section from known paths', () => {
		expect(getVideoSectionFromPath('/videos')).toBe('home');
		expect(getVideoSectionFromPath('/videos/clips')).toBe('clips');
		expect(getVideoSectionFromPath('/videos/series/breaking-bad')).toBe('series');
		expect(getVideoDetailSlugFromPath('/videos/series/breaking-bad')).toBe('breaking-bad');
		expect(getVideoDetailRoute('movies', 'the-matrix')).toBe('/videos/movies/the-matrix');
	});

	it('falls back to home metadata for unknown paths', () => {
		expect(getVideoSectionFromPath('/videos/unknown')).toBe('home');
		expect(getVideoSectionMeta('home')).toEqual(videoNavigationItems[0]);
	});
});
