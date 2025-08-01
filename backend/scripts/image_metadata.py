import json
import os
import sys
import traceback
import warnings
import ast
from io import BytesIO

from PIL import ExifTags, Image, ImageCms
from PIL.TiffImagePlugin import IFDRational

warnings.filterwarnings("ignore", category=UserWarning, module="PIL.TiffImagePlugin")

# Mapeia os nomes legíveis das tags EXIF
EXIF_TAGS = {v: k for k, v in ExifTags.TAGS.items()}

RESULT_DEFAULT = {
    "format": "",
    "mode": "",
    "width": 0,
    "height": 0,
    "dpi_x": 0,
    "dpi_y": 0,
    "x_resolution": 0,
    "y_resolution": 0,
    "resolution_unit": 0,
    "orientation": 0,
    "compression": 0,
    "photometric_interpretation": 0,
    "color_space": 0,
    "components_configuration": "",
    "icc_profile": "",
    "make": "",
    "model": "",
    "software": "",
    "lens_model": "",
    "serial_number": "",
    "datetime": "",
    "datetime_original": "",
    "datetime_digitized": "",
    "subsec_time": "",
    "exposure_time": 0,
    "f_number": 0,
    "iso": 0,
    "shutter_speed": 0,
    "aperture_value": 0,
    "brightness_value": 0,
    "exposure_bias": 0,
    "metering_mode": 0,
    "flash": 0,
    "focal_length": 0,
    "white_balance": 0,
    "exposure_program": 0,
    "max_aperture_value": 0,
    "gps_latitude": 0,
    "gps_longitude": 0,
    "gps_altitude": 0,
    "gps_date": "",
    "gps_time": "",
    "image_description": "",
    "user_comment": "",
    "copyright": "",
    "artist": "",
}


def safe_decode(value):
    try:
        if isinstance(value, bytes):
            return value.decode(errors="replace")
        elif isinstance(value, tuple):
            return [safe_decode(v) for v in value]
        elif hasattr(value, "numerator") and hasattr(value, "denominator"):
            # Para Rational: transforma em float, e arredonda para 6 casas
            return round(float(value.numerator) / float(value.denominator), 6) if value.denominator != 0 else 0
        elif isinstance(value, (int, float, str)):
            return value
        elif isinstance(value, IFDRational):
            # Para IFDRational: transforma em float, e arredonda para 6 casas
            return round(float(value.numerator) / float(value.denominator), 6) if value.denominator != 0 else 0
        else:
            return str(value)
    except Exception:
        save_traceback(f"{value}-{type(value)}")
        return ""


def parse_icc_profile(icc_bytes):
    try:
        profile = ImageCms.ImageCmsProfile(BytesIO(icc_bytes))
        return ImageCms.getProfileDescription(profile)
    except Exception:
        save_traceback("")
        return ""


def parse_coord(coord, ref, image_path=""):
    try:
        if not isinstance(coord, (list, tuple)) or len(coord) < 3:
            return 0

        def to_float(rational):
            if isinstance(rational, (list, tuple)) and len(rational) == 2:
                return rational[0] / rational[1] if rational[1] != 0 else 0
            elif isinstance(rational, (int, float)):
                return rational
            else:
                return 0

        deg = to_float(coord[0])
        minute = to_float(coord[1])
        second = to_float(coord[2])

        decimal = deg + (minute / 60.0) + (second / 3600.0)
        return round(decimal if ref in ["N", "E"] else -decimal, 8)

    except Exception:
        save_traceback(image_path)
        return 0


def format_gps_time(gps_time):
    if isinstance(gps_time, (list, tuple)) and len(gps_time) == 3:
        try:
            h = int(round(gps_time[0]))
            m = int(round(gps_time[1]))
            s = int(round(gps_time[2]))
            return f"{h:02}:{m:02}:{s:02}"
        except Exception:
            return ""
    return ""


def extract_metadata(image_path):
    result = RESULT_DEFAULT.copy()
    try:
        with Image.open(image_path) as img:
            result["format"] = img.format or ""
            result["mode"] = img.mode or ""
            result["width"] = img.width or 0
            result["height"] = img.height or 0

            dpi = img.info.get("dpi", (0, 0))
            dpi_x, dpi_y = dpi if isinstance(dpi, tuple) else (dpi, dpi)

            result["dpi_x"] = safe_decode(dpi_x)
            result["dpi_y"] = safe_decode(dpi_y)

            result["icc_profile"] = (
                parse_icc_profile(img.info.get("icc_profile", b"")) if "icc_profile" in img.info else ""
            )

            exif_data = {}
            if hasattr(img, "_getexif"):
                raw_exif = img._getexif()
                if raw_exif:
                    for tag, val in raw_exif.items():
                        tag_name = ExifTags.TAGS.get(tag, tag)
                        exif_data[tag_name] = safe_decode(val)

            def get(tag, default=""):
                return safe_decode(exif_data.get(tag, default))

            # Popula os campos
            result["x_resolution"] = get("XResolution", 0)
            result["y_resolution"] = get("YResolution", 0)
            result["resolution_unit"] = get("ResolutionUnit", 0)
            result["orientation"] = get("Orientation", 0)
            result["compression"] = get("Compression", 0)
            result["photometric_interpretation"] = get("PhotometricInterpretation", 0)
            result["color_space"] = get("ColorSpace", 0)
            result["components_configuration"] = get("ComponentsConfiguration")

            result["make"] = get("Make")
            result["model"] = get("Model")
            result["software"] = get("Software")
            result["lens_model"] = get("LensModel")
            result["serial_number"] = get("BodySerialNumber")

            result["datetime"] = get("DateTime")
            result["datetime_original"] = get("DateTimeOriginal")
            result["datetime_digitized"] = get("DateTimeDigitized")
            result["subsec_time"] = get("SubSecTime")

            result["exposure_time"] = get("ExposureTime", 0)
            result["f_number"] = get("FNumber", 0)
            result["iso"] = get("ISOSpeedRatings", 0)
            result["shutter_speed"] = get("ShutterSpeedValue", 0)
            result["aperture_value"] = get("ApertureValue", 0)
            result["brightness_value"] = get("BrightnessValue", 0)
            result["exposure_bias"] = get("ExposureBiasValue", 0)
            result["metering_mode"] = get("MeteringMode", 0)
            result["flash"] = get("Flash", 0)
            result["focal_length"] = get("FocalLength", 0)
            result["white_balance"] = get("WhiteBalance", 0)
            result["exposure_program"] = get("ExposureProgram", 0)
            result["max_aperture_value"] = get("MaxApertureValue", 0)

            # GPS
            gps = exif_data.get("GPSInfo", {})
            if gps:
                gps_dict = ast.literal_eval(gps)
                gps_tags = {}
                for key in gps_dict.keys():
                    decode = ExifTags.GPSTAGS.get(key, key)
                    gps_tags[decode] = gps_dict[key]

                result["gps_latitude"] = parse_coord(gps_tags.get("GPSLatitude", []), gps_tags.get("GPSLatitudeRef", 0))
                result["gps_longitude"] = parse_coord(
                    gps_tags.get("GPSLongitude", []), gps_tags.get("GPSLongitudeRef", 0)
                )
                result["gps_altitude"] = safe_decode(gps_tags.get("GPSAltitude", ""))
                result["gps_date"] = gps_tags.get("GPSDateStamp", "")
                result["gps_time"] = format_gps_time(gps_tags.get("GPSTimeStamp", ""))

            result["image_description"] = get("ImageDescription")
            result["user_comment"] = get("UserComment")
            result["copyright"] = get("Copyright")
            result["artist"] = get("Artist")

    except Exception:
        save_traceback(image_path)

    return result


def save_traceback(path):
    base_dir = os.path.dirname(os.path.abspath(__file__))
    logs_dir = os.path.join(base_dir, "logs")
    os.makedirs(logs_dir, exist_ok=True)
    log_path = os.path.join(logs_dir, "image_metadata.log")
    with open(log_path, "a", encoding="utf-8") as f:
        f.write(f"Erro ao processar imagem {path}:\n")
        f.write(traceback.format_exc())
        f.write("\n")


if __name__ == "__main__":
    try:
        path = sys.argv[1] if len(sys.argv) > 1 else ""
        metadata = extract_metadata(path)
        print(json.dumps(metadata, ensure_ascii=False))
    except Exception:
        save_traceback(path)
        print(json.dumps(RESULT_DEFAULT, ensure_ascii=False))
