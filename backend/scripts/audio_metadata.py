import sys
from mutagen import File


def extract_audio_metadata(audio_path):
    try:
        audio = File(audio_path, easy=True)
        if audio is None:
            return {
                "mime": "",
                "info": {},
                "tags": {},
            }
        metadata = {
            "mime": audio.mime[0] if hasattr(audio, "mime") and audio.mime else "",
            "info": dict(audio.info.__dict__) if hasattr(audio, "info") else {},
            "tags": dict(audio) if audio.tags else {},
        }
        return metadata
    except:
        return {
            "mime": "",
            "info": {},
            "tags": {},
        }


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Uso: python audio_metadata.py <caminho_do_audio>")
        sys.exit(1)
    audio_path = sys.argv[1]
    extract_audio_metadata(audio_path)
