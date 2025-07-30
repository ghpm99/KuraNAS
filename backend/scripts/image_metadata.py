import sys
import json
import warnings
from PIL import Image, ExifTags, ImageCms
from io import BytesIO

warnings.filterwarnings("ignore", category=UserWarning, module="PIL.TiffImagePlugin")


def serialize_value(value):
    try:
        if isinstance(value, bytes):
            return value.decode(errors="replace")
        elif hasattr(value, "numerator") and hasattr(value, "denominator"):
            return float(value)
        elif isinstance(value, (list, tuple)):
            return [serialize_value(v) for v in value]
        elif isinstance(value, dict):
            return {str(k): serialize_value(v) for k, v in value.items()}
        else:
            return value
    except Exception:
        return str(value)


def get_exif_dict(img):
    exif_data = img._getexif()
    if not exif_data:
        return {}

    exif = {}
    for tag_id, value in exif_data.items():
        tag = ExifTags.TAGS.get(tag_id, tag_id)
        exif[tag] = serialize_value(value)
    return exif


def parse_icc_profile(icc_bytes):
    try:
        profile = ImageCms.ImageCmsProfile(BytesIO(icc_bytes))
        return ImageCms.getProfileDescription(profile)
    except Exception:
        return ""


def extract_gps(exif):
    gps = exif.get("GPSInfo", {})
    gps_lat = gps_long = 0.0

    def convert_gps(coord, ref):
        try:
            deg, min_, sec = coord
            decimal = float(deg) + float(min_) / 60 + float(sec) / 3600
            if ref in ["S", "W"]:
                decimal = -decimal
            return decimal
        except Exception:
            return 0.0

    if gps:
        lat = gps.get(2)
        lat_ref = gps.get(1)
        lon = gps.get(4)
        lon_ref = gps.get(3)

        if lat and lat_ref:
            gps_lat = convert_gps(lat, lat_ref)
        if lon and lon_ref:
            gps_long = convert_gps(lon, lon_ref)

    return gps_lat, gps_long


def extract_image_metadata(image_path):
    try:
        with Image.open(image_path) as img:
            exif = get_exif_dict(img)
            gps_lat, gps_long = extract_gps(exif)

            dpi = img.info.get("dpi", (0, 0))
            icc_profile = parse_icc_profile(img.info.get("icc_profile", b""))

            metadata = {
                "format": img.format or "",
                "mode": img.mode or "",
                "width": img.width,
                "height": img.height,
                "capture_date": exif.get("DateTimeOriginal", ""),
                "software": exif.get("Software", ""),
                "make": exif.get("Make", ""),
                "model": exif.get("Model", ""),
                "lens_model": exif.get("LensModel", ""),
                "iso": exif.get("ISOSpeedRatings", 0),
                "exposure_time": exif.get("ExposureTime", ""),
                "dpi_x": dpi[0],
                "dpi_y": dpi[1],
                "icc_profile": icc_profile,
                "gps_latitude": gps_lat,
                "gps_longitude": gps_long,
            }
            return metadata
    except Exception:
        return {
            "format": "",
            "mode": "",
            "width": 0,
            "height": 0,
            "capture_date": "",
            "software": "",
            "make": "",
            "model": "",
            "lens_model": "",
            "iso": 0,
            "exposure_time": "",
            "dpi_x": 0,
            "dpi_y": 0,
            "icc_profile": "",
            "gps_latitude": "",
            "gps_longitude": "",
        }


if __name__ == "__main__":
    if len(sys.argv) < 2:
        sys.exit(1)

    image_path = sys.argv[1]
    metadata = extract_image_metadata(image_path)
    print(json.dumps(metadata, ensure_ascii=False))
