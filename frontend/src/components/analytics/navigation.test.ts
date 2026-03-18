import {
    getAnalyticsSectionFromPath,
    getAnalyticsSectionMeta,
} from './navigation';

describe('analytics navigation helpers', () => {
    it('resolves the current section from known paths', () => {
        expect(getAnalyticsSectionFromPath('/analytics')).toBe('overview');
        expect(getAnalyticsSectionFromPath('/analytics/library')).toBe('library');
        expect(getAnalyticsSectionFromPath('/analytics/library/errors')).toBe('library');
    });

    it('falls back to overview metadata for unknown paths', () => {
        expect(getAnalyticsSectionFromPath('/analytics/unknown')).toBe('overview');
        expect(getAnalyticsSectionMeta('overview')).toEqual(
            expect.objectContaining({ key: 'overview' })
        );
    });
});
