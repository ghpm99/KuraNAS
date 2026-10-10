import { fireEvent, render, screen } from '@testing-library/react';
import VideoSectionError from './VideoSectionError';

describe('VideoSectionError', () => {
    it('renders without any provider or service', () => {
        render(<VideoSectionError sectionTitleKey="VIDEO_ALL" failure={{ retry: jest.fn() }} />);

        expect(screen.getByRole('alert')).toBeInTheDocument();
    });

    it('shows the backend message verbatim and retries on click', () => {
        const retry = jest.fn();
        render(
            <VideoSectionError
                sectionTitleKey="VIDEO_ALL"
                failure={{ message: 'disk offline', retry }}
            />
        );

        expect(screen.getByText('disk offline')).toBeInTheDocument();
        fireEvent.click(screen.getByRole('button'));
        expect(retry).toHaveBeenCalledTimes(1);
    });
});
