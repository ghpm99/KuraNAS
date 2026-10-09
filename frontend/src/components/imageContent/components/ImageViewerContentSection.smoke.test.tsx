import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import ImageViewerContentSection from './ImageViewerContentSection';

describe('ImageViewerContentSection (no-mock render)', () => {
    it('renders nothing when the backend sent no content at all', () => {
        const { container } = renderWithoutBackend(
            <ImageViewerContentSection onSearchTag={jest.fn()} />
        );

        expect(container).toBeEmptyDOMElement();
    });

    it('renders nothing for a partial payload with empty fields', () => {
        const { container } = renderWithoutBackend(
            <ImageViewerContentSection content={{ tags: undefined }} onSearchTag={jest.fn()} />
        );

        expect(container).toBeEmptyDOMElement();
    });
});
