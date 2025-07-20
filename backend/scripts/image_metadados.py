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
            metadata = {
                "format": img.format,
                "mode": img.mode,
                "size": img.size,
                "info": img.info,
            }
            print(metadata)
            return metadata
    except Exception as e:
        print(f"Error extracting metadata from {image_path}: {e}")
        return None


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Uso: python image_metadados.py <caminho_da_imagem>")
        sys.exit(1)
    image_path = sys.argv[1]
    extract_image_metadata(image_path)  # Replace with your image path
