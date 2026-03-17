import { Box, Button, Tab, Tabs, Typography } from '@mui/material';
import { useInfiniteQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { getNotifications, markAllNotificationsAsRead, markNotificationAsRead } from '@/service/notifications';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { NotificationType } from '@/types/notification';
import NotificationItem from './NotificationItem';

type FilterTab = 'all' | 'unread' | NotificationType;

const PAGE_SIZE = 20;

export default function NotificationsScreen() {
    const { t } = useI18n();
    const queryClient = useQueryClient();
    const [activeTab, setActiveTab] = useState<FilterTab>('all');

    const getQueryParams = () => {
        const params: { pageSize: number; type?: string; is_read?: boolean } = { pageSize: PAGE_SIZE };
        if (activeTab === 'unread') {
            params.is_read = false;
        } else if (activeTab !== 'all') {
            params.type = activeTab;
        }
        return params;
    };

    const { data, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
        queryKey: ['notifications-page', activeTab],
        queryFn: async ({ pageParam = 1 }) => {
            const params = getQueryParams();
            return getNotifications({ page: pageParam, ...params });
        },
        initialPageParam: 1,
        getNextPageParam: (lastPage) =>
            lastPage.pagination.has_next ? lastPage.pagination.page + 1 : undefined,
    });

    const markOneMutation = useMutation({
        mutationFn: markNotificationAsRead,
        onSuccess: async () => {
            await Promise.all([
                queryClient.invalidateQueries({ queryKey: ['notifications-page'] }),
                queryClient.invalidateQueries({ queryKey: ['notifications-unread-count'] }),
                queryClient.invalidateQueries({ queryKey: ['notifications'] }),
            ]);
        },
    });

    const markAllMutation = useMutation({
        mutationFn: markAllNotificationsAsRead,
        onSuccess: async () => {
            await Promise.all([
                queryClient.invalidateQueries({ queryKey: ['notifications-page'] }),
                queryClient.invalidateQueries({ queryKey: ['notifications-unread-count'] }),
                queryClient.invalidateQueries({ queryKey: ['notifications'] }),
            ]);
        },
    });

    const allNotifications = data?.pages.flatMap((page) => page.items) ?? [];

    const handleItemClick = async (id: number, isRead: boolean) => {
        if (!isRead) {
            await markOneMutation.mutateAsync(id);
        }
    };

    return (
        <Box sx={{ p: { xs: 1.5, sm: 3 }, maxWidth: 720, mx: 'auto' }}>
            <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 2, flexWrap: 'wrap', gap: 1 }}>
                <Typography variant="h6" sx={{ fontWeight: 600 }}>
                    {t('NOTIFICATIONS')}
                </Typography>
                <Button
                    size="small"
                    onClick={() => markAllMutation.mutate()}
                    disabled={markAllMutation.isPending}
                    sx={{ textTransform: 'none' }}
                >
                    {t('MARK_ALL_AS_READ')}
                </Button>
            </Box>

            <Tabs
                value={activeTab}
                onChange={(_, v) => setActiveTab(v)}
                variant="scrollable"
                scrollButtons="auto"
                sx={{ mb: 2, minHeight: 36, '& .MuiTab-root': { minHeight: 36, py: 0.5 } }}
            >
                <Tab label={t('ALL')} value="all" />
                <Tab label={t('UNREAD')} value="unread" />
                <Tab label="Info" value="info" />
                <Tab label="Success" value="success" />
                <Tab label="Warning" value="warning" />
                <Tab label="Error" value="error" />
            </Tabs>

            {isLoading ? (
                <Box sx={{ p: 4, textAlign: 'center' }}>
                    <Typography variant="body2" color="text.secondary">
                        {t('LOADING')}
                    </Typography>
                </Box>
            ) : allNotifications.length === 0 ? (
                <Box sx={{ p: 4, textAlign: 'center' }}>
                    <Typography variant="body2" color="text.secondary">
                        {t('NO_NOTIFICATIONS')}
                    </Typography>
                </Box>
            ) : (
                <>
                    <Box sx={{ display: 'flex', flexDirection: 'column', gap: 0.5 }}>
                        {allNotifications.map((notification) => (
                            <NotificationItem
                                key={notification.id}
                                notification={notification}
                                onClick={() => handleItemClick(notification.id, notification.is_read)}
                            />
                        ))}
                    </Box>
                    {hasNextPage && (
                        <Box sx={{ textAlign: 'center', mt: 2 }}>
                            <Button
                                onClick={() => fetchNextPage()}
                                disabled={isFetchingNextPage}
                                sx={{ textTransform: 'none' }}
                            >
                                {isFetchingNextPage ? t('LOADING') : t('LOAD_MORE')}
                            </Button>
                        </Box>
                    )}
                </>
            )}
        </Box>
    );
}
