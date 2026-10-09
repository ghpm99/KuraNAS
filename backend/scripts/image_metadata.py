import json
import os
import sys
import traceback
import warnings
import ast
from io import BytesIO

from PIL import ExifTags, Image, ImageCms
from PIL.TiffImagePlugin import IFDRational

try:
    import pillow_heif

    pillow_heif.register_heif_opener()
except ImportError:
    pass

warnings.filterwarnings("ignore", category=UserWarning, module="PIL.TiffImagePlugin")

EXIF_TAGS = {v: k for k, v in ExifTags.TAGS.items()}

EXIF_IFD_POINTER = 0x8769
GPS_IFD_POINTER = 0x8825
RAW_EXTENSIONS = {".cr2", ".cr3", ".nef", ".arw", ".dng", ".orf", ".rw2", ".raf", ".srw", ".pef"}
RAW_SCAN_LIMIT_BYTES = 64 * 1024 * 1024
JPEG_SIGNATURE = b"\xff\xd8\xff"
JPEG_DECODABLE_FRAME_MARKERS = {0xC0, 0xC1, 0xC2}

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
            decoded = value.decode(errors="replace")
            return decoded.replace("\x00", "")
        elif isinstance(value, tuple):
            return [safe_decode(v) for v in value]
        elif hasattr(value, "numerator") and hasattr(value, "denominator"):
            # Para Rational: transforma em float, e arredonda para 6 casas
            return round(float(value.numerator) / float(value.denominator), 6) if value.denominator != 0 else 0
        elif isinstance(value, (int, float, str)):
            if isinstance(value, str):
                return value.replace("\x00", "")
            return value
        elif isinstance(value, IFDRational):
            # Para IFDRational: transforma em float, e arredonda para 6 casas
            return round(float(value.numerator) / float(value.denominator), 6) if value.denominator != 0 else 0
        else:
            return str(value).replace("\x00", "")
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


def measure_jpeg_stream(data, start):
    position = start + 2
    is_decodable = False
    while position + 2 <= len(data):
        if data[position] != 0xFF:
            return 0
        marker = data[position + 1]
        if marker == 0xFF:
            position += 1
            continue
        if marker == 0xD9:
            return position + 2 - start if is_decodable else 0
        if marker == 0x01 or 0xD0 <= marker <= 0xD7:
            position += 2
            continue
        if position + 4 > len(data):
            return 0
        segment_length = int.from_bytes(data[position + 2 : position + 4], "big")
        if segment_length < 2 or position + 2 + segment_length > len(data):
            return 0
        if 0xC0 <= marker <= 0xCF and marker not in (0xC4, 0xC8, 0xCC):
            is_decodable = marker in JPEG_DECODABLE_FRAME_MARKERS
        position += 2 + segment_length
        if marker == 0xDA:
            position = skip_entropy_coded_data(data, position)
    return 0


def skip_entropy_coded_data(data, position):
    while position + 1 < len(data):
        if data[position] != 0xFF:
            position += 1
            continue
        next_byte = data[position + 1]
        if next_byte != 0x00 and not 0xD0 <= next_byte <= 0xD7:
            return position
        position += 2
    return len(data)


def find_largest_embedded_jpeg(image_path):
    with open(image_path, "rb") as source:
        content = source.read(RAW_SCAN_LIMIT_BYTES)

    largest = b""
    search_from = 0
    while True:
        start = content.find(JPEG_SIGNATURE, search_from)
        if start < 0:
            break
        length = measure_jpeg_stream(content, start)
        if length == 0:
            search_from = start + 1
            continue
        if length > len(largest):
            largest = content[start : start + length]
        search_from = start + length
    return largest


def open_image(image_path):
    try:
        return Image.open(image_path)
    except Exception:
        if os.path.splitext(image_path)[1].lower() not in RAW_EXTENSIONS:
            raise
    embedded_jpeg = find_largest_embedded_jpeg(image_path)
    if not embedded_jpeg:
        raise ValueError("no embedded preview")
    return Image.open(BytesIO(embedded_jpeg))


def read_exif_tags(img):
    if hasattr(img, "_getexif"):
        raw_exif = img._getexif() or {}
        return {ExifTags.TAGS.get(tag, tag): safe_decode(value) for tag, value in raw_exif.items()}

    exif = img.getexif()
    raw_exif = dict(exif)
    raw_exif.update(exif.get_ifd(EXIF_IFD_POINTER))
    exif_data = {ExifTags.TAGS.get(tag, tag): safe_decode(value) for tag, value in raw_exif.items()}
    gps_ifd = exif.get_ifd(GPS_IFD_POINTER)
    if gps_ifd:
        exif_data["GPSInfo"] = str({key: tuple_of_floats(value) for key, value in gps_ifd.items()})
    return exif_data


def tuple_of_floats(value):
    if isinstance(value, tuple):
        return tuple(float(part) if hasattr(part, "numerator") else part for part in value)
    if hasattr(value, "numerator"):
        return float(value)
    return value


def read_gps_fields(exif_data, result):
    gps = exif_data.get("GPSInfo", {})
    if not gps:
        return
    try:
        gps_dict = ast.literal_eval(gps)
        gps_tags = {ExifTags.GPSTAGS.get(key, key): value for key, value in gps_dict.items()}

        result["gps_latitude"] = parse_coord(gps_tags.get("GPSLatitude", []), gps_tags.get("GPSLatitudeRef", 0))
        result["gps_longitude"] = parse_coord(gps_tags.get("GPSLongitude", []), gps_tags.get("GPSLongitudeRef", 0))
        result["gps_altitude"] = safe_decode(gps_tags.get("GPSAltitude", 0))
        result["gps_date"] = gps_tags.get("GPSDateStamp", "")
        result["gps_time"] = format_gps_time(gps_tags.get("GPSTimeStamp", ""))
    except Exception:
        save_traceback("gps")


def extract_metadata(image_path):
    result = RESULT_DEFAULT.copy()
    try:
        with open_image(image_path) as img:
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

            exif_data = read_exif_tags(img)

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
            result["flash"] = int(get("Flash", 0))
            result["focal_length"] = get("FocalLength", 0)
            result["white_balance"] = get("WhiteBalance", 0)
            result["exposure_program"] = get("ExposureProgram", 0)
            result["max_aperture_value"] = get("MaxApertureValue", 0)

            read_gps_fields(exif_data, result)

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
