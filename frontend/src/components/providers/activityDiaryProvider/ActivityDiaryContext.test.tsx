import { renderHook } from '@testing-library/react';
import React from 'react';
import { ActivityDiaryContextProvider, useActivityDiary } from './ActivityDiaryContext';

describe('providers/activityDiaryProvider/ActivityDiaryContext', () => {
	it('throws when used outside provider', () => {
		expect(() => renderHook(() => useActivityDiary())).toThrow(
			'useActivityDiary must be used within a ActivityDiaryProvider',
		);
	});

	it('returns data from provider', () => {
		const value = {
			form: { name: '', description: '' },
			handleSubmit: jest.fn(),
			handleNameChange: jest.fn(),
			handleDescriptionChange: jest.fn(),
			loading: true,
			data: null,
			getCurrentDuration: jest.fn(() => 0),
			currentTime: new Date(),
			copyActivity: jest.fn(),
		};
		const wrapper = ({ children }: { children: React.ReactNode }) => (
			<ActivityDiaryContextProvider value={value as any}>{children}</ActivityDiaryContextProvider>
		);

		const { result } = renderHook(() => useActivityDiary(), { wrapper });
		expect(result.current.loading).toBe(true);
	});
});
