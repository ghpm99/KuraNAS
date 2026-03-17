import { createContext, useContext } from 'react';
import type { Notification } from '@/types/notification';

export type NotificationContextType = {
    unreadCount: number;
    notifications: Notification[];
    isLoading: boolean;
    markAsRead: (id: number) => Promise<void>;
    markAllAsRead: () => Promise<void>;
    refetch: () => Promise<void>;
};

const NotificationContext = createContext<NotificationContextType | undefined>(undefined);

export const NotificationContextProvider = NotificationContext.Provider;

export const useNotifications = () => {
    const context = useContext(NotificationContext);
    if (!context) {
        throw new Error('useNotifications must be used within a NotificationProvider');
    }
    return context;
};
