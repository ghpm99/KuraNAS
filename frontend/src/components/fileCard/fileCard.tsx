import './fileCard.css';
const FileCard = ({
	title,
	metadata,
	thumbnail,
	onClick,
	starred,
	onClickStar,
}: {
	title: string;
	metadata: string;
	thumbnail: string;
	onClick: () => void;
	starred?: boolean;
	onClickStar?: () => void;
}) => {
	return (
		<div className='file-card'>
			<div className='file-thumbnail'>
				<img
					loading='lazy'
					src={thumbnail || '/placeholder.svg'}
					width={652}
					height={489}
					className='thumbnail-image'
					onClick={onClick}
				/>
				<span className='star-icon' onClick={onClickStar}>
					{starred ? '★' : '☆'}
				</span>
			</div>
			<div className='file-info'>
				<h3 className='file-title'>{title}</h3>
				<p className='file-metadata'>{metadata}</p>
			</div>
		</div>
	);
};

export default FileCard;
