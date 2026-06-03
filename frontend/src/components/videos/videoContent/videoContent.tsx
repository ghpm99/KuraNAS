import { VideoContentProvider } from '@/features/videos/providers/videoContentProvider';
import VideoContentScreen from './components/VideoContentScreen';

export default function VideoContent() {
    return (
        <VideoContentProvider>
            <VideoContentScreen />
        </VideoContentProvider>
    );
}
