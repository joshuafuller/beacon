#!/usr/bin/env python3
"""
Extract all RFC 2119 normative requirements from RFC 6762.
"""

import re
import sys

def extract_requirements(rfc_file):
    """Parse RFC and extract all normative statements."""

    with open(rfc_file, 'r', encoding='utf-8', errors='replace') as f:
        content = f.read()

    # Remove page headers/footers (lines starting with "Cheshire & Krochmal" or "RFC 6762")
    lines = content.split('\n')
    cleaned_lines = []
    current_section = None
    section_pattern = re.compile(r'^(\d+\.(?:\d+\.)*)\s+(.+)$')

    for line in lines:
        # Skip page headers/footers
        if line.strip().startswith('Cheshire & Krochmal') or \
           line.strip().startswith('RFC 6762') or \
           line.strip().startswith('[Page '):
            continue
        cleaned_lines.append(line)

    content = '\n'.join(cleaned_lines)

    # Find all sections
    sections = []
    lines = content.split('\n')
    current_section_num = None
    current_section_title = None
    current_text = []

    for i, line in enumerate(lines):
        section_match = section_pattern.match(line.strip())
        if section_match:
            # Save previous section
            if current_section_num:
                sections.append({
                    'number': current_section_num,
                    'title': current_section_title,
                    'text': '\n'.join(current_text)
                })

            current_section_num = section_match.group(1)
            current_section_title = section_match.group(2)
            current_text = []
        else:
            current_text.append(line)

    # Save last section
    if current_section_num:
        sections.append({
            'number': current_section_num,
            'title': current_section_title,
            'text': '\n'.join(current_text)
        })

    # Extract requirements from each section
    requirements = []
    req_id = 1

    # RFC 2119 keywords
    keywords = r'\b(MUST NOT|MUST|REQUIRED|SHALL NOT|SHALL|SHOULD NOT|SHOULD|RECOMMENDED|MAY|OPTIONAL)\b'

    for section in sections:
        section_num = section['number']
        section_title = section['title']
        text = section['text']

        # Split into sentences
        # Look for sentences containing RFC 2119 keywords
        sentences = re.split(r'(?<=[.!?])\s+', text)

        for sentence in sentences:
            if re.search(keywords, sentence):
                # Clean up the sentence
                sentence = ' '.join(sentence.split())
                if len(sentence) > 10:  # Ignore very short matches
                    # Determine requirement type
                    req_type = 'MAY'
                    if 'MUST NOT' in sentence or 'SHALL NOT' in sentence:
                        req_type = 'MUST NOT'
                    elif 'MUST' in sentence or 'SHALL' in sentence or 'REQUIRED' in sentence:
                        req_type = 'MUST'
                    elif 'SHOULD NOT' in sentence:
                        req_type = 'SHOULD NOT'
                    elif 'SHOULD' in sentence or 'RECOMMENDED' in sentence:
                        req_type = 'SHOULD'

                    priority = 'P0' if 'MUST' in req_type else ('P1' if 'SHOULD' in req_type else 'P2')

                    requirements.append({
                        'id': f'RFC6762-§{section_num}-REQ-{req_id:03d}',
                        'section': f'§{section_num} {section_title}',
                        'type': req_type,
                        'requirement': sentence,
                        'priority': priority
                    })
                    req_id += 1

    return requirements

def main():
    rfc_file = sys.argv[1] if len(sys.argv) > 1 else '/home/user/development/beacon/RFC Docs/RFC-6762-Multicast-DNS.txt'

    requirements = extract_requirements(rfc_file)

    print(f"Total requirements found: {len(requirements)}")
    print(f"MUST: {len([r for r in requirements if 'MUST' in r['type']])}")
    print(f"SHOULD: {len([r for r in requirements if 'SHOULD' in r['type']])}")
    print(f"MAY: {len([r for r in requirements if r['type'] == 'MAY'])}")
    print()

    for req in requirements[:10]:  # Print first 10 as sample
        print(f"{req['id']}")
        print(f"  Type: {req['type']}")
        print(f"  Section: {req['section']}")
        print(f"  Requirement: {req['requirement'][:100]}...")
        print()

if __name__ == '__main__':
    main()
