import { createFakeCanvasContext, type FakeCanvasContext } from './createFakeCanvasContext';

export const mockCanvasGetContext = (
    context: FakeCanvasContext | null = createFakeCanvasContext()
): jest.SpyInstance =>
    jest
        .spyOn(HTMLCanvasElement.prototype, 'getContext')
        .mockImplementation(() => context as unknown as CanvasRenderingContext2D);
