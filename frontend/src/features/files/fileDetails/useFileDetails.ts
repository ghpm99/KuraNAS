import { useContext } from 'react';
import { FileDetailsContext } from './fileDetailsContext';

export const useFileDetails = () => useContext(FileDetailsContext);

export default useFileDetails;
