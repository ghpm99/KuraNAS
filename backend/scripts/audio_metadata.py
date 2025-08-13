import os
import sys
import json
import traceback
from mutagen import File
from mutagen.id3 import ID3

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
    return str(value) if value is not None else ""


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
        id3_tags = ID3(path)
        if id3_tags is None:
            return output

        output["title"] = serialize_value(
            id3_tags.get("TIT2").text[0] if id3_tags.get("TIT2") and id3_tags.get("TIT2").text else ""
        )
        output["artist"] = serialize_value(
            id3_tags.get("TPE1").text[0] if id3_tags.get("TPE1") and id3_tags.get("TPE1").text else ""
        )
        output["album"] = serialize_value(
            id3_tags.get("TALB").text[0] if id3_tags.get("TALB") and id3_tags.get("TALB").text else ""
        )
        output["album_artist"] = serialize_value(
            id3_tags.get("TPE2").text[0] if id3_tags.get("TPE2") and id3_tags.get("TPE2").text else ""
        )
        output["track_number"] = serialize_value(
            id3_tags.get("TRCK").text[0] if id3_tags.get("TRCK") and id3_tags.get("TRCK").text else ""
        )
        output["genre"] = serialize_value(
            id3_tags.get("TCON").text[0] if id3_tags.get("TCON") and id3_tags.get("TCON").text else ""
        )
        output["composer"] = serialize_value(
            id3_tags.get("TCOM").text[0] if id3_tags.get("TCOM") and id3_tags.get("TCOM").text else ""
        )
        output["year"] = serialize_value(
            id3_tags.get("TYER").text[0] if id3_tags.get("TYER") and id3_tags.get("TYER").text else ""
        )
        output["recording_date"] = serialize_value(
            id3_tags.get("TDRC").text[0] if id3_tags.get("TDRC") and id3_tags.get("TDRC").text else ""
        )
        output["encoder"] = serialize_value(
            id3_tags.get("TENC").text[0] if id3_tags.get("TENC") and id3_tags.get("TENC").text else ""
        )
        output["publisher"] = serialize_value(
            id3_tags.get("TPUB").text[0] if id3_tags.get("TPUB") and id3_tags.get("TPUB").text else ""
        )
        output["original_release_date"] = serialize_value(
            id3_tags.get("TDOR").text[0] if id3_tags.get("TDOR") and id3_tags.get("TDOR").text else ""
        )
        output["original_artist"] = serialize_value(
            id3_tags.get("TOPE").text[0] if id3_tags.get("TOPE") and id3_tags.get("TOPE").text else ""
        )
        output["lyricist"] = serialize_value(
            id3_tags.get("TEXT").text[0] if id3_tags.get("TEXT") and id3_tags.get("TEXT").text else ""
        )

    except Exception:
        save_traceback(path)

    return output


def save_traceback(path):
    base_dir = os.path.dirname(os.path.abspath(__file__))
    logs_dir = os.path.join(base_dir, "logs")
    os.makedirs(logs_dir, exist_ok=True)
    log_path = os.path.join(logs_dir, "audio_metadata.log")
    with open(log_path, "a", encoding="utf-8") as f:
        f.write(f"Erro ao processar audio {path}:\n")
        f.write(traceback.format_exc())
        f.write("\n")


if __name__ == "__main__":
    try:
        path = sys.argv[1]
        metadata = extract_metadata(path)
        print(json.dumps(metadata, ensure_ascii=False))
    except Exception:
        save_traceback(path)
        print(json.dumps(OUTPUT_KEYS, ensure_ascii=False))
