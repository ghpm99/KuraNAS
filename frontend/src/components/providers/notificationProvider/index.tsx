import {
    getNotifications,
    getUnreadCount,
    markAllNotificationsAsRead,
    markNotificationAsRead,
} from '@/service/notifications';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useMemo } from 'react';
import { NotificationContextProvider } from './notificationContext';

export const NotificationProvider = ({ children }: { children: React.ReactNode }) => {
    const queryClient = useQueryClient();

    const unreadQuery = useQuery({
        queryKey: ['notifications-unread-count'],
        queryFn: getUnreadCount,
        refetchInterval: 30000,
    });

    const listQuery = useQuery({
        queryKey: ['notifications'],
        queryFn: () => getNotifications({ page: 1, pageSize: 5 }),
    });

    const markOneMutation = useMutation({
        mutationFn: markNotificationAsRead,
        onSuccess: async () => {
            await Promise.all([
                queryClient.invalidateQueries({ queryKey: ['notifications-unread-count'] }),
                queryClient.invalidateQueries({ queryKey: ['notifications'] }),
            ]);
        },
    });

    const markAllMutation = useMutation({
        mutationFn: markAllNotificationsAsRead,
        onSuccess: async () => {
            await Promise.all([
                queryClient.invalidateQueries({ queryKey: ['notifications-unread-count'] }),
                queryClient.invalidateQueries({ queryKey: ['notifications'] }),
            ]);
        },
    });

    const value = useMemo(
        () => ({
            unreadCount: unreadQuery.data?.unread_count ?? 0,
            notifications: listQuery.data?.items ?? [],
            isLoading: unreadQuery.isLoading || listQuery.isLoading,
            markAsRead: async (id: number) => {
                await markOneMutation.mutateAsync(id);
            },
            markAllAsRead: async () => {
                await markAllMutation.mutateAsync();
            },
            refetch: async () => {
                await Promise.all([unreadQuery.refetch(), listQuery.refetch()]);
            },
        }),
        [
            unreadQuery.data,
            unreadQuery.isLoading,
            listQuery.data,
            listQuery.isLoading,
            markOneMutation,
            markAllMutation,
            unreadQuery,
            listQuery,
        ]
    );

    return <NotificationContextProvider value={value}>{children}</NotificationContextProvider>;
};

export default NotificationProvider;
