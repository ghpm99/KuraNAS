import { renderHook } from '@testing-library/react';
import { createElement } from 'react';
import { useNotifications, NotificationContextProvider, type NotificationContextType } from './notificationContext';

describe('useNotifications', () => {
	it('throws when used outside a provider', () => {
		jest.spyOn(console, 'error').mockImplementation(() => {});

		expect(() => {
			renderHook(() => useNotifications());
		}).toThrow('useNotifications must be used within a NotificationProvider');

		(console.error as jest.Mock).mockRestore();
	});

	it('returns context value when inside provider', () => {
		const value: NotificationContextType = {
			unreadCount: 3,
			notifications: [],
			isLoading: false,
			markAsRead: jest.fn(),
			markAllAsRead: jest.fn(),
			refetch: jest.fn(),
		};
		const wrapper = ({ children }: { children: React.ReactNode }) =>
			createElement(NotificationContextProvider, { value }, children);

		const { result } = renderHook(() => useNotifications(), { wrapper });
		expect(result.current.unreadCount).toBe(3);
	});
});
