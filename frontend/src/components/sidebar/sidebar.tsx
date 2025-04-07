import { LayoutGrid } from 'lucide-react'
import NavItem from './components/navitem'
import FolderTree from './components/foldertree'

const Sidebar = () => {
    return (
        <div className='sidebar'>
				<div className='sidebar-header'>
					<h1 className='app-title'>KuraNAS</h1>
				</div>
				<nav className='nav'>
					<NavItem href='#' icon={<LayoutGrid className='icon' />} active>
						Todos Arquivos
					</NavItem>
					<NavItem
						href='#'
						icon={
							<svg className='icon' viewBox='0 0 24 24' fill='none' stroke='currentColor'>
								<path
									d='M15 3v18M12 3h7a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-7m0-18H5a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h7m0-18v18'
									strokeWidth='2'
									strokeLinecap='round'
								/>
							</svg>
						}
					>
						Presentations
					</NavItem>
					<NavItem
						href='#'
						icon={
							<svg className='icon' viewBox='0 0 24 24' fill='none' stroke='currentColor'>
								<path
									d='M9 5H7a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V7a2 2 0 0 0-2-2h-2M9 5a2 2 0 0 1 2-2h2a2 2 0 0 1 2 2M9 5h6m-3 4v6m-3-3h6'
									strokeWidth='2'
									strokeLinecap='round'
									strokeLinejoin='round'
								/>
							</svg>
						}
					>
						Analytics
					</NavItem>
					<FolderTree />
				</nav>
			</div>
    )
}

export default Sidebar