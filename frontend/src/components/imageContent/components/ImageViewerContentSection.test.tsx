import { fireEvent, render, screen } from '@testing-library/react';
import ImageViewerContentSection from './ImageViewerContentSection';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string, params?: Record<string, string>) =>
            params?.tag ? `${key}:${params.tag}` : key,
    }),
}));

describe('ImageViewerContentSection', () => {
    it('shows the caption and runs a search when a tag chip is clicked', () => {
        const onSearchTag = jest.fn();
        render(
            <ImageViewerContentSection
                content={{ caption: 'A dog in the park', tags: ['dog', 'park'], ocrText: '' }}
                onSearchTag={onSearchTag}
            />
        );

        expect(screen.getByText('A dog in the park')).toBeInTheDocument();
        fireEvent.click(screen.getByRole('button', { name: 'IMAGES_DETAIL_TAG_SEARCH:park' }));

        expect(onSearchTag).toHaveBeenCalledWith('park');
        expect(screen.queryByText('IMAGES_DETAIL_OCR_SHOW')).not.toBeInTheDocument();
    });

    it('keeps the OCR text collapsed until the user expands it', () => {
        render(
            <ImageViewerContentSection
                content={{ caption: '', tags: [], ocrText: 'STOP' }}
                onSearchTag={jest.fn()}
            />
        );

        expect(screen.queryByText('STOP')).not.toBeInTheDocument();
        const toggle = screen.getByRole('button', { name: 'IMAGES_DETAIL_OCR_SHOW' });
        expect(toggle).toHaveAttribute('aria-expanded', 'false');

        fireEvent.click(toggle);

        expect(screen.getByText('STOP')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'IMAGES_DETAIL_OCR_HIDE' })).toHaveAttribute(
            'aria-expanded',
            'true'
        );
    });
});
