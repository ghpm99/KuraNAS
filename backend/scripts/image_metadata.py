import sys
import json
from io import BytesIO
from PIL import Image, ImageCms
from PIL.ExifTags import TAGS


def serialize_value(v):
    if isinstance(v, bytes):
        try:
            return v.decode(errors="replace")
        except Exception:
            return str(v)
    elif hasattr(v, "numerator") and hasattr(v, "denominator"):
        try:
            return int(v)
        except ZeroDivisionError:
            return None
    elif isinstance(v, (list, tuple)):
        return [serialize_value(i) for i in v]
    elif isinstance(v, dict):
        return {serialize_value(k): serialize_value(val) for k, val in v.items()}
    else:
        return v


def serialize_info(info):
    return {k: serialize_value(v) for k, v in info.items()}


def get_exif_data(img):
    """
    Extrai os dados EXIF da imagem, se disponíveis.
    """
    exif = {}
    if hasattr(img, "_getexif"):
        raw_exif = img._getexif()
        if raw_exif:
            for tag, value in raw_exif.items():
                tag_name = TAGS.get(tag, tag)
                exif[tag_name] = value
    return exif


def parse_icc_profile(icc_bytes):
    """
    Tenta extrair informações legíveis de um perfil ICC.
    """
    try:
        profile = ImageCms.ImageCmsProfile(BytesIO(icc_bytes))
        desc = ImageCms.getProfileDescription(profile)
        return desc
    except Exception as e:
        return f"Perfil ICC não legível ({e})"


def extract_image_metadata(image_path):
    try:
        with Image.open(image_path) as img:
            exif = get_exif_data(img)
            data = {
                "format": img.format,
                "mode": img.mode,
                "width": int(img.width),
                "height": int(img.height),
                "info": {
                    "datetime": img.info.get("datetime", ""),
                    "exif": exif,
                    "icc_profile": parse_icc_profile(img.info["icc_profile"]) if "icc_profile" in img.info else "",
                    "dpi": img.info.get("dpi", (0, 0)),
                    "compression": img.info.get("compression", ""),
                    "transparency": img.info.get("transparency", ""),
                    "gamma": img.info.get("gamma", ""),
                    "background": img.info.get("background", ""),
                    "interlace": img.info.get("interlace", ""),
                    "palette": img.getpalette() if img.mode == "P" else "",
                    "channels": len(img.getbands()),
                    "has_alpha": "A" in img.getbands(),
                    "comments": img.info.get("comments", ""),
                },
            }
            return serialize_info(data)

    except Exception as e:
        return {
            "format": "",
            "mode": "",
            "width": 0,
            "height": 0,
            "info": {},
        }


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Uso: python image_metadata.py <caminho_da_imagem>")
        sys.exit(1)

    image_path = sys.argv[1]

    try:
        metadata = extract_image_metadata(image_path)
        print(json.dumps(metadata, indent=2, ensure_ascii=False))
    except Exception as e:
        print(
            {
                "format": "",
                "mode": "",
                "width": 0,
                "height": 0,
                "info": {},
            }
        )
