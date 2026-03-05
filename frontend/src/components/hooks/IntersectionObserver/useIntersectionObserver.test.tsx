import { act, renderHook } from '@testing-library/react';
import { useIntersectionObserver } from './useIntersectionObserver';

type Entry = { isIntersecting: boolean };
type Instance = {
	emit: (value: boolean) => void;
	emitEntries: (entries: Entry[]) => void;
	observe: jest.Mock;
	disconnect: jest.Mock;
	options: IntersectionObserverInit;
};

const instances: Instance[] = [];

class IntersectionObserverMock {
	private readonly callback: (entries: Entry[]) => void;
	public readonly observe = jest.fn();
	public readonly disconnect = jest.fn();
	public readonly options: IntersectionObserverInit;

	constructor(callback: (entries: Entry[]) => void, options: IntersectionObserverInit) {
		this.callback = callback;
		this.options = options;
		instances.push({
			emit: (value: boolean) => this.callback([{ isIntersecting: value }]),
			emitEntries: (entries: Entry[]) => this.callback(entries),
			observe: this.observe,
			disconnect: this.disconnect,
			options,
		});
	}
}

describe('useIntersectionObserver', () => {
	beforeEach(() => {
		instances.length = 0;
		(global as any).IntersectionObserver = IntersectionObserverMock;
	});

	it('observes node and reacts to intersection', () => {
		const onIntersect = jest.fn();
		const { result } = renderHook(() =>
			useIntersectionObserver<HTMLDivElement>({
				threshold: 0.5,
				rootMargin: '16px',
				onIntersect,
			}),
		);

		const node = document.createElement('div');

		let cleanup: undefined | (() => void);
		act(() => {
			cleanup = result.current.ref(node) as (() => void) | undefined;
		});

		expect(instances).toHaveLength(1);
		expect(instances[0]!.options).toEqual({ threshold: 0.5, rootMargin: '16px' });
		expect(instances[0]!.observe).toHaveBeenCalledWith(node);
		expect(result.current.isIntersecting).toBe(false);

		act(() => {
			instances[0]!.emit(true);
		});

		expect(result.current.isIntersecting).toBe(true);
		expect(onIntersect).toHaveBeenCalledTimes(1);

		act(() => {
			instances[0]!.emitEntries([]);
		});

		expect(result.current.isIntersecting).toBe(false);

		act(() => {
			result.current.ref(null);
		});

		expect(instances[0]!.disconnect).toHaveBeenCalledTimes(1);

		act(() => {
			cleanup?.();
		});
		expect(instances[0]!.disconnect).toHaveBeenCalledTimes(2);
	});

	it('does not create observer when disabled and disconnects existing observer', () => {
		const { result, rerender } = renderHook(
			({ enabled }) => useIntersectionObserver<HTMLDivElement>({ enabled }),
			{ initialProps: { enabled: true } },
		);
		const node = document.createElement('div');

		act(() => {
			result.current.ref(node);
		});
		expect(instances).toHaveLength(1);

		rerender({ enabled: false });

		act(() => {
			result.current.ref(node);
		});

		expect(instances[0]!.disconnect).toHaveBeenCalledTimes(1);
		expect(instances).toHaveLength(1);
	});

	it('does not call onIntersect when callback is not provided', () => {
		const { result } = renderHook(() => useIntersectionObserver<HTMLDivElement>());
		const node = document.createElement('div');

		act(() => {
			result.current.ref(node);
		});

		act(() => {
			instances[0]!.emit(true);
		});

		expect(result.current.isIntersecting).toBe(true);
	});
});
