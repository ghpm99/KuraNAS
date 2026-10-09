import { AboutProvider } from '../providers/aboutProvider';

const AboutLayout = ({ children }: { children: React.ReactNode }) => {
    return <AboutProvider>{children}</AboutProvider>;
};

export default AboutLayout;
