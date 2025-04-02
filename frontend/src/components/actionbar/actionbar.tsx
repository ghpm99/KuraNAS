import { Plus } from 'lucide-react'

export const ActionBar = () => {
    return ( <div className='action-bar'>
        <button className='button primary-button'>
            <Plus className='icon' />
            Create
        </button>
        <button className='button secondary-button'>
            <svg className='icon' viewBox='0 0 24 24' fill='none' stroke='currentColor'>
                <path
                    d='M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M7 10l5 5 5-5M12 15V3'
                    strokeWidth='2'
                    strokeLinecap='round'
                    strokeLinejoin='round'
                />
            </svg>
            Upload
        </button>
        <button className='button secondary-button'>
            <svg className='icon' viewBox='0 0 24 24' fill='none' stroke='currentColor'>
                <path
                    d='M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z'
                    strokeWidth='2'
                    strokeLinecap='round'
                    strokeLinejoin='round'
                />
            </svg>
            Create folder
        </button>
        <button className='button secondary-button'>
            <svg className='icon' viewBox='0 0 24 24' fill='none' stroke='currentColor'>
                <path
                    d='M12 18.5a6.5 6.5 0 100-13 6.5 6.5 0 000 13zM12 14a2 2 0 100-4 2 2 0 000 4z'
                    strokeWidth='2'
                    strokeLinecap='round'
                    strokeLinejoin='round'
                />
            </svg>
            Record
        </button>
    </div>)
}

export default ActionBar