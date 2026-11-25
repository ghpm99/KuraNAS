import { fireEvent, render, screen } from '@testing-library/react';
import VacuumMapCleaningActions from './VacuumMapCleaningActions';

jest.mock('@/components/i18n/provider/i18nContext', () =>
    require('./testSupport/interpolatingI18nMock').interpolatingI18nMock
);

const baseProps = {
    touchMode: 'select-rooms' as const,
    selectedRoomCount: 0,
    hasSpotPoint: false,
    isSpotRejected: false,
    canStartCleaning: true,
    isSending: false,
    onCleanRooms: jest.fn(),
    onCleanSpot: jest.fn(),
};

describe('VacuumMapCleaningActions', () => {
    beforeEach(() => jest.clearAllMocks());

    it('renders without any provider or service', () => {
        expect(() => render(<VacuumMapCleaningActions {...baseProps} />)).not.toThrow();
    });

    it('hides the clean selected button without a selection and prompts for one', () => {
        render(<VacuumMapCleaningActions {...baseProps} />);

        expect(screen.queryByRole('button')).not.toBeInTheDocument();
        expect(screen.getByText('VACUUM_MAP_ROOMS_PROMPT')).toBeInTheDocument();
    });

    it('shows the quantity on the clean selected button and sends on click', () => {
        render(<VacuumMapCleaningActions {...baseProps} selectedRoomCount={2} />);

        fireEvent.click(screen.getByRole('button', { name: 'VACUUM_MAP_CLEAN_SELECTED:2' }));

        expect(baseProps.onCleanRooms).toHaveBeenCalledTimes(1);
    });

    it('prompts for a point in spot mode and shows no confirm button yet', () => {
        render(<VacuumMapCleaningActions {...baseProps} touchMode="clean-spot" />);

        expect(screen.getByText('VACUUM_MAP_SPOT_PROMPT')).toBeInTheDocument();
        expect(screen.queryByRole('button', { name: 'VACUUM_MAP_CLEAN_SPOT' })).not.toBeInTheDocument();
    });

    it('warns when the tapped point is not on the floor', () => {
        render(<VacuumMapCleaningActions {...baseProps} touchMode="clean-spot" isSpotRejected />);

        expect(screen.getByText('VACUUM_MAP_SPOT_REJECTED')).toBeInTheDocument();
    });

    it('confirms the marked point with the clean here button', () => {
        render(<VacuumMapCleaningActions {...baseProps} touchMode="clean-spot" hasSpotPoint />);

        expect(screen.getByText('VACUUM_MAP_SPOT_CONFIRM')).toBeInTheDocument();
        fireEvent.click(screen.getByRole('button', { name: 'VACUUM_MAP_CLEAN_SPOT' }));

        expect(baseProps.onCleanSpot).toHaveBeenCalledTimes(1);
    });

    it('disables the buttons while sending', () => {
        render(<VacuumMapCleaningActions {...baseProps} selectedRoomCount={1} isSending />);

        expect(screen.getByRole('button', { name: 'VACUUM_MAP_CLEAN_SELECTED:1' })).toBeDisabled();
    });

    it('disables the buttons and explains when the robot cannot start a cleaning', () => {
        render(<VacuumMapCleaningActions {...baseProps} selectedRoomCount={1} canStartCleaning={false} />);

        expect(screen.getByRole('button', { name: 'VACUUM_MAP_CLEAN_SELECTED:1' })).toBeDisabled();
        expect(screen.getByText('VACUUM_MAP_ACTIONS_UNAVAILABLE')).toBeInTheDocument();
    });
});
