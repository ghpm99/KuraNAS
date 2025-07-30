import sys
import json
from mutagen import File

# Mapeia nomes técnicos (ID3v2) para chaves amigáveis
ID3_TAG_MAP = {
    "TIT2": "title",
    "TPE1": "artist",
    "TALB": "album",
    "TPE2": "album_artist",
    "TRCK": "track_number",
    "TCON": "genre",
    "TCOM": "composer",
    "TYER": "year",
    "TDRC": "recording_date",
    "TENC": "encoder",
    "TPUB": "publisher",
    "TDOR": "original_release_date",
    "TOPE": "original_artist",
    "TEXT": "lyricist",
    "USLT": "lyrics",
}

# Todas as chaves que sempre estarão presentes no JSON final
OUTPUT_KEYS = {
    # Técnicas
    "mime": "",
    "length": 0.0,
    "bitrate": 0,
    "sample_rate": 0,
    "channels": 0,
    "bitrate_mode": 0,
    "encoder_info": "",
    "bit_depth": 0,
    # Tags amigáveis
    "title": "",
    "artist": "",
    "album": "",
    "album_artist": "",
    "track_number": "",
    "genre": "",
    "composer": "",
    "year": "",
    "recording_date": "",
    "encoder": "",
    "publisher": "",
    "original_release_date": "",
    "original_artist": "",
    "lyricist": "",
    "lyrics": "",
}


def serialize_value(value):
    if isinstance(value, (str, int, float)):
        return value
    elif isinstance(value, list):
        return value[0] if len(value) > 0 and isinstance(value[0], (str, int, float)) else ""
    elif hasattr(value, "text"):
        return value.text[0] if isinstance(value.text, list) else value.text
    return ""


def extract_metadata(path):
    output = OUTPUT_KEYS.copy()

    try:
        audio = File(path, easy=False)
        if audio is None:
            return output

        # MIME
        if hasattr(audio, "mime") and audio.mime:
            output["mime"] = audio.mime[0]

        # Info técnica
        info = getattr(audio, "info", None)
        if info:
            output["length"] = getattr(info, "length", 0.0)
            output["bitrate"] = getattr(info, "bitrate", 0)
            output["sample_rate"] = getattr(info, "sample_rate", 0)
            output["channels"] = getattr(info, "channels", 0)
            output["bitrate_mode"] = getattr(info, "bitrate_mode", 0)
            output["encoder_info"] = getattr(info, "encoder_info", "")
            output["bit_depth"] = getattr(info, "bits_per_sample", 0)

        # Tags
        if audio.tags:
            for key, value in audio.tags.items():
                friendly = ID3_TAG_MAP.get(key, None)
                if friendly and friendly in output:
                    output[friendly] = serialize_value(value)

    except Exception:
        pass

    return output


if __name__ == "__main__":
    try:
        path = sys.argv[1]
        metadata = extract_metadata(path)
        print(json.dumps(metadata, ensure_ascii=False, indent=2))
    except Exception:
        print(json.dumps(OUTPUT_KEYS, ensure_ascii=False, indent=2))
