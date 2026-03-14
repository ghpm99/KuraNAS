import { getVideoSectionFromPath, getVideoSectionMeta, videoNavigationItems } from './navigation';

describe('video navigation helpers', () => {
	it('resolves the current section from known paths', () => {
		expect(getVideoSectionFromPath('/videos')).toBe('home');
		expect(getVideoSectionFromPath('/videos/clips')).toBe('clips');
	});

	it('falls back to home metadata for unknown paths', () => {
		expect(getVideoSectionFromPath('/videos/unknown')).toBe('home');
		expect(getVideoSectionMeta('home')).toEqual(videoNavigationItems[0]);
	});
});
