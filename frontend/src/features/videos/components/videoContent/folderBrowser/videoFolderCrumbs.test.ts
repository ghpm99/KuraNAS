import { buildVideoFolderCrumbs } from './videoFolderCrumbs';

describe('buildVideoFolderCrumbs', () => {
    it('returns no crumbs for the roots', () => {
        expect(buildVideoFolderCrumbs('')).toEqual([]);
    });

    it('builds cumulative paths and normalizes backslashes', () => {
        expect(buildVideoFolderCrumbs('\\Series\\S1')).toEqual([
            { name: 'Series', path: '/Series' },
            { name: 'S1', path: '/Series/S1' },
        ]);
    });
});
