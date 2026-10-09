import type { ImageCategory } from '@/types/imageLibrary';

const imageCategoryLabelKeys: Record<ImageCategory, string> = {
    capture: 'IMAGES_CLASSIFICATION_CAPTURE',
    photo: 'IMAGES_CLASSIFICATION_PHOTO',
    other: 'IMAGES_CLASSIFICATION_OTHER',
    document: 'IMAGES_CLASSIFICATION_DOCUMENT',
    receipt: 'IMAGES_CLASSIFICATION_RECEIPT',
    landscape: 'IMAGES_CLASSIFICATION_LANDSCAPE',
    portrait: 'IMAGES_CLASSIFICATION_PORTRAIT',
    meme: 'IMAGES_CLASSIFICATION_MEME',
    art: 'IMAGES_CLASSIFICATION_ART',
    screenshot_app: 'IMAGES_CLASSIFICATION_SCREENSHOT_APP',
};

export const getImageCategoryLabelKey = (category: string | null | undefined): string =>
    imageCategoryLabelKeys[category as ImageCategory] ?? imageCategoryLabelKeys.other;
