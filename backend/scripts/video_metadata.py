import sys
import subprocess
import json


def extract_video_metadata(video_path):
    try:
        result = subprocess.run(
            [
                "ffprobe",
                "-v",
                "error",
                "-show_entries",
                "format=duration,bit_rate,size",
                "-show_streams",
                "-of",
                "json",
                video_path,
            ],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
        )
        if result.returncode != 0:
            return {"format": {"duration": 0, "size": 0, "bit_rate": 0}}
        metadata = json.loads(result.stdout)
        return metadata
    except:
        return {"format": {"duration": 0, "size": 0, "bit_rate": 0}}


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Uso: python video_metadata.py <caminho_do_video>")
        sys.exit(1)
    video_path = sys.argv[1]
    metadata = extract_video_metadata(video_path)
    print(json.dumps(metadata, indent=2, ensure_ascii=False))
