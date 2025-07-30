import subprocess
import json
import sys
import os


def extract_video_metadata(path):
    if not os.path.isfile(path):
        return default_metadata(path)

    try:
        cmd = ["ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", path]
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=10)

        if result.returncode != 0:
            return default_metadata(path)

        data = json.loads(result.stdout)
        streams = data.get("streams", [])
        format_info = data.get("format", {})

        # Video stream
        video_stream = next((s for s in streams if s.get("codec_type") == "video"), {})
        # Audio stream
        audio_stream = next((s for s in streams if s.get("codec_type") == "audio"), {})

        return {
            "filename": os.path.basename(path),
            "format_name": format_info.get("format_name", ""),
            "size": format_info.get("size", ""),
            "duration": format_info.get("duration", ""),
            "width": video_stream.get("width", 0),
            "height": video_stream.get("height", 0),
            "frame_rate": _parse_frame_rate(video_stream.get("avg_frame_rate", "")),
            "nb_frames": int(video_stream.get("nb_frames", 0)) if video_stream.get("nb_frames") else 0,
            "bit_rate": video_stream.get("bit_rate", format_info.get("bit_rate", "")),
            "codec_name": video_stream.get("codec_name", ""),
            "codec_long_name": video_stream.get("codec_long_name", ""),
            "pix_fmt": video_stream.get("pix_fmt", ""),
            "level": video_stream.get("level", 0),
            "profile": video_stream.get("profile", ""),
            "aspect_ratio": video_stream.get("display_aspect_ratio", ""),
            "audio_codec": audio_stream.get("codec_name", ""),
            "audio_channels": audio_stream.get("channels", 0),
            "audio_sample_rate": audio_stream.get("sample_rate", ""),
            "audio_bit_rate": audio_stream.get("bit_rate", ""),
        }

    except Exception:
        return default_metadata(path)


def _parse_frame_rate(rate_str):
    if not rate_str or rate_str == "0/0":
        return 0
    try:
        num, denom = rate_str.split("/")
        return round(float(num) / float(denom), 2) if float(denom) != 0 else 0
    except:
        return 0


def default_metadata(path=""):
    return {
        "format_name": "",
        "size": "",
        "duration": "",
        "width": 0,
        "height": 0,
        "frame_rate": 0,
        "nb_frames": 0,
        "bit_rate": "",
        "codec_name": "",
        "codec_long_name": "",
        "pix_fmt": "",
        "level": 0,
        "profile": "",
        "aspect_ratio": "",
        "audio_codec": "",
        "audio_channels": 0,
        "audio_sample_rate": "",
        "audio_bit_rate": "",
    }


if __name__ == "__main__":
    video_path = sys.argv[1] if len(sys.argv) > 1 else ""
    metadata = extract_video_metadata(video_path)
    print(json.dumps(metadata, ensure_ascii=False))
