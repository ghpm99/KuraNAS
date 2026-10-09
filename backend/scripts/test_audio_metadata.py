import os
import sys
import unittest

try:
    import mutagen
except ImportError:
    raise unittest.SkipTest("mutagen is not installed")

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

import audio_metadata
from mutagen.id3 import ID3, TDRC, TIT2, TPE1, TPOS, TRCK, USLT
from mutagen.mp4 import MP4Tags
from mutagen.oggvorbis import OggVCommentDict


class ExtractYearTest(unittest.TestCase):
    def test_takes_first_four_digits_of_full_date(self):
        self.assertEqual(audio_metadata.extract_year("2019-05-01"), "2019")

    def test_falls_back_to_next_candidate(self):
        self.assertEqual(audio_metadata.extract_year("", None, "1987"), "1987")

    def test_returns_empty_without_digits(self):
        self.assertEqual(audio_metadata.extract_year("unknown", None), "")


class StringifyTagValueTest(unittest.TestCase):
    def test_mp4_pair_becomes_number_slash_total(self):
        self.assertEqual(audio_metadata.stringify_tag_value([(3, 12)]), "3/12")

    def test_mp4_pair_without_total(self):
        self.assertEqual(audio_metadata.stringify_tag_value([(3, 0)]), "3")

    def test_first_non_empty_list_entry_wins(self):
        self.assertEqual(audio_metadata.stringify_tag_value(["", "Second"]), "Second")


class ReadRawTagsTest(unittest.TestCase):
    class FakeAudio:
        def __init__(self, tags):
            self.tags = tags

    def test_id3_frames_including_lyrics_and_disc(self):
        id3_tags = ID3()
        id3_tags.add(TIT2(encoding=3, text="Song"))
        id3_tags.add(TPE1(encoding=3, text="Artist"))
        id3_tags.add(TRCK(encoding=3, text="3/12"))
        id3_tags.add(TPOS(encoding=3, text="1/2"))
        id3_tags.add(TDRC(encoding=3, text="2019-05-01"))
        id3_tags.add(USLT(encoding=3, lang="eng", desc="", text="la la la"))

        raw_tags = audio_metadata.read_raw_tags(self.FakeAudio(id3_tags))

        self.assertEqual(raw_tags["title"], "Song")
        self.assertEqual(raw_tags["artist"], "Artist")
        self.assertEqual(raw_tags["track_number"], "3/12")
        self.assertEqual(raw_tags["disc_number"], "1/2")
        self.assertEqual(raw_tags["lyrics"], "la la la")
        self.assertEqual(audio_metadata.extract_year(raw_tags["recording_date"]), "2019")

    def test_mp4_atoms(self):
        mp4_tags = MP4Tags()
        mp4_tags["\xa9nam"] = ["Song"]
        mp4_tags["trkn"] = [(3, 12)]
        mp4_tags["disk"] = [(1, 2)]
        mp4_tags["\xa9lyr"] = ["la la la"]

        raw_tags = audio_metadata.read_raw_tags(self.FakeAudio(mp4_tags))

        self.assertEqual(raw_tags["title"], "Song")
        self.assertEqual(raw_tags["track_number"], "3/12")
        self.assertEqual(raw_tags["disc_number"], "1/2")
        self.assertEqual(raw_tags["lyrics"], "la la la")

    def test_vorbis_lyrics_from_unsyncedlyrics(self):
        vorbis_tags = OggVCommentDict()
        vorbis_tags["UNSYNCEDLYRICS"] = ["la la la"]

        raw_tags = audio_metadata.read_raw_tags(self.FakeAudio(vorbis_tags))

        self.assertEqual(raw_tags["lyrics"], "la la la")

    def test_audio_without_tags_yields_empty_mapping(self):
        self.assertEqual(audio_metadata.read_raw_tags(self.FakeAudio(None)), {})


class ApplyTagsTest(unittest.TestCase):
    def test_easy_tags_override_raw_and_lyrics_are_truncated(self):
        merged_tags = audio_metadata.merge_tags(
            {"title": "Easy", "artist": ""},
            {"title": "Raw", "artist": "Raw Artist", "lyrics": "x" * 30000},
        )
        output = audio_metadata.OUTPUT_KEYS.copy()

        audio_metadata.apply_tags(merged_tags, output)

        self.assertEqual(output["title"], "Easy")
        self.assertEqual(output["artist"], "Raw Artist")
        self.assertEqual(len(output["lyrics"]), audio_metadata.MAX_LYRICS_LENGTH)

    def test_year_is_derived_from_recording_date(self):
        output = audio_metadata.OUTPUT_KEYS.copy()

        audio_metadata.apply_tags({"recording_date": "2019-05-01"}, output)

        self.assertEqual(output["year"], "2019")


class ExtractMetadataTest(unittest.TestCase):
    def test_unreadable_path_returns_all_output_keys_without_raising(self):
        output = audio_metadata.extract_metadata(os.path.join(os.devnull, "missing.flac"))

        self.assertEqual(set(output.keys()), set(audio_metadata.OUTPUT_KEYS.keys()))


if __name__ == "__main__":
    unittest.main()
