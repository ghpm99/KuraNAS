const FileCard = ({ title, metadata, thumbnail }: { title: string; metadata: string; thumbnail: string }) => {
	return (
		<div className='file-card'>
			<div className='file-thumbnail'>
				<image href={thumbnail || '/placeholder.svg'} width={400} height={300} className='thumbnail-image' />
			</div>
			<div className='file-info'>
				<h3 className='file-title'>{title}</h3>
				<p className='file-metadata'>{metadata}</p>
			</div>
		</div>
	);
}

export default FileCard