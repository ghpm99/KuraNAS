import { fireEvent, screen } from '@testing-library/react';
import FileProvider from '@/features/files/providers/fileProvider';
import { createTestFile } from '@/features/files/selection/testFileFactory';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import DocumentItemMenu from './DocumentItemMenu';

const renderMenu = (props: Partial<React.ComponentProps<typeof DocumentItemMenu>> = {}) =>
    renderWithoutBackend(
        <FileProvider>
            <DocumentItemMenu
                anchorPosition={{ top: 10, left: 10 }}
                targetFiles={[createTestFile(1, { name: 'a.pdf' })]}
                onClose={jest.fn()}
                {...props}
            />
        </FileProvider>
    );

describe('DocumentItemMenu', () => {
    it('renders without any backend, closed and open', () => {
        expect(() => renderMenu({ anchorPosition: null, targetFiles: [] })).not.toThrow();
        expect(screen.queryByRole('menu')).toBeNull();
        expect(() => renderMenu()).not.toThrow();
    });

    it('disables rename and details for several targets', async () => {
        renderMenu({ targetFiles: [createTestFile(1), createTestFile(2)] });

        expect(await screen.findByRole('menuitem', { name: 'RENAME' })).toHaveAttribute(
            'aria-disabled',
            'true'
        );
        expect(screen.queryByRole('menuitem', { name: 'FILES_DETAILS' })).toBeNull();
    });

    it('closes when an action is chosen', async () => {
        const onClose = jest.fn();
        renderMenu({ onClose });

        fireEvent.click(await screen.findByRole('menuitem', { name: 'MOVE' }));

        expect(onClose).toHaveBeenCalled();
    });
});
