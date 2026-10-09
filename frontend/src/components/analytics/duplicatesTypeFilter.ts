import type { DuplicatesType } from '@/types/analytics';

export const parseDuplicatesType = (search: string): DuplicatesType | undefined =>
    new URLSearchParams(search).get('type') === 'image' ? 'image' : undefined;
