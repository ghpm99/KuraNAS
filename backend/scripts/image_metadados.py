import sys
from PIL import Image


def extract_image_metadata(image_path):
    """
    Extracts metadata from an image file.

    Args:
        image_path (str): The path to the image file.

    Returns:
        dict: A dictionary containing the image metadata.
    """
    try:
        with Image.open(image_path) as img:
            return {
                "format": img.format,
                "mode": img.mode,
                "width": img.width,
                "height": img.height,
                "info": img.info,
            }

    except:
        return {
            "format": "",
            "mode": "",
            "width": 0,
            "height": 0,
            "info": {},
        }


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Uso: python image_metadados.py <caminho_da_imagem>")
        sys.exit(1)
    image_path = sys.argv[1]
    extract_image_metadata(image_path)  # Replace with your image path
