
const FolderItem = ({ href, children }: { href: string; children: React.ReactNode }) => {
	return (
		<a href={href} className='folder-item'>
			<svg className='icon' fill='none' stroke='currentColor' viewBox='0 0 24 24'>
				<path
					strokeLinecap='round'
					strokeLinejoin='round'
					strokeWidth={2}
					d='M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z'
				/>
			</svg>
			<span>{children}</span>
		</a>
	);
}

export default FolderItem