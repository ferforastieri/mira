#!/usr/bin/env python3
"""Draws Mira's logo, assets/mira.png, as the packages' icons.

Usage: icons.py png <size> <file>      a PNG, size pixels wide
       icons.py iconset <folder>       the PNGs of a macOS .iconset
       icons.py ico <file>             a Windows .ico, 16 to 256 pixels

The logo is resized with a box filter, keeping its transparency, so the
icons look the same as assets/mira.png at every size. Python's standard
library only: it runs on every runner.
"""
import os
import struct
import sys
import zlib

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
SOURCE = os.path.join(ROOT, "assets", "mira.png")

_source = None
_resized = {}


def read_png(path):
    """The RGBA rows of an 8-bit, non-interlaced PNG."""
    data = open(path, "rb").read()
    if data[:8] != b"\x89PNG\r\n\x1a\n":
        sys.exit("%s: not a PNG" % path)
    pos, idat = 8, []
    width = height = depth = color = interlace = None
    while pos < len(data):
        length = struct.unpack(">I", data[pos:pos + 4])[0]
        kind = data[pos + 4:pos + 8]
        body = data[pos + 8:pos + 8 + length]
        pos += 12 + length
        if kind == b"IHDR":
            width, height, depth, color, _, _, interlace = struct.unpack(">IIBBBBB", body)
        elif kind == b"IDAT":
            idat.append(body)
        elif kind == b"IEND":
            break
    if depth != 8 or color != 6 or interlace != 0:
        sys.exit("%s: expected an 8-bit RGBA PNG without interlacing" % path)
    raw = zlib.decompress(b"".join(idat))
    stride = width * 4
    rows, previous, at = [], bytearray(stride), 0
    for _ in range(height):
        filter_type = raw[at]
        at += 1
        line = bytearray(raw[at:at + stride])
        at += stride
        if filter_type == 1:
            for x in range(4, stride):
                line[x] = (line[x] + line[x - 4]) & 0xFF
        elif filter_type == 2:
            for x in range(stride):
                line[x] = (line[x] + previous[x]) & 0xFF
        elif filter_type == 3:
            for x in range(stride):
                left = line[x - 4] if x >= 4 else 0
                line[x] = (line[x] + ((left + previous[x]) >> 1)) & 0xFF
        elif filter_type == 4:
            for x in range(stride):
                left = line[x - 4] if x >= 4 else 0
                up = previous[x]
                corner = previous[x - 4] if x >= 4 else 0
                estimate = left + up - corner
                pa, pb, pc = abs(estimate - left), abs(estimate - up), abs(estimate - corner)
                nearest = left if pa <= pb and pa <= pc else (up if pb <= pc else corner)
                line[x] = (line[x] + nearest) & 0xFF
        elif filter_type != 0:
            sys.exit("%s: unknown PNG filter %d" % (path, filter_type))
        rows.append(bytes(line))
        previous = line
    return width, height, rows


def source():
    global _source
    if _source is None:
        _source = read_png(SOURCE)
    return _source


def resized(size):
    """The logo, size pixels wide, as rows of RGBA bytes."""
    if size in _resized:
        return _resized[size]
    width, height, rows = source()
    if size > width:
        sys.exit("icons.py: %d pixels is larger than %s" % (size, SOURCE))
    columns = [(x * width // size, max(x * width // size + 1, (x + 1) * width // size))
               for x in range(size)]
    out = []
    for y in range(size):
        top = y * height // size
        bottom = max(top + 1, (y + 1) * height // size)
        line = bytearray()
        for left, right in columns:
            red = green = blue = alpha = 0
            count = (right - left) * (bottom - top)
            for row_index in range(top, bottom):
                row = rows[row_index]
                for x in range(left, right):
                    at = x * 4
                    a = row[at + 3]
                    red += row[at] * a
                    green += row[at + 1] * a
                    blue += row[at + 2] * a
                    alpha += a
            if alpha == 0:
                line += b"\0\0\0\0"
            else:
                line += bytes((round(red / alpha), round(green / alpha), round(blue / alpha),
                               round(alpha / count)))
        out.append(bytes(line))
    _resized[size] = out
    return out


def png(size):
    rows = resized(size)
    raw = b"".join(b"\0" + row for row in rows)

    def chunk(kind, data):
        return struct.pack(">I", len(data)) + kind + data + struct.pack(">I", zlib.crc32(kind + data) & 0xFFFFFFFF)

    return (b"\x89PNG\r\n\x1a\n" + chunk(b"IHDR", struct.pack(">IIBBBBB", size, size, 8, 6, 0, 0, 0))
            + chunk(b"IDAT", zlib.compress(raw, 9)) + chunk(b"IEND", b""))


def write(path, data):
    with open(path, "wb") as f:
        f.write(data)


def main():
    kind = sys.argv[1] if len(sys.argv) > 1 else ""
    if kind == "png" and len(sys.argv) == 4:
        write(sys.argv[3], png(int(sys.argv[2])))
    elif kind == "iconset" and len(sys.argv) == 3:
        os.makedirs(sys.argv[2], exist_ok=True)
        made = {}
        for points in (16, 32, 128, 256, 512):
            for factor, suffix in ((1, ""), (2, "@2x")):
                size = points * factor
                made.setdefault(size, png(size))
                write(os.path.join(sys.argv[2], "icon_%dx%d%s.png" % (points, points, suffix)), made[size])
    elif kind == "ico" and len(sys.argv) == 3:
        # PNG images in an ICO, which Windows Vista and later read.
        images = [(size, png(size)) for size in (16, 24, 32, 48, 64, 128, 256)]
        header = struct.pack("<HHH", 0, 1, len(images))
        offset = len(header) + 16 * len(images)
        entries, data = b"", b""
        for size, image in images:
            entries += struct.pack("<BBBBHHII", size % 256, size % 256, 0, 0, 1, 32, len(image), offset + len(data))
            data += image
        write(sys.argv[2], header + entries + data)
    else:
        sys.exit(__doc__)


if __name__ == "__main__":
    main()
