import React from 'react';
import './folderItem.css';

const FolderSvg = () => {
	return (
		<svg className='icon' fill='none' stroke='currentColor' viewBox='0 0 24 24'>
			<path
				strokeLinecap='round'
				strokeLinejoin='round'
				strokeWidth={2}
				d='M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z'
			/>
		</svg>
	);
};

const FilerSvg = () => {
	return (
		<svg className='icon' fill='none' stroke='currentColor' viewBox='0 0 24 24'>
			<path
				strokeLinecap='round'
				strokeLinejoin='round'
				strokeWidth={2}
				d='M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 00-2-2V5a2 2 0 002-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 002 2z'
			/>
		</svg>
	);
};
const FolderItem = ({
	children,
	label,
	type,
	onClick,
}: {
	children: React.ReactNode;
	label: string;
	type: number;
	onClick: () => void;
}) => {
	return (
		<>
			<a className='folder-item' onClick={onClick}>
				{type === 1 ? <FolderSvg /> : <FilerSvg />}
				<span>{label}</span>
			</a>
			{children && <div className='folder-children'>{children}</div>}
		</>
	);
};

export default FolderItem;
