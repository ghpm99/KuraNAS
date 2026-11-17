import { duplicateCategories, type DuplicateCategory } from '@/types/analytics';

export const duplicateCategoryQueryParam = 'type';

export const parseDuplicateCategory = (search: string): DuplicateCategory | undefined => {
    const rawCategory = new URLSearchParams(search).get(duplicateCategoryQueryParam);
    return duplicateCategories.find((category) => category === rawCategory);
};
