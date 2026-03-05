import { ListItemButton, ListItemIcon, ListItemText } from '@mui/material';
import { File, Folder } from 'lucide-react';
import React from 'react';

const FolderItem = ({
	children,
	label,
	type,
	onClick,
	expanded,
	selected,
}: {
	children: React.ReactNode;
	label: string;
	type: number;
	onClick: () => void;
	expanded: boolean;
	selected: boolean;
}) => {
	const formatLabel = () => {
		if (label?.length <= 12 || type === 1) return label;
		const parts = label.split('.');
		const ext = parts[parts.length - 1] ?? '';
		const maxName = 12 - ext.length - 3;
		return `${label.substring(0, maxName)}...${ext}`;
	};

	return (
		<>
			<ListItemButton selected={selected} onClick={onClick} sx={{ borderRadius: 1, py: 0.5 }}>
				<ListItemIcon sx={{ minWidth: 28 }}>
					{type === 1 ? <Folder size={16} /> : <File size={16} />}
				</ListItemIcon>
				<ListItemText primary={formatLabel()} primaryTypographyProps={{ variant: 'body2', noWrap: true }} />
			</ListItemButton>
			{children && expanded && <>{children}</>}
		</>
	);
};

export default FolderItem;
