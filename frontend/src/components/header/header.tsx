import { Bell, Grid, Search } from 'lucide-react'


const Header = () => {
    return (
        <header className='header'>
					<div className='search-container'>
						<Search className='search-icon' />
						<input type='search' placeholder='Search files...' className='search-input' />
					</div>
					<div className='header-actions'>
						<button className='icon-button'>
							<Grid className='icon' />
						</button>
						<button className='icon-button'>
							<Bell className='icon' />
						</button>
						<div className='avatar'>
							<img src='/placeholder.svg' alt='Avatar' width={32} height={32} />
						</div>
					</div>
				</header>
    )
}

export default Header