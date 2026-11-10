import type { TrashItem } from '@/types/trash';
import { daysUntilExpiry, trashItemName, trashItemOriginFolder } from './trashItemPresentation';

const trashItemAt = (originalPath: string, name?: string): TrashItem => ({
    id: 1,
    name,
    original_path: originalPath,
    size: 0,
    deleted_at: '2026-06-11T10:00:00Z',
});

describe('trashItemName', () => {
    it('prefers the name sent by the server', () => {
        expect(trashItemName(trashItemAt('/data/docs/a.pdf', 'relatorio.pdf'))).toBe(
            'relatorio.pdf'
        );
    });

    it('falls back to the last path segment', () => {
        expect(trashItemName(trashItemAt('/data/docs/a.pdf'))).toBe('a.pdf');
        expect(trashItemName(trashItemAt('C:\\data\\fotos\\'))).toBe('fotos');
    });

    it('is empty when there is no path at all', () => {
        expect(trashItemName({ id: 1 } as TrashItem)).toBe('');
    });
});

describe('trashItemOriginFolder', () => {
    it('returns the folder the item was deleted from', () => {
        expect(trashItemOriginFolder(trashItemAt('/data/docs/a.pdf'))).toBe('/data/docs');
        expect(trashItemOriginFolder(trashItemAt('C:\\data\\a.pdf'))).toBe('C:\\data');
    });

    it('returns the root for top-level items and missing paths', () => {
        expect(trashItemOriginFolder(trashItemAt('/a.pdf'))).toBe('/');
        expect(trashItemOriginFolder({ id: 1 } as TrashItem)).toBe('/');
    });
});

describe('daysUntilExpiry', () => {
    const currentTime = new Date('2026-06-11T10:00:00Z');

    it('rounds a partial day up', () => {
        expect(daysUntilExpiry('2026-06-13T09:00:00Z', currentTime)).toBe(2);
        expect(daysUntilExpiry('2026-06-11T11:00:00Z', currentTime)).toBe(1);
    });

    it('is zero once the expiry has passed', () => {
        expect(daysUntilExpiry('2026-06-01T00:00:00Z', currentTime)).toBe(0);
    });

    it('is unknown without a readable expiry', () => {
        expect(daysUntilExpiry(undefined, currentTime)).toBeUndefined();
        expect(daysUntilExpiry('not a date', currentTime)).toBeUndefined();
    });
});
