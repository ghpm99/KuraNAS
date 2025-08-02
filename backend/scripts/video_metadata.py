import json
import os
import subprocess
import sys
import traceback

RESULT_DEFAULT = {
    "filename": "",
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


def extract_video_metadata(path):

    result = RESULT_DEFAULT.copy()
    result["filename"] = os.path.basename(path)

    if not os.path.isfile(path):
        return result

    try:
        cmd = ["ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", path]
        process = subprocess.run(cmd, capture_output=True, text=True, timeout=10)

        if process.returncode != 0:
            return result

        data = json.loads(process.stdout)
        streams = data.get("streams", [])
        format_info = data.get("format", {})

        # Video stream
        video_stream = next((s for s in streams if s.get("codec_type") == "video"), {})
        # Audio stream
        audio_stream = next((s for s in streams if s.get("codec_type") == "audio"), {})

        result["format_name"] = format_info.get("format_name", "")
        result["size"] = format_info.get("size", "")
        result["duration"] = format_info.get("duration", "")
        result["width"] = video_stream.get("width", 0)
        result["height"] = video_stream.get("height", 0)
        result["frame_rate"] = _parse_frame_rate(video_stream.get("avg_frame_rate", ""))
        result["nb_frames"] = int(video_stream.get("nb_frames", 0)) if video_stream.get("nb_frames") else 0
        result["bit_rate"] = video_stream.get("bit_rate", format_info.get("bit_rate", ""))
        result["codec_name"] = video_stream.get("codec_name", "")
        result["codec_long_name"] = video_stream.get("codec_long_name", "")
        result["pix_fmt"] = video_stream.get("pix_fmt", "")
        result["level"] = video_stream.get("level", 0)
        result["profile"] = video_stream.get("profile", "")
        result["aspect_ratio"] = video_stream.get("display_aspect_ratio", "")
        result["audio_codec"] = audio_stream.get("codec_name", "")
        result["audio_channels"] = audio_stream.get("channels", 0)
        result["audio_sample_rate"] = audio_stream.get("sample_rate", "")
        result["audio_bit_rate"] = audio_stream.get("bit_rate", "")

    except Exception:
        save_traceback(path)

    return result


def _parse_frame_rate(rate_str):
    if not rate_str or rate_str == "0/0":
        return 0
    try:
        num, denom = rate_str.split("/")
        return round(float(num) / float(denom), 2) if float(denom) != 0 else 0
    except:
        return 0


def save_traceback(path):
    base_dir = os.path.dirname(os.path.abspath(__file__))
    logs_dir = os.path.join(base_dir, "logs")
    os.makedirs(logs_dir, exist_ok=True)
    log_path = os.path.join(logs_dir, "video_metadata.log")
    with open(log_path, "a", encoding="utf-8") as f:
        f.write(f"Erro ao processar video {path}:\n")
        f.write(traceback.format_exc())
        f.write("\n")


if __name__ == "__main__":
    try:
        video_path = sys.argv[1] if len(sys.argv) > 1 else ""
        metadata = extract_video_metadata(video_path)
        print(json.dumps(metadata, ensure_ascii=False))
    except Exception:
        save_traceback(video_path)
        print(json.dumps(RESULT_DEFAULT, ensure_ascii=False))
