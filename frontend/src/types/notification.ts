export type NotificationType = 'info' | 'success' | 'warning' | 'error' | 'system';

export interface Notification {
    id: number;
    type: NotificationType;
    title: string;
    message: string;
    metadata?: Record<string, unknown>;
    is_read: boolean;
    created_at: string;
    group_key?: string;
    group_count: number;
    is_grouped: boolean;
}

export interface UnreadCount {
    unread_count: number;
}
