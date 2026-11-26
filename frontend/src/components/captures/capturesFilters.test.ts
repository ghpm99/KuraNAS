import { parseCapturesFilters, writeCapturesFilters } from './capturesFilters';

describe('parseCapturesFilters', () => {
    it('reads a valid status and the name', () => {
        expect(parseCapturesFilters(new URLSearchParams('q=naruto&status=failed'))).toEqual({
            name: 'naruto',
            status: 'failed',
        });
    });

    it('turns an invalid status into empty', () => {
        expect(parseCapturesFilters(new URLSearchParams('status=bogus')).status).toBe('');
    });

    it('defaults to empty filters when params are absent', () => {
        expect(parseCapturesFilters(new URLSearchParams())).toEqual({ name: '', status: '' });
    });
});

describe('writeCapturesFilters', () => {
    it('removes empty filters and keeps other params', () => {
        const written = writeCapturesFilters(new URLSearchParams('q=old&status=failed&x=1'), {
            name: '',
            status: '',
        });
        expect(written.toString()).toBe('x=1');
    });

    it('sets filled filters', () => {
        const written = writeCapturesFilters(new URLSearchParams('x=1'), {
            name: 'naruto',
            status: 'promoted',
        });
        expect(written.get('q')).toBe('naruto');
        expect(written.get('status')).toBe('promoted');
        expect(written.get('x')).toBe('1');
    });
});
