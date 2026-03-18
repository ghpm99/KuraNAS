import { fireEvent, render, screen } from '@testing-library/react';
import AboutLayout from './about/aboutLayout';
import ActivityDiaryLayout from './activityDiary/activityDiaryLayout';
import FilesLayout from './files/filesLayout';
import ImagesLayout from './images/imagesLayout';
import VideoLayout from './videos/videoLayout';
import MusicLayout from './music/musicLayout';
import MusicSidebar from './music/MusicSidebar';
import NavItem from './layout/Sidebar/components/navItem';
import Button from './ui/Button/Button';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

jest.mock('react-router-dom', () => ({
    Link: ({ to, children, ...props }: any) => (
        <a href={to} {...props}>
            {children}
        </a>
    ),
    Outlet: () => <div>OutletMock</div>,
    useLocation: () => ({ pathname: '/music/artists' }),
}));

describe('layout wrappers and export indexes', () => {
    it('layout wrappers return JSX containing their children', () => {
        const child = <span>child</span>;
        const layouts = [AboutLayout, ActivityDiaryLayout, FilesLayout, ImagesLayout, VideoLayout, MusicLayout];

        for (const Layout of layouts) {
            const result = (Layout as any)({ children: child });
            expect(result).toBeTruthy();
            expect(result.props).toBeDefined();
        }
    });

    it('renders music sidebar links', () => {
        render(<MusicSidebar />);
        expect(screen.getByRole('link', { name: /MUSIC_ARTISTS/i })).toHaveAttribute(
            'href',
            '/music/artists'
        );
        expect(screen.getByRole('link', { name: /MUSIC_PLAYLISTS/i })).toHaveAttribute(
            'href',
            '/music/playlists'
        );
    });

    it('renders nav item link and generic button', () => {
        render(
            <NavItem href="/images" icon={<span>icon</span>}>
                Images
            </NavItem>
        );
        expect(screen.getByRole('link', { name: /Images/ })).toBeInTheDocument();

        const onClick = jest.fn();
        render(<Button onClick={onClick}>Run</Button>);
        fireEvent.click(screen.getByRole('button', { name: 'Run' }));
        expect(onClick).toHaveBeenCalled();
    });

});
