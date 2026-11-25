export type FakeCanvasContext = ReturnType<typeof createFakeCanvasContext>;

export const createFakeCanvasContext = () => {
    const imageData = { data: new Uint8ClampedArray(0) };
    return {
        fillStyle: '' as string,
        strokeStyle: '' as string,
        lineWidth: 0,
        lineJoin: 'miter' as CanvasLineJoin,
        clearRect: jest.fn(),
        beginPath: jest.fn(),
        moveTo: jest.fn(),
        lineTo: jest.fn(),
        closePath: jest.fn(),
        arc: jest.fn(),
        fill: jest.fn(),
        stroke: jest.fn(),
        setLineDash: jest.fn(),
        createImageData: jest.fn((width: number, height: number) => {
            imageData.data = new Uint8ClampedArray(width * height * 4);
            return imageData;
        }),
        putImageData: jest.fn(),
        imageData,
    };
};
