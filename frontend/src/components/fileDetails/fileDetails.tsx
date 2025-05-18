import { FileType, formatDate, formatSize, getFileTypeInfo } from '@/utils';
import useFile from '../hooks/fileProvider/fileContext';
import './fileDetails.css';
const FileDetails = () => {
	const { selectedItem } = useFile();

	if (!selectedItem || selectedItem.type === FileType.Directory) return <></>;
	const fileType = getFileTypeInfo(selectedItem.format);
	return (
		<div className='file-details'>
			<div className='details-header'>
				<h2 className='details-title'>File Details</h2>
				<p className='details-subtitle'>Information about this file</p>
			</div>

			<div className='details-section'>
				<h3 className='section-title'>Properties</h3>
				<div className='detail-item'>
					<span className='detail-label'>Type</span>
					<span className='detail-value'>{fileType.description}</span>
				</div>
				<div className='detail-item'>
					<span className='detail-label'>Size</span>
					<span className='detail-value'>
						{formatSize(selectedItem.size)}({selectedItem.size} B)
					</span>
				</div>
				<div className='detail-item'>
					<span className='detail-label'>Created</span>
					<span className='detail-value'>{formatDate(selectedItem.created_at)}</span>
				</div>
				<div className='detail-item'>
					<span className='detail-label'>Modified</span>
					<span className='detail-value'>{formatDate(selectedItem.updated_at)}</span>
				</div>
				<div className='detail-item'>
					<span className='detail-label'>Path</span>
					<span className='detail-value'>{selectedItem.path}</span>
				</div>
			</div>

			<div className='details-section'>
				<h3 className='section-title'>Tags</h3>
				<div className='tag-list'>
					{/* {fileData.tags.map((tag, index) => (
        <span key={index} className="tag">
          {tag}
        </span>
      ))} */}
				</div>
			</div>

			<div className='details-section'>
				<h3 className='section-title'>Recent Activity</h3>
				<ul className='activity-list'>
					{/* {fileData.activities.map((activity, index) => (
        <li key={index} className="activity-item">
          <div className="activity-avatar">
            <Image src={activity.avatar || "/placeholder.svg"} alt={activity.user} width={32} height={32} />
          </div>
          <div className="activity-content">
            <p className="activity-user">{activity.user}</p>
            <p className="activity-action">{activity.action}</p>
            <p className="activity-time">{activity.time}</p>
          </div>
        </li>
      ))} */}
				</ul>
			</div>
		</div>
	);
};

export default FileDetails;
