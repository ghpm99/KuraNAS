export type FileLocation = {
    file_id: number;
    tier: 'hot' | 'cold';
    logical_path: string;
    disk_path: string;
    logical_disk_path: string;
    root_label: string;
    exists_on_disk: boolean;
};
