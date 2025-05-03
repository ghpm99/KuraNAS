const FileCard = ({
	title,
	metadata,
	thumbnail,
	onClick,
}: {
	title: string;
	metadata: string;
	thumbnail: string;
	onClick: () => void;
}) => {
	return (
		<div className='file-card' onClick={onClick}>
			<div className='file-thumbnail'>
				<img
					loading='lazy'
					src={thumbnail || '/placeholder.svg'}
					width={652}
					height={489}
					className='thumbnail-image'
				/>
			</div>
			<div className='file-info'>
				<h3 className='file-title'>{title}</h3>
				<p className='file-metadata'>{metadata}</p>
			</div>
		</div>
	);
};

export default FileCard;
