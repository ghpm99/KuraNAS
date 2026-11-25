import { act, fireEvent, screen, waitFor } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import {
    getVacuumCloudSettings,
    getVacuumMap,
    getVacuumPath,
    getVacuumRooms,
    getVacuumStatus,
    renameVacuumRoom,
    startVacuumRoomCleaning,
    startVacuumSpotCleaning,
} from '@/service/vacuum';
import { buildVacuumMapPayload } from './testSupport/buildVacuumMapPayload';
import { createFakeCanvasContext } from './testSupport/createFakeCanvasContext';
import { createServerFailure } from './testSupport/createServerFailure';
import { mockCanvasGetContext } from './testSupport/mockCanvasGetContext';
import { mockElementBounds } from './testSupport/mockElementBounds';
import { renderWithVacuumProviders } from './testSupport/renderWithVacuumProviders';
import {
    configuredCloudSettings,
    emptyPath,
    mapRooms,
} from './testSupport/vacuumMapServiceFixtures';
import {
    frameHeightPixels,
    frameWidthPixels,
    tapOnObstacle,
    tapOnRoomOne,
    tapOnRoomTwo,
    tapOnWall,
} from './testSupport/vacuumMapTaps';
import VacuumMapBlock from './VacuumMapBlock';
import { vacuumQueryKeys } from './vacuumQueryKeys';

jest.mock('@/components/i18n/provider/i18nContext', () =>
    require('./testSupport/interpolatingI18nMock').interpolatingI18nMock
);
jest.mock('@/service/vacuum', () => ({
    ...jest.requireActual('@/service/vacuum'),
    getVacuumStatus: jest.fn(),
    getVacuumCloudSettings: jest.fn(),
    getVacuumMap: jest.fn(),
    getVacuumRooms: jest.fn(),
    getVacuumPath: jest.fn(),
    renameVacuumRoom: jest.fn(),
    startVacuumRoomCleaning: jest.fn(),
    startVacuumSpotCleaning: jest.fn(),
}));

const mockedStatus = getVacuumStatus as jest.Mock;
const mockedCloudSettings = getVacuumCloudSettings as jest.Mock;
const mockedMap = getVacuumMap as jest.Mock;
const mockedRooms = getVacuumRooms as jest.Mock;
const mockedPath = getVacuumPath as jest.Mock;
const mockedRename = renameVacuumRoom as jest.Mock;
const mockedRoomCleaning = startVacuumRoomCleaning as jest.Mock;
const mockedSpotCleaning = startVacuumSpotCleaning as jest.Mock;

const canvas = () => screen.getByRole('img', { name: 'VACUUM_MAP_CANVAS_LABEL' });

const renderLoadedBlock = async (state = 'idle') => {
    mockedStatus.mockResolvedValue({ state });
    const rendered = renderWithVacuumProviders(<VacuumMapBlock />);
    await screen.findByRole('img', { name: 'VACUUM_MAP_CANVAS_LABEL' });
    return rendered;
};

describe('VacuumMapBlock', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedStatus.mockRejectedValue(new Error('backend absent'));
        mockedCloudSettings.mockResolvedValue(configuredCloudSettings);
        mockedMap.mockResolvedValue(buildVacuumMapPayload());
        mockedRooms.mockResolvedValue(mapRooms);
        mockedPath.mockResolvedValue(emptyPath);
        mockCanvasGetContext(createFakeCanvasContext());
        mockElementBounds(frameWidthPixels, frameHeightPixels);
    });

    afterEach(() => jest.restoreAllMocks());

    it('mounts without a backend', () => {
        expect(() => renderWithoutBackend(<VacuumMapBlock />)).not.toThrow();
    });

    it('survives empty cloud settings by explaining what is missing', async () => {
        mockedCloudSettings.mockResolvedValue({});

        renderWithVacuumProviders(<VacuumMapBlock />);

        expect(
            await screen.findByText('VACUUM_MAP_CLOUD_NOT_CONFIGURED_DESCRIPTION')
        ).toBeInTheDocument();
        expect(mockedMap).not.toHaveBeenCalled();
    });

    it('survives empty map, rooms and path payloads', async () => {
        mockedMap.mockResolvedValue({});
        mockedRooms.mockResolvedValue({});
        mockedPath.mockResolvedValue({});

        renderWithVacuumProviders(<VacuumMapBlock />);

        expect(await screen.findByText('VACUUM_MAP_EMPTY')).toBeInTheDocument();
    });

    it('survives a map without cells', async () => {
        mockedMap.mockResolvedValue(buildVacuumMapPayload({ cells: undefined }));

        renderWithVacuumProviders(<VacuumMapBlock />);

        expect(await screen.findByText('VACUUM_MAP_EMPTY')).toBeInTheDocument();
    });

    it('survives rooms and path points that are not lists', async () => {
        mockedRooms.mockResolvedValue({ map_id: 77, items: null });
        mockedPath.mockResolvedValue({ points: 'x', position: 'y' });

        renderWithVacuumProviders(<VacuumMapBlock />);

        expect(await screen.findByText('VACUUM_ROOMS_EMPTY')).toBeInTheDocument();
    });

    it('shows the loading state while the cloud settings are being read', () => {
        mockedCloudSettings.mockReturnValue(new Promise(() => undefined));

        renderWithVacuumProviders(<VacuumMapBlock />);

        expect(screen.getByRole('status')).toHaveTextContent('LOADING');
    });

    it('links to the cloud section when the credentials are missing', async () => {
        mockedCloudSettings.mockResolvedValue({ access_id: 'abc', region: 'eu', has_access_secret: false });

        renderWithVacuumProviders(<VacuumMapBlock />);

        expect(
            await screen.findByRole('link', { name: 'VACUUM_MAP_CLOUD_NOT_CONFIGURED_ACTION' })
        ).toHaveAttribute('href', expect.stringContaining('#vacuum-cloud'));
    });

    it('shows the server message with a retry when the cloud fails and retries only the map', async () => {
        mockedMap.mockRejectedValueOnce(createServerFailure(502, 'cloud is down'));
        renderWithVacuumProviders(<VacuumMapBlock />);
        expect(await screen.findByText('cloud is down')).toBeInTheDocument();

        fireEvent.click(screen.getByRole('button', { name: 'TRY_AGAIN' }));

        await screen.findByRole('img', { name: 'VACUUM_MAP_CANVAS_LABEL' });
        expect(mockedMap).toHaveBeenCalledTimes(2);
        expect(mockedCloudSettings).toHaveBeenCalledTimes(1);
    });

    it('shows an error when the rooms fail and retries them', async () => {
        mockedRooms.mockRejectedValueOnce(createServerFailure(502, 'rooms down'));
        renderWithVacuumProviders(<VacuumMapBlock />);
        expect(await screen.findByText('rooms down')).toBeInTheDocument();

        fireEvent.click(screen.getByRole('button', { name: 'TRY_AGAIN' }));

        await screen.findByRole('img', { name: 'VACUUM_MAP_CANVAS_LABEL' });
        expect(mockedRooms).toHaveBeenCalledTimes(2);
    });

    it('retries the cloud settings when they fail to load', async () => {
        mockedCloudSettings.mockRejectedValueOnce(createServerFailure(500, 'settings down'));
        renderWithVacuumProviders(<VacuumMapBlock />);
        expect(await screen.findByText('settings down')).toBeInTheDocument();

        fireEvent.click(screen.getByRole('button', { name: 'TRY_AGAIN' }));

        await screen.findByRole('img', { name: 'VACUUM_MAP_CANVAS_LABEL' });
    });

    it('keeps showing the last map when a refresh fails', async () => {
        const { queryClient } = await renderLoadedBlock();
        mockedMap.mockRejectedValue(createServerFailure(502, 'cloud is down'));

        await act(() => queryClient.invalidateQueries({ queryKey: vacuumQueryKeys.map }));

        expect(canvas()).toBeInTheDocument();
        expect(screen.queryByText('cloud is down')).not.toBeInTheDocument();
    });

    it('lists the rooms with names, numbered fallback and area', async () => {
        await renderLoadedBlock();

        expect(screen.getByText('VACUUM_ROOM_DEFAULT_NAME:1')).toBeInTheDocument();
        expect(screen.getByText('Cozinha')).toBeInTheDocument();
        expect(screen.getByText('VACUUM_ROOM_AREA:3.8')).toBeInTheDocument();
    });

    it('toggles a room by tapping on its floor and ignores walls, obstacles and the unknown', async () => {
        await renderLoadedBlock();
        const [firstCheckbox, secondCheckbox] = screen.getAllByRole('checkbox');

        fireEvent.click(canvas(), tapOnRoomTwo);
        expect(secondCheckbox).toBeChecked();
        expect(screen.getByRole('button', { name: 'VACUUM_MAP_CLEAN_SELECTED:1' })).toBeInTheDocument();

        fireEvent.click(canvas(), tapOnWall);
        fireEvent.click(canvas(), tapOnObstacle);
        expect(firstCheckbox).not.toBeChecked();
        expect(secondCheckbox).toBeChecked();

        fireEvent.click(canvas(), tapOnRoomTwo);
        expect(secondCheckbox).not.toBeChecked();
        expect(screen.queryByRole('button', { name: /VACUUM_MAP_CLEAN_SELECTED/ })).not.toBeInTheDocument();
    });

    it('keeps the list and the map selection in sync', async () => {
        await renderLoadedBlock();

        fireEvent.click(screen.getAllByRole('checkbox')[0] as HTMLElement);

        expect(screen.getByRole('button', { name: 'VACUUM_MAP_CLEAN_SELECTED:1' })).toBeInTheDocument();
        fireEvent.click(canvas(), tapOnRoomOne);
        expect(screen.getAllByRole('checkbox')[0]).not.toBeChecked();
    });

    it('sends the selected rooms with the map id, reports success and clears the selection', async () => {
        mockedRoomCleaning.mockResolvedValue(undefined);
        await renderLoadedBlock();
        fireEvent.click(canvas(), tapOnRoomOne);
        fireEvent.click(canvas(), tapOnRoomTwo);

        fireEvent.click(screen.getByRole('button', { name: 'VACUUM_MAP_CLEAN_SELECTED:2' }));

        await waitFor(() =>
            expect(mockedRoomCleaning).toHaveBeenCalledWith({ map_id: 77, room_ids: [1, 2] })
        );
        expect(await screen.findByText('VACUUM_MAP_CLEANING_SENT')).toBeInTheDocument();
        await waitFor(() => expect(mockedStatus.mock.calls.length).toBeGreaterThan(1));
        await waitFor(() => expect(mockedPath.mock.calls.length).toBeGreaterThan(1));
        expect(screen.queryByRole('button', { name: /VACUUM_MAP_CLEAN_SELECTED/ })).not.toBeInTheDocument();
    });

    it('disables the button while sending and ignores a double click', async () => {
        let finishCleaning: () => void = () => undefined;
        mockedRoomCleaning.mockReturnValue(
            new Promise<void>((resolve) => {
                finishCleaning = resolve;
            })
        );
        await renderLoadedBlock();
        fireEvent.click(canvas(), tapOnRoomOne);
        const cleanButton = screen.getByRole('button', { name: 'VACUUM_MAP_CLEAN_SELECTED:1' });

        fireEvent.click(cleanButton);
        fireEvent.click(cleanButton);

        await waitFor(() => expect(cleanButton).toBeDisabled());
        expect(mockedRoomCleaning).toHaveBeenCalledTimes(1);
        finishCleaning();
        await waitFor(() => expect(mockedRoomCleaning).toHaveBeenCalledTimes(1));
    });

    it('shows the server message and keeps the selection when the robot refuses', async () => {
        mockedRoomCleaning.mockRejectedValue(createServerFailure(502, 'robot unreachable'));
        await renderLoadedBlock();
        fireEvent.click(canvas(), tapOnRoomOne);

        fireEvent.click(screen.getByRole('button', { name: 'VACUUM_MAP_CLEAN_SELECTED:1' }));

        expect(await screen.findByText('robot unreachable')).toBeInTheDocument();
        expect(screen.getAllByRole('checkbox')[0]).toBeChecked();
    });

    it('reloads the map and rooms and clears the selection when the map changed', async () => {
        mockedRoomCleaning.mockRejectedValue(createServerFailure(409, 'map changed'));
        await renderLoadedBlock();
        fireEvent.click(canvas(), tapOnRoomOne);
        mockedMap.mockClear();
        mockedRooms.mockClear();

        fireEvent.click(screen.getByRole('button', { name: 'VACUUM_MAP_CLEAN_SELECTED:1' }));

        expect(await screen.findByText('map changed')).toBeInTheDocument();
        await waitFor(() => expect(mockedMap).toHaveBeenCalledTimes(1));
        await waitFor(() => expect(mockedRooms).toHaveBeenCalledTimes(1));
        expect(screen.getAllByRole('checkbox')[0]).not.toBeChecked();
    });

    it('drops a selection made on a previous map when the map id changes', async () => {
        const { queryClient } = await renderLoadedBlock();
        fireEvent.click(canvas(), tapOnRoomOne);
        mockedMap.mockResolvedValue(buildVacuumMapPayload({ map_id: 78 }));
        mockedRooms.mockResolvedValue({ ...mapRooms, map_id: 78 });

        await act(() => queryClient.invalidateQueries({ queryKey: vacuumQueryKeys.map }));
        await act(() => queryClient.invalidateQueries({ queryKey: vacuumQueryKeys.rooms }));

        await waitFor(() => expect(screen.getAllByRole('checkbox')[0]).not.toBeChecked());
    });

    it('disables the map commands in states where Start is not offered', async () => {
        await renderLoadedBlock('cleaning');
        await waitFor(() => expect(mockedStatus).toHaveBeenCalled());
        fireEvent.click(canvas(), tapOnRoomOne);

        await waitFor(() =>
            expect(screen.getByRole('button', { name: 'VACUUM_MAP_CLEAN_SELECTED:1' })).toBeDisabled()
        );
        expect(screen.getByText('VACUUM_MAP_ACTIONS_UNAVAILABLE')).toBeInTheDocument();
    });

    it('offers the map commands when the state is not known', async () => {
        mockedStatus.mockRejectedValue(new Error('offline'));
        renderWithVacuumProviders(<VacuumMapBlock />);
        await screen.findByRole('img', { name: 'VACUUM_MAP_CANVAS_LABEL' });
        fireEvent.click(canvas(), tapOnRoomOne);

        expect(screen.getByRole('button', { name: 'VACUUM_MAP_CLEAN_SELECTED:1' })).toBeEnabled();
    });

    describe('clean here', () => {
        const switchToSpotMode = () =>
            fireEvent.click(screen.getByRole('button', { name: 'VACUUM_MAP_MODE_CLEAN_SPOT' }));

        it('does not toggle rooms in clean here mode and prompts for a point', async () => {
            await renderLoadedBlock();
            switchToSpotMode();

            fireEvent.click(canvas(), tapOnRoomOne);

            expect(screen.getAllByRole('checkbox')[0]).not.toBeChecked();
            expect(screen.getByText('VACUUM_MAP_SPOT_CONFIRM')).toBeInTheDocument();
        });

        it('rejects a point that is not on the floor and shows the hint', async () => {
            await renderLoadedBlock();
            switchToSpotMode();

            fireEvent.click(canvas(), tapOnWall);

            expect(screen.getByText('VACUUM_MAP_SPOT_REJECTED')).toBeInTheDocument();
            expect(screen.queryByRole('button', { name: 'VACUUM_MAP_CLEAN_SPOT' })).not.toBeInTheDocument();
        });

        it('keeps the previous valid point when a later tap is rejected', async () => {
            await renderLoadedBlock();
            switchToSpotMode();
            fireEvent.click(canvas(), tapOnRoomOne);

            fireEvent.click(canvas(), tapOnWall);

            expect(screen.getByText('VACUUM_MAP_SPOT_REJECTED')).toBeInTheDocument();
            expect(screen.getByRole('button', { name: 'VACUUM_MAP_CLEAN_SPOT' })).toBeInTheDocument();
        });

        it('sends the confirmed point in millimeters with the map id', async () => {
            mockedSpotCleaning.mockResolvedValue(undefined);
            await renderLoadedBlock();
            switchToSpotMode();
            fireEvent.click(canvas(), tapOnRoomOne);

            fireEvent.click(screen.getByRole('button', { name: 'VACUUM_MAP_CLEAN_SPOT' }));

            await waitFor(() =>
                expect(mockedSpotCleaning).toHaveBeenCalledWith({ map_id: 77, x: -25, y: 25 })
            );
            expect(await screen.findByText('VACUUM_MAP_CLEANING_SENT')).toBeInTheDocument();
            await waitFor(() =>
                expect(screen.queryByRole('button', { name: 'VACUUM_MAP_CLEAN_SPOT' })).not.toBeInTheDocument()
            );
        });

        it('drops the marked point when going back to selecting rooms', async () => {
            await renderLoadedBlock();
            switchToSpotMode();
            fireEvent.click(canvas(), tapOnRoomOne);

            fireEvent.click(screen.getByRole('button', { name: 'VACUUM_MAP_MODE_SELECT_ROOMS' }));
            switchToSpotMode();

            expect(screen.queryByRole('button', { name: 'VACUUM_MAP_CLEAN_SPOT' })).not.toBeInTheDocument();
        });

        it('shows the server message when the spot command fails', async () => {
            mockedSpotCleaning.mockRejectedValue(createServerFailure(400, 'not on floor'));
            await renderLoadedBlock();
            switchToSpotMode();
            fireEvent.click(canvas(), tapOnRoomOne);

            fireEvent.click(screen.getByRole('button', { name: 'VACUUM_MAP_CLEAN_SPOT' }));

            expect(await screen.findByText('not on floor')).toBeInTheDocument();
        });
    });

    describe('renaming', () => {
        const openRenameDialog = async () => {
            fireEvent.click(screen.getByRole('button', { name: 'VACUUM_ROOM_RENAME_LABEL:VACUUM_ROOM_DEFAULT_NAME:1' }));
            return screen.findByLabelText('VACUUM_ROOM_RENAME_FIELD');
        };

        it('sends the new name with the current map id, closes and reloads the rooms', async () => {
            mockedRename.mockResolvedValue({ id: 1, name: 'Sala', area: 3.8 });
            await renderLoadedBlock();
            const field = await openRenameDialog();
            mockedRooms.mockClear();

            fireEvent.change(field, { target: { value: 'Sala' } });
            fireEvent.click(screen.getByRole('button', { name: 'VACUUM_ROOM_RENAME_SAVE' }));

            await waitFor(() => expect(mockedRename).toHaveBeenCalledWith(1, { map_id: 77, name: 'Sala' }));
            expect(await screen.findByText('VACUUM_ROOM_RENAMED')).toBeInTheDocument();
            await waitFor(() => expect(mockedRooms).toHaveBeenCalled());
            await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
        });

        it('shows the server message and keeps the dialog open on failure', async () => {
            mockedRename.mockRejectedValue(createServerFailure(400, 'name too long'));
            await renderLoadedBlock();
            await openRenameDialog();

            fireEvent.click(screen.getByRole('button', { name: 'VACUUM_ROOM_RENAME_SAVE' }));

            expect(await screen.findByText('name too long')).toBeInTheDocument();
            expect(screen.getByRole('dialog')).toBeInTheDocument();
        });

        it('reloads the map when the rename hits a map that changed', async () => {
            mockedRename.mockRejectedValue(createServerFailure(409, 'map changed'));
            await renderLoadedBlock();
            await openRenameDialog();
            mockedMap.mockClear();

            fireEvent.click(screen.getByRole('button', { name: 'VACUUM_ROOM_RENAME_SAVE' }));

            expect(await screen.findByText('map changed')).toBeInTheDocument();
            await waitFor(() => expect(mockedMap).toHaveBeenCalledTimes(1));
        });

        it('cancels without sending', async () => {
            await renderLoadedBlock();
            await openRenameDialog();

            fireEvent.click(screen.getByRole('button', { name: 'ACTION_CANCEL' }));

            await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
            expect(mockedRename).not.toHaveBeenCalled();
        });
    });

    describe('refresh intervals', () => {
        const refetchIntervalOf = (
            queryClient: ReturnType<typeof renderWithVacuumProviders>['queryClient'],
            queryKey: readonly string[]
        ) => {
            const observer = queryClient.getQueryCache().find({ queryKey })?.observers[0];
            return {
                interval: observer?.options.refetchInterval,
                isInBackground: observer?.options.refetchIntervalInBackground,
            };
        };

        it('refreshes the map every 30 s and the path every 5 s only while cleaning, never in background', async () => {
            const { queryClient } = await renderLoadedBlock('cleaning');
            await waitFor(() =>
                expect(refetchIntervalOf(queryClient, vacuumQueryKeys.map).interval).toBe(30000)
            );

            expect(refetchIntervalOf(queryClient, vacuumQueryKeys.rooms).interval).toBe(30000);
            expect(refetchIntervalOf(queryClient, vacuumQueryKeys.path).interval).toBe(5000);
            expect(refetchIntervalOf(queryClient, vacuumQueryKeys.map).isInBackground).toBe(false);
            expect(refetchIntervalOf(queryClient, vacuumQueryKeys.path).isInBackground).toBe(false);
        });

        it('does not poll the map nor the path when the robot is not cleaning', async () => {
            const { queryClient } = await renderLoadedBlock('idle');
            await waitFor(() => expect(mockedStatus).toHaveBeenCalled());

            expect(refetchIntervalOf(queryClient, vacuumQueryKeys.map).interval).toBe(false);
            expect(refetchIntervalOf(queryClient, vacuumQueryKeys.path).interval).toBe(false);
        });
    });

    it('draws the path and the robot position that the path endpoint reports', async () => {
        const context = createFakeCanvasContext();
        (HTMLCanvasElement.prototype.getContext as jest.Mock).mockImplementation(() => context);
        mockedPath.mockResolvedValue({
            path_id: 5,
            points: [
                [0, 0],
                [25, 25],
            ],
            position: { x: 25, y: 25 },
        });

        await renderLoadedBlock();

        await waitFor(() => expect(context.lineWidth).toBe(2));
        expect(context.arc).toHaveBeenCalledTimes(2);
    });
});
