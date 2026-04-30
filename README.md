# photofs

A virtual filesystem that organises photos from an SD card by month, served over [9P](https://en.wikipedia.org/wiki/9P_(protocol)).

## How it works

Point photofs at a directory (e.g. an SD card) and mount the 9P share. Photos are grouped by their EXIF `DateTimeOriginal` into a directory tree:

```
/
├── jpg/
│   ├── 2024-01/
│   │   ├── IMG_0001.jpg
│   │   └── IMG_0002.jpg
│   └── 2024-02/
│       └── IMG_0042.jpg
├── raw/
│   ├── 2024-01/
│   │   └── IMG_0001.cr3
│   └── 2024-02/
│       └── IMG_0042.cr3
└── undated/
    └── IMG_0099.jpg
```

Files are read-only and read directly from disk — nothing is copied or moved.

## Requirements

- [exiftool](https://exiftool.org/) must be installed and on `$PATH`

Supported RAW formats: `.cr2`, `.cr3`, `.nef`, `.arw`, `.dng`, `.raf`, `.orf`, `.rw2`

## Usage

```sh
go build
./photofs --path /path/to/sdcard --addr localhost:8000
```

Then mount the share. On Linux with `9pfuse` (from [plan9port](https://9fans.github.io/plan9port/)):

```sh
9pfuse localhost:8000 /mnt/photos
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--path` | *(required)* | Path to the SD card or photo directory |
| `--addr` | `localhost:8000` | TCP address to serve 9P on |

## Notes

- Files without a `DateTimeOriginal` EXIF tag appear in `undated/`
- The directory structure is built at startup; restart to pick up new files
