import sys
from mutagen import File


def extract_audio_metadata(audio_path):
    try:
        audio = File(audio_path, easy=True)
        if audio is None:
            print(f"Arquivo não suportado ou corrompido: {audio_path}")
            return None
        metadata = {
            "mime": audio.mime[0] if hasattr(audio, "mime") and audio.mime else "",
            "info": dict(audio.info.__dict__) if hasattr(audio, "info") else {},
            "tags": dict(audio) if audio.tags else {},
        }
        print(metadata)
        return metadata
    except Exception as e:
        print(f"Erro ao extrair metadados de {audio_path}: {e}")
        return None


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Uso: python audio_metadados.py <caminho_do_audio>")
        sys.exit(1)
    audio_path = sys.argv[1]
    extract_audio_metadata(audio_path)
