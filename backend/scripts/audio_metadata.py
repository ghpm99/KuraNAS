import json
import os
import re
import sys
import traceback

from mutagen import File

MAX_LYRICS_LENGTH = 20000

OUTPUT_KEYS = {
    "mime": "",
    "length": 0.0,
    "bitrate": 0,
    "sample_rate": 0,
    "channels": 0,
    "bitrate_mode": 0,
    "encoder_info": "",
    "bit_depth": 0,
    "title": "",
    "artist": "",
    "album": "",
    "album_artist": "",
    "track_number": "",
    "disc_number": "",
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

EASY_TAG_KEYS = {
    "title": ["title"],
    "artist": ["artist"],
    "album": ["album"],
    "album_artist": ["albumartist"],
    "track_number": ["tracknumber"],
    "disc_number": ["discnumber"],
    "genre": ["genre"],
    "composer": ["composer"],
    "recording_date": ["date"],
    "encoder": ["encodedby", "encoder"],
    "publisher": ["organization", "publisher", "label"],
    "original_release_date": ["originaldate"],
    "original_artist": ["originalartist"],
    "lyricist": ["lyricist"],
    "lyrics": ["lyrics", "unsyncedlyrics"],
}

ID3_FRAME_KEYS = {
    "title": ["TIT2"],
    "artist": ["TPE1"],
    "album": ["TALB"],
    "album_artist": ["TPE2"],
    "track_number": ["TRCK"],
    "disc_number": ["TPOS"],
    "genre": ["TCON"],
    "composer": ["TCOM"],
    "recording_date": ["TDRC", "TYER"],
    "encoder": ["TENC"],
    "publisher": ["TPUB"],
    "original_release_date": ["TDOR", "TORY"],
    "original_artist": ["TOPE"],
    "lyricist": ["TEXT"],
}

MP4_ATOM_KEYS = {
    "title": ["\xa9nam"],
    "artist": ["\xa9ART"],
    "album": ["\xa9alb"],
    "album_artist": ["aART"],
    "track_number": ["trkn"],
    "disc_number": ["disk"],
    "genre": ["\xa9gen"],
    "composer": ["\xa9wrt"],
    "recording_date": ["\xa9day"],
    "encoder": ["\xa9too"],
    "lyrics": ["\xa9lyr"],
}

YEAR_PATTERN = re.compile(r"\d{4}")


def stringify_tag_value(raw_value):
    if raw_value is None:
        return ""
    if isinstance(raw_value, bytes):
        return raw_value.decode("utf-8", errors="ignore").strip()
    if isinstance(raw_value, str):
        return raw_value.strip()
    if isinstance(raw_value, (int, float)):
        return str(raw_value)
    if isinstance(raw_value, tuple):
        numbers = [str(number) for number in raw_value if number]
        return "/".join(numbers)
    if hasattr(raw_value, "text"):
        return stringify_tag_value(raw_value.text)
    if isinstance(raw_value, (list, set)):
        for candidate in raw_value:
            candidate_text = stringify_tag_value(candidate)
            if candidate_text:
                return candidate_text
        return ""
    return str(raw_value).strip()


def first_non_empty_tag(tags, tag_keys):
    for tag_key in tag_keys:
        try:
            raw_value = tags.get(tag_key)
        except Exception:
            continue
        tag_text = stringify_tag_value(raw_value)
        if tag_text:
            return tag_text
    return ""


def extract_year(*date_candidates):
    for date_candidate in date_candidates:
        year_match = YEAR_PATTERN.search(date_candidate or "")
        if year_match:
            return year_match.group(0)
    return ""


def read_technical_info(audio, output):
    mime_types = getattr(audio, "mime", None)
    if mime_types:
        output["mime"] = mime_types[0]

    info = getattr(audio, "info", None)
    if not info:
        return

    output["length"] = getattr(info, "length", 0.0)
    output["bitrate"] = getattr(info, "bitrate", 0)
    output["sample_rate"] = getattr(info, "sample_rate", 0)
    output["channels"] = getattr(info, "channels", 0)
    output["bitrate_mode"] = int(getattr(info, "bitrate_mode", 0) or 0)
    output["encoder_info"] = str(getattr(info, "encoder_info", "") or "")
    output["bit_depth"] = getattr(info, "bits_per_sample", 0)


def read_easy_tags(path):
    easy_audio = File(path, easy=True)
    if easy_audio is None or easy_audio.tags is None:
        return {}

    return {
        output_key: first_non_empty_tag(easy_audio.tags, tag_keys)
        for output_key, tag_keys in EASY_TAG_KEYS.items()
    }


def read_id3_lyrics(tags):
    for lyrics_frame in tags.getall("USLT"):
        lyrics_text = stringify_tag_value(lyrics_frame.text)
        if lyrics_text:
            return lyrics_text
    return ""


def read_vorbis_lyrics(tags):
    return first_non_empty_tag(tags, ["lyrics", "unsyncedlyrics"])


def read_raw_tags(audio):
    raw_tags = getattr(audio, "tags", None)
    if raw_tags is None:
        return {}

    if hasattr(raw_tags, "getall"):
        id3_values = {
            output_key: first_non_empty_tag(raw_tags, frame_keys)
            for output_key, frame_keys in ID3_FRAME_KEYS.items()
        }
        id3_values["lyrics"] = read_id3_lyrics(raw_tags)
        id3_values["year"] = first_non_empty_tag(raw_tags, ["TYER", "TDRC"])
        return id3_values

    if any(isinstance(atom_key, str) and atom_key.startswith("\xa9") for atom_key in raw_tags.keys()):
        return {
            output_key: first_non_empty_tag(raw_tags, atom_keys)
            for output_key, atom_keys in MP4_ATOM_KEYS.items()
        }

    return {"lyrics": read_vorbis_lyrics(raw_tags)}


def merge_tags(preferred_tags, fallback_tags):
    merged_tags = dict(fallback_tags)
    for tag_name, tag_text in preferred_tags.items():
        if tag_text:
            merged_tags[tag_name] = tag_text
    return merged_tags


def apply_tags(tags, output):
    for output_key in OUTPUT_KEYS:
        if output_key in tags and isinstance(OUTPUT_KEYS[output_key], str):
            output[output_key] = tags[output_key]

    output["year"] = extract_year(
        tags.get("recording_date"),
        tags.get("year"),
        tags.get("original_release_date"),
    )
    output["lyrics"] = output["lyrics"][:MAX_LYRICS_LENGTH]


def extract_metadata(path):
    output = OUTPUT_KEYS.copy()

    try:
        audio = File(path, easy=False)
    except Exception:
        save_traceback(path)
        return output
    if audio is None:
        return output

    try:
        read_technical_info(audio, output)
    except Exception:
        save_traceback(path)

    easy_tags = {}
    try:
        easy_tags = read_easy_tags(path)
    except Exception:
        save_traceback(path)

    raw_tags = {}
    try:
        raw_tags = read_raw_tags(audio)
    except Exception:
        save_traceback(path)

    try:
        apply_tags(merge_tags(easy_tags, raw_tags), output)
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
    path = sys.argv[1] if len(sys.argv) > 1 else ""
    try:
        metadata = extract_metadata(path)
        print(json.dumps(metadata, ensure_ascii=False))
    except Exception:
        save_traceback(path)
        print(json.dumps(OUTPUT_KEYS, ensure_ascii=False))
