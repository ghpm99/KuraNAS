import { fireEvent, screen, within } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import VideoClassificationDialog from './VideoClassificationDialog';
import VideoConfirmDialog from './VideoConfirmDialog';
import VideoCorrectionMenu from './VideoCorrectionMenu';
import VideoEpisodeDialog from './VideoEpisodeDialog';
import VideoHideSeriesButton from './VideoHideSeriesButton';
import VideoSeriesWatchedActions from './VideoSeriesWatchedActions';

describe('video correction components without a backend', () => {
    it('renders the classification dialog and survives a failing save', async () => {
        renderWithoutBackend(
            <VideoClassificationDialog isOpen fileId={7} fileName="a.mkv" onClose={jest.fn()} />
        );

        fireEvent.click(screen.getByRole('button', { name: 'VIDEO_CORRECTION_SAVE' }));

        expect(await screen.findByRole('alert')).toBeInTheDocument();
    });

    it('renders the episode dialog with the current numbers and survives a failing save', async () => {
        renderWithoutBackend(
            <VideoEpisodeDialog
                isOpen
                fileId={7}
                fileName="a.mkv"
                initialSeason={2}
                initialEpisode={3}
                onClose={jest.fn()}
            />
        );

        expect(screen.getByLabelText('VIDEO_CORRECTION_SEASON_LABEL')).toHaveValue(2);
        expect(screen.getByLabelText('VIDEO_CORRECTION_EPISODE_LABEL')).toHaveValue(3);
        fireEvent.click(screen.getByRole('button', { name: 'VIDEO_CORRECTION_SAVE' }));

        expect(await screen.findByRole('alert')).toBeInTheDocument();
    });

    it('opens the right dialog from the menu', () => {
        renderWithoutBackend(<VideoCorrectionMenu fileId={7} fileName="a.mkv" canCorrectEpisode />);

        fireEvent.click(screen.getByRole('button'));
        fireEvent.click(screen.getByRole('menuitem', { name: 'VIDEO_CORRECT_EPISODE' }));

        expect(screen.getByRole('dialog')).toBeInTheDocument();
    });

    it('offers only the classification correction when the episode cannot be corrected', () => {
        renderWithoutBackend(<VideoCorrectionMenu fileId={7} fileName="a.mkv" />);

        fireEvent.click(screen.getByRole('button'));

        expect(
            screen.getByRole('menuitem', { name: 'VIDEO_CORRECT_CLASSIFICATION' })
        ).toBeInTheDocument();
        expect(screen.queryByRole('menuitem', { name: 'VIDEO_CORRECT_EPISODE' })).toBeNull();
    });

    it('renders the confirm dialog and confirms', () => {
        const onConfirm = jest.fn();
        renderWithoutBackend(
            <VideoConfirmDialog
                isOpen
                title="Titulo"
                message="Mensagem"
                confirmLabel="Sim"
                onConfirm={onConfirm}
                onClose={jest.fn()}
            />
        );

        fireEvent.click(screen.getByRole('button', { name: 'Sim' }));

        expect(onConfirm).toHaveBeenCalled();
    });

    it('renders the series actions without seasons and the hide button without handler', () => {
        const { container } = renderWithoutBackend(
            <>
                <VideoSeriesWatchedActions playlistId={1} playlistName="S" seasons={[]} />
                <VideoHideSeriesButton seriesName="S" isChanging={false} />
            </>
        );

        expect(container).toBeEmptyDOMElement();
    });

    it('confirms before hiding the series', () => {
        const onHide = jest.fn();
        renderWithoutBackend(
            <VideoHideSeriesButton seriesName="S" isChanging={false} onHide={onHide} />
        );

        fireEvent.click(screen.getByRole('button', { name: 'VIDEO_HIDE_SERIES' }));
        expect(onHide).not.toHaveBeenCalled();
        fireEvent.click(
            within(screen.getByRole('dialog')).getByRole('button', { name: 'VIDEO_HIDE_SERIES' })
        );

        expect(onHide).toHaveBeenCalledTimes(1);
    });
});
