import { fireEvent, render, screen } from '@testing-library/react';
import ImageCollectionsPanel from './ImageCollectionsPanel';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string, params?: Record<string, string>) => {
            if (key === 'IMAGES_COLLECTION_OPEN') {
                return `Abrir ${params?.name ?? ''}`.trim();
            }
            if (key === 'IMAGES_PHOTOS_COUNT') {
                return `${params?.count ?? '0'} fotos`;
            }
            return key;
        },
    }),
}));

jest.mock('@/service/apiUrl', () => ({
    getApiV1BaseUrl: () => '/api/v1',
}));

describe('ImageCollectionsPanel', () => {
    it('renders an empty state when there are no cards', () => {
        render(
            <ImageCollectionsPanel
                cards={[]}
                emptyTitle="Nada aqui"
                emptyDescription="Sem colecoes"
                onSelect={jest.fn()}
            />
        );

        expect(screen.getByText('Nada aqui')).toBeInTheDocument();
        expect(screen.getByText('Sem colecoes')).toBeInTheDocument();
    });

    it('renders cards with placeholder/thumbnail and handles selection', () => {
        const onSelect = jest.fn();
        render(
            <ImageCollectionsPanel
                cards={[
                    {
                        id: 'folder-a',
                        title: 'Folder A',
                        description: '/photos/folder-a',
                        imageCount: 3,
                    },
                    {
                        id: 'folder-b',
                        title: 'Folder B',
                        description: '/photos/folder-b',
                        imageCount: 4,
                        coverImageId: 10,
                    },
                ]}
                emptyTitle="Nada aqui"
                emptyDescription="Sem colecoes"
                selectedId="folder-a"
                onSelect={onSelect}
            />
        );

        expect(screen.getByText('F')).toBeInTheDocument();
        expect(screen.getByAltText('Folder B')).toHaveAttribute(
            'src',
            '/api/v1/files/thumbnail/10?width=960&height=720'
        );

        fireEvent.click(screen.getByRole('button', { name: /abrir folder b/i }));
        expect(onSelect).toHaveBeenCalledWith('folder-b');
    });
});
