import { parseCapturesFilters, writeCapturesFilters } from './capturesFilters';

describe('parseCapturesFilters', () => {
    it('reads a valid status and the name', () => {
        expect(parseCapturesFilters(new URLSearchParams('q=naruto&status=failed'))).toEqual({
            name: 'naruto',
            status: 'failed',
            platform: '',
        });
    });

    it('reads the platform', () => {
        expect(parseCapturesFilters(new URLSearchParams('platform=youtube')).platform).toBe(
            'youtube'
        );
    });

    it('turns an invalid status into empty', () => {
        expect(parseCapturesFilters(new URLSearchParams('status=bogus')).status).toBe('');
    });

    it('defaults to empty filters when params are absent', () => {
        expect(parseCapturesFilters(new URLSearchParams())).toEqual({
            name: '',
            status: '',
            platform: '',
        });
    });
});

describe('writeCapturesFilters', () => {
    it('removes empty filters and keeps other params', () => {
        const written = writeCapturesFilters(
            new URLSearchParams('q=old&status=failed&platform=youtube&x=1'),
            {
                name: '',
                status: '',
                platform: '',
            }
        );
        expect(written.toString()).toBe('x=1');
    });

    it('sets filled filters', () => {
        const written = writeCapturesFilters(new URLSearchParams('x=1'), {
            name: 'naruto',
            status: 'promoted',
            platform: 'youtube',
        });
        expect(written.get('q')).toBe('naruto');
        expect(written.get('status')).toBe('promoted');
        expect(written.get('platform')).toBe('youtube');
        expect(written.get('x')).toBe('1');
    });
});
