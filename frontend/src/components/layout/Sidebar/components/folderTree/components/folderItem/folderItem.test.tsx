import { fireEvent, render, screen } from '@testing-library/react';
import FolderItem from './index';

describe('folderItem', () => {
    it('renders folder and file labels with truncation and children expansion', () => {
        const onClick = jest.fn();
        const { rerender } = render(
            <FolderItem
                label="verylongfilename.mp3"
                type={2}
                onClick={onClick}
                expanded={false}
                selected={false}
            >
                <div>child</div>
            </FolderItem>
        );
        expect(screen.getByText(/\.mp3$/)).toBeInTheDocument();
        fireEvent.click(screen.getByRole('button'));
        expect(onClick).toHaveBeenCalled();
        expect(screen.queryByText('child')).not.toBeInTheDocument();

        rerender(
            <FolderItem label="Folder" type={1} onClick={onClick} expanded selected>
                <div>child</div>
            </FolderItem>
        );
        expect(screen.getByText('Folder')).toBeInTheDocument();
        expect(screen.getByText('child')).toBeInTheDocument();
    });
});
