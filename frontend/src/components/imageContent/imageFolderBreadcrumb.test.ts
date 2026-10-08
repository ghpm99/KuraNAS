import { buildFolderCrumbs } from './imageFolderBreadcrumb';

describe('buildFolderCrumbs', () => {
    it('builds one crumb per segment with cumulative paths', () => {
        expect(buildFolderCrumbs('/photos/trip/day1')).toEqual([
            { name: 'photos', path: '/photos' },
            { name: 'trip', path: '/photos/trip' },
            { name: 'day1', path: '/photos/trip/day1' },
        ]);
    });

    it('normalizes backslashes and yields nothing for the roots', () => {
        expect(buildFolderCrumbs('\\Fotos\\Casa')).toEqual([
            { name: 'Fotos', path: '/Fotos' },
            { name: 'Casa', path: '/Fotos/Casa' },
        ]);
        expect(buildFolderCrumbs('')).toEqual([]);
        expect(buildFolderCrumbs('/')).toEqual([]);
    });
});
