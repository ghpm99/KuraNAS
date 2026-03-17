import type { Pagination } from '@/types/pagination';
import type { Notification, UnreadCount } from '@/types/notification';
import { apiBase } from './index';

interface ListNotificationsParams {
    page?: number;
    pageSize?: number;
    type?: string;
    is_read?: boolean;
}

export const getNotifications = async ({
    page = 1,
    pageSize = 20,
    type: notifType,
    is_read,
}: ListNotificationsParams = {}) => {
    const response = await apiBase.get<Pagination<Notification>>('/notifications', {
        params: {
            page,
            page_size: pageSize,
            type: notifType,
            is_read,
        },
    });
    return response.data;
};

export const getUnreadCount = async () => {
    const response = await apiBase.get<UnreadCount>('/notifications/unread-count');
    return response.data;
};

export const markNotificationAsRead = async (id: number) => {
    await apiBase.put(`/notifications/${id}/read`);
};

export const markAllNotificationsAsRead = async () => {
    await apiBase.put('/notifications/read-all');
};
