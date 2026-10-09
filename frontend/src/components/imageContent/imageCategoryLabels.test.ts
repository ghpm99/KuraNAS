import { getImageCategoryLabelKey } from './imageCategoryLabels';

describe('imageCategoryLabels', () => {
    it('resolves a label key for every known and unknown category', () => {
        expect(getImageCategoryLabelKey('screenshot_app')).toBe(
            'IMAGES_CLASSIFICATION_SCREENSHOT_APP'
        );
        expect(getImageCategoryLabelKey(undefined)).toBe('IMAGES_CLASSIFICATION_OTHER');
        expect(getImageCategoryLabelKey('weird')).toBe('IMAGES_CLASSIFICATION_OTHER');
    });
});
