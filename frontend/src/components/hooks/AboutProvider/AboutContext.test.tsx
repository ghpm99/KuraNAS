import { renderHook } from '@testing-library/react';
import React from 'react';
import { AboutContextProvider, useAbout } from './AboutContext';

describe('hooks/AboutProvider/AboutContext', () => {
	it('throws when used outside provider', () => {
		expect(() => renderHook(() => useAbout())).toThrow('useAbout must be used within an AboutProvider');
	});

	it('returns context when provider is present', () => {
		const value = {
			version: '1.0.0',
			commit_hash: 'abc123',
			platform: 'linux',
			path: '/tmp',
			lang: 'pt-BR',
			enable_workers: true,
			uptime: '1m 0s',
			statup_time: '2025-01-01T00:00:00Z',
			gin_mode: 'release',
			gin_version: '1.0',
			go_version: '1.21',
			node_version: '20',
		};
		const wrapper = ({ children }: { children: React.ReactNode }) => (
			<AboutContextProvider value={value}>{children}</AboutContextProvider>
		);

		const { result } = renderHook(() => useAbout(), { wrapper });
		expect(result.current.version).toBe('1.0.0');
		expect(result.current.enable_workers).toBe(true);
	});
});
