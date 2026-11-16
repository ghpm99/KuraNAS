import { fireEvent, screen, waitFor } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import { getVacuumPreferences, updateVacuumPreferences } from '@/service/vacuum';
import type { VacuumPreferences } from '@/types/vacuum';
import { createServerFailure } from './testSupport/createServerFailure';
import { renderWithVacuumProviders } from './testSupport/renderWithVacuumProviders';
import VacuumPreferencesBlock from './VacuumPreferencesBlock';

jest.mock('@/components/i18n/provider/i18nContext', () =>
    require('./testSupport/interpolatingI18nMock').interpolatingI18nMock
);
jest.mock('@/service/vacuum', () => ({
    ...jest.requireActual('@/service/vacuum'),
    getVacuumPreferences: jest.fn(),
    updateVacuumPreferences: jest.fn(),
}));

const mockedGetPreferences = getVacuumPreferences as jest.Mock;
const mockedUpdatePreferences = updateVacuumPreferences as jest.Mock;

const fullPreferences: VacuumPreferences = {
    suction: 'auto',
    water_level: 'mid',
    volume: 4,
    is_do_not_disturb: false,
    is_carpet_boost: true,
    is_resume_after_charge: false,
    is_auto_empty: true,
    auto_empty_frequency: '2',
};

const allNullPreferences: VacuumPreferences = {
    suction: null,
    water_level: null,
    volume: null,
    is_do_not_disturb: null,
    is_carpet_boost: null,
    is_resume_after_charge: null,
    is_auto_empty: null,
    auto_empty_frequency: null,
};

describe('VacuumPreferencesBlock', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedGetPreferences.mockRejectedValue(new Error('backend absent'));
    });

    it('mounts without a backend', () => {
        expect(() => renderWithoutBackend(<VacuumPreferencesBlock />)).not.toThrow();
    });

    it('shows the load error with the server message', async () => {
        mockedGetPreferences.mockRejectedValue(createServerFailure(502, 'robot offline'));

        renderWithVacuumProviders(<VacuumPreferencesBlock />);

        expect(await screen.findByText('robot offline')).toBeInTheDocument();
    });

    it('renders every control the robot reports', async () => {
        mockedGetPreferences.mockResolvedValue(fullPreferences);

        renderWithVacuumProviders(<VacuumPreferencesBlock />);

        expect(await screen.findByText('VACUUM_PREFERENCE_SUCTION')).toBeInTheDocument();
        expect(screen.getByText('VACUUM_PREFERENCE_WATER_LEVEL')).toBeInTheDocument();
        expect(screen.getByRole('switch', { name: 'VACUUM_PREFERENCE_CARPET_BOOST' })).toBeChecked();
        expect(
            screen.getByRole('switch', { name: 'VACUUM_PREFERENCE_RESUME_AFTER_CHARGE' })
        ).not.toBeChecked();
        expect(screen.getByRole('slider', { name: 'VACUUM_PREFERENCE_VOLUME' })).toHaveValue('4');
        expect(screen.getByRole('switch', { name: 'VACUUM_PREFERENCE_DO_NOT_DISTURB' })).not.toBeChecked();
        expect(screen.getByRole('switch', { name: 'VACUUM_PREFERENCE_AUTO_EMPTY' })).toBeChecked();
        expect(screen.getByText('VACUUM_PREFERENCE_AUTO_EMPTY_FREQUENCY')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: '2' })).toHaveAttribute('aria-pressed', 'true');
        expect(screen.queryByText('VACUUM_PREFERENCES_EMPTY')).not.toBeInTheDocument();
    });

    it('does not render controls whose field is null', async () => {
        mockedGetPreferences.mockResolvedValue({ ...allNullPreferences, volume: 3 });

        renderWithVacuumProviders(<VacuumPreferencesBlock />);

        expect(await screen.findByRole('slider', { name: 'VACUUM_PREFERENCE_VOLUME' })).toBeInTheDocument();
        expect(screen.queryByText('VACUUM_PREFERENCE_SUCTION')).not.toBeInTheDocument();
        expect(screen.queryByText('VACUUM_PREFERENCE_WATER_LEVEL')).not.toBeInTheDocument();
        expect(screen.queryByRole('switch')).not.toBeInTheDocument();
        expect(screen.queryByText('VACUUM_PREFERENCE_AUTO_EMPTY_FREQUENCY')).not.toBeInTheDocument();
    });

    it('says the robot reported no setting when every field is null', async () => {
        mockedGetPreferences.mockResolvedValue(allNullPreferences);

        renderWithVacuumProviders(<VacuumPreferencesBlock />);

        expect(await screen.findByText('VACUUM_PREFERENCES_EMPTY')).toBeInTheDocument();
        expect(screen.queryByRole('switch')).not.toBeInTheDocument();
    });

    it('survives an empty payload', async () => {
        mockedGetPreferences.mockResolvedValue({});

        renderWithVacuumProviders(<VacuumPreferencesBlock />);

        expect(await screen.findByText('VACUUM_PREFERENCES_EMPTY')).toBeInTheDocument();
    });

    it('patches only the changed field and shows what the robot confirmed', async () => {
        mockedGetPreferences.mockResolvedValue(fullPreferences);
        mockedUpdatePreferences.mockResolvedValue({ ...fullPreferences, suction: 'max' });
        renderWithVacuumProviders(<VacuumPreferencesBlock />);
        await screen.findByText('VACUUM_PREFERENCE_SUCTION');

        fireEvent.click(screen.getByRole('button', { name: 'VACUUM_SUCTION_MAX' }));

        await waitFor(() => expect(mockedUpdatePreferences).toHaveBeenCalledWith({ suction: 'max' }));
        expect(mockedUpdatePreferences).toHaveBeenCalledTimes(1);
        await waitFor(() =>
            expect(screen.getByRole('button', { name: 'VACUUM_SUCTION_MAX' })).toHaveAttribute(
                'aria-pressed',
                'true'
            )
        );
        expect(await screen.findByText('VACUUM_PREFERENCE_SAVED')).toBeInTheDocument();
    });

    it('shows the value the robot returned even when it differs from the one requested', async () => {
        mockedGetPreferences.mockResolvedValue(fullPreferences);
        mockedUpdatePreferences.mockResolvedValue({ ...fullPreferences, water_level: 'low' });
        renderWithVacuumProviders(<VacuumPreferencesBlock />);
        await screen.findByText('VACUUM_PREFERENCE_WATER_LEVEL');

        fireEvent.click(screen.getByRole('button', { name: 'VACUUM_WATER_HIGH' }));

        await waitFor(() =>
            expect(screen.getByRole('button', { name: 'VACUUM_WATER_LOW' })).toHaveAttribute(
                'aria-pressed',
                'true'
            )
        );
    });

    it('patches a switch with its own field only', async () => {
        mockedGetPreferences.mockResolvedValue(fullPreferences);
        mockedUpdatePreferences.mockResolvedValue({ ...fullPreferences, is_carpet_boost: false });
        renderWithVacuumProviders(<VacuumPreferencesBlock />);

        fireEvent.click(await screen.findByRole('switch', { name: 'VACUUM_PREFERENCE_CARPET_BOOST' }));

        await waitFor(() =>
            expect(mockedUpdatePreferences).toHaveBeenCalledWith({ is_carpet_boost: false })
        );
    });

    it('patches the auto empty frequency as the text the contract uses', async () => {
        mockedGetPreferences.mockResolvedValue(fullPreferences);
        mockedUpdatePreferences.mockResolvedValue({ ...fullPreferences, auto_empty_frequency: '3' });
        renderWithVacuumProviders(<VacuumPreferencesBlock />);

        fireEvent.click(await screen.findByRole('button', { name: '3' }));

        await waitFor(() =>
            expect(mockedUpdatePreferences).toHaveBeenCalledWith({ auto_empty_frequency: '3' })
        );
    });

    it('sends the volume only when the control is released', async () => {
        mockedGetPreferences.mockResolvedValue(fullPreferences);
        mockedUpdatePreferences.mockResolvedValue({ ...fullPreferences, volume: 7 });
        const { container } = renderWithVacuumProviders(<VacuumPreferencesBlock />);
        await screen.findByText('VACUUM_PREFERENCE_AUTO_EMPTY_FREQUENCY');
        const sliderTrack = container.querySelector('.MuiSlider-root') as HTMLElement;
        sliderTrack.getBoundingClientRect = () =>
            ({ left: 0, width: 100, top: 0, height: 10, right: 100, bottom: 10 }) as DOMRect;

        fireEvent.mouseDown(sliderTrack, { clientX: 70, buttons: 1 });
        expect(mockedUpdatePreferences).not.toHaveBeenCalled();

        fireEvent.mouseUp(document);
        await waitFor(() => expect(mockedUpdatePreferences).toHaveBeenCalledWith({ volume: 7 }));
        expect(mockedUpdatePreferences).toHaveBeenCalledTimes(1);
    });

    it('shows the robot mop-only suction as a read-only current value and does not offer it', async () => {
        mockedGetPreferences.mockResolvedValue({ ...fullPreferences, suction: 'mop' });

        renderWithVacuumProviders(<VacuumPreferencesBlock />);

        expect(
            await screen.findByText('VACUUM_PREFERENCE_CURRENT_READ_ONLY:VACUUM_SUCTION_MOP')
        ).toBeInTheDocument();
        expect(screen.queryByRole('button', { name: 'VACUUM_SUCTION_MOP' })).not.toBeInTheDocument();
    });

    it('shows any other unlisted value as unknown', async () => {
        mockedGetPreferences.mockResolvedValue({ ...fullPreferences, water_level: 'turbo' });

        renderWithVacuumProviders(<VacuumPreferencesBlock />);

        expect(
            await screen.findByText('VACUUM_PREFERENCE_CURRENT_READ_ONLY:VACUUM_VALUE_UNKNOWN')
        ).toBeInTheDocument();
    });

    it('shows the server message and keeps the confirmed value when the change is refused', async () => {
        mockedGetPreferences.mockResolvedValue(fullPreferences);
        mockedUpdatePreferences.mockRejectedValue(createServerFailure(400, 'invalid setting'));
        renderWithVacuumProviders(<VacuumPreferencesBlock />);
        await screen.findByText('VACUUM_PREFERENCE_SUCTION');

        fireEvent.click(screen.getByRole('button', { name: 'VACUUM_SUCTION_MAX' }));

        expect(await screen.findByText('invalid setting')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'VACUUM_SUCTION_AUTO' })).toHaveAttribute(
            'aria-pressed',
            'true'
        );
    });

    it('disables only the control being saved', async () => {
        mockedGetPreferences.mockResolvedValue(fullPreferences);
        mockedUpdatePreferences.mockReturnValue(new Promise(() => undefined));
        renderWithVacuumProviders(<VacuumPreferencesBlock />);
        await screen.findByText('VACUUM_PREFERENCE_SUCTION');

        fireEvent.click(screen.getByRole('button', { name: 'VACUUM_SUCTION_MAX' }));

        await waitFor(() =>
            expect(screen.getByRole('button', { name: 'VACUUM_SUCTION_MAX' })).toBeDisabled()
        );
        expect(screen.getByRole('button', { name: 'VACUUM_WATER_LOW' })).toBeEnabled();
        expect(screen.getByRole('switch', { name: 'VACUUM_PREFERENCE_CARPET_BOOST' })).toBeEnabled();
    });
});
