import { act, renderHook } from '@testing-library/react';
import { UIProvider } from './index';
import { useUI } from './uiContext';

describe('providers/uiProvider/uiContext', () => {
	it('throws when used outside provider', () => {
		expect(() => renderHook(() => useUI())).toThrow('useUI must be used within a UIProvider');
	});

	it('provides state and updater through UIProvider', () => {
		const wrapper = ({ children }: { children: React.ReactNode }) => <UIProvider>{children}</UIProvider>;
		const { result } = renderHook(() => useUI(), { wrapper });

		expect(result.current.activePage).toBe('unknown');
		act(() => {
			result.current.setActivePage('music');
		});
		expect(result.current.activePage).toBe('music');
	});
});
