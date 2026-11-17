import type { DuplicateCopy, DuplicateGroupDetail } from '@/types/analytics';

export const buildDuplicateCopy = (
    id: number,
    overrides: Partial<DuplicateCopy> = {}
): DuplicateCopy => ({
    id,
    path: `/Fotos/copia-${id}.jpg`,
    updated_at: '2026-03-10T10:00:00Z',
    tier: 'hot',
    ...overrides,
});

export const buildDuplicateGroup = (
    signature: string,
    copies: DuplicateCopy[]
): DuplicateGroupDetail => ({
    signature,
    copies: copies.length,
    size_bytes: 4096,
    reclaimable_size: 4096 * (copies.length - 1),
    paths: copies.map((copy) => copy.path),
    files: copies,
});
