import { fireEvent, render, screen } from '@testing-library/react';
import ActionBarButton from './actionBarButton';

describe('ActionBarButton', () => {
    it('renders a labelled button without any provider or mock', () => {
        render(
            <ActionBarButton
                label="Upload"
                icon={<i aria-hidden="true" />}
                onClick={jest.fn()}
                isIconOnly={false}
            />
        );

        expect(screen.getByRole('button', { name: 'Upload' })).toBeInTheDocument();
        expect(screen.getByText('Upload')).toBeInTheDocument();
    });

    it('renders only an icon with an accessible name when icon-only', () => {
        const onClick = jest.fn();
        render(
            <ActionBarButton
                label="Upload"
                icon={<i aria-hidden="true" />}
                onClick={onClick}
                isIconOnly
            />
        );

        expect(screen.queryByText('Upload')).toBeNull();
        fireEvent.click(screen.getByRole('button', { name: 'Upload' }));
        expect(onClick).toHaveBeenCalledTimes(1);
    });

    it('supports primary and destructive variants', () => {
        render(
            <>
                <ActionBarButton label="Go" icon={null} onClick={jest.fn()} isIconOnly={false} isPrimary />
                <ActionBarButton
                    label="Remove"
                    icon={null}
                    onClick={jest.fn()}
                    isIconOnly={false}
                    isDestructive
                />
            </>
        );

        expect(screen.getByRole('button', { name: 'Go' })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'Remove' })).toBeInTheDocument();
    });
});
