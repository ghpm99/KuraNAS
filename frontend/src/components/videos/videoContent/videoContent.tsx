import { VideoContentProvider } from '@/components/providers/videoContentProvider';
import VideoContentScreen from './components/VideoContentScreen';

export default function VideoContent() {
    return (
        <VideoContentProvider>
            <VideoContentScreen />
        </VideoContentProvider>
    );
}
