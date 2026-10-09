import { fireEvent, render, screen } from '@testing-library/react';
import { expectRendersWithoutBackend } from '@/shared/test/renderWithoutBackend';
import FileSearchModeTabs from './FileSearchModeTabs';

describe('FileSearchModeTabs', () => {
    it('renders without props and without a backend', () => {
        expectRendersWithoutBackend(<FileSearchModeTabs />);
    });

    it('marks the current mode and reports a switch to content', () => {
        const onChange = jest.fn();
        render(<FileSearchModeTabs mode="name" onChange={onChange} />);

        expect(screen.getByRole('tab', { name: 'FILES_SEARCH_MODE_NAME' })).toHaveAttribute(
            'aria-selected',
            'true'
        );

        fireEvent.click(screen.getByRole('tab', { name: 'FILES_SEARCH_MODE_CONTENT' }));

        expect(onChange).toHaveBeenCalledWith('content');
    });
});
