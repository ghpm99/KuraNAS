import { fireEvent, screen, waitFor } from '@testing-library/react';
import { apiBase } from '@/service';
import { buildVacuumMapPayload } from './testSupport/buildVacuumMapPayload';
import { createFakeCanvasContext } from './testSupport/createFakeCanvasContext';
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
    tapOnRoomOne,
    tapOnRoomTwo,
} from './testSupport/vacuumMapTaps';
import VacuumMapBlock from './VacuumMapBlock';

jest.mock('@/service', () => ({
    apiBase: {
        get: jest.fn(),
        put: jest.fn(),
        post: jest.fn(),
    },
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; put: jest.Mock; post: jest.Mock };

const responsesByUrl: Record<string, unknown> = {
    '/vacuum/status': { state: 'idle' },
    '/vacuum/cloud-settings': configuredCloudSettings,
    '/vacuum/map': buildVacuumMapPayload(),
    '/vacuum/rooms': mapRooms,
    '/vacuum/path': emptyPath,
};

const canvas = () => screen.getByRole('img', { name: 'VACUUM_MAP_CANVAS_LABEL' });

const renderMapBlock = async () => {
    renderWithVacuumProviders(<VacuumMapBlock />);
    await screen.findByRole('img', { name: 'VACUUM_MAP_CANVAS_LABEL' });
};

describe('VacuumMapBlock (seam)', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedApi.get.mockImplementation((url: string) =>
            Promise.resolve({ data: responsesByUrl[url] })
        );
        mockedApi.put.mockResolvedValue({ data: { id: 1, name: 'Sala', area: 3.8 } });
        mockedApi.post.mockResolvedValue({ data: undefined });
        mockCanvasGetContext(createFakeCanvasContext());
        mockElementBounds(frameWidthPixels, frameHeightPixels);
    });

    afterEach(() => jest.restoreAllMocks());

    it('asks for the map, the rooms and the path once when opened', async () => {
        await renderMapBlock();

        expect(mockedApi.get).toHaveBeenCalledWith('/vacuum/cloud-settings');
        expect(mockedApi.get).toHaveBeenCalledWith('/vacuum/map');
        expect(mockedApi.get).toHaveBeenCalledWith('/vacuum/rooms');
        expect(mockedApi.get).toHaveBeenCalledWith('/vacuum/path');
    });

    it('clicking Clean selected issues POST /vacuum/room-cleaning with the map id and the room ids', async () => {
        await renderMapBlock();
        fireEvent.click(canvas(), tapOnRoomOne);
        fireEvent.click(canvas(), tapOnRoomTwo);

        fireEvent.click(screen.getByRole('button', { name: 'VACUUM_MAP_CLEAN_SELECTED' }));

        await waitFor(() => expect(mockedApi.post).toHaveBeenCalledTimes(1));
        expect(mockedApi.post).toHaveBeenCalledWith('/vacuum/room-cleaning', {
            map_id: 77,
            room_ids: [1, 2],
        });
    });

    it('clicking Clean here issues POST /vacuum/spot-cleaning with the map id and x, y in millimeters', async () => {
        await renderMapBlock();
        fireEvent.click(screen.getByRole('button', { name: 'VACUUM_MAP_MODE_CLEAN_SPOT' }));
        fireEvent.click(canvas(), tapOnRoomTwo);

        fireEvent.click(screen.getByRole('button', { name: 'VACUUM_MAP_CLEAN_SPOT' }));

        await waitFor(() => expect(mockedApi.post).toHaveBeenCalledTimes(1));
        expect(mockedApi.post).toHaveBeenCalledWith('/vacuum/spot-cleaning', {
            map_id: 77,
            x: 25,
            y: 25,
        });
    });

    it('saving a room name issues PUT /vacuum/rooms/:id with the map id and the name', async () => {
        await renderMapBlock();
        fireEvent.click(
            screen.getAllByRole('button', { name: 'VACUUM_ROOM_RENAME_LABEL' })[0] as HTMLElement
        );
        fireEvent.change(await screen.findByLabelText('VACUUM_ROOM_RENAME_FIELD'), {
            target: { value: 'Sala' },
        });

        fireEvent.click(screen.getByRole('button', { name: 'VACUUM_ROOM_RENAME_SAVE' }));

        await waitFor(() => expect(mockedApi.put).toHaveBeenCalledTimes(1));
        expect(mockedApi.put).toHaveBeenCalledWith('/vacuum/rooms/1', { map_id: 77, name: 'Sala' });
    });
});
