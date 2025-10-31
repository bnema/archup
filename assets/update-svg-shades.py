#!/usr/bin/env python3
"""
Script to update SVG logo by wrapping all '░' characters in tspan elements
with a specific color class while keeping other characters unchanged.
"""

import re
import sys

def process_svg_line(line):
    """
    Process a text line in the SVG, wrapping all '░' characters in tspan elements.

    Args:
        line: A line from the SVG file

    Returns:
        Modified line with '░' characters wrapped in tspan elements
    """
    # Check if this is a text element line
    if '<text x=' not in line:
        return line

    # Extract the text content between > and </text>
    match = re.match(r'(.*<text[^>]*>)(.*)(</text>.*)', line)
    if not match:
        return line

    prefix, content, suffix = match.groups()

    # First, strip any existing tspan elements to get clean text
    clean_content = re.sub(r'<tspan[^>]*>(.*?)</tspan>', r'\1', content)

    # Split content into segments: shade characters and other characters
    segments = []
    current_segment = ""
    current_is_shade = False

    for char in clean_content:
        is_shade = (char == '░')

        if current_segment == "":
            # Start new segment
            current_segment = char
            current_is_shade = is_shade
        elif is_shade == current_is_shade:
            # Continue current segment
            current_segment += char
        else:
            # Save current segment and start new one
            segments.append((current_segment, current_is_shade))
            current_segment = char
            current_is_shade = is_shade

    # Don't forget the last segment
    if current_segment:
        segments.append((current_segment, current_is_shade))

    # Build the new content with tspan wrappers for shade characters
    new_content = ""
    for segment, is_shade in segments:
        if is_shade:
            new_content += f'<tspan class="shade">{segment}</tspan>'
        else:
            new_content += segment

    return prefix + new_content + suffix


def main():
    input_file = "archup-logo.svg"
    output_file = "archup-logo-updated.svg"

    try:
        with open(input_file, 'r', encoding='utf-8') as f:
            lines = f.readlines()

        # Process each line
        processed_lines = [process_svg_line(line) for line in lines]

        # Write output
        with open(output_file, 'w', encoding='utf-8') as f:
            f.writelines(processed_lines)

        print(f"✓ Successfully processed {input_file}")
        print(f"✓ Output written to {output_file}")
        print(f"\nTo replace the original file, run:")
        print(f"  mv {output_file} {input_file}")

    except FileNotFoundError:
        print(f"Error: {input_file} not found", file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
