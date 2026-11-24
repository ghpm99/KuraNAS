import { fireEvent, screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import DocumentCard from './DocumentCard';

const document = {
    id: 1,
    name: 'relatorio.pdf',
    parent_path: '/docs',
    format: 'pdf',
    size: 10,
    modified_at: '2024-01-01T00:00:00Z',
    starred: true,
} as never;

describe('DocumentCard', () => {
    it('renders name, folder and star', () => {
        renderWithoutBackend(<DocumentCard document={document} />);

        expect(screen.getByText('relatorio.pdf')).toBeInTheDocument();
        expect(screen.getByText('/docs')).toBeInTheDocument();
        expect(screen.getByLabelText('DOCUMENTS_STARRED')).toBeInTheDocument();
    });

    it('opens the document from its name when an open handler is given', () => {
        const onOpen = jest.fn();
        renderWithoutBackend(<DocumentCard document={document} onOpen={onOpen} />);

        fireEvent.click(screen.getByRole('button', { name: 'relatorio.pdf' }));

        expect(onOpen).toHaveBeenCalledTimes(1);
    });

    it('omits the star when not starred', () => {
        renderWithoutBackend(
            <DocumentCard document={{ ...(document as object), starred: false } as never} />
        );

        expect(screen.queryByLabelText('DOCUMENTS_STARRED')).toBeNull();
    });
});

describe('DocumentCard with interaction', () => {
    it('renders the controls and reports clicks and right click', () => {
        const interaction = {
            isSelected: false,
            isSelectionActive: false,
            onToggleSelection: jest.fn(),
            onToggleStar: jest.fn(),
            onOpenMenu: jest.fn(),
            onContextMenu: jest.fn(),
        };
        renderWithoutBackend(
            <ul>
                <DocumentCard document={document} interaction={interaction} />
            </ul>
        );

        fireEvent.click(screen.getByRole('checkbox'));
        fireEvent.click(screen.getByRole('button', { name: 'FILES_UNFAVORITE_ITEM' }));
        fireEvent.click(screen.getByRole('button', { name: 'FILES_ITEM_MENU' }));
        fireEvent.contextMenu(screen.getByRole('listitem'));

        expect(interaction.onToggleSelection).toHaveBeenCalledTimes(1);
        expect(interaction.onToggleStar).toHaveBeenCalledTimes(1);
        expect(interaction.onOpenMenu).toHaveBeenCalledTimes(1);
        expect(interaction.onContextMenu).toHaveBeenCalledTimes(1);
    });
});
