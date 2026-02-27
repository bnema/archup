#!/usr/bin/env python3
"""
Generate arch-logo.png for Limine bootloader wallpaper.

Renders the Arch ASCII logo with the same cyan→slate-blue gradient used in
hyprlock.conf, on a deep navy background. Output is sized for 1920x1080.

Usage: python3 gen-limine-logo.py [--out path/to/arch-logo.png]
"""

import sys
import os
import argparse
from PIL import Image, ImageDraw, ImageFont

LOGO_LINES = [
    "                   -`",
    "                  .o+`",
    "                 `ooo/",
    "                `+oooo:",
    "               `+oooooo:",
    "               -+oooooo+:",
    "             `/:-:++oooo+:",
    "            `/++++/+++++++:",
    "           `/++++++++++++++:",
    "          `/+++ooooooooooooo/`",
    "         ./ooosssso++osssssso+`",
    "        .oossssso-````/ossssss+`",
    "       -osssssso.      :ssssssso.",
    "      :osssssss/        osssso+++.",
    "     /ossssssss/        +ssssooo/-",
    "   `/ossssso+/:-        -:/+osssso+-",
    "  `+sso+:-`                 `.-/+oso:",
    " `++:.                           `-/+/",
    " .`                                 `",
]

# Gradient: cyan at top → slate-blue at bottom (matching hyprlock.conf)
GRADIENT_COLORS = [
    (0,   212, 255),   # line 0  - bright cyan
    (0,   212, 255),   # line 1
    (0,   212, 255),   # line 2
    (0,   212, 255),   # line 3
    (0,   196, 239),   # line 4
    (0,   180, 223),   # line 5
    (34,  164, 207),   # line 6
    (68,  148, 191),   # line 7
    (85,  136, 204),   # line 8
    (85,  128, 192),   # line 9
    (80,  120, 180),   # line 10
    (74,  123, 167),   # line 11
    (74,  123, 167),   # line 12
    (90,  122, 154),   # line 13
    (96,  122, 148),   # line 14
    (101, 122, 144),   # line 15
    (106, 122, 140),   # line 16
    (110, 122, 136),   # line 17
    (112, 128, 144),   # line 18
]

# Background: deep navy from hyprlock
BG_COLOR = (5, 10, 20)

FONT_SIZE = 36
LINE_SPACING = 42  # px between lines
PADDING = 40       # px padding around the logo content

FONT_PATH = "/usr/share/fonts/TTF/JetBrainsMonoNerdFont-Thin.ttf"
FONT_PATH_FALLBACK = "/usr/share/fonts/TTF/JetBrainsMonoNerdFont-Regular.ttf"


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--out", default=os.path.join(os.path.dirname(__file__), "arch-logo.png"))
    args = parser.parse_args()

    # Load font
    font_path = FONT_PATH if os.path.exists(FONT_PATH) else FONT_PATH_FALLBACK
    try:
        font = ImageFont.truetype(font_path, FONT_SIZE)
    except Exception as e:
        print(f"Failed to load font {font_path}: {e}", file=sys.stderr)
        sys.exit(1)

    # Measure block dimensions using a throw-away draw surface
    tmp = Image.new("RGB", (4000, 2000), BG_COLOR)
    tmp_draw = ImageDraw.Draw(tmp)
    widths = [tmp_draw.textbbox((0, 0), line, font=font)[2] for line in LOGO_LINES]
    max_w = max(widths)
    total_h = len(LOGO_LINES) * LINE_SPACING

    # Canvas is just the logo + padding — Limine centers it on screen
    img_w = int(max_w) + PADDING * 2
    img_h = int(total_h) + PADDING * 2

    img = Image.new("RGB", (img_w, img_h), BG_COLOR)
    draw = ImageDraw.Draw(img)

    for i, line in enumerate(LOGO_LINES):
        color = GRADIENT_COLORS[i] if i < len(GRADIENT_COLORS) else GRADIENT_COLORS[-1]
        y = PADDING + i * LINE_SPACING
        draw.text((PADDING, y), line, font=font, fill=color)

    # Quantize to a small palette (no dithering) — cuts file size dramatically
    # while preserving the gradient colors faithfully at boot resolution.
    img = img.quantize(colors=32, dither=Image.Dither.NONE)
    img.save(args.out, "PNG", optimize=True)
    size_kb = os.path.getsize(args.out) / 1024
    print(f"Generated: {args.out} ({img_w}x{img_h}, {size_kb:.1f}K)")


if __name__ == "__main__":
    main()
